import { ChevronRight, FileText } from 'lucide-react'
import { Link } from 'react-router'
import { Panel } from '@/components/glass/Panel'
import { EmptyState, ErrorState, LoadingRows } from '@/components/glass/states'
import { useSession } from '@/features/auth/session'
import { formatDate } from '@/lib/dates'
import { useReportCards, useSnapshot } from '@/lib/queries'

export function ReportPage() {
  const { data: snap, error, refetch, isPending } = useSnapshot()
  const { reportCards: portal } = useSession()
  const cards = useReportCards(portal)

  if (isPending) return <Panel><LoadingRows rows={4} /></Panel>
  if (!snap) return <ErrorState error={error!} onRetry={() => void refetch()} />

  const p = snap.profile
  const name = [p.firstName, p.middleName, p.lastName].filter(Boolean).join(' ')
  const initials = [p.firstName, p.lastName].map((n) => n.charAt(0)).join('')
  const term = snap.reportingTerms.find((t) => t.current)
  const academic = snap.courses.filter((c) => c.academic).length
  const absences = snap.attendance.filter((a) => a.kind === 'absent').length

  return (
    <div className="grid grid-cols-1 gap-3 sm:gap-4 lg:grid-cols-12">
      <Panel className="lg:col-span-5">
        <div className="flex items-center gap-4">
          <span className="grid size-16 shrink-0 place-items-center bead rounded-full bg-primary text-xl font-semibold text-primary-foreground" aria-hidden>
            {initials}
          </span>
          <div className="min-w-0">
            <h2 className="truncate text-xl font-semibold tracking-tight">{name}</h2>
            <p className="text-sm text-muted-foreground">
              Grade {p.gradeLevel}
              {p.school && ` · ${p.school}`}
            </p>
          </div>
        </div>
        <dl className="mt-5 grid grid-cols-3 gap-2">
          <Stat label="Current term" value={term?.title ?? (p.currentTerm || '—')} />
          <Stat label="Courses" value={String(academic)} />
          <Stat label="Absences" value={String(absences)} />
        </dl>
        {p.gpa && <p className="mt-3 text-sm text-muted-foreground">GPA {p.gpa}</p>}
      </Panel>

      <Panel title="Report cards" className="lg:col-span-7">
        {!portal ? (
          <EmptyState icon={FileText} title="Report cards unavailable">
            The PowerSchool web portal didn’t accept this sign-in. Sign out and in again to retry.
          </EmptyState>
        ) : cards.isPending ? (
          <LoadingRows rows={4} />
        ) : cards.error ? (
          <ErrorState error={cards.error} onRetry={() => void cards.refetch()} />
        ) : cards.data.length === 0 ? (
          <EmptyState icon={FileText} title="No report cards published yet" />
        ) : (
          <ul className="-mx-2">
            {cards.data.map((rc) => (
              <li key={rc.id}>
                <Link to={`/report/${rc.id}`} className="flex items-center gap-3 rounded-md px-2 py-3 transition-colors hover:bg-accent/60">
                  <FileText className="size-5 shrink-0 text-muted-foreground" strokeWidth={1.6} />
                  <span className="min-w-0 flex-1 truncate font-medium">{rc.name}</span>
                  <span className="tabular shrink-0 text-sm text-muted-foreground">{rc.date && formatDate(rc.date, { day: 'numeric', month: 'short', year: 'numeric' })}</span>
                  <ChevronRight className="size-4 shrink-0 text-faint" />
                </Link>
              </li>
            ))}
          </ul>
        )}
      </Panel>
    </div>
  )
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="glass-sunk rounded-lg px-3 py-2.5">
      <dd className="tabular truncate text-lg font-semibold">{value}</dd>
      <dt className="text-xs text-muted-foreground">{label}</dt>
    </div>
  )
}
