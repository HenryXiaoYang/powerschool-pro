package student

import (
	"context"
	"errors"
	"log/slog"
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
}

// ErrPortalUnavailable means this session has no portal sign-in (report cards are off).
var ErrPortalUnavailable = errors.New("student: portal session unavailable")

type Service struct {
	mobile Mobile
	portal Portal
	loc    *time.Location
	now    func() time.Time
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

func (s *Service) Snapshot(ctx context.Context, c Credentials) (Snapshot, error) {
	data, err := s.mobile.StudentData(ctx, c.Mobile)
	if err != nil {
		return Snapshot{}, err
	}
	// shortcut: first student only; the target accounts have one. Add a student switcher for parent accounts.
	return Build(data[0], s.loc, s.now()), nil
}

func (s *Service) Schedule(ctx context.Context, c Credentials, year, month int) (Schedule, error) {
	data, err := s.mobile.SectionTimes(ctx, c.Mobile, year, month)
	if err != nil {
		return Schedule{}, err
	}
	return BuildSchedule(data[0].Sections, s.loc, year, month), nil
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
