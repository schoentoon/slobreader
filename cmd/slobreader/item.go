package main

import (
	"bytes"
	"fmt"
	"strings"

	goslob "github.com/schoentoon/go-slob"
	"golang.org/x/net/html"
)

type WordEntry struct {
	Input         string
	Pronunciation string
	Output        []*Word
}

type Word struct {
	Word   string
	Gender string
}

func hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key != "class" {
			continue
		}

		for _, classname := range strings.Split(attr.Val, " ") {
			if classname == class {
				return true
			}
		}
	}

	return false
}

func parseWord(n *html.Node) (*Word, error) {
	out := &Word{}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.FirstChild != nil {
			switch n.Data {
			case "div":
				if hasClass(n, "gen") {
					out.Gender = n.FirstChild.Data
				}
			case "li":
				if hasClass(n, "quote") {
					out.Word = n.FirstChild.Data
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	if out.Word == "" {
		return nil, fmt.Errorf("No word found")
	}

	return out, nil
}

func ParseItem(item *goslob.Item) (*WordEntry, error) {
	doc, err := html.Parse(bytes.NewReader(item.Content))
	if err != nil {
		return nil, err
	}
	out := &WordEntry{
		Output: make([]*Word, 0),
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.FirstChild != nil {
			switch n.Data {
			case "div":
				if hasClass(n, "orth") {
					out.Input = n.FirstChild.Data
				} else if hasClass(n, "pron") {
					out.Pronunciation = n.FirstChild.Data
				}
			case "li":
				if hasClass(n, "sense") {
					word, err := parseWord(n)
					if err == nil {
						out.Output = append(out.Output, word)
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return out, nil
}

func (w *WordEntry) Render(cfg *Config) string {
	out := fmt.Sprintf("# %s\n\n", w.Input)

	for _, word := range w.Output {
		if word.Gender != "" {
			gender := cfg.Gender(word.Gender)
			if gender != "" {
				out += fmt.Sprintf("* %s %s\n", gender, word.Word)
			} else {
				out += fmt.Sprintf("* %s (%s)\n", word.Word, word.Gender)
			}
		} else {
			out += fmt.Sprintf("* %s\n", word.Word)
		}
	}

	return out
}
