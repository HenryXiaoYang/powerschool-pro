// Package fake is a synthetic PowerSchool for MOCK=1 and tests. fixture.json is shaped
// like a real getStudentData response but every personal value is invented. Dates are
// shifted by whole A/B cycles so the fixture's timetable always covers "today".
package fake

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/mobile"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/portal"
)

//go:embed fixture.json
var fixture []byte

// anchor is the Monday (local) that starts the fixture's one recorded A/B cycle.
var anchor = time.Date(2026, 11, 2, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))

const cycle = 14 * 24 * time.Hour

// Password "wrong" fails sign-in; ticket "expired" fails data calls. Anything else works.
const (
	WrongPassword = "wrong"
	ExpiredTicket = "expired"
)

type Mobile struct{ Now func() time.Time }

func (m Mobile) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now()
}

func (Mobile) Login(_ context.Context, username, password string) (mobile.Session, error) {
	if username == "" || password == "" || password == WrongPassword {
		return mobile.Session{}, powerschool.ErrInvalidCredentials
	}
	return mobile.Session{
		ServiceTicket: "fake-ticket", StudentIDs: mobile.List[int64]{1001},
		UserID: mobile.Text{Value: "1", Valid: true}, UserType: mobile.Text{Value: "2", Valid: true},
		ServerInfo: mobile.ServerInfo{APIVersion: mobile.Text{Value: "3.0.0", Valid: true}},
	}, nil
}

func (m Mobile) StudentData(_ context.Context, s mobile.Session) ([]mobile.StudentData, error) {
	if s.ServiceTicket == ExpiredTicket {
		return nil, powerschool.ErrSessionExpired
	}
	d, err := m.load()
	if err != nil {
		return nil, err
	}
	return []mobile.StudentData{d}, nil
}

// SectionTimes repeats the fixture's single cycle across the requested month.
func (m Mobile) SectionTimes(_ context.Context, s mobile.Session, year, month int) ([]mobile.StudentData, error) {
	if s.ServiceTicket == ExpiredTicket {
		return nil, powerschool.ErrSessionExpired
	}
	d, err := m.load()
	if err != nil {
		return nil, err
	}
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, anchor.Location())
	for i := range d.Sections {
		var out mobile.List[mobile.StartStop]
		for _, ss := range d.Sections[i].StartStopDates {
			for k := -4; k <= 4; k++ {
				start, stop := ss.Start.Add(time.Duration(k)*cycle), ss.Stop.Add(time.Duration(k)*cycle)
				if local := start.In(anchor.Location()); local.Year() == first.Year() && local.Month() == first.Month() {
					out = append(out, mobile.StartStop{SectionEnrollmentID: ss.SectionEnrollmentID, Start: &start, Stop: &stop})
				}
			}
		}
		d.Sections[i].StartStopDates = out
	}
	return []mobile.StudentData{d}, nil
}

var isoTime = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T`)

// load decodes the fixture with every timestamp moved forward by whole cycles.
func (m Mobile) load() (mobile.StudentData, error) {
	var tree any
	if err := json.Unmarshal(fixture, &tree); err != nil {
		return mobile.StudentData{}, fmt.Errorf("fake: fixture: %w", err)
	}
	shift := time.Duration(int(m.now().Sub(anchor)/cycle)) * cycle
	if m.now().Before(anchor) {
		shift -= cycle
	}
	b, _ := json.Marshal(shiftTimes(tree, shift))
	var d mobile.StudentData
	if err := json.Unmarshal(b, &d); err != nil {
		return mobile.StudentData{}, fmt.Errorf("fake: fixture: %w", err)
	}
	return d, nil
}

func shiftTimes(v any, by time.Duration) any {
	switch t := v.(type) {
	case map[string]any:
		for k, x := range t {
			t[k] = shiftTimes(x, by)
		}
	case []any:
		for i, x := range t {
			t[i] = shiftTimes(x, by)
		}
	case string:
		if isoTime.MatchString(t) {
			if ts, err := time.Parse(time.RFC3339, t); err == nil {
				return ts.Add(by).UTC().Format("2006-01-02T15:04:05.000Z")
			}
		}
	}
	return v
}

type Portal struct{}

func (Portal) Login(_ context.Context, username, password string) (portal.Session, error) {
	if username == "" || password == "" || password == WrongPassword {
		return portal.Session{}, powerschool.ErrInvalidCredentials
	}
	return portal.Session{Cookies: []portal.Cookie{{Name: "JSESSIONID", Value: "fake"}}}, nil
}

var reportCards = []portal.ReportCardRef{
	{ID: "3003", Name: "Y10 Y1 Report Card", Date: "06/05/2026"},
	{ID: "3002", Name: "Y10 S1 Report Card", Date: "01/30/2026"},
	{ID: "3001", Name: "Y10 M1 Report Card", Date: "11/14/2025"},
}

func (Portal) ReportCards(context.Context, portal.Session) ([]portal.ReportCardRef, error) {
	return reportCards, nil
}

func (Portal) ReportCard(_ context.Context, _ portal.Session, id string) (string, error) {
	for _, rc := range reportCards {
		if rc.ID == id {
			return fmt.Sprintf(reportCardHTML, rc.Name, rc.Date), nil
		}
	}
	return "", powerschool.ErrNotFound
}

const reportCardHTML = `<!doctype html><html><head><meta charset="utf-8"><style>
body{font:10pt sans-serif;margin:0;background:#fff;color:#111}.page{max-width:720px;margin:24px auto;padding:32px;border:1px solid #ccc}
h1{font-size:14pt;text-align:center}table{width:100%%;border-collapse:collapse;margin-top:16px}td,th{border:1px solid #888;padding:3px 6px}th{background:#ddd}
</style></head><body><div class="page"><h1>%s</h1><p>Synthetic High School · Generated on %s</p>
<p><b>Alex Rivera</b> · Grade 10 · synthetic demo data</p>
<table><tr><th>Course</th><th>Teacher</th><th>Grade</th><th>Absences</th></tr>
<tr><td>10 Biology</td><td>Morgan, Avery</td><td>A</td><td>1</td></tr>
<tr><td>10 Mathematics</td><td>Lin, Jia</td><td>A*</td><td>0</td></tr>
<tr><td>10 English</td><td>Okafor, Tunde</td><td>B</td><td>0</td></tr></table></div></body></html>`
