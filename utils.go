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
	"bytes"
	"strings"
	"unicode"

	"github.com/go-shiori/dom"
	"github.com/markusmobius/go-htmldate/internal/regexp"
	"golang.org/x/net/html"
)

// cleanDocument cleans the document by discarding unwanted elements.
func cleanDocument(doc *html.Node) *html.Node {
	// Clone doc
	clone := dom.Clone(doc, true)

	// Remove comments
	// removeHtmlCommentNode(clone)

	// Remove useless nodes
	tagNames := []string{
		// Embed elements
		"object", "embed", "applet",
		// Frame elements
		"frame", "frameset", "noframes", "iframe",
		// Others
		"label", "map", "math",
		"audio", "canvas", "datalist",
		"picture", "rdf", "svg", "track", "video",
		// TODO: to be considered
		// "figure", "input", "layer", "param", "source"
	}

	for _, node := range dom.GetAllNodesWithTag(clone, tagNames...) {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}

	return clone
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

func rxFindNamedStringSubmatch(rx *regexp.Regexp, s string) (map[string]string, string) {
	names := rx.SubexpNames()
	result := make(map[string]string)
	matches := rx.FindStringSubmatch(s)

	var lastMatchedName string
	for i, match := range matches {
		if i > 0 && match != "" {
			result[names[i]] = match
			lastMatchedName = names[i]
		}
	}

	return result, lastMatchedName
}

// isLeapYear check if year is leap year.
func isLeapYear(year int) bool {
	// If year is not divisible by 4, then it is not a leap year
	if year%4 != 0 {
		return false
	}

	// If year is not divisible by 100, then it is a leap year
	if year%100 != 0 {
		return true
	}

	// If year is not divisible by 400, then it is not a leap year
	if year%400 != 0 {
		return false
	}

	// If all passed, it's leap year
	return true
}
