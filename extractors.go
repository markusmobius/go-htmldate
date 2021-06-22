package htmldate

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/go-shiori/dom"
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
	if len(parts) == 0 {
		return timeZero
	}

	// Create date from the extracted parts
	year, _ := strconv.Atoi(parts[1])
	month, _ := strconv.Atoi(parts[2])
	day, _ := strconv.Atoi(parts[3])

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if !validateDate(date, opts) {
		return timeZero
	}

	log.Info().Msgf("found date in url: %s", parts[0])
	return date
}

// tryYmdDate tries to extract date which contains year, month and day using
// a series of heuristics and rules.
func tryYmdDate(s string, opts Options) time.Time {
	// If string less than 6 runes, stop
	if len(s) < 6 {
		return timeZero
	}

	// Count how many digit number in this string
	nDigit := getDigitCount(s)
	if nDigit < 4 || nDigit > 18 {
		return timeZero
	}

	// Check if string only contains time/single year or digits and not a date
	if !rxTextDatePattern.MatchString(s) || rxNoTextDatePattern.MatchString(s) {
		return timeZero
	}

	// Try to parse date
	parseResult := fastParse(s, opts)
	if !parseResult.IsZero() {
		return parseResult
	}

	if opts.UseExtensiveSearch {
		// TODO: NEED-DATEPARSER
		// In original library they can run extensive (but slow) date parsing using
		// `scrapinghub/dateparser` which can parse date from almost any string in
		// many languages. Unfortunately we haven't ported it so we will skip it.
	}

	return timeZero
}

// fastParse parse the string into time.Time.
// In the original Python library, this function is named `custom_parse`, but I
// renamed it to `fastParse` because I think it's more suitable to its purpose.
func fastParse(s string, opts Options) time.Time {
	dt, err := dateparse.ParseAny(s)
	if err != nil {
		log.Error().Msgf("failed to parse \"%s\": %v", s, err)
	}

	if !validateDate(dt, opts) {
		return timeZero
	}

	return dt
}

// jsonSearch looks for JSON time patterns in JSON sections of the document.
func jsonSearch(doc *html.Node, opts Options) time.Time {
	// Determine pattern
	var rxJson *regexp.Regexp
	if opts.UseOriginalDate {
		rxJson = rxJsonPatternPublished
	} else {
		rxJson = rxJsonPatternModified
	}

	// Look throughout the HTML tree
	ldJsonScripts := dom.QuerySelectorAll(doc, `script[type="application/ld+json"]`)
	settingsJsonScripts := dom.QuerySelectorAll(doc, `script[type="application/settings+json"]`)
	scriptNodes := append(ldJsonScripts, settingsJsonScripts...)

	for _, elem := range scriptNodes {
		// Get the json text inside the script
		jsonText := dom.TextContent(elem)
		jsonText = strings.TrimSpace(jsonText)
		if jsonText == "" || !strings.Contains(jsonText, `"date`) {
			continue
		}

		parts := rxJson.FindStringSubmatch(jsonText)
		if len(parts) != 0 {
			dt, err := time.Parse("2006-01-02", parts[1])
			if err == nil && validateDate(dt, opts) {
				return dt
			}
		}
	}

	return timeZero
}

// timestampSearch looks for timestamps throughout the html string.
func timestampSearch(htmlString string, opts Options) time.Time {
	parts := rxTimestampPattern.FindStringSubmatch(htmlString)
	if len(parts) != 0 {
		return fastParse(parts[1], opts)
	}

	return timeZero
}
