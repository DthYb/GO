# TP2 — Inventaire de fichiers en mémoire

**Durée :** 3h30 (J2 après-midi)
**Objectif :** modéliser un inventaire de fichiers avec structs, slices et maps, écrire des fonctions de filtrage à retours multiples, des statistiques et un tri — le tout sur des données en dur.

> 🧵 **Fil rouge `gopack` :** c'est la **première brique** de l'outil de sauvegarde que vous construirez toute la semaine. Aujourd'hui, l'inventaire vit en mémoire avec des données simulées ; au jour 4, les mêmes structures seront remplies par un vrai scan de répertoire. Soignez ce code : vous le réutiliserez.

## Pré-requis

- TP1 validé : environnement Go + VSCode fonctionnels
- Cours du matin : fonctions (retours multiples), `for` / `range`, pointeurs, slices, maps, structs

## Mise en place

Créez un dossier `tp2/` et copiez-y le fichier de départ [`starter/main.go`](starter/main.go). Il contient les **données en dur** (12 fichiers simulés) et les TODO de chaque étape. Vérifiez qu'il compile avant de commencer :

```bash
go run main.go
# → TP2 — inventaire de fichiers : à vous de jouer !
```

## Plan du TP

| # | Étape | Durée |
|---|-------|-------|
| 1 | La struct `File` | 20 min |
| 2 | L'inventaire : un slice de structs | 20 min |
| 3 | Filtrage avec retours multiples `(résultat, erreur)` | 45 min |
| 4 | Statistiques par extension avec une map | 40 min |
| 5 | Tri avec `sort.Slice` | 25 min |
| 6 | Affichage formaté | 20 min |
| 7 | Mini-challenge : méthodes & pointer receiver | 40 min |

Si vous bloquez, le corrigé complet est dans [`corrige/main.go`](corrige/main.go).

---

## Étape 1 — La struct `File` (20 min)

Définissez le type central du projet :

```go
type File struct {
	Name      string    // nom du fichier, ex : "rapport.pdf"
	Size      int64     // taille en octets
	Extension string    // extension avec le point, ex : ".pdf"
	Modified  time.Time // date de dernière modification
	Tag       string    // étiquette libre, vide par défaut (utilisée à l'étape 7)
}
```

Pourquoi ces choix ?

- `Size` est un `int64` : c'est le type que renvoie la stdlib (`os.FileInfo.Size()`) — autant être compatibles dès maintenant pour le jour 4
- `Tag` illustre la **zero value** : un champ non renseigné vaut `""` automatiquement

Dans `main`, instanciez **un** fichier à la main et affichez-le :

```go
f := File{Name: "test.txt", Size: 100, Extension: ".txt", Modified: time.Now()}
fmt.Println(f)
fmt.Printf("%+v\n", f) // %+v affiche aussi les noms des champs
```

**Checkpoints ✅**
- [ ] Le programme compile et affiche votre fichier de test
- [ ] Vous savez expliquer la différence d'affichage entre `%v` et `%+v`
- [ ] Que vaut `f.Tag` alors que vous ne l'avez jamais renseigné ? Pourquoi ?

---

## Étape 2 — L'inventaire : un slice de structs (20 min)

Décommentez le bloc `var inventaire = []File{...}` du starter (et ajoutez `"time"` aux imports). Ces 12 fichiers simulent le contenu d'un dossier de projet.

Dans `main` :

1. Affichez le nombre de fichiers avec `len(inventaire)`
2. Parcourez l'inventaire avec `for i, f := range inventaire` et affichez `numéro. nom (taille octets)` pour chaque fichier
3. Ajoutez un 13ᵉ fichier avec `append` :
   ```go
   inventaire = append(inventaire, File{Name: "notes-perso.txt", Size: 256, Extension: ".txt", Modified: time.Now()})
   ```
   ⚠️ N'oubliez pas de **réaffecter** le résultat : `append` peut retourner un nouveau slice (réallocation du backing array, vue en démo ce matin).

**Checkpoints ✅**
- [ ] 12 fichiers affichés, puis 13 après l'`append`
- [ ] Vous savez pourquoi on écrit `inventaire = append(inventaire, ...)` et jamais `append(inventaire, ...)` tout seul
- [ ] Question : dans `for i, f := range`, `f` est-il le fichier du slice ou une copie ? *(testez : modifiez `f.Name` dans la boucle, puis réaffichez l'inventaire)*

---

## Étape 3 — Filtrage avec retours multiples (45 min)

Écrivez deux fonctions de filtrage. Chacune retourne **deux valeurs** : le résultat **et** une erreur — le pattern `(valeur, error)` omniprésent en Go (on le creusera demain, aujourd'hui on prend le réflexe).

```go
// filtrerParExtension retourne les fichiers dont l'extension correspond.
// Erreur si ext est vide ou ne commence pas par un point.
func filtrerParExtension(fichiers []File, ext string) ([]File, error)

// filtrerParTailleMin retourne les fichiers d'au moins tailleMin octets.
// Erreur si tailleMin est négative.
func filtrerParTailleMin(fichiers []File, tailleMin int64) ([]File, error)
```

### Consignes

1. En cas de paramètre invalide, retournez `nil` et une erreur créée avec `errors.New("...")` (import `"errors"`) :
   ```go
   if ext == "" {
       return nil, errors.New("extension vide")
   }
   ```
2. Sinon, construisez le résultat avec un slice vide + `append` dans un `range`, puis retournez `resultat, nil`
3. **Important :** la fonction retourne un **nouveau slice**, elle ne modifie jamais `fichiers`
4. Côté appelant (dans `main`), gérez l'erreur systématiquement :
   ```go
   pdfs, err := filtrerParExtension(inventaire, ".pdf")
   if err != nil {
       fmt.Println("Erreur de filtrage :", err)
       os.Exit(1)
   }
   ```

### Tests attendus

| Appel | Résultat |
|---|---|
| `filtrerParExtension(inventaire, ".go")` | 2 fichiers |
| `filtrerParExtension(inventaire, "go")` | erreur (pas de point) |
| `filtrerParTailleMin(inventaire, 1_000_000)` | 5 fichiers (≥ 1 Mo) |
| `filtrerParTailleMin(inventaire, -1)` | erreur |

**Checkpoints ✅**
- [ ] Les 4 tests ci-dessus donnent le résultat attendu
- [ ] Chaîner les deux filtres fonctionne : les `.png` de plus de 1 Mo → 1 fichier (`photo-equipe.png`)
- [ ] Aucun filtre ne modifie l'inventaire d'origine (réaffichez `len(inventaire)` après)

---

## Étape 4 — Statistiques par extension (40 min)

Objectif : produire un récapitulatif par extension grâce à une **map**.

1. Définissez une petite struct pour les valeurs :
   ```go
   type Stats struct {
       Nombre       int   // nombre de fichiers
       TailleTotale int64 // somme des tailles en octets
   }
   ```
2. Écrivez la fonction :
   ```go
   func statistiquesParExtension(fichiers []File) map[string]Stats
   ```
   Parcourez les fichiers ; pour chaque extension, incrémentez le compteur et cumulez la taille.

   > 💡 Une map de structs ne permet pas `m[ext].Nombre++` (la valeur d'une map n'est pas adressable). Le pattern : lire la valeur, la modifier, la réécrire :
   > ```go
   > s := stats[f.Extension] // zero value {0, 0} si absente : parfait !
   > s.Nombre++
   > s.TailleTotale += f.Size
   > stats[f.Extension] = s
   > ```
3. Dans `main`, affichez le tableau des stats avec `for ext, s := range stats`
4. Testez aussi l'**idiome `ok`** sur une extension précise :
   ```go
   if s, ok := stats[".zip"]; ok {
       fmt.Println("Les .zip pèsent", s.TailleTotale, "octets")
   }
   ```

**Checkpoints ✅**
- [ ] `.go` → 2 fichiers ; `.zip` → 2 fichiers pour 57 Mo ; `.md` → 2 fichiers
- [ ] `stats[".xyz"]` (extension absente) ne panique pas — que retourne-t-elle et pourquoi ?
- [ ] Lancez le programme deux fois : l'ordre d'affichage des extensions change-t-il ? *(l'itération d'une map est volontairement non déterministe en Go — on règle ça à l'étape 5)*

---

## Étape 5 — Tri avec `sort.Slice` (25 min)

1. Écrivez :
   ```go
   func trierParTaille(fichiers []File) []File
   ```
   qui retourne les fichiers triés **par taille décroissante**, **sans modifier** le slice d'origine :
   ```go
   tries := make([]File, len(fichiers))
   copy(tries, fichiers) // copie de travail : l'original reste intact

   sort.Slice(tries, func(i, j int) bool {
       return tries[i].Size > tries[j].Size // > : ordre décroissant
   })
   return tries
   ```
   `sort.Slice` prend une **fonction anonyme** (vue ce matin) qui répond à : « l'élément `i` doit-il passer avant l'élément `j` ? »
2. Dans `main`, affichez le **top 3** des fichiers les plus lourds
3. Bonus : rendez l'affichage des stats de l'étape 4 déterministe — collectez les clés de la map dans un slice, triez-les avec `sort.Strings`, puis itérez sur les clés triées

**Checkpoints ✅**
- [ ] Top 3 : `archive-2025.zip`, `backup.zip`, `presentation.pdf`
- [ ] L'inventaire d'origine n'est PAS trié après l'appel (vérifiez son premier élément)
- [ ] Vous savez expliquer le rôle de la fonction `func(i, j int) bool`

---

## Étape 6 — Affichage formaté (20 min)

Écrivez une fonction d'affichage propre pour l'inventaire :

```go
func afficher(fichiers []File)
```

Sortie attendue (colonnes alignées, tailles lisibles) :

```text
NOM                        TAILLE     MODIFIÉ       TAG
main.go                    4.0 Ko     2026-07-01
archive-2025.zip           42.0 Mo    2025-12-31
...
```

### Consignes

1. Alignez avec les largeurs de `fmt.Printf` : `%-25s` (aligné à gauche sur 25 caractères), `%-10s`…
2. Écrivez un helper `tailleLisible(octets int64) string` qui convertit en `o` / `Ko` / `Mo` selon la valeur (1 Ko = 1024 o). C'est un `switch` sans expression parfait.
3. Formatez la date avec `f.Modified.Format("2006-01-02")` — la date de référence magique de Go

**Checkpoints ✅**
- [ ] Les colonnes sont alignées quel que soit le nom de fichier
- [ ] `512` → `512 o`, `4096` → `4.0 Ko`, `44040192` → `42.0 Mo`

---

## Étape 7 — Mini-challenge : méthodes & pointer receiver (40 min)

L'inventaire doit pouvoir être **modifié** : renommer un fichier, marquer des fichiers pour la future sauvegarde. C'est le rôle des **méthodes** — et le choix du receiver (valeur ou pointeur) devient crucial.

### Cahier des charges

1. **Méthode en lecture** (value receiver — pas de mutation) :
   ```go
   func (f File) TailleLisible() string
   ```
   Réutilisez le helper de l'étape 6 ; remplacez son usage dans `afficher` par `f.TailleLisible()`.

2. **Méthodes en écriture** (pointer receiver — mutation du struct) :
   ```go
   // Renommer change le nom ET met à jour l'extension en conséquence.
   func (f *File) Renommer(nouveauNom string)

   // Marquer pose une étiquette sur le fichier (champ Tag).
   func (f *File) Marquer(tag string)
   ```
   💡 Pour l'extension, `filepath.Ext("rapport.pdf")` → `".pdf"` (import `"path/filepath"`).

3. **Fonction de marquage en masse** :
   ```go
   // marquerLesLourds pose le tag "a-archiver" sur tous les fichiers
   // d'au moins tailleMin octets, et retourne le nombre de fichiers marqués.
   func marquerLesLourds(fichiers []File, tailleMin int64) int
   ```
   ⚠️ **Piège du jour :** `for _, f := range fichiers { f.Marquer(...) }` marque… une **copie**, et l'inventaire reste vierge. Modifiez l'élément du slice via son **index** : `fichiers[i].Marquer(...)` — Go prend automatiquement l'adresse (`&fichiers[i]`) pour appeler la méthode à pointer receiver.

4. Scénario final dans `main` :
   - Renommez `todo.txt` en `todo-2026.md` → vérifiez que l'extension suit
   - Marquez les fichiers ≥ 1 Mo → la fonction doit retourner `5`
   - Réaffichez l'inventaire : la colonne TAG montre les fichiers marqués

**Checkpoints ✅**
- [ ] `Renommer` sur `todo.txt` change bien `Name` **et** `Extension` dans l'inventaire (pas sur une copie)
- [ ] `marquerLesLourds(inventaire, 1_000_000)` retourne 5 et les tags sont visibles à l'affichage
- [ ] Expérience : passez `Marquer` en value receiver `(f File)` — que se passe-t-il ? Remettez le pointeur et sachez expliquer.

---

## Livrable J2

À la fin du TP vous devez avoir un `main.go` qui, exécuté d'un seul `go run main.go`, déroule :

- ✅ Struct `File` + inventaire de 13 fichiers (12 en dur + 1 `append`)
- ✅ Les 2 filtres à retours multiples, avec au moins un cas d'erreur géré
- ✅ Les statistiques par extension (map + idiome `ok`)
- ✅ Le top 3 par taille (`sort.Slice`, original préservé)
- ✅ L'affichage formaté avec tailles lisibles
- ✅ Le scénario renommage + marquage en masse (pointer receivers)

**Démo de fin :** être capable de rejouer en live :
- Un filtrage chaîné (extension puis taille) avec gestion d'erreur
- Le piège `range`-copie vs `fichiers[i]` sur le marquage

## Pour aller plus loin

- Implémentez `func supprimer(fichiers []File, nom string) ([]File, error)` qui retire un fichier par son nom (astuce : `append(fichiers[:i], fichiers[i+1:]...)` — dessinez ce qui se passe sur le backing array !)
- Ajoutez `func plusRecentQue(fichiers []File, date time.Time) []File` avec `f.Modified.After(date)`
- Remplacez `sort.Slice` par le plus moderne `slices.SortFunc` (package `slices`, Go 1.21+) et comparez les signatures
