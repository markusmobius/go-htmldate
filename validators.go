package htmldate

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type yearCandidate struct {
	Pattern    string
	Occurences int
}

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
// Unlike in the original, here we sort it as well by the highest frequency.
func plausibleYearFilter(htmlString string, pattern, yearPattern *regexp.Regexp, toComplete bool, opts Options) []yearCandidate {
	// Prepare min and max year
	minYear := opts.MinDate.Year()
	maxYear := opts.MaxDate.Year()

	// Find all matches in html string
	uniqueMatches := []string{}
	mapMatchCount := make(map[string]int)

	for _, parts := range pattern.FindAllStringSubmatch(htmlString, -1) {
		match := parts[0]
		if len(parts) > 1 {
			match = parts[1]
		}

		if _, exist := mapMatchCount[match]; !exist {
			uniqueMatches = append(uniqueMatches, match)
		}
		mapMatchCount[match]++
	}

	// Check if matched item is invalid and can be ignored
	validOccurences := []yearCandidate{}
	for _, match := range uniqueMatches {
		// Check if match fulfill the year pattern as well
		var err error
		yearVal := -1
		yearParts := yearPattern.FindStringSubmatch(match)

		if len(yearParts) >= 2 {
			yearVal, err = strconv.Atoi(yearParts[1])
			if err != nil {
				log.Debug().Msgf("not year pattern: %s", match)
				delete(mapMatchCount, match)
				continue
			}
		}

		if yearVal == -1 {
			log.Debug().Msgf("not year pattern: %s (nothing found)", match)
			delete(mapMatchCount, match)
			continue
		}

		// Make sure the year is valid
		var potentialYear int
		if !toComplete {
			potentialYear = yearVal
		} else {
			if yearVal < 100 {
				if yearVal >= 90 {
					yearVal += 1900
				} else {
					yearVal += 2000
				}
			}
		}

		if potentialYear < minYear || potentialYear > maxYear {
			log.Debug().Msgf("not potential year %d: %s", potentialYear, match)
			delete(mapMatchCount, match)
			continue
		}

		// Save the valid matches
		validOccurences = append(validOccurences, yearCandidate{
			Pattern:    match,
			Occurences: mapMatchCount[match],
		})
	}

	return validOccurences
}

// filterYmdCandidate filters free text candidates in the YMD format.
func filterYmdCandidate(bestMatch []string, pattern *regexp.Regexp, copYear int, opts Options) time.Time {
	if len(bestMatch) < 4 {
		return timeZero
	}

	str := fmt.Sprintf("%s-%s-%s", bestMatch[1], bestMatch[2], bestMatch[3])
	dt, err := time.Parse("2006-1-2", str)
	if err != nil || !validateDate(dt, opts) {
		return timeZero
	}

	if copYear == 0 || dt.Year() >= copYear {
		log.Debug().Msgf("date found for pattern %s: %s", pattern.String(), str)
		return dt
	}

	// TODO: test and improve
	// if opts.UseOriginalDate {
	// 	if copYear == 0 || dt.Year() <= copYear {
	// 		log.Debug().Msgf("original date found for pattern %s: %s", pattern.String(), str)
	// 		return dt
	// 	}
	// } else {
	// 	if copYear == 0 || dt.Year() >= copYear {
	// 		log.Debug().Msgf("date found for pattern %s: %s", pattern.String(), str)
	// 		return dt
	// 	}
	// }

	return timeZero
}

func createCandidates(items ...string) []yearCandidate {
	uniqueItems := []string{}
	mapItemCount := make(map[string]int)
	for _, item := range items {
		if _, exist := mapItemCount[item]; !exist {
			uniqueItems = append(uniqueItems, item)
		}
		mapItemCount[item]++
	}

	var candidates []yearCandidate
	for _, item := range uniqueItems {
		candidates = append(candidates, yearCandidate{
			Pattern:    item,
			Occurences: mapItemCount[item],
		})
	}

	return candidates
}

func normalizeCandidates(candidates []yearCandidate, opts Options) []yearCandidate {
	normalizedItems := []string{}
	for _, item := range candidates {
		dt := fastParse(item.Pattern, opts)
		if dt.IsZero() {
			continue
		}

		strDt := dt.Format("2006-01-02")
		for i := 0; i < item.Occurences; i++ {
			normalizedItems = append(normalizedItems, strDt)
		}
	}

	return createCandidates(normalizedItems...)
}
