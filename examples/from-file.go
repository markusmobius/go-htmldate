//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	"github.com/markusmobius/go-htmldate"
)

func main() {
	// Open the file
	f, err := os.Open("test-files/comparison/talent.ch.5031.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Prepare configuration
	// Request the publication date rather than the last-modified date.
	opts := htmldate.Options{
		UseOriginalDate:     true,
		EnableLog:           true,
		SkipExtensiveSearch: false,
	}

	// Extract date
	res, err := htmldate.FromReader(f, opts)
	if err != nil {
		panic(err)
	}

	// Print the result if a date was found.
	if !res.IsZero() {
		fmt.Println(res.Format("2006-01-02"))
	}
}
