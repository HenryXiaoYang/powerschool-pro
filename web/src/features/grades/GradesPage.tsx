import { GraduationCap } from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router'
import { CourseDot, GradeBadge } from '@/components/glass/badges'
import { Panel } from '@/components/glass/Panel'
import { EmptyState, ErrorState, LoadingRows } from '@/components/glass/states'
import { Pill } from '@/components/glass/Pill'
import { formatPoints } from '@/lib/grades'
import { transition } from '@/lib/motion'
import { useCourses, useSnapshot } from '@/lib/queries'
import { cn } from '@/lib/utils'
import { gradedTerms, gradeTable } from './termGrades'

export function GradesPage() {
  const { data: snap, error, refetch, isPending } = useSnapshot()
  const courses = useCourses(snap)
  const [picked, setPicked] = useState<number | null>(null)

  if (isPending) return <Panel><LoadingRows rows={8} /></Panel>
  if (!snap) return <ErrorState error={error!} onRetry={() => void refetch()} />

  const terms = gradedTerms(snap)
  const current = snap.reportingTerms.find((t) => t.current)
  if (terms.length === 0) {
    return (
      <Panel>
        <EmptyState icon={GraduationCap} title={`No term grades posted${current ? ` for ${current.title}` : ''} yet`}>
          PowerSchool hasn’t published course grades. Scores for individual work are in{' '}
          <Link to="/assignments?status=graded" className="text-foreground underline">Assignments</Link>.
        </EmptyState>
      </Panel>
    )
  }

  const table = gradeTable(snap)
  const rows = snap.courses.filter((c) => table.has(c.id))
  const term = terms.find((t) => t.id === picked) ?? terms.find((t) => t.current) ?? terms.at(-1)!
  const percentsHidden = terms.every((t) => t.suppressPercents)

  return (
    <div className="space-y-3 sm:space-y-4">
      {/* Phone: one term at a time. */}
      <div className="glass flex gap-1 overflow-x-auto rounded-full p-1 md:hidden" role="group" aria-label="Reporting term">
        {terms.map((t) => (
          <button
            key={t.id}
            onClick={() => transition('local', () => setPicked(t.id))}
            aria-pressed={t.id === term.id}
            className={cn('relative h-11 shrink-0 rounded-full px-4 text-sm font-medium transition duration-200 active:scale-[0.96]', t.id === term.id ? 'text-primary-foreground' : 'text-muted-foreground')}
          >
            {t.id === term.id && <Pill name="term-pill" />}
            <span className="relative" style={{ viewTransitionName: `term-${t.id}` }}>{t.title}</span>
          </button>
        ))}
      </div>
      <Panel className="md:hidden">
        <ul className="divide-y divide-border">
          {rows.map((c) => {
            const g = table.get(c.id)?.get(term.id)
            return (
              <li key={c.id} className="flex items-center gap-3 py-2.5">
                <CourseDot hue={courses.hues.get(c.id)} />
                <Link to={`/classes/${c.id}`} className="line-clamp-2 min-w-0 flex-1 font-medium">{c.title}</Link>
                {g ? <GradeBadge grade={g.grade} /> : <span className="text-faint">—</span>}
              </li>
            )
          })}
        </ul>
      </Panel>

      {/* iPad and desktop: the whole year as a grid. */}
      <Panel className="hidden overflow-x-auto p-0 md:block">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left text-muted-foreground">
              <th className="px-5 py-3 font-medium">Course</th>
              {terms.map((t) => (
                <th key={t.id} className={cn('px-3 py-3 text-center font-medium', t.current && 'text-today')}>
                  {t.title}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((c) => (
              <tr key={c.id} className="border-b border-border/60 last:border-0">
                <td className="px-5 py-2.5">
                  <Link to={`/classes/${c.id}`} className="flex items-center gap-3 font-medium hover:underline">
                    <CourseDot hue={courses.hues.get(c.id)} /> {c.title}
                  </Link>
                </td>
                {terms.map((t) => {
                  const g = table.get(c.id)?.get(t.id)
                  return (
                    <td key={t.id} className="px-3 py-2.5 text-center">
                      {g ? (
                        <span className="inline-flex flex-col items-center gap-0.5">
                          <GradeBadge grade={g.grade} />
                          {g.percent !== null && <span className="tabular text-[11px] text-muted-foreground">{formatPoints(g.percent)}%</span>}
                        </span>
                      ) : (
                        <span className="text-faint">—</span>
                      )}
                    </td>
                  )
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </Panel>
      {percentsHidden && <p className="px-1 text-xs text-muted-foreground">Your school hides term percentages; only grades are shown.</p>}
    </div>
  )
}
