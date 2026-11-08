package mobile

import "time"

// Session is what loginToPublicPortal returns and getStudentData expects back.
// Values keep the server's lexical form because they are echoed verbatim.
type Session struct {
	ServerCurrentTime Text        `json:"serverCurrentTime"`
	ServerInfo        ServerInfo  `json:"serverInfo"`
	ServiceTicket     string      `json:"serviceTicket"`
	StudentIDs        List[int64] `json:"studentIDs"`
	UserID            Text        `json:"userId"`
	UserType          Text        `json:"userType"`
}

type ServerInfo struct {
	APIVersion Text `json:"apiVersion"`
}

type message struct {
	Description Text `json:"description"`
	MsgCode     Text `json:"msgCode"`
}

type results struct {
	MessageVOs     List[message]     `json:"messageVOs"`
	StudentDataVOs List[StudentData] `json:"studentDataVOs"`
	UserSessionVO  *Session          `json:"userSessionVO"`
}

// StudentData is one student's record from getStudentData. Only fields the app uses are declared.
type StudentData struct {
	StudentID            int64                    `json:"studentId"`
	YearID               int                      `json:"yearId"`
	Student              *Student                 `json:"student"`
	Schools              List[School]             `json:"schools"`
	Sections             List[Section]            `json:"sections"`
	Teachers             List[Teacher]            `json:"teachers"`
	Terms                List[Term]               `json:"terms"`
	ReportingTerms       List[ReportingTerm]      `json:"reportingTerms"`
	Periods              List[Period]             `json:"periods"`
	AssignmentCategories List[AssignmentCategory] `json:"assignmentCategories"`
	Assignments          List[Assignment]         `json:"assignments"`
	AssignmentScores     List[AssignmentScore]    `json:"assignmentScores"`
	FinalGrades          List[FinalGrade]         `json:"finalGrades"`
	Attendance           List[Attendance]         `json:"attendance"`
	AttendanceCodes      List[AttendanceCode]     `json:"attendanceCodes"`
	NotInSessionDays     List[NotInSessionDay]    `json:"notInSessionDays"`
}

type Student struct {
	ID          int64      `json:"id"`
	FirstName   Text       `json:"firstName"`
	MiddleName  Text       `json:"middleName"`
	LastName    Text       `json:"lastName"`
	GradeLevel  int        `json:"gradeLevel"`
	CurrentGPA  Text       `json:"currentGPA"`
	CurrentTerm Text       `json:"currentTerm"`
	DOB         *time.Time `json:"dob"`
}

type School struct {
	SchoolNumber  int64 `json:"schoolNumber"`
	Name          Text  `json:"name"`
	Abbreviation  Text  `json:"abbreviation"`
	CurrentTermID int64 `json:"currentTermId"`
}

type Section struct {
	ID                int64            `json:"id"`
	CourseCode        Text             `json:"courseCode"`
	SchoolCourseTitle Text             `json:"schoolCourseTitle"`
	Expression        Text             `json:"expression"`
	PeriodSort        int              `json:"periodSort"`
	RoomName          Text             `json:"roomName"`
	SectionNum        Text             `json:"sectionNum"`
	TeacherID         int64            `json:"teacherID"`
	TermID            int64            `json:"termID"`
	Enrollments       List[Enrollment] `json:"enrollments"`
	StartStopDates    List[StartStop]  `json:"startStopDates"`
}

type Enrollment struct {
	ID        int64      `json:"id"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

type StartStop struct {
	SectionEnrollmentID int64      `json:"sectionEnrollmentId"`
	Start               *time.Time `json:"start"`
	Stop                *time.Time `json:"stop"`
}

type Teacher struct {
	ID          int64 `json:"id"`
	FirstName   Text  `json:"firstName"`
	LastName    Text  `json:"lastName"`
	Email       Text  `json:"email"`
	SchoolPhone Text  `json:"schoolPhone"`
}

type Term struct {
	ID           int64      `json:"id"`
	Title        Text       `json:"title"`
	Abbrev       Text       `json:"abbrev"`
	ParentTermID int64      `json:"parentTermId"`
	StartDate    *time.Time `json:"startDate"`
	EndDate      *time.Time `json:"endDate"`
}

type ReportingTerm struct {
	ID               int64      `json:"id"`
	Title            Text       `json:"title"`
	Abbreviation     Text       `json:"abbreviation"`
	TermID           int64      `json:"termid"`
	StartDate        *time.Time `json:"startDate"`
	EndDate          *time.Time `json:"endDate"`
	SortOrder        int        `json:"sortOrder"`
	SuppressGrades   bool       `json:"suppressGrades"`
	SuppressPercents bool       `json:"suppressPercents"`
}

type Period struct {
	ID           int64 `json:"id"`
	Name         Text  `json:"name"`
	Abbreviation Text  `json:"abbreviation"`
	PeriodNumber int   `json:"periodNumber"`
	SortOrder    int   `json:"sortOrder"`
}

type AssignmentCategory struct {
	ID           int64 `json:"id"`
	Name         Text  `json:"name"`
	Abbreviation Text  `json:"abbreviation"`
}

type Assignment struct {
	ID                   int64      `json:"id"`
	SectionID            int64      `json:"sectionid"`
	CategoryID           int64      `json:"categoryId"`
	Name                 Text       `json:"name"`
	Description          Text       `json:"description"`
	DueDate              *time.Time `json:"dueDate"`
	PointsPossible       Number     `json:"pointspossible"`
	Weight               Number     `json:"weight"`
	IncludeInFinalGrades int        `json:"includeinfinalgrades"`
	PublishScores        int        `json:"publishscores"`
}

type AssignmentScore struct {
	ID           int64  `json:"id"`
	AssignmentID int64  `json:"assignmentId"`
	Score        Text   `json:"score"`
	Percent      Number `json:"percent"`
	LetterGrade  Text   `json:"letterGrade"`
	Comment      Text   `json:"comment"`
	Collected    bool   `json:"collected"`
	Late         bool   `json:"late"`
	Missing      bool   `json:"missing"`
	Exempt       bool   `json:"exempt"`
	Incomplete   bool   `json:"incomplete"`
}

type FinalGrade struct {
	ID              int64      `json:"id"`
	SectionID       int64      `json:"sectionid"`
	ReportingTermID int64      `json:"reportingTermId"`
	Grade           Text       `json:"grade"`
	Percent         Number     `json:"percent"`
	CommentValue    Text       `json:"commentValue"`
	DateStored      *time.Time `json:"dateStored"`
}

type Attendance struct {
	ID           int64      `json:"id"`
	AttCodeID    int64      `json:"attCodeid"`
	AttDate      *time.Time `json:"attDate"`
	AttComment   Text       `json:"attComment"`
	CCID         int64      `json:"ccid"`
	PeriodID     int64      `json:"periodid"`
	TotalMinutes Number     `json:"totalMinutes"`
}

type AttendanceCode struct {
	ID          int64 `json:"id"`
	AttCode     Text  `json:"attCode"`
	CodeType    int   `json:"codeType"`
	Description Text  `json:"description"`
	SortOrder   int   `json:"sortorder"`
}

type NotInSessionDay struct {
	ID          int64      `json:"id"`
	CalendarDay *time.Time `json:"calendarDay"`
	CalType     Text       `json:"calType"`
	Description Text       `json:"description"`
}
