package student

import (
	"context"
	"testing"
	"time"

	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/fake"
	"github.com/HenryXiaoYang/powerschool-mirror/server/internal/powerschool/mobile"
)

func txt(s string) mobile.Text { return mobile.Text{Value: s, Valid: true} }

func TestPlainText(t *testing.T) {
	got := plainText("<p>Doors open at <b>6pm</b>&nbsp;&amp; close at 9.</p><p>Bring ID.</p>")
	if want := "Doors open at 6pm & close at 9.\nBring ID."; got != want {
		t.Errorf("plainText = %q, want %q", got, want)
	}
}

func TestAnnouncements(t *testing.T) {
	data := []mobile.StudentData{{Bulletins: mobile.List[mobile.Bulletin]{
		{ID: 1, Name: txt("Current"), StartDate: txt("2026-11-01Z"), EndDate: txt("2026-12-01")},
		{ID: 2, Name: txt("Expired"), EndDate: txt("2026-10-01")},
		{ID: 3, Name: txt("Future"), StartDate: txt("2026-12-01")},
		{ID: 4}, // empty
		{ID: 5, Body: txt("Open ended"), StartDate: txt("2026-11-05")}, // newer
	}}}
	got := announcements([][]mobile.StudentData{data, data}, "2026-11-08") // the same set from two sessions
	if len(got) != 2 || got[0].ID != 5 || got[1].ID != 1 || got[1].Start != "2026-11-01" {
		t.Errorf("announcements = %+v", got)
	}
}

func TestLinks(t *testing.T) {
	data := []mobile.StudentData{{CustomPage: mobile.List[mobile.CustomPage]{
		{FormTitle: txt("Lunch menu"), RedirectURL: txt("/menu")},
		{FormTitle: txt("Book a conference"), RedirectURL: txt("https://school.example.org/ptc")},
		{FormTitle: txt("Bad"), RedirectURL: txt("javascript:alert(1)")},
		{FormTitle: txt("Protocol-relative"), RedirectURL: txt("//evil.example/x")},
		{FormTitle: txt("Dup"), RedirectURL: txt("https://school.example.org/ptc")},
		{FormTitle: txt("家长会预约"), RedirectURL: txt("http://a.example/p")},
	}}}
	got := links([][]mobile.StudentData{data}, "https://school.example.org")
	if len(got) != 3 {
		t.Fatalf("links = %+v", got)
	}
	if !got[0].Conference || !got[1].Conference || got[2].Conference || got[2].URL != "https://school.example.org/menu" {
		t.Errorf("order or resolution wrong: %+v", got)
	}
	if got := links([][]mobile.StudentData{data}, ""); len(got) != 2 { // relative link needs a portal
		t.Errorf("without a portal: %+v", got)
	}
}

func TestParentLinkFlow(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 4, 10, 0, 0, 0, shanghai)
	svc := NewService(fake.Mobile{Now: func() time.Time { return now }}, fake.Portal{}, shanghai).WithPortalURL("https://school.example.org")
	svc.now = func() time.Time { return now }

	creds, err := svc.Login(ctx, "alex", "pw")
	if err != nil {
		t.Fatal(err)
	}
	snap, err := svc.Snapshot(ctx, creds, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Children) != 1 || snap.ParentLinked || len(snap.Links) != 0 || len(snap.Announcements) != 1 {
		t.Fatalf("student alone: children=%v linked=%v links=%v ann=%d", snap.Children, snap.ParentLinked, snap.Links, len(snap.Announcements))
	}

	if _, err := svc.LinkParent(ctx, creds, "parent1", fake.WrongPassword); err == nil {
		t.Fatal("a wrong parent password must fail")
	}
	linked, err := svc.LinkParent(ctx, creds, "parent1", "pw")
	if err != nil {
		t.Fatal(err)
	}
	snap, err = svc.Snapshot(ctx, linked, 1002)
	if err != nil {
		t.Fatal(err)
	}
	if snap.ChildID != 1002 || snap.Profile.FirstName != "Jamie" || !snap.ParentLinked {
		t.Errorf("child 1002: id=%d name=%q linked=%v", snap.ChildID, snap.Profile.FirstName, snap.ParentLinked)
	}
	if len(snap.Children) != 2 || snap.Children[0].ID != 1001 || snap.Children[1].Name != "Jamie" {
		t.Errorf("children (own student first, parent's second deduped): %+v", snap.Children)
	}
	if len(snap.Links) != 2 || !snap.Links[0].Conference || snap.Links[1].URL != "https://school.example.org/menu" {
		t.Errorf("links: %+v", snap.Links)
	}
	if len(snap.Announcements) != 1 {
		t.Errorf("announcements are school-wide and must not double: %d", len(snap.Announcements))
	}

	// An id this sign-in cannot open falls back to the first student.
	if snap, _ = svc.Snapshot(ctx, linked, 4242); snap.ChildID != 1001 {
		t.Errorf("unknown child → %d", snap.ChildID)
	}
	if _, err := svc.Schedule(ctx, linked, 1002, 2026, 11); err != nil {
		t.Errorf("schedule for the parent's child: %v", err)
	}
	if snap, _ = svc.Snapshot(ctx, svc.UnlinkParent(linked), 1002); snap.ChildID != 1001 || len(snap.Links) != 0 {
		t.Errorf("after unlink: id=%d links=%d", snap.ChildID, len(snap.Links))
	}
}
