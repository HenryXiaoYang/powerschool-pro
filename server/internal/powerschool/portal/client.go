// Package portal drives the PowerSchool parent/student web portal (cookie session),
// used only for what the mobile API lacks: archived report cards.
package portal

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool"
)

const maxPage = 8 << 20

// Session is the portal's cookie set after a successful sign-in.
type Session struct {
	Cookies []Cookie `json:"cookies"`
}

type Cookie struct {
	Name  string `json:"n"`
	Value string `json:"v"`
}

// Client talks to one school's portal. It never follows redirects off that host.
type Client struct {
	origin    *url.URL
	transport http.RoundTripper
}

func New(origin string, httpClient *http.Client) (*Client, error) {
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return nil, errors.New("portal: origin must be an absolute URL")
	}
	c := &Client{origin: &url.URL{Scheme: u.Scheme, Host: u.Host}}
	if httpClient != nil {
		c.transport = httpClient.Transport
	}
	return c, nil
}

// browser returns an HTTP client with a fresh jar holding s's cookies.
func (c *Client) browser(s Session) *http.Client {
	jar, _ := cookiejar.New(nil)
	cookies := make([]*http.Cookie, len(s.Cookies))
	for i, ck := range s.Cookies {
		cookies[i] = &http.Cookie{Name: ck.Name, Value: ck.Value, Path: "/"}
	}
	jar.SetCookies(c.origin, cookies)
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: c.transport,
		Jar:       jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Host != c.origin.Host || len(via) > 5 {
				return errors.New("portal: refused redirect")
			}
			return nil
		},
	}
}

// Login signs in through the portal's form. The login page sends the plain password
// over HTTPS in pw, dbpw and ldappassword (see /admin/javascript/signin-script.js).
func (c *Client) Login(ctx context.Context, username, password string) (Session, error) {
	hc := c.browser(Session{})
	if _, _, err := c.get(ctx, hc, "/public/home.html"); err != nil {
		return Session{}, err
	}
	form := url.Values{
		"account": {username}, "pw": {password}, "dbpw": {password}, "ldappassword": {password},
		"serviceName": {"PS Parent Portal"}, "pcasServerUrl": {"/"},
		"credentialType":      {"User Id and Password Credential"},
		"translator_username": {""}, "translator_password": {""}, "translator_ldappassword": {""},
		"returnUrl": {""}, "serviceTicket": {""}, "translatorpw": {""},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.origin.String()+"/guardian/home.html", strings.NewReader(form.Encode()))
	if err != nil {
		return Session{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	final, body, err := c.do(hc, req)
	if err != nil {
		return Session{}, err
	}
	if isLoginPage(final, body) {
		return Session{}, powerschool.ErrInvalidCredentials
	}
	var s Session
	for _, ck := range hc.Jar.Cookies(c.origin) {
		s.Cookies = append(s.Cookies, Cookie{ck.Name, ck.Value})
	}
	return s, nil
}

// page fetches an authenticated portal page.
func (c *Client) page(ctx context.Context, s Session, path string) (*http.Client, string, error) {
	hc := c.browser(s)
	final, body, err := c.get(ctx, hc, path)
	if err != nil {
		return nil, "", err
	}
	if isLoginPage(final, body) {
		return nil, "", powerschool.ErrSessionExpired
	}
	return hc, body, nil
}

func (c *Client) get(ctx context.Context, hc *http.Client, path string) (*url.URL, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.origin.String()+path, nil)
	if err != nil {
		return nil, "", err
	}
	return c.do(hc, req)
}

func (c *Client) do(hc *http.Client, req *http.Request) (*url.URL, string, error) {
	resp, err := hc.Do(req)
	if err != nil {
		if req.Context().Err() != nil {
			return nil, "", req.Context().Err()
		}
		return nil, "", powerschool.Upstreamf("cannot reach the PowerSchool portal")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", powerschool.Upstreamf("portal returned HTTP %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxPage))
	if err != nil {
		return nil, "", powerschool.Upstreamf("portal: reading response failed")
	}
	return resp.Request.URL, string(b), nil
}

func isLoginPage(final *url.URL, body string) bool {
	return strings.HasPrefix(final.Path, "/public/") || strings.Contains(body, `id="LoginForm"`)
}
