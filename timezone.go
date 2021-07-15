package htmldate

import (
	"strconv"
	"strings"
	"time"
)

// parseTimezoneCode returns the location for the specified timezone code.
func parseTimezoneCode(tzCode string) *time.Location {
	// If it's equal to Z, it's UTC
	tzCode = strings.ToUpper(tzCode)
	if tzCode == "Z" {
		return time.UTC
	}

	// Try ISO timezone format
	parts := rxTzCode.FindStringSubmatch(tzCode)
	if len(parts) > 0 {
		hour, _ := strconv.Atoi(parts[2])
		minute, _ := strconv.Atoi(parts[3])

		offset := hour*3_600 + minute*60
		if parts[1] == "-" {
			offset *= -1
		}

		return time.FixedZone(tzCode, offset)
	}

	// If nothing found, return nil
	return nil
}
