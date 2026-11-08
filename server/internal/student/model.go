// Package student turns raw PowerSchool records into the joined, typed models the API
// serves. It owns the school rules: joins, local dates, the A/B day cycle, and grade
// suppression. Nothing here invents a grade.
package student

import "time"

// Snapshot is everything the app shows, from one getStudentData call.
type Snapshot struct {
	Profile         Profile          `json:"profile"`
	Terms           []Term           `json:"terms"`
	ReportingTerms  []ReportingTerm  `json:"reportingTerms"`
	Courses         []Course         `json:"courses"`
	Assignments     []Assignment     `json:"assignments"`
	FinalGrades     []FinalGrade     `json:"finalGrades"`
	Attendance      []Attendance     `json:"attendance"`
	AttendanceCodes []AttendanceCode `json:"attendanceCodes"`
	Periods         []Period         `json:"periods"`
	Holidays        []Holiday        `json:"holidays"`
	// Schedule covers the current month; other months come from the schedule endpoint.
	Schedule  Schedule  `json:"schedule"`
	FetchedAt time.Time `json:"fetchedAt"`
	TimeZone  string    `json:"timeZone"`
}

type Profile struct {
	FirstName   string `json:"firstName"`
	MiddleName  string `json:"middleName"`
	LastName    string `json:"lastName"`
	GradeLevel  int    `json:"gradeLevel"`
	GPA         string `json:"gpa"`
	CurrentTerm string `json:"currentTerm"` // reporting-term abbreviation, e.g. "P2"
	School      string `json:"school"`
}

// Term is a scheduling term (year, semester, trimester). Distinct from ReportingTerm.
type Term struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Abbrev   string `json:"abbrev"`
	ParentID int64  `json:"parentId"`
	Start    string `json:"start"` // local date YYYY-MM-DD
	End      string `json:"end"`
}

// ReportingTerm is a grading period that course grades are stored against.
type ReportingTerm struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	Abbrev           string `json:"abbrev"`
	Start            string `json:"start"`
	End              string `json:"end"`
	Current          bool   `json:"current"`
	SuppressGrades   bool   `json:"suppressGrades"`
	SuppressPercents bool   `json:"suppressPercents"`
}

type Course struct {
	ID         int64    `json:"id"`
	Code       string   `json:"code"`
	Title      string   `json:"title"`
	Expression string   `json:"expression"` // as PowerSchool prints it, e.g. "1(MA,FB) 2(FA)"
	Slots      []Slot   `json:"slots"`
	Room       string   `json:"room"`
	Teacher    *Teacher `json:"teacher"`
	TermID     int64    `json:"termId"`
	PeriodSort int      `json:"periodSort"`
	// Academic is false for boarding, lunch, assembly and similar blocks.
	Academic bool `json:"academic"`
	Absences int  `json:"absences"`
	Tardies  int  `json:"tardies"`
}

// Slot is one "periods(days)" group of a section expression.
type Slot struct {
	Periods string   `json:"periods"` // "1", "5-6", "Tu2-LT2" as written
	Days    []string `json:"days"`    // expanded cycle days, e.g. ["MA","TA","WA"]
}

type Teacher struct {
	Name  string `json:"name"` // "Last, First" as the portal prints it
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type Assignment struct {
	ID          int64    `json:"id"`
	CourseID    int64    `json:"courseId"`
	CourseTitle string   `json:"courseTitle"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	DueAt       *string  `json:"dueAt"`   // RFC 3339 instant
	DueDate     string   `json:"dueDate"` // local date YYYY-MM-DD, "" when undated
	Points      *float64 `json:"points"`
	Weight      *float64 `json:"weight"`
	// Counted mirrors includeinfinalgrades; false shows as "not counted".
	Counted bool `json:"counted"`
	// Score is nil when PowerSchool has no score record (ungraded), never zero-filled.
	Score *Score `json:"score"`
}

type Score struct {
	Value      string   `json:"value"` // server display value: "86", "0", "--"
	Percent    *float64 `json:"percent"`
	Grade      string   `json:"grade"` // letterGrade, e.g. IB "6"
	Comment    string   `json:"comment"`
	Late       bool     `json:"late"`
	Missing    bool     `json:"missing"`
	Exempt     bool     `json:"exempt"`
	Collected  bool     `json:"collected"`
	Incomplete bool     `json:"incomplete"`
}

// FinalGrade is a stored course grade for one reporting term.
type FinalGrade struct {
	ID              int64    `json:"id"`
	CourseID        int64    `json:"courseId"`
	ReportingTermID int64    `json:"reportingTermId"`
	Grade           string   `json:"grade"`
	Percent         *float64 `json:"percent"` // nil when the term suppresses percents
	Comment         string   `json:"comment"`
}

type AttendanceKind string

const (
	KindAbsent  AttendanceKind = "absent"
	KindTardy   AttendanceKind = "tardy"
	KindPresent AttendanceKind = "present"
	KindOther   AttendanceKind = "other"
)

type Attendance struct {
	ID          int64          `json:"id"`
	Date        string         `json:"date"` // local date
	Code        string         `json:"code"`
	Description string         `json:"description"`
	Kind        AttendanceKind `json:"kind"`
	CourseID    int64          `json:"courseId"` // 0 when the record is not tied to a section
	Period      string         `json:"period"`
	Minutes     *float64       `json:"minutes"`
	Comment     string         `json:"comment"`
}

type AttendanceCode struct {
	Code        string         `json:"code"`
	Description string         `json:"description"`
	Kind        AttendanceKind `json:"kind"`
}

type Period struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Abbrev string `json:"abbrev"`
	Sort   int    `json:"sort"`
}

type Holiday struct {
	Date        string `json:"date"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// Schedule is one month of real meeting times.
type Schedule struct {
	Year     int        `json:"year"`
	Month    int        `json:"month"`
	Meetings []Meeting  `json:"meetings"`
	Days     []CycleDay `json:"days"`
}

type Meeting struct {
	CourseID int64  `json:"courseId"`
	Start    string `json:"start"` // RFC 3339 instant
	End      string `json:"end"`
	Date     string `json:"date"` // local date of Start
}

// CycleDay names a school day in the A/B rotation, inferred from which sections meet.
type CycleDay struct {
	Date string `json:"date"`
	Day  string `json:"day"`  // e.g. "MA"; "" when the meetings do not pin it down
	Week string `json:"week"` // e.g. "A"
}

// ReportCard is one archived report card.
type ReportCard struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Date string `json:"date"` // local date YYYY-MM-DD, "" if unparseable
}
