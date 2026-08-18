package main

import (
	"fmt"
	"os"
	"strconv"
)

func usage() {
	fmt.Println("Usage : calc <nombre> <opérateur> <nombre>")
	fmt.Println("Opérateurs : + - x /")
}

func main() {
	// len(os.Args) est vérifié AVANT d'accéder aux indices
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println("calc v1.0")
		return
	}

	if len(os.Args) != 4 {
		usage()
		os.Exit(1)
	}

	operateur := os.Args[2]

	// Modulo : n'existe que sur les entiers, on le traite à part
	if operateur == "%" {
		a, errA := strconv.Atoi(os.Args[1])
		b, errB := strconv.Atoi(os.Args[3])
		if errA != nil || errB != nil {
			fmt.Println("Erreur : le modulo nécessite deux entiers")
			os.Exit(1)
		}
		if b == 0 {
			fmt.Println("Erreur : modulo par zéro impossible")
			os.Exit(1)
		}
		fmt.Printf("%d %% %d = %d\n", a, b, a%b)
		return
	}

	a, errA := strconv.ParseFloat(os.Args[1], 64)
	b, errB := strconv.ParseFloat(os.Args[3], 64)
	if errA != nil || errB != nil {
		fmt.Println("Erreur : les opérandes doivent être des nombres")
		os.Exit(1)
	}

	var resultat float64

	switch operateur {
	case "+":
		resultat = a + b
	case "-":
		resultat = a - b
	case "x", "*":
		resultat = a * b
	case "/":
		if b == 0 {
			fmt.Println("Erreur : division par zéro impossible")
			os.Exit(1)
		}
		resultat = a / b
	default:
		fmt.Printf("Erreur : opérateur inconnu %q\n", operateur)
		os.Exit(1)
	}

	fmt.Printf("%g %s %g = %g\n", a, operateur, b, resultat)
}
