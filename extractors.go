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

package htmldate

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-shiori/dom"
	dps "github.com/markusmobius/go-dateparser"
	"golang.org/x/net/html"
)

// discardUnwanted removes unwanted sections of an HTML document and
// return the discarded elements as a list.
func discardUnwanted(doc *html.Node) []*html.Node {
	var discardedElements []*html.Node
	for _, elem := range findElementsWithRule(doc, discardSelectorRule) {
		if elem.Parent != nil {
			elem.Parent.RemoveChild(elem)
			discardedElements = append(discardedElements, elem)
		}
	}

	return discardedElements
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

// extractPartialUrlDate extract the date out of an URL string complying
// with the Y-M format.
func extractPartialUrlDate(url string, opts Options) time.Time {
	// Extract date component using regex
	parts := rxPartialUrl.FindStringSubmatch(url)
	if len(parts) != 3 {
		return timeZero
	}

	// Create date from the extracted parts
	year, _ := strconv.Atoi(parts[1])
	month, _ := strconv.Atoi(parts[2])

	if month > 12 {
		return timeZero
	}

	date, valid := validateDateParts(year, month, 1, opts)
	if !valid {
		return timeZero
	}

	log.Debug().Msgf("found partial date in url: %s", parts[0])
	return date
}

// tryYmdDate tries to extract date which contains year, month and day using
// a series of heuristics and rules.
func tryYmdDate(s string, opts Options) (string, time.Time) {
	// If string less than 6 runes, stop
	if utf8.RuneCountInString(s) < 6 {
		return s, timeZero
	}

	// Count how many digit number in this string
	nDigit := getDigitCount(s)
	if nDigit < 4 || nDigit > 18 {
		return s, timeZero
	}

	// Check if string only contains time/single year or digits and not a date
	if !rxTextDatePattern.MatchString(s) || rxNoTextDatePattern.MatchString(s) {
		return s, timeZero
	}

	// Try to parse date
	parseResult := fastParse(s, opts)
	if !parseResult.IsZero() {
		return s, parseResult
	}

	if !opts.SkipExtensiveSearch {
		// Use dateparser to extensively parse the date
		dt, _ := dps.Parse(externalDpsConfig, s)
		if !dt.IsZero() && validateDate(dt.Time, opts) {
			return s, dt.Time
		}
	}

	return s, timeZero
}

// fastParse parse the string into time.Time.
// In the original Python library, this function is named `custom_parse`, but I
// renamed it to `fastParse` because I think it's more suitable to its purpose.
func fastParse(s string, opts Options) time.Time {
	// 1. Try YYYYMMDD first
	// This also handle '201709011234' which not covered by dateparser
	if len(s) >= 8 && isDigit(s[:8]) {
		year, _ := strconv.Atoi(s[:4])
		month, _ := strconv.Atoi(s[4:6])
		day, _ := strconv.Atoi(s[6:8])

		if dt, valid := validateDateParts(year, month, day, opts); valid {
			log.Debug().Msgf("fast parse found Y-M-D without separator: %s", s[:8])
			return dt
		}
	}

	// 2. Try Y-M-D pattern since it's the one used in ISO-8601
	parts := rxYmdPattern.FindStringSubmatch(s)
	if len(parts) == 4 {
		year, _ := strconv.Atoi(parts[1])
		month, _ := strconv.Atoi(parts[2])
		day, _ := strconv.Atoi(parts[3])

		// Make sure month is at most 12, because if not then it's not YMD
		dt, valid := validateDateParts(year, month, day, opts)
		if valid {
			log.Debug().Msgf("fast parse found Y-M-D date: %s", parts[0])
			return dt
		}
	}

	// 3. Try the D-M-Y pattern since it's the most common date format in the world
	parts = rxDmyPattern.FindStringSubmatch(s)
	if len(parts) == 4 {
		day, _ := strconv.Atoi(parts[1])
		month, _ := strconv.Atoi(parts[2])
		year, _ := strconv.Atoi(parts[3])

		// Append year if necessary
		if year < 100 {
			if year >= 90 {
				year += 1900
			} else {
				year += 2000
			}
		}

		// If month is more than 12, swap it with the day
		if month > 12 && day <= 12 {
			day, month = month, day
		}

		// Make sure month is at most 12, because if not then it's not D-M-Y
		dt, valid := validateDateParts(year, month, day, opts)
		if valid {
			log.Debug().Msgf("fast parse found D-M-Y date: %s", parts[0])
			return dt
		}
	}

	// 4. Try the Y-M pattern
	parts = rxYmPattern.FindStringSubmatch(s)
	if len(parts) == 3 {
		year, _ := strconv.Atoi(parts[1])
		month, _ := strconv.Atoi(parts[2])

		// Make sure month is at most 12, because if not then it's not D-M-Y
		dt, valid := validateDateParts(year, month, 1, opts)
		if valid {
			log.Debug().Msgf("fast parse found Y-M date: %s", parts[0])
			return dt
		}
	}

	// 5. Try the other regex pattern
	dt := regexParse(s, opts)
	if !dt.IsZero() {
		log.Debug().Msgf("fast parse found regex date: %s", dt.Format("2006-01-02"))
		return dt
	}

	log.Error().Msgf("failed to parse \"%s\"", s)
	return timeZero
}

// jsonSearch looks for JSON time patterns in JSON sections of the document.
func jsonSearch(doc *html.Node, opts Options) (string, time.Time) {
	// Prepare function to capture dates recursively
	var findDates func(dst map[string]string, obj map[string]interface{})
	findDates = func(dst map[string]string, obj map[string]interface{}) {
		for key, value := range obj {
			switch v := value.(type) {
			case string:
				v = normalizeSpaces(v)
				lowerKey := strings.ToLower(key)
				if strings.Contains(lowerKey, "date") && v != "" {
					dst[key] = v
					return
				}

			case map[string]interface{}:
				findDates(dst, v)

			case []interface{}:
				for _, item := range v {
					itemObject, isObject := item.(map[string]interface{})
					if isObject {
						findDates(dst, itemObject)
					}
				}
			}
		}
	}

	// Look throughout the HTML tree
	ldJsonScripts := dom.QuerySelectorAll(doc, `script[type="application/ld+json"]`)
	settingsJsonScripts := dom.QuerySelectorAll(doc, `script[type="application/settings+json"]`)
	scriptNodes := append(ldJsonScripts, settingsJsonScripts...)

	jsonDates := make(map[string]string)
	for _, elem := range scriptNodes {
		// Get the json text inside the script
		jsonText := dom.TextContent(elem)
		jsonText = strings.TrimSpace(jsonText)

		// Decode JSON text, assuming it is an object
		data := map[string]interface{}{}
		err := json.Unmarshal([]byte(jsonText), &data)
		if err != nil {
			continue
		}

		// Find all dates recursively
		findDates(jsonDates, data)
	}

	// Prepare keys to look for
	var keys []string
	if opts.UseOriginalDate {
		keys = []string{"datePublished", "dateCreated"}
	} else {
		keys = []string{"dateModified"}
	}

	// Check each key
	for _, key := range keys {
		keyValue, exist := jsonDates[key]
		if !exist {
			continue
		}

		dt := fastParse(keyValue, opts)
		if validateDate(dt, opts) {
			return keyValue, dt
		}
	}

	return "", timeZero
}

// timestampSearch looks for timestamps throughout the html string.
func timestampSearch(htmlString string, opts Options) (string, time.Time) {
	idxs := rxTimestampPattern.FindStringSubmatchIndex(htmlString)
	if len(idxs) == 0 {
		return "", timeZero
	}

	rawString := strLimit(htmlString[idxs[0]:], 100)
	group1 := htmlString[idxs[2]:idxs[3]]
	dt := fastParse(group1, opts)
	return rawString, dt
}

// idiosyncrasiesSearch looks for author-written dates throughout the web page.
func idiosyncrasiesSearch(htmlString string, opts Options) (string, time.Time) {
	// Do it in order of DE-EN-TR
	rawString, result := extractIdiosyncrasy(rxDePattern, htmlString, opts)

	if result.IsZero() {
		rawString, result = extractIdiosyncrasy(rxEnPattern, htmlString, opts)
	}

	if result.IsZero() {
		rawString, result = extractIdiosyncrasy(rxTrPattern, htmlString, opts)
	}

	return rawString, result
}

// metaImgSearch looks for url in <meta> image elements.
func metaImgSearch(doc *html.Node, opts Options) (string, time.Time) {
	for _, elem := range dom.QuerySelectorAll(doc, `meta[property="og:image"]`) {
		content := strings.TrimSpace(dom.GetAttribute(elem, "content"))
		if content != "" {
			result := extractUrlDate(content, opts)
			if !result.IsZero() {
				return content, result
			}
		}
	}

	return "", timeZero
}

// extractIdiosyncrasy looks for a precise pattern throughout the web page.
func extractIdiosyncrasy(rxIdiosyncrasy *regexp.Regexp, htmlString string, opts Options) (string, time.Time) {
	var candidate time.Time
	parts := rxIdiosyncrasy.FindStringSubmatch(htmlString)
	if len(parts) == 0 {
		return "", timeZero
	}

	var groups []int
	if len(parts) >= 4 && parts[3] != "" {
		groups = []int{0, 1, 2, 3}
	} else if len(parts) >= 7 && parts[6] != "" {
		groups = []int{0, 4, 5, 6}
	} else {
		return "", timeZero
	}

	if len(parts[1]) == 4 {
		year, _ := strconv.Atoi(parts[groups[1]])
		month, _ := strconv.Atoi(parts[groups[2]])
		day, _ := strconv.Atoi(parts[groups[3]])
		candidate, _ = validateDateParts(year, month, day, opts)
	} else if tmp := len(parts[groups[3]]); tmp == 2 || tmp == 4 {
		year, _ := strconv.Atoi(parts[groups[3]])
		month, _ := strconv.Atoi(parts[groups[2]])
		day, _ := strconv.Atoi(parts[groups[1]])

		// Switch to MM/DD/YY if necessary
		if month > 12 {
			day, month = month, day
		}

		// Append year if necessary
		if year < 100 {
			if year >= 90 {
				year += 1900
			} else {
				year += 2000
			}
		}

		candidate, _ = validateDateParts(year, month, day, opts)
	}

	if candidate.IsZero() {
		return "", timeZero
	}

	// Get raw string
	idxs := rxIdiosyncrasy.FindStringIndex(htmlString)
	rawString := strLimit(htmlString[idxs[0]:], 100)

	// Return candidate
	log.Debug().Msgf("idiosyncratic pattern found: %s", parts[0])
	return rawString, candidate
}

// regexParse is full-text parse using a series of regular expressions.
func regexParse(s string, opts Options) time.Time {
	dt := regexParseDe(s, opts)
	if dt.IsZero() {
		dt = regexParseMultilingual(s, opts)
	}

	return dt
}

// regexParseDe tries full-text parse for German date elements.
func regexParseDe(s string, opts Options) time.Time {
	parts := rxGermanTextSearch.FindStringSubmatch(s)
	if len(parts) == 0 {
		return timeZero
	}

	year, _ := strconv.Atoi(parts[3])
	day, _ := strconv.Atoi(parts[1])
	month, exist := monthNumber[parts[2]]
	if !exist {
		return timeZero
	}

	dt, valid := validateDateParts(year, month, day, opts)
	if valid {
		log.Debug().Msgf("German text found: %s", s)
		return dt
	}

	return timeZero
}

// regexParseMultilingual tries full-text parse for English date elements.
func regexParseMultilingual(s string, opts Options) time.Time {
	var exist bool
	var parts []string
	var year, month, day int

	// In original code they handle Month-Day-Year here.
	// However, since it already handled by fastParser I don't repeat it here again.

	// MMMM D YYYY pattern
	parts = rxLongMdyPattern.FindStringSubmatch(s)
	if len(parts) >= 4 {
		month, exist = monthNumber[parts[1]]
		if exist {
			year, _ = strconv.Atoi(parts[3])
			day, _ = strconv.Atoi(parts[2])
			goto regex_finish
		}
	}

	// D MMMM YYYY pattern
	parts = rxLongDmyPattern.FindStringSubmatch(s)
	if len(parts) >= 4 {
		month, exist = monthNumber[parts[2]]
		if exist {
			year, _ = strconv.Atoi(parts[3])
			day, _ = strconv.Atoi(parts[1])
			goto regex_finish
		}
	}

regex_finish:
	if year < 100 {
		if year >= 90 {
			year += 1900
		} else {
			year += 2000
		}
	}

	dt, valid := validateDateParts(year, month, day, opts)
	if valid {
		log.Debug().Msgf("multilingual text found: %s", s)
		return dt
	}

	return timeZero
}
