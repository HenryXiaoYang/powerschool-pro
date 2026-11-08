import { BookOpen, ChevronRight } from 'lucide-react'
import { Link, useSearchParams } from 'react-router'
import { CourseDot, GradeBadge } from '@/components/glass/badges'
import { Panel } from '@/components/glass/Panel'
import { EmptyState, ErrorState, LoadingRows } from '@/components/glass/states'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { gradeTable, latestGrade } from '@/features/grades/termGrades'
import { useCourses, useSnapshot } from '@/lib/queries'
import { cn } from '@/lib/utils'

export function ClassesPage() {
  const { data: snap, error, refetch, isPending } = useSnapshot()
  const courses = useCourses(snap)
  const [params, setParams] = useSearchParams()
  const showAll = params.get('all') === '1'

  if (isPending) return <Panel><LoadingRows rows={8} /></Panel>
  if (!snap) return <ErrorState error={error!} onRetry={() => void refetch()} />

  const table = gradeTable(snap)
  const scored = new Map<number, number>()
  for (const a of snap.assignments) if (a.score) scored.set(a.courseId, (scored.get(a.courseId) ?? 0) + 1)
  const list = snap.courses.filter((c) => showAll || c.academic)

  return (
    <div className="space-y-3 sm:space-y-4">
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          {list.length} {showAll ? 'blocks' : 'courses'} this year
        </p>
        <div className="flex items-center gap-2">
          <Checkbox id="show-all" checked={showAll} onCheckedChange={(v) => setParams(v === true ? { all: '1' } : {}, { replace: true })} />
          <Label htmlFor="show-all" className="py-3 text-sm font-normal">Show all blocks</Label>
        </div>
      </div>
      <Panel className="p-2 sm:p-3">
        {list.length === 0 ? (
          <EmptyState icon={BookOpen} title="No courses" />
        ) : (
          <ul className="grid grid-cols-1 gap-x-3 lg:grid-cols-2">
            {list.map((c) => {
              const latest = latestGrade(snap, table, c.id)
              return (
                <li key={c.id}>
                  <Link to={`/classes/${c.id}`} className="flex items-center gap-3 rounded-lg px-3 py-3 transition-colors hover:bg-accent/60">
                    <CourseDot hue={courses.hues.get(c.id)} className="size-3" />
                    <div className="min-w-0 flex-1">
                      <p className={cn('line-clamp-2 font-medium', !c.academic && 'text-muted-foreground')}>{c.title}</p>
                      <p className="truncate text-xs text-muted-foreground">
                        {[c.teacher?.name, c.room && `Room ${c.room}`].filter(Boolean).join(' · ') || c.expression}
                      </p>
                    </div>
                    <div className="hidden shrink-0 text-right text-xs text-muted-foreground sm:block">
                      {(c.absences > 0 || c.tardies > 0) && (
                        <p>
                          {c.absences > 0 && <span className="text-alert">{c.absences} absent</span>}
                          {c.absences > 0 && c.tardies > 0 && ' · '}
                          {c.tardies > 0 && <span className="text-warn">{c.tardies} tardy</span>}
                        </p>
                      )}
                      {scored.get(c.id) ? <p>{scored.get(c.id)} scored</p> : null}
                    </div>
                    {latest ? (
                      <span className="flex shrink-0 items-center gap-1.5">
                        <span className="text-[11px] text-faint">{latest.term.abbrev}</span>
                        <GradeBadge grade={latest.grade.grade} />
                      </span>
                    ) : (
                      <span className="w-8 shrink-0" />
                    )}
                    <ChevronRight className="size-4 shrink-0 text-faint" />
                  </Link>
                </li>
              )
            })}
          </ul>
        )}
      </Panel>
    </div>
  )
}
