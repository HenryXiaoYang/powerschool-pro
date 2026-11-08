package student

import (
	"regexp"
	"strings"
)

// A cycle day code is a weekday prefix plus a rotation letter: MA, THA, FB.
var weekdayIndex = map[string]int{"M": 0, "T": 1, "TU": 1, "W": 2, "TH": 3, "H": 3, "F": 4, "SA": 5, "SU": 6}

type cycleDay struct {
	weekday int  // 0 = Monday
	week    byte // rotation letter
}

func parseDay(code string) (cycleDay, bool) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) < 2 {
		return cycleDay{}, false
	}
	wd, ok := weekdayIndex[code[:len(code)-1]]
	week := code[len(code)-1]
	return cycleDay{wd, week}, ok && week >= 'A' && week <= 'Z'
}

var slotRe = regexp.MustCompile(`^([^()]+)\(([^()]*)\)$`)

// parseExpression splits a section expression such as "1(MA,FB) 2(FA) 5-6(TA)" into
// slots with day ranges expanded (FA-MB is FA then MB: Monday–Friday per rotation letter).
func parseExpression(expr string) []Slot {
	slots := []Slot{}
	for _, tok := range strings.Fields(expr) {
		m := slotRe.FindStringSubmatch(tok)
		if m == nil {
			slots = append(slots, Slot{Periods: tok, Days: []string{}})
			continue
		}
		slot := Slot{Periods: m[1], Days: []string{}}
		for _, part := range strings.Split(m[2], ",") {
			slot.Days = append(slot.Days, expandDays(part)...)
		}
		slots = append(slots, slot)
	}
	return slots
}

func expandDays(part string) []string {
	from, to, isRange := strings.Cut(strings.TrimSpace(part), "-")
	if !isRange {
		if from == "" {
			return nil
		}
		return []string{strings.ToUpper(from)}
	}
	a, okA := parseDay(from)
	b, okB := parseDay(to)
	if !okA || !okB || a.week > b.week || (a.week == b.week && a.weekday > b.weekday) {
		return []string{strings.ToUpper(part)} // unknown shape: keep as written
	}
	var out []string
	for wk := a.week; wk <= b.week; wk++ {
		first, last := 0, 4
		if wk == a.week {
			first = a.weekday
		}
		if wk == b.week {
			last = b.weekday
		}
		for wd := first; wd <= last; wd++ {
			out = append(out, dayCode(wd, wk))
		}
	}
	return out
}

var weekdayPrefix = [...]string{"M", "T", "W", "TH", "F", "SA", "SU"}

func dayCode(weekday int, week byte) string { return weekdayPrefix[weekday] + string(week) }

// inferCycleDay picks the rotation day for a date from the day sets of the sections that
// meet on it: each meeting section allows only the letters it has on that weekday.
func inferCycleDay(weekday int, meetingDays [][]string) (code string, week string) {
	var candidates map[byte]bool
	for _, days := range meetingDays {
		allowed := map[byte]bool{}
		for _, d := range days {
			if cd, ok := parseDay(d); ok && cd.weekday == weekday {
				allowed[cd.week] = true
			}
		}
		if len(allowed) == 0 {
			continue // section without usable days says nothing
		}
		if candidates == nil {
			candidates = allowed
			continue
		}
		for k := range candidates {
			if !allowed[k] {
				delete(candidates, k)
			}
		}
	}
	if len(candidates) != 1 {
		return "", ""
	}
	for k := range candidates {
		return dayCode(weekday, k), string(k)
	}
	return "", ""
}
