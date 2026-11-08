import { ClipboardCheck, Mail, SearchX, UserCheck } from 'lucide-react'
import { Link, useParams } from 'react-router'
import { CourseDot, GradeBadge } from '@/components/glass/badges'
import { Panel } from '@/components/glass/Panel'
import { EmptyState, ErrorState, LoadingRows } from '@/components/glass/states'
import { AssignmentRow } from '@/features/assignments/AssignmentRow'
import { AttendanceList } from '@/features/attendance/AttendancePage'
import { gradeTable } from '@/features/grades/termGrades'
import { useCourses, useSnapshot } from '@/lib/queries'

export function CoursePage() {
  const { id } = useParams()
  const { data: snap, error, refetch, isPending } = useSnapshot()
  const courses = useCourses(snap)

  if (isPending) return <Panel><LoadingRows rows={6} /></Panel>
  if (!snap) return <ErrorState error={error!} onRetry={() => void refetch()} />

  const c = courses.byId.get(Number(id))
  if (!c) {
    return (
      <Panel>
        <EmptyState icon={SearchX} title="Course not found">
          It may belong to another year. <Link to="/classes" className="underline">Back to classes</Link>
        </EmptyState>
      </Panel>
    )
  }
  const hue = courses.hues.get(c.id)
  const work = snap.assignments.filter((a) => a.courseId === c.id)
  const row = gradeTable(snap).get(c.id)
  const attendance = snap.attendance.filter((a) => a.courseId === c.id)
  const term = snap.terms.find((t) => t.id === c.termId)

  return (
    <div className="space-y-3 sm:space-y-4">
      <div>
        <div className="flex items-start gap-3">
          <CourseDot hue={hue} className="mt-2.5 size-3" />
          <h1 className="text-2xl font-semibold tracking-tight text-balance">{c.title}</h1>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-3 sm:gap-4 lg:grid-cols-12">
        <Panel title="Assignments" action={<span className="tabular text-xs text-muted-foreground">{work.length}</span>} className="lg:col-span-8 lg:row-span-3">
          {work.length === 0 ? (
            <EmptyState icon={ClipboardCheck} title="No assignments yet">Work appears here when the teacher publishes it.</EmptyState>
          ) : (
            <ul className="-mx-2">
              {work.map((a) => <AssignmentRow key={a.id} a={a} hue={hue} showCourse={false} />)}
            </ul>
          )}
        </Panel>

        <Panel title="Details" className="lg:col-span-4">
          <dl className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-4 gap-y-2 text-sm wrap-anywhere">
            <dt className="text-muted-foreground">Teacher</dt>
            <dd>
              {c.teacher?.name ?? '—'}
              {c.teacher?.email && (
                <a href={`mailto:${c.teacher.email}`} className="mt-0.5 flex items-center gap-1 text-xs text-muted-foreground underline decoration-border hover:text-foreground hover:decoration-current">
                  <Mail className="size-3.5" /> {c.teacher.email}
                </a>
              )}
            </dd>
            <dt className="text-muted-foreground">Room</dt>
            <dd>{c.room || '—'}</dd>
            <dt className="text-muted-foreground">Course code</dt>
            <dd className="tabular">{c.code || '—'}</dd>
            <dt className="text-muted-foreground">Term</dt>
            <dd>{term?.title ?? '—'}</dd>
            <dt className="text-muted-foreground">Meets</dt>
            <dd className="space-y-0.5">
              {c.slots.length === 0
                ? '—'
                : c.slots.map((s) => (
                    <p key={s.periods + s.days.join()} className="tabular">
                      <span className="font-medium">Period {s.periods}</span> <span className="text-muted-foreground">{s.days.join(' ')}</span>
                    </p>
                  ))}
            </dd>
          </dl>
        </Panel>

        <Panel title="Term grades" className="lg:col-span-4">
          {!row ? (
            <p className="text-sm text-muted-foreground">No term grades posted yet.</p>
          ) : (
            <div className="flex flex-wrap gap-3">
              {snap.reportingTerms
                .filter((t) => row.has(t.id))
                .map((t) => (
                  <div key={t.id} className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">{t.abbrev}</span>
                    <GradeBadge grade={row.get(t.id)!.grade} />
                  </div>
                ))}
            </div>
          )}
        </Panel>

        <Panel title="Attendance" className="lg:col-span-4">
          {attendance.length === 0 ? (
            <EmptyState icon={UserCheck} title="Nothing recorded" className="py-4" />
          ) : (
            <AttendanceList records={attendance} snap={snap} compact />
          )}
        </Panel>
      </div>
    </div>
  )
}
