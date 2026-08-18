package main

import (
	"fmt"
	"os"
	"strconv"
)

const justePrix = 42

// classifyIfElse : version 1, chaîne if / else if / else
func classifyIfElse(note float64) string {
	if note < 10 {
		return "Ajourné"
	} else if note < 12 {
		return "Passable"
	} else if note < 14 {
		return "Assez bien"
	} else if note < 16 {
		return "Bien"
	} else {
		return "Très bien"
	}
}

// classifySwitch : version 2, switch sans expression
func classifySwitch(note float64) string {
	var mention string
	switch {
	case note < 10:
		mention = "Ajourné"
	case note < 12:
		mention = "Passable"
	case note < 14:
		mention = "Assez bien"
	case note < 16:
		mention = "Bien"
	default:
		mention = "Très bien"
	}
	return mention
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage : go run classification.go <note> [proposition juste prix]")
		os.Exit(1)
	}

	// if avec instruction courte : note et err ne vivent que dans ce bloc
	if note, err := strconv.ParseFloat(os.Args[1], 64); err != nil {
		fmt.Println("Erreur : la note doit être un nombre")
		os.Exit(1)
	} else if note < 0 || note > 20 {
		fmt.Println("Erreur : la note doit être comprise entre 0 et 20")
		os.Exit(1)
	}

	// on reparse ici pour avoir "note" dans le scope de main (cf. remarque sur les scopes)
	note, _ := strconv.ParseFloat(os.Args[1], 64)

	mentionIfElse := classifyIfElse(note)
	mentionSwitch := classifySwitch(note)

	fmt.Printf("Note : %g/20 — Mention %s\n", note, mentionSwitch)
	if mentionIfElse != mentionSwitch {
		fmt.Println("⚠️ Les deux versions divergent !")
	}

	// Bonus : le juste prix
	if len(os.Args) >= 3 {
		proposition, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Erreur : la proposition doit être un entier")
			os.Exit(1)
		}

		switch {
		case proposition < justePrix:
			fmt.Println("C'est plus !")
		case proposition > justePrix:
			fmt.Println("C'est moins !")
		default:
			fmt.Println("Gagné !")
		}
	}
}
