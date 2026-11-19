package student

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/mobile"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/portal"
)

// Mobile is the PowerSchool mobile API (implemented by mobile.Client and fake.Mobile).
type Mobile interface {
	Login(ctx context.Context, username, password string) (mobile.Session, error)
	StudentData(ctx context.Context, s mobile.Session) ([]mobile.StudentData, error)
	SectionTimes(ctx context.Context, s mobile.Session, year, month int) ([]mobile.StudentData, error)
}

// Portal is the PowerSchool web portal (implemented by portal.Client and fake.Portal).
type Portal interface {
	Login(ctx context.Context, username, password string) (portal.Session, error)
	ReportCards(ctx context.Context, s portal.Session) ([]portal.ReportCardRef, error)
	ReportCard(ctx context.Context, s portal.Session, id string) (string, error)
}

// Credentials are both upstream sessions. Portal is nil when the portal sign-in failed;
// everything except report cards still works.
type Credentials struct {
	Mobile mobile.Session  `json:"m"`
	Portal *portal.Session `json:"p,omitempty"`
	// Parent is an optional second sign-in (a parent account) used alongside the first, so
	// a student can unlock parent-only data and open the parent's children.
	Parent *Linked `json:"x,omitempty"`
}

// Linked is the parent sign-in attached to a session.
type Linked struct {
	Mobile mobile.Session  `json:"m"`
	Portal *portal.Session `json:"p,omitempty"`
}

// ErrPortalUnavailable means this session has no portal sign-in (report cards are off).
var ErrPortalUnavailable = errors.New("student: portal session unavailable")

type Service struct {
	mobile    Mobile
	portal    Portal
	loc       *time.Location
	now       func() time.Time
	portalURL string
}

// WithPortalURL sets the school's PowerSchool origin, used to resolve the school's relative links.
func (s *Service) WithPortalURL(origin string) *Service {
	s.portalURL = strings.TrimRight(origin, "/")
	return s
}

func NewService(m Mobile, p Portal, loc *time.Location) *Service {
	return &Service{mobile: m, portal: p, loc: loc, now: time.Now}
}

// Login signs in to both channels concurrently. The mobile sign-in decides success.
func (s *Service) Login(ctx context.Context, username, password string) (Credentials, error) {
	// Portal first, never in parallel: a portal sign-in revokes any mobile ticket issued before it.
	ps, portErr := s.portal.Login(ctx, username, password)
	ms, err := s.mobile.Login(ctx, username, password)
	if err != nil {
		return Credentials{}, err
	}
	creds := Credentials{Mobile: ms}
	if portErr == nil {
		creds.Portal = &ps
	} else {
		slog.Warn("portal sign-in failed; report cards disabled for this session", "err", portErr)
	}
	return creds, nil
}

// LinkParent signs in a parent account and attaches it to the session. The student sign-in
// stays the primary one; the parent's children join the switcher and its data fills the
// parent-only sections. A wrong parent password fails here and leaves c unchanged.
func (s *Service) LinkParent(ctx context.Context, c Credentials, username, password string) (Credentials, error) {
	ps, portErr := s.portal.Login(ctx, username, password)
	ms, err := s.mobile.Login(ctx, username, password)
	if err != nil {
		return Credentials{}, err
	}
	linked := &Linked{Mobile: ms}
	if portErr == nil {
		linked.Portal = &ps
	}
	c.Parent = linked
	return c, nil
}

// UnlinkParent detaches the parent sign-in.
func (s *Service) UnlinkParent(c Credentials) Credentials {
	c.Parent = nil
	return c
}

// Snapshot builds the view of one child. child 0, or an id this sign-in cannot open, shows
// the primary session's first student.
func (s *Service) Snapshot(ctx context.Context, c Credentials, child int64) (Snapshot, error) {
	primary, err := s.mobile.StudentData(ctx, c.Mobile)
	if err != nil {
		return Snapshot{}, err
	}
	var parent []mobile.StudentData
	if c.Parent != nil {
		if parent, err = s.mobile.StudentData(ctx, c.Parent.Mobile); err != nil {
			return Snapshot{}, err
		}
	}
	d, ok := pick(child, primary, parent)
	if !ok {
		return Snapshot{}, fmt.Errorf("%w: no student on this account", powerschool.ErrUpstream)
	}
	now := s.now()
	snap := Build(d, s.loc, now)
	sets := [][]mobile.StudentData{primary, parent}
	snap.Children = children(primary, parent)
	snap.ChildID = recordID(d)
	snap.ParentLinked = c.Parent != nil
	snap.Announcements = announcements(sets, now.In(s.loc).Format(time.DateOnly))
	snap.Links = links(sets, s.portalURL)
	snap.PortalURL = s.portalURL
	return snap, nil
}

func (s *Service) Schedule(ctx context.Context, c Credentials, child int64, year, month int) (Schedule, error) {
	sess := c.Mobile
	if c.Parent != nil && child != 0 && !slices.Contains(c.Mobile.StudentIDs, child) && slices.Contains(c.Parent.Mobile.StudentIDs, child) {
		sess = c.Parent.Mobile
	}
	data, err := s.mobile.SectionTimes(ctx, sess, year, month)
	if err != nil {
		return Schedule{}, err
	}
	d, ok := pick(child, data)
	if !ok {
		return Schedule{}, fmt.Errorf("%w: no student on this account", powerschool.ErrUpstream)
	}
	return BuildSchedule(d.Sections, s.loc, year, month), nil
}

func (s *Service) ReportCards(ctx context.Context, c Credentials) ([]ReportCard, error) {
	if c.Portal == nil {
		return nil, ErrPortalUnavailable
	}
	refs, err := s.portal.ReportCards(ctx, *c.Portal)
	if err != nil {
		return nil, err
	}
	out := make([]ReportCard, 0, len(refs))
	for _, r := range refs {
		rc := ReportCard{ID: r.ID, Name: r.Name}
		if d, err := time.Parse("01/02/2006", r.Date); err == nil {
			rc.Date = d.Format(time.DateOnly)
		}
		out = append(out, rc)
	}
	return out, nil
}

func (s *Service) ReportCard(ctx context.Context, c Credentials, id string) (string, error) {
	if c.Portal == nil {
		return "", ErrPortalUnavailable
	}
	return s.portal.ReportCard(ctx, *c.Portal, id)
}

// IsSessionError reports whether err means the client should sign in again.
func IsSessionError(err error) bool {
	return errors.Is(err, powerschool.ErrSessionExpired) || errors.Is(err, ErrPortalUnavailable)
}
