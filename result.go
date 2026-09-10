package htmldate

import "time"

var resultZero = Result{}

// Result contains an extracted date and optional time and timezone information.
type Result struct {
	// DateTime is the extracted date, including the time and timezone when found.
	DateTime time.Time
	// HasTime reports whether a time was extracted.
	HasTime bool
	// HasTimezone reports whether an explicit timezone was extracted.
	// It distinguishes a detected UTC timezone from the UTC fallback.
	HasTimezone bool
	// SrcString is the normalized source text used for extraction.
	SrcString string
}

// IsZero reports whether DateTime is zero.
func (r Result) IsZero() bool {
	return r.DateTime.IsZero()
}

// Format formats DateTime using a Go reference-time layout.
func (r Result) Format(layout string) string {
	return r.DateTime.Format(layout)
}
