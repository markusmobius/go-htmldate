package htmldate

import (
	"fmt"
	"io"
	"os"
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
