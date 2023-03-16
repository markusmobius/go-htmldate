// This file is part of go-htmldate, Go package for extracting publication dates from a web page.
// Source available in <https://github.com/markusmobius/go-htmldate>.
// Copyright (C) 2022 Markus Mobius
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
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	htmlxpath "github.com/antchfx/htmlquery"
	"github.com/go-shiori/dom"
	"github.com/rs/zerolog"
	"golang.org/x/net/html"
)

var log zerolog.Logger

func init() {
	log = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04",
	}).With().Timestamp().Logger().Level(zerolog.Disabled)
}

// FromReader extract publish date from the specified reader.
func FromReader(r io.Reader, opts Options) (Result, error) {
	// Parse html document
	doc, err := dom.Parse(r)
	if err != nil {
		return resultZero, err
	}

	return FromDocument(doc, opts)
}

// FromDocument extract publish date from the specified html document.
func FromDocument(doc *html.Node, opts Options) (Result, error) {
	// Make sure document exist
	if doc == nil {
		return resultZero, fmt.Errorf("document is empty")
	}

	// Clone document so the original kept untouched
	doc = dom.Clone(doc, true)

	// Set default options
	if opts.MinDate.IsZero() {
		opts.MinDate = defaultMinDate
	}

	if opts.MaxDate.IsZero() {
		opts.MaxDate = defaultMaxDate
	}

	// If URL is not defined in options, look in elements
	if opts.URL == "" {
		links := dom.QuerySelectorAll(doc, `link[rel="canonical"]`)

		for _, elem := range links {
			attrName := "href"
			if dom.TagName(elem) == "meta" {
				attrName = "content"
			}

			href := dom.GetAttribute(elem, attrName)
			href = strings.TrimSpace(href)
			if href != "" {
				opts.URL = href
				break
			}
		}
	}

	// Prepare logger
	if opts.EnableLog {
		log = log.Level(zerolog.DebugLevel)
	}

	// Extract date
	rawString, date, err := findDate(doc, opts)
	if err != nil {
		return resultZero, err
	}

	// Extract time if required
	var timeFound bool
	var timezoneFound bool

	if opts.ExtractTime {
		h, m, s, tz, found := findTime(rawString)
		if found {
			timeFound = true
			date = date.Add(time.Hour * time.Duration(h))
			date = date.Add(time.Minute * time.Duration(m))
			date = date.Add(time.Second * time.Duration(s))
		}

		if tz != nil {
			timezoneFound = true
			date = time.Date(date.Year(), date.Month(), date.Day(),
				date.Hour(), date.Minute(), date.Second(), 0, tz)
		}
	}

	return Result{
		DateTime:    date,
		HasTime:     timeFound,
		HasTimezone: timezoneFound,
		SrcString:   normalizeSpaces(rawString),
	}, nil
}

// findDate extract publish date from the specified html document.
func findDate(doc *html.Node, opts Options) (string, time.Time, error) {
	// Check URL
	if opts.URL != "" {
		urlDate := extractUrlDate(opts.URL, opts)
		if !urlDate.IsZero() {
			return opts.URL, urlDate, nil
		}
	}

	// Try from head elements
	rawString, metaResult := examineMetaElements(doc, opts)
	if !metaResult.IsZero() {
		return rawString, metaResult, nil
	}

	// Try to use JSON data
	rawString, jsonResult := jsonSearch(doc, opts)
	if !jsonResult.IsZero() {
		return rawString, jsonResult, nil
	}

	// Try <abbr> elements
	rawString, abbrResult := examineAbbrElements(doc, opts)
	if !abbrResult.IsZero() {
		return rawString, abbrResult, nil
	}

	// Use selectors + text content
	var finalDateXpath string
	if !opts.SkipExtensiveSearch {
		finalDateXpath = slowPrependXpath + dateXpath
	} else {
		finalDateXpath = fastPrependXpath + dateXpath
	}

	// First try in pruned document
	prunedDoc := dom.Clone(doc, true)
	discardUnwanted(prunedDoc)
	dateElements := htmlxpath.Find(prunedDoc, finalDateXpath)
	rawString, dateResult := examineOtherElements(dateElements, opts)
	if !dateResult.IsZero() {
		return rawString, dateResult, nil
	}

	// TODO: for now, we'll stop searching in discarded elements
	// Search in the discarded elements (currently: footers and archive.org banner)
	// for _, subTree := range discarded {
	// 	dateElements := htmlxpath.Find(subTree, dateXpathQuery)
	// 	rawString, dateResult := examineOtherElements(dateElements, opts)
	// 	if !dateResult.IsZero() {
	// 		return rawString, dateResult, nil
	// 	}
	// }

	// Try <time> elements
	rawString, timeResult := examineTimeElements(prunedDoc, opts)
	if !timeResult.IsZero() {
		return rawString, timeResult, nil
	}

	// Try string search
	cleanedDoc := cleanDocument(doc)

	var htmlString string
	htmlNode := dom.QuerySelector(cleanedDoc, "html")
	if htmlNode != nil {
		htmlString = dom.InnerHTML(htmlNode)
	} else {
		htmlString = dom.InnerHTML(cleanedDoc)
	}

	// String search using regex timestamp
	rawString, timestampResult := timestampSearch(htmlString, opts)
	if !timestampResult.IsZero() {
		return rawString, timestampResult, nil
	}

	// Precise patterns and idiosyncrasies
	rawString, textResult := idiosyncrasiesSearch(htmlString, opts)
	if !textResult.IsZero() {
		return rawString, textResult, nil
	}

	// Try title elements
	for _, titleElem := range dom.GetAllNodesWithTag(doc, "title", "h1") {
		textContent := normalizeSpaces(dom.TextContent(titleElem))
		_, attempt := tryYmdDate(textContent, opts)
		if !attempt.IsZero() {
			log.Debug().Msgf("found date in title: %s", textContent)
			return textContent, attempt, nil
		}
	}

	// Try partial URL
	if opts.URL != "" {
		urlDate := extractPartialUrlDate(opts.URL, opts)
		if !urlDate.IsZero() {
			return opts.URL, urlDate, nil
		}
	}

	// Try URL from image metadata
	rawString, imgResult := metaImgSearch(doc, opts)
	if !imgResult.IsZero() {
		return rawString, imgResult, nil
	}

	// Last resort: do extensive search.
	if !opts.SkipExtensiveSearch {
		log.Debug().Msg("extensive search started")

		// TODO: further tests & decide according to original_date
		var refValue int64
		var refString string
		for _, segment := range htmlxpath.Find(doc, freeTextXpath) {
			text := normalizeSpaces(segment.Data)
			if nText := len(text); nText > 6 && nText < 60 {
				refString, refValue = compareReference(refString, refValue, text, opts)
			}
		}

		// Return
		converted := checkExtractedReference(refValue, opts)
		if !converted.IsZero() {
			return refString, converted, nil
		}

		// Search page HTML
		rawString, searchResult := searchPage(htmlString, opts)
		if !searchResult.IsZero() {
			return rawString, searchResult, nil
		}
	}

	return "", timeZero, nil
}

func findTime(rawString string) (hour, minute, second int, timezone *time.Location, timeFound bool) {
	// If raw string is empty, return early
	rawString = normalizeSpaces(rawString)
	if rawString == "" {
		return
	}

	// Try ISO-8601 time format.
	// While looking for ISO-8601, remove the matches so the later regex not confused.
	rawString = rxIsoTime.ReplaceAllStringFunc(rawString, func(match string) string {
		if !timeFound {
			log.Debug().Msgf("found ISO-8601 time: %s", rawString)

			parts := rxIsoTime.FindStringSubmatch(match)
			hour, _ = strconv.Atoi(parts[1])
			minute, _ = strconv.Atoi(parts[2])
			second, _ = strconv.Atoi(parts[3])
			timezone = parseTimezoneCode(parts[4])
			timeFound = true
		}

		return " "
	})

	if timeFound && timezone != nil {
		return
	}

	// If timezone not exist in ISO time, looks for the common TZ code (e.g. UTC +07:00)
	// Like before, while looking for timezone code, remove the matches so the later
	// regex not confused.
	if timezone == nil {
		rawString = rxTzCode.ReplaceAllStringFunc(rawString, func(match string) string {
			if timezone == nil {
				timezone = parseTimezoneCode(match)
			}
			return " "
		})
	}

	if timeFound && timezone != nil {
		return
	}

	// If timezone still not found, try to use the named timezone
	if timezone == nil {
		timezone = findNamedTimezone(rawString)
	}

	if timeFound && timezone != nil {
		return
	}

	// At this point we have no more cards to play for extracting timezone, so now we
	// switch to capturing time (if it still hasn't found).
	if !timeFound {
		parts := rxCommonTime.FindStringSubmatch(rawString)

		if len(parts) > 0 {
			// Convert string to int
			hour, _ = strconv.Atoi(parts[1])
			minute, _ = strconv.Atoi(parts[2])
			second, _ = strconv.Atoi(parts[3])

			// Convert 12-hour clock to 24-hour
			h12 := strings.ToLower(parts[4])
			h12 = strings.ReplaceAll(h12, ".", "")
			if h12 == "pm" {
				hour += 12
			}

			log.Debug().Msgf("found common format time: %s", rawString)
			timeFound = true
		}
	}

	return
}

// examineMetaElements parse meta elements to find date cues.
func examineMetaElements(doc *html.Node, opts Options) (string, time.Time) {
	var tMeta, tReserve time.Time
	var strMeta, strReserve string

	// Loop through all meta elements
	for _, elem := range dom.QuerySelectorAll(doc, "meta") {
		// Safeguard
		if len(elem.Attr) == 0 {
			continue
		}

		// Fetch attributes
		name := strings.TrimSpace(dom.GetAttribute(elem, "name"))
		content := strings.TrimSpace(dom.GetAttribute(elem, "content"))
		property := strings.TrimSpace(dom.GetAttribute(elem, "property"))
		pubDate := strings.TrimSpace(dom.GetAttribute(elem, "pubdate"))
		itemProp := strings.TrimSpace(dom.GetAttribute(elem, "itemprop"))
		dateTime := strings.TrimSpace(dom.GetAttribute(elem, "datetime"))
		httpEquiv := strings.TrimSpace(dom.GetAttribute(elem, "http-equiv"))
		outerHtml := dom.OuterHTML(elem)

		if property != "" && content != "" { // Property attribute
			attribute := strings.ToLower(property)
			inModifiedProps := inMap(attribute, propertyModified)
			inDateAttributes := inMap(attribute, dateAttributes)

			if (opts.UseOriginalDate && inDateAttributes) ||
				(!opts.UseOriginalDate && (inModifiedProps || inDateAttributes)) {
				log.Debug().Msgf("examining meta property for publish date: %s", outerHtml)
				strMeta, tMeta = tryYmdDate(content, opts)
			}
		} else if name != "" && content != "" { // Name attribute
			name = strings.ToLower(name)
			if name == "og:url" { // url
				strMeta = content
				tMeta = extractUrlDate(content, opts)
			} else if inMap(name, dateAttributes) { // date
				log.Debug().Msgf("examining meta name: %s", outerHtml)
				strMeta, tMeta = tryYmdDate(content, opts)
			} else if strIn(name, modifiedAttrKeys...) { // modified
				log.Debug().Msgf("examining meta name: %s", outerHtml)
				if !opts.UseOriginalDate {
					strMeta, tMeta = tryYmdDate(content, opts)
				} else {
					strReserve, tReserve = tryYmdDate(content, opts)
				}
			}
		} else if strings.ToLower(pubDate) == "pubdate" { // Publish date
			log.Debug().Msgf("examining meta pubdate: %s", outerHtml)
			strMeta, tMeta = tryYmdDate(content, opts)
		} else if itemProp != "" { // Item scope
			attribute := strings.ToLower(itemProp)
			if strIn(attribute, itemPropAttrKeys...) {
				var strAttempt string
				var tAttempt time.Time
				log.Debug().Msgf("examining meta itemprop: %s", outerHtml)

				if dateTime != "" {
					strAttempt, tAttempt = tryYmdDate(dateTime, opts)
				} else if content != "" {
					strAttempt, tAttempt = tryYmdDate(content, opts)
				}

				if !tAttempt.IsZero() {
					if (strIn(attribute, itemPropOriginal...) && opts.UseOriginalDate) ||
						(strIn(attribute, itemPropModified...) && !opts.UseOriginalDate) {
						strMeta, tMeta = strAttempt, tAttempt
					} else {
						// TODO: put on hold, hurts precision
						// strReserve, tReserve = strAttempt, tAttempt
					}
				}
			} else if attribute == "copyrightyear" { // reserve with copyrightyear
				log.Debug().Msgf("examining meta itemprop: %s", outerHtml)
				if content != "" {
					attempt := content + "-01-01"
					tAttempt, err := time.Parse("2006-01-02", attempt)
					if err == nil && validateDate(tAttempt, opts) {
						strReserve, tReserve = content, tAttempt
					}
				}
			}
		} else if httpEquiv != "" && content != "" { // http-equiv, rare http://www.standardista.com/html5/http-equiv-the-meta-attribute-explained/
			attribute := strings.ToLower(httpEquiv)
			if attribute == "date" {
				log.Debug().Msgf("examining meta httpequiv: %s", outerHtml)
				if opts.UseOriginalDate {
					strMeta, tMeta = tryYmdDate(content, opts)
				} else {
					strReserve, tReserve = tryYmdDate(content, opts)
				}
			} else if attribute == "last-modified" {
				log.Debug().Msgf("examining meta httpequiv: %s", outerHtml)
				if !opts.UseOriginalDate {
					strMeta, tMeta = tryYmdDate(content, opts)
				} else {
					strReserve, tReserve = tryYmdDate(content, opts)
				}
			}
		}

		// Exit loop
		if !tMeta.IsZero() {
			return strMeta, tMeta
		}
	}

	// If nothing was found, look for lower granularity (so far: "copyright year")
	log.Debug().Msg("opting for reserve date with less granularity")
	return strReserve, tReserve
}

// examineAbbrElements scans the page for <abbr> elements and check if their content
// contains an eligible date.
func examineAbbrElements(doc *html.Node, opts Options) (string, time.Time) {
	elements := dom.GetElementsByTagName(doc, "abbr")

	// Make sure elements exist and less than `maxPossibleCandidates`
	if nElements := len(elements); nElements == 0 || nElements >= maxPossibleCandidates {
		return "", timeZero
	}

	var refValue int64
	var refString string
	for _, elem := range elements {
		class := strings.TrimSpace(dom.GetAttribute(elem, "class"))
		dataUtime := strings.TrimSpace(dom.GetAttribute(elem, "data-utime"))

		// Handle data-utime (mostly Facebook)
		if dataUtime != "" {
			candidate, err := strconv.ParseInt(dataUtime, 10, 64)
			if err != nil {
				continue
			}
			log.Debug().Msgf("data-utime found: %d", candidate)

			if opts.UseOriginalDate { // Look for original date
				if refValue == 0 || candidate < refValue {
					refValue = candidate
					refString = dataUtime
				}
			} else { // Look for newest (i.e. largest time delta)
				if candidate > refValue {
					refValue = candidate
					refString = dataUtime
				}
			}
		}

		// Handle class
		if class != "" && strIn(class, classAttrKeys...) {
			text := normalizeSpaces(etreeText(elem))
			title := strings.TrimSpace(dom.GetAttribute(elem, "title"))

			// Other attributes
			if title != "" {
				tryText := title
				log.Debug().Msgf("abbr published-title found: %s", tryText)

				if opts.UseOriginalDate {
					_, attempt := tryYmdDate(tryText, opts)
					if !attempt.IsZero() {
						return tryText, attempt
					}
				} else {
					refString, refValue = compareReference(refString, refValue, tryText, opts)
					if refValue > 0 {
						break
					}
				}
			}

			// Dates, not times of the day
			if len(text) > 10 {
				tryText := strings.TrimPrefix(text, "am ")
				log.Debug().Msgf("abbr published found: %s", tryText)
				refString, refValue = compareReference(refString, refValue, tryText, opts)
			}
		}
	}

	// Convert and return
	converted := checkExtractedReference(refValue, opts)
	if !converted.IsZero() {
		return refString, converted
	}

	// Try rescue in abbr content
	abbrElements := dom.GetElementsByTagName(doc, "abbr")
	rawString, dateResult := examineOtherElements(abbrElements, opts)
	if !dateResult.IsZero() {
		return rawString, dateResult
	}

	return "", timeZero
}

// examineTimeElements scans the page for <time> elements and check if their content
// contains an eligible date.
func examineTimeElements(doc *html.Node, opts Options) (string, time.Time) {
	elements := dom.GetElementsByTagName(doc, "time")

	// Make sure elements exist and less than `maxPossibleCandidates`
	if nElements := len(elements); nElements == 0 || nElements >= maxPossibleCandidates {
		return "", timeZero
	}

	// Scan all the tags and look for the newest one
	var refValue int64
	var refString string
	for _, elem := range elements {
		var shortcutFlag bool
		text := normalizeSpaces(etreeText(elem))
		class := strings.TrimSpace(dom.GetAttribute(elem, "class"))
		dateTime := strings.TrimSpace(dom.GetAttribute(elem, "datetime"))
		pubDate := strings.TrimSpace(dom.GetAttribute(elem, "pubdate"))

		if len(dateTime) > 6 { // Go for datetime attribute
			if strings.ToLower(pubDate) == "pubdate" { // Shortcut: time pubdate
				log.Debug().Msgf("time pubdate found: %s", dateTime)
				if opts.UseOriginalDate {
					shortcutFlag = true
				}
			} else if class != "" { // First choice: entry-date + datetime attribute
				if strings.HasPrefix(class, "entry-date") || strings.HasPrefix(class, "entry-time") {
					log.Debug().Msgf("time/datetime found: %s", dateTime)
					if opts.UseOriginalDate {
						shortcutFlag = true
					}
				} else if class == "updated" && !opts.UseOriginalDate {
					log.Debug().Msgf("updated time/datetime found: %s", dateTime)
				}
			} else { // Datetime attribute
				log.Debug().Msgf("time/datetime found: %s", dateTime)
			}

			// Analyze attribute
			if shortcutFlag {
				_, attempt := tryYmdDate(dateTime, opts)
				if !attempt.IsZero() {
					return dateTime, attempt
				}
			} else {
				refString, refValue = compareReference(refString, refValue, dateTime, opts)
				if refValue > 0 {
					break
				}
			}
		} else if len(text) > 6 { // Bare text in element
			log.Debug().Msgf("time/datetime found in text: %s", text)
			refString, refValue = compareReference(refString, refValue, text, opts)
		}
	}

	// Return
	converted := checkExtractedReference(refValue, opts)
	if !converted.IsZero() {
		return refString, converted
	}

	return "", timeZero
}

// examineOtherElements scans the specified elements and check if their content
// contains an eligible date.
func examineOtherElements(elements []*html.Node, opts Options) (string, time.Time) {
	// Make sure elements exist and less than `maxPossibleCandidates`
	if nElements := len(elements); nElements == 0 || nElements >= maxPossibleCandidates {
		return "", timeZero
	}

	// Loop through the elements to analyze
	var attempt time.Time
	for _, elem := range elements {
		// Trim text content
		textContent := normalizeSpaces(dom.TextContent(elem))

		// Simple length heuristics
		if len(textContent) > 6 {
			// Shorten and try the beginning of the string.
			toExamine := strLimit(textContent, 48)
			toExamine = rxLastNonDigits.ReplaceAllString(toExamine, "")

			// Log the examined element
			elemHTML := dom.OuterHTML(elem)
			elemHTML = strLimit(normalizeSpaces(elemHTML), 100)
			elemHTML = strings.TrimSpace(elemHTML)
			log.Debug().Msgf("analyzing HTML: %s (%s)", elemHTML, toExamine)

			// Attempt to extract date
			_, attempt = tryYmdDate(toExamine, opts)
			if !attempt.IsZero() {
				return toExamine, attempt
			}
		}

		// Try link title (Blogspot)
		titleAttr := strings.TrimSpace(dom.GetAttribute(elem, "title"))
		if titleAttr != "" {
			toExamine := strLimit(titleAttr, 48)
			toExamine = rxLastNonDigits.ReplaceAllString(toExamine, "")
			_, attempt = tryYmdDate(toExamine, opts)
			if !attempt.IsZero() {
				return toExamine, attempt
			}
		}
	}

	return "", timeZero
}

// searchPage opportunistically search the HTML text for common text patterns.
func searchPage(htmlString string, opts Options) (string, time.Time) {
	// Copyright symbol
	log.Debug().Msg("looking for copyright/footer information")

	var copYear int
	var copRawString string
	rawString, bestMatch := searchPattern(htmlString, rxCopyrightPattern, rxYearPattern, rxYearPattern, opts)
	if len(bestMatch) > 0 {
		log.Debug().Msgf("copyright detected: %s", bestMatch[0])
		bestMatchVal, err := strconv.Atoi(bestMatch[0])
		if err == nil && bestMatchVal >= opts.MinDate.Year() && bestMatchVal <= opts.MaxDate.Year() {
			log.Debug().Msgf("copyright year/footer pattern found: %d", bestMatchVal)
			copRawString = rawString
			copYear = bestMatchVal
		}
	}

	// 3 components
	log.Debug().Msg("3 components")

	// Target URL characteristics
	rawString, bestMatch = searchPattern(htmlString, rxThreePattern, rxThreeCatch, rxYearPattern, opts)
	result := filterYmdCandidate(bestMatch, rxThreePattern, copYear, opts)
	if !result.IsZero() {
		return rawString, result
	}

	// More loosely structured date
	rawString, bestMatch = searchPattern(htmlString, rxThreeLoosePattern, rxThreeLooseCatch, rxYearPattern, opts)
	result = filterYmdCandidate(bestMatch, rxThreeLoosePattern, copYear, opts)
	if !result.IsZero() {
		return rawString, result
	}

	// Handle YYYY-MM-DD/DD-MM-YYYY, normalize candidates first
	candidates := plausibleYearFilter(htmlString, rxSelectYmdPattern, rxSelectYmdYear, false, opts)
	candidates = normalizeCandidates(candidates, opts)

	rawString, bestMatch = selectCandidate(candidates, rxYmdPattern, rxYmdYear, opts)
	result = filterYmdCandidate(bestMatch, rxSelectYmdPattern, copYear, opts)
	if !result.IsZero() {
		return rawString, result
	}

	// Valid dates string
	rawString, bestMatch = searchPattern(htmlString, rxDateStringsPattern, rxDateStringsCatch, rxYearPattern, opts)
	result = filterYmdCandidate(bestMatch, rxDateStringsPattern, copYear, opts)
	if !result.IsZero() {
		return rawString, result
	}

	// Handle DD?/MM?/YYYY, normalize candidates first
	candidates = plausibleYearFilter(htmlString, rxSlashesPattern, rxSlashesYear, true, opts)
	candidates = normalizeCandidates(candidates, opts)

	rawString, bestMatch = selectCandidate(candidates, rxYmdPattern, rxYmdYear, opts)
	result = filterYmdCandidate(bestMatch, rxSlashesPattern, copYear, opts)
	if !result.IsZero() {
		return rawString, result
	}

	// 2 components
	log.Debug().Msg("switching to 2 components")

	// First option
	rawString, bestMatch = searchPattern(htmlString, rxYyyyMmPattern, rxYyyyMmCatch, rxYearPattern, opts)
	if len(bestMatch) >= 3 {
		str := fmt.Sprintf("%s-%s-1", bestMatch[1], bestMatch[2])
		dt, err := time.Parse("2006-1-2", str)
		if err == nil && validateDate(dt, opts) {
			if copYear == 0 || dt.Year() >= copYear {
				log.Debug().Msgf("date found for pattern \"%s\": %s", rxYyyyMmPattern.String(), str)
				return rawString, dt
			}
		}
	}

	// Second option
	candidates = plausibleYearFilter(htmlString, rxMmYyyyPattern, rxMmYyyyYear, false, opts)

	// Revert DD-MM-YYYY patterns before sorting
	uniquePatterns := []string{}
	mapPatternCount := make(map[string]int)
	mapPatternRawString := make(map[string]string)

	for _, candidate := range candidates {
		parts := rxMyPattern.FindStringSubmatch(candidate.Pattern)
		if len(parts) < 3 {
			continue
		}

		year, _ := strconv.Atoi(parts[2])
		month, _ := strconv.Atoi(parts[1])
		newPattern := fmt.Sprintf("%04d-%02d-01", year, month)

		if _, exist := mapPatternCount[newPattern]; !exist {
			uniquePatterns = append(uniquePatterns, newPattern)
			mapPatternRawString[newPattern] = candidate.RawString
		}

		mapPatternCount[newPattern] += candidate.Count
	}

	candidates = make([]yearCandidate, len(uniquePatterns))
	for i, pattern := range uniquePatterns {
		candidates[i] = yearCandidate{
			Pattern:   pattern,
			Count:     mapPatternCount[pattern],
			RawString: mapPatternRawString[pattern],
		}
	}

	rawString, bestMatch = selectCandidate(candidates, rxYmdPattern, rxYmdYear, opts)
	result = filterYmdCandidate(bestMatch, rxMmYyyyPattern, copYear, opts)
	if !result.IsZero() {
		return rawString, result
	}

	// Catch all
	if copYear != 0 {
		log.Debug().Msg("using copyright year as default")
		return copRawString, time.Date(copYear, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	// 1 component, last try
	log.Debug().Msg("switching to one component")
	rawString, bestMatch = searchPattern(htmlString, rxSimplePattern, rxYearPattern, rxYearPattern, opts)
	if len(bestMatch) >= 2 {
		str := fmt.Sprintf("%s-1-1", bestMatch[1])
		dt, err := time.Parse("2006-1-2", str)
		if err == nil && validateDate(dt, opts) && dt.Year() >= copYear {
			log.Debug().Msgf("date found for pattern \"%s\": %s", rxSimplePattern.String(), str)
			return rawString, dt
		}
	}

	return "", timeZero
}

// compareReference compares candidate to current date reference
// (includes date validation and older/newer test)
func compareReference(refString string, refValue int64, expression string, opts Options) (string, int64) {
	newRefString, attempt := tryExpression(expression, opts)
	if attempt.IsZero() {
		return refString, refValue
	}

	refValue, changed := compareValues(refValue, attempt, opts)
	if changed {
		refString = newRefString
	}

	return refString, refValue
}

// tryExpression checks if the text string could be a valid date expression.
func tryExpression(expression string, opts Options) (string, time.Time) {
	// Trim expression
	expression = normalizeSpaces(expression)
	if expression == "" || getDigitCount(expression) < 4 {
		return "", timeZero
	}

	// Try the beginning of the string
	expression = strLimit(expression, 48)
	return tryYmdDate(expression, opts)
}

// searchPattern runs chained candidate filtering and selection.
func searchPattern(htmlString string, rxPattern, rxCatchPattern, rxYearPattern *regexp.Regexp, opts Options) (string, []string) {
	candidates := plausibleYearFilter(htmlString, rxPattern, rxYearPattern, false, opts)
	return selectCandidate(candidates, rxCatchPattern, rxYearPattern, opts)
}

// selectCandidate selects a candidate among the most frequent matches.
func selectCandidate(candidates []yearCandidate, catchPattern, yearPattern *regexp.Regexp, opts Options) (string, []string) {
	// Make sure candidates exist and less than `maxPossibleCandidates`
	nCandidates := len(candidates)
	if nCandidates == 0 || nCandidates >= maxPossibleCandidates {
		return "", nil
	}

	// If there is only one candidates, check it immediately
	if nCandidates == 1 {
		for _, item := range candidates {
			matches := catchPattern.FindStringSubmatch(item.Pattern)
			if len(matches) > 0 {
				return item.RawString, matches
			}
		}
	}

	// Get 10 most frequent candidates
	sort.SliceStable(candidates, func(a, b int) bool {
		return candidates[a].Count > candidates[b].Count
	})

	if len(candidates) > 10 {
		candidates = candidates[:10]
	}

	log.Debug().Msgf("top ten occurences: %v", candidates)

	// Sort and find probable candidates
	if !opts.UseOriginalDate {
		sort.SliceStable(candidates, func(a, b int) bool {
			return candidates[a].Pattern > candidates[b].Pattern
		})
	} else {
		sort.SliceStable(candidates, func(a, b int) bool {
			return candidates[a].Pattern < candidates[b].Pattern
		})
	}

	candidate1 := candidates[0]
	candidate2 := candidates[1]
	log.Debug().Msgf("best candidate: %v, %v", candidate1, candidate2)

	// Use plausability heuristics
	year1Parts := yearPattern.FindStringSubmatch(candidate1.Pattern)
	year2Parts := yearPattern.FindStringSubmatch(candidate2.Pattern)
	if len(year1Parts) < 2 || len(year2Parts) < 2 {
		return "", nil
	}

	year1, _ := strconv.Atoi(year1Parts[1])
	year2, _ := strconv.Atoi(year2Parts[1])
	_, year1isValid := validateDateParts(year1, 1, 1, opts)
	_, year2isValid := validateDateParts(year2, 1, 1, opts)

	var matches []string
	var rawString string

	// Safety net: plausibility
	if year1isValid && year2isValid {
		if candidate1.Count == candidate2.Count {
			// Same number of occurrences: always take top of the pile?
			rawString = candidate1.RawString
			matches = catchPattern.FindStringSubmatch(candidate1.Pattern)
		} else if year2 != year1 && float64(candidate1.Count)/float64(candidate2.Count) > 0.5 {
			// Safety net: newer date but up to 50% less frequent
			rawString = candidate2.RawString
			matches = catchPattern.FindStringSubmatch(candidate2.Pattern)
		} else {
			// Not newer or hopefully not significant
			rawString = candidate1.RawString
			matches = catchPattern.FindStringSubmatch(candidate1.Pattern)
		}
	} else if !year1isValid && year2isValid {
		rawString = candidate2.RawString
		matches = catchPattern.FindStringSubmatch(candidate2.Pattern)
	} else if year1isValid && !year2isValid {
		rawString = candidate1.RawString
		matches = catchPattern.FindStringSubmatch(candidate1.Pattern)
	} else {
		log.Debug().Msgf("no suitable candidate: %d %d", year1, year2)
	}

	return rawString, matches
}
