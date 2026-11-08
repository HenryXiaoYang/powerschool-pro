// Package mobile speaks the PowerSchool mobile app's SOAP-in, JSON-out protocol
// (PublicPortalServiceJSON), as documented in powerschool-api-analysis/PROTOCOL.md.
package mobile

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool"
)

const (
	servicePath = "/pearson-rest/services/PublicPortalServiceJSON?response=application/json"
	soapNS      = "http://schemas.xmlsoap.org/soap/envelope/"
	portalNS    = "http://publicportal.rest.powerschool.pearson.com/xsd"
	maxBody     = 16 << 20

	// Public app constants shipped in the PowerSchool Android APK; not user secrets.
	DefaultServiceUser     = "pearson"
	DefaultServicePassword = "m0bApP5"
)

// Client calls one school's mobile service endpoint.
type Client struct {
	endpoint string
	uri      string // request target used in the Digest response
	http     *http.Client
	digest   *digestAuth
}

// New returns a client for origin (e.g. https://school.example.org). httpClient may be nil.
func New(origin string, httpClient *http.Client, serviceUser, servicePassword string) (*Client, error) {
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return nil, errors.New("mobile: origin must be an absolute URL")
	}
	hc := &http.Client{Timeout: 30 * time.Second}
	if httpClient != nil {
		hc.Transport = httpClient.Transport
	}
	// Redirects could forward credentials to an HTML login page or another host.
	hc.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{
		endpoint: u.Scheme + "://" + u.Host + servicePath,
		uri:      servicePath,
		http:     hc,
		digest:   newDigestAuth(serviceUser, servicePassword),
	}, nil
}

// Login authenticates the portal account and returns its session.
func (c *Client) Login(ctx context.Context, username, password string) (Session, error) {
	var res results
	if err := c.call(ctx, "loginToPublicPortal", &res, leaf("username", username), leaf("password", password)); err != nil {
		return Session{}, err
	}
	if res.UserSessionVO == nil || res.UserSessionVO.ServiceTicket == "" {
		return Session{}, powerschool.ErrInvalidCredentials
	}
	if len(res.UserSessionVO.StudentIDs) == 0 {
		return Session{}, powerschool.Upstreamf("account has no accessible students")
	}
	return *res.UserSessionVO, nil
}

// StudentData returns every accessible student's full record (qil ALL).
func (c *Client) StudentData(ctx context.Context, s Session) ([]StudentData, error) {
	params := append(sessionParams(s), group("qil", leaf("includes", "1")))
	var res results
	if err := c.call(ctx, "getStudentData", &res, params...); err != nil {
		return nil, err
	}
	if len(res.StudentDataVOs) == 0 {
		// The server answers an expired ticket with an empty result, not an error.
		return nil, powerschool.ErrSessionExpired
	}
	return res.StudentDataVOs, nil
}

// SectionTimes returns each student's sections with meeting times for one month.
func (c *Client) SectionTimes(ctx context.Context, s Session, year, month int) ([]StudentData, error) {
	params := append(sessionParams(s), leaf("month", strconv.Itoa(month)), leaf("year", strconv.Itoa(year)))
	var res results
	if err := c.call(ctx, "getStartStopTimeForAllSections", &res, params...); err != nil {
		return nil, err
	}
	if len(res.StudentDataVOs) == 0 {
		return nil, powerschool.ErrSessionExpired
	}
	return res.StudentDataVOs, nil
}

// sessionParams serializes the minimal session the tested server accepts, in WSDL order,
// followed by the explicit student IDs.
func sessionParams(s Session) []node {
	var fields []node
	if s.ServerCurrentTime.Valid {
		fields = append(fields, leaf("serverCurrentTime", s.ServerCurrentTime.Value))
	}
	// Echoing the whole serverInfo object made the server return HTTP 500.
	fields = append(fields, group("serverInfo", leaf("apiVersion", s.ServerInfo.APIVersion.Value)))
	fields = append(fields, leaf("serviceTicket", s.ServiceTicket))
	if s.UserID.Valid {
		fields = append(fields, leaf("userId", s.UserID.Value))
	}
	if s.UserType.Valid {
		fields = append(fields, leaf("userType", s.UserType.Value))
	}
	params := []node{group("userSessionVO", fields...)}
	for _, id := range s.StudentIDs {
		params = append(params, leaf("studentIDs", strconv.FormatInt(id, 10)))
	}
	return params
}

func (c *Client) call(ctx context.Context, method string, out *results, params ...node) error {
	body := envelope(method, params)
	resp, err := c.post(ctx, method, body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized && c.digest.setChallenge(resp.Header.Get("WWW-Authenticate")) {
		resp.Body.Close()
		if resp, err = c.post(ctx, method, body); err != nil {
			return err
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return powerschool.Upstreamf("%s returned HTTP %d", method, resp.StatusCode)
	}
	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return powerschool.Upstreamf("%s: reading response failed", method)
	}
	if len(payload) > maxBody {
		return powerschool.Upstreamf("%s: oversized response", method)
	}
	return decodeResult(method, payload, out)
}

func (c *Client) post(ctx context.Context, method string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Close = true
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "urn:"+method)
	req.Header.Set("User-Agent", "PSMobileApp")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("stamp-of-time", time.Now().UTC().Format("2006-01-02 15:04:05")+" +00000")
	if auth := c.digest.authorization(http.MethodPost, c.uri); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, powerschool.Upstreamf("cannot reach PowerSchool")
	}
	return resp, nil
}

// decodeResult unwraps <method>Response.return after normalizing Axis metadata.
func decodeResult(method string, payload []byte, out *results) error {
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	var tree any
	if err := dec.Decode(&tree); err != nil {
		return powerschool.Upstreamf("%s: response is not JSON", method)
	}
	root, _ := normalize(tree).(map[string]any)
	wrapper, _ := root[method+"Response"].(map[string]any)
	ret, ok := wrapper["return"].(map[string]any)
	if !ok {
		return powerschool.Upstreamf("%s: unexpected response or SOAP fault", method)
	}
	b, err := json.Marshal(ret)
	if err == nil {
		err = json.Unmarshal(b, out)
	}
	if err != nil {
		return powerschool.Upstreamf("%s: malformed data", method)
	}
	return nil
}

// node is a minimal XML element: text when kids is nil.
type node struct {
	name string
	text string
	kids []node
}

func leaf(name, text string) node          { return node{name: name, text: text} }
func group(name string, kids ...node) node { return node{name: name, kids: kids} }

func (n node) write(b *bytes.Buffer) {
	b.WriteString("<" + n.name + ">")
	if n.kids == nil {
		_ = xml.EscapeText(b, []byte(n.text))
	}
	for _, k := range n.kids {
		k.write(b)
	}
	b.WriteString("</" + n.name + ">")
}

// envelope builds a SOAP 1.1 request. The operation is namespace-qualified; its
// parameters are not (elementFormDefault="unqualified").
func envelope(method string, params []node) []byte {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="utf-8"?><soap:Envelope xmlns:soap="` + soapNS + `"><soap:Body><p:` + method + ` xmlns:p="` + portalNS + `">`)
	for _, p := range params {
		p.write(&b)
	}
	b.WriteString(`</p:` + method + `></soap:Body></soap:Envelope>`)
	return b.Bytes()
}
