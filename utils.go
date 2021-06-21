package htmldate

import (
	"bytes"
	"io"
	"io/ioutil"
	"unicode"

	"github.com/gogs/chardet"
	"github.com/pingcap/parser/charset"
	"golang.org/x/net/html"
	xunicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// normalizeTextEncoding convert text encoding from NFD to NFC.
// It also remove soft hyphen since apparently it's useless in web.
// See: https://web.archive.org/web/19990117011731/http://www.hut.fi/~jkorpela/shy.html
func normalizeTextEncoding(r io.Reader) io.Reader {
	fnSoftHyphen := func(r rune) bool { return r == '\u00AD' }
	softHyphenSet := runes.Predicate(fnSoftHyphen)
	transformer := transform.Chain(norm.NFD, runes.Remove(softHyphenSet), norm.NFC)
	return transform.NewReader(r, transformer)
}

// parseHTMLDocument parses a reader and try to convert the character encoding into UTF-8.
func parseHTMLDocument(r io.Reader) (*html.Node, error) {
	// Split the reader using tee
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Detect page encoding
	res, err := chardet.NewHtmlDetector().DetectBest(content)
	if err != nil {
		return nil, err
	}

	pageEncoding, _ := charset.Lookup(res.Charset)
	if pageEncoding == nil {
		pageEncoding = xunicode.UTF8
	}

	// Parse HTML using the page encoding
	r = bytes.NewReader(content)
	r = transform.NewReader(r, pageEncoding.NewDecoder())
	r = normalizeTextEncoding(r)
	return html.Parse(r)
}

func isDigit(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func inMap(key string, mapString map[string]struct{}) bool {
	_, exist := mapString[key]
	return exist
}

func strIn(s string, args ...string) bool {
	for _, arg := range args {
		if s == arg {
			return true
		}
	}
	return false
}
