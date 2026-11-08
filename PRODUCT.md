# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

- `server/`: Go + gin. API-only service; deployment is out of scope.
- `web/`: Vite + React + Tailwind 4 + shadcn, React Router 8, TanStack Query persisted to IndexedDB.
- Types generated from Go DTOs with tygo. Mock mode (`MOCK=1`) serves a synthetic fixture from Go.

## Users

High-school students (and classmates) at one school (Asia/Shanghai, boarding, IB grades 1–7). They check grades, today's classes, and attendance in short moments between classes, on phone, iPad, and desktop.

## Product Purpose

A fast, pleasant mirror of the student's PowerSchool data. The official portal is ugly and hard to use; this exists so checking a new score or the next class takes seconds. Success: students open this instead of PowerSchool.

## Positioning

Reads the same data PowerSchool's mobile app uses (the SOAP `getStudentData` call) plus archived report cards from the web portal. It presents that data in one place, joined and in local time.

## Operating Context

- 2-week A/B timetable: section expressions like `1(MA,FB)`, day codes MA…FB. Boarding blocks such as Breakfast, Study Hall and Lights Out appear as sections.
- Dates arrive in UTC; local midnight is `16:00Z`.
- Attendance returns exception records only. There are no "present" rows.
- Term grades (`finalGrades`) can be absent until the school posts them.
- Report cards are school-generated HTML pages, bilingual English/中文, whose template changes between years.
- Course names are bilingual and shown exactly as sent.

## Capabilities and Constraints

- Views: Home (today), Classes, Grades, Assignments, Attendance, Schedule (real meeting times, week view), Student report (profile + archived report cards).
- Read-only. No grade calculation: never invent a course average; show server values only. A missing score is not zero.
- Credentials live in the browser, encrypted with a non-extractable WebCrypto key. They are kept in sessionStorage for the tab, or persisted with "Remember me". An expired session is refreshed silently.
- The Go relay is stateless and talks to the one school host only. Its session token is AES-GCM sealed. It never logs bodies or tickets.
- English UI. Multi-student switching is not needed now (the account has one student).

## Brand Commitments

The user wants a modern, clean, convenient UI. Binding style: glassmorphism, like a visionOS-style dashboard of frosted panels over a rich backdrop. Dark and light themes follow the system setting. Not PowerSchool's look.

## Evidence on Hand

Real data shapes were observed from a live account (kept outside the repo). The repo fixture is synthetic only. No real student data, names, or report cards may be committed.

## Product Principles

1. Seconds, not clicks: the most likely question (next class, newest score) is answered on first view.
2. Server truth only: show what PowerSchool says, label what is missing, never fabricate.
3. Private by default: student data stays on the device and in the stateless relay, never stored server-side.
4. One system everywhere: phone, iPad and desktop share one component language.

## Accessibility & Inclusion

Glass must stay legible: text contrast meets WCAG AA over any backdrop state. Respect reduced-motion (the backdrop drifts) and reduced-transparency preferences.
