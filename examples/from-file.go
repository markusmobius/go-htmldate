//+build ignore

package main

import (
	"fmt"
	"os"

	"github.com/markusmobius/go-htmldate"
)

func main() {
	// Open the file
	f, err := os.Open("test-files/mock/befifty.montauk.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Prepare configuration
	// Here we want the publish date instead of last modified
	opts := htmldate.Options{
		UseOriginalDate: false,
		EnableLog:       true,
	}

	// Extract date
	res, err := htmldate.FromReader(f, opts)
	if err != nil {
		panic(err)
	}

	// Print result if date found
	if !res.IsZero() {
		fmt.Println(res.Format("2006-01-02"))
	}
}
