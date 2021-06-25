package htmldate

import "time"

// Options is configuration for the extractor.
type Options struct {
	// UseExtensiveSearch specify whether to skip pattern-based opportunistic
	// text search or not. By the way, the extensive search in this Go port is
	// not as good as the original because over there they use a powerful date
	// parser named `scrapinghub/dateparser`. However, even despite that, the
	// extensive search in this port should be good enough to use.
	// TODO: NEED-DATEPARSER.
	SkipExtensiveSearch bool

	// UseOriginalDate specify whether to extract the original date (e.g. publication
	// date) instead of most recent one (e.g. last modified, updated time).
	UseOriginalDate bool

	// URL is the URL for the webpage.
	URL string

	// MinDate is the earliest acceptable date.
	MinDate time.Time

	// MaxDate is the latest acceptable date.
	MaxDate time.Time

	// EnableLog specify whether log should be enabled or not.
	EnableLog bool
}
