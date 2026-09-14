package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-shiori/dom"
	dps "github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-htmldate"
	"golang.org/x/net/html"
)

type benchmarkFixture struct {
	File   string `json:"file"`
	Bytes  int    `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type benchmarkPage struct {
	document *html.Node
}

type benchmarkOutcome struct {
	File  string `json:"file"`
	Date  string `json:"date"`
	Error string `json:"error"`
}

type benchmarkAttempt struct {
	date time.Time
	err  error
}

type benchmarkMetadata struct {
	Cohort        string `json:"cohort"`
	Cases         int    `json:"cases"`
	CorpusBytes   int    `json:"corpus_bytes"`
	CorpusSHA256  string `json:"corpus_sha256"`
	ResultsSHA256 string `json:"results_sha256"`
	Parsed        int    `json:"parsed"`
	CurrentTime   string `json:"current_time"`
	MinDate       string `json:"min_date"`
	MaxDate       string `json:"max_date"`
	GoVersion     string `json:"go_version"`
	Timezone      string `json:"timezone"`
}

func runBenchmark(cohort string, passes int, root, currentTime string, options htmldate.Options) error {
	if cohort != "document" || passes != 1 || runtime.GOMAXPROCS(0) != 1 || os.Getenv("TZ") != "UTC" {
		return fmt.Errorf("benchmark requires -benchmark document -passes 1, GOMAXPROCS=1 and TZ=UTC")
	}
	if options.MinDate.IsZero() || options.MaxDate.IsZero() || options.MinDate.After(options.MaxDate) {
		return fmt.Errorf("benchmark requires ordered explicit date bounds")
	}
	reference, err := time.Parse(time.RFC3339, currentTime)
	if err != nil {
		return err
	}
	options.DateParserConfig = &dps.Configuration{
		CurrentTime: reference, StrictParsing: true, PreferredDateSource: dps.Past,
	}
	entries := append(append([]comparisonEntry{}, mediacloudData...), defaultComparisonData...)
	pages := make([]benchmarkPage, len(entries))
	fixtures := make([]benchmarkFixture, len(entries))
	corpusBytes := 0
	for index, entry := range entries {
		var content []byte
		var filename string
		for _, directory := range []string{"mediacloud", "comparison", "mock"} {
			filename = filepath.Join("test-files", directory, entry.File)
			content, err = os.ReadFile(filepath.Join(root, filename))
			if err == nil {
				break
			}
		}
		if err != nil {
			return fmt.Errorf("load %s: %w", entry.File, err)
		}
		content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
		fixtures[index] = benchmarkFixture{
			File: filepath.ToSlash(filename), Bytes: len(content), SHA256: fmt.Sprintf("%x", sha256.Sum256(content)),
		}
		corpusBytes += len(content)
		pages[index].document, err = dom.Parse(bytes.NewReader(content))
		if err != nil {
			return fmt.Errorf("parse %s: %w", filename, err)
		}
	}
	attempts := make([]benchmarkAttempt, len(pages))
	execute := func() {
		for index, page := range pages {
			result, err := htmldate.FromDocument(page.document, options)
			attempts[index] = benchmarkAttempt{result.DateTime, err}
			runtime.KeepAlive(result)
		}
	}
	outcomes := func() ([]benchmarkOutcome, string, int) {
		results := make([]benchmarkOutcome, len(attempts))
		parsed := 0
		for index, attempt := range attempts {
			results[index].File = fixtures[index].File
			if !attempt.date.IsZero() {
				results[index].Date = attempt.date.Format("2006-01-02")
				parsed++
			}
			if attempt.err != nil {
				results[index].Error = attempt.err.Error()
			}
		}
		encoded, _ := json.Marshal(results)
		return results, fmt.Sprintf("%x", sha256.Sum256(encoded)), parsed
	}
	encodedFixtures, err := json.Marshal(fixtures)
	if err != nil {
		return err
	}
	metadata := benchmarkMetadata{
		Cases: len(pages), CorpusBytes: corpusBytes,
		CorpusSHA256: fmt.Sprintf("%x", sha256.Sum256(encodedFixtures)),
		CurrentTime:  reference.Format(time.RFC3339), MinDate: options.MinDate.Format(time.RFC3339Nano),
		MaxDate: options.MaxDate.Format(time.RFC3339Nano), GoVersion: runtime.Version(), Timezone: os.Getenv("TZ"),
	}
	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(struct {
		Metadata benchmarkMetadata  `json:"metadata"`
		Fixtures []benchmarkFixture `json:"fixtures"`
	}{metadata, fixtures}); err != nil {
		return err
	}
	decoder := json.NewDecoder(os.Stdin)
	expected := make(map[string]string)
	for {
		var request struct {
			Cohort string `json:"cohort"`
			Warmup bool   `json:"warmup"`
		}
		if err := decoder.Decode(&request); err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		parts := strings.Split(request.Cohort, "-")
		if len(parts) != 3 || parts[0] != "document" ||
			parts[1] != "original" && parts[1] != "modified" || parts[2] != "fast" && parts[2] != "extensive" {
			return fmt.Errorf("invalid benchmark cohort %q", request.Cohort)
		}
		if !request.Warmup && expected[request.Cohort] == "" {
			return fmt.Errorf("warm up %s before timing", request.Cohort)
		}
		options.UseOriginalDate = parts[1] == "original"
		options.SkipExtensiveSearch = parts[2] == "fast"
		var elapsed float64
		var allocatedBytes, allocations uint64
		if request.Warmup {
			execute()
		} else {
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			started := time.Now()
			execute()
			elapsed = float64(time.Since(started)) / float64(time.Millisecond)
			runtime.ReadMemStats(&after)
			allocatedBytes = after.TotalAlloc - before.TotalAlloc
			allocations = after.Mallocs - before.Mallocs
		}
		results, actualHash, parsed := outcomes()
		for _, result := range results {
			if result.Error != "" {
				return fmt.Errorf("extract %s: %s", result.File, result.Error)
			}
		}
		if request.Warmup {
			expected[request.Cohort] = actualHash
		} else if actualHash != expected[request.Cohort] {
			return fmt.Errorf("%s outputs changed during timed pass", request.Cohort)
		} else {
			results = nil
		}
		metadata.Cohort, metadata.ResultsSHA256, metadata.Parsed = request.Cohort, actualHash, parsed
		runtime.KeepAlive(pages)
		if err := encoder.Encode(struct {
			Metadata       benchmarkMetadata  `json:"metadata"`
			Warmup         bool               `json:"warmup"`
			PassMS         float64            `json:"pass_ms"`
			AllocatedBytes uint64             `json:"allocated_bytes"`
			Allocations    uint64             `json:"allocations"`
			Results        []benchmarkOutcome `json:"results,omitempty"`
		}{metadata, request.Warmup, elapsed, allocatedBytes, allocations, results}); err != nil {
			return err
		}
	}
}
