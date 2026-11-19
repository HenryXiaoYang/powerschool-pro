package student

import (
	"html"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/mobile"
)

var (
	breakTag = regexp.MustCompile(`(?i)<\s*(br|/p|/div|/li)\s*/?>`)
	anyTag   = regexp.MustCompile(`(?s)<[^>]*>`)
	blanks   = regexp.MustCompile(`\n{3,}`)
	// Titles a school uses for parent-teacher conference booking, in English and Chinese.
	conference = regexp.MustCompile(`(?i)conference|\bptcs?\b|parent[\s-]*teacher|家长会|家校`)
)

// plainText reduces school-authored markup to readable text. Clients render it as text,
// never as HTML, so this is for readability rather than safety.
func plainText(s string) string {
	s = breakTag.ReplaceAllString(s, "\n")
	s = anyTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\r", "")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSpace(l)
	}
	return strings.TrimSpace(blanks.ReplaceAllString(strings.Join(lines, "\n"), "\n\n"))
}

// bulletinDate keeps the YYYY-MM-DD the API sends ("2026-11-02", "2026-11-02Z", full timestamps).
func bulletinDate(s string) string {
	if len(s) >= 10 {
		if _, err := time.Parse(time.DateOnly, s[:10]); err == nil {
			return s[:10]
		}
	}
	return ""
}

// announcements merges the bulletins every session returned (they are school-wide, so a
// parent and a student session carry the same ones), drops expired and empty ones, and
// lists the newest first.
func announcements(sets [][]mobile.StudentData, today string) []Announcement {
	out := []Announcement{}
	seen := map[int64]bool{}
	for _, set := range sets {
		for _, d := range set {
			for _, b := range d.Bulletins {
				if seen[b.ID] {
					continue
				}
				a := Announcement{ID: b.ID, Title: plainText(b.Name.Value), Body: plainText(b.Body.Value),
					Start: bulletinDate(b.StartDate.Value), End: bulletinDate(b.EndDate.Value)}
				if a.Title == "" && a.Body == "" {
					continue
				}
				if a.End != "" && a.End < today || a.Start != "" && a.Start > today {
					continue
				}
				seen[b.ID] = true
				out = append(out, a)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Start > out[j].Start })
	return out
}

// links merges the schools' custom links. Only http(s) URLs survive; a path starting with
// "/" is resolved against the portal origin. Conference-looking links sort first.
func links(sets [][]mobile.StudentData, portal string) []Link {
	out := []Link{}
	seen := map[string]bool{}
	for _, set := range sets {
		for _, d := range set {
			for _, p := range d.CustomPage {
				raw := strings.TrimSpace(p.RedirectURL.Value)
				if strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") && portal != "" {
					raw = portal + raw
				}
				u, err := url.Parse(raw)
				if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || seen[raw] {
					continue
				}
				seen[raw] = true
				title := plainText(p.FormTitle.Value)
				if title == "" {
					title = u.Host
				}
				out = append(out, Link{Title: title, URL: raw, Conference: conference.MatchString(title + " " + raw)})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Conference && !out[j].Conference })
	return out
}

// children lists every student the sign-in can open: the primary session's, then any the
// attached parent login adds. The same student reached both ways appears once.
func children(sets ...[]mobile.StudentData) []Child {
	out := []Child{}
	seen := map[int64]bool{}
	for _, set := range sets {
		for _, d := range set {
			id := recordID(d)
			if seen[id] {
				continue
			}
			seen[id] = true
			c := Child{ID: id, Name: "Student"}
			if st := d.Student; st != nil {
				if st.FirstName.Value != "" {
					c.Name = st.FirstName.Value
				}
				c.Grade = st.GradeLevel
			}
			out = append(out, c)
		}
	}
	return out
}

func recordID(d mobile.StudentData) int64 {
	if d.StudentID != 0 {
		return d.StudentID
	}
	if d.Student != nil {
		return d.Student.ID
	}
	return 0
}

// pick returns the record for child, searching the sets in order; child 0 or an unknown id
// gives the first record, so a removed parent link never strands the app.
func pick(child int64, sets ...[]mobile.StudentData) (mobile.StudentData, bool) {
	var first *mobile.StudentData
	for _, set := range sets {
		for i := range set {
			if first == nil {
				first = &set[i]
			}
			if child != 0 && recordID(set[i]) == child {
				return set[i], true
			}
		}
	}
	if first == nil {
		return mobile.StudentData{}, false
	}
	return *first, true
}
