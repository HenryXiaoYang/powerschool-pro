---
name: Mirror
description: A calm frosted-glass window over a drifting deep-ocean backdrop, where the class happening now is the brightest thing on screen.
colors:
  primary: "oklch(1 0 0)"
  primary-dark: "oklch(0.97 0.008 250)"
  on-primary: "oklch(0.26 0.06 262)"
  on-primary-dark: "oklch(0.24 0.06 262)"
  now: "oklch(0.49 0.14 228)"
  now-dark: "oklch(0.82 0.13 215)"
  today: "oklch(0.5 0.15 258)"
  today-dark: "oklch(0.86 0.09 235)"
  warn: "oklch(0.52 0.12 70)"
  warn-dark: "oklch(0.85 0.14 80)"
  alert: "oklch(0.55 0.19 20)"
  alert-dark: "oklch(0.75 0.16 20)"
  grade-7: "oklch(0.86 0.12 128)"
  grade-6: "oklch(0.86 0.12 150)"
  grade-5: "oklch(0.86 0.12 176)"
  grade-4: "oklch(0.86 0.12 88)"
  grade-3: "oklch(0.86 0.12 58)"
  grade-2: "oklch(0.86 0.12 30)"
  grade-1: "oklch(0.86 0.12 18)"
  ground: "oklch(0.95 0.02 250)"
  ground-dark: "oklch(0.2 0.045 260)"
  mesh-sky: "oklch(0.85 0.08 245)"
  mesh-pale-cyan: "oklch(0.9 0.07 210)"
  mesh-periwinkle: "oklch(0.85 0.07 278)"
  mesh-ice: "oklch(0.91 0.05 230)"
  mesh-cobalt-dark: "oklch(0.37 0.12 262)"
  mesh-azure-dark: "oklch(0.39 0.09 238)"
  mesh-cyan-dark: "oklch(0.4 0.075 212)"
  mesh-indigo-dark: "oklch(0.3 0.1 276)"
  glass: "oklch(1 0 0 / 0.46)"
  glass-dark: "oklch(0.18 0.035 260 / 0.38)"
  glass-strong: "oklch(1 0 0 / 0.78)"
  glass-strong-dark: "oklch(0.3 0.035 258 / 0.82)"
  glass-sunk: "oklch(0.2 0.03 240 / 0.05)"
  glass-sunk-dark: "oklch(0 0 0 / 0.16)"
  glass-edge: "oklch(1 0 0 / 0.7)"
  glass-edge-dark: "oklch(1 0 0 / 0.12)"
  foreground: "oklch(0.24 0.03 245)"
  foreground-dark: "oklch(0.97 0.005 235)"
  muted-foreground: "oklch(0.45 0.03 245)"
  muted-foreground-dark: "oklch(0.8 0.015 235)"
  faint: "oklch(0.5 0.025 245)"
  faint-dark: "oklch(0.68 0.02 235)"
  border: "oklch(0.3 0.03 245 / 0.12)"
  border-dark: "oklch(1 0 0 / 0.1)"
  ring: "oklch(0.6 0.13 225)"
  ring-dark: "oklch(0.82 0.13 215)"
typography:
  display:
    fontFamily: "Geist Variable, PingFang SC, Hiragino Sans GB, Microsoft YaHei, system-ui, sans-serif"
    fontSize: "1.875rem"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "-0.025em"
    fontFeature: "tnum"
  headline:
    fontFamily: "Geist Variable, PingFang SC, Hiragino Sans GB, Microsoft YaHei, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 600
    lineHeight: 1.33
    letterSpacing: "-0.025em"
  title:
    fontFamily: "Geist Variable, PingFang SC, Hiragino Sans GB, Microsoft YaHei, system-ui, sans-serif"
    fontSize: "15px"
    fontWeight: 600
    lineHeight: 1.4
    letterSpacing: "-0.025em"
  body:
    fontFamily: "Geist Variable, PingFang SC, Hiragino Sans GB, Microsoft YaHei, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 500
    lineHeight: 1.43
  label:
    fontFamily: "Geist Variable, PingFang SC, Hiragino Sans GB, Microsoft YaHei, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 400
    lineHeight: 1.33
  label-small:
    fontFamily: "Geist Variable, PingFang SC, Hiragino Sans GB, Microsoft YaHei, system-ui, sans-serif"
    fontSize: "11px"
    fontWeight: 500
    lineHeight: 1.33
rounded:
  md: "12px"
  lg: "16px"
  xl: "20px"
  2xl: "28px"
  full: "9999px"
spacing:
  row: "10px"
  gap: "12px"
  gap-wide: "16px"
  tile: "16px"
  tile-wide: "20px"
  gutter: "20px"
  gutter-wide: "32px"
components:
  window:
    backgroundColor: "{colors.glass}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.2xl}"
  panel:
    backgroundColor: "{colors.glass}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.xl}"
    padding: "20px"
  well:
    backgroundColor: "{colors.glass-sunk}"
    rounded: "{rounded.lg}"
    padding: "10px 12px"
  bead-selected:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.on-primary}"
    rounded: "{rounded.full}"
  segment:
    textColor: "{colors.muted-foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.full}"
    height: "36px"
    padding: "0 16px"
  segment-selected:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.on-primary}"
    rounded: "{rounded.full}"
    height: "36px"
  nav-rail-item:
    textColor: "{colors.muted-foreground}"
    rounded: "{rounded.full}"
    size: "44px"
  tab-bar-item:
    textColor: "{colors.muted-foreground}"
    typography: "{typography.label-small}"
    rounded: "{rounded.full}"
    height: "56px"
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.on-primary}"
    typography: "{typography.title}"
    rounded: "{rounded.md}"
    height: "44px"
  field-glass:
    backgroundColor: "{colors.glass}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.full}"
    height: "40px"
    padding: "0 12px"
  grade-badge:
    backgroundColor: "{colors.grade-7}"
    textColor: "{colors.foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    height: "32px"
    width: "32px"
  flag-chip:
    textColor: "{colors.alert}"
    typography: "{typography.label-small}"
    rounded: "{rounded.full}"
    height: "20px"
    padding: "0 8px"
---

# Design System: Mirror

## Overview

**Creative North Star: "The Window Over Deep Water"**

Mirror is one frosted window floating over slowly drifting ocean colour. Everything a student touches is the same glass, a single material at different depths, and the backdrop moves so slowly you barely notice it. The screen has one bright object, the class happening now, lit from within in its course's hue. Selection is frosted white: a bead of brighter glass that slides from one choice to the next. There are two blues for time. Cyan means "now" and moves; sky blue means "today" and stays put.

The density suits quick checks between classes. Tiles hold a few facts each, numbers use tabular figures, and headings are plain sentence-case titles with no ornament. Phone, iPad and desktop all use the same pieces. A floating icon rail sits beside the window on wide screens, and a glass tab bar takes its place on phones. Light and dark follow the system setting: ice blue with milky glass and ink text, or navy with smoked glass and white text.

Motion is the window settling. New pages come up from a soft 6px blur, the selection bead travels to where you tapped, week pages slide the way you paged, and the timeline fills down to the present moment while a dot pulses at its tip. Reduced motion stops all of it. Reduced transparency swaps in near-opaque glass.

**Key Characteristics:**
- One surface material (frosted glass) over a four-blob deep-ocean mesh that drifts over about 60 to 90 seconds.
- Frosted-white selection bead that slides between choices: a CSS transform in the rail and tab bar, view transitions for in-page filters and chips.
- Cyan "now" for live time, sky-blue "today" for date labels, and course hues taken in order from a 10-hue OKLCH ring.
- IB grades 7 to 1 on a fixed ramp from lime through teal and amber to rose. Lime appears nowhere else.
- Geist Variable for every role, with tabular figures. Chinese falls back to PingFang SC.

## Colors

The palette is cool and watery: navy or ice-blue ground, colourless glass, one white for selection, two time blues, and saturated colour only where it carries information (course hue, grade, flag).

### Primary
- **Frosted White Bead** (light `primary`, dark `primary-dark`): fills the current selection. That means the active rail or tab item, the chosen status filter or term, the selected day in Schedule, today in the Home week strip, and the sign-in submit button. It always carries the bead's top-lit inner highlight and soft lift (see Elevation) and sits behind deep indigo ink (`on-primary`).

### Secondary
- **Now Cyan** (`now` / `now-dark`): live time only. It colours the timeline fill and its pulsing dot, the line across free time in the week grid, the class-progress bar, the "Ends in" countdown, and the text caret. It also tones the "collected" flag.
- **Today Sky** (`today` / `today-dark`): text tint for today's date in the Schedule week header and for the current term column in Grades. It is a label colour, not a fill.

### Tertiary
- **Course Ring**: each academic course gets one hue, in timetable order, from a fixed ring of ten (130, 200, 25, 270, 165, 60, 305, 235, 95, 340). Lightness and chroma come from the theme (0.66/0.14 light, 0.80/0.13 dark). Non-academic blocks (Breakfast, Study Hall) stay neutral in `faint`. The hue shows up as a 10px dot beside the course name and as the inner glow of the live class.
- **Grade Ramp** (`grade-7` … `grade-1`): badge fills at L 0.86 / C 0.12 with ink at L 0.27 / C 0.06 in the same hue. 7 is lime (128), 6 is green (150), 5 is teal (176), 4 is yellow (88), 3 is amber (58), 2 is coral (30), 1 is rose (18). IGCSE letters map onto the same steps. A grade that has no step gets a neutral sunk badge.
- **Warn Amber / Alert Rose** (`warn`, `alert`): late, incomplete and tardy use amber. Missing and absent use rose. Both appear as tinted text inside a 40%-opacity border of the same colour, never as a solid fill.

### Neutral
- **Ice Ground / Navy Ground** (`ground` / `ground-dark`): the page colour under the mesh. It also paints `html`, so the edges never flash.
- **Ocean Mesh** (`mesh-*`): four blurred 70vmax blobs (80px blur, 0.85 opacity, with the fourth at 0.5). Light uses sky, pale cyan, periwinkle and ice. Dark uses cobalt, azure, cyan and indigo.
- **Glass family** (`glass`, `glass-strong`, `glass-sunk`, `glass-edge`): the surface fill, the reduced-transparency or popover fill, the nested-well tint, and the 1px rim. In dark, `glass` is a navy pane (`oklch(0.18 0.035 260 / 0.38)`), not a white wash: white glass lifts the cobalt blobs until secondary text fails AA.
- **Ink** (`foreground`, `muted-foreground`, `faint`): primary text, secondary text and labels, and tertiary text (placeholder icons, "/ points", neutral course dots).
- **Hairline** (`border`) for dividers and well rims. **Ring** (`ring`) for the 2px focus outline at a 2px offset.

### Named Rules
**The One Bead Rule.** Frosted white marks what is selected, and nothing else. A screen shows one bead per choice group, and it moves between options rather than appearing in a new place.

**The Two Clocks Rule.** Cyan is time in motion: the now-line, now-dot, countdown and progress bar. Sky blue is today's date or the current term, applied as a text tint. Neither stands in for the other.

**The Lime Is a Grade Rule.** Lime appears only as IB grade 7 on the ramp. It never marks selection, focus or brand.

**The Hue Belongs to the Course Rule.** A course keeps one ring hue on every view (dot, lit tile, week-grid block). Hues never encode status.

## Typography

**Display Font:** Geist Variable (with PingFang SC, Hiragino Sans GB, Microsoft YaHei, system-ui)
**Body Font:** the same family
**Label/Mono Font:** the same family with tabular figures

**Character:** One neutral grotesque throughout, set tight at the top of the scale and medium weight in running text. Bilingual course names display as sent, and Chinese glyphs come from the fallback stack.

### Hierarchy
- **Display** (600, 1.875rem, 1.2, -0.025em): the Now/Next course title from `sm` up (1.5rem on phones), the sign-in wordmark, and the large attendance counts.
- **Headline** (600, 1.5rem, 1.33, -0.025em): the page title and greeting in the window header, and the Home counts.
- **Title** (600, 15px, -0.025em): Panel headings. These are plain sentence-case headings.
- **Body** (500 for item names, 400 for prose, 0.875rem, 1.43): list rows, facts, table cells and filter labels. Explanatory copy stays within about 24rem (`max-w-sm`).
- **Label** (400, 0.75rem): fact labels, timestamps, secondary lines and the "Updated 3 min ago" freshness note.
- **Label Small** (500, 11px): tab bar labels, flag chips and week-strip day codes. 11px is the floor.

### Named Rules
**The One Family Rule.** Every role uses Geist Variable. Hierarchy comes from size, weight (400/500/600) and tight tracking. Never add a separate display face.

**The Tabular Figures Rule.** Every number that can change or line up (times, scores, counts, dates, grade badges) uses tabular figures.

## Layout

The rail appears only with room and a hovering pointer: at least 768px wide, 480px tall, and a mouse or trackpad (the `rail` variant). Touch tablets, landscape phones and short windows keep the labelled tab bar, capped at 32rem and centred, because rail labels live in hover tooltips. Compact control sizes (36–40px) also apply only with a fine pointer; touch keeps 44px at every width. The week grid needs about 150px per day, so it starts at `lg` (1024px); below that Schedule uses the day list. Every class block, day-list row and timeline row opens its course page, so a truncated name is always one tap from in full. On `md` (768px) and wider, a floating glass rail (pill-shaped, 44px icon targets) sits detached to the left of one large glass window, with a 20px gap (24px at `lg`). The shell is centred at a maximum of 1500px with a 20px outer gutter (32px at `lg`), and the window is at least full-viewport height. Below `md` the window has no glass of its own. The page sits straight on the mesh and a glass tab bar floats 12px above the bottom safe area. Content keeps 112px of bottom padding so the bar never covers it.

The window header runs at least 76px tall (88px at `md`), with the title on the left and freshness and refresh on the right. Content padding is 12px on phones and 24px from `sm`. Tiles stack with a 12px gap (16px from `sm`). At `lg` (1024px), Home becomes a 12-column grid split 7/5, and the columns size to their content. On phones the same tiles follow task order (Now/Next, timeline, week, scores, term). Tile padding is 16px (20px from `sm`) and list rows sit on 10px vertical padding. From `md` up the scrollbar gutter is reserved so page width never shifts.

## Elevation & Depth

Depth comes from a single glass material. There are no shadow levels to choose between. Every surface uses the same recipe: translucent fill, 24px backdrop blur with 150% saturation, a 1px rim, a 1px inner top highlight, and one long, soft drop shadow that reads as a pane floating over water. Content nested inside glass sits in sunk wells that add a tint and a hairline, with no second blur. Two things rise above the plane. The selection bead is brighter glass with its own highlight and lift. The live class glows in its course hue.

### Shadow Vocabulary
- **Glass float** (`box-shadow: inset 0 1px 0 var(--glass-highlight), 0 18px 40px -18px var(--glass-shadow)`): every glass surface (window, rail, tab bar, Panels, filter bars, fields, sheet).
- **Bead lift** (`box-shadow: inset 0 1px 0 oklch(1 0 0 / 0.9), inset 0 0 0 1px oklch(1 0 0 / 0.6), 0 6px 18px -6px oklch(0.2 0.06 262 / 0.45)`): the frosted-white selection and the primary button.
- **Lit from within** (radial course-hue wash from the top-left corner (0.42 light, 0.2 dark), flat 0.14 hue tint, hue rim at 0.5, `0 18px 40px -16px` hue glow): the live class only, wherever it appears (the Now tile, its timeline row, its week-grid block, and its row in the day list).

### Named Rules
**The One Material Rule.** Glass is the only surface, at two depths at most: the window, and the tiles and controls on it (on phones the tiles sit straight on the mesh). Anything inside a tile becomes a sunk well. Dropping the tiles' own blur was measured: it saved little and lightened them, so they keep it.

**The Lit From Within Rule.** The live class is the brightest object on screen. No other element may use a coloured glow.

## Shapes

Corners are soft and nest by depth. The window and sign-in card use 28px, Panels and the Now tile use 20px, wells and day cells use 16px, and rows, badges and the sign-in button use 12px. Anything you pick from a set (rail, tab bar, segmented filters, select triggers, search field, bead, flag chips, course dots) is fully round. Borders are always 1px: a glass rim, a hairline, or a 40% tone rim on chips and attendance codes. The phone sheet rises with 28px top corners.

## Components

### Selection Bead (signature)
A frosted-white pill behind the chosen item, with the bead lift shadow. In the rail and tab bar it is one element that slides by CSS transform (380ms, `cubic-bezier(0.16, 1, 0.3, 1)`): 50px steps in the rail, one fifth of the bar per tab. In-page choices (assignment status, phone term and day chips) use a view-transition name (`status-pill`, `term-pill`, `day-pill`) and travel for 340ms; those swaps drop the old view immediately and their labels carry their own names, so only the bead moves and it never covers its label. The label switches to `on-primary`.

### Segmented Filters
- **Style:** a glass pill bar (4px inset) holding round segments, 44px on phones and 36px from `sm`, with 16px side padding. The text is medium-weight 14px in muted ink.
- **State:** the selected segment carries the bead. Others brighten to full ink on hover, and every segment presses to 0.96 scale.

### Buttons
- **Shape:** 12px radius for the primary button. Icon and toolbar buttons are fully round.
- **Primary:** frosted white with indigo ink and the bead lift, 44px tall, 15px semibold. It is used for sign-in. Icon and toolbar buttons (refresh, week paging, retry) are 44px on phones and 36px from `md`.
- **Ghost:** transparent, with a muted tint on hover. Used for refresh, week paging and "Today".
- **Secondary on glass:** a round glass button, used for printing report cards.
- **Focus:** a 2px `ring` outline at a 2px offset everywhere.

### Chips
- **Flag chip:** 20px tall, round, 11px medium capitalised word, with tinted text in a 40% rim of the same tone (rose missing, amber late/incomplete, cyan collected, muted exempt and not counted).
- **Attendance code:** 24px tall, 12px radius, semibold tabular code using the same tinted-rim treatment: rose absent, amber tardy, neutral ink for other codes and present.

### Cards / Containers
- **Corner Style:** 20px (Panel), 28px (window).
- **Background:** `glass`, or `glass-strong` under reduced transparency.
- **Shadow Strategy:** Glass float.
- **Border:** 1px `glass-edge`.
- **Internal Padding:** 16px, or 20px from `sm`. The Panel header is a plain 15px semibold title with an optional muted action link ("All ›", "Week ›") on the right.

### Inputs / Fields
- **Style:** filter fields are round glass pills, 44px on phones and 40px from `sm` (select triggers, and a search field with a faint leading icon). Sign-in fields are 44px with a 12px radius on the hairline input stroke.
- **Focus:** the 2px ring outline. The caret is Now Cyan.
- **Error:** a rose message in a 40% rose rim, with no fill.

### Navigation
- **Rail (768px+, 480px+ tall, fine pointer):** a vertical glass pill with 44px round icon buttons (20px Lucide icons at stroke 1.8), muted ink, an accent tint on hover, and a 0.94 press. The active item carries the bead. Tooltips appear on the right, and sign-out sits below a hairline.
- **Tab bar (phones, touch tablets, short windows):** a floating glass pill with five 56px items (icon plus 11px label). "More" opens a glass bottom sheet listing the other sections, and the bead marks whichever item is active.

### Grade Badge
A tabular semibold grade on its ramp colour with same-hue ink, in three sizes (24, 32 and 48px squares that grow with the text, 12px radius, 16px at large). Grades that have no ramp step use a neutral sunk badge.

### Now Line and Timeline
A 2px hairline track. A Now Cyan fill grows down it to the present moment (750ms on entry, then a 600ms glide each minute), and a 10px cyan dot rides its tip with a pulse every 2.4s. Past rows fade to 50%. The live row is lit. In the week grid the cyan line sweeps across free time only, because while a class is running its lit block already marks now.

### Course Dot
A 10px circle in the course's ring hue, used as the course's identity mark on every list.

### States
Empty states show a faint 24px line icon, a medium title, and a muted sentence about what is missing and why. Errors use "Couldn’t load this" with the server message and a "Try again" button. Loading rows are 48px muted skeletons with a 12px radius.

## Do's and Don'ts

### Do:
- **Do** build every surface from glass (24px blur, 150% saturation, 1px rim, top highlight, glass float shadow), and use sunk wells inside it.
- **Do** mark selection with the frosted-white bead and make it travel to the new choice, never just reappear.
- **Do** tie Now Cyan to time that is passing and Today Sky to today's date or the current term.
- **Do** give each course its ring hue on every view, and keep non-academic blocks neutral.
- **Do** settle new route content over 280ms (`cubic-bezier(0.16, 1, 0.3, 1)`): the page rises 4px, and each glass surface fades in from a 6px blur on its own. Never fade or filter the page container: an ancestor with opacity or a filter hides the backdrop from the glass inside, and the panels snap to colour when it ends. Under reduced motion the glass only fades (200ms). Route changes take no view-transition snapshot: the old page leaves at once, because two translucent glass pages cross-fading double-expose.
- **Do** use tabular figures for every time, score, count and grade.
- **Do** stop the mesh drift, settle and timeline animations under reduced motion, and switch to `glass-strong` under reduced transparency.

### Don't:
- **Don't** use lime for selection, focus or brand. Lime belongs to grade 7.
- **Don't** blur a third layer: content inside a tile is a sunk well, never more glass.
- **Don't** give anything but the live class a coloured glow.
- **Don't** fill a surface with Today Sky or Now Cyan. They are text tints, lines and dots.
- **Don't** set uppercase tracked labels above headings. Panel titles are plain sentence-case headings.
- **Don't** set text below 11px.
- **Don't** add a second typeface for display.
