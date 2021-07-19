//+build ignore

package main

import (
	"fmt"
	"net/http"

	"github.com/markusmobius/go-htmldate"
)

func main() {
	// Download URL
	resp, err := http.Get("https://www.politico.com/states/california/story/2020/04/01/newsom-california-school-year-closure-the-right-thing-to-do-1270260")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Prepare configuration
	// Here we want the publish date instead of last modified
	opts := htmldate.Options{
		UseOriginalDate: false,
		EnableLog:       true,
		ExtractTime:     true,
	}

	// Extract date
	res, err := htmldate.FromReader(resp.Body, opts)
	if err != nil {
		panic(err)
	}

	// Print result if date found
	if !res.IsZero() {
		fmt.Printf("Date        : %s\n", res.Format("2006-01-02"))
		fmt.Printf("Has time    : %v\n", res.HasTime)
		fmt.Printf("Time        : %s\n", res.Format("15:04:05"))
		fmt.Printf("Has timezone: %v\n", res.HasTimezone)

		name, offset := res.DateTime.Zone()
		fmt.Printf("Timezone    : %s (offset %d seconds)\n", name, offset)
	}
}
