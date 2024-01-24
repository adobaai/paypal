package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/exp/slices"
)

const (
	maxColumn = 100
	tabWidth  = 4
)

var (
	NotFound       = errors.New("not found")
	TooManyMatches = errors.New("too many matches")
)

type WebhookGroup struct {
	Title       string
	Description Comment
	Webhooks    map[string][]Webhook // Key is version
}

type Webhook struct {
	Repeated      bool
	ID            string
	Event         string
	Trigger       Comment
	RelatedMethod Comment
}

func (wh *Webhook) IsRef() bool {
	return bytes.HasPrefix(wh.Trigger.Content, []byte("See"))
}

type Comment struct {
	Content []byte
	Links   []Link
}

type Link struct {
	Title string
	URL   string
}

func sortedRangeMap[K ~string, V any](m map[K]V, f func(k K, v V)) {
	var keys []K
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, key := range keys {
		f(key, m[key])
	}
}

var allEvents [][]byte

func parsePayPal(bs []byte) (res []WebhookGroup, err error) {
	var i = 0
	var hookCount = map[string]int{}
	var step string
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %w", step, err)
		}
	}()
	for {
		var titleBytes, descBytes, webhooksBytes []byte
		var title, desc, version []byte
		titleBytes = findRegion(bs, []byte("<h2 "), []byte("</h2>"), &i)
		if titleBytes == nil {
			break
		}

		title, err = parseTitle(titleBytes)
		if err != nil {
			step = "title"
			break
		}

		descBytes = findRegion(bs, []byte("<p>"), []byte("</p>"), &i)
		if descBytes != nil {
			// Description is optional
			desc, err = parseDescription(descBytes)
			if err != nil {
				step = "description"
				break
			}
		}

		step = "webhooks"
		group := WebhookGroup{
			Title:       string(title),
			Description: NewParser(desc).Comment(),
			Webhooks:    map[string][]Webhook{},
		}
		res = append(res, group)

		// Find versions
		for {
			versionBytes := findRegion(bs, []byte("<h3 "), []byte("</h3>"), &i)
			if versionBytes == nil {
				break
			}

			version, err = getFirstMatch(reVersion, versionBytes)
			if err != nil {
				break
			}

			webhooksBytes := findRegion(bs, []byte("<tbody>"), []byte("</tbody>"), &i)
			if webhooksBytes == nil {
				break
			}

			var webhooks []Webhook
			webhooks, err = parseWebhooks(webhooksBytes, hookCount)
			if err != nil {
				break
			}

			group.Webhooks[string(version)] = webhooks
		}

		webhooksBytes = findRegion(bs, []byte("<tbody>"), []byte("</tbody>"), &i)
		if webhooksBytes == nil {
			continue
		}

		var webhooks []Webhook
		webhooks, err = parseWebhooks(webhooksBytes, hookCount)
		if err != nil {
			break
		} else {
			group.Webhooks[""] = webhooks
		}
	}
	return
}

func findRegion(bs []byte, start []byte, end []byte, i *int) []byte {
	startIndex := findBytes(bs[*i:], start)
	if startIndex == -1 {
		return nil
	}
	endDiff := findBytes(bs[*i+startIndex:], end)
	if endDiff == -1 {
		return nil
	}
	nextH2 := findBytes(bs[*i:], []byte("<h2 "))

	if nextH2 != 0 && nextH2 != startIndex && nextH2 < startIndex+endDiff {
		return nil
	}

	i2 := *i
	*i += startIndex + endDiff + len(end)
	return bs[i2+startIndex : *i]
}

func findBytes(bs []byte, pattern []byte) (index int) {
	for i := 0; i < len(bs); i++ {
		if bs[i] == pattern[0] && bytes.HasPrefix(bs[i:], pattern) {
			return i
		}
	}
	return -1
}

var (
	reTitle    = regexp.MustCompile(`(?s)<h2.*?>.+<\/div>\s*(.+?)\s*<\/h2>`)
	reDesc     = regexp.MustCompile(`(?s)<p>\s*(.+)\s*<\/p>`)
	reVersion  = regexp.MustCompile(`(?s)<h3.+div>\s*(.+?)\s*<\/h3>`)
	reWebhooks = regexp.MustCompile(`(?s)<tr>.+?<\/tr>`)
	reWebhook  = regexp.MustCompile(`(?s)<tr>\s*<td.*>\s*<code.*>(.+)<\/code\s*>\s*<\/td>\s*<td.*?>\s*(.+)\s*<\/td>\s*<td.*?>\s*(.*?)\s*<\/td>`)
	reA        = regexp.MustCompile(`(?s)<a.*href="(.+)".*>\s*(.*?)\s*<\/a.*>`)
	reCode     = regexp.MustCompile(`(?s)<code.*?>(.+?)<\/code.*?>`)
	reACode    = regexp.MustCompile(`(?s)<a.*?>\s*<code.*?>(\S+).*?>\s*<\/a\s*>`)
)

func getFirstMatch(r *regexp.Regexp, bs []byte) (res []byte, err error) {
	matches := r.FindSubmatch(bs)
	if matches == nil {
		return nil, NotFound
	}
	return matches[1], nil
}

func parseTitle(bs []byte) (res []byte, err error) {
	return getFirstMatch(reTitle, bs)
}

func parseDescription(bs []byte) (res []byte, err error) {
	res, err = getFirstMatch(reDesc, bs)
	res = NewParser(res).RemoveWhitespaces().Bytes()
	if res[len(res)-1] == ':' {
		res = res[:len(res)-1]
	}
	return
}

func parseWebhooks(bs []byte, hookCount map[string]int) (res []Webhook, err error) {
	matches := reWebhooks.FindAll(bs, -1)
	for _, m := range matches {
		webhook, err := parseWebhook(m)
		if err != nil {
			return nil, err
		}
		count := hookCount[webhook.Event]
		if !webhook.IsRef() {
			hookCount[webhook.Event]++
		}
		webhook.Repeated = count != 0
		res = append(res, webhook)
	}
	return
}

func parseWebhook(bs []byte) (Webhook, error) {
	matches := reWebhook.FindSubmatch(bs)
	if matches == nil {
		return Webhook{}, fmt.Errorf("no match")
	}

	allEvents = append(allEvents, matches[1])

	return Webhook{
		ID:    string(eventToGolangIdentifier(matches[1])),
		Event: string(matches[1]),
		Trigger: NewParser(matches[2]).ReplaceACode().
			ReplaceCode().RemoveWhitespaces().Comment(),
		RelatedMethod: NewParser(matches[3]).ReplaceCode().RemoveWhitespaces().Comment(),
	}, nil
}

func wrap(bs []byte, length int) (res [][]byte) {
	if len(bs) == 0 {
		return nil
	}
	if len(bs) <= length {
		return [][]byte{bs}
	}

	lastSpace := 0
	for i := 0; i < len(bs); i++ {
		if bs[i] == ' ' {
			if i > length {
				res = append(res, bs[:lastSpace])
				res = append(res, wrap(bs[lastSpace+1:], length)...)
				return
			}
			lastSpace = i
		}
	}
	res = append(res, bs[:lastSpace])
	res = append(res, bs[lastSpace+1:])
	return
}

func eventToGolangIdentifier(s []byte) (res []byte) {
	for _, sp := range bytes.Split(s, []byte(".")) {
		for _, sp2 := range bytes.Split(sp, []byte("-")) {
			res = append(res, byte(unicode.ToTitle(rune(sp2[0]))))
			res = append(res, bytes.ToLower(sp2[1:])...)
		}
	}
	return
}

type Parser struct {
	bs []byte
}

func NewParser(bs []byte) *Parser {
	return &Parser{
		bs: bs,
	}
}

func (p *Parser) Bytes() []byte {
	return p.bs
}

func (p *Parser) String() string {
	return string(p.bs)
}

func (p *Parser) Comment() (res Comment) {
	old := []byte("<strong>Deprecation notice</strong>")
	bs := bytes.ReplaceAll(p.bs, old, []byte("Deprecated"))
	t := []byte{}
	links := reA.FindAll(bs, -1)
	if links == nil {
		res.Content = bs
		return
	}

	for _, link := range links {
		matches := reA.FindSubmatch(link)
		if matches == nil {
			t = append(t, link...)
			continue
		}
		link := []byte("[")
		link = append(link, matches[2]...)
		link = append(link, ']')
		t = append(t, reA.ReplaceAll(bs, link)...)
		res.Links = append(res.Links, Link{
			Title: string(matches[2]),
			URL:   docBaseURL + string(matches[1]),
		})
	}
	res.Content = t
	return
}

func (p *Parser) ReplaceCode() (res *Parser) {
	res = p
	codes := reCode.FindAll(p.bs, -1)
	if codes == nil {
		return
	}

	for _, code := range codes {
		matches := reCode.FindSubmatch(code)
		if matches == nil {
			continue
		}

		p.bs = reCode.ReplaceAll(p.bs, []byte("`$1`"))
	}
	return
}

func (p *Parser) ReplaceACode() (res *Parser) {
	res = p
	matches := reACode.FindAll(p.bs, -1)
	if matches == nil {
		return
	}

	if len(matches) > 1 {
		return
	}

	submatches := reCode.FindSubmatch(p.bs)
	repl := append([]byte("["), eventToGolangIdentifier(submatches[1])...)
	repl = append(repl, ']')
	p.bs = reACode.ReplaceAll(p.bs, repl)
	return
}

func (p *Parser) RemoveWhitespaces() (res *Parser) {
	res = p
	if len(p.bs) == 0 {
		return
	}

	var last byte = ' '
	var tmp []byte
	for _, c := range p.bs {
		if isWhitespace(c) {
			if !isWhitespace(last) {
				tmp = append(tmp, ' ')
			}
		} else {
			tmp = append(tmp, c)
		}
		last = c
	}
	if isWhitespace(last) {
		tmp = tmp[:len(tmp)-1]
	}
	p.bs = tmp
	return
}

func isWhitespace(r byte) bool {
	switch r {
	case ' ', '\n', '\t':
		return true
	default:
		return false
	}
}

type Generator struct {
	Package string
	sb      strings.Builder
}

func (g *Generator) Build(gs []WebhookGroup) string {
	g.AppendHeader()
	for _, wg := range gs {
		g.AppendGroup(wg)
	}
	return g.sb.String()
}

func (g *Generator) WriteFile(gs []WebhookGroup, file string) error {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	_, err = f.WriteString(g.Build(gs))
	return err
}

func (g *Generator) AppendHeader() {
	g.Writeln("// Code generated by paypalc. DO NOT EDIT.")
	g.Writef("package %s\n", g.Package)
	g.Writeln()
	g.Writeln("// EventType is the type of webhook.")
	g.Writeln("//")
	g.Writeln("// See https://developer.paypal.com/api/rest/webhooks/event-names/")
	g.Writeln("type EventType string")
	g.Writeln()
}

func (g *Generator) AppendGroup(wg WebhookGroup) {
	g.Writeln()
	g.Writeln("// " + wg.Title)
	if len(wg.Description.Content) > 0 {
		g.Writeln("//")
		g.AppendComment(wg.Description, 0)
	}
	g.Writeln()
	g.AppendWebhooks(wg.Webhooks)
}

func (g *Generator) AppendWebhooks(webhooks map[string][]Webhook) {
	sortedRangeMap(webhooks, func(version string, v []Webhook) {
		if version != "" {
			g.Writeln("// " + version)
			g.Writeln()
		}
		g.sb.WriteString("const (")
		for _, hook := range v {
			g.Writeln()
			trigger := hook.Trigger
			trigger.Links = nil
			hook.RelatedMethod.Links = append(hook.Trigger.Links, hook.RelatedMethod.Links...)
			g.AppendComment(trigger, 1)
			g.AppendRelatedMethod(hook.RelatedMethod)
			if hook.Repeated || hook.IsRef() {
				g.WritefIndent(1, "// (redeclared) %s EventType = \"%s\"\n", hook.ID, hook.Event)
			} else {
				g.WritefIndent(1, "%s EventType = \"%s\"\n", hook.ID, hook.Event)
			}
		}
		g.Writeln(")")
		g.Writeln()
	})
}

func (g *Generator) AppendRelatedMethod(rm Comment) {
	if len(rm.Content) == 0 {
		return
	}
	g.WritelnIndent(1, "//")
	rm.Content = append([]byte("Related method: "), rm.Content...)
	g.AppendComment(rm, 1)
}

func (g *Generator) AppendComment(t Comment, tabs int) {
	lines := wrap([]byte(t.Content), maxColumn-tabs*tabWidth-3)
	for _, line := range lines {
		g.WritelnIndent(tabs, "// "+string(line))
	}
	if len(t.Links) > 0 {
		g.WritelnIndent(tabs, "//")
		for _, link := range t.Links {
			g.WritefIndent(tabs, "// [%s]: %s\n", link.Title, link.URL)
		}
	}
}

func (g *Generator) Writef(format string, args ...any) {
	g.sb.WriteString(fmt.Sprintf(format, args...))
}

func (g *Generator) Writeln(a ...any) {
	g.sb.WriteString(fmt.Sprintln(a...))
}

func (g *Generator) WritefIndent(indent int, format string, args ...any) {
	prefix := strings.Repeat("	", indent)
	g.Writef(prefix+format, args...)
}

func (g *Generator) WritelnIndent(indent int, a ...any) {
	prefix := strings.Repeat("	", indent)
	g.sb.WriteString(prefix + fmt.Sprintln(a...))
}
