---
version: 1
slug: "web-src-app-tsx"
primary_target: "web/src/App.tsx"
related_targets: ["web/src"]
---

# App shell and views (web/src)

Mode: Operate. Students check their next class, newest score, or attendance in seconds, between classes, on phone, iPad, and desktop. Seven destinations: Home, Schedule, Classes (+ course detail), Grades, Assignments, Attendance, Report.

Constraints: show server values only, label what is missing, never average. Course names are bilingual and stay as sent. Light and dark themes follow the system setting. Respect reduced motion and reduced transparency.

## Direction contract

THESIS: A calm glass "window" floating over slowly drifting colour, where today's timetable reads like light: the class you are in right now is the brightest object on the screen. It refuses PowerSchool's grey table-of-links portal and the generic admin-dashboard card grid.

OWN-WORLD: Ground is a drifting OKLCH gradient mesh (ink-teal, indigo and moss in dark; mist, sky and pistachio in light). Surfaces are frosted glass: 8–14% white fill, 24px backdrop blur, a 1px top-lit inner highlight, and a soft offset shadow. The radius family is 28 (window), 20 (tile) and 12 (control). Lime is selection and primary; cyan is time and "now"; each course has one hue chip from a 10-hue OKLCH ring. IB grades 7→1 run along a lime→teal→amber→rose ramp. Geist with tabular figures, falling back to PingFang SC for Chinese.

STORY: The visitor opens the app and sees, without tapping, which class is now or next, where, and the countdown, then today's run of classes, then the newest scores. Every deeper view (week, course, term grades, flags, absences, report cards) is one tap from the rail or tab bar.

FIRST VIEWPORT: Desktop/iPad: a floating glass icon rail on the left, detached from a large glass window. The window header holds the greeting and date (left) and "Updated 3 min ago" with a refresh control (right). Inside: a wide Now/Next tile top-left with course hue, title, room, teacher and countdown; Today's timeline as a vertical track with a cyan now-line on the right; latest scores and a week strip (Mon–Fri with A/B day codes) underneath. Phone: header, then the Now/Next tile full-width, then the timeline, with a glass bottom tab bar (Home, Schedule, Classes, Assignments, More).

FORM: Pinned by the user brief (glassmorphism, visionOS-style dashboard reference); no concept roll. Signature move: the live now-line. The current meeting's glass tile is lit from within in its course hue, and the same now-line crosses the week grid in Schedule. Motion grammar: route content settles from 6px blur to sharp over 220ms with ease-out; the now-line advances every minute; the mesh drifts over about 60s and freezes under reduced motion.

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
