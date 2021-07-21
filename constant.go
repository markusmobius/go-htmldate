// This file is part of go-htmldate, Go package for extracting publication dates from a web page.
// Source available in <https://github.com/markusmobius/go-trafilatura>.
// Copyright (C) 2021 Markus Mobius
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
)

var (
	timeZero       = time.Time{}
	defaultMinDate = time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC)
	defaultMaxDate = time.Now().AddDate(1, 0, 0)
)

const (
	maxPossibleCandidates = 300
	defaultDateFormat     = "2006-1-2"
)

var (
	rxLastNonDigits = regexp.MustCompile(`\D+$`)

	rxYmdNoSepPattern = regexp.MustCompile(`(?:\D|^)(\d{8})(?:\D|$)`)
	rxYmdPattern      = regexp.MustCompile(`(?:\D|^)(\d{4})[\-/.](\d{1,2})[\-/.](\d{1,2})(?:\D|$)`)
	rxDmyPattern      = regexp.MustCompile(`(?:\D|^)(\d{1,2})[\-/.](\d{1,2})[\-/.](\d{2,4})(?:\D|$)`)
	rxYmPattern       = regexp.MustCompile(`(?:\D|^)(\d{4})[\-/.](\d{1,2})(?:\D|$)`)
	rxMyPattern       = regexp.MustCompile(`(?:\D|^)(\d{1,2})[\-/.](\d{4})(?:\D|$)`)

	rxLongMdyPattern = regexp.MustCompile(`(` +
		`January|February|March|April|May|June|July|August|September|October|November|December|` +
		`Januari|Februari|Maret|Mei|Juni|Juli|Agustus|Oktober|Desember|` +
		`Jan|Feb|Mar|Apr|Jun|Jul|Aug|Sep|Oct|Nov|Dec|` +
		`Januar|Jänner|Februar|Feber|März|Mai|Dezember|` +
		`janvier|février|mars|avril|mai|juin|juillet|aout|septembre|octobre|novembre|décembre|` +
		`Ocak|Şubat|Mart|Nisan|Mayıs|Haziran|Temmuz|Ağustos|Eylül|Ekim|Kasım|Aralık|` +
		`Oca|Şub|Mar|Nis|Haz|Tem|Ağu|Eyl|Eki|Kas|Ara) ` +
		`([0-9]{1,2})(?:st|nd|rd|th)?,? ([0-9]{4})`)

	rxLongDmyPattern = regexp.MustCompile(`([0-9]{1,2})(?:st|nd|rd|th)? (?:of )?(` +
		`January|February|March|April|May|June|July|August|September|October|November|December|` +
		`Januari|Februari|Maret|Mei|Juni|Juli|Agustus|Oktober|Desember|` +
		`Jan|Feb|Mar|Apr|Jun|Jul|Aug|Sep|Oct|Nov|Dec|` +
		`Januar|Jänner|Februar|Feber|März|Mai|Dezember|` +
		`janvier|février|mars|avril|mai|juin|juillet|aout|septembre|octobre|novembre|décembre|` +
		`Ocak|Şubat|Mart|Nisan|Mayıs|Haziran|Temmuz|Ağustos|Eylül|Ekim|Kasım|Aralık|` +
		`Oca|Şub|Mar|Nis|Haz|Tem|Ağu|Eyl|Eki|Kas|Ara),? ` +
		`([0-9]{4})`)

	rxCompleteUrl = regexp.MustCompile(`(?i)([0-9]{4})[/-]([0-9]{1,2})[/-]([0-9]{1,2})`)
	rxPartialUrl  = regexp.MustCompile(`(?i)/([0-9]{4})/([0-9]{1,2})/`)

	rxGermanTextSearch = regexp.MustCompile(`(?i)` +
		`([0-9]{1,2})\.? (Januar|Jänner|Februar|Feber|März|April|` +
		`Mai|Juni|Juli|August|September|Oktober|November|Dezember) ` +
		`([0-9]{4})`)

	rxTimestampPattern  = regexp.MustCompile(`(?i)([0-9]{4}-[0-9]{2}-[0-9]{2}|[0-9]{2}\.[0-9]{2}\.[0-9]{4}).[0-9]{2}:[0-9]{2}:[0-9]{2}`)
	rxTextDatePattern   = regexp.MustCompile(`(?i)[.:,_/ -]|^[0-9]+$`)
	rxNoTextDatePattern = regexp.MustCompile(`(?i)^(?:[0-9]{3,}\D+[0-9]{3,}|[0-9]{2}:[0-9]{2}(:| )|\D*[0-9]{4}\D*$)`)

	rxEnPattern = regexp.MustCompile(`(?i)(?:[Dd]ate[^0-9"]{0,20}|updated|published) *?(?:in)? *?:? *?([0-9]{1,4})[./]([0-9]{1,2})[./]([0-9]{2,4})`)
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
)

// English + German + Turkish months cache
var monthNumber = map[string]int{
	"Januar": 1, "Jänner": 1, "January": 1, "Januari": 1, "Jan": 1, "Ocak": 1, "Oca": 1, "janvier": 1,
	"Februar": 2, "Feber": 2, "February": 2, "Februari": 2, "Feb": 2, "Şubat": 2, "Şub": 2, "février": 2,
	"März": 3, "March": 3, "Maret": 3, "Mar": 3, "Mart": 3, "mars": 3,
	"April": 4, "Apr": 4, "Nisan": 4, "Nis": 4, "avril": 4,
	"Mai": 5, "May": 5, "Mei": 5, "Mayıs": 5, "mai": 5,
	"Juni": 6, "June": 6, "Jun": 6, "Haziran": 6, "Haz": 6, "juin": 6,
	"Juli": 7, "July": 7, "Jul": 7, "Temmuz": 7, "Tem": 7, "juillet": 7,
	"August": 8, "Agustus": 8, "Aug": 8, "Ağustos": 8, "Ağu": 8, "aout": 8,
	"September": 9, "Sep": 9, "Eylül": 9, "Eyl": 9, "septembre": 9,
	"Oktober": 10, "October": 10, "Oct": 10, "Ekim": 10, "Eki": 10, "octobre": 10,
	"November": 11, "Nov": 11, "Kasım": 11, "Kas": 11, "novembre": 11,
	"Dezember": 12, "December": 12, "Desember": 12, "Dec": 12, "Aralık": 12, "Ara": 12, "décembre": 12,
}

var dateAttributes = sliceToMap(
	"article.created", "article_date_original",
	"article.published", "created", "cxenseparse:recs:publishtime",
	"date", "date_published", "dc.date", "dc.date.created",
	"dc.date.issued", "dcterms.date", "gentime",
	"originalpublicationdate", "parsely-pub-date",
	"pubdate", "publishdate", "publish_date",
	"published-date", "publication_date", "sailthru.date",
	"timestamp",
	"article:published_time", "bt:pubdate", "datecreated",
	"dateposted", "datepublished", "dc:created", "dc:date",
	"og:article:published_time", "og:published_time", "og:pubdate",
	"sailthru.date", "rnews:datepublished")

var propertyModified = sliceToMap(
	"article:modified_time", "datemodified", "modified_time",
	"og:article:modified_time", "og:updated_time", "og:modified_time",
	"release_date", "updated_time")

func sliceToMap(strings ...string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, s := range strings {
		result[s] = struct{}{}
	}
	return result
}
