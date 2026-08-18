package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	kmVersMiles = 0.621371
	kmVersM     = 1000.0

	celsiusVersFahrenheitFacteur = 9.0 / 5.0
	celsiusVersFahrenheitOffset  = 32.0

	moVersKo     = 1024
	moVersOctets = 1024 * 1024
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage : go run convertisseur.go <valeur>")
		os.Exit(1)
	}

	valeur, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		fmt.Println("Erreur : la valeur doit être un nombre")
		os.Exit(1)
	}

	fmt.Println("=== Convertisseur d'unités ===")
	fmt.Printf("Valeur saisie : %g\n\n", valeur)

	fmt.Printf("-- Distance (%g interprété en km) --\n", valeur)
	fmt.Printf("%g km = %.2f miles\n", valeur, valeur*kmVersMiles)
	fmt.Printf("%g km = %g m\n\n", valeur, valeur*kmVersM)

	fmt.Printf("-- Température (%g interprété en °C) --\n", valeur)
	fahrenheit := valeur*celsiusVersFahrenheitFacteur + celsiusVersFahrenheitOffset
	fmt.Printf("%g °C = %.1f °F\n\n", valeur, fahrenheit)

	fmt.Printf("-- Stockage (%g interprété en Mo) --\n", valeur)
	ko := int(valeur * moVersKo)
	octets := int(valeur * moVersOctets)
	fmt.Printf("%g Mo = %d Ko\n", valeur, ko)
	fmt.Printf("%g Mo = %d octets\n", valeur, octets)
}
