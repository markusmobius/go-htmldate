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

package htmldate

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-shiori/dom"
	dps "github.com/markusmobius/go-dateparser"
	cpydatetime "github.com/markusmobius/go-dateutil/v2/compat/datetime"
	dateutilparser "github.com/markusmobius/go-dateutil/v2/parser"
	"github.com/markusmobius/go-htmldate/internal/re2go"
	"github.com/markusmobius/go-htmldate/internal/selector"
	"golang.org/x/net/html"
)

// discardUnwanted removes unwanted sections of an HTML document.
func discardUnwanted(doc *html.Node) {
	for _, elem := range selector.QueryAll(doc, selector.Discard) {
		if elem.Parent != nil {
			elem.Parent.RemoveChild(elem)
		}
	}
}

// extractUrlDate extract the date out of an URL string complying
// with the Y-M-D format.
func extractUrlDate(url string, opts Options) time.Time {
	// Extract date component using regex
	parts := rxCompleteUrl.FindStringSubmatch(url)
	if len(parts) != 4 {
		return timeZero
	}

	// Create date from the extracted parts
	year, _ := strconv.Atoi(parts[1])
	month, _ := strconv.Atoi(parts[2])
	day, _ := strconv.Atoi(parts[3])

	date, valid := validateDateParts(year, month, day, opts)
	if !valid {
		return timeZero
	}

	log.Debug().Msgf("found date in url: %s", parts[0])
	return date
}

// tryDateExpr tries to extract date which contains year, month and day using
// a series of heuristics and rules.
func tryDateExpr(s string, opts Options) (string, time.Time) {
	// Trim
	s = normalizeSpaces(s)
	s = strLimit(s, maxSegmentLen)

	// Formal constraint: 4 to 18 digits
	nDigit := getDigitCount(s)
	if nDigit < 4 || nDigit > 18 {
		return s, timeZero
	}

	// Check if string only contains time/single year or digits and not a date
	if rxDiscardPattern.MatchString(s) {
		return s, timeZero
	}

	// Try to parse date using the faster method
	parseResult := fastParse(s, opts)
	if !parseResult.IsZero() {
		return s, parseResult
	}

	// Use slow but extensive search, using dateparser
	if !opts.SkipExtensiveSearch {
		// Additional filters to prevent computational cost
		if !rxTextDatePattern.MatchString(s) {
			return s, timeZero
		}

		dt := externalDateParser(s, opts)
		if !dt.IsZero() {
			return s, dt
		}
	}

	return s, timeZero
}

// fastParse parse the string into time.Time.
// In the original Python library, this function is named `custom_parse`, but I
// renamed it to `fastParse` because I think it's more suitable to its purpose.
func fastParse(s string, opts Options) time.Time {
	prefix := []rune(strLimit(s, 8))
	if isDigit(string(prefix[:min(4, len(prefix))])) {
		if isDigit(string(prefix[min(4, len(prefix)):])) {
			if digits, decimal := dateutilparser.ASCIIDecimal(string(prefix)); decimal && len(digits) > 6 {
				year, _ := strconv.Atoi(digits[:4])
				month, _ := strconv.Atoi(digits[4:6])
				day, _ := strconv.Atoi(digits[6:])
				if candidate, valid := validateDateParts(year, month, day, opts); valid {
					return candidate
				}
			}
		} else {
			parsed, err := cpydatetime.FromISOFormat(s)
			if err != nil {
				if candidate := dateutilFallback(s, opts); !candidate.IsZero() {
					return candidate
				}
			} else if validateParsedDate(parsed, opts) {
				return time.Date(parsed.Time.Year(), parsed.Time.Month(), parsed.Time.Day(), 0, 0, 0, 0, time.UTC)
			}
		}
	}

	// 2. Try YYYYMMDD with regex
	parts := rxYmdNoSepPattern.FindStringSubmatch(s)
	if len(parts) == 2 {
		text := parts[1]
		year, _ := strconv.Atoi(text[:4])
		month, _ := strconv.Atoi(text[4:6])
		day, _ := strconv.Atoi(text[6:8])

		if dt, valid := validateDateParts(year, month, day, opts); valid {
			log.Debug().Msgf("fast parse found Y-M-D without separator: %s", text)
			return dt
		}
	}

	// 3. Try the very common YMD, Y-M-D, and D-M-Y patterns
	namedParts, lastMatchedName := rxFindNamedStringSubmatch(rxYmdPattern, s)
	if len(namedParts) != 0 {
		year, _ := strconv.Atoi(namedParts["year"])
		month, _ := strconv.Atoi(namedParts["month"])
		day, _ := strconv.Atoi(namedParts["day"])

		if lastMatchedName != "day" { // handle D-M-Y formats
			year = correctYear(year)
			day, month = trySwapValues(day, month)
		}

		// Make sure month is at most 12, because if not then it's not YMD
		dt, valid := validateDateParts(year, month, day, opts)
		if valid {
			log.Debug().Msgf("fast parse found Y-M-D date: %s", s)
			return dt
		}
	}

	// 4. Try the Y-M and M-Y patterns
	namedParts, _ = rxFindNamedStringSubmatch(rxYmPattern, s)
	if len(namedParts) != 0 {
		year, _ := strconv.Atoi(namedParts["year"])
		month, _ := strconv.Atoi(namedParts["month"])

		// Make sure month is at most 12, because if not then it's not D-M-Y
		dt, valid := validateDateParts(year, month, 1, opts)
		if valid {
			log.Debug().Msgf("fast parse found Y-M date: %s", s)
			return dt
		}
	}

	// 5. Try the other regex pattern
	dt := regexParse(s, opts)
	if validateDate(dt, opts) {
		log.Debug().Msgf("fast parse found regex date: %s", dt.Format("2006-01-02"))
		return dt
	}

	log.Error().Msgf("failed to parse \"%s\"", s)
	return timeZero
}

func dateutilFallback(s string, opts Options) time.Time {
	environment := localDateutilEnvironment
	now := time.Now().In(environment.location)
	if opts.DateParserConfig != nil && !opts.DateParserConfig.CurrentTime.IsZero() {
		now = opts.DateParserConfig.CurrentTime
	} else if !externalDpsConfig.CurrentTime.IsZero() {
		now = externalDpsConfig.CurrentTime
	}
	_, offset := now.Zone()
	instant, instantErr := cpydatetime.Timestamp(dateutilparser.Result{
		Time: now, Aware: true, Offset: time.Duration(offset) * time.Second,
	}, environment.location, false)
	first, firstErr := cpydatetime.Timestamp(dateutilparser.Result{Time: now}, environment.location, false)
	fold := false
	if instantErr == nil && firstErr == nil && instant != first {
		second, secondErr := cpydatetime.Timestamp(dateutilparser.Result{Time: now}, environment.location, true)
		fold = secondErr == nil && instant == second
	}
	return dateutilFallbackInEnvironment(s, opts, now, fold, environment)
}

func dateutilFallbackInEnvironment(s string, opts Options, now time.Time, fold bool, environment dateutilEnvironment) time.Time {
	defaultDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	parsed, err := dateutilparser.ParseWithLocalTimezone(s, defaultDate, environment.parserYear, environment.timezone, fold)
	if err != nil || !validateParsedDateInLocation(parsed, opts, environment.location) {
		return timeZero
	}
	return time.Date(parsed.Time.Year(), parsed.Time.Month(), parsed.Time.Day(), 0, 0, 0, 0, time.UTC)
}

// externalDateParser uses go-dateparser package to extensively look for date.
func externalDateParser(s string, opts Options) time.Time {
	var cfg *dps.Configuration
	if opts.DateParserConfig != nil {
		cfg = opts.DateParserConfig
	} else {
		cfg = externalDpsConfig
	}

	dt, _ := externalParser.Parse(cfg, s)
	if !dt.Time.IsZero() {
		formatted := time.Date(dt.Time.Year(), dt.Time.Month(), dt.Time.Day(), 0, 0, 0, 0, time.UTC)
		if validateDate(formatted, opts) {
			return formatted
		}
	}

	return timeZero
}

// jsonSearch looks for JSON time patterns in JSON sections of the document.
func jsonSearch(doc *html.Node, opts Options) (string, time.Time) {
	pattern := rxJSONModified
	if opts.UseOriginalDate {
		pattern = rxJSONPublished
	}

	for _, elem := range dom.QuerySelectorAll(doc, `script[type="application/ld+json"], script[type="application/settings+json"]`) {
		text := etreeText(elem)
		if !strings.Contains(text, `"date`) {
			continue
		}
		indexes := pattern.FindStringSubmatchIndex(text)
		if len(indexes) < 4 {
			continue
		}
		date, err := time.Parse("2006-1-2", text[indexes[2]:indexes[3]])
		if err != nil || !validateDate(date, opts) {
			continue
		}
		end := indexes[3]
		if closing := strings.IndexByte(text[indexes[2]:], '"'); closing >= 0 {
			end = indexes[2] + closing
		}
		return normalizeSpaces(text[indexes[2]:end]), date
	}
	return "", timeZero
}

// idiosyncrasiesSearch looks for author-written dates throughout the web page.
func idiosyncrasiesSearch(htmlString string, opts Options) (string, time.Time) {
	// Extract date parts
	var candidate time.Time
	parts, startIdx := re2go.IdiosyncracyPatternSubmatch(htmlString)
	if len(parts) == 0 {
		return "", timeZero
	}

	// Process parts
	if len(parts[1]) == 4 { // YYYY/MM/DD
		year, _ := strconv.Atoi(parts[1])
		month, _ := strconv.Atoi(parts[2])
		day, _ := strconv.Atoi(parts[3])
		candidate, _ = validateDateParts(year, month, day, opts)
	} else if tmp := len(parts[3]); tmp == 2 || tmp == 4 { // DD/MM/YY or MM/DD/YY
		year, _ := strconv.Atoi(parts[3])
		month, _ := strconv.Atoi(parts[2])
		day, _ := strconv.Atoi(parts[1])

		year = correctYear(year)
		day, month = trySwapValues(day, month)
		candidate, _ = validateDateParts(year, month, day, opts)
	}

	if !validateDate(candidate, opts) {
		return "", timeZero
	}

	// Get raw string
	rawString := strLimit(htmlString[startIdx:], 100)

	// Return candidate
	log.Debug().Msgf("idiosyncratic pattern found: %s", parts[0])
	return rawString, candidate
}

// metaImgSearch looks for url in <meta> image elements.
func metaImgSearch(doc *html.Node, opts Options) (string, time.Time) {
	for _, elem := range dom.QuerySelectorAll(doc, `meta[property="og:image"]`) {
		content := strings.TrimSpace(dom.GetAttribute(elem, "content"))
		if content != "" {
			result := extractUrlDate(content, opts)
			if validateDate(result, opts) {
				return content, result
			}
		}
	}

	return "", timeZero
}

// regexPatternSearch looks for date expressions using a regular expression on a string of text.
func regexPatternSearch(
	text string,
	patternName string,
	dateSubmatchFinder func(string) ([]string, int),
	opts Options,
) (string, time.Time) {
	parts, _ := dateSubmatchFinder(text)
	if len(parts) < 2 {
		return "", timeZero
	}

	dt := fastParse(parts[1], opts)
	if validateDate(dt, opts) {
		log.Debug().Msgf("regex found: %q %q", patternName, parts[0])
		return parts[0], dt
	}

	return "", timeZero
}

// regexParse try full-text parse for date elements using a series of regular
// expressions with particular emphasis on English, French, German and Turkish.
func regexParse(s string, opts Options) time.Time {
	var exist bool
	var year, month, day int

	// Multilingual day-month-year pattern + American English patterns
	strYear, strMonth, strDay, ok := re2go.FindLongTextPattern(s)
	if ok {
		strMonth = strings.ToLower(strMonth)
		month, exist = monthNumber[strMonth]
		if exist {
			year, _ = strconv.Atoi(strYear)
			day, _ = strconv.Atoi(strDay)
		}
	}

	year = correctYear(year)
	day, month = trySwapValues(day, month)
	dt, valid := validateDateParts(year, month, day, opts)
	if valid {
		log.Debug().Msgf("multilingual text found: %s", s)
		return dt
	}

	return timeZero
}

// trySwapValues swap day and month values if it seems feaaible.
func trySwapValues(day, month int) (int, int) {
	if month > 12 && day <= 12 {
		return month, day
	}
	return day, month
}
