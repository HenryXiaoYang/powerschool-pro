import { CalendarOff, ChevronLeft, ChevronRight } from 'lucide-react'
import { useState, type CSSProperties } from 'react'
import { Link, useSearchParams } from 'react-router'
import type { Snapshot } from '@/api/student.gen'
import { CourseDot } from '@/components/glass/badges'
import { Panel } from '@/components/glass/Panel'
import { EmptyState, ErrorState, LoadingRows } from '@/components/glass/states'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { addDays, clockTime, formatDate, minuteOfDay, mondayOf, schoolDate, weekday } from '@/lib/dates'
import { hueColor } from '@/lib/grades'
import { Pill } from '@/components/glass/Pill'
import { markTransition } from '@/lib/motion'
import { useCourses, useSchedule, useSnapshot } from '@/lib/queries'
import { useNow } from '@/lib/useNow'
import { cn } from '@/lib/utils'
import { gridHours, placeDay, weekPlan, type DayPlan } from './week'

type Courses = ReturnType<typeof useCourses>
const PX_PER_MIN = 1.1

export function SchedulePage() {
  const { data: snap, error, refetch, isPending } = useSnapshot()
  if (isPending) return <Panel><LoadingRows rows={6} /></Panel>
  if (!snap) return <ErrorState error={error!} onRetry={() => void refetch()} />
  return <Week snap={snap} />
}

function Week({ snap }: { snap: Snapshot }) {
  const now = useNow()
  const todayDate = schoolDate(now, snap.timeZone)
  const [params, setParams] = useSearchParams()
  // On a weekend the coming week is the useful one, as on Home.
  const home = weekday(todayDate) >= 5 ? addDays(mondayOf(todayDate), 7) : todayDate
  const selected = params.get('day') ?? home
  const monday = mondayOf(selected)
  const [classesOnly, setClassesOnly] = useState(false)
  const courses = useCourses(snap)

  // A week can span two months; each month is one schedule request.
  const sunday = addDays(monday, 6)
  const first = useSchedule(snap, Number(monday.slice(0, 4)), Number(monday.slice(5, 7)))
  const second = useSchedule(snap, Number(sunday.slice(0, 4)), Number(sunday.slice(5, 7)))
  const sameMonth = monday.slice(0, 7) === sunday.slice(0, 7)
  const schedules = (sameMonth ? [first.data] : [first.data, second.data]).filter((s) => s !== undefined)
  const loading = first.isPending || second.isPending
  const failed = first.error ?? second.error

  const plans = weekPlan(monday, schedules, snap.holidays).map((p) => ({
    ...p,
    meetings: classesOnly ? p.meetings.filter((m) => courses.byId.get(m.courseId)?.academic) : p.meetings,
  }))
  const go = (date: string) => {
    markTransition(mondayOf(date) > monday ? 'next' : mondayOf(date) < monday ? 'prev' : 'local')
    setParams({ day: date }, { replace: true, viewTransition: true })
  }
  const weekLetter = plans.find((p) => p.week)?.week

  return (
    <div className="space-y-3 sm:space-y-4">
      <div className="flex flex-wrap items-center gap-2">
        <div className="glass flex items-center rounded-full p-1">
          <Button variant="ghost" size="icon" className="size-11 rounded-full md:pointer-fine:size-9" aria-label="Previous week" onClick={() => go(addDays(monday, -7))}>
            <ChevronLeft />
          </Button>
          <Button variant="ghost" className="h-11 rounded-full px-3 md:pointer-fine:h-9" onClick={() => go(home)} disabled={mondayOf(home) === monday && selected === home}>
            Today
          </Button>
          <Button variant="ghost" size="icon" className="size-11 rounded-full md:pointer-fine:size-9" aria-label="Next week" onClick={() => go(addDays(monday, 7))}>
            <ChevronRight />
          </Button>
        </div>
        <p className="font-medium">
          {formatDate(monday, { day: 'numeric', month: 'short' })} – {formatDate(addDays(monday, 4), { day: 'numeric', month: 'short' })}
          {weekLetter && <span className="text-muted-foreground"> · Week {weekLetter}</span>}
        </p>
        <div className="ml-auto flex items-center gap-2">
          <Checkbox id="classes-only" checked={classesOnly} onCheckedChange={(v) => setClassesOnly(v === true)} />
          <Label htmlFor="classes-only" className="py-3 text-sm font-normal">Classes only</Label>
        </div>
      </div>

      {failed && !loading && schedules.length === 0 ? (
        <Panel><ErrorState error={failed} onRetry={() => { first.refetch(); second.refetch() }} /></Panel>
      ) : loading && schedules.length === 0 ? (
        <Panel><LoadingRows rows={6} /></Panel>
      ) : (
        <div style={{ viewTransitionName: 'week' }}>
          <WeekGrid plans={plans} snap={snap} courses={courses} now={now} today={todayDate} />
          <DayList plans={plans} snap={snap} courses={courses} now={now} today={todayDate} selected={selected} onSelect={go} />
        </div>
      )}
    </div>
  )
}

interface ViewProps {
  plans: DayPlan[]
  snap: Snapshot
  courses: Courses
  now: Date
  today: string
}

function WeekGrid({ plans, snap, courses, now, today }: ViewProps) {
  const tz = snap.timeZone
  const { start, end } = gridHours(plans, tz)
  const height = (end - start) * PX_PER_MIN
  const hours = Array.from({ length: (end - start) / 60 + 1 }, (_, i) => start + i * 60)
  const nowMin = minuteOfDay(now.toISOString(), tz)

  return (
    <section className="glass hidden overflow-hidden rounded-xl lg:block" aria-label="Week timetable">
      <div className="grid border-b border-border" style={{ gridTemplateColumns: `3.25rem repeat(${plans.length}, minmax(0, 1fr))` }}>
        <span />
        {plans.map((p) => (
          <div key={p.date} className="flex items-baseline justify-center gap-2 px-2 py-3">
            <span className={cn('font-medium', p.date === today && 'text-today')}>{formatDate(p.date, { weekday: 'short', day: 'numeric' })}</span>
            <span className="text-xs text-muted-foreground">{p.holiday ? 'No school' : p.code}</span>
          </div>
        ))}
      </div>
      <div className="grid py-3" style={{ gridTemplateColumns: `3.25rem repeat(${plans.length}, minmax(0, 1fr))` }}>
        <div className="relative" style={{ height }}>
          {plans.some((p) => p.date === today) && nowMin >= start && nowMin <= end && (
            <span aria-hidden className="absolute right-0 size-2 translate-x-1/2 -translate-y-1/2 rounded-full bg-now" style={{ top: (nowMin - start) * PX_PER_MIN }} />
          )}
          {hours.map((h) => (
            <span key={h} className="tabular absolute right-2 -translate-y-1/2 text-[11px] text-faint" style={{ top: (h - start) * PX_PER_MIN }}>
              {String(h / 60).padStart(2, '0')}:00
            </span>
          ))}
        </div>
        {plans.map((p) => (
          <div
            key={p.date}
            role="group"
            aria-label={formatDate(p.date, { weekday: 'long', day: 'numeric', month: 'long' })}
            className={cn('relative border-l border-border', p.date === today && 'bg-accent/40')}
            style={{ height }}
          >
            {hours.map((h) => (
              <span key={h} className="absolute inset-x-0 border-t border-border/60" style={{ top: (h - start) * PX_PER_MIN }} />
            ))}
            {p.holiday && <p className="absolute inset-x-0 top-6 text-center text-sm text-muted-foreground">{p.holiday.description || 'Holiday'}</p>}
            {placeDay(p.meetings, tz, start).map(({ meeting: m, top, height: h, lane, lanes }) => {
              const c = courses.byId.get(m.courseId)
              const hue = courses.hues.get(m.courseId)
              const live = p.date === today && Date.parse(m.start) <= now.getTime() && now.getTime() < Date.parse(m.end)
              return (
                <Link
                  key={`${m.courseId}-${m.start}-${m.end}`}
                  to={`/classes/${m.courseId}`}
                  className={cn('absolute z-[1] block overflow-hidden rounded-md px-2 text-xs leading-snug focus-visible:z-[2]', h < 30 ? 'py-0.5 text-[11px] leading-tight' : 'py-1', live ? 'lit border' : 'glass-sunk')}
                  style={
                    {
                      '--hue': hue ?? 220,
                      top: top * PX_PER_MIN + 1,
                      height: h * PX_PER_MIN - 2,
                      left: `calc(${(lane / lanes) * 100}% + 3px)`,
                      width: `calc(${100 / lanes}% - 6px)`,
                      background: hue === undefined || live ? undefined : `color-mix(in oklch, ${hueColor(hue)} 16%, transparent)`,
                    } as CSSProperties
                  }
                  title={`${c?.title ?? ''} ${clockTime(m.start, tz)}–${clockTime(m.end, tz)}${c?.room ? ` · ${c.room}` : ''}`}
                >
                  <p className={cn('font-medium', h < 60 ? 'truncate' : 'line-clamp-2', !c?.academic && 'text-muted-foreground')}>{c?.title}</p>
                  {h >= 40 && (
                    <p className="tabular text-muted-foreground">
                      {clockTime(m.start, tz)}
                      {c?.room && ` · ${c.room}`}
                    </p>
                  )}
                </Link>
              )
            })}
            {/* While a class is live its lit tile marks now; the line only crosses free time. */}
            {p.date === today && nowMin >= start && nowMin <= end && !p.meetings.some((m) => Date.parse(m.start) <= now.getTime() && now.getTime() < Date.parse(m.end)) && (
              <div className="now-sweep pointer-events-none absolute inset-x-0 z-0 h-0.5 bg-now" style={{ top: (nowMin - start) * PX_PER_MIN }} aria-hidden />
            )}
          </div>
        ))}
      </div>
    </section>
  )
}

function DayList({ plans, snap, courses, now, today, selected, onSelect }: ViewProps & { selected: string; onSelect: (d: string) => void }) {
  const tz = snap.timeZone
  const day = plans.find((p) => p.date === selected) ?? plans[0]
  const t = now.getTime()
  return (
    <div className="space-y-3 lg:hidden">
      <div className="grid gap-1.5" style={{ gridTemplateColumns: `repeat(${plans.length}, minmax(0, 1fr))` }}>
        {plans.map((p) => (
          <button
            key={p.date}
            onClick={() => onSelect(p.date)}
            aria-pressed={p.date === day.date}
            className={cn(
              'glass relative flex flex-col items-center rounded-lg py-2 transition duration-200 active:scale-[0.95]',
              p.date === day.date && 'text-primary-foreground',
            )}
          >
            {p.date === day.date && <Pill name="day-pill" className="rounded-lg" />}
            <span className="relative flex flex-col items-center" style={{ viewTransitionName: `day-${p.date}` }}>
            <span className="text-[11px]">{formatDate(p.date, { weekday: 'short' })}</span>
            <span className={cn('tabular text-base font-semibold', p.date === today && p.date !== day.date && 'text-today')}>{formatDate(p.date, { day: 'numeric' })}</span>
            <span className="text-[11px] opacity-80">{p.holiday ? 'Off' : p.code || '·'}</span>
            </span>
          </button>
        ))}
      </div>
      <Panel>
        {day.meetings.length === 0 ? (
          <EmptyState icon={CalendarOff} title={day.holiday ? (day.holiday.description || 'No school') : 'No classes'} />
        ) : (
          <ol className="divide-y divide-border">
            {day.meetings.map((m) => {
              const c = courses.byId.get(m.courseId)
              const hue = courses.hues.get(m.courseId)
              const live = Date.parse(m.start) <= t && t < Date.parse(m.end)
              return (
                <li key={`${m.courseId}-${m.start}-${m.end}`}>
                  <Link
                    to={`/classes/${m.courseId}`}
                    className={cn('flex items-center gap-3 py-2.5', live && 'lit glass-sunk -mx-2 rounded-md px-2', Date.parse(m.end) <= t && 'opacity-50')}
                    style={{ '--hue': hue ?? 220 } as CSSProperties}
                  >
                    <span className="tabular w-11 shrink-0 text-sm leading-tight">
                      <span className="block text-muted-foreground">{clockTime(m.start, tz)}</span>
                      <span className="block text-xs text-faint">{clockTime(m.end, tz)}</span>
                    </span>
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
    </div>
  )
}
