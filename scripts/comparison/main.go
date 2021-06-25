// This file is part of go-htmldate, Go package for extracting publication dates from a web page.
// Source available in <https://github.com/markusmobius/go-trafilatura>.
// Copyright (C) 2021 Markus Mobius
//
// This program is free software: you can redistribute it and/or modify it under the terms of
// the GNU General Public License as published by the Free Software Foundation, either version 3
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with this program.
// If not, see <https://www.gnu.org/licenses/>.

// Code in this file is ported from <https://github.com/adbar/htmldate> which available under
// GNU GPL v3 license.

package main

import (
	"fmt"
	"os"
	fp "path/filepath"
	"time"

	"github.com/markusmobius/go-htmldate"
	"github.com/rs/zerolog"
	"golang.org/x/net/html"
)

var log = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: "2006-01-02 15:04",
}).With().Timestamp().Logger()

func main() {
	var (
		nDocument   int
		evFast      evaluationResult
		evExtensive evaluationResult
	)

	for _, entry := range comparisonData {
		// Open file
		doc, err := openFile(entry.File)
		if err != nil {
			log.Error().Msgf("failed to open %s: %v", entry.File, err)
			continue
		}

		// Fast htmldate
		start := time.Now()
		fastResult, err := runHtmlDate(doc, false)
		if err != nil {
			log.Error().Msgf("fast error in %s: %v", entry.URL, err)
		}

		duration := time.Now().Sub(start)
		ev := evaluateResult(fastResult, entry)
		evFast = mergeEvaluationResult(evFast, ev)
		evFast.Duration += duration

		// Extensive htmldate
		start = time.Now()
		extensiveResult, err := runHtmlDate(doc, true)
		if err != nil {
			log.Error().Msgf("extensive error in %s: %v", entry.URL, err)
		}

		duration = time.Now().Sub(start)
		ev = evaluateResult(extensiveResult, entry)
		evExtensive = mergeEvaluationResult(evExtensive, ev)
		evExtensive.Duration += duration

		// Log the difference with original code
		if fastResult != entry.Fast || extensiveResult != entry.Extensive {
			log.Debug().Msgf("%s: want \"%s\"", entry.URL, entry.Date)

			if fastResult != entry.Fast {
				log.Debug().Msgf("\tfast different with original: \"%s\" vs \"%s\"", entry.Fast, fastResult)
			} else {
				log.Debug().Msgf("\textensive different with original: \"%s\" vs \"%s\"", entry.Extensive, extensiveResult)
			}
		}

		// Counter
		nDocument++
	}

	fmt.Printf("N Documents: %d\n\n", nDocument)

	fmt.Printf("Fast: %s\n", evFast.info())
	fmt.Printf("\t%s\n\n", evFast.scoreInfo())

	fmt.Printf("Extensive: %s\n", evExtensive.info())
	fmt.Printf("\t%s\n\n", evExtensive.scoreInfo())
}

func openFile(name string) (*html.Node, error) {
	// Open file
	var f *os.File
	var err error
	pathList := []string{
		fp.Join("test-files", "comparison", name),
		fp.Join("test-files", "mock", name),
	}

	for _, path := range pathList {
		f, err = os.Open(path)
		if err == nil {
			break
		}
	}

	if f == nil || err != nil {
		return nil, err
	} else {
		defer f.Close()
	}

	// Parse document
	return parseHTMLDocument(f)
}

func runHtmlDate(doc *html.Node, extensive bool) (string, error) {
	opts := htmldate.Options{
		UseOriginalDate:     true,
		SkipExtensiveSearch: !extensive,
	}

	dt, err := htmldate.FromDocument(doc, opts)
	if err != nil {
		return "", err
	}

	var result string
	if !dt.IsZero() {
		result = dt.Format("2006-01-02")
	}

	return result, nil
}

func evaluateResult(result string, entry comparisonEntry) evaluationResult {
	var ev evaluationResult

	if result == "" && entry.Date == "" {
		ev.TrueNegatives++
	}

	switch {
	case result == "" && entry.Date == "":
		ev.TrueNegatives++
	case result == "" && entry.Date != "":
		ev.FalseNegatives++
	case result == entry.Date:
		ev.TruePositives++
	default:
		ev.FalsePositives++
	}

	return ev
}
