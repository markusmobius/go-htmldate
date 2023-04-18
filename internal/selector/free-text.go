package selector

import (
	"github.com/go-shiori/dom"
	"golang.org/x/net/html"
)

// .//*[(self::div or self::li or self::p or self::span)]/text()
func FreeText(n *html.Node) bool {
	switch dom.TagName(n) {
	case "div", "li", "p", "span":
		return true
	default:
		return false
	}
}
