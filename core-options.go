package htmldate

import "time"

// Options is configuration for the extractor.
type Options struct {
	// UseExtensiveSearch specify whether to activate pattern-based opportunistic
	// text search or not. Unfortunately, to do it we need to port `scrapinghub/dateparser`
	// first, so for now this options is useless. TODO: NEED-DATEPARSER.
	UseExtensiveSearch bool

	// UseOriginalDate specify whether to extract the original date (e.g. publication
	// date) instead of most recent one (e.g. last modified, updated time).
	UseOriginalDate bool

	// DateFormat is the format of date that want to be returned.
	DateFormat string

	// URL is the URL for the webpage.
	URL string

	// MinDate is the earliest acceptable date.
	MinDate time.Time

	// MaxDate is the latest acceptable date.
	MaxDate time.Time

	// EnableLog specify whether log should be enabled or not.
	EnableLog bool
}
