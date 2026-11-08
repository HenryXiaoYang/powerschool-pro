import { CalendarOff, ChevronRight, ClipboardCheck } from 'lucide-react'
import { useLayoutEffect, useRef, useState, type CSSProperties } from 'react'
import { Link } from 'react-router'
import type { Snapshot } from '@/api/student.gen'
import { CourseDot, GradeBadge } from '@/components/glass/badges'
import { Panel } from '@/components/glass/Panel'
import { EmptyState, ErrorState, LoadingRows } from '@/components/glass/states'
import { addDays, clockTime, daysBetween, formatDate, inMinutes, mondayOf } from '@/lib/dates'
import { formatPoints } from '@/lib/grades'
import { useCourses, useSnapshot } from '@/lib/queries'
import { useNow } from '@/lib/useNow'
import { cn } from '@/lib/utils'
import { latestScores, today, type Today } from './today'

type Courses = ReturnType<typeof useCourses>

export function HomePage() {
  const { data: snap, error, refetch, isPending } = useSnapshot()
  const courses = useCourses(snap)
  const now = useNow()
  if (isPending) return <HomeSkeleton />
  if (!snap) return <ErrorState error={error!} onRetry={() => void refetch()} />

  const day = today(snap.schedule, snap.timeZone, now, (m) => !!courses.byId.get(m.courseId)?.academic)
  return (
    <div className="grid grid-cols-1 gap-3 sm:gap-4 lg:grid-cols-12 lg:items-start">
      {/* Phone: one stream in task order. Desktop/iPad landscape: two columns that size to content. */}
      <div className="contents lg:col-span-7 lg:flex lg:flex-col lg:gap-4">
        <NowNext snap={snap} day={day} now={now} courses={courses} />
        <WeekStrip snap={snap} date={day.upcomingDate ?? day.date} today={day.date} />
        <LatestScores snap={snap} courses={courses} />
      </div>
      <div className="contents lg:col-span-5 lg:flex lg:flex-col lg:gap-4">
        <Timeline snap={snap} day={day} now={now} courses={courses} />
        <TermAndAttendance snap={snap} date={day.date} />
      </div>
    </div>
  )
}

function HomeSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-12">
      <Panel className="lg:col-span-7"><LoadingRows rows={3} /></Panel>
      <Panel className="lg:col-span-5 lg:row-span-2"><LoadingRows rows={6} /></Panel>
      <Panel className="lg:col-span-7"><LoadingRows rows={2} /></Panel>
    </div>
  )
}

function NowNext({ snap, day, now, courses }: { snap: Snapshot; day: Today; now: Date; courses: Courses }) {
  const m = day.current ?? day.next
  if (!m) {
    return (
      <Panel className="order-1 lg:order-none">
        <EmptyState icon={CalendarOff} title="Nothing else scheduled this month">
          PowerSchool lists no more class meetings. Check the Schedule for next month.
        </EmptyState>
      </Panel>
    )
  }
  const c = courses.byId.get(m.courseId)
  const hue = courses.hues.get(m.courseId)
  const t = now.getTime()
  const start = Date.parse(m.start)
  const end = Date.parse(m.end)
  const live = m === day.current
  const minsTo = Math.max(0, Math.round((start - t) / 60_000))
  const then = live ? snap.schedule.meetings.find((x) => x.date === day.date && Date.parse(x.start) >= end) : null
  const thenCourse = then && courses.byId.get(then.courseId)

  const minsLeft = Math.max(1, Math.round((end - t) / 60_000))
  const progress = Math.min(100, ((t - start) / (end - start)) * 100)
  const startsIn = m.date === day.date ? inMinutes(minsTo) : `${formatDate(m.date)}`

  return (
    <section
      className={cn('glass relative order-1 overflow-hidden rounded-xl p-5 sm:p-6 lg:order-none', live && 'lit')}
      style={{ '--hue': hue ?? 220 } as CSSProperties}
      aria-label={live ? 'Current class' : 'Next class'}
    >
      <div className="flex items-start gap-3">
        <CourseDot hue={hue} className="mt-3 size-3" />
        <h2 className="text-2xl leading-tight font-semibold tracking-tight text-balance sm:text-3xl">{c?.title ?? 'Class'}</h2>
      </div>
      <dl className="mt-4 grid grid-cols-[auto_1fr] gap-x-4 gap-y-1.5 text-sm sm:grid-cols-4 sm:gap-x-6">
        <Fact label={live ? 'Ends' : 'Starts'} value={live ? inMinutes(minsLeft) : startsIn} accent />
        <Fact label="Time" value={`${clockTime(m.start, snap.timeZone)}–${clockTime(m.end, snap.timeZone)}`} />
        <Fact label="Room" value={c?.room || '—'} />
        <Fact label="Teacher" value={c?.teacher?.name || '—'} />
      </dl>
      {live && (
        <div className="mt-5 flex items-center gap-3">
          <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted" role="progressbar" aria-label="Class progress" aria-valuenow={Math.round(progress)}>
            <div className="h-full rounded-full bg-now transition-[width] duration-700" style={{ width: `${progress}%` }} />
          </div>
          <span className="tabular shrink-0 text-xs font-medium text-now">{minsLeft} min left</span>
        </div>
      )}
      {then && thenCourse && (
        <p className="mt-3 text-sm text-muted-foreground">
          Then <span className="text-foreground">{thenCourse.title}</span> at {clockTime(then.start, snap.timeZone)}
          {thenCourse.room && ` · ${thenCourse.room}`}
        </p>
      )}
    </section>
  )
}

function Fact({ label, value, accent = false }: { label: string; value: string; accent?: boolean }) {
  return (
    <div className="contents sm:block">
      <dt className="text-muted-foreground sm:text-xs">{label}</dt>
      <dd className={cn('tabular font-medium sm:mt-0.5 sm:text-[15px]', accent && 'text-now sm:text-lg sm:leading-tight')}>{value}</dd>
    </div>
  )
}

function Timeline({ snap, day, now, courses }: { snap: Snapshot; day: Today; now: Date; courses: Courses }) {
  const t = now.getTime()
  const list = day.meetings.length ? day.meetings : day.upcoming
  const nowIndex = day.meetings.findIndex((m) => Date.parse(m.start) > t)
  const ol = useRef<HTMLOListElement>(null)
  const [nowY, setNowY] = useState<number | null>(null)
  const [fill, setFill] = useState(0)

  // Where "now" falls on the track: inside the live row at its progress, else at the next row's top.
  useLayoutEffect(() => {
    const rows = ol.current ? [...ol.current.querySelectorAll<HTMLElement>('[data-row]')] : []
    if (!day.meetings.length || !rows.length) return setNowY(null)
    const live = day.current ? day.meetings.indexOf(day.current) : -1
    if (live >= 0) {
      const r = rows[live]
      const f = (t - Date.parse(day.current!.start)) / (Date.parse(day.current!.end) - Date.parse(day.current!.start))
      setNowY(r.offsetTop + f * r.offsetHeight)
    } else if (nowIndex >= 0) {
      setNowY(Math.max(0, rows[nowIndex].offsetTop - 2))
    } else {
      const last = rows[rows.length - 1]
      setNowY(last.offsetTop + last.offsetHeight)
    }
  }, [t, day, nowIndex])
  useLayoutEffect(() => {
    if (nowY !== null && ol.current?.offsetHeight) setFill(nowY / ol.current.offsetHeight)
  }, [nowY])

  return (
    <Panel
      title={day.upcomingDate ? formatDate(day.upcomingDate, { weekday: 'long', day: 'numeric', month: 'short' }) : 'Today'}
      action={<Link to="/schedule" className="-my-3 py-3 text-sm text-muted-foreground hover:text-foreground">Week <ChevronRight className="inline size-4" /></Link>}
      className="order-2 lg:order-none"
    >
      {day.upcomingDate && <p className="-mt-2 mb-2 text-xs text-muted-foreground">No classes today. Next school day:</p>}
      {list.length === 0 ? (
        <EmptyState icon={CalendarOff} title="No classes today" />
      ) : (
        <ol ref={ol} className="relative pl-4">
          <span aria-hidden className="absolute top-0 bottom-0 left-[5px] w-0.5 rounded-full bg-border" />
          {nowY !== null && (
            <>
              <span
                aria-hidden
                className="day-fill absolute top-0 bottom-0 left-[5px] w-0.5 rounded-full bg-now"
                style={{ '--fill': fill } as CSSProperties}
              />
              <span aria-hidden className="now-dot absolute top-0 left-[6px] size-2.5 rounded-full bg-now" style={{ '--y': `${nowY}px` } as CSSProperties} />
            </>
          )}
          {list.map((m) => {
            const c = courses.byId.get(m.courseId)
            const hue = courses.hues.get(m.courseId)
            const past = Date.parse(m.end) <= t
            const live = m === day.current
            return (
              <li key={`${m.courseId}-${m.start}-${m.end}`}>
                <Link
                  to={`/classes/${m.courseId}`}
                  data-row
                  className={cn('relative flex items-center gap-3 rounded-md px-2 py-2 transition-colors hover:bg-accent/50', live && 'lit', past && 'opacity-50')}
                  style={{ '--hue': hue ?? 220 } as CSSProperties}
                >
                  <span className="tabular w-11 shrink-0 text-sm text-muted-foreground">{clockTime(m.start, snap.timeZone)}</span>
                  <CourseDot hue={hue} />
                  <span className={cn('line-clamp-2 min-w-0 flex-1', c?.academic ? 'font-medium' : 'text-muted-foreground')}>{c?.title}</span>
                  {c?.room && <span className="shrink-0 text-xs text-muted-foreground">{c.room}</span>}
                </Link>
              </li>
            )
          })}
        </ol>
      )}
    </Panel>
  )
}

/** The week of the day Home is showing: next week on a weekend, with that day beaded. */
function WeekStrip({ snap, date, today }: { snap: Snapshot; date: string; today: string }) {
  const monday = mondayOf(date)
  const cycle = new Map(snap.schedule.days.map((d) => [d.date, d]))
  const holidays = new Set(snap.holidays.map((h) => h.date))
  return (
    <Panel title={monday > mondayOf(today) ? 'Next week' : 'This week'} className="order-3 lg:order-none">
      <div className="grid grid-cols-5 gap-1.5 sm:gap-2">
        {[0, 1, 2, 3, 4].map((i) => {
          const d = addDays(monday, i)
          const shown = d === date
          const code = cycle.get(d)?.day
          return (
            <Link
              key={d}
              to={`/schedule?day=${d}`}
              className={cn(
                'flex flex-col items-center gap-0.5 rounded-lg py-2.5 transition duration-200 active:scale-[0.96]',
                shown ? 'bead bg-primary text-primary-foreground' : 'hover:bg-accent',
              )}
            >
              <span className={cn('text-xs', !shown && 'text-muted-foreground')}>{formatDate(d, { weekday: 'short' })}</span>
              <span className="tabular text-lg font-semibold">{formatDate(d, { day: 'numeric' })}</span>
              <span className={cn('text-[11px] font-medium', !shown && 'text-muted-foreground')}>
                {holidays.has(d) ? 'Off' : (code ?? '·')}
              </span>
            </Link>
          )
        })}
      </div>
    </Panel>
  )
}

function LatestScores({ snap, courses }: { snap: Snapshot; courses: Courses }) {
  const recent = latestScores(snap.assignments)
  return (
    <Panel
      title="Latest scores"
      action={<Link to="/assignments" className="-my-3 py-3 text-sm text-muted-foreground hover:text-foreground">All <ChevronRight className="inline size-4" /></Link>}
      className="order-4 lg:order-none"
    >
      {recent.length === 0 ? (
        <EmptyState icon={ClipboardCheck} title="No scores yet">Scored work appears here as teachers publish it.</EmptyState>
      ) : (
        <ul className="divide-y divide-border">
          {recent.map((a) => (
            <li key={a.id} className="flex items-center gap-3 py-2.5">
              <CourseDot hue={courses.hues.get(a.courseId)} />
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium">{a.name}</p>
                <p className="line-clamp-2 text-xs text-muted-foreground">
                  {a.courseTitle} · {formatDate(a.dueDate, { day: 'numeric', month: 'short' })}
                </p>
              </div>
              <span className="tabular shrink-0 text-sm text-muted-foreground">
                {a.score!.value}
                {a.points !== null && <span className="text-faint"> / {formatPoints(a.points)}</span>}
              </span>
              {a.score!.grade ? <GradeBadge grade={a.score!.grade} /> : <span className="w-8" />}
            </li>
          ))}
        </ul>
      )}
    </Panel>
  )
}

function TermAndAttendance({ snap, date }: { snap: Snapshot; date: string }) {
  const term = snap.reportingTerms.find((r) => r.current)
  const counts = { absent: 0, tardy: 0, other: 0 }
  for (const a of snap.attendance) if (a.kind in counts) counts[a.kind as keyof typeof counts]++
  const left = term ? daysBetween(date, term.end) : 0
  return (
    <Panel title="Term & attendance" className="order-5 lg:order-none">
      {term && (
        <div className="mb-4">
          <p className="text-lg font-semibold">{term.title}</p>
          <p className="text-sm text-muted-foreground">
            {formatDate(term.start, { day: 'numeric', month: 'short' })} – {formatDate(term.end, { day: 'numeric', month: 'short' })}
            {' · '}
            {left > 0 ? `${left} days left` : 'ended'}
          </p>
        </div>
      )}
      <Link to="/attendance" className="grid grid-cols-3 gap-2 rounded-lg transition-colors hover:bg-accent/50">
        <Count n={counts.absent} label="Absences" tone={counts.absent ? 'text-alert' : undefined} />
        <Count n={counts.tardy} label="Tardies" tone={counts.tardy ? 'text-warn' : undefined} />
        <Count n={counts.other} label="Other" />
      </Link>
    </Panel>
  )
}

function Count({ n, label, tone }: { n: number; label: string; tone?: string }) {
  return (
    <div className="glass-sunk rounded-lg px-3 py-2.5">
      <p className={cn('tabular text-2xl font-semibold', tone)}>{n}</p>
      <p className="text-xs text-muted-foreground">{label}</p>
    </div>
  )
}

