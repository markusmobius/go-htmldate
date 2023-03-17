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
	"os"
	"path/filepath"
	"strings"
)

// openMockFile is used to open HTML document from specified mock file.
// Make sure to close the reader later.
func openMockFile(url string) io.ReadCloser {
	// Open file
	path := mapMockFiles[url]
	path = filepath.Join("test-files", "mock", path)

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	return f
}

// extractMockFile open then extract content from a mock file.
func extractMockFile(url string, customOpts ...Options) Result {
	// Open mock file
	f := openMockFile(url)
	defer f.Close()

	// Extract
	opts := Options{}
	if len(customOpts) > 0 {
		opts = mergeOpts(opts, customOpts[0])
	}

	result, err := FromReader(f, opts)
	if err != nil {
		panic(err)
	}

	return result
}

func extractFromString(s string, customOpts ...Options) Result {
	opts := Options{}
	if len(customOpts) > 0 {
		opts = mergeOpts(opts, customOpts[0])
	}

	r := strings.NewReader(s)
	result, err := FromReader(r, opts)
	if err != nil {
		panic(err)
	}

	return result
}

func extractFromURL(url string, customOpts ...Options) Result {
	opts := Options{URL: url}
	if len(customOpts) > 0 {
		opts = mergeOpts(opts, customOpts[0])
	}

	r := strings.NewReader("")
	result, err := FromReader(r, opts)
	if err != nil {
		panic(err)
	}

	return result
}

func mergeOpts(opt1, opt2 Options) Options {
	opt1.EnableLog = opt1.EnableLog || opt2.EnableLog
	opt1.ExtractTime = opt1.ExtractTime || opt2.ExtractTime
	opt1.UseOriginalDate = opt1.UseOriginalDate || opt2.UseOriginalDate
	opt1.SkipExtensiveSearch = opt1.SkipExtensiveSearch || opt2.SkipExtensiveSearch
	opt1.DeferUrlExtractor = opt1.DeferUrlExtractor || opt2.DeferUrlExtractor

	if opt2.URL != "" {
		opt1.URL = opt2.URL
	}

	if !opt2.MinDate.IsZero() {
		opt1.MinDate = opt2.MinDate
	}

	if !opt2.MaxDate.IsZero() {
		opt1.MaxDate = opt2.MaxDate
	}

	return opt1
}

var mapMockFiles = map[string]string{
	"http://blog.kinra.de/?p=959/":                                              "kinra.de.html",
	"http://blog.python.org/2016/12/python-360-is-now-available.html":           "blog.python.org.html",
	"http://blog.todamax.net/2018/midp-emulator-kemulator-und-brick-challenge/": "blog.todamax.net.html",
	"http://carta.info/der-neue-trend-muss-statt-wunschkoalition/":              "carta.info.html",
	"https://500px.com/photo/26034451/spring-in-china-by-alexey-kruglov":        "500px.com.spring.html",
	"https://bayern.de/":                 "bayern.de.html",
	"https://creativecommons.org/about/": "creativecommons.org.html",
	"https://die-partei.net/sh/":         "die-partei.net.sh.html",
	"https://en.blog.wordpress.com/":     "blog.wordpress.com.html",
	"https://en.support.wordpress.com/":  "support.wordpress.com.html",
	"https://futurezone.at/digital-life/wie-creativecommons-richtig-genutzt-wird/24.600.504": "futurezone.at.cc.html",
	"https://github.com/adbar/htmldate": "github.com.html",
	"https://netzpolitik.org/2016/die-cider-connection-abmahnungen-gegen-nutzer-von-creative-commons-bildern/": "netzpolitik.org.abmahnungen.html",
	"https://pixabay.com/en/service/terms/":                   "pixabay.com.tos.html",
	"https://www.austria.info/":                               "austria.info.html",
	"https://www.befifty.de/home/2017/7/12/unter-uns-montauk": "befifty.montauk.html",
	"https://www.beltz.de/fachmedien/paedagogik/didacta_2019_in_koeln_19_23_februar/beltz_veranstaltungen_didacta_2016/veranstaltung.html?tx_news_pi1%5Bnews%5D=14392&tx_news_pi1%5Bcontroller%5D=News&tx_news_pi1%5Baction%5D=detail&cHash=10b1a32fb5b2b05360bdac257b01c8fa": "beltz.de.didakta.html",
	"https://www.channelpartner.de/a/sieben-berufe-die-zukunft-haben,3050673": "channelpartner.de.berufe.html",
	"https://www.creativecommons.at/faircoin-hackathon":                       "creativecommons.at.faircoin.html",
	"https://www.deutschland.de/en":                                           "deutschland.de.en.html",
	"https://www.eff.org/files/annual-report/2015/index.html":                 "eff.org.2015.html",
	"https://www.facebook.com/visitaustria/":                                  "facebook.com.visitaustria.html",
	"https://www.gnu.org/licenses/gpl-3.0.en.html":                            "gnu.org.gpl.html",
	"https://www.goodform.ch/blog/schattiges_plaetzchen":                      "goodform.ch.blog.html",
	"https://www.horizont.net/marketing/kommentare/influencer-marketing-was-sich-nach-dem-vreni-frost-urteil-aendert-und-aendern-muss-172529":                         "horizont.net.html",
	"https://www.intel.com/content/www/us/en/legal/terms-of-use.html":                                                                                                 "intel.com.tos.html",
	"https://www.pferde-fuer-unsere-kinder.de/unsere-projekte/":                                                                                                       "pferde.projekte.de.html",
	"https://www.rosneft.com/business/Upstream/Licensing/":                                                                                                            "rosneft.com.licensing.html",
	"https://www.scs78.de/news/items/warm-war-es-schoen-war-es.html":                                                                                                  "scs78.de.html",
	"https://www.tagesausblick.de/Analyse/USA/DOW-Jones-Jahresendrally-ade__601.html":                                                                                 "tagesausblick.de.dow.html",
	"https://www.transgen.de/aktuell/2687.afrikanische-schweinepest-genome-editing.html":                                                                              "transgen.de.aktuell.html",
	"https://www.weltwoche.ch/ausgaben/2019-4/artikel/forbes-die-weltwoche-ausgabe-4-2019.html":                                                                       "weltwoche.ch.html",
	"https://www.wunderweib.de/manuela-reimann-hochzeitsueberraschung-in-bayern-107930.html":                                                                          "wunderweib.html",
	"http://unexpecteduser.blogspot.de/2011/":                                                                                                                         "unexpecteduser.2011.html",
	"http://viehbacher.com/de/spezialisierung/internationale-forderungsbeitreibung":                                                                                   "viehbacher.com.forderungsbetreibung.html",
	"http://www.eza.gv.at/das-ministerium/presse/aussendungen/2018/07/aussenministerin-karin-kneissl-beim-treffen-der-deutschsprachigen-aussenminister-in-luxemburg/": "eza.gv.at.html",
	"http://www.freundeskreis-videoclips.de/waehlen-sie-car-player-tipps-zur-auswahl-der-besten-car-cd-player/":                                                       "freundeskreis-videoclips.de.html",
	"http://www.greenpeace.org/international/en/campaigns/forests/asia-pacific/":                                                                                      "greenpeace.org.forests.html",
	"http://www.heimicke.de/chronik/zahlen-und-daten/":                                                                                                                "heimicke.de.zahlen.html",
	"http://www.hobby-werkstatt-blog.de/arduino/424-eine-arduino-virtual-wall-fuer-den-irobot-roomba.php":                                                             "hobby-werkstatt-blog.de.roomba.html",
	"http://www.hundeverein-kreisunna.de/termine.html":                                                                                                                "hundeverein-kreisunna.de.html",
	"http://www.hundeverein-querfurt.de/index.php?option=com_content&view=article&id=54&Itemid=50":                                                                    "hundeverein-querfurt.de.html",
	"http://www.jovelstefan.de/2012/05/11/parken-in-paris/":                                                                                                           "jovelstefan.de.parken.html",
	"http://www.klimawandel-global.de/klimaschutz/energie-sparen/elektromobilitat-der-neue-trend/":                                                                    "klimawandel-global.de.html",
	"http://www.medef.com/en/content/alternative-dispute-resolution-for-antitrust-damages":                                                                            "medef.fr.dispute.html",
	"http://www.pbrunst.de/news/2011/12/kein-cyberterrorismus-diesmal/":                                                                                               "pbrunst.de.html",
	"http://www.stuttgart.de/": "stuttgart.de.html",
	"https://paris-luttes.info/quand-on-comprend-que-les-grenades-12355":                                                    "paris-luttes.info.html",
	"https://www.brigitte.de/aktuell/riverdale--so-ehrt-die-serie-luke-perry-in-staffel-vier-11602344.html":                 "brigitte.de.riverdale.html",
	"https://www.ldt.de/ldtblog/fall-in-love-with-black/":                                                                   "ldt.de.fallinlove.html",
	"http://www.loldf.org/spip.php?article717":                                                                              "loldf.org.html",
	"https://www.beltz.de/sachbuch_ratgeber/buecher/produkt_produktdetails/37219-12_wege_zu_guter_pflege.html":              "beltz.de.12wege.html",
	"https://www.oberstdorf-resort.de/interaktiv/blog/unser-kraeutergarten-wannenkopfhuette.html":                           "oberstdorfresort.de.kraeuter.html",
	"https://www.wienbadminton.at/news/119843/Come-Together":                                                                "wienbadminton.at.html",
	"https://blog.wikimedia.org/2018/06/28/interactive-maps-now-in-your-language/":                                          "blog.wikimedia.interactivemaps.html",
	"https://blogs.mediapart.fr/elba/blog/260619/violences-policieres-bombe-retardement-mediatique":                         "mediapart.fr.violences.html",
	"https://verfassungsblog.de/the-first-decade/":                                                                          "verfassungsblog.de.decade.html",
	"https://cric-grenoble.info/infos-locales/article/putsh-en-cours-a-radio-kaleidoscope-1145":                             "cric-grenoble.info.radio.html",
	"https://www.sebastian-kurz.at/magazin/wasserstoff-als-schluesseltechnologie":                                           "kurz.at.wasserstoff.html",
	"https://la-bas.org/la-bas-magazine/chroniques/Didier-Porte-souhaite-la-Sante-a-Balkany":                                "la-bas.org.porte.html",
	"https://exporo.de/wiki/europaeische-zentralbank-ezb/":                                                                  "exporo.de.ezb.html",
	"https://www.revolutionpermanente.fr/Antonin-Bernanos-en-prison-depuis-pres-de-deux-mois-en-raison-de-son-militantisme": "revolutionpermanente.fr.antonin.html",
	"http://www.wara-enforcement.org/guinee-un-braconnier-delephant-interpelle-et-condamne-a-la-peine-maximale/":            "wara-enforcement.org.guinee.html",
	"https://ebene11.com/die-arbeit-mit-fremden-dwg-dateien-in-autocad":                                                     "ebene11.com.autocad.html",
	"https://www.acredis.com/schoenheitsoperationen/augenlidstraffung/":                                                     "acredis.com.augenlidstraffung.html",
	"https://www.hertie-school.org/en/debate/detail/content/whats-on-the-cards-for-von-der-leyen/":                          "hertie-school.org.leyen.html",
	"https://www.adac.de/rund-ums-fahrzeug/tests/kindersicherheit/kindersitztest-2018/":                                     "adac.de.kindersitztest.html",
	"http://web.archive.org/web/20210916140120/https://www.kath.ch/die-insel-der-klosterzoeglinge/":                         "archive.org.kath.ch.html",
}
