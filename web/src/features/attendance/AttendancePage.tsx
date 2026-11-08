import { Check, UserCheck } from 'lucide-react'
import type { Attendance, AttendanceKind, Snapshot } from '@/api/student.gen'
import { CourseDot } from '@/components/glass/badges'
import { Panel } from '@/components/glass/Panel'
import { EmptyState, ErrorState, LoadingRows } from '@/components/glass/states'
import { formatDate } from '@/lib/dates'
import { useCourses, useSnapshot } from '@/lib/queries'
import { cn } from '@/lib/utils'

const KIND_TONE: Record<AttendanceKind, string> = {
  absent: 'text-alert border-alert/40',
  tardy: 'text-warn border-warn/40',
  present: 'text-muted-foreground border-border',
  other: 'text-foreground border-foreground/25',
}

function CodeTag({ code, kind }: { code: string; kind: AttendanceKind }) {
  return <span className={cn('tabular inline-flex h-6 shrink-0 items-center rounded-md border px-1.5 text-xs font-semibold', KIND_TONE[kind])}>{code || <Check className="size-3.5" aria-label="Present" />}</span>
}

export function AttendancePage() {
  const { data: snap, error, refetch, isPending } = useSnapshot()
  if (isPending) return <Panel><LoadingRows rows={6} /></Panel>
  if (!snap) return <ErrorState error={error!} onRetry={() => void refetch()} />

  const byCode = new Map<string, { kind: AttendanceKind; description: string; n: number }>()
  for (const a of snap.attendance) {
    const key = a.code || a.description
    const entry = byCode.get(key) ?? { kind: a.kind, description: a.description, n: 0 }
    entry.n++
    byCode.set(key, entry)
  }
  const totals = { absent: 0, tardy: 0 }
  for (const a of snap.attendance) if (a.kind === 'absent' || a.kind === 'tardy') totals[a.kind]++

  return (
    <div className="space-y-3 sm:space-y-4">
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-[repeat(2,minmax(0,10rem))_1fr] sm:gap-4">
        <Total n={totals.absent} label="Absences" tone={totals.absent ? 'text-alert' : undefined} />
        <Total n={totals.tardy} label="Tardies" tone={totals.tardy ? 'text-warn' : undefined} />
        <Panel className="col-span-2 sm:col-span-1" title="By code">
          {byCode.size === 0 ? (
            <p className="text-sm text-muted-foreground">Nothing recorded this year.</p>
          ) : (
            <ul className="flex flex-wrap gap-2">
              {[...byCode].map(([code, e]) => (
                <li key={code} className="glass-sunk flex items-center gap-2 rounded-lg py-1 pr-3 pl-1">
                  <CodeTag code={code} kind={e.kind} />
                  <span className="text-sm">{e.description}</span>
                  <span className="tabular text-sm font-semibold">{e.n}</span>
                </li>
              ))}
            </ul>
          )}
        </Panel>
      </div>

      <Panel title="All records">
        <p className="-mt-2 mb-3 text-xs text-muted-foreground">PowerSchool records exceptions only. Days not listed were ordinary attendance.</p>
        {snap.attendance.length === 0 ? (
          <EmptyState icon={UserCheck} title="No attendance records">PowerSchool lists only absences, tardies and other exceptions.</EmptyState>
        ) : (
          <AttendanceList records={snap.attendance} snap={snap} />
        )}
      </Panel>
    </div>
  )
}

function Total({ n, label, tone }: { n: number; label: string; tone?: string }) {
  return (
    <div className="glass rounded-xl px-5 py-4">
      <p className={cn('tabular text-3xl font-semibold', tone)}>{n}</p>
      <p className="text-sm text-muted-foreground">{label} this year</p>
    </div>
  )
}

/** Records grouped by month. Shared with the course page. */
export function AttendanceList({ records, snap, compact = false }: { records: Attendance[]; snap: Snapshot; compact?: boolean }) {
  const courses = useCourses(snap)
  const months = new Map<string, Attendance[]>()
  for (const r of records) months.set(r.date.slice(0, 7), [...(months.get(r.date.slice(0, 7)) ?? []), r])

  return (
    <div className="space-y-4">
      {[...months].map(([month, list]) => (
        <section key={month}>
          {!compact && <h3 className="mb-1 text-xs font-medium text-muted-foreground">{formatDate(`${month}-01`, { month: 'long', year: 'numeric' })}</h3>}
          <ul className="divide-y divide-border">
            {list.map((r) => {
              const c = courses.byId.get(r.courseId)
              return (
                <li key={r.id} className="flex items-start gap-3 py-2.5">
                  <span className="tabular w-14 shrink-0 pt-0.5 text-sm text-muted-foreground">{formatDate(r.date, { day: 'numeric', month: 'short' })}</span>
                  <CodeTag code={r.code} kind={r.kind} />
                  <div className="min-w-0 flex-1 wrap-anywhere">
                    <p className="text-sm font-medium">{r.description}</p>
                    {!compact && c && (
                      <p className="flex items-start gap-1.5 text-xs text-muted-foreground">
                        <CourseDot hue={courses.hues.get(c.id)} className="mt-1 size-2" />
                        <span className="min-w-0">
                          {c.title}
                          {r.period && ` · Period ${r.period}`}
                        </span>
                      </p>
                    )}
                    {r.comment && <p className="mt-0.5 text-xs text-muted-foreground">“{r.comment}”</p>}
                  </div>
                </li>
              )
            })}
          </ul>
        </section>
      ))}
    </div>
  )
}
