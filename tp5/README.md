# TP5 — Projet final : `gopack`

**Jour 5 — Durée : 2h45** (suivi du QCM final, 1h)

Assemblage du fil rouge de la semaine : `gopack`, un outil CLI de sauvegarde et transfert de fichiers, livré en binaires cross-compilés pour 3 plateformes. Vous réutilisez les briques des TP2 → TP4.

## Cahier des charges

`gopack` est une CLI à sous-commandes, construite avec `flag.NewFlagSet` :

| Commande | Rôle |
|---|---|
| `gopack scan <dir>` | Inventaire du répertoire : parcours récursif (`filepath.WalkDir`), stats par extension (nombre de fichiers + taille cumulée), total général |
| `gopack backup <src> <dst>` | Copie récursive avec **préservation des permissions** des fichiers et répertoires |
| `gopack fetch [-o dir] [-n N] <url>...` | Téléchargement concurrent (goroutines + channels + sémaphore) — réutilise le TP4 |
| `gopack version` / `gopack help` | Version injectée au build, aide globale |

**Exigences transverses :**

- **Aide** : menu d'aide global (`gopack help`) **et** par sous-commande (`gopack scan -h`...) via un `flag.Usage` custom
- **Progression** : barre de progression ou compteur d'avancement (`\r` ou `[i/total]`) sur `backup` et `fetch`
- **Erreurs** : remontées jusqu'au `main` (wrapping `%w`), affichées sur **stderr**, codes de sortie : `0` succès, `1` erreur, `2` erreur d'usage ; sur `fetch`, une URL en échec n'arrête pas le lot
- **Organisation** : un `main.go` de dispatch + packages `internal/scan`, `internal/backup`, `internal/fetch`
- **Cross-compilation obligatoire** : script `build.sh` livrant les binaires `linux/amd64`, `windows/amd64` (`.exe`), `linux/arm64` dans `dist/`

## Structure du TP

| Dossier | Rôle |
|---|---|
| `starter/` | `go.mod` + `main.go` (dispatch fourni, TODO) + `build.sh` à compléter |
| `solution/` | Projet complet (distribué après l'évaluation) |

## Prérequis

- Go 1.22+, vos codes des TP2, TP3 et TP4
- Pour tester `fetch` : le serveur du TP4 (`go run ../tp4-telechargeur/server`) ou n'importe quelle URL publique

---

## Étape 0 — Mise en place (15 min)

```bash
cd starter
go run . help          # le squelette compile et affiche l'aide
go run . scan .        # → "pas encore implémenté"
```

Lisez `main.go` : le dispatch des sous-commandes (`switch os.Args[1]`), la convention d'erreurs (remontée au `main`, stderr, codes de sortie) et les trois `flag.NewFlagSet` sont déjà en place.

**✅ Checkpoint 0 :** `go run . help` et `go run . scan -h` fonctionnent.

---

## Étape 1 — `gopack scan` (40 min)

Créez `internal/scan/scan.go` (réutilisez vos TP2/TP3) :

- `filepath.WalkDir` sur le répertoire ; ignorez ce qui n'est pas un fichier régulier
- Agrégation dans une `map[string]...` par extension (`filepath.Ext`, pensez au cas « sans extension »)
- Affichage : tableau trié (par taille décroissante, par exemple), tailles lisibles (Ko/Mo), total général
- Séparez le **calcul** (fonction qui retourne un rapport) de l'**affichage** : c'est ce qui rend le code testable

```bash
go run . scan ~/dev
```

**✅ Checkpoint 1 :** le scan d'un vrai répertoire affiche les stats par extension + un total cohérent avec `du -sh`.

---

## Étape 2 — `gopack backup` (40 min)

Créez `internal/backup/backup.go` :

- `filepath.WalkDir` sur `src`, chemin cible reconstruit avec `filepath.Rel` + `filepath.Join`
- Répertoires : `os.MkdirAll(target, info.Mode().Perm())`
- Fichiers : `os.Open` + `os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)` + `io.Copy`, puis `os.Chmod` (le umask modifie les permissions à la création !)
- Compteur d'avancement pendant la copie (`\rCopie... N fichier(s)`)

```bash
go run . backup ./testdata /tmp/sauvegarde
ls -lR /tmp/sauvegarde     # comparez les permissions avec la source
```

> ⚠️ Pensez au piège classique : que se passe-t-il si `<dst>` est *dans* `<src>` ?

**✅ Checkpoint 2 :** un `chmod 600` sur un fichier source se retrouve à l'identique dans la copie ; l'arborescence est complète.

---

## Étape 3 — `gopack fetch` (45 min)

Créez `internal/fetch/fetch.go` en **réutilisant votre TP4** :

- Une goroutine par URL, channel de résultats, `sync.WaitGroup`, `close` dans une goroutine dédiée
- Sémaphore `make(chan struct{}, workers)` branché sur le flag `-n`
- Écriture en `0644` dans le répertoire `-o`, streaming `io.Copy`
- Erreur par URL sans arrêter le lot, compteur `[i/total]`, bilan final

```bash
# terminal 1 : le serveur du TP4
go run ../../tp4-telechargeur/server

# terminal 2 :
go run . fetch -o /tmp/dl -n 3 \
  http://localhost:8080/files/notes.txt \
  http://localhost:8080/files/archive.zip \
  http://localhost:8080/files/inexistant.txt
echo $?   # → 1 (une URL en échec)
```

**✅ Checkpoint 3 :** les 2 fichiers valides sont téléchargés, l'URL invalide est signalée sur stderr, code de sortie 1, et `go run -race .` ne détecte rien.

---

## Étape 4 — Aide & UX (15 min)

- `flag.Usage` custom sur **chaque** `FlagSet` : une ligne d'usage, une description, `fs.PrintDefaults()`
- `gopack` sans argument et commande inconnue → aide + code de sortie 2
- Messages d'erreur : préfixe `gopack :`, sur stderr uniquement

**✅ Checkpoint 4 :** `gopack fetch -h` documente `-o` et `-n` ; `gopack toto ; echo $?` affiche l'aide et `2`.

---

## Étape 5 — Cross-compilation (10 min)

Complétez `build.sh` : matrice de build sur les 3 plateformes cibles, `-ldflags "-s -w -X main.version=..."`, extension `.exe` pour Windows, checksums `SHA256SUMS`.

```bash
./build.sh 1.0.0
ls -lh dist/
file dist/*                          # vérifie les architectures
./dist/gopack-linux-amd64 version    # → gopack 1.0.0
```

**✅ Checkpoint 5 :** `dist/` contient `gopack-linux-amd64`, `gopack-windows-amd64.exe`, `gopack-linux-arm64` et `SHA256SUMS`.

---

## Barème (sur 20)

| Critère | Points | Détail |
|---|---|---|
| **Fonctionnalité** | **10** | |
| `scan` | 3 | WalkDir, stats par extension, total |
| `backup` | 3 | Copie récursive, permissions préservées |
| `fetch` | 4 | Concurrence réelle (goroutines/channels/WaitGroup), sémaphore `-n`, erreurs par URL |
| **Qualité du code / organisation** | **4** | Packages `internal/` cohérents, fonctions courtes, nommage, code idiomatique (`gofmt`, `go vet` propres) |
| **Gestion d'erreurs** | **3** | Wrapping `%w`, remontée au `main`, stderr, codes de sortie corrects, jamais d'erreur avalée |
| **Aide + UX** | **2** | Aide globale + par sous-commande, progression/compteur, messages clairs |
| **Cross-compilation** | **1** | `build.sh` fonctionnel, 3 binaires + checksums |

**Validation : ≥ 10/20.** Bonus possibles (+1 max) : flag `-v` verbose, `scan` avec filtre d'extension, `fetch` qui lit la liste `/files` du serveur TP4.

## Livrables

1. Le repo du projet (`go build .` doit passer sans erreur)
2. Les binaires `dist/gopack-linux-amd64`, `dist/gopack-windows-amd64.exe`, `dist/gopack-linux-arm64` + `SHA256SUMS`
3. `build.sh` complété

> 🕐 Après le rendu : QCM final (1h, 20 questions, validation C31/C32).
