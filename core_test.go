package htmldate

import (
	"io/ioutil"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_HtmlDate(t *testing.T) {
	// Variables
	var str, url string
	var dt time.Time
	useOriginalDate := Options{UseOriginalDate: true}
	skipExtensiveSearch := Options{SkipExtensiveSearch: true}

	// Helper function
	format := func(t time.Time) string {
		if !t.IsZero() {
			return t.Format("2006-01-02")
		}
		return ""
	}

	// ===================================================
	// Tests below these point should return an exact date
	// ===================================================

	// These pages shouldn't return any date
	url = "https://www.intel.com/content/www/us/en/legal/terms-of-use.html"
	dt = extractMockFile(url)
	assert.Equal(t, "", format(dt))

	url = "https://en.support.wordpress.com/"
	dt = extractMockFile(url)
	assert.Equal(t, "", format(dt))

	str = "<html><body>XYZ</body></html>"
	dt = extractFromString(str)
	assert.Equal(t, "", format(dt))

	// HTML document tree
	str = `<html><head><meta property="dc:created" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta property="dc:created" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str, useOriginalDate)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta http-equiv="date" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str, useOriginalDate)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta name="last-modified" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta property="OG:Updated_Time" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta property="og:updated_time" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta name="created" content="2017-01-09"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-01-09", format(dt))

	str = `<html><head><meta itemprop="copyrightyear" content="2017"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-01-01", format(dt))

	// Original date
	str = `<html><head><meta property="OG:Updated_Time" content="2017-09-01"/><meta property="OG:Original_Time" content="2017-07-02"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-09-01", format(dt))

	dt = extractFromString(str, useOriginalDate)
	assert.Equal(t, "2017-07-02", format(dt))

	// Link in header
	url = "http://www.jovelstefan.de/2012/05/11/parken-in-paris/"
	dt = extractMockFile(url)
	assert.Equal(t, "2012-05-11", format(dt))

	// Meta in header
	str = `<html><head><meta/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "", format(dt))

	url = "https://500px.com/photo/26034451/spring-in-china-by-alexey-kruglov"
	dt = extractMockFile(url)
	assert.Equal(t, "2013-02-16", format(dt))

	str = `<html><head><meta name="og:url" content="http://www.example.com/2018/02/01/entrytitle"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-02-01", format(dt))

	str = `<html><head><meta itemprop="datecreated" datetime="2018-02-02"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-02-02", format(dt))

	str = `<html><head><meta itemprop="datemodified" content="2018-02-04"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-02-04", format(dt))

	str = `<html><head><meta http-equiv="last-modified" content="2018-02-05"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-02-05", format(dt))

	str = `<html><head><meta name="Publish_Date" content="02.02.2004"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2004-02-02", format(dt))

	str = `<html><head><meta name="pubDate" content="2018-02-06"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-02-06", format(dt))

	str = `<html><head><meta pubdate="pubDate" content="2018-02-06"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-02-06", format(dt))

	str = `<html><head><meta itemprop="DateModified" datetime="2018-02-06"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-02-06", format(dt))

	// Time in document body
	url = "https://www.facebook.com/visitaustria/"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-10-08", format(dt))

	dt = extractMockFile(url, useOriginalDate)
	assert.Equal(t, "2017-10-06", format(dt))

	url = "http://www.medef.com/en/content/alternative-dispute-resolution-for-antitrust-damages"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><body><time datetime="08:00"></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "", format(dt))

	str = `<html><body><time datetime="2014-07-10 08:30:45.687"></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2014-07-10", format(dt))

	str = `<html><head></head><body><time class="entry-time" itemprop="datePublished" datetime="2018-04-18T09:57:38+00:00"></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-04-18", format(dt))

	str = `<html><body><footer class="article-footer"><p class="byline">Veröffentlicht am <time class="updated" datetime="2019-01-03T14:56:51+00:00">3. Januar 2019 um 14:56 Uhr.</time></p></footer></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2019-01-03", format(dt))

	str = `<html><body><footer class="article-footer"><p class="byline">Veröffentlicht am <time class="updated" datetime="2019-01-03T14:56:51+00:00"></time></p></footer></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2019-01-03", format(dt))

	// Removed from HTML5 https://www.w3schools.com/TAgs/att_time_datetime_pubdate.asp
	str = `<html><body><time datetime="2011-09-28" pubdate="pubdate"></time></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2011-09-28", format(dt))

	dt = extractFromString(str, useOriginalDate)
	assert.Equal(t, "2011-09-28", format(dt))

	str = `<html><body><time datetime="2011-09-28" class="entry-date"></time></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2011-09-28", format(dt))

	// Precise pattern in document body
	str = `<html><body><font size="2" face="Arial,Geneva,Helvetica">Bei <a href="../../sonstiges/anfrage.php"><b>Bestellungen</b></a> bitte Angabe der Titelnummer nicht vergessen!<br><br>Stand: 03.04.2019</font></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2019-04-03", format(dt))

	str = `<html><body>Datum: 10.11.2017</body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-11-10", format(dt))

	url = `https://www.tagesausblick.de/Analyse/USA/DOW-Jones-Jahresendrally-ade__601.html`
	dt = extractMockFile(url)
	assert.Equal(t, "2012-12-22", format(dt))

	url = `http://blog.todamax.net/2018/midp-emulator-kemulator-und-brick-challenge/`
	dt = extractMockFile(url)
	assert.Equal(t, "2018-02-15", format(dt))

	// JSON date published
	url = `https://www.acredis.com/schoenheitsoperationen/augenlidstraffung/`
	dt = extractMockFile(url, useOriginalDate)
	assert.Equal(t, "2018-02-28", format(dt))

	// JSON date modified
	url = `https://www.channelpartner.de/a/sieben-berufe-die-zukunft-haben,3050673`
	dt = extractMockFile(url)
	assert.Equal(t, "2019-04-03", format(dt))

	// Meta in document body
	url = "https://futurezone.at/digital-life/wie-creativecommons-richtig-genutzt-wird/24.600.504"
	dt = extractMockFile(url, useOriginalDate)
	assert.Equal(t, "2013-08-09", format(dt))

	url = "https://www.horizont.net/marketing/kommentare/influencer-marketing-was-sich-nach-dem-vreni-frost-urteil-aendert-und-aendern-muss-172529"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-01-29", format(dt))

	url = "http://www.klimawandel-global.de/klimaschutz/energie-sparen/elektromobilitat-der-neue-trend/"
	dt = extractMockFile(url)
	assert.Equal(t, "2013-05-03", format(dt))

	url = "http://www.hobby-werkstatt-blog.de/arduino/424-eine-arduino-virtual-wall-fuer-den-irobot-roomba.php"
	dt = extractMockFile(url)
	assert.Equal(t, "2015-12-14", format(dt))

	url = "https://www.beltz.de/fachmedien/paedagogik/didacta_2019_in_koeln_19_23_februar/beltz_veranstaltungen_didacta_2016/veranstaltung.html?tx_news_pi1%5Bnews%5D=14392&tx_news_pi1%5Bcontroller%5D=News&tx_news_pi1%5Baction%5D=detail&cHash=10b1a32fb5b2b05360bdac257b01c8fa"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-02-20", format(dt))

	url = "https://www.wienbadminton.at/news/119843/Come-Together"
	dt = extractMockFile(url, skipExtensiveSearch)
	assert.Equal(t, "", format(dt))

	url = "https://www.wienbadminton.at/news/119843/Come-Together"
	dt = extractMockFile(url)
	assert.Equal(t, "2018-05-06", format(dt))

	// Abbr in document body
	url = "http://blog.kinra.de/?p=959/"
	dt = extractMockFile(url)
	assert.Equal(t, "2012-12-16", format(dt))

	str = `<html><body><abbr class="published">am 12.11.16</abbr></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2016-11-12", format(dt))

	str = `<html><body><abbr class="published">am 12.11.16</abbr></body></html>`
	dt = extractFromString(str, useOriginalDate)
	assert.Equal(t, "2016-11-12", format(dt))

	str = `<html><body><abbr class="published" title="2016-11-12">XYZ</abbr></body></html>`
	dt = extractFromString(str, useOriginalDate)
	assert.Equal(t, "2016-11-12", format(dt))

	str = `<html><body><abbr class="date-published">8.11.2016</abbr></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2016-11-08", format(dt))

	// Valid vs invalid data-utime
	str = `<html><body><abbr data-utime="1438091078" class="something">A date</abbr></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2015-07-28", format(dt))

	str = `<html><body><abbr data-utime="143809-1078" class="something">A date</abbr></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "", format(dt))

	// Time in document body
	str = `<html><body><time>2018-01-04</time></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-01-04", format(dt))

	url = "https://www.adac.de/rund-ums-fahrzeug/tests/kindersicherheit/kindersitztest-2018/"
	dt = extractMockFile(url)
	assert.Equal(t, "2018-10-23", format(dt))

	// Additional selector rules
	str = `<html><body><div class="fecha">2018-01-04</div></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-01-04", format(dt))

	// Other expressions in document body
	str = `<html><body>"datePublished":"2018-01-04"</body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-01-04", format(dt))

	str = `<html><body>Stand: 1.4.18</body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-04-01", format(dt))

	url = "http://www.stuttgart.de/"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-10-09", format(dt))

	// In document body
	url = "https://github.com/adbar/htmldate"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-01-01", format(dt))

	url = "https://en.blog.wordpress.com/"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-08-30", format(dt))

	url = "https://www.austria.info/"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-09-07", format(dt))

	url = "https://www.eff.org/files/annual-report/2015/index.html"
	dt = extractMockFile(url)
	assert.Equal(t, "2016-05-04", format(dt))

	url = "http://unexpecteduser.blogspot.de/2011/"
	dt = extractMockFile(url)
	assert.Equal(t, "2011-03-30", format(dt))

	url = "https://die-partei.net/sh/"
	dt = extractMockFile(url)
	assert.Equal(t, "2014-07-19", format(dt))

	url = "https://www.rosneft.com/business/Upstream/Licensing/"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-02-27", format(dt)) // most probably 2014-12-31, found in text

	url = "http://www.freundeskreis-videoclips.de/waehlen-sie-car-player-tipps-zur-auswahl-der-besten-car-cd-player/"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-07-12", format(dt))

	url = "https://www.scs78.de/news/items/warm-war-es-schoen-war-es.html"
	dt = extractMockFile(url)
	assert.Equal(t, "2018-06-10", format(dt))

	url = "https://www.goodform.ch/blog/schattiges_plaetzchen"
	dt = extractMockFile(url)
	assert.Equal(t, "2018-06-27", format(dt))

	url = "https://www.transgen.de/aktuell/2687.afrikanische-schweinepest-genome-editing.html"
	dt = extractMockFile(url)
	assert.Equal(t, "2018-01-18", format(dt))

	url = "http://www.eza.gv.at/das-ministerium/presse/aussendungen/2018/07/aussenministerin-karin-kneissl-beim-treffen-der-deutschsprachigen-aussenminister-in-luxemburg/"
	dt = extractMockFile(url)
	assert.Equal(t, "2018-07-03", format(dt))

	url = "https://www.weltwoche.ch/ausgaben/2019-4/artikel/forbes-die-weltwoche-ausgabe-4-2019.html"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-01-23", format(dt))

	// Free text
	str = `<html><body>&copy; 2017</body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-01-01", format(dt))

	str = `<html><body>© 2017</body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-01-01", format(dt))

	str = `<html><body><p>Dieses Datum ist leider ungültig: 30. Februar 2018.</p></body></html>`
	dt = extractFromString(str, skipExtensiveSearch)
	assert.Equal(t, "", format(dt))

	str = `<html><body><p>Dieses Datum ist leider ungültig: 30. Februar 2018.</p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-01-01", format(dt))

	// Other format
	url = "http://unexpecteduser.blogspot.de/2011/"
	dt = extractMockFile(url)
	assert.Equal(t, "30 March 2011", dt.Format("2 January 2006"))

	url = "http://blog.python.org/2016/12/python-360-is-now-available.html"
	dt = extractMockFile(url)
	assert.Equal(t, "23 December 2016", dt.Format("2 January 2006"))

	// Additional list
	url = "http://carta.info/der-neue-trend-muss-statt-wunschkoalition/"
	dt = extractMockFile(url)
	assert.Equal(t, "2012-05-08", format(dt))

	url = "https://www.wunderweib.de/manuela-reimann-hochzeitsueberraschung-in-bayern-107930.html"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-06-20", format(dt))

	url = "https://www.befifty.de/home/2017/7/12/unter-uns-montauk"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-07-12", format(dt))

	url = "https://www.brigitte.de/aktuell/riverdale--so-ehrt-die-serie-luke-perry-in-staffel-vier-11602344.html"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-06-20", format(dt))

	url = "http://www.loldf.org/spip.php?article717"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-06-27", format(dt))

	url = "https://www.beltz.de/sachbuch_ratgeber/buecher/produkt_produktdetails/37219-12_wege_zu_guter_pflege.html"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-02-07", format(dt))

	url = "https://www.oberstdorf-resort.de/interaktiv/blog/unser-kraeutergarten-wannenkopfhuette.html"
	dt = extractMockFile(url)
	assert.Equal(t, "2018-06-20", format(dt))

	url = "https://www.wienbadminton.at/news/119843/Come-Together"
	dt = extractMockFile(url)
	assert.Equal(t, "2018-05-06", format(dt))

	url = "https://www.ldt.de/ldtblog/fall-in-love-with-black/"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-08-08", format(dt))

	url = "https://paris-luttes.info/quand-on-comprend-que-les-grenades-12355"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-06-29", format(dt)) // here we are better than the original

	url = "https://verfassungsblog.de/the-first-decade/"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-07-13", format(dt))

	url = "https://cric-grenoble.info/infos-locales/article/putsh-en-cours-a-radio-kaleidoscope-1145"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-06-09", format(dt))

	url = "https://www.sebastian-kurz.at/magazin/wasserstoff-als-schluesseltechnologie"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-07-30", format(dt))

	url = "https://exporo.de/wiki/europaeische-zentralbank-ezb/"
	dt = extractMockFile(url, useOriginalDate)
	assert.Equal(t, "2018-01-01", format(dt))

	// Only found by extensive search
	url = "https://ebene11.com/die-arbeit-mit-fremden-dwg-dateien-in-autocad"
	dt = extractMockFile(url, skipExtensiveSearch)
	assert.Equal(t, "", format(dt))

	url = "https://ebene11.com/die-arbeit-mit-fremden-dwg-dateien-in-autocad"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-01-12", format(dt))

	url = "https://www.hertie-school.org/en/debate/detail/content/whats-on-the-cards-for-von-der-leyen/"
	dt = extractMockFile(url, skipExtensiveSearch)
	assert.Equal(t, "", format(dt))

	url = "https://www.hertie-school.org/en/debate/detail/content/whats-on-the-cards-for-von-der-leyen/"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-12-02", format(dt)) // Or maybe 2019-02-12

	// Date not in footer but at the start of the article
	url = "http://www.wara-enforcement.org/guinee-un-braconnier-delephant-interpelle-et-condamne-a-la-peine-maximale/"
	dt = extractMockFile(url)
	assert.Equal(t, "2016-09-27", format(dt))

	// URL from meta image in header
	str = `<html><meta property="og:image" content="https://example.org/img/2019-05-05/test.jpg"><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2019-05-05", format(dt))

	str = `<html><meta property="og:image" content="https://example.org/img/test.jpg"><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "", format(dt))

	// URL from <img> in body
	str = `<html><body><img src="https://example.org/img/2019-05-05/test.jpg"/></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2019-05-05", format(dt))

	str = `<html><body><img src="https://example.org/img/test.jpg"/></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "", format(dt))

	str = `<html><body><img src="https://example.org/img/2019-05-03/test.jpg"/><img src="https://example.org/img/2019-05-04/test.jpg"/><img src="https://example.org/img/2019-05-05/test.jpg"/></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2019-05-05", format(dt))

	str = `<html><body><img src="https://example.org/img/2019-05-05/test.jpg"/><img src="https://example.org/img/2019-05-04/test.jpg"/><img src="https://example.org/img/2019-05-03/test.jpg"/></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2019-05-05", format(dt))

	// Title content
	str = `<html><head><title>Bericht zur Coronalage vom 22.04.2020 – worauf wartet die Politik? – DIE ACHSE DES GUTEN. ACHGUT.COM</title></head></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-04-22", format(dt))

	// In unknown div
	str = `<html><body><div class="spip spip-block-right" style="text-align:right;">Le 26 juin 2019</div></body></html>`
	dt = extractFromString(str, skipExtensiveSearch)
	assert.Equal(t, "", format(dt))

	str = `<html><body><div class="spip spip-block-right" style="text-align:right;">Le 26 juin 2019</div></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2019-06-26", format(dt))

	// In link title
	str = `<html><body><a class="ribbon date " title="12th December 2018" href="https://example.org/" itemprop="url">Text</a></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2018-12-12", format(dt))

	// =========================================================
	// Tests below these point should return an approximate date
	// =========================================================

	// Copyright text
	url = "http://viehbacher.com/de/spezialisierung/internationale-forderungsbeitreibung"
	dt = extractMockFile(url)
	assert.Equal(t, "2016-01-01", format(dt)) // somewhere in 2016

	// Other
	url = "https://creativecommons.org/about/"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-08-11", format(dt)) // or '2017-08-03'

	url = "https://creativecommons.org/about/"
	dt = extractMockFile(url, useOriginalDate)
	assert.Equal(t, "2016-05-22", format(dt)) // or '2017-08-03'

	// In original code, they have problem on Windows
	url = "https://www.deutschland.de/en"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-08-01", format(dt))

	url = "http://www.greenpeace.org/international/en/campaigns/forests/asia-pacific/"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-04-28", format(dt))

	url = "https://www.creativecommons.at/faircoin-hackathon"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-07-24", format(dt))

	url = "https://pixabay.com/en/service/terms/"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-08-09", format(dt))

	url = "https://bayern.de/"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-10-06", format(dt)) // most probably 2017-10-06

	url = "https://www.pferde-fuer-unsere-kinder.de/unsere-projekte/"
	dt = extractMockFile(url)
	assert.Equal(t, "2016-07-20", format(dt)) // most probably 2016-07-15

	url = "http://www.hundeverein-querfurt.de/index.php?option=com_content&view=article&id=54&Itemid=50"
	dt = extractMockFile(url)
	assert.Equal(t, "2016-12-04", format(dt)) // 2010-11-01 in meta, 2016 more plausible

	url = "http://www.pbrunst.de/news/2011/12/kein-cyberterrorismus-diesmal/"
	dt = extractMockFile(url)
	assert.Equal(t, "2011-12-01", format(dt))

	// TODO: problematic, should take URL instead
	url = "http://www.pbrunst.de/news/2011/12/kein-cyberterrorismus-diesmal/"
	dt = extractMockFile(url, useOriginalDate)
	assert.Equal(t, "2010-06-01", format(dt))

	// Dates in table
	url = "http://www.hundeverein-kreisunna.de/termine.html"
	dt = extractMockFile(url)
	assert.Equal(t, "2017-03-29", format(dt)) // probably newer

	// ===================================================
	// Tests below these point are for URL with exact date
	// ===================================================

	url = "http://example.com/category/2016/07/12/key-words"
	dt = extractFromURL(url)
	assert.Equal(t, "2016-07-12", format(dt))

	url = "http://example.com/2016/key-words"
	dt = extractFromURL(url)
	assert.Equal(t, "", format(dt))

	url = "http://www.kreditwesen.org/widerstand-berlin/2012-11-29/keine-kurzung-bei-der-jugend-klubs-konnen-vorerst-aufatmen-bvv-beschliest-haushaltsplan/"
	dt = extractFromURL(url)
	assert.Equal(t, "2012-11-29", format(dt))

	url = "http://www.kreditwesen.org/widerstand-berlin/2012-11/keine-kurzung-bei-der-jugend-klubs-konnen-vorerst-aufatmen-bvv-beschliest-haushaltsplan/"
	dt = extractFromURL(url)
	assert.Equal(t, "", format(dt))

	url = "http://www.kreditwesen.org/widerstand-berlin/6666-42-87/"
	dt = extractFromURL(url)
	assert.Equal(t, "", format(dt))

	url = "https://www.pamelaandersonfoundation.org/news/2019/6/26/dm4wjh7skxerzzw8qa8cklj8xdri5j"
	dt = extractFromURL(url)
	assert.Equal(t, "2019-06-26", format(dt))

	// =========================================================
	// Tests below these point are for URL with approximate date
	// =========================================================

	url = "http://example.com/blog/2016/07/key-words"
	dt = extractFromURL(url)
	assert.Equal(t, "2016-07-01", format(dt))

	url = "http://example.com/category/2016/"
	dt = extractFromURL(url)
	assert.Equal(t, "", format(dt))

	// ==============================================
	// Tests below these point are for idiosyncrasies
	// ==============================================

	str = `<html><body><p><em>Last updated: 5/5/20</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-05-05", format(dt))

	str = `<html><body><p><em>Last updated: 01/23/2021</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2021-01-23", format(dt))

	str = `<html><body><p><em>Published: 5/5/2020</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-05-05", format(dt))

	str = `<html><body><p><em>Published in: 05.05.2020</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-05-05", format(dt))

	str = `<html><body><p><em>Son güncelleme: 5/5/20</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-05-05", format(dt))

	str = `<html><body><p><em>Son güncellenme: 5/5/2020</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-05-05", format(dt))

	str = `<html><body><p><em>Yayımlama tarihi: 05.05.2020</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-05-05", format(dt))

	str = `<html><body><p><em>Son güncelleme tarihi: 5/5/20</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-05-05", format(dt))

	str = `<html><body><p><em>5/5/20 tarihinde güncellendi.</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-05-05", format(dt))

	str = `<html><body><p><em>5/5/2020 tarihinde yayımlandı.</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-05-05", format(dt))

	str = `<html><body><p><em>5/5/2020 tarihinde yayımlandı.</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-05-05", format(dt))

	str = `<html><body><p><em>05.05.2020 tarihinde yayınlandı.</em></p></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2020-05-05", format(dt))

	// =========================================================
	// Tests below these point are from original htmldate README
	// =========================================================

	url = "http://blog.python.org/2016/12/python-360-is-now-available.html"
	dt = extractMockFile(url)
	assert.Equal(t, "2016-12-23", format(dt))

	url = "https://creativecommons.org/about/"
	dt = extractMockFile(url, skipExtensiveSearch)
	assert.Equal(t, "", format(dt))

	str = `<html><body><span class="entry-date">July 12th, 2016</span></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2016-07-12", format(dt))

	str = `<html><body><span class="entry-date">July 12th, 2016</span></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2016-07-12", format(dt))

	url = "https://www.gnu.org/licenses/gpl-3.0.en.html"
	dt = extractMockFile(url)
	assert.Equal(t, "2016-11-18", format(dt))

	url = "https://netzpolitik.org/2016/die-cider-connection-abmahnungen-gegen-nutzer-von-creative-commons-bildern/"
	dt = extractMockFile(url, useOriginalDate)
	assert.Equal(t, "2016-06-23", format(dt))

	url = "https://blog.wikimedia.org/2018/06/28/interactive-maps-now-in-your-language/"
	dt = extractMockFile(url)
	assert.Equal(t, "2018-06-28", format(dt))

	// =================================================================
	// Tests below these point are for fancy dates which in the original
	// htmldate can only be parsed by `scrapinghub/dateparser`
	// =================================================================

	url = "https://blogs.mediapart.fr/elba/blog/260619/violences-policieres-bombe-retardement-mediatique"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-06-27", format(dt))

	url = "https://la-bas.org/la-bas-magazine/chroniques/Didier-Porte-souhaite-la-Sante-a-Balkany"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-06-28", format(dt))

	url = "https://www.revolutionpermanente.fr/Antonin-Bernanos-en-prison-depuis-pres-de-deux-mois-en-raison-de-son-militantisme"
	dt = extractMockFile(url)
	assert.Equal(t, "2019-06-13", format(dt))
}

func Test_compareReference(t *testing.T) {
	opts := Options{
		MinDate: defaultMinDate,
		MaxDate: defaultMaxDate,
	}

	res := compareReference(0, "AAAA", opts)
	assert.Equal(t, int64(0), res)

	res = compareReference(1517500000, "2018-33-01", opts)
	assert.Equal(t, int64(1517500000), res)

	res = compareReference(0, "2018-02-01", opts)
	assert.Less(t, int64(1517400000), res)
	assert.Greater(t, int64(1517500000), res)

	res = compareReference(1517500000, "2018-02-01", opts)
	assert.Equal(t, int64(1517500000), res)
}

func Test_selectCandidate(t *testing.T) {
	// Initiate variables and helper function
	rxYear := regexp.MustCompile(`^([0-9]{4})`)
	rxCatch := regexp.MustCompile(`([0-9]{4})-([0-9]{2})-([0-9]{2})`)
	opts := Options{MinDate: defaultMinDate, MaxDate: defaultMaxDate}

	// Candidate exist
	candidates := createCandidates("2016-12-23", "2016-12-23", "2016-12-23",
		"2016-12-23", "2017-08-11", "2016-07-12", "2017-11-28")
	result := selectCandidate(candidates, rxCatch, rxYear, opts)
	assert.NotEmpty(t, result)
	assert.Equal(t, "2017-11-28", result[0])

	// Candidates not exist
	candidates = createCandidates("20208956", "20208956", "20208956",
		"19018956", "209561", "22020895607-12", "2-28")
	result = selectCandidate(candidates, rxCatch, rxYear, opts)
	assert.Empty(t, result)
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

	bt, _ := ioutil.ReadAll(f)
	dt = searchPage(string(bt), opts)
	assert.Equal(t, "2019-04-06", format(dt))

	// From string
	dt = searchPage(`<html><body><p>The date is 5/2010</p></body></html>`, opts)
	assert.Equal(t, "2010-05-01", format(dt))

	dt = searchPage(`<html><body><p>The date is 5.5.2010</p></body></html>`, opts)
	assert.Equal(t, "2010-05-05", format(dt))

	dt = searchPage(`<html><body><p>The date is 11/10/99</p></body></html>`, opts)
	assert.Equal(t, "1999-10-11", format(dt))

	dt = searchPage(`<html><body><p>The date is 3/3/11</p></body></html>`, opts)
	assert.Equal(t, "2011-03-03", format(dt))

	dt = searchPage(`<html><body><p>The date is 06.12.06</p></body></html>`, opts)
	assert.Equal(t, "2006-12-06", format(dt))

	dt = searchPage(`<html><body><p>The timestamp is 20140915D15:23H</p></body></html>`, opts)
	assert.Equal(t, "2014-09-15", format(dt))

	dt = searchPage(`<html><body><p>It could be 2015-04-30 or 2003-11-24.</p></body></html>`, opts)
	assert.Equal(t, "2015-04-30", format(dt))

	dt = searchPage(`<html><body><p>It could be 03/03/2077 or 03/03/2013.</p></body></html>`, opts)
	assert.Equal(t, "2013-03-03", format(dt))

	dt = searchPage(`<html><body><p>It could not be 03/03/2077 or 03/03/1988.</p></body></html>`, opts)
	assert.Equal(t, "", format(dt))

	dt = searchPage(`<html><body><p>© The Web Association 2013.</p></body></html>`, opts)
	assert.Equal(t, "2013-01-01", format(dt))

	dt = searchPage(`<html><body><p>Next © Copyright 2018</p></body></html>`, opts)
	assert.Equal(t, "2018-01-01", format(dt))

	dt = searchPage(`<html><body><p> © Company 2014-2019 </p></body></html>`, opts)
	assert.Equal(t, "2019-01-01", format(dt))
}

func Test_searchPattern(t *testing.T) {
	// Variables
	opts := Options{MinDate: defaultMinDate, MaxDate: defaultMaxDate}

	// First pattern, YYYY MM
	pattern := regexp.MustCompile(`\D([0-9]{4}[/.-][0-9]{2})\D`)
	catchPattern := regexp.MustCompile(`([0-9]{4})[/.-]([0-9]{2})`)
	yearPattern := regexp.MustCompile(`^([12][0-9]{3})`)

	str := "It happened on the 202.E.19, the day when it all began."
	res := searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.Empty(t, res)

	str = "The date is 2002.02.15."
	res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.NotEmpty(t, res)
	assert.Equal(t, "2002.02", res[0])

	str = "http://www.url.net/index.html"
	res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.Empty(t, res)

	str = "http://www.url.net/2016/01/index.html"
	res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.NotEmpty(t, res)
	assert.Equal(t, "2016/01", res[0])

	// Second pattern, MM YYYY
	pattern = regexp.MustCompile(`\D([0-9]{2}[/.-][0-9]{4})\D`)
	catchPattern = regexp.MustCompile(`([0-9]{2})[/.-]([0-9]{4})`)
	yearPattern = regexp.MustCompile(`([12][0-9]{3})$`)

	str = "It happened on the 202.E.19, the day when it all began."
	res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.Empty(t, res)

	str = "It happened on the 15.02.2002, the day when it all began."
	res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.NotEmpty(t, res)
	assert.Equal(t, "02.2002", res[0])

	// Third pattern, YYYY only
	pattern = regexp.MustCompile(`\D(2[01][0-9]{2})\D`)
	catchPattern = regexp.MustCompile(`(2[01][0-9]{2})`)
	yearPattern = regexp.MustCompile(`^(2[01][0-9]{2})`)

	str = "It happened in the film 300."
	res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.Empty(t, res)

	str = "It happened in 2002."
	res = searchPattern(str, pattern, catchPattern, yearPattern, opts)
	assert.NotEmpty(t, res)
	assert.Equal(t, "2002", res[0])
}
