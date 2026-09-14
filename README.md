# Go-HtmlDate [![Go Reference][ref-badge]][ref-link]

Go-HtmlDate is a Go package and command-line tool for extracting publication and last-modified dates from web pages. It is based on [`htmldate`][0], a Python package by [Adrien Barbaresi][1].

Its extraction pipeline follows the original Python implementation to make upstream changes easier to port, while preserving a Go API and optional time and timezone extraction.

## Table of Contents

- [Features](#features)
- [Python Compatibility](#python-compatibility)
- [Usage as a Go Package](#usage-as-a-go-package)
- [Usage as a CLI Application](#usage-as-a-cli-application)
- [Python Verification](#python-verification)
- [Performance](#performance)
- [Comparison with Original](#comparison-with-original)
- [Additional Notes](#additional-notes)
- [Acknowledgements](#acknowledgements)
- [License](#license)

## Features

- Extracts publication or last-modified dates from web pages.
- **Experimental:** also extracts the time and timezone when available.

Like the original, Go-HtmlDate has two modes:

- Fast mode searches metadata and selected elements using targeted patterns.
- Extensive mode also searches additional text and date candidates, using an external date parser and disambiguation rules.

By default, Go-HtmlDate uses extensive mode and looks for the most recent date. Set `UseOriginalDate` to `true` to request the publication date, and `SkipExtensiveSearch` to `true` to select fast mode.

## Python Compatibility

Go-HtmlDate v1.10.1 follows the Python `htmldate` implementation at commit [b895282][3], with Python dateparser 1.4.3 as its reference. It uses [Go-DateParser v1.4.7](https://github.com/markusmobius/go-dateparser/releases/tag/v1.4.7) and [Go-Dateutil v2.9.1](https://github.com/markusmobius/go-dateutil/releases/tag/v2.9.1), without local replacements or runtime Python. The shared Dateutil library owns date parsing, Unicode helpers and the separate CPython ISO/timestamp compatibility APIs.

The implementation follows Python's ISO-before-Dateutil shortcut, character-based digit gates, timestamp bounds, JSON/script order, attribute order and HTML repair. All 9,614 independently generated Python cases agree. Another 24 Python-checked HTML regressions cover local Unix references and date bounds in UTC, Eastern and Kolkata contexts. This evidence is not a proof for arbitrary inputs: a different selected date under an equivalent Python context is a bug, not a supported deviation.

An unspecified `MinDate` means local midnight on January 1, 1995. `MaxDate` defaults to the end of the current local calendar day at microsecond precision and is recalculated per extraction. Explicit Go bounds are instants in their supplied locations; validation uses their wall years and inclusive Python-compatible floating timestamps. Naive candidates use the local environment, while aware ISO/Dateutil candidates use their parsed offsets. Returned dates retain wall-calendar fields in UTC. An out-of-bounds full date can still fall back to a valid month-only date, following Python's branch order.

Unix references such as `abbr[data-utime]` are converted to their local calendar date before validation. Text candidates in the same reference selection use local-midnight timestamps, following Python in both directions.

When time extraction is enabled:

- If no time is found, the result contains only the date, with the time set to `00:00:00` and `HasTime` set to `false`.
- If no timezone is found, the result uses `time.UTC` and `HasTimezone` is `false`. This distinguishes an unspecified timezone from an explicitly detected UTC timezone.

Time and timezone extraction have unit tests, but the saved-page comparison does not measure their accuracy.

## Usage as a Go Package

Requires Go 1.26.0 or newer. The preferred toolchain for working in this repository is Go 1.27.1.

To install the package, run this command inside your Go project:

```sh
go get github.com/markusmobius/go-htmldate@latest
```

Then import the package:

```go
import "github.com/markusmobius/go-htmldate"
```

Use `htmldate.FromReader` to extract a date from an HTML reader, or `htmldate.FromDocument` to reuse a parsed HTML document without modifying the caller's DOM. See [examples/from-file.go](examples/from-file.go), [examples/from-url.go](examples/from-url.go), and [examples/extract-time.go](examples/extract-time.go). The URL examples access live pages, so their results may change over time.

## Usage as a CLI Application

To install the CLI, use Go 1.26.0 or newer:

```sh
go install github.com/markusmobius/go-htmldate/cmd/go-htmldate@latest
```

Once installed, you can use it from your terminal:

```text
$ go-htmldate -h
Extract publication or modification dates from an HTML file or URL

Usage:
  go-htmldate [flags] [source]

Flags:
  -f, --format string       set custom date output format (default "2006-01-02")
  -h, --help                help for go-htmldate
      --ori                 extract the publication date instead of the most recent date
      --skip-tls            skip X.509 (TLS) certificate verification
      --time                extract the time as well as the date
  -t, --timeout int         web page download timeout in seconds (default 30)
  -u, --user-agent string   set custom user agent (default "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
  -v, --verbose             enable debug logging
```

The CLI uses extensive mode. Use `--ori` to request the publication date. Output formats use Go's reference-time layouts; for example, `--time --format '2006-01-02T15:04:05Z07:00'` includes the extracted time and timezone. The default format remains date-only even when `--time` is enabled.

## Python Verification

[test-files/python-reference.json](test-files/python-reference.json) contains 992 shortcut, 929 regex, 1,876 expression, 929 URL, 848 synthetic HTML and 4,040 saved-page cases. The original 9,533 expected results are unchanged; 81 additional cases cover ISO weeks, compact and Unicode digits, aware bounds, local timezone names and both repeated-hour folds. Four recorded Python exceptions are checked for safe no-date handling, not identical exception text. Time/timezone extraction remains a Go extension tested separately.

The self-contained [Python exporter](scripts/python-reference/export.py) pins CPython 3.14.6 and [all dependency versions](scripts/python-reference/requirements.txt). It verifies LF-normalized hashes of all eight installed HtmlDate modules against the recorded upstream source, freezes the parser year and reference date, checks original corpus bytes, and executes Python rather than deriving expectations from Go. Local-context cases run in fresh processes with explicit timestamp environments and cleared validation caches.

```sh
GOWORK=off TZ=UTC go test -mod=readonly -count=1 ./...
GOWORK=off go vet -mod=readonly ./...
python -m pip install -r scripts/python-reference/requirements.txt
python scripts/python-reference/export.py --check
```

Normal Go tests require neither Python nor a neighboring checkout. `--check --kind fast --kind try` rechecks only those existing Python slices without rewriting the fixture. Regeneration without `--check` still rejects any changed historical expectation.

Saved pages are replayed after CRLF-to-LF normalization, with separate SHA-256 hashes for those normalized bytes. The original raw hashes remain in the fixture as provenance; normalization avoids Git's platform-dependent checkout conversion without changing corpus files or expected dates.

`DateParserConfig.CurrentTime` can freeze incomplete-date defaults even in fast mode; otherwise the local calendar date is used. The Dateutil parser-year and timezone-name/offset snapshots are initialized with the process environment. Configure `TZ` before process startup; changing it afterward is unsupported. Windows uses its native long standard/daylight names, independently of the naive timestamp location. Comparisons with Python must use equivalent timezone names, rules, database versions and supported timestamp ranges; matching a zone name alone does not establish that context.

The API is native Go, not a drop-in Python signature: bounds and results are typed, the zero time means no date, and callers format results using Go layouts. The optional time extraction is a Go extension. Python's string-bound and input-object APIs are not reproduced. HTML parsing and byte decoding use native libraries, so untested inputs still need differential checks. These API differences do not relax the date-selection algorithm. No internal worker threads are created; callers control concurrency.

## Performance

Go-HtmlDate uses Go's `regexp` package alongside selected matchers generated by [re2go]. The generated Go files in `internal/re2go` are checked in and compiled as ordinary Go code, without cgo or a runtime regex engine for those matchers. The `.re` files are generation inputs, not runtime dependencies. The indirect `github.com/wasilibs/go-re2` dependency is a separate runtime regex library, not the generator for these matchers.

The checked-in matchers were generated with re2go 4.1. To regenerate them, install that version and run `make generate` from the repository root in a POSIX-compatible shell. Ordinary builds and tests do not require the generator. Changes belong in the `.re` sources, followed by regeneration and tests, rather than manual edits to generated Go code.

### Version Comparison

Measured on **2026-09-14**, comparing Go-HtmlDate v1.9.3 (`79c39d1`) with v1.10.1 (`0f04a39`) on the same **1,000 pre-parsed DOMs**. Each cell is the median of **eight timed runs**, with one complete corpus pass per run: **64 timed passes total**. HTML parsing, file loading and network requests are excluded.

| Mode | Old Go v1.9.3, ms/1,000 pages | New Go v1.10.1, ms/1,000 pages | Old / New |
| :--- | ---------------------------: | ----------------------------: | --------: |
| Publication, fast | 733.54 | 451.89 | 1.62x |
| Publication, extensive | 1,275.83 | 933.30 | 1.37x |
| Last modified, fast | 760.18 | 408.96 | 1.86x |
| Last modified, extensive | 1,392.00 | 1,044.31 | 1.33x |

The multiplier is the old median divided by the new median; above 1 means the new version takes less time. Individual timings vary and some old/new ranges overlap, so these are corpus measurements, not guaranteed production speedups. The [raw report](scripts/comparison/benchmark-v1.9.3-v1.10.1.json) retains all 64 samples, ranges, execution order, allocations, output differences and source identities. The complete benchmark, including builds and setup, finished in 116.3 seconds.

Both versions use Go 1.27.1 on an AMD Ryzen AI 7 PRO 350 under Linux/WSL2, with one extraction caller pinned to logical CPU 2, `GOMAXPROCS=1`, `CGO_ENABLED=0`, `GOAMD64=v1`, default garbage collection and no optional build tags. Each revision retains its original module graph without replacements, including Go-DateParser v1.2.4 versus v1.4.7. This measures the combined date-extraction changes, not DateParser's contribution in isolation.

The [runner](scripts/comparison/benchmark.py) keeps one process per version alive. Each process loads and parses the corpus once, then reuses its caller-owned DOMs across all four modes. One untimed warmup per version/mode precedes the measurements, for eight warmups total. Only one version executes a pass at a time; old/new order alternates within adjacent pairs and mode order varies between rounds. Every completed result is saved immediately. There is no additional repetition multiplier or separate preflight process.

The saved HTML is normalized from CRLF to LF before hashing and parsing. Bounds are fixed at `1995-01-01` through the end of `2026-09-13`, with `DateParserConfig.CurrentTime=2026-09-13T12:00:00Z`, UTC, no explicit URL and time extraction disabled. Only `FromDocument` is timed, including its date parsing, DOM operations and allocations; validation and result formatting are outside the timer. Exact outputs remained stable across all passes within each version. Five publication dates and one modified date differ between versions in each search mode; all inputs are retained. These DOM-only outcomes do not exercise `FromReader`'s HTML repair.

To reproduce under Linux/WSL, use Python 3.12+, Git, Go 1.27.1 and a full Git checkout containing both source revisions:

```sh
python3 scripts/comparison/benchmark.py --runs 8 --cpu 2 --output /tmp/htmldate-comparison.json
```

Choose an available logical CPU with `--cpu` and a new output path for each run. Individual operations have a 60-second timeout and measurement requests share a six-minute overall deadline. Failed runs retain their completed samples and are not retried automatically.

## Comparison with Original

Checked on **2026-09-14**, Go-HtmlDate **v1.10.1** and the [Python reference][3] produce identical dates on the 1,000-entry comparison corpus in all four modes: **4,000/4,000 matching outputs**. The corpus contains 725 BBAW entries and 275 from the [Data Culture Group][dcg]; repeated entries are retained.

Publication-date scores use the same corrected upstream labels, including the 42 label corrections since v1.9.3. Remaining `NaN` labels count as nonmatching references, following the upstream evaluation.

| Implementation | Mode | Precision | Recall | Accuracy | F-Score |
| :------------- | :--- | --------: | -----: | -------: | ------: |
| Python reference (`b895282`) | Fast | 0.924 | 0.927 | 0.861 | 0.925 |
| Python reference (`b895282`) | Extensive | 0.908 | 0.993 | 0.903 | 0.949 |
| Go v1.10.1 | Fast | 0.924 | 0.927 | 0.861 | 0.925 |
| Go v1.10.1 | Extensive | 0.908 | 0.993 | 0.903 | 0.949 |

This is a correctness comparison on saved HTML, including HTML repair, not a timing measurement. It uses Go 1.27.1, CPython 3.14.6 and Python dateparser 1.4.3, with UTC, a fixed reference time of `2026-09-13T12:00:00Z`, bounds from `1995-01-01` through `2026-09-13`, no explicit URL and time extraction disabled. Python expectations come from the independently verified [reference fixture](test-files/python-reference.json), with the four corpus files absent from that fixture checked directly against the pinned Python implementation. All current Go outputs were checked again.

### Changes Since v1.9.3

- Updated Go-DateParser from v1.2.4 to v1.4.7 and adopted shared Go-Dateutil v2.9.1 for parsing, Unicode helpers and CPython-compatible ISO/timestamp handling.
- Matched Python's JSON/script precedence, attribute order and HTML repair, resolving the earlier date-selection differences.
- Corrected ISO-week and compact-date handling, Unicode digit gates, incomplete-date defaults, timezone-aware bounds and local Unix-reference dates.
- Removed unnecessary DOM copies while preserving caller-owned documents and surrounding text; the default maximum date now follows the current local day on each call.

The public Go API, CLI and optional time/timezone extraction are retained. The measured extraction speed improvements are shown in the [version comparison](#version-comparison).

## Additional Notes

The accuracy scores above apply to publication dates only. The upstream corpus has no modified-date ground truth, so Go/Python agreement in modified mode does not establish accuracy. Matching Python means reproducing its date choices, including cases where those choices are not the correct publication dates.

Evaluate the library on representative pages for your application, especially when modification dates, times, or timezones matter. Use explicit `MinDate` and `MaxDate` values when reproducible date bounds are required.

## Acknowledgements

This package would not exist without the work of Adrien Barbaresi, the author of the original Python package. He created `htmldate` as part of an effort to build [text databases for research][k-web]. For some web pages, neither the URL nor the server response provides a reliable way to determine when a document was published or modified. For more information:

```bibtex
@article{barbaresi-2020-htmldate,
  title = {{htmldate: A Python package to extract publication dates from web pages}},
  author = "Barbaresi, Adrien",
  journal = "Journal of Open Source Software",
  volume = 5,
  number = 51,
  pages = 2439,
  url = {https://doi.org/10.21105/joss.02439},
  publisher = {The Open Journal},
  year = 2020,
}
```

- Barbaresi, A. ["htmldate: A Python package to extract publication dates from web pages"][paper-1], Journal of Open Source Software, 5(51), 2439, 2020. DOI: 10.21105/joss.02439
- Barbaresi, A. ["Generic Web Content Extraction with Open-Source Software"][paper-2], Proceedings of KONVENS 2019, Kaleidoscope Abstracts, 2019.
- Barbaresi, A. ["Efficient construction of metadata-enhanced web corpora"][paper-3], Proceedings of the [10th Web as Corpus Workshop (WAC-X)][wac-x], 2016.

## License

Like the original, Go-HtmlDate is distributed under the [Apache License 2.0](LICENSE).

[0]: https://github.com/adbar/htmldate
[1]: https://github.com/adbar
[3]: https://github.com/adbar/htmldate/commit/b895282
[dcg]: https://dataculturegroup.org
[ref-badge]: https://pkg.go.dev/badge/github.com/markusmobius/go-htmldate.svg
[ref-link]: https://pkg.go.dev/github.com/markusmobius/go-htmldate
[paper-1]: https://doi.org/10.21105/joss.02439
[paper-2]: https://hal.archives-ouvertes.fr/hal-02447264/document
[paper-3]: https://hal.archives-ouvertes.fr/hal-01371704v2/document
[wac-x]: https://www.sigwac.org.uk/wiki/WAC-X
[k-web]: https://www.dwds.de/d/k-web
[re2go]: https://re2c.org/manual/manual_go.html
