package mobile

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool"
)

func TestDigestRFC2617Vector(t *testing.T) {
	d := newDigestAuth("Mufasa", "Circle Of Life")
	d.cnonce = func() string { return "0a4f113b" }
	if !d.setChallenge(`Digest realm="testrealm@host.com", qop="auth,auth-int", nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093", opaque="5ccc069c403ebaf9f0171e9517f40e41"`) {
		t.Fatal("challenge rejected")
	}
	got := d.authorization("GET", "/dir/index.html")
	if !strings.Contains(got, `response="6629fae49393a05397450978507c4ef1"`) || !strings.Contains(got, "nc=00000001") {
		t.Fatalf("authorization = %s", got)
	}
	if strings.Contains(d.authorization("GET", "/dir/index.html"), "nc=00000001") {
		t.Fatal("nonce count must advance")
	}
}

func TestDecodeShapes(t *testing.T) {
	payload := `{"getStudentDataResponse":{"return":{"@type":"x","studentDataVOs":{
		"studentId":1,
		"sections":{"id":10,"expression":"1(MA)","roomName":{"@nil":"true"}},
		"periods":[{"id":5,"name":1},null,{"id":6,"name":"BF"}],
		"assignmentScores":[{"assignmentId":7,"score":"--","percent":null},{"assignmentId":8,"score":0,"percent":"85.5","comment":42}]
	}}}}`
	var res results
	if err := decodeResult("getStudentData", []byte(payload), &res); err != nil {
		t.Fatal(err)
	}
	d := res.StudentDataVOs[0]
	if len(d.Sections) != 1 || d.Sections[0].RoomName.Valid || d.Sections[0].Expression.Value != "1(MA)" {
		t.Errorf("singleton section / nil room: %+v", d.Sections)
	}
	if len(d.Periods) != 2 || d.Periods[0].Name.Value != "1" || d.Periods[1].Name.Value != "BF" {
		t.Errorf("periods: %+v", d.Periods)
	}
	s := d.AssignmentScores
	if s[0].Score.Value != "--" || s[0].Percent.Valid {
		t.Errorf("non-numeric score: %+v", s[0])
	}
	if !s[1].Score.Valid || s[1].Score.Value != "0" || s[1].Percent.Value != 85.5 || s[1].Comment.Value != "42" {
		t.Errorf("zero score must stay a score: %+v", s[1])
	}
	if err := decodeResult("getStudentData", []byte(`{"Fault":{}}`), &res); !errors.Is(err, powerschool.ErrUpstream) {
		t.Errorf("fault: %v", err)
	}
}

// fakeService challenges once, then answers login and getStudentData.
func fakeService(t *testing.T, studentData string) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Digest ") {
			w.Header().Set("WWW-Authenticate", `Digest realm="Protected", qop="auth", nonce="n1"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		body, _ := io.ReadAll(r.Body)
		switch r.Header.Get("SOAPAction") {
		case "urn:loginToPublicPortal":
			if !strings.Contains(string(body), "<password>p&amp;w&lt;</password>") {
				io.WriteString(w, `{"loginToPublicPortalResponse":{"return":{"userSessionVO":{"serviceTicket":{"@nil":true}}}}}`)
				return
			}
			io.WriteString(w, `{"loginToPublicPortalResponse":{"return":{"userSessionVO":{"serviceTicket":"T","studentIDs":[1,2],"userId":9,"userType":2,"serverInfo":{"apiVersion":"3.0.0"}}}}}`)
		case "urn:getStudentData":
			if !strings.Contains(string(body), "<studentIDs>1</studentIDs><studentIDs>2</studentIDs><qil><includes>1</includes></qil>") {
				t.Errorf("request body: %s", body)
			}
			io.WriteString(w, `{"getStudentDataResponse":{"return":`+studentData+`}}`)
		}
	}))
}

func TestClientLoginAndData(t *testing.T) {
	srv := fakeService(t, `{"studentDataVOs":{"studentId":1}}`)
	defer srv.Close()
	c, err := New(srv.URL, srv.Client(), "u", "p")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := c.Login(ctx, "a", "wrong"); !errors.Is(err, powerschool.ErrInvalidCredentials) {
		t.Fatalf("wrong password: %v", err)
	}
	s, err := c.Login(ctx, "a", "p&w<")
	if err != nil {
		t.Fatal(err)
	}
	if s.ServiceTicket != "T" || len(s.StudentIDs) != 2 || s.UserID.Value != "9" {
		t.Fatalf("session: %+v", s)
	}
	// The session survives a JSON round trip (it travels inside the sealed token).
	b, _ := json.Marshal(s)
	var back Session
	if err := json.Unmarshal(b, &back); err != nil || back.UserID != s.UserID || back.ServerInfo != s.ServerInfo {
		t.Fatalf("round trip: %v %+v", err, back)
	}
	data, err := c.StudentData(ctx, back)
	if err != nil || data[0].StudentID != 1 {
		t.Fatalf("data: %v %+v", err, data)
	}
}

func TestClientExpiredTicket(t *testing.T) {
	srv := fakeService(t, `{"studentDataVOs":{"@nil":true}}`)
	defer srv.Close()
	c, _ := New(srv.URL, srv.Client(), "u", "p")
	_, err := c.StudentData(context.Background(), Session{ServiceTicket: "old", StudentIDs: List[int64]{1, 2}})
	if !errors.Is(err, powerschool.ErrSessionExpired) {
		t.Fatalf("got %v", err)
	}
}
