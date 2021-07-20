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
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

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
		metas := dom.QuerySelectorAll(doc, `meta[property="og:url"]`)
		elements := append(links, metas...)

		for _, elem := range elements {
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
	}, nil
}

// findDate extract publish date from the specified html document.
func findDate(doc *html.Node, opts Options) (string, time.Time, error) {
	// If URL is defined, extract baseline date from it
	var urlDate time.Time
	if opts.URL != "" {
		urlDate = extractUrlDate(opts.URL, opts)
	}

	validateResult := func(result time.Time) bool {
		// URL date is the baseline for original date, so if URL date exist
		// and for some reason the result is different with URL date, most
		// likely that result is invalid.
		if opts.UseOriginalDate && !urlDate.IsZero() && !result.Equal(urlDate) {
			return false
		}
		return !result.IsZero()
	}

	// Try from head elements
	rawString, metaResult := examineMetaElements(doc, opts)
	metaResult = fixVagueDM(rawString, metaResult, urlDate, opts)
	if validateResult(metaResult) {
		return rawString, metaResult, nil
	}

	// Try to use JSON data
	rawString, jsonResult := jsonSearch(doc, opts)
	jsonResult = fixVagueDM(rawString, jsonResult, urlDate, opts)
	if validateResult(jsonResult) {
		return rawString, jsonResult, nil
	}

	// Try <abbr> elements
	rawString, abbrResult := examineAbbrElements(doc, opts)
	abbrResult = fixVagueDM(rawString, abbrResult, urlDate, opts)
	if validateResult(abbrResult) {
		return rawString, abbrResult, nil
	}

	// Use selectors + text content
	// First try in pruned document
	prunedDoc := dom.Clone(doc, true)
	discarded := discardUnwanted(prunedDoc)
	dateElements := findElementsWithRule(prunedDoc, dateSelectorRule)
	rawString, dateResult := examineOtherElements(dateElements, opts)
	dateResult = fixVagueDM(rawString, dateResult, urlDate, opts)
	if validateResult(dateResult) {
		return rawString, dateResult, nil
	}

	// Search in the discarded elements (currently only footer)
	for _, subTree := range discarded {
		dateElements := findElementsWithRule(subTree, dateSelectorRule)
		rawString, dateResult := examineOtherElements(dateElements, opts)
		dateResult = fixVagueDM(rawString, dateResult, urlDate, opts)
		if validateResult(dateResult) {
			return rawString, dateResult, nil
		}
	}

	// Supply more expressions.
	if !opts.SkipExtensiveSearch {
		dateElements := findElementsWithRule(doc, additionalSelectorRule)
		rawString, dateResult := examineOtherElements(dateElements, opts)
		dateResult = fixVagueDM(rawString, dateResult, urlDate, opts)
		if validateResult(dateResult) {
			return rawString, dateResult, nil
		}
	}

	// Try <time> elements
	rawString, timeResult := examineTimeElements(doc, opts)
	timeResult = fixVagueDM(rawString, timeResult, urlDate, opts)
	if validateResult(timeResult) {
		return rawString, timeResult, nil
	}

	// Try string search
	cleanDocument(doc)

	var htmlString string
	htmlNode := dom.QuerySelector(doc, "html")
	if htmlNode != nil {
		htmlString = dom.InnerHTML(htmlNode)
	} else {
		htmlString = dom.InnerHTML(doc)
	}

	// String search using regex timestamp
	rawString, timestampResult := timestampSearch(htmlString, opts)
	timestampResult = fixVagueDM(rawString, timestampResult, urlDate, opts)
	if validateResult(timestampResult) {
		return rawString, timestampResult, nil
	}

	// Precise patterns and idiosyncrasies
	rawString, textResult := idiosyncrasiesSearch(htmlString, opts)
	textResult = fixVagueDM(rawString, textResult, urlDate, opts)
	if validateResult(textResult) {
		return rawString, textResult, nil
	}

	// Try title elements
	for _, titleElem := range dom.GetAllNodesWithTag(doc, "title", "h1") {
		textContent := normalizeSpaces(dom.TextContent(titleElem))
		_, attempt := tryYmdDate(textContent, opts)
		attempt = fixVagueDM(textContent, attempt, urlDate, opts)
		if validateResult(attempt) {
			log.Debug().Msgf("found date in title: %s", textContent)
			return textContent, attempt, nil
		}
	}

	// Last resort: do extensive search.
	if !opts.SkipExtensiveSearch {
		log.Debug().Msg("extensive search started")

		// Process div and p elements
		// TODO: check all and decide according to original_date
		var refValue int64
		var refString string
		for _, elem := range dom.GetAllNodesWithTag(doc, "div", "p") {
			for _, child := range dom.ChildNodes(elem) {
				if child.Type != html.TextNode {
					continue
				}

				text := normalizeSpaces(child.Data)
				if nText := len(text); nText > 0 && nText < 80 {
					refString, refValue = compareReference(refString, refValue, text, opts)
				}
			}
		}

		// Return
		converted := checkExtractedReference(refValue, opts)
		converted = fixVagueDM(refString, converted, urlDate, opts)
		if validateResult(converted) {
			return refString, converted, nil
		}

		// Search page HTML
		rawString, searchResult := searchPage(htmlString, opts)
		searchResult = fixVagueDM(rawString, searchResult, urlDate, opts)
		if validateResult(searchResult) {
			return rawString, searchResult, nil
		}
	}

	// If nothing else found, try from URL
	if urlDate.IsZero() {
		urlDate = extractPartialUrlDate(opts.URL, opts)
	}

	if !urlDate.IsZero() {
		log.Debug().Msgf("nothing found, just use date from url")
		return opts.URL, urlDate, nil
	}

	// If url doesn't have any date, try to use URL from image metadata
	if urlDate.IsZero() {
		rawString, imgResult := metaImgSearch(doc, opts)
		if !imgResult.IsZero() {
			return rawString, imgResult, nil
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
	// Extract dates
	var metaDate time.Time
	var metaString string

	if opts.UseOriginalDate {
		metaString, metaDate = examineMetaPublish(doc, opts)
	} else {
		metaString, metaDate = examineMetaModified(doc, opts)
	}

	// If date still not found, look for copyright year in metadata
	for _, meta := range dom.QuerySelectorAll(doc, `meta[itemprop]`) {
		itemProp := dom.GetAttribute(meta, "itemprop")
		itemProp = strings.TrimSpace(itemProp)
		itemProp = strings.ToLower(itemProp)
		if itemProp != "copyrightyear" {
			continue
		}

		content := dom.GetAttribute(meta, "content")
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}

		log.Debug().Msgf("examining meta copyright year: %s", dom.OuterHTML(meta))
		attempt := content + "-01-01"
		dt, err := time.Parse("2006-01-02", attempt)
		if err == nil && validateDate(dt, opts) {
			metaString, metaDate = attempt, dt
			break
		}
	}

	return metaString, metaDate
}

// examineMetaPublish parse meta elements to find publish date (original date).
func examineMetaPublish(doc *html.Node, opts Options) (string, time.Time) {
	var metaDate time.Time
	var metaString string

	for _, elem := range dom.QuerySelectorAll(doc, "meta") {
		// Fetch attributes
		name := strings.TrimSpace(dom.GetAttribute(elem, "name"))
		content := strings.TrimSpace(dom.GetAttribute(elem, "content"))
		property := strings.TrimSpace(dom.GetAttribute(elem, "property"))
		pubDate := strings.TrimSpace(dom.GetAttribute(elem, "pubdate"))
		itemProp := strings.TrimSpace(dom.GetAttribute(elem, "itemprop"))
		dateTime := strings.TrimSpace(dom.GetAttribute(elem, "datetime"))
		httpEquiv := strings.TrimSpace(dom.GetAttribute(elem, "http-equiv"))

		if property != "" && content != "" { // Handle property
			attribute := strings.ToLower(property)
			if inMap(attribute, dateAttributes) {
				log.Debug().Msgf("examining meta property for publish date: %s", dom.OuterHTML(elem))
				metaString, metaDate = tryYmdDate(content, opts)
			}
		} else if name != "" && content != "" { // Handle name
			attribute := strings.ToLower(name)
			if inMap(attribute, dateAttributes) {
				log.Debug().Msgf("examining meta name for publish date: %s", dom.OuterHTML(elem))
				metaString, metaDate = tryYmdDate(content, opts)
			}
		} else if pubDate != "" { // Handle publish date
			if strings.ToLower(pubDate) == "pubdate" {
				log.Debug().Msgf("examining meta pubdate: %s", dom.OuterHTML(elem))
				metaString, metaDate = tryYmdDate(content, opts)
			}
		} else if itemProp != "" { // Handle item props
			attribute := strings.ToLower(itemProp)
			if strIn(attribute, "datecreated", "datepublished", "pubyear") {
				log.Debug().Msgf("examining meta itemprop for publish date: %s", dom.OuterHTML(elem))

				var strDate string
				if dateTime != "" {
					strDate = dateTime
				} else if content != "" {
					strDate = content
				}

				if strDate != "" {
					metaString, metaDate = tryYmdDate(strDate, opts)
				}
			}
		} else if httpEquiv != "" { // Handle http-equiv, rare
			// See http://www.standardista.com/html5/http-equiv-the-meta-attribute-explained/
			attribute := strings.ToLower(httpEquiv)
			if attribute == "date" && content != "" {
				log.Debug().Msgf("examining meta http-equiv for publish date: %s", dom.OuterHTML(elem))
				metaString, metaDate = tryYmdDate(content, opts)
			}
		}

		// Exit loop
		if !metaDate.IsZero() {
			break
		}
	}

	return metaString, metaDate
}

// examineMetaModified parse meta elements to find modified date.
func examineMetaModified(doc *html.Node, opts Options) (string, time.Time) {
	var metaDate time.Time
	var metaString string

	for _, elem := range dom.QuerySelectorAll(doc, "meta") {
		// Fetch attributes
		name := strings.TrimSpace(dom.GetAttribute(elem, "name"))
		content := strings.TrimSpace(dom.GetAttribute(elem, "content"))
		property := strings.TrimSpace(dom.GetAttribute(elem, "property"))
		itemProp := strings.TrimSpace(dom.GetAttribute(elem, "itemprop"))
		dateTime := strings.TrimSpace(dom.GetAttribute(elem, "datetime"))
		httpEquiv := strings.TrimSpace(dom.GetAttribute(elem, "http-equiv"))

		if property != "" && content != "" { // Handle property
			attribute := strings.ToLower(property)
			if inMap(attribute, propertyModified) {
				log.Debug().Msgf("examining meta property for modified date: %s", dom.OuterHTML(elem))
				metaString, metaDate = tryYmdDate(content, opts)
			}
		} else if name != "" && content != "" { // Handle name
			attribute := strings.ToLower(name)
			if strIn(attribute, "lastmodified", "last-modified", "lastmod") {
				log.Debug().Msgf("examining meta name for modified date: %s", dom.OuterHTML(elem))
				metaString, metaDate = tryYmdDate(content, opts)
			}
		} else if itemProp != "" { // Handle item props
			attribute := strings.ToLower(itemProp)
			if attribute == "datemodified" {
				log.Debug().Msgf("examining meta itemprop for modified date: %s", dom.OuterHTML(elem))

				var strDate string
				if dateTime != "" {
					strDate = dateTime
				} else if content != "" {
					strDate = content
				}

				if strDate != "" {
					metaString, metaDate = tryYmdDate(strDate, opts)
				}
			}
		} else if httpEquiv != "" { // Handle http-equiv, rare
			// See http://www.standardista.com/html5/http-equiv-the-meta-attribute-explained/
			attribute := strings.ToLower(httpEquiv)
			if attribute == "last-modified" && content != "" {
				log.Debug().Msgf("examining meta http-equiv for modified date: %s", dom.OuterHTML(elem))
				metaString, metaDate = tryYmdDate(content, opts)
			}
		}

		// Exit loop
		if !metaDate.IsZero() {
			break
		}
	}

	return metaString, metaDate
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

			if opts.UseOriginalDate {
				// Look for original date
				if refValue == 0 || candidate < refValue {
					refValue = candidate
					refString = dataUtime
				}
			} else {
				// Look for newest (i.e. largest time delta)
				if candidate > refValue {
					refValue = candidate
					refString = dataUtime
				}
			}
		}

		// Handle class
		if class != "" {
			if strIn(class, "published", "date-published", "time-published") {
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
	candidates := plausibleYearFilterx(htmlString, rxSelectYmdPattern, rxSelectYmdYear, false, opts)
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
	candidates = plausibleYearFilterx(htmlString, rxSlashesPattern, rxSlashesYear, true, opts)
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
	candidates = plausibleYearFilterx(htmlString, rxMmYyyyPattern, rxMmYyyyYear, false, opts)

	// Revert DD-MM-YYYY patterns before sorting
	uniquePatterns := []string{}
	mapPatternCount := make(map[string]int)
	mapPatternRawString := make(map[string]string)

	for _, candidate := range candidates {
		parts := rxMyPattern.FindStringSubmatch(candidate.Patternz)
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

		mapPatternCount[newPattern] += candidate.Occurences
	}

	candidates = make([]yearCandidate, len(uniquePatterns))
	for i, pattern := range uniquePatterns {
		candidates[i] = yearCandidate{
			Patternz:   pattern,
			Occurences: mapPatternCount[pattern],
			RawString:  mapPatternRawString[pattern],
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
		if err == nil && validateDate(dt, opts) {
			if copYear == 0 || dt.Year() >= copYear {
				log.Debug().Msgf("date found for pattern \"%s\": %s", rxSimplePattern.String(), str)
				return rawString, dt
			}
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
	candidates := plausibleYearFilterx(htmlString, rxPattern, rxYearPattern, false, opts)
	return selectCandidate(candidates, rxCatchPattern, rxYearPattern, opts)
}

// selectCandidate selects a candidate among the most frequent matches.
func selectCandidate(candidates []yearCandidate, catchPattern, yearPattern *regexp.Regexp, opts Options) (string, []string) {
	// Prepare variables
	minYear := opts.MinDate.Year()
	maxYear := opts.MaxDate.Year()

	// Make sure candidates exist and less than `maxPossibleCandidates`
	nCandidates := len(candidates)
	if nCandidates == 0 || nCandidates >= maxPossibleCandidates {
		return "", nil
	}

	// If there is only one candidates, check it immediately
	if nCandidates == 1 {
		for _, item := range candidates {
			matches := catchPattern.FindStringSubmatch(item.Patternz)
			if len(matches) > 0 {
				return item.RawString, matches
			}
		}
	}

	// Get 10 most frequent candidates
	sort.SliceStable(candidates, func(a, b int) bool {
		return candidates[a].Occurences > candidates[b].Occurences
	})

	if len(candidates) > 10 {
		candidates = candidates[:10]
	}

	log.Debug().Msgf("top ten occurences: %v", candidates)

	// Sort and find probable candidates
	if !opts.UseOriginalDate {
		sort.SliceStable(candidates, func(a, b int) bool {
			return candidates[a].Patternz > candidates[b].Patternz
		})
	} else {
		sort.SliceStable(candidates, func(a, b int) bool {
			return candidates[a].Patternz < candidates[b].Patternz
		})
	}

	firstCandidate := candidates[0]
	secondCandidate := candidates[1]
	log.Debug().Msgf("best candidate: %v, %v", firstCandidate, secondCandidate)

	// If there are same number of occurences, use the first one
	var matches []string
	var rawString string

	if firstCandidate.Occurences == secondCandidate.Occurences {
		rawString = firstCandidate.RawString
		matches = catchPattern.FindStringSubmatch(firstCandidate.Patternz)
	} else {
		// Get year from the candidate
		year1Parts := yearPattern.FindStringSubmatch(firstCandidate.Patternz)
		year2Parts := yearPattern.FindStringSubmatch(secondCandidate.Patternz)
		if len(year1Parts) < 2 || len(year2Parts) < 2 {
			return "", nil
		}

		year1, _ := strconv.Atoi(year1Parts[1])
		year2, _ := strconv.Atoi(year2Parts[1])

		// Safety net: plausibility
		if year1 < minYear || year1 > maxYear {
			if year2 >= minYear && year2 <= maxYear {
				rawString = secondCandidate.RawString
				matches = catchPattern.FindStringSubmatch(secondCandidate.Patternz)
			} else {
				log.Debug().Msgf("no suitable candidate: %d %d", year1, year2)
			}
		}

		// Safety net: newer date but up to 50% less frequent
		if year2 != year1 && float64(secondCandidate.Occurences)/float64(firstCandidate.Occurences) > 0.5 {
			rawString = secondCandidate.RawString
			matches = catchPattern.FindStringSubmatch(secondCandidate.Patternz)
		} else {
			rawString = firstCandidate.RawString
			matches = catchPattern.FindStringSubmatch(firstCandidate.Patternz)
		}
	}

	return rawString, matches
}

// fixVagueDm fix vague date that uses MM-DD-YYYY format instead of the common DD-MM-YYYY.
func fixVagueDM(rawString string, date, urlDate time.Time, opts Options) time.Time {
	// If date or url date is empty, or date equal url date, return early
	if date.IsZero() || urlDate.IsZero() || date.Equal(urlDate) {
		return date
	}

	// If raw string doesn't contain DMY, return early
	if !rxDmyPattern.MatchString(rawString) {
		return date
	}

	// If date is different with URL date, try to swap day and month
	if !date.Equal(urlDate) {
		swapped := time.Date(date.Year(), time.Month(date.Day()),
			int(date.Month()), 0, 0, 0, 0, date.Location())

		if swapped.Equal(urlDate) {
			return swapped
		}
	}

	// If we are looking for modified date and date is before URL date,
	// try to swap day and month as well.
	if !opts.UseOriginalDate && date.Before(urlDate) {
		swapped := time.Date(date.Year(), time.Month(date.Day()),
			int(date.Month()), 0, 0, 0, 0, date.Location())

		if !swapped.Before(urlDate) {
			return swapped
		}
	}

	// If nothing else, return the original date
	return date
}
