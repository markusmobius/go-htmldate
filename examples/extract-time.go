//+build ignore

package main

import (
	"fmt"
	"net/http"

	"github.com/markusmobius/go-htmldate"
)

func main() {
	// This has complete date + time + timezone
	url := "https://edition.cnn.com/2021/07/13/politics/donald-trump-books-last-days-2020/index.html"
	result, err := processURL(url)
	checkError(err)

	fmt.Println("Complete date + time + timezone")
	fmt.Printf("Date        : %s\n", result.Format("2006-01-02"))
	fmt.Printf("Has time    : %v\n", result.HasTime)
	fmt.Printf("Time        : %s\n", result.Format("15:04:05"))
	fmt.Printf("Has timezone: %v\n", result.HasTimezone)

	name, offset := result.DateTime.Zone()
	fmt.Printf("Timezone: %s (offset %d seconds)\n", name, offset)
	fmt.Println()

	// This has date + time, no timezone
	url = "https://arstechnica.com/gaming/2021/07/steam-deck-is-valves-switch-like-portable-pc-starting-at-399-this-december/"
	result, err = processURL(url)
	checkError(err)

	fmt.Println("Date + time, no timezone")
	fmt.Printf("Date        : %s\n", result.Format("2006-01-02"))
	fmt.Printf("Has time    : %v\n", result.HasTime)
	fmt.Printf("Time        : %s\n", result.Format("15:04:05"))
	fmt.Printf("Has timezone: %v\n", result.HasTimezone)
	fmt.Println()

	// This has date only
	url = "https://www.steamdeck.com/en/"
	result, err = processURL(url)
	checkError(err)

	fmt.Println("Date only, no time and timezone")
	fmt.Printf("Date        : %s\n", result.Format("2006-01-02"))
	fmt.Printf("Has time    : %v\n", result.HasTime)
	fmt.Printf("Has timezone: %v\n", result.HasTimezone)

	// Output should look like this:
	//
	// Complete date + time + timezone
	// Date        : 2021-07-13
	// Has time    : true
	// Time        : 19:25:31
	// Has timezone: true
	// Timezone: UTC (offset 0 seconds)
	//
	// Date + time, no timezone
	// Date        : 2021-07-15
	// Has time    : true
	// Time        : 05:08:00
	// Has timezone: false
	//
	// Date only, no time and timezone
	// Date        : 2021-01-01
	// Has time    : false
	// Has timezone: false
}

func processURL(url string) (result htmldate.Result, err error) {
	// Download page
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Extract date and time
	opts := htmldate.Options{
		ExtractTime:     true,
		UseOriginalDate: false,
		EnableLog:       false,
	}

	result, err = htmldate.FromReader(resp.Body, opts)
	return
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
