package htmldate

import (
	"regexp"
	"strconv"
	"time"
)

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

// plausibleYearFilter filters the date patterns to find plausible years only.
func plausibleYearFilter(htmlString string, pattern, yearPattern *regexp.Regexp, toComplete bool, opts Options) map[string]struct{} {
	minYear := opts.MinDate.Year()
	maxYear := opts.MaxDate.Year()
	allMatches := pattern.FindAllString(htmlString, -1)
	occurences := sliceToMap(allMatches...)

	for item := range occurences {
		yearParts := yearPattern.FindStringSubmatch(item)

		var err error
		var yearVal int
		if len(yearParts) >= 2 {
			yearVal, err = strconv.Atoi(yearParts[1])
			if err != nil {
				log.Debug().Msgf("not year pattern: %s", item)
				continue
			}
		}

		var potentialYear int
		if !toComplete {
			potentialYear = yearVal
		} else {
			if yearVal < 100 {
				if yearVal >= 90 {
					yearVal = 1900 + yearVal
				} else {
					yearVal = 2000 + yearVal
				}
			}
		}

		if potentialYear < minYear || potentialYear > maxYear {
			log.Debug().Msgf("not potential year: %s", item)
			delete(occurences, item)
		}
	}

	return occurences
}
