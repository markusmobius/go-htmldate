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

// FromReader extract publish date from the specified reader.
func FromReader(r io.Reader, opts Options) (time.Time, error) {
	// Parse html document
	doc, err := parseHTMLDocument(r)
	if err != nil {
		return timeZero, err
	}

	return FromDocument(doc, opts)
}

// FromDocument extract publish date from the specified html document.
func FromDocument(doc *html.Node, opts Options) (time.Time, error) {
	// Set default options
	if opts.DateFormat == "" {
		opts.DateFormat = defaultDateFormat
	}

	if opts.MinDate.IsZero() {
		opts.MinDate = defaultMinDate
	}

	if opts.MaxDate.IsZero() {
		opts.MaxDate = defaultMaxDate
	}

	// Make sure document exist
	if doc == nil {
		return timeZero, fmt.Errorf("document is empty")
	}

	// Prepare logger
	log = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04",
	}).With().Timestamp().Logger()

	if !opts.EnableLog {
		log = log.Level(zerolog.Disabled)
	}

	// If URL is not defined in options, use canonical link
	if opts.URL == "" {
		for _, link := range dom.QuerySelectorAll(doc, `link[rel="canonical"]`) {
			href := dom.GetAttribute(link, "href")
			href = strings.TrimSpace(href)
			if href != "" {
				opts.URL = href
				break
			}
		}
	}

	// If URL is defined, extract date from it
	if opts.URL != "" {
		dateResult := extractUrlDate(opts.URL, opts)
		if !dateResult.IsZero() {
			return dateResult, nil
		}
	}

	// Try from head elements
	headerResult := examineHeader(doc, opts)
	if !headerResult.IsZero() {
		return headerResult, nil
	}

	// Try to use JSON data
	jsonResult := jsonSearch(doc, opts)
	if !jsonResult.IsZero() {
		return jsonResult, nil
	}

	// Try <abbr> elements
	abbrResult := examineAbbrElements(doc, opts)
	if !abbrResult.IsZero() {
		return abbrResult, nil
	}

	// Use selectors + text content
	// First try in pruned document
	prunedDoc := dom.Clone(doc, true)
	discarded := discardUnwanted(prunedDoc)
	dateElements := findElementsWithRule(prunedDoc, dateSelectorRule)
	dateResult := examineOtherElements(dateElements, opts)
	if !dateResult.IsZero() {
		return dateResult, nil
	}

	// Search in the discarded elements (currently only footer)
	for _, subTree := range discarded {
		dateElements := findElementsWithRule(subTree, dateSelectorRule)
		dateResult := examineOtherElements(dateElements, opts)
		if !dateResult.IsZero() {
			return dateResult, nil
		}
	}

	// Supply more expressions.
	if !opts.SkipExtensiveSearch {
		dateElements := findElementsWithRule(doc, additionalSelectorRule)
		dateResult := examineOtherElements(dateElements, opts)
		if !dateResult.IsZero() {
			return dateResult, nil
		}
	}

	// Try <time> elements
	timeResult := examineTimeElements(doc, opts)
	if !timeResult.IsZero() {
		return timeResult, nil
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
	timestampResult := timestampSearch(htmlString, opts)
	if !timestampResult.IsZero() {
		return timestampResult, nil
	}

	// Precise patterns and idiosyncrasies
	textResult := idiosyncrasiesSearch(htmlString, opts)
	if !textResult.IsZero() {
		return textResult, nil
	}

	// Try title elements
	for _, titleElem := range dom.GetAllNodesWithTag(doc, "title", "h1") {
		textContent := normalizeSpaces(dom.TextContent(titleElem))
		attempt := tryYmdDate(textContent, opts)
		if !attempt.IsZero() {
			log.Debug().Msgf("found date in title: %s", textContent)
			return attempt, nil
		}
	}

	// Retry partial URL
	if opts.URL != "" {
		dateResult := extractPartialUrlDate(opts.URL, opts)
		if !dateResult.IsZero() {
			return dateResult, nil
		}
	}

	// Try meta images
	imgResult := metaImgSearch(doc, opts)
	if !imgResult.IsZero() {
		return imgResult, nil
	}

	// Last resort: do extensive search.
	if !opts.SkipExtensiveSearch {
		log.Debug().Msg("extensive search started")

		// Process div and p elements
		// TODO: check all and decide according to original_date
		var reference int64
		for _, elem := range dom.GetAllNodesWithTag(doc, "div", "p") {
			for _, child := range dom.ChildNodes(elem) {
				if child.Type != html.TextNode {
					continue
				}

				text := normalizeSpaces(child.Data)
				if nText := len(text); nText > 0 && nText < 80 {
					reference = compareReference(reference, text, opts)
				}
			}
		}

		// Return
		converted := checkExtractedReference(reference, opts)
		if !converted.IsZero() {
			return converted, nil
		}

		// Search page HTML
		return searchPage(htmlString, opts), nil
	}

	return timeZero, nil
}

// examineHeader parse meta elements to find date cues.
func examineHeader(doc *html.Node, opts Options) time.Time {
	var headerDate time.Time
	var reserveDate time.Time

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

			if opts.UseOriginalDate {
				if inMap(attribute, dateAttributes) {
					log.Debug().Msgf("examining meta property: %s", dom.OuterHTML(elem))
					headerDate = tryYmdDate(content, opts)
				}
			} else {
				if inMap(attribute, propertyModified) {
					log.Debug().Msgf("examining meta property: %s", dom.OuterHTML(elem))
					headerDate = tryYmdDate(content, opts)
				} else if inMap(attribute, dateAttributes) {
					log.Debug().Msgf("examining meta property: %s", dom.OuterHTML(elem))
					headerDate = tryYmdDate(content, opts)
				}
			}
		} else if name != "" && content != "" { // Handle name
			lowerName := strings.ToLower(name)

			if lowerName == "og:url" {
				headerDate = extractUrlDate(content, opts)
			} else if inMap(lowerName, dateAttributes) {
				log.Debug().Msgf("examining meta name: %s", dom.OuterHTML(elem))
				headerDate = tryYmdDate(content, opts)
			} else if strIn(lowerName, "lastmodified", "last-modified") {
				log.Debug().Msgf("examining meta name: %s", dom.OuterHTML(elem))
				if !opts.UseOriginalDate {
					headerDate = tryYmdDate(content, opts)
				} else {
					reserveDate = tryYmdDate(content, opts)
				}
			}
		} else if pubDate != "" { // Handle publish date
			if strings.ToLower(pubDate) == "pubdate" {
				log.Debug().Msgf("examining meta pubdate: %s", dom.OuterHTML(elem))
				headerDate = tryYmdDate(content, opts)
			}
		} else if itemProp != "" { // Handle item props
			attribute := strings.ToLower(itemProp)

			if strIn(attribute, "datecreated", "datepublished", "pubyear") {
				log.Debug().Msgf("examining meta itemprop: %s", dom.OuterHTML(elem))
				if dateTime != "" {
					headerDate = tryYmdDate(dateTime, opts)
				} else if content != "" {
					headerDate = tryYmdDate(content, opts)
				}
			} else if attribute == "datemodified" && !opts.UseOriginalDate {
				log.Debug().Msgf("examining meta itemprop: %s", dom.OuterHTML(elem))
				if dateTime != "" {
					headerDate = tryYmdDate(dateTime, opts)
				} else if content != "" {
					headerDate = tryYmdDate(content, opts)
				}
			} else if attribute == "copyrightyear" { // reserve with copyrightyear
				log.Debug().Msgf("examining meta itemprop: %s", dom.OuterHTML(elem))
				if content != "" {
					attempt := content + "-01-01"
					dt, err := time.Parse("2006-01-02", attempt)
					if err == nil && validateDate(dt, opts) {
						reserveDate = dt
					}
				}
			}
		} else if httpEquiv != "" { // Handle http-equiv, rare
			// See http://www.standardista.com/html5/http-equiv-the-meta-attribute-explained/
			attribute := strings.ToLower(httpEquiv)

			if attribute == "date" && content != "" {
				log.Debug().Msgf("examining meta http-equiv: %s", dom.OuterHTML(elem))
				headerDate = tryYmdDate(content, opts)
			} else if attribute == "last-modified" && content != "" {
				log.Debug().Msgf("examining meta http-equiv: %s", dom.OuterHTML(elem))
				if !opts.UseOriginalDate {
					headerDate = tryYmdDate(content, opts)
				} else {
					reserveDate = tryYmdDate(content, opts)
				}
			}
		}

		// Exit loop
		if !headerDate.IsZero() {
			break
		}
	}

	// If nothing was found, look for lower granularity (so far: "copyright year")
	if headerDate.IsZero() && !reserveDate.IsZero() {
		log.Debug().Msg("opting for reserve date with less granularity")
		return reserveDate
	}

	return headerDate
}

// examineAbbrElements scans the page for <abbr> elements and check if their content
// contains an eligible date.
func examineAbbrElements(doc *html.Node, opts Options) time.Time {
	elements := dom.GetElementsByTagName(doc, "abbr")

	// Make sure elements exist and less than `maxPossibleCandidates`
	if nElements := len(elements); nElements == 0 || nElements >= maxPossibleCandidates {
		return timeZero
	}

	var reference int64
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
				if reference == 0 {
					reference = candidate
				} else if candidate < reference {
					reference = candidate
				}
			} else {
				// Look for newest (i.e. largest time delta)
				if candidate > reference {
					reference = candidate
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
						attempt := tryYmdDate(tryText, opts)
						if !attempt.IsZero() {
							return attempt
						}
					} else {
						reference = compareReference(reference, tryText, opts)
						if reference > 0 {
							break
						}
					}
				}

				// Dates, not times of the day
				if len(text) > 10 {
					tryText := strings.TrimPrefix(text, "am ")
					log.Debug().Msgf("abbr published found: %s", tryText)
					reference = compareReference(reference, tryText, opts)
				}
			}
		}
	}

	// Convert and return
	converted := checkExtractedReference(reference, opts)
	if !converted.IsZero() {
		return converted
	}

	// Try rescue in abbr content
	abbrElements := dom.GetElementsByTagName(doc, "abbr")
	dateResult := examineOtherElements(abbrElements, opts)
	if !dateResult.IsZero() {
		return dateResult
	}

	return timeZero
}

// examineTimeElements scans the page for <time> elements and check if their content
// contains an eligible date.
func examineTimeElements(doc *html.Node, opts Options) time.Time {
	elements := dom.GetElementsByTagName(doc, "time")

	// Make sure elements exist and less than `maxPossibleCandidates`
	if nElements := len(elements); nElements == 0 || nElements >= maxPossibleCandidates {
		return timeZero
	}

	// Scan all the tags and look for the newest one
	var reference int64
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
				attempt := tryYmdDate(dateTime, opts)
				if !attempt.IsZero() {
					return attempt
				}
			} else {
				reference = compareReference(reference, dateTime, opts)
				if reference > 0 {
					break
				}
			}
		} else if len(text) > 6 { // Bare text in element
			log.Debug().Msgf("time/datetime found in text: %s", text)
			reference = compareReference(reference, text, opts)
		}
	}

	// Return
	converted := checkExtractedReference(reference, opts)
	return converted
}

// examineOtherElements scans the specified elements and check if their content
// contains an eligible date.
func examineOtherElements(elements []*html.Node, opts Options) time.Time {
	// Make sure elements exist and less than `maxPossibleCandidates`
	if nElements := len(elements); nElements == 0 || nElements >= maxPossibleCandidates {
		return timeZero
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

			// Trim non-digits at the end of the string.
			toExamine = rxLastNonDigits.ReplaceAllString(toExamine, "")

			// Log the examined element
			elemHTML := dom.OuterHTML(elem)
			elemHTML = strLimit(normalizeSpaces(elemHTML), 100)
			log.Debug().Msgf("analyzing HTML: %s", elemHTML)

			// Attempt to extract date
			attempt = tryYmdDate(toExamine, opts)
			if !attempt.IsZero() {
				return attempt
			}
		}

		// Try link title (Blogspot)
		titleAttr := strings.TrimSpace(dom.GetAttribute(elem, "title"))
		if titleAttr != "" {
			toExamine := strLimit(titleAttr, 48)
			toExamine = rxLastNonDigits.ReplaceAllString(toExamine, "")
			attempt = tryYmdDate(toExamine, opts)
			if !attempt.IsZero() {
				return attempt
			}
		}
	}

	return timeZero
}

// searchPage opportunistically search the HTML text for common text patterns.
func searchPage(htmlString string, opts Options) time.Time {
	// Copyright symbol
	log.Debug().Msg("looking for copyright/footer information")

	var copYear int
	bestMatch := searchPattern(htmlString, rxCopyrightPattern, rxYearPattern, rxYearPattern, opts)
	if len(bestMatch) > 0 {
		log.Debug().Msgf("copyright detected: %s", bestMatch[0])
		bestMatchVal, err := strconv.Atoi(bestMatch[0])
		if err == nil && bestMatchVal >= opts.MinDate.Year() && bestMatchVal <= opts.MaxDate.Year() {
			log.Debug().Msgf("copyright year/footer pattern found: %d", bestMatchVal)
			copYear = bestMatchVal
		}
	}

	// 3 components
	log.Debug().Msg("3 components")

	// Target URL characteristics
	bestMatch = searchPattern(htmlString, rxThreePattern, rxThreeCatch, rxYearPattern, opts)
	result := filterYmdCandidate(bestMatch, rxThreePattern, copYear, opts)
	if !result.IsZero() {
		return result
	}

	// More loosely structured date
	bestMatch = searchPattern(htmlString, rxThreeLoosePattern, rxThreeLooseCatch, rxYearPattern, opts)
	result = filterYmdCandidate(bestMatch, rxThreeLoosePattern, copYear, opts)
	if !result.IsZero() {
		return result
	}

	// Handle YYYY-MM-DD/DD-MM-YYYY, normalize candidates first
	candidates := plausibleYearFilter(htmlString, rxSelectYmdPattern, rxSelectYmdYear, false, opts)
	candidates = normalizeCandidates(candidates, opts)

	bestMatch = selectCandidate(candidates, rxYmdPattern, rxYmdYear, opts)
	result = filterYmdCandidate(bestMatch, rxSelectYmdPattern, copYear, opts)
	if !result.IsZero() {
		return result
	}

	// Valid dates string
	bestMatch = searchPattern(htmlString, rxDateStringsPattern, rxDateStringsCatch, rxYearPattern, opts)
	result = filterYmdCandidate(bestMatch, rxDateStringsPattern, copYear, opts)
	if !result.IsZero() {
		return result
	}

	// Handle DD?/MM?/YYYY, normalize candidates first
	candidates = plausibleYearFilter(htmlString, rxSlashesPattern, rxSlashesYear, true, opts)
	candidates = normalizeCandidates(candidates, opts)

	bestMatch = selectCandidate(candidates, rxYmdPattern, rxYmdYear, opts)
	result = filterYmdCandidate(bestMatch, rxSlashesPattern, copYear, opts)
	if !result.IsZero() {
		return result
	}

	// 2 components
	log.Debug().Msg("switching to 2 components")

	// First option
	bestMatch = searchPattern(htmlString, rxYyyyMmPattern, rxYyyyMmCatch, rxYearPattern, opts)
	if len(bestMatch) >= 3 {
		str := fmt.Sprintf("%s-%s-1", bestMatch[1], bestMatch[2])
		dt, err := time.Parse("2006-1-2", str)
		if err == nil && validateDate(dt, opts) {
			if copYear == 0 || dt.Year() >= copYear {
				log.Debug().Msgf("date found for pattern \"%s\": %s", rxYyyyMmPattern.String(), str)
				return dt
			}
		}
	}

	// Second option
	candidates = plausibleYearFilter(htmlString, rxMmYyyyPattern, rxMmYyyyYear, false, opts)
	candidates = normalizeCandidates(candidates, opts)

	bestMatch = selectCandidate(candidates, rxYmdPattern, rxYmdYear, opts)
	result = filterYmdCandidate(bestMatch, rxMmYyyyPattern, copYear, opts)
	if !result.IsZero() {
		return result
	}

	// Catch all
	if copYear != 0 {
		log.Debug().Msg("using copyright year as default")
		return time.Date(copYear, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	// 1 component, last try
	log.Debug().Msg("switching to one component")
	bestMatch = searchPattern(htmlString, rxSimplePattern, rxYearPattern, rxYearPattern, opts)
	if len(bestMatch) >= 2 {
		str := fmt.Sprintf("%s-1-1", bestMatch[1])
		dt, err := time.Parse("2006-1-2", str)
		if err == nil && validateDate(dt, opts) {
			if copYear == 0 || dt.Year() >= copYear {
				log.Debug().Msgf("date found for pattern \"%s\": %s", rxSimplePattern.String(), str)
				return dt
			}
		}
	}

	return timeZero
}

// compareReference compares candidate to current date reference
// (includes date validation and older/newer test)
func compareReference(reference int64, expression string, opts Options) int64 {
	attempt := tryExpression(expression, opts)
	if !attempt.IsZero() {
		return compareValues(reference, attempt, opts)
	} else {
		return reference
	}
}

// tryExpression checks if the text string could be a valid date expression.
func tryExpression(expression string, opts Options) time.Time {
	// Trim expression
	expression = normalizeSpaces(expression)
	if expression == "" || getDigitCount(expression) < 4 {
		return timeZero
	}

	// Try the beginning of the string
	expression = strLimit(expression, 48)
	return tryYmdDate(expression, opts)
}

// searchPattern runs chained candidate filtering and selection.
func searchPattern(htmlString string, pattern, catchPattern, yearPattern *regexp.Regexp, opts Options) []string {
	candidates := plausibleYearFilter(htmlString, pattern, yearPattern, false, opts)
	return selectCandidate(candidates, catchPattern, yearPattern, opts)
}

// selectCandidate selects a candidate among the most frequent matches.
func selectCandidate(candidates []yearCandidate, catchPattern, yearPattern *regexp.Regexp, opts Options) []string {
	// Prepare variables
	minYear := opts.MinDate.Year()
	maxYear := opts.MaxDate.Year()

	// Make sure candidates exist and less than `maxPossibleCandidates`
	nCandidates := len(candidates)
	if nCandidates == 0 || nCandidates >= maxPossibleCandidates {
		return nil
	}

	// If there is only one candidates, check it immediately
	if nCandidates == 1 {
		for _, item := range candidates {
			matches := catchPattern.FindStringSubmatch(item.Pattern)
			if len(matches) > 0 {
				return matches
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
			return candidates[a].Pattern > candidates[b].Pattern
		})
	} else {
		sort.SliceStable(candidates, func(a, b int) bool {
			return candidates[a].Pattern < candidates[b].Pattern
		})
	}

	firstCandidate := candidates[0]
	secondCandidate := candidates[1]
	log.Debug().Msgf("best candidate: %v, %v", firstCandidate, secondCandidate)

	// If there are same number of occurences, use the first one
	var matches []string
	if firstCandidate.Occurences == secondCandidate.Occurences {
		matches = catchPattern.FindStringSubmatch(firstCandidate.Pattern)
	} else {
		// Get year from the candidate
		year1Parts := yearPattern.FindStringSubmatch(firstCandidate.Pattern)
		year2Parts := yearPattern.FindStringSubmatch(secondCandidate.Pattern)
		if len(year1Parts) < 2 || len(year2Parts) < 2 {
			return nil
		}

		year1, _ := strconv.Atoi(year1Parts[1])
		year2, _ := strconv.Atoi(year2Parts[1])

		// Safety net: plausibility
		if year1 < minYear || year1 > maxYear {
			if year2 >= minYear && year2 <= maxYear {
				matches = catchPattern.FindStringSubmatch(secondCandidate.Pattern)
			} else {
				log.Debug().Msgf("no suitable candidate: %d %d", year1, year2)
			}
		}

		// Safety net: newer date but up to 50% less frequent
		if year2 != year1 && float64(secondCandidate.Occurences)/float64(firstCandidate.Occurences) > 0.5 {
			matches = catchPattern.FindStringSubmatch(secondCandidate.Pattern)
		} else {
			matches = catchPattern.FindStringSubmatch(firstCandidate.Pattern)
		}
	}

	return matches
}
