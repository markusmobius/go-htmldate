// This file is part of go-htmldate, Go package for extracting publication dates from a web page.
// Source available in <https://github.com/markusmobius/go-trafilatura>.
// Copyright (C) 2022 Markus Mobius
//
// This program is free software: you can redistribute it and/or modify it under the terms of
// the GNU General Public License as published by the Free Software Foundation, either version 3
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with this program.
// If not, see <https://www.gnu.org/licenses/>.

// Code in this file is ported from <https://github.com/adbar/htmldate> which available under
// GNU GPL v3 license.

package htmldate

import (
	"regexp"
	"strconv"
	"time"
)

type yearCandidate struct {
	Patternz   string
	Occurences int
	RawString  string
}

// validateDateParts checks if date parts can be used to generate a valid date
func validateDateParts(year, month, day int, opts Options) (time.Time, bool) {
	// Make sure year is in Gregorian era
	if year < 1582 {
		return timeZero, false
	}

	// Make sure month is valid
	if month < 1 || month > 12 {
		return timeZero, false
	}

	// Make sure day is valid
	if day < 1 {
		return timeZero, false
	}

	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		if day > 31 {
			return timeZero, false
		}

	case 4, 6, 9, 11:
		if day > 30 {
			return timeZero, false
		}

	case 2:
		isLeap := isLeapYear(year)
		if (isLeap && day > 29) || (!isLeap && day > 28) {
			return timeZero, false
		}
	}

	// Generate date
	dt := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	valid := validateDate(dt, opts)
	return dt, valid
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
func compareValues(reference int64, attempt time.Time, opts Options) (int64, bool) {
	changed := false
	timestamp := attempt.Unix()

	if (opts.UseOriginalDate && (reference == 0 || timestamp < reference)) ||
		(!opts.UseOriginalDate && timestamp > reference) {
		changed = true
		reference = timestamp
	}

	return reference, changed
}

// checkExtractedReference tests if the extracted reference date can be returned.
func checkExtractedReference(reference int64, opts Options) time.Time {
	if reference > 0 {
		dt := time.Unix(reference, 0).UTC()
		if validateDate(dt, opts) {
			return dt
		}
	}
	return timeZero
}

// plausibleYearFilter filters the date patterns to find plausible years only.
// Unlike in the original, here we sort it as well by the highest frequency.
func plausibleYearFilter(htmlString string, rxPattern, rxYearPattern *regexp.Regexp, toComplete bool, opts Options) []yearCandidate {
	// Prepare min and max year
	minYear := opts.MinDate.Year()
	maxYear := opts.MaxDate.Year()

	// Find all matches in html string
	uniqueMatches := []string{}
	mapMatchCount := make(map[string]int)
	mapMatchRawString := make(map[string]string)

	for _, idxs := range rxPattern.FindAllStringSubmatchIndex(htmlString, -1) {
		var match string
		if len(idxs) > 2 {
			match = htmlString[idxs[2]:idxs[3]]
		} else {
			match = htmlString[idxs[0]:idxs[1]]
		}

		if _, exist := mapMatchCount[match]; !exist {
			rawString := strLimit(htmlString[idxs[0]:], 100)
			uniqueMatches = append(uniqueMatches, match)
			mapMatchRawString[match] = rawString
		}

		mapMatchCount[match]++
	}

	// Check if matched item is invalid and can be ignored
	validOccurences := []yearCandidate{}
	for _, match := range uniqueMatches {
		// Check if match fulfill the year pattern as well
		var err error
		yearVal := -1
		yearParts := rxYearPattern.FindStringSubmatch(match)

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
		} else if yearVal < 100 {
			if yearVal >= 90 {
				potentialYear = 1900 + yearVal
			} else {
				potentialYear = 2000 + yearVal
			}
		}

		if potentialYear < minYear || potentialYear > maxYear {
			log.Debug().Msgf("not potential year %d: %s", potentialYear, match)
			delete(mapMatchCount, match)
			continue
		}

		// Save the valid matches
		validOccurences = append(validOccurences, yearCandidate{
			Patternz:   match,
			Occurences: mapMatchCount[match],
			RawString:  mapMatchRawString[match],
		})
	}

	return validOccurences
}

// filterYmdCandidate filters free text candidates in the YMD format.
func filterYmdCandidate(bestMatch []string, pattern *regexp.Regexp, copYear int, opts Options) time.Time {
	if len(bestMatch) < 4 {
		return timeZero
	}

	year, _ := strconv.Atoi(bestMatch[1])
	month, _ := strconv.Atoi(bestMatch[2])
	day, _ := strconv.Atoi(bestMatch[3])
	dt, valid := validateDateParts(year, month, day, opts)
	if !valid {
		return timeZero
	}

	if copYear == 0 || dt.Year() >= copYear {
		s := dt.Format("2006-01-02")
		log.Debug().Msgf("date found for pattern %s: %s", pattern.String(), s)
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

func normalizeCandidates(candidates []yearCandidate, opts Options) []yearCandidate {
	uniquePatterns := []string{}
	mapPatternCount := make(map[string]int)
	mapPatternRawString := make(map[string]string)

	for _, candidate := range candidates {
		dt := fastParse(candidate.Patternz, opts)
		if dt.IsZero() {
			continue
		}

		newPattern := dt.Format("2006-01-02")
		if _, exist := mapPatternCount[newPattern]; !exist {
			uniquePatterns = append(uniquePatterns, newPattern)
			mapPatternRawString[newPattern] = candidate.RawString
		}

		mapPatternCount[newPattern] += candidate.Occurences
	}

	normalizedCandidates := make([]yearCandidate, len(uniquePatterns))
	for i, pattern := range uniquePatterns {
		normalizedCandidates[i] = yearCandidate{
			Patternz:   pattern,
			Occurences: mapPatternCount[pattern],
			RawString:  mapPatternRawString[pattern],
		}
	}

	return normalizedCandidates
}
