package student

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/fake"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/mobile"
)

var shanghai, _ = time.LoadLocation("Asia/Shanghai")

func TestParseExpression(t *testing.T) {
	cases := map[string][]Slot{
		"1(MA,FB) 2(FA)":     {{"1", []string{"MA", "FB"}}, {"2", []string{"FA"}}},
		"4(TA,FA-MB,WB)":     {{"4", []string{"TA", "FA", "MB", "WB"}}},
		"LT2(MA-THA,MB-THB)": {{"LT2", []string{"MA", "TA", "WA", "THA", "MB", "TB", "WB", "THB"}}},
		"Tu2-LT2(FA,FB)":     {{"Tu2-LT2", []string{"FA", "FB"}}},
		"5-6(TA)":            {{"5-6", []string{"TA"}}},
		"BF(MA-FB)":          {{"BF", []string{"MA", "TA", "WA", "THA", "FA", "MB", "TB", "WB", "THB", "FB"}}},
		"":                   {},
		"7":                  {{"7", []string{}}},
	}
	for expr, want := range cases {
		if got := parseExpression(expr); !reflect.DeepEqual(got, want) {
			t.Errorf("parseExpression(%q) = %v, want %v", expr, got, want)
		}
	}
}

func TestInferCycleDay(t *testing.T) {
	// Monday: one section meets only on MA, another on MA and MB → A.
	if day, week := inferCycleDay(0, [][]string{{"MA", "MB"}, {"MA", "TB"}}); day != "MA" || week != "A" {
		t.Errorf("got %q %q", day, week)
	}
	// Thursday letters conflict → unknown rather than a guess.
	if day, _ := inferCycleDay(3, [][]string{{"THA"}, {"THB"}}); day != "" {
		t.Errorf("conflict should be unknown, got %q", day)
	}
}

func TestBuildFromFixture(t *testing.T) {
	now := time.Date(2026, 11, 4, 10, 0, 0, 0, shanghai) // Wednesday of the fixture's cycle
	data, err := fake.Mobile{Now: func() time.Time { return now }}.StudentData(context.Background(), mobile.Session{})
	if err != nil {
		t.Fatal(err)
	}
	snap := Build(data[0], shanghai, now)

	byID := map[int64]Course{}
	for _, c := range snap.Courses {
		byID[c.ID] = c
	}
	for _, c := range snap.Courses {
		if c.Title == "Breakfast 早餐" && c.Academic {
			t.Error("boarding block marked academic")
		}
	}
	scored, ungraded := 0, 0
	for _, a := range snap.Assignments {
		if a.CourseTitle == "" || byID[a.CourseID].Title != a.CourseTitle {
			t.Fatalf("assignment %d not joined to its course", a.ID)
		}
		if a.Score == nil {
			ungraded++
		} else {
			scored++
		}
	}
	if scored == 0 || ungraded == 0 {
		t.Errorf("want both scored and ungraded assignments, got %d/%d", scored, ungraded)
	}
	for _, g := range snap.FinalGrades {
		if g.Percent != nil {
			t.Error("fixture terms suppress percents; percent must be hidden")
		}
	}
	for _, a := range snap.Attendance {
		if a.CourseID == 0 || a.Date == "" {
			t.Errorf("attendance %d not joined: %+v", a.ID, a)
		}
	}
	current := 0
	for _, rt := range snap.ReportingTerms {
		if rt.Current {
			current++
		}
	}
	if current != 1 {
		t.Errorf("want exactly one current reporting term, got %d", current)
	}

	if len(snap.Schedule.Meetings) == 0 {
		t.Fatal("no meetings this month")
	}
	labelled := 0
	for _, d := range snap.Schedule.Days {
		if d.Day != "" {
			labelled++
		}
	}
	if labelled < len(snap.Schedule.Days)/2 {
		t.Errorf("only %d of %d days got a cycle label", labelled, len(snap.Schedule.Days))
	}
}

func TestLocalDates(t *testing.T) {
	due := time.Date(2026, 11, 6, 16, 0, 0, 0, time.UTC) // local midnight, Nov 7 in Shanghai
	d := mobile.StudentData{Assignments: mobile.List[mobile.Assignment]{{ID: 1, DueDate: &due}}}
	snap := Build(d, shanghai, due)
	if got := snap.Assignments[0].DueDate; got != "2026-11-07" {
		t.Errorf("due date = %s, want 2026-11-07", got)
	}
}
