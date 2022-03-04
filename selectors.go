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
	"strings"

	"github.com/go-shiori/dom"
	"golang.org/x/net/html"
)

type selectorRule func(*html.Node) bool

func findElementsWithRule(doc *html.Node, rule selectorRule) []*html.Node {
	var elements []*html.Node
	for _, elem := range dom.GetElementsByTagName(doc, "*") {
		if rule(elem) {
			elements = append(elements, elem)
		}
	}

	return elements
}

func dateSelectorRule(n *html.Node) bool {
	id := strings.ToLower(dom.GetAttribute(n, "id"))
	class := strings.ToLower(dom.GetAttribute(n, "class"))
	itemProp := strings.ToLower(dom.GetAttribute(n, "itemprop"))
	tagName := dom.TagName(n)

	switch {
	// Rule 1
	case strings.Contains(id, "date"),
		strings.Contains(id, "datum"),
		strings.Contains(id, "time"),
		strings.Contains(class, "post-meta-time"):

	// Rule 2
	case strings.Contains(class, "date"),
		strings.Contains(class, "datum"):

	// Rule 3
	case strings.Contains(class, "postmeta"),
		strings.Contains(class, "post-meta"),
		strings.Contains(class, "post_meta"),
		strings.Contains(class, "post__meta"),
		strings.Contains(class, "entry-meta"),
		strings.Contains(class, "entry-date"),
		strings.Contains(class, "article__date"),
		strings.Contains(class, "post_detail"),
		class == "meta",
		class == "meta-before",
		class == "asset-meta",
		strings.Contains(id, "article-metadata"),
		strings.Contains(class, "article-metadata"),
		strings.Contains(class, "block-content"),
		strings.Contains(class, "byline"),
		strings.Contains(class, "dateline"),
		strings.Contains(class, "subline"),
		strings.Contains(class, "published"),
		strings.Contains(class, "posted"),
		strings.Contains(class, "submitted"),
		strings.Contains(class, "updated"),
		strings.Contains(class, "created-post"),
		strings.Contains(id, "post-timestamp"),
		strings.Contains(class, "post-timestamp"):

	// Rule 4
	case strings.Contains(id, "lastmod"),
		strings.Contains(itemProp, "date"),
		strings.Contains(class, "time"),
		strings.Contains(id, "metadata"),
		strings.Contains(id, "publish"):

	// Rule 5
	case tagName == "footer",
		class == "post-footer",
		class == "footer",
		id == "footer":

	// Rule 6
	case tagName == "small":

	// Rule 7
	case strings.Contains(class, "author"),
		strings.Contains(class, "autor"),
		strings.Contains(class, "field-content"),
		class == "meta",
		strings.Contains(class, "info"),
		strings.Contains(class, "fa-clock-o"),
		strings.Contains(class, "fa-calendar"),
		strings.Contains(class, "publication"):

	default:
		return false
	}

	return true
}

func additionalSelectorRule(n *html.Node) bool {
	class := strings.ToLower(dom.GetAttribute(n, "class"))

	switch {
	case strings.Contains(class, "fecha"),
		strings.Contains(class, "parution"):
		return true
	default:
		return false
	}
}

func discardSelectorRule(n *html.Node) bool {
	id := strings.ToLower(dom.GetAttribute(n, "id"))
	class := strings.ToLower(dom.GetAttribute(n, "class"))
	tagName := dom.TagName(n)

	switch {
	// Rule 1
	case tagName == "footer":
	// Rule 2
	case (tagName == "div" || tagName == "section") && (id == "footer" || class == "footer"):
	// Rule 3
	case tagName == "div" && (id == "wm-ipp-base" || id == "wm-ipp"):

	default:
		return false
	}

	return true
}
