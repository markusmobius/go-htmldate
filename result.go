package htmldate

import "time"

var resultZero = Result{}

// Result is the result of date time extraction.
type Result struct {
	DateTime    time.Time
	HasTime     bool
	HasTimezone bool
}

// IsZero reports whether the result is empty or not.
// Wrapper for `Result.DateTime.IsZero`.
func (r Result) IsZero() bool {
	return r.DateTime.IsZero()
}

// Format returns a textual representation of the time value formatted according to
// the specified layout. Wrapper for `Result.DateTime.Format`.
func (r Result) Format(layout string) string {
	return r.DateTime.Format(layout)
}
