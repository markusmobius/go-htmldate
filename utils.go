package htmldate

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"unicode"

	"github.com/go-shiori/dom"
	"github.com/gogs/chardet"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	xunicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// cleanDocument cleans the document by discarding unwanted elements.
func cleanDocument(doc *html.Node) {
	// Remove comments
	removeHtmlCommentNode(doc)

	// Remove useless nodes
	tagNames := []string{
		// Embed elements
		"object", "embed", "applet",
		// Frame elements
		"frame", "iframe",
		// Others
		"audio", "canvas", "datalist",
		"figure", "label", "map", "math",
		"picture", "rdf", "svg", "video",
	}

	for _, node := range dom.GetAllNodesWithTag(doc, tagNames...) {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

// removeHtmlCommentNode removes all `html.CommentNode` in document.
func removeHtmlCommentNode(doc *html.Node) {
	// Find all comment nodes
	var finder func(*html.Node)
	var commentNodes []*html.Node

	finder = func(node *html.Node) {
		if node.Type == html.CommentNode {
			commentNodes = append(commentNodes, node)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			finder(child)
		}
	}

	for child := doc.FirstChild; child != nil; child = child.NextSibling {
		finder(child)
	}

	// Remove it
	dom.RemoveNodes(commentNodes, nil)
}

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

// isDigit check if string only consisted of digit number.
func isDigit(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

// getDigitCount returns count of digit number in the specified string.
func getDigitCount(s string) int {
	var nDigit int
	for _, r := range s {
		if unicode.IsDigit(r) {
			nDigit++
		}
	}
	return nDigit
}

// etreeText returns texts before first subelement. If there was no text,
// this function will returns an empty string.
func etreeText(element *html.Node) string {
	if element == nil {
		return ""
	}

	buffer := bytes.NewBuffer(nil)
	for child := element.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			break
		} else if child.Type == html.TextNode {
			buffer.WriteString(child.Data)
		}
	}

	return buffer.String()
}

// inMap check if keys exist in map.
func inMap(key string, mapString map[string]struct{}) bool {
	_, exist := mapString[key]
	return exist
}

// strIn check if string exists in slice.
func strIn(s string, args ...string) bool {
	for _, arg := range args {
		if s == arg {
			return true
		}
	}
	return false
}

// strLimit cut a string until the specified limit.
func strLimit(s string, limit int) string {
	if len(s) > limit {
		s = s[:limit]
	}

	return s
}

// normalizeSpaces converts all whitespaces to normal spaces, remove multiple adjacent
// whitespaces and trim the string.
func normalizeSpaces(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}
