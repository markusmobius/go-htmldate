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

// compareValues compares the date expression to a reference.
func compareValues(reference int64, attempt time.Time, opts Options) int64 {
	timestamp := attempt.Unix()
	if opts.UseOriginalDate {
		if reference == 0 || timestamp < reference {
			reference = timestamp
		}
	} else {
		if timestamp > reference {
			reference = timestamp
		}
	}

	return reference
}

// checkExtractedReference tests if the extracted reference date can be returned.
func checkExtractedReference(reference int64, opts Options) time.Time {
	if reference > 0 {
		dt := time.Unix(reference, 0)
		if validateDate(dt, opts) {
			return dt
		}
	}
	return timeZero
}
