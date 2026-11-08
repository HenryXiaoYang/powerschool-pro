package portal

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool"
)

// ReportCardRef is one archived report card on reportCardsSelect.html.
type ReportCardRef struct {
	ID   string
	Name string
	Date string // as the portal prints it, MM/DD/YYYY
}

const reportCardPath = "/guardian/rcBuilder/savedReportCard.html?rcID="

var digits = regexp.MustCompile(`^[0-9]{1,12}$`)

// ReportCards lists the student's archived report cards, newest first as the portal orders them.
func (c *Client) ReportCards(ctx context.Context, s Session) ([]ReportCardRef, error) {
	_, body, err := c.page(ctx, s, "/guardian/reportCardsSelect.html")
	if err != nil {
		return nil, err
	}
	return parseReportCardList(body)
}

func parseReportCardList(body string) ([]ReportCardRef, error) {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, powerschool.Upstreamf("portal: unparseable report list")
	}
	var refs []ReportCardRef
	seen := map[string]bool{}
	for a := range doc.Descendants() {
		if a.DataAtom != atom.A {
			continue
		}
		href := attr(a, "href")
		i := strings.Index(href, reportCardPath)
		if i < 0 {
			continue
		}
		id := href[i+len(reportCardPath):]
		if !digits.MatchString(id) || seen[id] {
			continue
		}
		seen[id] = true
		ref := ReportCardRef{ID: id, Name: text(a)}
		// The date is the next cell in the same row.
		if td := a.Parent; td != nil && td.DataAtom == atom.Td {
			for sib := td.NextSibling; sib != nil; sib = sib.NextSibling {
				if sib.DataAtom == atom.Td {
					ref.Date = text(sib)
					break
				}
			}
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

// ReportCard returns one report card as a standalone, script-free HTML document.
// Same-host images are inlined so the page renders without a portal cookie.
func (c *Client) ReportCard(ctx context.Context, s Session, id string) (string, error) {
	if !digits.MatchString(id) {
		return "", powerschool.ErrNotFound
	}
	hc, body, err := c.page(ctx, s, reportCardPath+id)
	if err != nil {
		return "", err
	}
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return "", powerschool.Upstreamf("portal: unparseable report card")
	}
	return c.standalone(ctx, hc, doc)
}

func (c *Client) standalone(ctx context.Context, hc *http.Client, doc *html.Node) (string, error) {
	var container *html.Node
	var styles []*html.Node
	dropLastStyles := 0
	for n := range doc.Descendants() {
		switch {
		case n.DataAtom == atom.Div && attr(n, "id") == "reportCardContainer" && container == nil:
			container = n
		case n.DataAtom == atom.Style:
			styles = append(styles, n)
		case n.DataAtom == atom.Script:
			// The page's own script removes its trailing default styles when stored styles
			// exist ("if(1 == 1){ $j('style:last').remove(); ... }"); scripts are stripped, so mirror it.
			if js := text(n); strings.Contains(js, "if(1 == 1)") {
				dropLastStyles = strings.Count(js, "$j('style:last').remove()")
			}
		}
	}
	if container == nil {
		return "", powerschool.ErrNotFound
	}
	styles = styles[:max(0, len(styles)-dropLastStyles)]

	var out bytes.Buffer
	out.WriteString(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">`)
	for _, st := range styles {
		out.WriteString("<style>" + strings.ReplaceAll(rawText(st), "</", `<\/`) + "</style>")
	}
	out.WriteString(`<style>@media print{@page{margin:0}}body{margin:0;background:#fff}</style></head><body>`)

	c.scrub(ctx, hc, container)
	if err := html.Render(&out, container); err != nil {
		return "", err
	}
	out.WriteString("</body></html>")
	return out.String(), nil
}

var dropped = map[atom.Atom]bool{
	atom.Script: true, atom.Noscript: true, atom.Iframe: true, atom.Object: true, atom.Embed: true,
	atom.Form: true, atom.Input: true, atom.Button: true, atom.Select: true, atom.Textarea: true,
	atom.Link: true, atom.Meta: true, atom.Base: true,
}

// scrub removes active content and inlines same-host images in place.
func (c *Client) scrub(ctx context.Context, hc *http.Client, root *html.Node) {
	var remove []*html.Node
	for n := range root.Descendants() {
		if n.Type != html.ElementNode {
			continue
		}
		if dropped[n.DataAtom] {
			remove = append(remove, n)
			continue
		}
		kept := n.Attr[:0]
		for _, a := range n.Attr {
			key := strings.ToLower(a.Key)
			v := strings.ToLower(strings.TrimSpace(a.Val))
			if strings.HasPrefix(key, "on") || strings.HasPrefix(v, "javascript:") || key == "srcset" {
				continue
			}
			if key == "href" {
				continue // links point into the portal; the mirror has no use for them
			}
			if key == "src" && n.DataAtom == atom.Img {
				a.Val = c.inlineImage(ctx, hc, a.Val)
				if a.Val == "" {
					continue
				}
			}
			kept = append(kept, a)
		}
		n.Attr = kept
	}
	for _, n := range remove {
		n.Parent.RemoveChild(n)
	}
}

// inlineImage turns a same-host image URL into a data URI, or "" to drop it.
func (c *Client) inlineImage(ctx context.Context, hc *http.Client, src string) string {
	if strings.HasPrefix(src, "data:image/") {
		return src
	}
	u, err := c.origin.Parse(src)
	if err != nil || u.Host != c.origin.Host {
		return ""
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ""
	}
	resp, err := hc.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	mt, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(mt, "image/") || mt == "image/svg+xml" {
		return ""
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return ""
	}
	return "data:" + mt + ";base64," + base64.StdEncoding.EncodeToString(b)
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func rawText(n *html.Node) string {
	var b strings.Builder
	for d := range n.Descendants() {
		if d.Type == html.TextNode {
			b.WriteString(d.Data)
		}
	}
	return b.String()
}

// text returns n's text content with whitespace collapsed.
func text(n *html.Node) string { return strings.Join(strings.Fields(rawText(n)), " ") }
