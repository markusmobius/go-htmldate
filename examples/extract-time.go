//+build ignore

package main

import (
	"fmt"
	"net/http"

	"github.com/markusmobius/go-htmldate"
)

func main() {
	// Download URL
	url := "https://edition.cnn.com/2021/07/13/politics/donald-trump-books-last-days-2020/index.html"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Prepare configuration
	opts := htmldate.Options{
		ExtractTime:     true,
		UseOriginalDate: false,
		EnableLog:       true,
	}

	// Extract date
	dt, err := htmldate.FromReader(resp.Body, opts)
	if err != nil {
		panic(err)
	}

	// Print result if date found
	if dt.IsZero() {
		fmt.Println("date not found")
		return
	}

	fmt.Println("Date:", dt.Format("2006-01-02"))
	fmt.Println("Time:", dt.Format("15:04:05 MST"))

	// The output should be:
	// Date: 2021-07-13
	// Time: 19:25:31 UTC
}
