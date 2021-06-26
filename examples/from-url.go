//+build ignore

package main

import (
	"fmt"
	"net/http"

	"github.com/markusmobius/go-htmldate"
)

func main() {
	// Download URL
	resp, err := http.Get("https://www.sueddeutsche.de/bayern/wuerzburg-messer-attacke-verletzte-aktuell-eindruecke-1.5334175")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Prepare configuration
	// Here we want the publish date instead of last modified
	opts := htmldate.Options{
		UseOriginalDate: true,
		EnableLog:       true,
	}

	// Extract date
	dt, err := htmldate.FromReader(resp.Body, opts)
	if err != nil {
		panic(err)
	}

	// Print result if date found
	if !dt.IsZero() {
		fmt.Println(dt.Format("2006-01-02"))
	}
}
