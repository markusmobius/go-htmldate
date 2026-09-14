package htmldate

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-shiori/dom"
	"github.com/markusmobius/go-htmldate/internal/selector"
	"golang.org/x/net/html"
)

type prunedDocument struct {
	root          *html.Node
	alreadyPruned bool
}

func (view prunedDocument) skip(node *html.Node) bool {
	return !view.alreadyPruned && node != view.root && node.Type == html.ElementNode &&
		(isCleanedTag(node.Data) || selector.Discard(node))
}

func (view prunedDocument) walk(root *html.Node, visit func(*html.Node) bool) {
	for node := root; node != nil; {
		if !view.skip(node) {
			if !visit(node) {
				return
			}
			if node.FirstChild != nil {
				node = node.FirstChild
				continue
			}
		}
		for node != root && node.NextSibling == nil {
			node = node.Parent
		}
		if node == root {
			return
		}
		node = node.NextSibling
	}
}

func (view prunedDocument) elements() []*html.Node {
	return view.queryAll(func(*html.Node) bool { return true })
}

func (view prunedDocument) queryAll(rule selector.Rule) []*html.Node {
	var elements []*html.Node
	view.walk(view.root, func(node *html.Node) bool {
		if node != view.root && node.Type == html.ElementNode && rule(node) {
			elements = append(elements, node)
		}
		return true
	})
	return elements
}

func (view prunedDocument) tagged(tag string) []*html.Node {
	return view.queryAll(func(node *html.Node) bool { return node.Data == tag })
}

func (view prunedDocument) querySelectorAll(selectors string) []*html.Node {
	elements := dom.QuerySelectorAll(view.root, selectors)
	if view.alreadyPruned {
		return elements
	}
	visible := elements[:0]
	for _, element := range elements {
		hidden := false
		for ancestor := element; ancestor != nil && ancestor != view.root; ancestor = ancestor.Parent {
			if view.skip(ancestor) {
				hidden = true
				break
			}
		}
		if !hidden {
			visible = append(visible, element)
		}
	}
	return visible
}

func (view prunedDocument) queryAllTextNodes(rule selector.Rule) []*html.Node {
	var matches []*html.Node
	for _, element := range view.queryAll(rule) {
		for child := element.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.TextNode && child.Data != "" {
				matches = append(matches, child)
			}
		}
	}
	return matches
}

func (view prunedDocument) text(node *html.Node) string {
	if view.alreadyPruned {
		return dom.TextContent(node)
	}
	var text strings.Builder
	view.walk(node, func(current *html.Node) bool {
		if current.Type == html.TextNode {
			text.WriteString(current.Data)
		}
		return true
	})
	return text.String()
}

func (view prunedDocument) initialText(node *html.Node) string {
	if view.alreadyPruned {
		return etreeText(node)
	}
	if node == nil {
		return ""
	}
	var text strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if view.skip(child) {
			continue
		}
		if child.Type == html.ElementNode {
			break
		}
		if child.Type == html.TextNode {
			text.WriteString(child.Data)
		}
	}
	return text.String()
}

func (view prunedDocument) outerHTML(node *html.Node) string {
	if view.alreadyPruned {
		return dom.OuterHTML(node)
	}
	var output bytes.Buffer
	if _, err := view.render(&output, node); err != nil {
		return ""
	}
	return output.String()
}

func (view prunedDocument) innerHTML(node *html.Node) string {
	if view.alreadyPruned {
		return dom.InnerHTML(node)
	}
	var output bytes.Buffer
	if node != nil {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if _, err := view.render(&output, child); err != nil {
				return ""
			}
		}
	}
	return strings.TrimSpace(output.String())
}

func (view prunedDocument) render(output *bytes.Buffer, node *html.Node) (bool, error) {
	if node == nil || view.skip(node) {
		return false, nil
	}
	if node.Type != html.ElementNode && node.Type != html.DocumentNode {
		return false, html.Render(output, node)
	}
	var closing string
	var literal bool
	if node.Type == html.ElementNode {
		shell := html.Node{Type: node.Type, DataAtom: node.DataAtom, Data: node.Data, Attr: node.Attr}
		var probe *html.Node
		switch node.Data {
		case "iframe", "noembed", "noframes", "noscript", "plaintext", "script", "style", "xmp":
			probe = &html.Node{Type: html.TextNode, Data: "<"}
			shell.FirstChild, shell.LastChild = probe, probe
		}
		start := output.Len()
		if err := html.Render(output, &shell); err != nil {
			return false, err
		}
		suffix := "</" + node.Data + ">"
		if bytes.HasSuffix(output.Bytes()[start:], []byte(suffix)) {
			closing = suffix
			output.Truncate(output.Len() - len(suffix))
		}
		if probe != nil {
			literal = output.Bytes()[output.Len()-1] == '<'
			if literal {
				output.Truncate(output.Len() - 1)
			} else {
				output.Truncate(output.Len() - len("&lt;"))
			}
		}
		first := node.FirstChild
		for first != nil && view.skip(first) {
			first = first.NextSibling
		}
		if closing == "" && probe == nil {
			if first != nil {
				return false, fmt.Errorf("html: void element <%s> has child nodes", node.Data)
			}
			return false, nil
		}
		if first != nil && first.Type == html.TextNode && strings.HasPrefix(first.Data, "\n") {
			switch node.Data {
			case "pre", "listing", "textarea":
				output.WriteByte('\n')
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if literal && child.Type == html.TextNode {
			output.WriteString(child.Data)
		} else if stopped, err := view.render(output, child); stopped || err != nil {
			return stopped, err
		}
	}
	output.WriteString(closing)
	return node.Type == html.ElementNode && closing == "", nil
}
