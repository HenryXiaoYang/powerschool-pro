//go:build live

// Live checks against the real school, using the repo-root .env
// (SCHOOL_BASE_URL, USERNAME, PASSWORD). Run: make live-test. Prints counts only.
package powerschool_test

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/mobile"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/portal"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/session"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/student"
)

func liveEnv(t *testing.T) map[string]string {
	f, err := os.Open("../../../.env")
	if err != nil {
		t.Skip("no .env at repo root")
	}
	defer f.Close()
	env := map[string]string{}
	for sc := bufio.NewScanner(f); sc.Scan(); {
		k, v, ok := strings.Cut(strings.TrimSpace(sc.Text()), "=")
		if ok {
			env[k] = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	return env
}

func TestLiveMobile(t *testing.T) {
	env := liveEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	c, err := mobile.New(env["SCHOOL_BASE_URL"], nil, mobile.DefaultServiceUser, mobile.DefaultServicePassword)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Login(ctx, env["USERNAME"], env["PASSWORD"])
	if err != nil {
		t.Fatal(err)
	}
	data, err := c.StudentData(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	d := data[0]
	loc, _ := time.LoadLocation("Asia/Shanghai")
	snap := student.Build(d, loc, time.Now())
	academic, labelled := 0, 0
	for _, c := range snap.Courses {
		if c.Academic {
			academic++
		}
	}
	for _, day := range snap.Schedule.Days {
		if day.Day != "" {
			labelled++
		}
	}
	t.Logf("snapshot: courses=%d academic=%d meetings=%d days=%d labelled=%d", len(snap.Courses), academic, len(snap.Schedule.Meetings), len(snap.Schedule.Days), labelled)
	t.Logf("sections=%d assignments=%d scores=%d finalGrades=%d attendance=%d",
		len(d.Sections), len(d.Assignments), len(d.AssignmentScores), len(d.FinalGrades), len(d.Attendance))

	next := time.Now().AddDate(0, 1, 0)
	times, err := c.SectionTimes(ctx, s, next.Year(), int(next.Month()))
	if err != nil {
		t.Fatalf("SectionTimes: %v", err)
	}
	n := 0
	for _, sec := range times[0].Sections {
		n += len(sec.StartStopDates)
	}
	t.Logf("SectionTimes %d-%02d: sections=%d meetings=%d", next.Year(), next.Month(), len(times[0].Sections), n)
}

func TestLivePortal(t *testing.T) {
	env := liveEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	c, err := portal.New(env["SCHOOL_BASE_URL"], nil)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Login(ctx, env["USERNAME"], env["PASSWORD"])
	if err != nil {
		t.Fatal(err)
	}
	refs, err := c.ReportCards(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("report cards=%d", len(refs))
	if len(refs) == 0 {
		return
	}
	page, err := c.ReportCard(ctx, s, refs[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("first report card: %d bytes, inline images=%d", len(page), strings.Count(page, "data:image/"))
	if out := os.Getenv("LIVE_OUT"); out != "" {
		_ = os.WriteFile(out, []byte(page), 0o600)
	}
}

// TestLiveService follows the API's real path: mobile and portal sign-in together, then the snapshot.
func TestLiveService(t *testing.T) {
	env := liveEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	m, err := mobile.New(env["SCHOOL_BASE_URL"], nil, mobile.DefaultServiceUser, mobile.DefaultServicePassword)
	if err != nil {
		t.Fatal(err)
	}
	p, err := portal.New(env["SCHOOL_BASE_URL"], nil)
	if err != nil {
		t.Fatal(err)
	}
	svc := student.NewService(m, p, time.UTC)
	creds, err := svc.Login(ctx, env["USERNAME"], env["PASSWORD"])
	if err != nil {
		t.Fatal(err)
	}
	// The API keeps credentials sealed in the bearer token: round-trip them the same way.
	sealer, _ := session.NewSealer(make([]byte, 32), time.Hour)
	token, err := sealer.Seal(creds)
	if err != nil || sealer.Open(token, &creds) != nil {
		t.Fatal("seal round-trip failed", err)
	}
	snap, err := svc.Snapshot(ctx, creds)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("service snapshot: courses=%d portal=%v", len(snap.Courses), creds.Portal != nil)
}
