package portal

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool"
)

const listPage = `<html><body><table>
<tr class="archiveRow"><td><a href="/guardian/rcBuilder/savedReportCard.html?rcID=42">Y10 S1 Report Card</a></td><td>01/30/2026</td><td></td></tr>
<tr class="archiveRow"><td><a href="/guardian/rcBuilder/savedReportCard.html?rcID=41">Y10 M1 Report Card</a></td><td>11/14/2025</td></tr>
<tr><td><a href="/guardian/rcBuilder/savedReportCard.html?rcID=41">duplicate</a></td></tr>
<tr><td><a href="/guardian/rcBuilder/savedReportCard.html?rcID=1;evil">bad id</a></td></tr>
</table></body></html>`

const cardPage = `<html><head>
<style>.keep{color:red}</style>
<script>$j(document).ready(function(){ if(1 == 1){ $j('style:last').remove(); } });</script>
</head><body>
<div id="buttonDiv"><button>Print</button></div>
<div id="reportCardContainer">
  <p class="keep" onclick="steal()">Grades</p>
  <img src="/images/logo.png"><img src="https://elsewhere.example/track.png">
  <a href="javascript:alert(1)">x</a><script>alert(1)</script>
</div>
<style>.default{}</style>
</body></html>`

func fakePortal(t *testing.T) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggedIn := false
		if c, err := r.Cookie("JSESSIONID"); err == nil && c.Value == "ok" {
			loggedIn = true
		}
		switch {
		case r.URL.Path == "/public/home.html":
			io.WriteString(w, `<form id="LoginForm"></form>`)
		case r.Method == http.MethodPost && r.URL.Path == "/guardian/home.html":
			r.ParseForm()
			if r.Form.Get("pw") != "secret" {
				http.Redirect(w, r, "/public/home.html", http.StatusFound)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "JSESSIONID", Value: "ok", Path: "/"})
			io.WriteString(w, "welcome")
		case !loggedIn:
			http.Redirect(w, r, "/public/home.html", http.StatusFound)
		case r.URL.Path == "/guardian/reportCardsSelect.html":
			io.WriteString(w, listPage)
		case r.URL.Path == "/guardian/rcBuilder/savedReportCard.html":
			io.WriteString(w, cardPage)
		case r.URL.Path == "/images/logo.png":
			w.Header().Set("Content-Type", "image/png")
			w.Write([]byte("PNG"))
		}
	}))
}

func TestPortalFlow(t *testing.T) {
	srv := fakePortal(t)
	defer srv.Close()
	c, _ := New(srv.URL, srv.Client())
	ctx := context.Background()

	if _, err := c.Login(ctx, "a", "nope"); !errors.Is(err, powerschool.ErrInvalidCredentials) {
		t.Fatalf("bad login: %v", err)
	}
	s, err := c.Login(ctx, "a", "secret")
	if err != nil {
		t.Fatal(err)
	}
	refs, err := c.ReportCards(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 2 || refs[0] != (ReportCardRef{"42", "Y10 S1 Report Card", "01/30/2026"}) || refs[1].ID != "41" {
		t.Fatalf("refs: %+v", refs)
	}

	page, err := c.ReportCard(ctx, s, "42")
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"<script", "onclick", "javascript:", "elsewhere.example", "Print", ".default"} {
		if strings.Contains(page, bad) {
			t.Errorf("sanitized page still contains %q", bad)
		}
	}
	for _, good := range []string{".keep{color:red}", "Grades", "data:image/png;base64,UE5H"} {
		if !strings.Contains(page, good) {
			t.Errorf("sanitized page lost %q", good)
		}
	}

	if _, err := c.ReportCards(ctx, Session{}); !errors.Is(err, powerschool.ErrSessionExpired) {
		t.Errorf("expired: %v", err)
	}
	if _, err := c.ReportCard(ctx, s, "../x"); !errors.Is(err, powerschool.ErrNotFound) {
		t.Errorf("bad id: %v", err)
	}
}
