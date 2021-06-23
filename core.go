package htmldate

import (
	"fmt"
	"io"
	"os"
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
	discarded := discardUnwanted(doc)
	dateElements := findElementsWithRule(doc, dateSelectorRule)
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
	// Unfortunately, extensive search is not working right now since it
	// requires `scrapinghub/dateparser`, but I keep this code here to make
	// it similar with the original htmldate, so if in future dateparser is
	// ported to Go, there are not many to change. TODO: NEED-DATEPARSER.
	if opts.UseExtensiveSearch {
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
	htmlString := dom.OuterHTML(doc)

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
	// Like before, unfortunately extensive search is not working right now since it
	// requires `scrapinghub/dateparser`, but I keep this code here so if in future
	// dateparser is ported to Go, there are not many to change. TODO: NEED-DATEPARSER.
	if opts.UseExtensiveSearch {
		log.Debug().Msg("extensive search started")

		// Process div and p elements
		// TODO: check all and decide according to original_date
		var reference int64
		for _, elem := range dom.GetAllNodesWithTag(doc, "div", "p") {
			for _, child := range dom.ChildNodes(elem) {
				if child.Type == html.TextNode {
					reference = compareReference(reference, child.Data, opts)
				}
			}
		}

		// Return
		converted := checkExtractedReference(reference, opts)
		if !converted.IsZero() {
			return converted, nil
		}
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
			if len(elemHTML) > 100 {
				elemHTML = elemHTML[:100]
			}
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
