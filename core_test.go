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
	"io"
	"testing"
	"time"

	"github.com/markusmobius/go-htmldate/internal/regexp"
	"github.com/stretchr/testify/assert"
)

func Test_HtmlDate(t *testing.T) {
	// Variables
	var str, url string
	useOriginalDate := Options{UseOriginalDate: true}
	skipExtensiveSearch := Options{SkipExtensiveSearch: true}

	// Helper function
	output := func(res Result) string {
		if !res.IsZero() {
			return res.Format("2006-01-02")
		}
		return ""
	}

	checkMockFile := func(url string, expected string, opts ...Options) {
		res := extractMockFile(url, opts...)
		assert.Equal(t, expected, output(res), "file "+url)
	}

	checkString := func(str string, expected string, opts ...Options) {
		res := extractFromString(str, opts...)
		assert.Equal(t, expected, output(res), "str "+str)
	}

	checkURL := func(url string, expected string, opts ...Options) {
		res := extractFromURL(url, opts...)
		assert.Equal(t, expected, output(res), "url "+url)
	}

	// ===================================================
	// Tests below these point should return an exact date
	// ===================================================

	// These pages shouldn't return any date
	url = "https://www.intel.com/content/www/us/en/legal/terms-of-use.html"
	checkMockFile(url, "")

	url = "https://en.support.wordpress.com/"
	checkMockFile(url, "")

	str = "<html><body>XYZ</body></html>"
	checkString(str, "")

	str = "<html><body><time></time></body></html>"
	checkString(str, "")

	str = `<html><body><abbr class="published"></abbr></body></html>`
	checkString(str, "")

	// HTML document tree
	str = `<html><head><meta property="dc:created" content="2017-09-01"></head><body><p>HELLO</p></body></html>`
	checkString(str, "2017-09-01", useOriginalDate)

	str = `<html><head><meta property="dc:created" content="2017-09-01"/></head><body></body></html>`
	checkString(str, "2017-09-01", useOriginalDate)

	str = `<html><head><meta property="og:published_time" content="2017-09-01"/></head><body></body></html>`
	checkString(str, "2017-09-01", useOriginalDate)

	str = `<html><head><meta name="last-modified" content="2017-09-01"/></head><body></body></html>`
	checkString(str, "2017-09-01")

	str = `<html><head><meta property="OG:Updated_Time" content="2017-09-01"/></head><body></body></html>`
	checkString(str, "2017-09-01")

	str = `<html><head><meta property="og:updated_time" content="2017-09-01"/></head><body></body></html>`
	checkString(str, "2017-09-01")

	str = `<html><head><meta property="og:regDate" content="20210820030646"></head><body></body></html>`
	checkString(str, "2021-08-20")

	str = `<html><head><meta name="created" content="2017-01-09"/></head><body></body></html>`
	checkString(str, "2017-01-09")

	str = `<html><head><meta name="citation_publication_date" content="2017-01-09"/></head><body></body></html>`
	checkString(str, "2017-01-09")

	str = `<html><head><meta itemprop="copyrightyear" content="2017"/></head><body></body></html>`
	checkString(str, "2017-01-01")

	// Original date
	str = `<html><head>
	<meta property="OG:Updated_Time" content="2017-09-01"/>
	<meta property="OG:DatePublished" content="2017-07-02"/>
	</head><body/></html>`
	checkString(str, "2017-09-01")
	checkString(str, "2017-07-02", useOriginalDate)

	str = `<html><head>
	<meta property="article:modified_time" content="2021-04-06T06:32:14+00:00" />
	<meta property="article:published_time" content="2020-07-21T00:17:28+00:00" />
	</head><body/></html>`
	checkString(str, "2021-04-06")
	checkString(str, "2020-07-21", useOriginalDate)

	str = `<html><head>
	<meta property="article:published_time" content="2020-07-21T00:17:28+00:00" />
	<meta property="article:modified_time" content="2021-04-06T06:32:14+00:00" />
	</head><body/></html>`
	checkString(str, "2021-04-06")
	checkString(str, "2020-07-21", useOriginalDate)

	// Link in header
	url = "http://www.jovelstefan.de/2012/05/11/parken-in-paris/"
	checkMockFile(url, "2012-05-11")

	// Meta in header
	str = `<html><head><meta/></head><body></body></html>`
	checkString(str, "")

	url = "https://500px.com/photo/26034451/spring-in-china-by-alexey-kruglov"
	checkMockFile(url, "2013-02-16")

	str = `<html><head><meta name="og:url" content="http://www.example.com/2018/02/01/entrytitle"/></head><body></body></html>`
	checkString(str, "2018-02-01")

	str = `<html><head><meta itemprop="datecreated" datetime="2018-02-02"/></head><body></body></html>`
	checkString(str, "2018-02-02")

	str = `<html><head><meta itemprop="datemodified" content="2018-02-04"/></head><body></body></html>`
	checkString(str, "2018-02-04")

	str = `<html><head><meta http-equiv="last-modified" content="2018-02-05"/></head><body></body></html>`
	checkString(str, "2018-02-05")

	str = `<html><head><meta name="lastmodified" content="2018-02-05"/></head><body></body></html>`
	checkString(str, "2018-02-05", useOriginalDate)

	str = `<html><head><meta name="lastmodified" content="2018-02-05"/></head><body></body></html>`
	checkString(str, "2018-02-05")

	str = `<html><head><meta http-equiv="date" content="2017-09-01"/></head><body></body></html>`
	checkString(str, "2017-09-01", useOriginalDate)

	str = `<html><head><meta http-equiv="last-modified" content="2018-10-01"/><meta http-equiv="date" content="2017-09-01"/></head><body></body></html>`
	checkString(str, "2017-09-01", useOriginalDate)

	str = `<html><head><meta http-equiv="last-modified" content="2018-10-01"/><meta http-equiv="date" content="2017-09-01"/></head><body></body></html>`
	checkString(str, "2018-10-01")

	str = `<html><head><meta http-equiv="date" content="2017-09-01"/><meta http-equiv="last-modified" content="2018-10-01"/></head><body></body></html>`
	checkString(str, "2017-09-01", useOriginalDate)

	str = `<html><head><meta http-equiv="date" content="2017-09-01"/><meta http-equiv="last-modified" content="2018-10-01"/></head><body></body></html>`
	checkString(str, "2018-10-01")

	str = `<html><head><meta name="Publish_Date" content="02.02.2004"/></head><body></body></html>`
	checkString(str, "2004-02-02")

	str = `<html><head><meta name="pubDate" content="2018-02-06"/></head><body></body></html>`
	checkString(str, "2018-02-06")

	str = `<html><head><meta pubdate="pubDate" content="2018-02-06"/></head><body></body></html>`
	checkString(str, "2018-02-06")

	str = `<html><head><meta itemprop="DateModified" datetime="2018-02-06"/></head><body></body></html>`
	checkString(str, "2018-02-06")

	str = `<html><head><meta name="DC.Issued" content="2021-07-13"/></head><body></body></html>`
	checkString(str, "2021-07-13")

	str = `<html><head><meta itemprop="dateUpdate" datetime="2018-02-06"/></head><body></body></html>`
	checkString(str, "2018-02-06", useOriginalDate)

	str = `<html><head><meta itemprop="dateUpdate" datetime="2018-02-06"/></head><body></body></html>`
	checkString(str, "2018-02-06")

	// Time in document body
	url = "https://www.facebook.com/visitaustria/"
	checkMockFile(url, "2017-10-08")

	checkMockFile(url, "2017-10-06", useOriginalDate)

	url = "http://www.medef.com/en/content/alternative-dispute-resolution-for-antitrust-damages"
	checkMockFile(url, "2017-09-01")

	str = `<html><body><time datetime="08:00"></body></html>`
	checkString(str, "")

	str = `<html><body><time datetime="2014-07-10 08:30:45.687"></body></html>`
	checkString(str, "2014-07-10")

	str = `<html><head></head><body><time class="entry-time" itemprop="datePublished" datetime="2018-04-18T09:57:38+00:00"></body></html>`
	checkString(str, "2018-04-18")

	str = `<html><body><footer class="article-footer"><p class="byline">Veröffentlicht am <time class="updated" datetime="2019-01-03T14:56:51+00:00">3. Januar 2019 um 14:56 Uhr.</time></p></footer></body></html>`
	checkString(str, "2019-01-03")

	str = `<html><body><footer class="article-footer"><p class="byline">Veröffentlicht am <time class="updated" datetime="2019-01-03T14:56:51+00:00"></time></p></footer></body></html>`
	checkString(str, "2019-01-03")

	str = `<html><body><time datetime="2011-09-27" class="entry-date"></time><time datetime="2011-09-28" class="updated"></time></body></html>`
	checkString(str, "2011-09-27", useOriginalDate)

	// Updated vs original in time elements
	str = `<html><body><time datetime="2011-09-27" class="entry-date"></time><time datetime="2011-09-28" class="updated"></time></body></html>`
	checkString(str, "2011-09-28")

	str = `<html><body><time datetime="2011-09-28" class="updated"></time><time datetime="2011-09-27" class="entry-date"></time></body></html>`
	checkString(str, "2011-09-27", useOriginalDate)

	str = `<html><body><time datetime="2011-09-28" class="updated"></time><time datetime="2011-09-27" class="entry-date"></time></body></html>`
	checkString(str, "2011-09-28")

	// Removed from HTML5 https://www.w3schools.com/TAgs/att_time_datetime_pubdate.asp
	str = `<html><body><time datetime="2011-09-28" pubdate="pubdate"></time></body></html>`
	checkString(str, "2011-09-28")
	checkString(str, "2011-09-28", useOriginalDate)

	str = `<html><body><time datetime="2011-09-28" class="entry-date"></time></body></html>`
	checkString(str, "2011-09-28")

	// Bug #54 in original Python library
	// Their issues doesn't really affect us since our dateparser are different
	str = `<html><body><time class="Feed-module--feed__item-meta-time--3t1fg" dateTime="November 29, 2020">November 2020</time></body></html>`
	checkString(str, "2020-11-29")

	// Precise pattern in document body
	str = `<html><body><font size="2" face="Arial,Geneva,Helvetica">Bei <a href="../../sonstiges/anfrage.php"><b>Bestellungen</b></a> bitte Angabe der Titelnummer nicht vergessen!<br><br>Stand: 03.04.2019</font></body></html>`
	checkString(str, "2019-04-03")

	str = `<html><body><div>Erstausstrahlung: 01.01.2020</div><div>Preisstand: 03.02.2022 03:00 GMT+1</div></body></html>`
	checkString(str, "2022-02-03")

	str = `<html><body>Datum: 10.11.2017</body></html>`
	checkString(str, "2017-11-10")

	url = `https://www.tagesausblick.de/Analyse/USA/DOW-Jones-Jahresendrally-ade__601.html`
	checkMockFile(url, "2012-12-22")

	url = `http://blog.todamax.net/2018/midp-emulator-kemulator-und-brick-challenge/`
	checkMockFile(url, "2018-02-15")

	// JSON date published
	url = `https://www.acredis.com/schoenheitsoperationen/augenlidstraffung/`
	checkMockFile(url, "2018-02-28", useOriginalDate)

	// JSON date modified
	url = `https://www.channelpartner.de/a/sieben-berufe-die-zukunft-haben,3050673`
	checkMockFile(url, "2019-04-03")

	// Meta in document body
	url = "https://futurezone.at/digital-life/wie-creativecommons-richtig-genutzt-wird/24.600.504"
	checkMockFile(url, "2013-08-09", useOriginalDate)

	url = "https://www.horizont.net/marketing/kommentare/influencer-marketing-was-sich-nach-dem-vreni-frost-urteil-aendert-und-aendern-muss-172529"
	checkMockFile(url, "2019-01-29")

	url = "http://www.klimawandel-global.de/klimaschutz/energie-sparen/elektromobilitat-der-neue-trend/"
	checkMockFile(url, "2013-05-03")

	url = "http://www.hobby-werkstatt-blog.de/arduino/424-eine-arduino-virtual-wall-fuer-den-irobot-roomba.php"
	checkMockFile(url, "2015-12-14")

	url = "https://www.beltz.de/fachmedien/paedagogik/didacta_2019_in_koeln_19_23_februar/beltz_veranstaltungen_didacta_2016/veranstaltung.html?tx_news_pi1%5Bnews%5D=14392&tx_news_pi1%5Bcontroller%5D=News&tx_news_pi1%5Baction%5D=detail&cHash=10b1a32fb5b2b05360bdac257b01c8fa"
	checkMockFile(url, "2019-02-20")

	url = "https://www.wienbadminton.at/news/119843/Come-Together"
	checkMockFile(url, "", skipExtensiveSearch)

	url = "https://www.wienbadminton.at/news/119843/Come-Together"
	checkMockFile(url, "2018-05-06")

	// Abbr in document body
	url = "http://blog.kinra.de/?p=959/"
	checkMockFile(url, "2012-12-16")

	str = `<html><body><abbr class="published">am 12.11.16</abbr></body></html>`
	checkString(str, "2016-11-12")

	str = `<html><body><abbr class="published">am 12.11.16</abbr></body></html>`
	checkString(str, "2016-11-12", useOriginalDate)

	str = `<html><body><abbr class="published" title="2016-11-12">XYZ</abbr></body></html>`
	checkString(str, "2016-11-12", useOriginalDate)

	str = `<html><body><abbr class="date-published">8.11.2016</abbr></body></html>`
	checkString(str, "2016-11-08")

	// Valid vs invalid data-utime
	str = `<html><body><abbr data-utime="1438091078" class="something">A date</abbr></body></html>`
	checkString(str, "2015-07-28")

	str = `<html><body><abbr data-utime="143809-1078" class="something">A date</abbr></body></html>`
	checkString(str, "")

	// Time in document body
	str = `<html><body><time>2018-01-04</time></body></html>`
	checkString(str, "2018-01-04")

	url = "https://www.adac.de/rund-ums-fahrzeug/tests/kindersicherheit/kindersitztest-2018/"
	checkMockFile(url, "2018-10-23")

	// Additional selector rules
	str = `<html><body><div class="fecha">2018-01-04</div></body></html>`
	checkString(str, "2018-01-04")

	// Other expressions in document body
	str = `<html><body>"datePublished":"2018-01-04"</body></html>`
	checkString(str, "2018-01-04")

	str = `<html><body>Stand: 1.4.18</body></html>`
	checkString(str, "2018-04-01")

	url = "http://www.stuttgart.de/"
	checkMockFile(url, "2017-10-09")

	// In document body
	url = "https://github.com/adbar/htmldate"
	checkMockFile(url, "2017-11-28") // was '2019-01-01'

	url = "https://github.com/adbar/htmldate"
	checkMockFile(url, "2016-07-12", useOriginalDate)

	url = "https://en.blog.wordpress.com/"
	checkMockFile(url, "2017-08-30")

	url = "https://www.austria.info/"
	checkMockFile(url, "2017-09-07")

	url = "https://www.eff.org/files/annual-report/2015/index.html"
	checkMockFile(url, "2016-05-04")

	url = "http://unexpecteduser.blogspot.de/2011/"
	checkMockFile(url, "2011-03-30")

	url = "https://die-partei.net/sh/"
	checkMockFile(url, "2014-07-19")

	url = "https://www.rosneft.com/business/Upstream/Licensing/"
	checkMockFile(url, "2017-02-27") // most probably 2014-12-31, found in text

	url = "http://www.freundeskreis-videoclips.de/waehlen-sie-car-player-tipps-zur-auswahl-der-besten-car-cd-player/"
	checkMockFile(url, "2017-07-12")

	url = "https://www.scs78.de/news/items/warm-war-es-schoen-war-es.html"
	checkMockFile(url, "2018-06-10")

	url = "https://www.goodform.ch/blog/schattiges_plaetzchen"
	checkMockFile(url, "2018-06-27")

	url = "https://www.transgen.de/aktuell/2687.afrikanische-schweinepest-genome-editing.html"
	checkMockFile(url, "2018-01-18")

	url = "http://www.eza.gv.at/das-ministerium/presse/aussendungen/2018/07/aussenministerin-karin-kneissl-beim-treffen-der-deutschsprachigen-aussenminister-in-luxemburg/"
	checkMockFile(url, "2018-07-03")

	url = "https://www.weltwoche.ch/ausgaben/2019-4/artikel/forbes-die-weltwoche-ausgabe-4-2019.html"
	checkMockFile(url, "2019-01-23")

	// Free text
	str = `<html><body>&copy; 2017</body></html>`
	checkString(str, "2017-01-01")

	str = `<html><body>© 2017</body></html>`
	checkString(str, "2017-01-01")

	str = `<html><body><p>Dieses Datum ist leider ungültig: 30. Februar 2018.</p></body></html>`
	checkString(str, "", skipExtensiveSearch)

	str = `<html><body><p>Dieses Datum ist leider ungültig: 30. Februar 2018.</p></body></html>`
	checkString(str, "2018-01-01")

	url = "http://unexpecteduser.blogspot.de/2011/"
	checkMockFile(url, "2011-03-30")

	url = "http://blog.python.org/2016/12/python-360-is-now-available.html"
	checkMockFile(url, "2016-12-23")

	// Additional list
	url = "http://carta.info/der-neue-trend-muss-statt-wunschkoalition/"
	checkMockFile(url, "2012-05-08")

	url = "https://www.wunderweib.de/manuela-reimann-hochzeitsueberraschung-in-bayern-107930.html"
	checkMockFile(url, "2019-06-20")

	url = "https://www.befifty.de/home/2017/7/12/unter-uns-montauk"
	checkMockFile(url, "2017-07-12")

	url = "https://www.brigitte.de/aktuell/riverdale--so-ehrt-die-serie-luke-perry-in-staffel-vier-11602344.html"
	checkMockFile(url, "2019-06-20")

	url = "http://www.loldf.org/spip.php?article717"
	checkMockFile(url, "2019-06-27")

	url = "https://www.beltz.de/sachbuch_ratgeber/buecher/produkt_produktdetails/37219-12_wege_zu_guter_pflege.html"
	checkMockFile(url, "2019-02-07")

	url = "https://www.oberstdorf-resort.de/interaktiv/blog/unser-kraeutergarten-wannenkopfhuette.html"
	checkMockFile(url, "2018-06-20")

	url = "https://www.wienbadminton.at/news/119843/Come-Together"
	checkMockFile(url, "2018-05-06")

	url = "https://www.ldt.de/ldtblog/fall-in-love-with-black/"
	checkMockFile(url, "2017-08-08")

	url = "https://paris-luttes.info/quand-on-comprend-que-les-grenades-12355"
	checkMockFile(url, "2019-06-29") // here we are better than the original

	url = "https://verfassungsblog.de/the-first-decade/"
	checkMockFile(url, "2019-07-13")

	url = "https://cric-grenoble.info/infos-locales/article/putsh-en-cours-a-radio-kaleidoscope-1145"
	checkMockFile(url, "2019-06-09")

	url = "https://www.sebastian-kurz.at/magazin/wasserstoff-als-schluesseltechnologie"
	checkMockFile(url, "2019-07-30")

	url = "https://exporo.de/wiki/europaeische-zentralbank-ezb/"
	checkMockFile(url, "2018-01-01", useOriginalDate)

	// Only found by extensive search
	url = "https://ebene11.com/die-arbeit-mit-fremden-dwg-dateien-in-autocad"
	checkMockFile(url, "", skipExtensiveSearch)

	url = "https://ebene11.com/die-arbeit-mit-fremden-dwg-dateien-in-autocad"
	checkMockFile(url, "2017-01-12")

	url = "https://www.hertie-school.org/en/debate/detail/content/whats-on-the-cards-for-von-der-leyen/"
	checkMockFile(url, "", skipExtensiveSearch)

	url = "https://www.hertie-school.org/en/debate/detail/content/whats-on-the-cards-for-von-der-leyen/"
	checkMockFile(url, "2019-12-02") // Or maybe 2019-02-12

	// Date not in footer but at the start of the article
	url = "http://www.wara-enforcement.org/guinee-un-braconnier-delephant-interpelle-et-condamne-a-la-peine-maximale/"
	checkMockFile(url, "2016-09-27")

	// URL from meta image in header
	str = `<html><meta property="og:image" content="https://example.org/img/2019-05-05/test.jpg"><body></body></html>`
	checkString(str, "2019-05-05")

	str = `<html><meta property="og:image" content="https://example.org/img/test.jpg"><body></body></html>`
	checkString(str, "")

	// URL from <img> in body
	str = `<html><body><img src="https://example.org/img/2019-05-05/test.jpg"/></body></html>`
	checkString(str, "2019-05-05")

	str = `<html><body><img src="https://example.org/img/test.jpg"/></body></html>`
	checkString(str, "")

	str = `<html><body><img src="https://example.org/img/2019-05-03/test.jpg"/><img src="https://example.org/img/2019-05-04/test.jpg"/><img src="https://example.org/img/2019-05-05/test.jpg"/></body></html>`
	checkString(str, "2019-05-05")

	str = `<html><body><img src="https://example.org/img/2019-05-05/test.jpg"/><img src="https://example.org/img/2019-05-04/test.jpg"/><img src="https://example.org/img/2019-05-03/test.jpg"/></body></html>`
	checkString(str, "2019-05-05")

	// Title content
	str = `<html><head><title>Bericht zur Coronalage vom 22.04.2020 – worauf wartet die Politik? – DIE ACHSE DES GUTEN. ACHGUT.COM</title></head></html>`
	checkString(str, "2020-04-22")

	// In unknown div
	str = `<html><body><div class="spip spip-block-right" style="text-align:right;">Le 26 juin 2019</div></body></html>`
	checkString(str, "", skipExtensiveSearch)

	str = `<html><body><div class="spip spip-block-right" style="text-align:right;">Le 26 juin 2019</div></body></html>`
	checkString(str, "2019-06-26")

	// In link title
	str = `<html><body><a class="ribbon date " title="12th December 2018" href="https://example.org/" itemprop="url">Text</a></body></html>`
	checkString(str, "2018-12-12")

	// Archive.org documents
	url = "http://web.archive.org/web/20210916140120/https://www.kath.ch/die-insel-der-klosterzoeglinge/"
	checkMockFile(url, "", skipExtensiveSearch)
	checkMockFile(url, "2021-07-13")

	// Min date
	str = `<html><meta><meta property="article:published_time" content="1991-01-02T01:01:00+01:00"></meta><body></body></html>`
	checkString(str, "", Options{MinDate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)})
	checkString(str, "1991-01-02", Options{MinDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)})

	// Wild text in body
	str = `<html><body>Wed, 19 Oct 2022 14:24:05 +0000</body></html>`
	checkString(str, "2022-10-19")

	// =========================================================
	// Tests below these point should return an approximate date
	// =========================================================

	// Copyright text
	url = "http://viehbacher.com/de/spezialisierung/internationale-forderungsbeitreibung"
	checkMockFile(url, "2016-01-01") // somewhere in 2016

	// Other
	url = "https://creativecommons.org/about/"
	checkMockFile(url, "2017-08-11") // or '2017-08-03'

	url = "https://creativecommons.org/about/"
	checkMockFile(url, "2016-05-22", useOriginalDate) // or '2017-08-03'

	// In original code, they have problem on Windows
	url = "https://www.deutschland.de/en"
	checkMockFile(url, "2017-08-01")

	url = "http://www.greenpeace.org/international/en/campaigns/forests/asia-pacific/"
	checkMockFile(url, "2017-04-28")

	url = "https://www.creativecommons.at/faircoin-hackathon"
	checkMockFile(url, "2017-07-24")

	url = "https://pixabay.com/en/service/terms/"
	checkMockFile(url, "2017-08-09")

	url = "https://bayern.de/"
	checkMockFile(url, "2017-10-06")

	url = "https://www.pferde-fuer-unsere-kinder.de/unsere-projekte/"
	checkMockFile(url, "2016-07-20") // most probably 2016-07-15

	url = "http://www.hundeverein-querfurt.de/index.php?option=com_content&view=article&id=54&Itemid=50"
	checkMockFile(url, "2016-12-04") // 2010-11-01 in meta, 2016 more plausible

	url = "http://www.pbrunst.de/news/2011/12/kein-cyberterrorismus-diesmal/"
	checkMockFile(url, "2011-12-01")

	// TODO: problematic, should take URL instead
	url = "http://www.pbrunst.de/news/2011/12/kein-cyberterrorismus-diesmal/"
	checkMockFile(url, "2010-06-01", useOriginalDate)

	// Dates in table
	url = "http://www.hundeverein-kreisunna.de/termine.html"
	checkMockFile(url, "2017-03-29") // probably newer

	// ===================================================
	// Tests below these point are for URL with exact date
	// ===================================================

	url = "http://example.com/category/2016/07/12/key-words"
	checkURL(url, "2016-07-12")

	url = "http://example.com/2016/key-words"
	checkURL(url, "")

	url = "http://www.kreditwesen.org/widerstand-berlin/2012-11-29/keine-kurzung-bei-der-jugend-klubs-konnen-vorerst-aufatmen-bvv-beschliest-haushaltsplan/"
	checkURL(url, "2012-11-29")

	url = "http://www.kreditwesen.org/widerstand-berlin/6666-42-87/"
	checkURL(url, "")

	url = "https://www.pamelaandersonfoundation.org/news/2019/6/26/dm4wjh7skxerzzw8qa8cklj8xdri5j"
	checkURL(url, "2019-06-26")

	// =========================================================
	// Tests below these point are for URL with approximate date
	// =========================================================

	url = "http://example.com/category/2016/"
	checkURL(url, "")

	// ==============================================
	// Tests below these point are for idiosyncrasies
	// ==============================================

	str = `<html><body><p><em>Last updated: 5/5/20</em></p></body></html>`
	checkString(str, "2020-05-05")

	str = `<html><body><p><em>Last updated: 01/23/2021</em></p></body></html>`
	checkString(str, "2021-01-23")

	str = `<html><body><p><em>Last updated: 01/23/21</em></p></body></html>`
	checkString(str, "2021-01-23")

	str = `<html><body><p><em>Last updated: 1/23/21</em></p></body></html>`
	checkString(str, "2021-01-23")

	str = `<html><body><p><em>Last updated: 23/1/21</em></p></body></html>`
	checkString(str, "2021-01-23")

	str = `<html><body><p><em>Last updated: 23/01/21</em></p></body></html>`
	checkString(str, "2021-01-23")

	str = `<html><body><p><em>Last updated: 23/01/2021</em></p></body></html>`
	checkString(str, "2021-01-23")

	str = `<html><body><p><em>Last updated: 33/23/3033</em></p></body></html>`
	checkString(str, "")

	str = `<html><body><p><em>Published: 5/5/2020</em></p></body></html>`
	checkString(str, "2020-05-05")

	str = `<html><body><p><em>Published in: 05.05.2020</em></p></body></html>`
	checkString(str, "2020-05-05")

	str = `<html><body><p><em>Son güncelleme: 5/5/20</em></p></body></html>`
	checkString(str, "2020-05-05")

	str = `<html><body><p><em>Son güncellenme: 5/5/2020</em></p></body></html>`
	checkString(str, "2020-05-05")

	str = `<html><body><p><em>Yayımlama tarihi: 05.05.2020</em></p></body></html>`
	checkString(str, "2020-05-05")

	str = `<html><body><p><em>Son güncelleme tarihi: 5/5/20</em></p></body></html>`
	checkString(str, "2020-05-05")

	str = `<html><body><p><em>5/5/20 tarihinde güncellendi.</em></p></body></html>`
	checkString(str, "2020-05-05")

	str = `<html><body><p><em>5/5/2020 tarihinde yayımlandı.</em></p></body></html>`
	checkString(str, "2020-05-05")

	str = `<html><body><p><em>5/5/2020 tarihinde yayımlandı.</em></p></body></html>`
	checkString(str, "2020-05-05")

	str = `<html><body><p><em>05.05.2020 tarihinde yayınlandı.</em></p></body></html>`
	checkString(str, "2020-05-05")

	// =========================================================
	// Tests below these point are from original htmldate README
	// =========================================================

	url = "http://blog.python.org/2016/12/python-360-is-now-available.html"
	checkMockFile(url, "2016-12-23")

	url = "https://creativecommons.org/about/"
	checkMockFile(url, "", skipExtensiveSearch)

	str = `<html><body><span class="entry-date">July 12th, 2016</span></body></html>`
	checkString(str, "2016-07-12")

	str = `<html><body><span class="entry-date">July 12th, 2016</span></body></html>`
	checkString(str, "2016-07-12")

	url = "https://www.gnu.org/licenses/gpl-3.0.en.html"
	checkMockFile(url, "2016-11-18") // could also be: 29 June 2007

	url = "https://netzpolitik.org/2016/die-cider-connection-abmahnungen-gegen-nutzer-von-creative-commons-bildern/"
	checkMockFile(url, "2019-06-24")

	url = "https://netzpolitik.org/2016/die-cider-connection-abmahnungen-gegen-nutzer-von-creative-commons-bildern/"
	checkMockFile(url, "2016-06-23", useOriginalDate)

	url = "https://blog.wikimedia.org/2018/06/28/interactive-maps-now-in-your-language/"
	checkMockFile(url, "2018-06-28")

	// =================================================================
	// Tests below these point are for fancy dates which in the original
	// htmldate can only be parsed by `scrapinghub/dateparser`
	// =================================================================

	url = "https://blogs.mediapart.fr/elba/blog/260619/violences-policieres-bombe-retardement-mediatique"
	checkMockFile(url, "2019-06-27")

	url = "https://la-bas.org/la-bas-magazine/chroniques/Didier-Porte-souhaite-la-Sante-a-Balkany"
	checkMockFile(url, "2019-06-28")

	url = "https://www.revolutionpermanente.fr/Antonin-Bernanos-en-prison-depuis-pres-de-deux-mois-en-raison-de-son-militantisme"
	checkMockFile(url, "2019-06-13")

	// ==================================================
	// Tests below these point are for deferred URL dates
	// ==================================================

	str = `<!doctype html>
	<html lang="en-CA" class="no-js">
	
	<head>
		<link rel="canonical" href="https://www.fool.ca/2022/10/20/3-stable-stocks-id-buy-if-the-market-tanks-further/" />
		<meta property="article:published_time" content="2022-10-20T18:45:00+00:00" />
		<meta property="article:modified_time" content="2022-10-20T18:35:08+00:00" />
		<script type="application/ld+json" class="yoast-schema-graph">{"@context":"https://schema.org","@graph":[{"@type":"WebPage","@id":"https://www.fool.ca/2022/10/20/3-stable-stocks-id-buy-if-the-market-tanks-further/#webpage","url":"https://www.fool.ca/2022/10/20/3-stable-stocks-id-buy-if-the-market-tanks-further/","name":"3 Stable Stocks I'd Buy if the Market Tanks Further | The Motley Fool Canada","isPartOf":{"@id":"https://www.fool.ca/#website"},"datePublished":"2022-10-20T18:45:00+00:00","dateModified":"2022-10-20T18:35:08+00:00","description":"Dividend aristocrats contain stable stocks that any investor should consider, but these three offer the best chance at future growth as well.","breadcrumb":{"@id":"https://www.fool.ca/2022/10/20/3-stable-stocks-id-buy-if-the-market-tanks-further/#breadcrumb"},"inLanguage":"en-CA"},{"@type":"NewsArticle","@id":"https://www.fool.ca/2022/10/20/3-stable-stocks-id-buy-if-the-market-tanks-further/#article","isPartOf":{"@id":"https://www.fool.ca/2022/10/20/3-stable-stocks-id-buy-if-the-market-tanks-further/#webpage"},"author":{"@id":"https://www.fool.ca/#/schema/person/e0d452bd1e82135f310295e7dc650aca"},"headline":"3 Stable Stocks I&#8217;d Buy if the Market Tanks Further","datePublished":"2022-10-20T18:45:00+00:00","dateModified":"2022-10-20T18:35:08+00:00"}]}</script>
	</head>
	
	<body class="post-template-default single single-post postid-1378278 single-format-standard mega-menu-main-menu-2020 mega-menu-footer-2020" data-has-main-nav="true"> <span class="posted-on">Published <time class="entry-date published" datetime="2022-10-20T14:45:00-04:00">October 20, 2:45 pm EDT</time></span> </body>
	
	</html>`

	opts := Options{ExtractTime: true, UseOriginalDate: true}
	opts.DeferUrlExtractor = true
	res := extractFromString(str, opts)
	assert.Equal(t, "2022-10-20 18:45", res.Format("2006-01-02 15:04"))

	opts.DeferUrlExtractor = false
	res = extractFromString(str)
	assert.Equal(t, "2022-10-20 00:00", res.Format("2006-01-02 15:04"))
}

func Test_findTime(t *testing.T) {
	// Helper function
	check := func(expectedOutput string, input string, tzExist bool) {
		var output string
		h, m, s, tz, found := findTime(input)
		if found {
			loc := tz
			if loc == nil {
				loc = time.UTC
			}

			dt := time.Date(1, 1, 1, h, m, s, 0, loc)
			output = dt.Format("15:04:05 -0700")
		}

		assert.Equal(t, expectedOutput, output, input)
		assert.Equal(t, tzExist, tz != nil, input)
	}

	// ISO-8601 format
	check("12:00:00 +0000", "12:00", false)
	check("12:00:10 +0000", "12:00:10", false)
	check("12:00:10 +0000", "12:00:10.372", false)
	check("10:21:00 +0000", "10:21Z", true)
	check("10:21:40 +0000", "10:21:40Z", true)
	check("10:21:40 +0000", "10:21:40.462Z", true)
	check("16:14:00 +0200", "16:14+02:00", true)
	check("16:14:51 +0200", "16:14:51+02:00", true)
	check("16:14:51 +0200", "16:14:51.075+02:00", true)
	check("16:14:51 +0200", "16:14:51.075+0200", true)
	check("16:14:51 +0200", "16:14:51.075+02", true)

	// Common format
	check("07:08:00 +0000", "7:8", false)
	check("07:08:09 +0000", "7:8:9", false)
	check("07:08:00 +0000", "7:8 am", false)
	check("07:08:09 +0000", "7:8:9 am", false)
	check("19:08:00 +0000", "7:8 pm", false)
	check("19:08:09 +0000", "7:8:9 pm", false)
	check("07:08:00 +0000", "7:8 a.m.", false)
	check("07:08:09 +0000", "7:8:9 a.m.", false)
	check("19:08:00 +0000", "7:8 p.m.", false)
	check("19:08:09 +0000", "7:8:9 p.m.", false)
	check("07:08:00 +0000", "07:08", false)
	check("07:08:09 +0000", "07:08:09", false)
	check("07:08:00 +0000", "07:08 am", false)
	check("07:08:09 +0000", "07:08:09 am", false)
	check("19:08:00 +0000", "07:08 pm", false)
	check("19:08:09 +0000", "07:08:09 pm", false)
	check("07:08:00 +0000", "07:08 a.m.", false)
	check("07:08:09 +0000", "07:08:09 a.m.", false)
	check("19:08:00 +0000", "07:08 p.m.", false)
	check("19:08:09 +0000", "07:08:09 p.m.", false)
	check("07:08:00 +0100", "07:08 a.m. +0100", true)
	check("07:08:09 +0100", "07:08:09 a.m. +0100", true)
	check("19:08:00 +0100", "07:08 p.m. +0100", true)
	check("19:08:09 +0100", "07:08:09 p.m. +0100", true)

	// French format
	check("07:08:00 +0100", "07h08 a.m. +0100", true)
	check("19:08:00 +0100", "07h08 p.m. +0100", true)
}

func Test_compareReference(t *testing.T) {
	opts := Options{
		MinDate: defaultMinDate,
		MaxDate: defaultMaxDate,
	}

	_, res := compareReference("", 0, "AAAA", opts)
	assert.Equal(t, int64(0), res)

	_, res = compareReference("", 1517500000, "2018-33-01", opts)
	assert.Equal(t, int64(1517500000), res)

	_, res = compareReference("", 0, "2018-02-01", opts)
	assert.Less(t, int64(1517400000), res)
	assert.Greater(t, int64(1517500000), res)

	_, res = compareReference("", 1517500000, "2018-02-01", opts)
	assert.Equal(t, int64(1517500000), res)
}

func Test_selectCandidate(t *testing.T) {
	// Initiate variables and helper function
	rxYear := regexp.MustCompile(`^([0-9]{4})`)
	rxCatch := regexp.MustCompile(`([0-9]{4})-([0-9]{2})-([0-9]{2})`)
	opts := Options{MinDate: defaultMinDate, MaxDate: defaultMaxDate}

	// Nonsense
	candidates := createCandidates("20208956", "20208956", "20208956",
		"19018956", "209561", "22020895607-12", "2-28")
	_, result := selectCandidate(candidates, rxCatch, rxYear, opts)
	assert.Empty(t, result)

	// Plausible
	candidates = createCandidates("2016-12-23", "2016-12-23", "2016-12-23",
		"2016-12-23", "2017-08-11", "2016-07-12", "2017-11-28")
	_, result = selectCandidate(candidates, rxCatch, rxYear, opts)
	assert.Equal(t, "2017-11-28", result[0])

	opts.UseOriginalDate = true
	_, result = selectCandidate(candidates, rxCatch, rxYear, opts)
	assert.Equal(t, "2016-07-12", result[0])

	// Mix plausible and implausible
	candidates = createCandidates("2116-12-23", "2116-12-23", "2116-12-23",
		"2017-08-11", "2017-08-11")
	_, result = selectCandidate(candidates, rxCatch, rxYear, opts)
	assert.Equal(t, "2017-08-11", result[0])

	opts.UseOriginalDate = false
	_, result = selectCandidate(candidates, rxCatch, rxYear, opts)
	assert.Equal(t, "2017-08-11", result[0])

	// Taking date present twice, corner case
	candidates = createCandidates("2016-12-23", "2016-12-23", "2017-08-11",
		"2017-08-11", "2017-08-11")
	_, result = selectCandidate(candidates, rxCatch, rxYear, opts)
	assert.Equal(t, "2016-12-23", result[0])

}

func Test_searchPage(t *testing.T) {
	// Variables
	var dt time.Time
	opts := Options{
		MinDate: defaultMinDate,
		MaxDate: defaultMaxDate,
	}

	// Helper function
	format := func(t time.Time) string {
		if !t.IsZero() {
			return t.Format("2006-01-02")
		}
		return ""
	}

	// From file
	f := openMockFile("http://www.heimicke.de/chronik/zahlen-und-daten/")
	defer f.Close()

	bt, _ := io.ReadAll(f)
	_, dt = searchPage(string(bt), opts)
	assert.Equal(t, "2019-04-06", format(dt))

	// From string
	_, dt = searchPage(`<html><body><p>The date is 5/2010</p></body></html>`, opts)
	assert.Equal(t, "2010-05-01", format(dt))

	_, dt = searchPage(`<html><body><p>The date is 5.5.2010</p></body></html>`, opts)
	assert.Equal(t, "2010-05-05", format(dt))

	_, dt = searchPage(`<html><body><p>The date is 11/10/99</p></body></html>`, opts)
	assert.Equal(t, "1999-10-11", format(dt))

	_, dt = searchPage(`<html><body><p>The date is 3/3/11</p></body></html>`, opts)
	assert.Equal(t, "2011-03-03", format(dt))

	_, dt = searchPage(`<html><body><p>The date is 06.12.06</p></body></html>`, opts)
	assert.Equal(t, "2006-12-06", format(dt))

	_, dt = searchPage(`<html><body><p>The timestamp is 20140915D15:23H</p></body></html>`, opts)
	assert.Equal(t, "2014-09-15", format(dt))

	_, dt = searchPage(`<html><body><p>It could be 2015-04-30 or 2003-11-24.</p></body></html>`, opts)
	assert.Equal(t, "2015-04-30", format(dt))

	useOriginal := mergeOpts(Options{UseOriginalDate: true}, opts)
	_, dt = searchPage(`<html><body><p>It could be 2015-04-30 or 2003-11-24.</p></body></html>`, useOriginal)
	assert.Equal(t, "2003-11-24", format(dt))

	_, dt = searchPage(`<html><body><p>It could be 03/03/2077 or 03/03/2013.</p></body></html>`, opts)
	assert.Equal(t, "2013-03-03", format(dt))

	_, dt = searchPage(`<html><body><p>It could not be 03/03/2077 or 03/03/1988.</p></body></html>`, opts)
	assert.Equal(t, "", format(dt))

	_, dt = searchPage(`<html><body><p>© The Web Association 2013.</p></body></html>`, opts)
	assert.Equal(t, "2013-01-01", format(dt))

	_, dt = searchPage(`<html><body><p>Next © Copyright 2018</p></body></html>`, opts)
	assert.Equal(t, "2018-01-01", format(dt))

	_, dt = searchPage(`<html><body><p> © Company 2014-2019 </p></body></html>`, opts)
	assert.Equal(t, "2019-01-01", format(dt))

	_, dt = searchPage(`<html><head><link xmlns="http://www.w3.org/1999/xhtml"/></head></html>`, opts)
	assert.Equal(t, "", format(dt))

	_, dt = searchPage(`<html><body><link href="//homepagedesigner.telekom.de/.cm4all/res/static/beng-editor/5.1.98/css/deploy.css"/></body></html>`, opts)
	assert.Equal(t, "", format(dt))
}

func Test_searchPattern(t *testing.T) {
	// Variables
	opts := Options{MinDate: defaultMinDate, MaxDate: defaultMaxDate}

	// First pattern, YYYY MM
	pattern := regexp.MustCompile(`\D([0-9]{4}[/.-][0-9]{2})\D`)
	catchPattern := regexp.MustCompile(`([0-9]{4})[/.-]([0-9]{2})`)
	yearPattern := regexp.MustCompile(`^([12][0-9]{3})`)

	str := "It happened on the 202.E.19, the day when it all began."
	_, res := searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.Empty(t, res)

	str = "The date is 2002.02.15."
	_, res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.NotEmpty(t, res)
	assert.Equal(t, "2002.02", res[0])

	str = "http://www.url.net/index.html"
	_, res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.Empty(t, res)

	str = "http://www.url.net/2016/01/index.html"
	_, res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.NotEmpty(t, res)
	assert.Equal(t, "2016/01", res[0])

	// Second pattern, MM YYYY
	pattern = regexp.MustCompile(`\D([0-9]{2}[/.-][0-9]{4})\D`)
	catchPattern = regexp.MustCompile(`([0-9]{2})[/.-]([0-9]{4})`)
	yearPattern = regexp.MustCompile(`([12][0-9]{3})$`)

	str = "It happened on the 202.E.19, the day when it all began."
	_, res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.Empty(t, res)

	str = "It happened on the 15.02.2002, the day when it all began."
	_, res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.NotEmpty(t, res)
	assert.Equal(t, "02.2002", res[0])

	// Third pattern, YYYY only
	pattern = regexp.MustCompile(`\D(2[01][0-9]{2})\D`)
	catchPattern = regexp.MustCompile(`(2[01][0-9]{2})`)
	yearPattern = regexp.MustCompile(`^(2[01][0-9]{2})`)

	str = "It happened in the film 300."
	_, res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.Empty(t, res)

	str = "It happened in 2002."
	_, res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.NotEmpty(t, res)
	assert.Equal(t, "2002", res[0])
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
			Pattern: item,
			Count:   mapItemCount[item],
		})
	}

	return candidates
}
