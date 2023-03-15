// This file is part of go-htmldate, Go package for extracting publication dates from a web page.
// Source available in <https://github.com/markusmobius/go-htmldate>.
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
	"time"

	dps "github.com/markusmobius/go-dateparser"
)

// Options is configuration for the extractor.
type Options struct {
	// ExtractTime specify if we want to extract publish time as well along the date. Still WIP.
	ExtractTime bool

	// UseOriginalDate specify whether to extract the original date (e.g. publication date) instead
	// of most recent one (e.g. last modified, updated time).
	UseOriginalDate bool

	// URL is the URL for the webpage.
	URL string

	// MinDate is the earliest acceptable date.
	MinDate time.Time

	// MaxDate is the latest acceptable date.
	MaxDate time.Time

	// EnableLog specify whether log should be enabled or not.
	EnableLog bool

	// SkipExtensiveSearch specify whether to skip pattern-based opportunistic text search or not
	// using the external `dateparser` library. Note: this extensive search might be quite slow,
	// so use as necessary.
	SkipExtensiveSearch bool

	// DateParserConfig is configuration for the external `dateparser`. Only used extensive search
	// is enabled (`SkipExtensiveSearch=false`).
	DateParserConfig *dps.Configuration
}
