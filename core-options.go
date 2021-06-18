package htmldate

import "time"

// Options is configuration for the extractor.
type Options struct {
	// UseExtensiveSearch specify whether to activate pattern-based opportunistic
	// text search or not.
	UseExtensiveSearch bool

	// UseOriginalDate specify whether to extract the original date (e.g. publication
	// date) instead of most recent one (e.g. last modified, updated time).
	UseOriginalDate bool

	// URL is the pattern that used to search date in URL.
	URL string

	// MinDate is the earliest acceptable date.
	MinDate time.Time

	// MaxDate is the latest acceptable date.
	MaxDate time.Time

	// EnableLog specify whether log should be enabled or not.
	EnableLog bool
}
