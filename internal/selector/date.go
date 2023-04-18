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

// .//*[(self::div or self::li or self::p or self::span)], then date selector.
func FastDate(n *html.Node) bool {
	tagName := dom.TagName(n)

	switch tagName {
	case "div", "li", "p", "span":
	case "footer", "small":
		return true
	default:
		return false
	}

	return dateRule(n)
}

// [contains(translate(@id, "D", "d"), 'date')
// or contains(translate(@class, "D", "d"), 'date')
// or contains(translate(@itemprop, "D", "d"), 'date')
// or contains(translate(@id, "D", "d"), 'datum')
// or contains(translate(@class, "D", "d"), 'datum')
// or contains(@id, 'time') or contains(@class, 'time')
// or @class='meta' or contains(translate(@id, "M", "m"), 'metadata')
// or contains(translate(@class, "M", "m"), 'meta-')
// or contains(translate(@class, "M", "m"), '-meta')
// or contains(translate(@id, "M", "m"), '-meta')
// or contains(translate(@class, "M", "m"), '_meta')
// or contains(translate(@class, "M", "m"), 'postmeta')
// or contains(@class, 'info') or contains(@class, 'post_detail')
// or contains(@class, 'block-content')
// or contains(@class, 'byline') or contains(@class, 'subline')
// or contains(@class, 'posted') or contains(@class, 'submitted')
// or contains(@class, 'created-post')
// or contains(@id, 'publish') or contains(@class, 'publish')
// or contains(@class, 'publication')
// or contains(@class, 'author') or contains(@class, 'autor')
// or contains(@class, 'field-content')
// or contains(@class, 'fa-clock-o') or contains(@class, 'fa-calendar')
// or contains(@class, 'fecha') or contains(@class, 'parution')
// or contains(@class, 'footer') or contains(@id, 'footer')]
// |
// .//footer|.//small`
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

	switch {
	case strings.Contains(strings.ToLower(id), "date"),
		strings.Contains(strings.ToLower(class), "date"),
		strings.Contains(strings.ToLower(itemProp), "date"),
		strings.Contains(strings.ToLower(id), "datum"),
		strings.Contains(strings.ToLower(class), "datum"),
		strings.Contains(id, "time"),
		strings.Contains(class, "time"),
		class == "meta",
		strings.Contains(strings.ToLower(id), "metadata"),
		strings.Contains(strings.ToLower(class), "meta-"),
		strings.Contains(strings.ToLower(class), "-meta"),
		strings.Contains(strings.ToLower(id), "-meta"),
		strings.Contains(strings.ToLower(class), "_meta"),
		strings.Contains(strings.ToLower(class), "postmeta"),
		strings.Contains(class, "info"),
		strings.Contains(class, "post_detail"),
		strings.Contains(class, "block-content"),
		strings.Contains(class, "byline"),
		strings.Contains(class, "subline"),
		strings.Contains(class, "posted"),
		strings.Contains(class, "submitted"),
		strings.Contains(class, "created-post"),
		strings.Contains(id, "publish"),
		strings.Contains(class, "publish"),
		strings.Contains(class, "publication"),
		strings.Contains(class, "author"),
		strings.Contains(class, "autor"),
		strings.Contains(class, "field-content"),
		strings.Contains(class, "fa-clock-o"),
		strings.Contains(class, "fa-calendar"),
		strings.Contains(class, "fecha"),
		strings.Contains(class, "parution"),
		strings.Contains(class, "footer"),
		strings.Contains(id, "footer"):
	default:
		return false
	}

	return true
}
