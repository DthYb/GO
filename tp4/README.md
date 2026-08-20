# TP4 — Téléchargeur concurrent + serveur de fichiers

**Jour 4 — Durée : 3h30**

Deuxième brique du fil rouge `gopack` : un serveur HTTP de fichiers vous est fourni, vous écrivez le client qui télécharge tous ses fichiers **en parallèle**.

## Objectifs

- Faire des requêtes HTTP avec `net/http` (GET + décodage JSON)
- Écrire des fichiers sur disque avec les bonnes permissions (`0644`)
- Streamer avec `io.Copy` (sans charger le fichier en mémoire)
- Paralléliser avec goroutines + channels + `sync.WaitGroup`
- Gérer les erreurs **par fichier** sans arrêter le lot
- (Bonus) Limiter la concurrence avec un channel sémaphore

## Prérequis

- Go 1.22+ (`go version`)
- Les acquis du cours du matin : goroutines, channels, `select`, `sync.WaitGroup`, fichiers (`os`, `io`, `filepath`)

## Structure du TP

| Dossier | Rôle |
|---|---|
| `server/` | Serveur HTTP de fichiers **fourni** (ne pas modifier) |
| `starter/` | Squelette du client, avec des TODO — **votre travail** |
| `corrige/` | Corrigé complet (distribué en fin de TP) |

---

## Étape 1 — Lancer et explorer le serveur (20 min)

Dans un **premier terminal**, depuis la racine du TP :

```bash
go run ./server
```

Le serveur génère ~10 fichiers de tailles variées (2 Ko à 5 Mo) dans un répertoire temporaire et expose :

| Endpoint | Rôle |
|---|---|
| `GET /files` | Liste des fichiers en JSON (`[{"name": "...", "size": ...}]`) |
| `GET /files/{name}` | Contenu du fichier |

Testez dans un **second terminal** :

```bash
curl http://localhost:8080/files
curl -o /tmp/notes.txt http://localhost:8080/files/notes.txt
curl -i http://localhost:8080/files/inexistant.txt   # → 404
```

> ⏱️ Vous remarquerez que chaque téléchargement prend ~0,5 s : le serveur ajoute une **latence artificielle** pour simuler un vrai réseau. C'est elle qui rendra le gain de la concurrence spectaculaire.

Prenez 5 minutes pour **lire `server/main.go`** : routage `GET /files/{name}` (Go 1.22), `r.PathValue`, `http.ServeFile`, protection contre le path traversal.

**✅ Checkpoint 1 :** `curl http://localhost:8080/files` renvoie un tableau JSON de ~10 fichiers.

---

## Étape 2 — Télécharger UN fichier (45 min)

Ouvrez `starter/main.go`. La fonction `fetchList` (liste JSON) est fournie. À vous d'implémenter `downloadFile` puis de l'appeler **une seule fois** depuis `main` (sur `files[0]`) :

1. `http.Get(serverURL + "/files/" + name)` — n'oubliez pas `defer resp.Body.Close()` et le contrôle de `resp.StatusCode`
2. Créez le répertoire de destination : `os.MkdirAll(outDir, 0o755)`
3. Créez le fichier avec les permissions exactes `0644` :

```go
out, err := os.OpenFile(filepath.Join(outDir, name),
    os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
```

> ❓ Pourquoi pas `os.Create` ? Parce que `os.Create` utilise `0666` (modulé par le umask). Le TP exige `0644` : lecture/écriture pour le propriétaire, lecture seule pour les autres.

4. Copiez le corps de la réponse : `io.Copy(out, resp.Body)` — le fichier est **streamé**, jamais chargé entièrement en mémoire

Test :

```bash
go run ./starter
ls -l downloads/
# → -rw-r--r--  (= 0644)
```

**✅ Checkpoint 2 :** un fichier est présent dans `downloads/` avec les permissions `-rw-r--r--`, et sa taille correspond au champ `size` du JSON.

---

## Étape 3 — Paralléliser avec goroutines + channels (1h)

Téléchargez maintenant **tous** les fichiers en parallèle. Architecture attendue :

```
main ──┬─ go download(f1) ─┐
       ├─ go download(f2) ─┤──▶ chan Result ──▶ range (affichage au fil de l'eau)
       ├─ go download(f…) ─┘
       └─ go { wg.Wait(); close(results) }
```

1. Un channel de résultats : `results := make(chan Result)`
2. Un `sync.WaitGroup` ; pour chaque fichier : `wg.Add(1)` puis `go func(f FileInfo) { defer wg.Done(); ... }(f)`
3. Chaque goroutine appelle `downloadFile` et envoie un `Result` dans le channel — **succès comme erreur**
4. Une goroutine dédiée fait `wg.Wait()` puis `close(results)` — c'est ce `close` qui termine le `range`
5. Dans `main` : `for res := range results { ... }` affiche chaque résultat dès qu'il arrive

Mesurez le gain :

```bash
time go run ./starter        # ~10 fichiers × ~0,5s de latence...
```

> 🔍 Vérifiez l'absence de condition de course : `go run -race ./starter`

**✅ Checkpoint 3 :** tous les fichiers sont téléchargés, le temps total est proche du fichier le plus lent (~1-2 s) au lieu de la somme (~5-8 s), et `-race` ne détecte rien.

---

## Étape 4 — Erreurs par fichier + bilan (35 min)

Un lot de téléchargements ne doit **jamais** s'arrêter à la première erreur.

1. Le champ `Err` de `Result` transporte l'erreur de la goroutine vers `main` — la goroutine ne fait ni `log.Fatal`, ni `panic`
2. Dans le `range`, comptez succès et échecs ; affichez `❌ nom : erreur` sans interrompre la boucle
3. En fin de programme : bilan (`X succès, Y échecs`) et `os.Exit(1)` si au moins un échec (les erreurs vont sur `os.Stderr`)

Pour tester le cas d'erreur, ajoutez temporairement un fichier piégé à la liste :

```go
files = append(files, FileInfo{Name: "inexistant.txt"})
```

**✅ Checkpoint 4 :** avec le fichier piégé, le programme télécharge les 10 vrais fichiers, affiche 1 échec, et son code de sortie est 1 (`echo $?`).

---

## Étape 5 — Bonus : sémaphore + progression (40 min)

Lancer 10 000 goroutines sur 10 000 fichiers saturerait le serveur. Limitez la concurrence à `n` téléchargements simultanés avec un **channel sémaphore** :

```go
sem := make(chan struct{}, n)   // n jetons disponibles

// dans chaque goroutine :
sem <- struct{}{}        // prend un jeton (bloque si n téléchargements en cours)
defer func() { <-sem }() // rend le jeton
```

1. Ajoutez un flag `-n` (défaut : 4) et branchez le sémaphore
2. Affichez un **compteur d'avancement** au fil des résultats : `[3/10] ✅ rapport.pdf`
3. Comparez `-n 1` (séquentiel), `-n 4` et `-n 10` avec `time`

**✅ Checkpoint bonus :** avec `-n 1` le temps total redevient ~la somme des latences ; avec `-n 10` il redescend au niveau de l'étape 3.

---

## Livrable

Votre `starter/main.go` complété :

- [ ] Télécharge tous les fichiers du serveur en parallèle (goroutines + channel de résultats + `WaitGroup`)
- [ ] Fichiers écrits dans `downloads/` en `0644`, streamés via `io.Copy`
- [ ] Une erreur sur un fichier n'arrête pas le lot ; bilan final + code de sortie ≠ 0 en cas d'échec
- [ ] `go run -race ./starter` ne détecte aucune condition de course
- [ ] (Bonus) flag `-n` + sémaphore + compteur `[i/total]`

> 💡 Gardez ce code sous la main : la sous-commande `fetch` du projet final `gopack` (TP5) le réutilise directement.
