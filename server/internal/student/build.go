package student

import (
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/mobile"
)

// Build joins one student's raw record into a Snapshot. Joins follow PROTOCOL.md §9:
// score.assignmentId → assignment.id, assignment.sectionid → section.id,
// attendance.ccid → section enrollment id, finalGrade.reportingTermId → reportingTerm.id.
func Build(d mobile.StudentData, loc *time.Location, now time.Time) Snapshot {
	localDate := func(t *time.Time) string {
		if t == nil {
			return ""
		}
		return t.In(loc).Format(time.DateOnly)
	}

	snap := Snapshot{
		FetchedAt: now.UTC(), TimeZone: loc.String(),
		Terms: []Term{}, ReportingTerms: []ReportingTerm{}, Courses: []Course{}, Assignments: []Assignment{},
		FinalGrades: []FinalGrade{}, Attendance: []Attendance{}, AttendanceCodes: []AttendanceCode{},
		Periods: []Period{}, Holidays: []Holiday{},
	}
	if st := d.Student; st != nil {
		snap.Profile = Profile{
			FirstName: st.FirstName.Value, MiddleName: st.MiddleName.Value, LastName: st.LastName.Value,
			GradeLevel: st.GradeLevel, GPA: st.CurrentGPA.Value, CurrentTerm: st.CurrentTerm.Value,
		}
	}
	if len(d.Schools) > 0 {
		snap.Profile.School = d.Schools[0].Name.Value
	}

	for _, t := range d.Terms {
		snap.Terms = append(snap.Terms, Term{ID: t.ID, Title: t.Title.Value, Abbrev: t.Abbrev.Value,
			ParentID: t.ParentTermID, Start: localDate(t.StartDate), End: localDate(t.EndDate)})
	}
	sort.SliceStable(snap.Terms, func(i, j int) bool { return snap.Terms[i].Start < snap.Terms[j].Start })

	rterms := map[int64]mobile.ReportingTerm{}
	for _, rt := range d.ReportingTerms {
		rterms[rt.ID] = rt
		snap.ReportingTerms = append(snap.ReportingTerms, ReportingTerm{
			ID: rt.ID, Title: rt.Title.Value, Abbrev: rt.Abbreviation.Value,
			Start: localDate(rt.StartDate), End: localDate(rt.EndDate),
			Current:        rt.Abbreviation.Valid && rt.Abbreviation.Value == snap.Profile.CurrentTerm,
			SuppressGrades: rt.SuppressGrades, SuppressPercents: rt.SuppressPercents,
		})
	}
	sort.SliceStable(snap.ReportingTerms, func(i, j int) bool {
		a, b := snap.ReportingTerms[i], snap.ReportingTerms[j]
		return a.Start < b.Start || (a.Start == b.Start && a.End > b.End)
	})

	teachers := map[int64]*Teacher{}
	for _, t := range d.Teachers {
		name := t.LastName.Value
		if t.FirstName.Value != "" {
			name += ", " + t.FirstName.Value
		}
		teachers[t.ID] = &Teacher{Name: name, Email: t.Email.Value, Phone: t.SchoolPhone.Value}
	}

	periods := map[int64]string{}
	for _, p := range d.Periods {
		periods[p.ID] = p.Abbreviation.Value
		snap.Periods = append(snap.Periods, Period{ID: p.ID, Name: p.Name.Value, Abbrev: p.Abbreviation.Value, Sort: p.SortOrder})
	}
	sort.SliceStable(snap.Periods, func(i, j int) bool { return snap.Periods[i].Sort < snap.Periods[j].Sort })

	categories := map[int64]string{}
	for _, c := range d.AssignmentCategories {
		categories[c.ID] = c.Name.Value
	}

	codes := map[int64]AttendanceCode{}
	for _, c := range d.AttendanceCodes {
		code := AttendanceCode{Code: c.AttCode.Value, Description: c.Description.Value, Kind: attendanceKind(c)}
		codes[c.ID] = code
		snap.AttendanceCodes = append(snap.AttendanceCodes, code)
	}

	// Courses, and the enrollment → course map attendance needs.
	courseIdx := map[int64]int{}
	enrollmentCourse := map[int64]int64{}
	hasAssignments := map[int64]bool{}
	for _, a := range d.Assignments {
		hasAssignments[a.SectionID] = true
	}
	for _, s := range d.Sections {
		c := Course{
			ID: s.ID, Code: s.CourseCode.Value, Title: s.SchoolCourseTitle.Value,
			Expression: s.Expression.Value, Slots: parseExpression(s.Expression.Value),
			Room: s.RoomName.Value, Teacher: teachers[s.TeacherID], TermID: s.TermID, PeriodSort: s.PeriodSort,
		}
		c.Academic = hasAssignments[s.ID] || (c.Room != "" && hasNumericPeriod(c.Slots))
		courseIdx[s.ID] = len(snap.Courses)
		snap.Courses = append(snap.Courses, c)
		for _, e := range s.Enrollments {
			enrollmentCourse[e.ID] = s.ID
		}
	}
	sort.SliceStable(snap.Courses, func(i, j int) bool { return snap.Courses[i].PeriodSort < snap.Courses[j].PeriodSort })
	for i, c := range snap.Courses {
		courseIdx[c.ID] = i
	}

	scores := map[int64]mobile.AssignmentScore{}
	for _, s := range d.AssignmentScores {
		scores[s.AssignmentID] = s
	}
	for _, a := range d.Assignments {
		out := Assignment{
			ID: a.ID, CourseID: a.SectionID, Name: a.Name.Value, Description: a.Description.Value,
			Category: categories[a.CategoryID], DueDate: localDate(a.DueDate),
			Points: optional(a.PointsPossible), Weight: optional(a.Weight), Counted: a.IncludeInFinalGrades == 1,
		}
		if i, ok := courseIdx[a.SectionID]; ok {
			out.CourseTitle = snap.Courses[i].Title
		}
		if a.DueDate != nil {
			s := a.DueDate.UTC().Format(time.RFC3339)
			out.DueAt = &s
		}
		if s, ok := scores[a.ID]; ok {
			out.Score = &Score{
				Value: s.Score.Value, Percent: optional(s.Percent), Grade: s.LetterGrade.Value, Comment: s.Comment.Value,
				Late: s.Late, Missing: s.Missing, Exempt: s.Exempt, Collected: s.Collected, Incomplete: s.Incomplete,
			}
		}
		snap.Assignments = append(snap.Assignments, out)
	}
	sort.SliceStable(snap.Assignments, func(i, j int) bool { return snap.Assignments[i].DueDate > snap.Assignments[j].DueDate })

	for _, g := range d.FinalGrades {
		rt, known := rterms[g.ReportingTermID]
		if known && rt.SuppressGrades {
			continue
		}
		fg := FinalGrade{ID: g.ID, CourseID: g.SectionID, ReportingTermID: g.ReportingTermID,
			Grade: g.Grade.Value, Percent: optional(g.Percent), Comment: g.CommentValue.Value}
		if known && rt.SuppressPercents {
			fg.Percent = nil
		}
		snap.FinalGrades = append(snap.FinalGrades, fg)
	}

	for _, a := range d.Attendance {
		code := codes[a.AttCodeID]
		rec := Attendance{
			ID: a.ID, Date: localDate(a.AttDate), Code: code.Code, Description: code.Description, Kind: code.Kind,
			CourseID: enrollmentCourse[a.CCID], Period: periods[a.PeriodID],
			Minutes: optional(a.TotalMinutes), Comment: a.AttComment.Value,
		}
		if rec.Kind == "" {
			rec.Kind = KindOther
		}
		if i, ok := courseIdx[rec.CourseID]; ok && rec.CourseID != 0 {
			switch rec.Kind {
			case KindAbsent:
				snap.Courses[i].Absences++
			case KindTardy:
				snap.Courses[i].Tardies++
			}
		}
		snap.Attendance = append(snap.Attendance, rec)
	}
	sort.SliceStable(snap.Attendance, func(i, j int) bool { return snap.Attendance[i].Date > snap.Attendance[j].Date })

	for _, h := range d.NotInSessionDays {
		snap.Holidays = append(snap.Holidays, Holiday{Date: localDate(h.CalendarDay), Type: h.CalType.Value, Description: h.Description.Value})
	}
	sort.SliceStable(snap.Holidays, func(i, j int) bool { return snap.Holidays[i].Date < snap.Holidays[j].Date })

	local := now.In(loc)
	snap.Schedule = BuildSchedule(d.Sections, loc, local.Year(), int(local.Month()))
	return snap
}

// BuildSchedule lists one month's meetings from the sections' start/stop times and
// labels each school day with its rotation day.
func BuildSchedule(sections []mobile.Section, loc *time.Location, year, month int) Schedule {
	sched := Schedule{Year: year, Month: month, Meetings: []Meeting{}, Days: []CycleDay{}}
	daysByDate := map[string][][]string{}
	weekdayOf := map[string]int{}
	seen := map[Meeting]bool{} // PowerSchool repeats some start/stop rows
	for _, s := range sections {
		var days []string
		for _, slot := range parseExpression(s.Expression.Value) {
			days = append(days, slot.Days...)
		}
		for _, ss := range s.StartStopDates {
			if ss.Start == nil || ss.Stop == nil {
				continue
			}
			start := ss.Start.In(loc)
			if start.Year() != year || int(start.Month()) != month {
				continue
			}
			date := start.Format(time.DateOnly)
			m := Meeting{CourseID: s.ID, Date: date, Start: ss.Start.UTC().Format(time.RFC3339), End: ss.Stop.UTC().Format(time.RFC3339)}
			if seen[m] {
				continue
			}
			seen[m] = true
			sched.Meetings = append(sched.Meetings, m)
			daysByDate[date] = append(daysByDate[date], days)
			weekdayOf[date] = (int(start.Weekday()) + 6) % 7
		}
	}
	sort.SliceStable(sched.Meetings, func(i, j int) bool { return sched.Meetings[i].Start < sched.Meetings[j].Start })
	for date, sets := range daysByDate {
		day, week := inferCycleDay(weekdayOf[date], sets)
		sched.Days = append(sched.Days, CycleDay{Date: date, Day: day, Week: week})
	}
	sort.Slice(sched.Days, func(i, j int) bool { return sched.Days[i].Date < sched.Days[j].Date })
	return sched
}

func attendanceKind(c mobile.AttendanceCode) AttendanceKind {
	switch {
	case c.CodeType == 2:
		return KindAbsent
	case strings.TrimSpace(c.AttCode.Value) == "":
		return KindPresent
	case c.AttCode.Value == "T" || c.AttCode.Value == "TA" || strings.Contains(strings.ToLower(c.Description.Value), "tardy"):
		return KindTardy
	}
	return KindOther
}

// hasNumericPeriod tells teaching periods (1–7) from named blocks (BF, LT2, Tu1, CA1, SH).
func hasNumericPeriod(slots []Slot) bool {
	for _, s := range slots {
		if s.Periods != "" && unicode.IsDigit(rune(s.Periods[0])) {
			return true
		}
	}
	return false
}

func optional(n mobile.Number) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Value
	return &v
}
