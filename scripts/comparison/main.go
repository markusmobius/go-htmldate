// Copyright (C) 2022 Markus Mobius
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code in this file is ported from <https://github.com/adbar/htmldate>
// which available under Apache 2.0 license.

package main

import (
	"fmt"
	"os"
	fp "path/filepath"
	"time"

	"github.com/go-shiori/dom"
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
		nDocument      int
		evFast         evaluationResult
		evExtensive    evaluationResult
		comparisonData []comparisonEntry
	)

	comparisonData = append(comparisonData, mediacloudData...)
	comparisonData = append(comparisonData, defaultComparisonData...)

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

		duration := time.Since(start)
		ev := evaluateResult(fastResult, entry)
		evFast = mergeEvaluationResult(evFast, ev)
		evFast.Duration += duration

		// Extensive htmldate
		start = time.Now()
		extensiveResult, err := runHtmlDate(doc, true)
		if err != nil {
			log.Error().Msgf("extensive error in %s: %v", entry.URL, err)
		}

		duration = time.Since(start)
		ev = evaluateResult(extensiveResult, entry)
		evExtensive = mergeEvaluationResult(evExtensive, ev)
		evExtensive.Duration += duration

		// Log the difference with original code
		if fastResult != entry.Fast || extensiveResult != entry.Extensive {
			log.Debug().Msgf("%s: want \"%s\"", entry.URL, entry.Date)

			if fastResult != entry.Fast {
				log.Debug().Msgf("\tfast different: original %q, our %q", entry.Fast, fastResult)
			}

			if extensiveResult != entry.Extensive {
				log.Debug().Msgf("\textensive different: original %q, our %q", entry.Extensive, extensiveResult)
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
		fp.Join("test-files", "mediacloud", name),
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
	return dom.Parse(f)
}

func runHtmlDate(doc *html.Node, extensive bool) (string, error) {
	opts := htmldate.Options{
		UseOriginalDate:     true,
		SkipExtensiveSearch: !extensive,
	}

	res, err := htmldate.FromDocument(doc, opts)
	if err != nil {
		return "", err
	}

	var output string
	if !res.IsZero() {
		output = res.Format("2006-01-02")
	}

	return output, nil
}

func evaluateResult(result string, entry comparisonEntry) evaluationResult {
	var ev evaluationResult

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
