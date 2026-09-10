// Copyright (C) 2022 Markus Mobius
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code in this file is ported from <https://github.com/adbar/htmldate>
// which available under Apache 2.0 license.

package htmldate

import (
	"time"

	dps "github.com/markusmobius/go-dateparser"
)

// Options configures date extraction.
type Options struct {
	// ExtractTime enables experimental extraction of the time and timezone
	// along with the date.
	ExtractTime bool

	// UseOriginalDate requests the original publication date instead of the
	// most recent date, such as the last-modified date.
	UseOriginalDate bool

	// URL is the page URL. If empty, the extractor looks for a canonical link.
	URL string

	// MinDate is the earliest acceptable date.
	MinDate time.Time

	// MaxDate is the latest acceptable date. If zero, it defaults to the end of
	// the current local calendar day, represented in UTC.
	MaxDate time.Time

	// EnableLog enables debug logging.
	EnableLog bool

	// SkipExtensiveSearch selects fast mode, skipping additional text searches
	// and calls to the external date parser.
	SkipExtensiveSearch bool

	// DeferUrlExtractor defers use of a URL date until metadata and JSON dates
	// have been checked.
	DeferUrlExtractor bool

	// DateParserConfig configures the external go-dateparser library. It is
	// used only when SkipExtensiveSearch is false.
	DateParserConfig *dps.Configuration
}
