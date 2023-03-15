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
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extractPartialUrlDate(t *testing.T) {
	opts := Options{
		MinDate: defaultMinDate,
		MaxDate: defaultMaxDate,
	}

	extract := func(s string) string {
		dt := extractPartialUrlDate(s, opts)
		if !dt.IsZero() {
			return dt.Format("2006-01-02")
		}
		return ""
	}

	assert.Equal(t, "2018-01-01", extract("https://testsite.org/2018/01/test"))
	assert.Equal(t, "", extract("https://testsite.org/2018/33/test"))
}

func Test_tryYmdDate(t *testing.T) {
	// Helper function
	opts := Options{
		MinDate: defaultMinDate,
		MaxDate: defaultMaxDate,
	}

	try := func(s string) string {
		_, dt := tryYmdDate(s, opts)
		if !dt.IsZero() {
			return dt.Format("2006-01-02")
		}
		return ""
	}

	// Extensive search disabled
	opts.SkipExtensiveSearch = true
	assert.Equal(t, "", try("Fri, Sept 1, 2017"))

	// Extensive search enabled
	opts.SkipExtensiveSearch = false
	assert.Equal(t, "2017-09-01", try("Friday, September 01, 2017"))
	assert.Equal(t, "2017-09-01", try("Fr, 1 Sep 2017 16:27:51 MESZ"))
	assert.Equal(t, "2017-09-01", try("Freitag, 01. September 2017"))
	assert.Equal(t, "2017-09-01", try("Am 1. September 2017 um 15:36 Uhr schrieb"))
	assert.Equal(t, "2017-09-01", try("Fri - September 1 - 2017"))
	assert.Equal(t, "2017-09-01", try("1.9.2017"))
	assert.Equal(t, "2017-09-01", try("1/9/17"))
	assert.Equal(t, "2017-09-01", try("201709011234"))

	// Wrong date
	assert.Equal(t, "", try("201"))
	assert.Equal(t, "", try("14:35:10"))
	assert.Equal(t, "", try("12:00 h"))
	assert.Equal(t, "", try("2005-2006"))
}

func Test_fastParse(t *testing.T) {
	opts := Options{
		MinDate:   defaultMinDate,
		MaxDate:   defaultMaxDate,
		EnableLog: true,
	}

	parse := func(s string) string {
		dt := fastParse(s, opts)
		if !dt.IsZero() {
			return dt.Format("2006-01-02")
		}
		return ""
	}

	assert.Equal(t, "2004-12-12", parse("20041212"))
	assert.Equal(t, "2004-12-12", parse("12.12.2004"))
	assert.Equal(t, "2004-12-12", parse("2004-12-12"))
	assert.Equal(t, "2004-01-12", parse("12.01.2004"))
	assert.Equal(t, "2020-01-12", parse("12.01.20"))
	assert.Equal(t, "2016-03-14", parse("3/14/2016"))
	assert.Equal(t, "2020-01-01", parse("2020-1"))
	assert.Equal(t, "2020-01-01", parse("2020.01"))
	assert.Equal(t, "1998-01-01", parse("1998-01"))
	assert.Equal(t, "1998-10-10", parse("10.10.98"))
	assert.Equal(t, "2004-12-12", parse("abcd 20041212 efgh"))
	assert.Equal(t, "2004-02-12", parse("abcd 2004-2-12 efgh"))
	assert.Equal(t, "2004-02-01", parse("abcd 2004-2 efgh"))
	assert.Equal(t, "2004-02-01", parse("abcd 2004-2 efgh"))
	assert.Equal(t, "", parse("2020.13"))
	assert.Equal(t, "", parse("12122004"))
	assert.Equal(t, "", parse("1212-20-04"))
	assert.Equal(t, "", parse("33.20.2004"))
	assert.Equal(t, "", parse("36/14/2016"))
	assert.Equal(t, "", parse("2019 28 meh"))
	assert.Equal(t, "", parse("January 12 1098"))
	assert.Equal(t, "", parse("abcd 32. Januar 2020 efgh"))
}

func Test_regexParse(t *testing.T) {
	opts := Options{
		MinDate: defaultMinDate,
		MaxDate: defaultMaxDate,
	}

	parse := func(s string) string {
		dt := regexParse(s, opts)
		if !dt.IsZero() {
			return dt.Format("2006-01-02")
		}
		return ""
	}

	assert.Equal(t, "2008-12-03", parse("3. Dezember 2008"))
	assert.Equal(t, "", parse("33. Dezember 2008"))
	assert.Equal(t, "2008-12-03", parse("3 Aralık 2008 Çarşamba"))
	assert.Equal(t, "2008-12-03", parse("3 Aralık 2008"))
	assert.Equal(t, "2019-03-26", parse("Tuesday, March 26th, 2019"))
	assert.Equal(t, "2019-03-26", parse("March 26, 2019"))
	assert.Equal(t, "", parse("3rd Tuesday in March"))
	assert.Equal(t, "2019-03-26", parse("Mart 26, 2019"))
	assert.Equal(t, "2019-03-26", parse("Salı, Mart 26, 2019"))
	assert.Equal(t, "", parse("36/14/2016"))
	assert.Equal(t, "", parse("January 36 1998"))
	assert.Equal(t, "1998-01-01", parse("January 1st, 1998"))
	assert.Equal(t, "1998-02-01", parse("February 1st, 1998"))
	assert.Equal(t, "1998-03-01", parse("March 1st, 1998"))
	assert.Equal(t, "1998-04-01", parse("April 1st, 1998"))
	assert.Equal(t, "1998-05-01", parse("May 1st, 1998"))
	assert.Equal(t, "1998-06-01", parse("June 1st, 1998"))
	assert.Equal(t, "1998-07-01", parse("July 1st, 1998"))
	assert.Equal(t, "1998-08-01", parse("August 1st, 1998"))
	assert.Equal(t, "1998-09-01", parse("September 1st, 1998"))
	assert.Equal(t, "1998-10-01", parse("October 1st, 1998"))
	assert.Equal(t, "1998-11-01", parse("November 1st, 1998"))
	assert.Equal(t, "1998-12-01", parse("December 1st, 1998"))
	assert.Equal(t, "1998-01-01", parse("Jan 1st, 1998"))
	assert.Equal(t, "1998-02-01", parse("Feb 1st, 1998"))
	assert.Equal(t, "1998-03-01", parse("Mar 1st, 1998"))
	assert.Equal(t, "1998-04-01", parse("Apr 1st, 1998"))
	assert.Equal(t, "1998-06-01", parse("Jun 1st, 1998"))
	assert.Equal(t, "1998-07-01", parse("Jul 1st, 1998"))
	assert.Equal(t, "1998-08-01", parse("Aug 1st, 1998"))
	assert.Equal(t, "1998-09-01", parse("Sep 1st, 1998"))
	assert.Equal(t, "1998-10-01", parse("Oct 1st, 1998"))
	assert.Equal(t, "1998-11-01", parse("Nov 1st, 1998"))
	assert.Equal(t, "1998-12-01", parse("Dec 1st, 1998"))
	assert.Equal(t, "1998-01-01", parse("Januar 1, 1998"))
	assert.Equal(t, "1998-01-01", parse("Jänner 1, 1998"))
	assert.Equal(t, "1998-02-01", parse("Februar 1, 1998"))
	assert.Equal(t, "1998-02-01", parse("Feber 1, 1998"))
	assert.Equal(t, "1998-03-01", parse("März 1, 1998"))
	assert.Equal(t, "1998-04-01", parse("April 1, 1998"))
	assert.Equal(t, "1998-05-01", parse("Mai 1, 1998"))
	assert.Equal(t, "1998-06-01", parse("Juni 1, 1998"))
	assert.Equal(t, "1998-07-01", parse("Juli 1, 1998"))
	assert.Equal(t, "1998-08-01", parse("August 1, 1998"))
	assert.Equal(t, "1998-09-01", parse("September 1, 1998"))
	assert.Equal(t, "1998-10-01", parse("Oktober 1, 1998"))
	assert.Equal(t, "1998-11-01", parse("November 1, 1998"))
	assert.Equal(t, "1998-12-01", parse("Dezember 1, 1998"))
	assert.Equal(t, "1998-01-01", parse("Ocak 1, 1998"))
	assert.Equal(t, "1998-02-01", parse("Şubat 1, 1998"))
	assert.Equal(t, "1998-03-01", parse("Mart 1, 1998"))
	assert.Equal(t, "1998-04-01", parse("Nisan 1, 1998"))
	assert.Equal(t, "1998-05-01", parse("Mayıs 1, 1998"))
	assert.Equal(t, "1998-06-01", parse("Haziran 1, 1998"))
	assert.Equal(t, "1998-07-01", parse("Temmuz 1, 1998"))
	assert.Equal(t, "1998-08-01", parse("Ağustos 1, 1998"))
	assert.Equal(t, "1998-09-01", parse("Eylül 1, 1998"))
	assert.Equal(t, "1998-10-01", parse("Ekim 1, 1998"))
	assert.Equal(t, "1998-11-01", parse("Kasım 1, 1998"))
	assert.Equal(t, "1998-12-01", parse("Aralık 1, 1998"))
	assert.Equal(t, "1998-01-01", parse("Oca 1, 1998"))
	assert.Equal(t, "1998-02-01", parse("Şub 1, 1998"))
	assert.Equal(t, "1998-03-01", parse("Mar 1, 1998"))
	assert.Equal(t, "1998-04-01", parse("Nis 1, 1998"))
	assert.Equal(t, "1998-05-01", parse("May 1, 1998"))
	assert.Equal(t, "1998-06-01", parse("Haz 1, 1998"))
	assert.Equal(t, "1998-07-01", parse("Tem 1, 1998"))
	assert.Equal(t, "1998-08-01", parse("Ağu 1, 1998"))
	assert.Equal(t, "1998-09-01", parse("Eyl 1, 1998"))
	assert.Equal(t, "1998-10-01", parse("Eki 1, 1998"))
	assert.Equal(t, "1998-11-01", parse("Kas 1, 1998"))
	assert.Equal(t, "1998-12-01", parse("Ara 1, 1998"))
	assert.Equal(t, "1998-01-01", parse("1 January 1998"))
	assert.Equal(t, "1998-02-01", parse("1 February 1998"))
	assert.Equal(t, "1998-03-01", parse("1 March 1998"))
	assert.Equal(t, "1998-04-01", parse("1 April 1998"))
	assert.Equal(t, "1998-05-01", parse("1 May 1998"))
	assert.Equal(t, "1998-06-01", parse("1 June 1998"))
	assert.Equal(t, "1998-07-01", parse("1 July 1998"))
	assert.Equal(t, "1998-08-01", parse("1 August 1998"))
	assert.Equal(t, "1998-09-01", parse("1 September 1998"))
	assert.Equal(t, "1998-10-01", parse("1 October 1998"))
	assert.Equal(t, "1998-11-01", parse("1 November 1998"))
	assert.Equal(t, "1998-12-01", parse("1 December 1998"))
	assert.Equal(t, "1998-01-01", parse("1 Jan 1998"))
	assert.Equal(t, "1998-02-01", parse("1 Feb 1998"))
	assert.Equal(t, "1998-03-01", parse("1 Mar 1998"))
	assert.Equal(t, "1998-04-01", parse("1 Apr 1998"))
	assert.Equal(t, "1998-06-01", parse("1 Jun 1998"))
	assert.Equal(t, "1998-07-01", parse("1 Jul 1998"))
	assert.Equal(t, "1998-08-01", parse("1 Aug 1998"))
	assert.Equal(t, "1998-09-01", parse("1 Sep 1998"))
	assert.Equal(t, "1998-10-01", parse("1 Oct 1998"))
	assert.Equal(t, "1998-11-01", parse("1 Nov 1998"))
	assert.Equal(t, "1998-12-01", parse("1 Dec 1998"))
	assert.Equal(t, "1998-01-01", parse("1 Januar 1998"))
	assert.Equal(t, "1998-01-01", parse("1 Jänner 1998"))
	assert.Equal(t, "1998-02-01", parse("1 Februar 1998"))
	assert.Equal(t, "1998-02-01", parse("1 Feber 1998"))
	assert.Equal(t, "1998-03-01", parse("1 März 1998"))
	assert.Equal(t, "1998-04-01", parse("1 April 1998"))
	assert.Equal(t, "1998-05-01", parse("1 Mai 1998"))
	assert.Equal(t, "1998-06-01", parse("1 Juni 1998"))
	assert.Equal(t, "1998-07-01", parse("1 Juli 1998"))
	assert.Equal(t, "1998-08-01", parse("1 August 1998"))
	assert.Equal(t, "1998-09-01", parse("1 September 1998"))
	assert.Equal(t, "1998-10-01", parse("1 Oktober 1998"))
	assert.Equal(t, "1998-11-01", parse("1 November 1998"))
	assert.Equal(t, "1998-12-01", parse("1 Dezember 1998"))
	assert.Equal(t, "1998-01-01", parse("1 Ocak 1998"))
	assert.Equal(t, "1998-02-01", parse("1 Şubat 1998"))
	assert.Equal(t, "1998-03-01", parse("1 Mart 1998"))
	assert.Equal(t, "1998-04-01", parse("1 Nisan 1998"))
	assert.Equal(t, "1998-05-01", parse("1 Mayıs 1998"))
	assert.Equal(t, "1998-06-01", parse("1 Haziran 1998"))
	assert.Equal(t, "1998-07-01", parse("1 Temmuz 1998"))
	assert.Equal(t, "1998-08-01", parse("1 Ağustos 1998"))
	assert.Equal(t, "1998-09-01", parse("1 Eylül 1998"))
	assert.Equal(t, "1998-10-01", parse("1 Ekim 1998"))
	assert.Equal(t, "1998-11-01", parse("1 Kasım 1998"))
	assert.Equal(t, "1998-12-01", parse("1 Aralık 1998"))
	assert.Equal(t, "1998-01-01", parse("1 Oca 1998"))
	assert.Equal(t, "1998-02-01", parse("1 Şub 1998"))
	assert.Equal(t, "1998-03-01", parse("1 Mar 1998"))
	assert.Equal(t, "1998-04-01", parse("1 Nis 1998"))
	assert.Equal(t, "1998-05-01", parse("1 May 1998"))
	assert.Equal(t, "1998-06-01", parse("1 Haz 1998"))
	assert.Equal(t, "1998-07-01", parse("1 Tem 1998"))
	assert.Equal(t, "1998-08-01", parse("1 Ağu 1998"))
	assert.Equal(t, "1998-09-01", parse("1 Eyl 1998"))
	assert.Equal(t, "1998-10-01", parse("1 Eki 1998"))
	assert.Equal(t, "1998-11-01", parse("1 Kas 1998"))
	assert.Equal(t, "1998-12-01", parse("1 Ara 1998"))
}
