package htmldate

import "time"

// validateDate checks if date is valid and within the possible date.
func validateDate(date time.Time, opts Options) bool {
	// If time is zero, it's not valid
	if date.IsZero() {
		return false
	}

	// If min date specified, make sure our date is after that
	if !opts.MinDate.IsZero() && date.Before(opts.MinDate) {
		return false
	}

	// If max date specified, make sure our date is before that
	if !opts.MaxDate.IsZero() && date.After(opts.MaxDate) {
		return false
	}

	return true
}
