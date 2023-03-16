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
	"regexp"
	"time"

	dps "github.com/markusmobius/go-dateparser"
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
	maxTextSize           = 48
	maxPossibleCandidates = 150
	defaultDateFormat     = "2006-1-2"
)

var (
	rxLastNonDigits = regexp.MustCompile(`\D+$`)

	rxYmdNoSepPattern = regexp.MustCompile(`(?:\D|^)(\d{8})(?:\D|$)`)
	rxYmdPattern      = regexp.MustCompile(`(?:\D|^)(\d{4})[\-/.](\d{1,2})[\-/.](\d{1,2})(?:\D|$)`)
	rxDmyPattern      = regexp.MustCompile(`(?:\D|^)(\d{1,2})[\-/.](\d{1,2})[\-/.](\d{2,4})(?:\D|$)`)
	rxYmPattern       = regexp.MustCompile(`(?:\D|^)(\d{4})[\-/.](\d{1,2})(?:\D|$)`)
	rxMyPattern       = regexp.MustCompile(`(?:\D|^)(\d{1,2})[\-/.](\d{4})(?:\D|$)`)

	// # TODO: check "août"
	rxMonths = `` +
		`January|February|March|April|May|June|July|August|September|October|November|December|` +
		`Januari|Februari|Maret|Mei|Juni|Juli|Agustus|Oktober|Desember|` +
		`Jan|Feb|Mar|Apr|Jun|Jul|Aug|Sep|Oct|Nov|Dec|` +
		`Januar|Jänner|Februar|Feber|März|Mai|Dezember|` +
		`janvier|février|mars|avril|mai|juin|juillet|aout|septembre|octobre|novembre|décembre|` +
		`Ocak|Şubat|Mart|Nisan|Mayıs|Haziran|Temmuz|Ağustos|Eylül|Ekim|Kasım|Aralık|` +
		`Oca|Şub|Mar|Nis|Haz|Tem|Ağu|Eyl|Eki|Kas|Ara`
	rxLongMdyPattern = regexp.MustCompile(
		`(?i)(` + rxMonths + `)\s([0-9]{1,2})(?:st|nd|rd|th)?,? ([0-9]{4})`)
	rxLongDmyPattern = regexp.MustCompile(
		`(?i)([0-9]{1,2})(?:st|nd|rd|th|\.)? (?:of )?(` + rxMonths + `),? ([0-9]{4})`)

	rxCompleteUrl = regexp.MustCompile(`(?i)\D([0-9]{4})[/_-]([0-9]{1,2})[/_-]([0-9]{1,2})(?:\D|$)`)
	rxPartialUrl  = regexp.MustCompile(`(?i)\D([0-9]{4})[/_-]([0-9]{2})(?:\D|$)`)

	rxTimestampPattern  = regexp.MustCompile(`(?i)([0-9]{4}-[0-9]{2}-[0-9]{2}|[0-9]{2}\.[0-9]{2}\.[0-9]{4}).[0-9]{2}:[0-9]{2}:[0-9]{2}`)
	rxTextDatePattern   = regexp.MustCompile(`(?i)[.:,_/ -]|^\d+$`)
	rxNoTextDatePattern = regexp.MustCompile(`(?i)^(?:\d{3,}\D+\d{3,}|\d{2}:\d{2}(:| )|\+\d{2}\D+|\D*\d{4}\D*$)`)

	rxDiscardPattern = regexp.MustCompile(`(?i)[$€¥Ұ£¢₽₱฿#]|CNY|EUR|GBP|JPY|USD|http|\.(com|net|org)|IBAN|\+\d{2}\b`)
	// TODO: further testing required:
	// \d[,.]\d+  // currency amounts
	// # \b\d{5}\s  // postal codes

	rxEnPattern = regexp.MustCompile(`(?i)(?:date[^0-9"]{0,20}|updated|published) *?(?:in)? *?:? *?([0-9]{1,4})[./]([0-9]{1,2})[./]([0-9]{2,4})`)
	rxDePattern = regexp.MustCompile(`(?i)(?:Datum|Stand): ?([0-9]{1,2})\.([0-9]{1,2})\.([0-9]{2,4})`)
	rxTrPattern = regexp.MustCompile(`(?i)` +
		`(?:güncellen?me|yayı(?:m|n)lan?ma) *?(?:tarihi)? *?:? *?([0-9]{1,2})[./]([0-9]{1,2})[./]([0-9]{2,4})|` +
		`([0-9]{1,2})[./]([0-9]{1,2})[./]([0-9]{2,4}) *?(?:'de|'da|'te|'ta|’de|’da|’te|’ta|tarihinde) *(?:güncellendi|yayı(?:m|n)landı)`)

	// Extensive search patterns
	rxYearPattern        = regexp.MustCompile(`^\D?(199[0-9]|20[0-9]{2})`)
	rxCopyrightPattern   = regexp.MustCompile(`(?:©|\&copy;|Copyright|\(c\))\D*(?:[12][0-9]{3}-)?([12][0-9]{3})\D`)
	rxThreePattern       = regexp.MustCompile(`/([0-9]{4}/[0-9]{2}/[0-9]{2})[01/]`)
	rxThreeCatch         = regexp.MustCompile(`([0-9]{4})/([0-9]{2})/([0-9]{2})`)
	rxThreeLoosePattern  = regexp.MustCompile(`\D([0-9]{4}[/.-][0-9]{2}[/.-][0-9]{2})\D`)
	rxThreeLooseCatch    = regexp.MustCompile(`([0-9]{4})[/.-]([0-9]{2})[/.-]([0-9]{2})`)
	rxSelectYmdPattern   = regexp.MustCompile(`\D([0-3]?[0-9][/.-][01]?[0-9][/.-][0-9]{4})\D`)
	rxSelectYmdYear      = regexp.MustCompile(`(19[0-9]{2}|20[0-9]{2})\D?$`)
	rxYmdYear            = regexp.MustCompile(`^([0-9]{4})`)
	rxDateStringsPattern = regexp.MustCompile(`(\D19[0-9]{2}[01][0-9][0-3][0-9]\D|\D20[0-9]{2}[01][0-9][0-3][0-9]\D)`)
	rxDateStringsCatch   = regexp.MustCompile(`([12][0-9]{3})([01][0-9])([0-3][0-9])`)
	rxSlashesPattern     = regexp.MustCompile(`\D([0-3]?[0-9][/.][01]?[0-9][/.][0129][0-9])\D`)
	rxSlashesYear        = regexp.MustCompile(`([0-9]{2})$`)
	rxYyyyMmPattern      = regexp.MustCompile(`\D([12][0-9]{3}[/.-][01][0-9])\D`)
	rxYyyyMmCatch        = regexp.MustCompile(`([12][0-9]{3})[/.-]([01][0-9])`)
	rxMmYyyyPattern      = regexp.MustCompile(`\D([01]?[0-9][/.-][12][0-9]{3})\D`)
	rxMmYyyyYear         = regexp.MustCompile(`([12][0-9]{3})\D?$`)
	rxSimplePattern      = regexp.MustCompile(`\D(199[0-9]|20[0-9]{2})\D`)

	// Time patterns
	rxTzCode     = regexp.MustCompile(`(?i)(?:\s|^)([-+])(\d{2})(?::?(\d{2}))?`)
	rxIsoTime    = regexp.MustCompile(`(?i)(\d{2}):(\d{2})(?::(\d{2})(?:\.\d+)?)?(Z|[+-]\d{2}(?::?\d{2})?)`)
	rxCommonTime = regexp.MustCompile(`(?i)(?:\D|^)(\d{1,2})(?::|\s*h\s*)(\d{1,2})(?::(\d{1,2})(?:\.\d+)?)?(?:\s*((?:a|p)\.?m\.?))?`)

	rxLastJsonBracket = regexp.MustCompile(`(?i)\s*\}$`)
)

// English, French, German, Indonesian and Turkish dates cache
var monthNumber = map[string]int{
	"januar": 1, "jänner": 1, "january": 1, "januari": 1, "jan": 1, "ocak": 1, "oca": 1, "janvier": 1,
	"februar": 2, "feber": 2, "february": 2, "februari": 2, "feb": 2, "şubat": 2, "şub": 2, "février": 2,
	"märz": 3, "march": 3, "maret": 3, "mar": 3, "mart": 3, "mars": 3,
	"april": 4, "apr": 4, "nisan": 4, "nis": 4, "avril": 4,
	"mai": 5, "may": 5, "mei": 5, "mayıs": 5,
	"juni": 6, "june": 6, "jun": 6, "haziran": 6, "haz": 6, "juin": 6,
	"juli": 7, "july": 7, "jul": 7, "temmuz": 7, "tem": 7, "juillet": 7,
	"august": 8, "agustus": 8, "aug": 8, "ağustos": 8, "ağu": 8, "aout": 8,
	"september": 9, "sep": 9, "eylül": 9, "eyl": 9, "septembre": 9,
	"oktober": 10, "october": 10, "oct": 10, "ekim": 10, "eki": 10, "octobre": 10,
	"november": 11, "nov": 11, "kasım": 11, "kas": 11, "novembre": 11,
	"dezember": 12, "december": 12, "desember": 12, "dec": 12, "aralık": 12, "ara": 12, "décembre": 12,
}

var dateAttributes = sliceToMap(
	"article.created", "article_date_original",
	"article.published", "article:published_time",
	"article:publicationdate",
	"bt:pubdate", "citation_date", "citation_publication_date",
	"created", "cxenseparse:recs:publishtime",
	"date", "date_published",
	"datecreated", "dateposted", "datepublished",
	// Dublin Core: https://wiki.whatwg.org/wiki/MetaExtensions
	"dc.date", "dc.created", "dc.date.created",
	"dc.date.issued", "dc.date.publication",
	"dcterms.created", "dcterms.date",
	"dcterms.issued", "dc:created", "dc:date",
	"gentime",
	// Open Graph: https://opengraphprotocol.org/
	"og:published_time", "og:article:published_time",
	"originalpublicationdate", "parsely-pub-date",
	"pubdate", "publishdate", "publish_date",
	"published-date", "publication_date", "rnews:datepublished",
	"sailthru.date", "shareaholic:article_published_time", "timestamp")

var propertyModified = sliceToMap(
	"article:modified_time", "datemodified", "modified_time",
	"og:article:modified_time", "og:updated_time", "og:modified_time",
	"release_date", "updated_time")

var (
	modifiedAttrKeys = []string{"lastmodified", "last-modified", "lastmod"}
	classAttrKeys    = []string{"published", "date-published", "time-published"}
	itemPropAttrKeys = []string{"datecreated", "datepublished", "pubyear", "datemodified", "dateupdate"}
	itemPropOriginal = itemPropAttrKeys[:3]
	itemPropModified = itemPropAttrKeys[3:]
)

const (
	slowPrependXpath = `.//*`
	fastPrependXpath = `.//*[(self::b or self::div or self::em or self::font or self::i or self::li
		or self::p or self::span or self::strong)]`

	dateXpath = `
		[contains(translate(@id, "D", "d"), 'date')
		or contains(translate(@class, "D", "d"), 'date')
		or contains(translate(@itemprop, "D", "d"), 'date')
		or contains(translate(@id, "D", "d"), 'datum')
		or contains(translate(@class, "D", "d"), 'datum')
		or contains(@id, 'time') or contains(@class, 'time')
		or @class='meta' or contains(translate(@id, "M", "m"), 'metadata')
		or contains(translate(@class, "M", "m"), 'meta-')
		or contains(translate(@class, "M", "m"), '-meta')
		or contains(translate(@id, "M", "m"), '-meta')
		or contains(translate(@class, "M", "m"), '_meta')
		or contains(translate(@class, "M", "m"), 'postmeta')
		or contains(@class, 'info') or contains(@class, 'post_detail')
		or contains(@class, 'block-content')
		or contains(@class, 'byline') or contains(@class, 'subline')
		or contains(@class, 'posted') or contains(@class, 'submitted')
		or contains(@class, 'created-post')
		or contains(@id, 'publish') or contains(@class, 'publish')
		or contains(@class, 'publication')
		or contains(@class, 'author') or contains(@class, 'autor')
		or contains(@class, 'field-content')
		or contains(@class, 'fa-clock-o') or contains(@class, 'fa-calendar')
		or contains(@class, 'fecha') or contains(@class, 'parution')
		or contains(@class, 'footer') or contains(@id, 'footer')]
		|
		.//footer|.//small`
	// Further tests needed:
	// or contains(@class, 'article')
	// or contains(@class, 'footer') or contains(@id, 'footer')
	// or contains(@id, 'lastmod') or contains(@class, 'updated')

	// archive.org banner
	discardXpath = `.//div[@id="wm-ipp-base" or @id="wm-ipp"]`
	// .//footer
	// .//*[(self::div or self::section)][@id="footer" or @class="footer"]

	freeTextXpath = fastPrependXpath + "/text()"
)

func sliceToMap(strings ...string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, s := range strings {
		result[s] = struct{}{}
	}
	return result
}
