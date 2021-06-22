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
	if len(parts) != 4 {
		return timeZero
	}

	// Create date from the extracted parts
	year, _ := strconv.Atoi(parts[1])
	month, _ := strconv.Atoi(parts[2])
	day, _ := strconv.Atoi(parts[3])

	if month > 12 {
		return timeZero
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if !validateDate(date, opts) {
		return timeZero
	}

	log.Info().Msgf("found date in url: %s", parts[0])
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

	date := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	if !validateDate(date, opts) {
		return timeZero
	}

	log.Info().Msgf("found partial date in url: %s", parts[0])
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
	// Use regex first
	// First try Y-M-D pattern since it's the one used in ISO-8601
	parts := rxYmdPattern.FindStringSubmatch(s)
	if len(parts) == 4 {
		year, _ := strconv.Atoi(parts[1])
		month, _ := strconv.Atoi(parts[2])
		day, _ := strconv.Atoi(parts[3])

		// Make sure month is at most 12, because if not then it's not YMD
		if month <= 12 {
			dt := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			if validateDate(dt, opts) {
				return dt
			}
		}
	}

	// Next try the D-M-Y pattern since it's the most common date format in the world
	parts = rxDmyPattern.FindStringSubmatch(s)
	if len(parts) == 4 {
		day, _ := strconv.Atoi(parts[1])
		month, _ := strconv.Atoi(parts[2])
		year, _ := strconv.Atoi(parts[3])

		// Append year if necessary
		if year < 100 {
			year = 2000 + year
		}

		// If month is more than 12, swap it with the day
		if month > 12 && day <= 12 {
			day, month = month, day
		}

		// Make sure month is at most 12, because if not then it's one of the bizzare date
		// format, so just leave it to dateparse.
		if month <= 12 {
			dt := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			if validateDate(dt, opts) {
				return dt
			}
		}
	}

	// Finally, just try using dateparse
	dt, err := dateparse.ParseAny(s, dateparse.PreferMonthFirst(false))
	if err == nil && validateDate(dt, opts) {
		return dt
	}

	log.Error().Msgf("failed to parse \"%s\"", s)
	return timeZero
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

// idiosyncrasiesSearch looks for author-written dates throughout the web page.
func idiosyncrasiesSearch(htmlString string, opts Options) time.Time {
	// Do it in order of DE-EN-TR
	result := extractIdiosyncrasy(rxDePattern, htmlString, opts)

	if result.IsZero() {
		result = extractIdiosyncrasy(rxEnPattern, htmlString, opts)
	}

	if result.IsZero() {
		result = extractIdiosyncrasy(rxTrPattern, htmlString, opts)
	}

	return result
}

// extractIdiosyncrasy looks for a precise pattern throughout the web page.
func extractIdiosyncrasy(rxIdiosyncrasy *regexp.Regexp, htmlString string, opts Options) time.Time {
	var candidate time.Time
	parts := rxIdiosyncrasy.FindStringSubmatch(htmlString)
	if len(parts) == 0 {
		return timeZero
	}

	var groups []int
	if len(parts) >= 4 && parts[3] != "" {
		groups = []int{0, 1, 2, 3}
	} else if len(parts) >= 7 && parts[6] != "" {
		groups = []int{0, 4, 5, 6}
	} else {
		return timeZero
	}

	if len(parts[1]) == 4 {
		year, _ := strconv.Atoi(parts[groups[1]])
		month, _ := strconv.Atoi(parts[groups[2]])
		day, _ := strconv.Atoi(parts[groups[3]])
		candidate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
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
			year = 2000 + year
		}

		candidate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	}

	if validateDate(candidate, opts) {
		return candidate
	}

	return timeZero
}
