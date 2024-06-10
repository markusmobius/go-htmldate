package selector

import (
	"strings"

	"github.com/go-shiori/dom"
	"golang.org/x/net/html"
)

// `.//*`, then date selector.
func SlowDate(n *html.Node) bool {
	return dateRule(n)
}

// .//*[(self::div or self::h2 or self::h3 or self::h4 or self::li or self::p or self::span or self::time or self::ul)], then date selector.
func FastDate(n *html.Node) bool {
	tagName := dom.TagName(n)

	switch tagName {
	case "div", "h2", "h3", "h4", "li", "p", "span", "time", "ul":
	default:
		return false
	}

	return dateRule(n)
}

// [
// contains(translate(@id|@class|@itemprop, "D", "d"), 'date') or
// contains(translate(@id|@class|@itemprop, "D", "d"), 'datum') or
// contains(translate(@id|@class, "M", "m"), 'meta') or
// contains(@id|@class, 'time') or
// contains(@id|@class, 'publish') or
// contains(@id|@class, 'footer') or
// contains(@class, 'info') or
// contains(@class, 'post_detail') or
// contains(@class, 'block-content') or
// contains(@class, 'byline') or
// contains(@class, 'subline') or
// contains(@class, 'posted') or
// contains(@class, 'submitted') or
// contains(@class, 'created-post') or
// contains(@class, 'publication') or
// contains(@class, 'author') or
// contains(@class, 'autor') or
// contains(@class, 'field-content') or
// contains(@class, 'fa-clock-o') or
// contains(@class, 'fa-calendar') or
// contains(@class, 'fecha') or
// contains(@class, 'parution')
// ] |
// .//footer | .//small
//
// Further tests needed:
// or contains(@class, 'article')
// or contains(@class, 'footer') or contains(@id, 'footer')
// or contains(@id, 'lastmod') or contains(@class, 'updated')
func dateRule(n *html.Node) bool {
	id := dom.ID(n)
	class := dom.ClassName(n)
	tagName := dom.TagName(n)
	itemProp := dom.GetAttribute(n, "itemprop")

	switch tagName {
	case "footer", "small":
		return true
	}

	lowId := strings.ToLower(id)
	lowClass := strings.ToLower(class)
	lowItemProp := strings.ToLower(itemProp)
	switch {
	case strings.Contains(lowId, "date"),
		strings.Contains(lowClass, "date"),
		strings.Contains(lowItemProp, "date"),
		strings.Contains(lowId, "datum"),
		strings.Contains(lowClass, "datum"),
		strings.Contains(lowItemProp, "datum"),
		strings.Contains(lowId, "meta"),
		strings.Contains(lowClass, "meta"),
		strings.Contains(id, "time"),
		strings.Contains(class, "time"),
		strings.Contains(lowId, "publish"),
		strings.Contains(lowClass, "publish"),
		strings.Contains(lowId, "footer"),
		strings.Contains(lowClass, "footer"),
		strings.Contains(class, "info"),
		strings.Contains(class, "post_detail"),
		strings.Contains(class, "block-content"),
		strings.Contains(class, "byline"),
		strings.Contains(class, "subline"),
		strings.Contains(class, "posted"),
		strings.Contains(class, "submitted"),
		strings.Contains(class, "created-post"),
		strings.Contains(class, "publication"),
		strings.Contains(class, "author"),
		strings.Contains(class, "autor"),
		strings.Contains(class, "field-content"),
		strings.Contains(class, "fa-clock-o"),
		strings.Contains(class, "fa-calendar"),
		strings.Contains(class, "fecha"),
		strings.Contains(class, "parution"):
	default:
		return false
	}

	return true
}
