package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"time"

	"github.com/markusmobius/go-htmldate"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		nDocument   int
		evNothing   evaluationResult
		evFast      evaluationResult
		evExtensive evaluationResult
	)

	for _, entry := range comparisonData {
		// Open file
		fContent, err := openFile(entry.File)
		if err != nil {
			logrus.Errorf("failed to open %s: %v", entry.File, err)
			continue
		}

		// Null hypotheses
		ev := evaluateResult("", entry)
		evNothing = mergeEvaluationResult(evNothing, ev)

		// Fast htmldate
		start := time.Now()
		result, err := runHtmlDate(entry.URL, fContent, false)
		if err != nil {
			logrus.Errorf("fast htmldate error in %s: %v", entry.URL, err)
		}

		// if result != entry.Fast && result != entry.Date {
		// 	logrus.Warnf("fast got different result in %s: %s vs %s, want %s", entry.URL, entry.Fast, result, entry.Date)
		// }

		duration := time.Now().Sub(start)
		ev = evaluateResult(result, entry)
		evFast = mergeEvaluationResult(evFast, ev)
		evFast.Duration += duration

		// Extensive htmldate
		start = time.Now()
		result, err = runHtmlDate(entry.URL, fContent, true)
		if err != nil {
			logrus.Errorf("extensive htmldate error in %s: %v", entry.URL, err)
		}

		if result != entry.Extensive && result != entry.Date {
			logrus.Warnf("extensive got different result in %s: %s vs %s, want %s", entry.URL, entry.Extensive, result, entry.Date)
		}

		duration = time.Now().Sub(start)
		ev = evaluateResult(result, entry)
		evExtensive = mergeEvaluationResult(evExtensive, ev)
		evExtensive.Duration += duration

		// Counter
		nDocument++
	}

	fmt.Printf("N Documents: %d\n\n", nDocument)
	fmt.Printf("Nothing: %s\n\n", evNothing.info())

	fmt.Printf("Fast: %s\n", evFast.info())
	fmt.Printf("\t%s\n\n", evFast.scoreInfo())

	fmt.Printf("Extensive: %s\n", evExtensive.info())
	fmt.Printf("\t%s\n\n", evExtensive.scoreInfo())
}

func openFile(name string) ([]byte, error) {
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

	// Read file content
	bt, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if len(bt) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	return bt, nil
}

func runHtmlDate(url string, bt []byte, extensive bool) (string, error) {
	r := bytes.NewReader(bt)
	opts := htmldate.Options{
		UseOriginalDate:     true,
		SkipExtensiveSearch: !extensive,
	}

	dt, err := htmldate.FromReader(r, opts)
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
