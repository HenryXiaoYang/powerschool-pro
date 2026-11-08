package httpapi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/fake"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/session"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/student"
)

func testServer(t *testing.T) (*gin.Engine, *session.Sealer) {
	gin.SetMode(gin.TestMode)
	sealer, err := session.NewSealer(bytes.Repeat([]byte{7}, 32), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	svc := student.NewService(fake.Mobile{}, fake.Portal{}, time.UTC)
	return New(svc, sealer, Options{}), sealer
}

func do(r http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func errorCode(t *testing.T, w *httptest.ResponseRecorder) ErrorCode {
	var body ErrorBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("not an error body: %s", w.Body)
	}
	return body.Error.Code
}

func login(t *testing.T, r http.Handler) string {
	w := do(r, http.MethodPost, "/api/session", "", `{"username":"alex","password":"pw"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("login: %d %s", w.Code, w.Body)
	}
	var res LoginResponse
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res.Token == "" || !res.ReportCards {
		t.Fatalf("login response: %+v", res)
	}
	return res.Token
}

func TestAPIFlow(t *testing.T) {
	r, _ := testServer(t)
	token := login(t, r)

	w := do(r, http.MethodGet, "/api/student", token, "")
	if w.Code != http.StatusOK || w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatalf("student: %d %v", w.Code, w.Header())
	}
	var snap student.Snapshot
	if err := json.Unmarshal(w.Body.Bytes(), &snap); err != nil || len(snap.Courses) == 0 {
		t.Fatalf("snapshot: %v", err)
	}

	now := time.Now()
	w = do(r, http.MethodGet, "/api/schedule?year="+now.Format("2006")+"&month="+now.Format("1"), token, "")
	if w.Code != http.StatusOK {
		t.Fatalf("schedule: %d %s", w.Code, w.Body)
	}

	w = do(r, http.MethodGet, "/api/report-cards", token, "")
	var list ReportCardList
	if _ = json.Unmarshal(w.Body.Bytes(), &list); w.Code != http.StatusOK || len(list.ReportCards) != 3 || list.ReportCards[0].Date != "2026-06-05" {
		t.Fatalf("report cards: %d %s", w.Code, w.Body)
	}
	w = do(r, http.MethodGet, "/api/report-cards/"+list.ReportCards[0].ID, token, "")
	if w.Code != http.StatusOK || !strings.Contains(w.Header().Get("Content-Security-Policy"), "default-src 'none'") {
		t.Fatalf("report card: %d %v", w.Code, w.Header())
	}
	if w = do(r, http.MethodGet, "/api/report-cards/999", token, ""); w.Code != http.StatusNotFound {
		t.Errorf("missing report card: %d", w.Code)
	}
}

func TestAPIErrors(t *testing.T) {
	r, sealer := testServer(t)

	w := do(r, http.MethodPost, "/api/session", "", `{"username":"alex","password":"`+fake.WrongPassword+`"}`)
	if w.Code != http.StatusUnauthorized || errorCode(t, w) != CodeInvalidCredentials {
		t.Errorf("wrong password: %d %s", w.Code, w.Body)
	}
	if w = do(r, http.MethodPost, "/api/session", "", `{}`); w.Code != http.StatusBadRequest {
		t.Errorf("empty body: %d", w.Code)
	}
	if w = do(r, http.MethodGet, "/api/student", "", ""); w.Code != http.StatusUnauthorized || errorCode(t, w) != CodeSessionExpired {
		t.Errorf("no token: %d", w.Code)
	}
	if w = do(r, http.MethodGet, "/api/student", "garbage", ""); errorCode(t, w) != CodeSessionExpired {
		t.Errorf("bad token: %s", w.Body)
	}

	// A valid token whose upstream ticket expired asks the client to sign in again.
	creds := student.Credentials{}
	creds.Mobile.ServiceTicket = fake.ExpiredTicket
	expired, _ := sealer.Seal(creds)
	if w = do(r, http.MethodGet, "/api/student", expired, ""); errorCode(t, w) != CodeSessionExpired {
		t.Errorf("expired ticket: %s", w.Body)
	}
	// No portal sign-in in the token: report cards say sign in again.
	if w = do(r, http.MethodGet, "/api/report-cards", expired, ""); errorCode(t, w) != CodeSessionExpired {
		t.Errorf("no portal: %s", w.Body)
	}
	if w = do(r, http.MethodGet, "/api/schedule?year=x", login(t, r), ""); w.Code != http.StatusBadRequest {
		t.Errorf("bad schedule query: %d", w.Code)
	}
}

func TestGzip(t *testing.T) {
	r, _ := testServer(t)
	token := login(t, r)
	req := httptest.NewRequest(http.MethodGet, "/api/student", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("want gzip 200, got %d %q", w.Code, w.Header().Get("Content-Encoding"))
	}
	zr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	var snap student.Snapshot
	if err := json.NewDecoder(zr).Decode(&snap); err != nil || len(snap.Courses) == 0 {
		t.Fatalf("decoded snapshot: %v, %d courses", err, len(snap.Courses))
	}
}

// Classmates behind one school IP must not lock each other out.
func TestWebApp(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "index.html"), []byte("app"), 0o644)
	_ = os.Mkdir(filepath.Join(dir, "assets"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "assets", "a.js"), []byte("js"), 0o644)
	_ = os.WriteFile(filepath.Join(t.TempDir(), "secret"), []byte("no"), 0o644)
	_, sealer := testServer(t)
	r := New(student.NewService(fake.Mobile{}, fake.Portal{}, time.UTC), sealer, Options{WebDir: dir})

	for path, want := range map[string]string{"/": "app", "/schedule": "app", "/assets/a.js": "js"} {
		if w := do(r, http.MethodGet, path, "", ""); w.Code != http.StatusOK || w.Body.String() != want {
			t.Errorf("GET %s = %d %q, want %q", path, w.Code, w.Body, want)
		}
	}
	if w := do(r, http.MethodGet, "/../secret", "", ""); w.Body.String() == "no" {
		t.Error("served a file outside the web dir")
	}
	if w := do(r, http.MethodGet, "/api/nope", "", ""); w.Code != http.StatusNotFound || errorCode(t, w) != CodeNotFound {
		t.Errorf("unknown API route = %d, want JSON 404", w.Code)
	}
	if got := do(r, http.MethodGet, "/assets/a.js", "", "").Header().Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Errorf("assets Cache-Control = %q", got)
	}
}
