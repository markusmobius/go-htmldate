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
	"fmt"
	stdregex "regexp"
	"time"

	dps "github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-htmldate/internal/regexp"
)

var (
	timeZero       = time.Time{}
	defaultMinDate = time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC)
	defaultMaxDate = time.Now().AddDate(1, 0, 0)

	externalParser = &dps.Parser{
		ParserTypes: []dps.ParserType{
			dps.CustomFormat,
			dps.AbsoluteTime,
		},
	}

	externalDpsConfig = &dps.Configuration{
		// DateOrder:           dps.DMY,
		// PreferredDayOfMonth: dps.First,
		PreferredDateSource: dps.Past,
		StrictParsing:       true,
	}
)

const (
	minSegmentLen         = 6
	maxSegmentLen         = 52
	maxPossibleCandidates = 150
	defaultDateFormat     = "2006-1-2"
)

var (
	rxLastNonDigits = regexp.MustCompile(`\D+$`)

	rxDay   = `[0-3]?[0-9]`
	rxMonth = `[0-1]?[0-9]`
	rxYear  = `199[0-9]|20[0-3][0-9]`

	rxYmdNoSepPattern = regexp.MustCompile(`(?:\D|^)(\d{8})(?:\D|$)`)
	rxYmdPattern      = compileRegexF(`(?i)`+
		`(?:\D|^)(?:`+
		`(?P<year>%[1]s)[\-/.](?P<month>%[2]s)[\-/.](?P<day>%[3]s)`+
		`|`+
		`(?P<day>%[3]s)[\-/.](?P<month>%[2]s)[\-/.](?P<year>\d{2,4})`+
		`)(?:\D|$)`, rxYear, rxMonth, rxDay)
	rxYmPattern = compileRegexF(`(?i)`+
		`(?:\D|^)(?:`+
		`(?P<year>%[1]s)[\-/.](?P<month>%[2]s)`+
		`|`+
		`(?P<month>%[2]s)[\-/.](?P<year>%[1]s)`+
		`)(?:\D|$)`, rxYear, rxMonth)

	// TODO: check "août"
	rxMonths = `` +
		`January?|February?|March|A[pv]ril|Ma[iy]|Jun[ei]|Jul[iy]|August|September|O[ck]tober|November|De[csz]ember|` +
		`Jan|Feb|M[aä]r|Apr|Jun|Jul|Aug|Sep|O[ck]t|Nov|De[cz]|` +
		`Januari|Februari|Maret|Mei|Agustus|` +
		`Jänner|Feber|März|` +
		`janvier|février|mars|juin|juillet|aout|septembre|octobre|novembre|décembre|` +
		`Ocak|Şubat|Mart|Nisan|Mayıs|Haziran|Temmuz|Ağustos|Eylül|Ekim|Kasım|Aralık|` +
		`Oca|Şub|Mar|Nis|Haz|Tem|Ağu|Eyl|Eki|Kas|Ara`
	rxLongTextPattern = compileRegexF(`(?i)`+
		`(?P<month>%[2]s)\s(?P<day>%[3]s)(?:st|nd|rd|th)?,? (?P<year>%[1]s)`+
		`|`+
		`(?P<day>%[3]s)(?:st|nd|rd|th|\.)? (?:of )?(?P<month>%[2]s)[,.]? (?P<year>%[1]s)`,
		rxYear, rxMonths, rxDay)

	rxCompleteUrl = compileRegexF(`(?i)\D(%[1]s)[/_-](%[2]s)[/_-](%[3]s)(?:\D|$)`,
		rxYear, rxMonth, rxDay)

	rxTimestampPattern = compileRegexF(`(?i)(%[1]s-%[2]s-%[3]s).[0-9]{2}:[0-9]{2}:[0-9]{2}`,
		rxYear, rxMonth, rxDay)

	rxTextDatePattern = regexp.MustCompile(`(?i)[.:,_/ -]|^\d+$`)

	rxDiscardPattern = regexp.MustCompile(`` +
		`^\d{2}:\d{2}(?: |:|$)|` +
		`^\D*\d{4}\D*$|` +
		`[$€¥Ұ£¢₽₱฿#₹]|` + // currency symbols and special characters
		`[A-Z]{3}[^A-Z]|` + // currency codes
		`(?:^|\D)(?:\+\d{2}|\d{3}|\d{5})\D|` + // tel./IPs/postal codes
		`ftps?|https?|sftp|` + // protocols
		`\.(?:com|net|org|info|gov|edu|de|fr|io)(?:\z|[^\pL\pM\d_])|` + // TLDs
		`IBAN|[A-Z]{2}[0-9]{2}|` + // bank accounts
		`®` + // ©
		``)

	rxEnPattern = regexp.MustCompile(`(?i)(?:date[^0-9"]{0,20}|updated|published|on)(?:[ :])*?([0-9]{1,4})[./]([0-9]{1,2})[./]([0-9]{2,4})`)
	rxDePattern = regexp.MustCompile(`(?i)(?:Datum|Stand|Veröffentlicht am):? ?([0-9]{1,2})\.([0-9]{1,2})\.([0-9]{2,4})`)
	rxTrPattern = regexp.MustCompile(`(?i)` +
		`(?:güncellen?me|yayı(?:m|n)lan?ma) *?(?:tarihi)? *?:? *?([0-9]{1,2})[./]([0-9]{1,2})[./]([0-9]{2,4})` +
		`|` +
		`([0-9]{1,2})[./]([0-9]{1,2})[./]([0-9]{2,4}) *?(?:'de|'da|'te|'ta|’de|’da|’te|’ta|tarihinde) *(?:güncellendi|yayı(?:m|n)landı)`)

	// TODO: merge all idiosyncracy pattern
	// rxIdiosyncracyPattern = regexp.MustCompile(`(?i)` +
	// `(?:date[^0-9"]{0,20}|updated|published|on)(?:[ :])*?([0-9]{1,4})[./]([0-9]{1,2})[./]([0-9]{2,4})` + // EN
	// `|` +
	// `(?i)(?:Datum|Stand|Veröffentlicht am):? ?([0-9]{1,2})\.([0-9]{1,2})\.([0-9]{2,4})` + // DE
	// `|` +
	// `(?:güncellen?me|yayı(?:m|n)lan?ma) *?(?:tarihi)? *?:? *?([0-9]{1,2})[./]([0-9]{1,2})[./]([0-9]{2,4})` +
	// `|` +
	// `([0-9]{1,2})[./]([0-9]{1,2})[./]([0-9]{2,4}) *?(?:'de|'da|'te|'ta|’de|’da|’te|’ta|tarihinde) *(?:güncellendi|yayı(?:m|n)landı)`) // TR

	// Extensive search patterns
	rxYearPattern        = compileRegexF(`^\D?(%s)`, rxYear)
	rxCopyrightPattern   = compileRegexF(`(?:©|\&copy;|Copyright|\(c\))\D*(?:%[1]s-)?(%[1]s)\D`, rxYear)
	rxThreePattern       = regexp.MustCompile(`/([0-9]{4}/[0-9]{2}/[0-9]{2})[01/]`)
	rxThreeCatch         = regexp.MustCompile(`([0-9]{4})/([0-9]{2})/([0-9]{2})`)
	rxThreeLoosePattern  = regexp.MustCompile(`\D([0-9]{4}[/.-][0-9]{2}[/.-][0-9]{2})\D`)
	rxThreeLooseCatch    = regexp.MustCompile(`([0-9]{4})[/.-]([0-9]{2})[/.-]([0-9]{2})`)
	rxSelectYmdPattern   = regexp.MustCompile(`\D([0-3]?[0-9][/.-][01]?[0-9][/.-][0-9]{4})\D`)
	rxSelectYmdYear      = compileRegexF(`(%s)\D?$`, rxYear)
	rxYmdYear            = compileRegexF(`^(%s)`, rxYear)
	rxDateStringsPattern = regexp.MustCompile(`(\D19[0-9]{2}[01][0-9][0-3][0-9]\D|\D20[0-9]{2}[01][0-9][0-3][0-9]\D)`)
	rxDateStringsCatch   = compileRegexF(`(%s)([01][0-9])([0-3][0-9])`, rxYear)
	rxSlashesPattern     = regexp.MustCompile(`\D([0-3]?[0-9]/[01]?[0-9]/[0129][0-9]|[0-3][0-9]\.[01][0-9]\.[0129][0-9])\D`)
	rxSlashesYear        = regexp.MustCompile(`([0-9]{2})$`)
	rxYyyyMmPattern      = regexp.MustCompile(`\D([12][0-9]{3}[/.-][01][0-9])\D`)
	rxYyyyMmCatch        = compileRegexF(`(%s)[/.-]([01][0-9])`, rxYear)
	rxMmYyyyPattern      = regexp.MustCompile(`\D([01]?[0-9][/.-][12][0-9]{3})\D`)
	rxMmYyyyYear         = compileRegexF(`(%s)\D?$`, rxYear)
	rxSimplePattern      = compileRegexF(`\D(%s)\D`, rxYear)
	rxSimpleW3Cleaner    = compileRegexF(`w3.org\D(%s)\D`, rxYear)

	// Time patterns
	rxCommonTime = regexp.MustCompile(`(?i)(?:\D|^)(\d{1,2})(?::|\s*h\s*)(\d{1,2})(?::(\d{1,2})(?:\.\d+)?)?(?:\s*((?:a|p)\.?m\.?))?`)
	rxTzCode     = stdregex.MustCompile(`(?i)(?:\s|^)([-+])(\d{2})(?::?(\d{2}))?`)
	rxIsoTime    = stdregex.MustCompile(`(?i)(\d{2}):(\d{2})(?::(\d{2})(?:\.\d+)?)?(Z|[+-]\d{2}(?::?\d{2})?)`)

	rxLastJsonBracket = regexp.MustCompile(`(?i)\s*\}$`)
)

// English, French, German, Indonesian and Turkish dates cache
var monthNumber = func() map[string]int {
	var monthNames = [][]string{
		{"jan", "januar", "jänner", "january", "januari", "janvier", "ocak", "oca"},
		{"feb", "februar", "feber", "february", "februari", "février", "şubat", "şub"},
		{"mar", "mär", "märz", "march", "maret", "mart", "mars"},
		{"apr", "april", "avril", "nisan", "nis"},
		{"may", "mai", "mei", "mayıs"},
		{"jun", "juni", "june", "juin", "haziran", "haz"},
		{"jul", "juli", "july", "juillet", "temmuz", "tem"},
		{"aug", "august", "agustus", "ağustos", "ağu", "aout"},
		{"sep", "september", "septembre", "eylül", "eyl"},
		{"oct", "oktober", "october", "octobre", "okt", "ekim", "eki"},
		{"nov", "november", "kasım", "kas", "novembre"},
		{"dec", "dez", "dezember", "december", "desember", "décembre", "aralık", "ara"},
	}

	mapNameNumber := make(map[string]int)
	for i, names := range monthNames {
		for _, name := range names {
			mapNameNumber[name] = i + 1
		}
	}

	return mapNameNumber
}()

var dateAttributes = sliceToMap(
	"analyticsattributes.articledate",
	"article.created",
	"article_date_original",
	"article:post_date",
	"article.published",
	"article:published",
	"article:published_date",
	"article:published_time",
	"article:publicationdate",
	"bt:pubdate",
	"citation_date",
	"citation_publication_date",
	"content_create_date",
	"created",
	"cxenseparse:recs:publishtime",
	"date",
	"date_created",
	"date_published",
	"datecreated",
	"dateposted",
	"datepublished",
	// Dublin Core: https://wiki.whatwg.org/wiki/MetaExtensions
	"dc.date",
	"dc.created",
	"dc.date.created",
	"dc.date.issued",
	"dc.date.publication",
	"dcsext.articlefirstpublished",
	"dcterms.created",
	"dcterms.date",
	"dcterms.issued",
	"dc:created",
	"dc:date",
	"displaydate",
	"doc_date",
	"field-name-post-date",
	"gentime",
	"mediator_published_time",
	"meta", // too loose?
	// Open Graph: https://opengraphprotocol.org/
	"og:article:published",
	"og:article:published_time",
	"og:datepublished",
	"og:pubdate",
	"og:publish_date",
	"og:published_time",
	"og:question:published_time",
	"originalpublicationdate",
	"parsely-pub-date",
	"pdate",
	"ptime",
	"pubdate",
	"publishdate",
	"publish_date",
	"publish_time",
	"publish-date",
	"published-date",
	"published_date",
	"published_time",
	"publisheddate",
	"publication_date",
	"rbpubdate",
	"release_date",
	"rnews:datepublished",
	"sailthru.date",
	"shareaholic:article_published_time",
	"timestamp",
	"twt-published-at",
	"video:release_date",
	"vr:published_time")

var propertyModified = sliceToMap(
	"article:modified",
	"article:modified_date",
	"article:modified_time",
	"article:post_modified",
	"bt:moddate",
	"datemodified",
	"dc.modified",
	"dcterms.modified",
	"lastmodified",
	"modified_time",
	"modificationdate",
	"og:article:modified_time",
	"og:modified_time",
	"og:updated_time",
	"release_date",
	"revision_date",
	"updated_time")

var (
	attrModifiedNames = sliceToMap(
		"lastdate",
		"lastmod",
		"lastmodified",
		"last-modified",
		"modified",
		"utime")
	attrPublishClasses = sliceToMap("published", "date-published", "time-published")

	listItemPropAttrs = []string{"datecreated", "datepublished", "pubyear", "datemodified", "dateupdate"}
	itemPropAttrKeys  = sliceToMap(listItemPropAttrs...)
	itemPropOriginal  = sliceToMap(listItemPropAttrs[:3]...)
	itemPropModified  = sliceToMap(listItemPropAttrs[3:]...)
)

func sliceToMap(strings ...string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, s := range strings {
		result[s] = struct{}{}
	}
	return result
}

func compileRegexF(pattern string, args ...any) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(pattern, args...))
}
