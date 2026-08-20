// Serveur HTTP de fichiers — fourni pour le TP4 (ne pas modifier).
//
// Au démarrage, le serveur génère une dizaine de fichiers de tailles
// variées dans un répertoire temporaire, puis expose deux endpoints :
//
//	GET /files          → liste des fichiers disponibles (JSON)
//	GET /files/{name}   → contenu du fichier demandé
//
// Une latence artificielle est ajoutée sur chaque téléchargement pour
// rendre visible le gain apporté par la concurrence côté client.
//
// Lancement : go run ./server        (ou go run ./server -port 9090)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

// FileInfo décrit un fichier téléchargeable, tel qu'exposé par /files.
type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// Latence artificielle appliquée à chaque téléchargement :
// entre delayMin et delayMin+delayJitter.
const (
	delayMin    = 300 * time.Millisecond
	delayJitter = 500 * time.Millisecond
)

// catalogue : noms et tailles (en octets) des fichiers générés au démarrage.
var catalogue = []FileInfo{
	{Name: "notes.txt", Size: 2 << 10},         // 2 Ko
	{Name: "config.json", Size: 16 << 10},      // 16 Ko
	{Name: "export.csv", Size: 64 << 10},       // 64 Ko
	{Name: "journal.log", Size: 128 << 10},     // 128 Ko
	{Name: "rapport.pdf", Size: 256 << 10},     // 256 Ko
	{Name: "photo-01.jpg", Size: 512 << 10},    // 512 Ko
	{Name: "photo-02.jpg", Size: 1 << 20},      // 1 Mo
	{Name: "archive.zip", Size: 2 << 20},       // 2 Mo
	{Name: "sauvegarde.tar", Size: 3 << 20},    // 3 Mo
	{Name: "video-extrait.mp4", Size: 5 << 20}, // 5 Mo
}

func main() {
	port := flag.Int("port", 8080, "port d'écoute du serveur")
	noDelay := flag.Bool("no-delay", false, "désactiver la latence artificielle")
	flag.Parse()

	// Génération des fichiers dans un répertoire temporaire.
	dir, err := os.MkdirTemp("", "tp4-files-")
	if err != nil {
		log.Fatalf("création du répertoire temporaire : %v", err)
	}
	if err := generateFiles(dir); err != nil {
		log.Fatalf("génération des fichiers : %v", err)
	}
	log.Printf("%d fichiers générés dans %s", len(catalogue), dir)

	// Nettoyage du répertoire temporaire quand on arrête le serveur (Ctrl+C).
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		os.RemoveAll(dir)
		log.Println("arrêt du serveur, fichiers temporaires supprimés")
		os.Exit(0)
	}()

	// Routage (syntaxe "méthode /chemin/{param}" — Go 1.22+).
	mux := http.NewServeMux()
	mux.HandleFunc("GET /files", handleList)
	mux.HandleFunc("GET /files/{name}", handleFile(dir, !*noDelay))

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("serveur de fichiers prêt sur http://localhost%s", addr)
	log.Printf("  liste     : http://localhost%s/files", addr)
	log.Printf("  exemple   : http://localhost%s/files/notes.txt", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("serveur : %v", err)
	}
}

// generateFiles crée chaque fichier du catalogue avec un contenu
// pseudo-aléatoire de la taille annoncée.
func generateFiles(dir string) error {
	for _, f := range catalogue {
		content := make([]byte, f.Size)
		rand.Read(content) // pseudo-aléatoire : suffisant pour des données de test
		path := filepath.Join(dir, f.Name)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return fmt.Errorf("écriture de %s : %w", f.Name, err)
		}
	}
	return nil
}

// handleList renvoie le catalogue des fichiers en JSON.
func handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(catalogue); err != nil {
		log.Printf("encodage de la liste : %v", err)
	}
	log.Printf("%s %s → liste (%d fichiers)", r.Method, r.URL.Path, len(catalogue))
}

// handleFile renvoie un handler qui sert le contenu d'un fichier du
// répertoire dir, avec une latence artificielle optionnelle.
func handleFile(dir string, withDelay bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		// Sécurité : on refuse tout nom contenant un chemin ("../../etc/passwd").
		if name != filepath.Base(name) {
			http.Error(w, "nom de fichier invalide", http.StatusBadRequest)
			return
		}

		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			http.Error(w, "fichier inconnu : "+name, http.StatusNotFound)
			log.Printf("%s %s → 404", r.Method, r.URL.Path)
			return
		}

		// Latence artificielle : simule un vrai serveur distant et rend
		// le gain de la concurrence mesurable côté client.
		if withDelay {
			time.Sleep(delayMin + time.Duration(rand.Int63n(int64(delayJitter))))
		}

		log.Printf("%s %s → envoi", r.Method, r.URL.Path)
		http.ServeFile(w, r, path)
	}
}
