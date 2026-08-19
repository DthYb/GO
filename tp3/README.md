# TP3 — Refactor de l'inventaire en projet structuré

**Durée :** 3h30 (J3 après-midi)
**Objectif :** transformer le programme monolithique du TP2 en vrai projet Go : module `gopack`, packages `inventory` et `format`, erreurs propres remontées jusqu'au `main` (`errors.New`, `fmt.Errorf` + `%w`, `errors.Is`), interface `Formatter` avec deux implémentations (texte et JSON), et un premier package externe (`fatih/color`).

> 🧵 **Fil rouge `gopack` :** aujourd'hui votre inventaire devient un **module** nommé `gopack`. C'est cette structure (domaine dans `inventory/`, affichage dans `format/`) qui accueillera le scan disque et le réseau au jour 4, puis la CLI finale au jour 5. Le nom du module n'est pas négociable : les corrigés des jours suivants s'appuient dessus.

## Pré-requis

- TP2 terminé : votre `main.go` d'inventaire fonctionne (sinon, partez du corrigé [`../tp2-inventaire/corrige/main.go`](../tp2-inventaire/corrige/main.go))
- Cours du matin : gestion des erreurs, interfaces, `go mod` / packages
- Accès internet pour `go get` (proxy Go : `proxy.golang.org`)

## Plan du TP

| # | Étape | Durée |
|---|-------|-------|
| 1 | `go mod init gopack` + arborescence | 30 min |
| 2 | Extraire le package `inventory` | 45 min |
| 3 | Erreurs sentinelles, wrapping `%w`, `errors.Is` | 45 min |
| 4 | Interface `Formatter` : texte & JSON | 1h |
| 5 | Package externe : `fatih/color` | 30 min |

Le projet corrigé complet est dans [`corrige/`](corrige/) — arborescence, `go.mod` inclus.

---

## Étape 1 — Naissance du module `gopack` (30 min)

### 1.1 — Initialiser le module

```bash
mkdir gopack && cd gopack
go mod init gopack
cat go.mod
```

Le fichier `go.mod` déclare le **nom du module** (la racine de tous vos imports internes) et la version de Go. C'est lui qui fait de ce dossier un projet.

> 💡 Dans la vraie vie on nomme un module par son URL de dépôt (`github.com/vous/gopack`) pour qu'il soit installable par d'autres. `gopack` court suffit pour la formation.

### 1.2 — Arborescence cible

```text
gopack/
├── go.mod
├── main.go              ← point d'entrée : orchestration + gestion des erreurs
├── inventory/
│   └── inventory.go     ← le DOMAINE : File, filtres, stats, tri
└── format/
    ├── format.go        ← l'interface Formatter
    ├── text.go          ← implémentation texte
    └── json.go          ← implémentation JSON
```

La règle de découpage : **un package = une responsabilité**. `inventory` ne sait pas afficher ; `format` ne sait pas filtrer ; `main` assemble les deux.

### 1.3 — Premier déplacement

1. Copiez votre `main.go` du TP2 à la racine de `gopack/`
2. Créez les dossiers `inventory/` et `format/` (vides pour l'instant)
3. Vérifiez que tout compile encore : `go run .`

> 💡 Notez le `go run .` (le **package** courant) et non plus `go run main.go` : à partir de maintenant on raisonne en packages, plus en fichiers.

**Checkpoints ✅**
- [ ] `go.mod` existe et contient `module gopack`
- [ ] `go run .` exécute votre programme du TP2 à l'identique

---

## Étape 2 — Extraire le package `inventory` (45 min)

Déplacez tout le **domaine** dans `inventory/inventory.go` :

- la struct `File` (et `Stats`)
- les filtres, les statistiques, le tri
- les méthodes `TailleLisible`, `Renommer`, `Marquer`

### Consignes

1. Le fichier commence par `package inventory` (plus `package main`) et un commentaire de package :
   ```go
   // Package inventory contient le domaine de gopack : le type File,
   // les filtres, les statistiques et le tri de l'inventaire.
   package inventory
   ```
2. **Visibilité par la casse** — c'est LE point de l'étape :
   - tout ce que `main` doit utiliser prend une **Majuscule** : `File`, `Stats`, `FiltrerParExtension`, `FiltrerParTailleMin`, `StatistiquesParExtension`, `TrierParTaille`
   - les helpers internes restent en **minuscule** : `tailleLisible` par exemple n'a aucune raison de sortir du package
3. Dans `main.go`, importez le package par son chemin complet depuis la racine du module, et préfixez chaque usage :
   ```go
   import "gopack/inventory"

   fichiers := []inventory.File{ ... }
   pdfs, err := inventory.FiltrerParExtension(fichiers, ".pdf")
   ```
4. Les **données en dur** restent dans `main.go` (fonction `donnees()`) : ce sont des données d'exemple, pas du domaine. Au jour 4, cette fonction sera remplacée par un vrai scan disque — le package `inventory` n'aura pas à bouger.
5. La fonction d'affichage du TP2 peut rester provisoirement dans `main.go` : elle déménagera dans `format/` à l'étape 4.

**Checkpoints ✅**
- [ ] `go run .` fonctionne toujours, `go vet ./...` ne râle pas
- [ ] Depuis `main.go`, tapez `inventory.` : l'autocomplétion ne propose QUE les identifiants exportés
- [ ] Essayez d'appeler `inventory.tailleLisible(42)` depuis `main` : quelle erreur obtenez-vous ? *(c'est la visibilité par la casse en action)*

---

## Étape 3 — Erreurs sentinelles, wrapping, `errors.Is` (45 min)

Au TP2, vos filtres retournaient des `errors.New(...)` anonymes : impossible pour l'appelant de distinguer « paramètre invalide » de « aucun résultat » sans comparer des strings (fragile !). On professionnalise.

### 3.1 — Erreurs sentinelles dans `inventory`

En tête de `inventory.go`, déclarez des erreurs **exportées** et réutilisables :

```go
var (
	ErrExtensionInvalide = errors.New("extension invalide (attendu : \".ext\")")
	ErrTailleNegative    = errors.New("taille minimale négative")
	ErrAucunResultat     = errors.New("aucun fichier ne correspond")
)
```

### 3.2 — Wrapping avec `%w`

Dans les filtres, **enrichissez** la sentinelle avec le contexte, sans la perdre — c'est tout l'intérêt du verbe `%w` :

```go
if !strings.HasPrefix(ext, ".") {
	return nil, fmt.Errorf("filtre extension %q : %w", ext, ErrExtensionInvalide)
}
```

Nouvelle règle métier : un filtre qui ne trouve **rien** retourne `ErrAucunResultat` (wrappée avec l'extension demandée).

### 3.3 — Remontée jusqu'au `main`

Les couches intermédiaires ne décident rien : elles **wrappent et remontent**. Seul `main` choisit la réaction, avec `errors.Is` qui traverse toute la chaîne de wrapping :

```go
rapport, err := construireRapport(fichiers, sortie, ext)
if err != nil {
	switch {
	case errors.Is(err, inventory.ErrAucunResultat):
		// Pas grave : simple avertissement, code de sortie 0
	case errors.Is(err, inventory.ErrExtensionInvalide):
		// Faute de frappe utilisateur : usage + os.Exit(1)
	default:
		// Inattendu : message brut + os.Exit(1)
	}
}
```

**Checkpoints ✅**
- [ ] `go run . texte .xyz` affiche un avertissement « aucun fichier » et sort avec le code 0 (`echo $?`)
- [ ] `go run . texte xyz` (sans point) affiche l'usage et sort avec le code 1
- [ ] Le message d'erreur complet montre la chaîne de wrapping, ex : `construction du rapport : filtre extension "xyz" : extension invalide (...)`
- [ ] Vous savez expliquer pourquoi `err == inventory.ErrAucunResultat` (comparaison directe) échouerait ici, alors qu'`errors.Is` fonctionne

---

## Étape 4 — Interface `Formatter` : texte & JSON (1h)

L'affichage devient un vrai sous-système interchangeable : c'est l'étape « interfaces ».

### 4.1 — L'interface (`format/format.go`)

```go
package format

import "gopack/inventory"

// Formatter transforme un inventaire en chaîne prête à afficher.
type Formatter interface {
	Format(fichiers []inventory.File) (string, error)
}
```

Une interface **petite** (1 méthode), définie **côté consommateur**. Rappel du cours : l'implémentation est **implicite** — aucune struct ne déclarera `implements Formatter`, il suffit d'avoir la méthode.

### 4.2 — Implémentation texte (`format/text.go`)

```go
type TextFormatter struct {
	Colored bool // couleurs activées ou non (utilisé à l'étape 5)
}

func (t TextFormatter) Format(fichiers []inventory.File) (string, error)
```

Reprenez votre affichage en colonnes du TP2, mais **construisez une string** au lieu d'imprimer : utilisez `strings.Builder` (`b.WriteString(...)` / `fmt.Fprintf(&b, ...)`, puis `b.String()`). Un inventaire vide retourne l'erreur sentinelle `format.ErrRienAFormater`.

### 4.3 — Implémentation JSON (`format/json.go`)

```go
type JSONFormatter struct{}

func (j JSONFormatter) Format(fichiers []inventory.File) (string, error)
```

1. Utilisez `json.MarshalIndent(fichiers, "", "  ")` (package `encoding/json`)
2. En cas d'erreur de marshalling, wrappez-la : `fmt.Errorf("encodage JSON : %w", err)`
3. Ajoutez des **tags JSON** sur la struct `File` pour des clés propres en minuscules :
   ```go
   type File struct {
       Name string `json:"name"`
       Size int64  `json:"size"`
       // ...
       Tag  string `json:"tag,omitempty"` // omitempty : absent si vide
   }
   ```
   ❓ Question au passage : que se passerait-il si un champ de `File` était en minuscule ? *(indice : `encoding/json` est un package externe à `inventory`… la casse, toujours la casse)*

### 4.4 — Sélection au lancement

Dans `main.go`, la sortie se choisit en argument : `go run . [texte|json] [extension]`

```go
var fmtr format.Formatter // variable typée par l'INTERFACE
switch sortie {
case "texte":
	fmtr = format.TextFormatter{Colored: true}
case "json":
	fmtr = format.JSONFormatter{}
default:
	// erreur + usage
}
rapport, err := fmtr.Format(fichiers) // main ignore l'implémentation concrète
```

**Checkpoints ✅**
- [ ] `go run .` et `go run . texte` affichent le tableau texte
- [ ] `go run . json` sort du JSON indenté valide (testez `go run . json | python3 -m json.tool` ou collez-le dans un validateur)
- [ ] `go run . json .pdf` ne sort que les 2 PDF
- [ ] Les fichiers non marqués n'ont **pas** de clé `"tag"` dans le JSON (`omitempty`)
- [ ] Vous savez répondre : où est écrit que `TextFormatter` implémente `Formatter` ? *(nulle part — et c'est voulu)*

---

## Étape 5 — Package externe : `fatih/color` (30 min)

Premier `go get` du cours : de la couleur dans la sortie texte.

### 5.1 — Installation

```bash
go get github.com/fatih/color
cat go.mod   # une section require est apparue (+ dépendances indirectes)
cat go.sum   # les empreintes cryptographiques de chaque dépendance
```

`go.sum` garantit que tout le monde télécharge exactement les mêmes octets : il se **committe** avec `go.mod`.

### 5.2 — Utilisation

1. Dans `format/text.go`, colorisez quand `Colored` vaut `true` :
   ```go
   enTete := color.New(color.FgCyan, color.Bold).Sprint(...)
   tag    := color.New(color.FgYellow).Sprint(...)
   ```
   Les fonctions `Sprint*` retournent la string colorisée (séquences ANSI) — parfait puisque `Format` construit une string.
2. Dans `main.go`, les messages d'erreur passent en rouge (`color.New(color.FgRed)` sur stderr) et l'avertissement « aucun résultat » en jaune.
3. La couleur doit rester **désactivable** : `TextFormatter{Colored: false}` sort du texte brut. *(La lib se désactive d'ailleurs seule si la sortie n'est pas un terminal.)*

**Checkpoints ✅**
- [ ] `go run . texte` affiche l'en-tête en cyan et les tags en jaune
- [ ] `go run . json` reste **sans** couleur (le JSON doit rester parsable !)
- [ ] `go.mod` et `go.sum` référencent `github.com/fatih/color`
- [ ] `go build -o gopack .` produit un binaire — la dépendance est embarquée dedans, rien à installer sur la machine cible

---

## Livrable J3

À la fin du TP vous devez avoir un module `gopack` avec :

- ✅ `go.mod` (module `gopack`) + `go.sum`
- ✅ Package `inventory` : domaine complet, erreurs sentinelles exportées, helpers non exportés
- ✅ Erreurs wrappées avec `%w` et discriminées dans `main` avec `errors.Is` (codes de sortie corrects)
- ✅ Interface `Formatter` + `TextFormatter` + `JSONFormatter`, sélection par argument
- ✅ Sortie texte colorée via `fatih/color`, désactivable
- ✅ `go vet ./...` silencieux et `go build -o gopack .` fonctionnel

**Démo de fin :** être capable de rejouer en live :
- `go run . json .pdf` puis `go run . texte .xyz` avec explication des codes de sortie
- Montrer la chaîne de wrapping d'une erreur et le `errors.Is` qui la rattrape

## Pour aller plus loin

- Implémentez `fmt.Stringer` sur `inventory.Stats` (`func (s Stats) String() string`) et constatez que `fmt.Println(stats)` l'utilise automatiquement — encore une interface implicite
- Ajoutez un `CSVFormatter` (package `encoding/csv`) : combien de lignes avez-vous touchées dans `main.go` ? C'est la promesse des interfaces
- Réorganisez selon le layout standard : `cmd/gopack/main.go` + `internal/inventory/` — que change `internal/` pour ceux qui importeraient votre module ?
- Regardez `errors.As` : dans quel cas préférerait-on créer un **type** d'erreur custom plutôt qu'une sentinelle ?
