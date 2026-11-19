import { CalendarHeart, ExternalLink, Mail, Megaphone, Phone, School, UsersRound } from 'lucide-react'
import { useMemo } from 'react'
import { Link } from 'react-router'
import type { Snapshot } from '@/api/student.gen'
import { Panel } from '@/components/glass/Panel'
import { EmptyState, ErrorState, LoadingRows } from '@/components/glass/states'
import { formatDate } from '@/lib/dates'
import { useSnapshot } from '@/lib/queries'

export function SchoolPage() {
  const { data: snap, error, refetch, isPending } = useSnapshot()
  if (isPending) return <Panel><LoadingRows rows={4} /></Panel>
  if (!snap) return <ErrorState error={error!} onRetry={() => void refetch()} />
  return (
    <div className="mx-auto grid max-w-3xl gap-3 sm:gap-4">
      <Conferences snap={snap} />
      <Announcements snap={snap} />
      <OtherLinks snap={snap} />
      <Teachers snap={snap} />
    </div>
  )
}

/** Opens in a new tab: the school's own pages are not ours to frame or proxy. */
function External({ href, children, className }: { href: string; children: React.ReactNode; className?: string }) {
  return (
    <a href={href} target="_blank" rel="noopener noreferrer" className={className}>
      {children}
    </a>
  )
}

const row = 'flex min-h-12 items-center gap-3 rounded-lg px-2 py-2 transition-colors hover:bg-accent/50'

function Conferences({ snap }: { snap: Snapshot }) {
  const found = snap.links.filter((l) => l.conference)
  return (
    <Panel title="Parent-teacher conferences">
      {found.length > 0 ? (
        <ul>
          {found.map((l) => (
            <li key={l.url}>
              <External href={l.url} className={row}>
                <CalendarHeart className="size-5 shrink-0 text-now" strokeWidth={1.8} />
                <span className="min-w-0 flex-1 font-medium">{l.title}</span>
                <ExternalLink className="size-4 shrink-0 text-faint" />
              </External>
            </li>
          ))}
        </ul>
      ) : (
        <EmptyState icon={CalendarHeart} title="No conference link yet">
          {snap.parentLinked
            ? 'Your school hasn’t put a conference booking link in the PowerSchool app.'
            : 'Schools usually share the booking link with parent accounts. Add a parent login in Settings to look for it.'}
          {snap.portalUrl && (
            <>
              {' '}
              <External href={snap.portalUrl} className="text-foreground underline">Open the school portal</External>
            </>
          )}
          {!snap.parentLinked && (
            <>
              {' '}
              <Link to="/settings" className="text-foreground underline">Settings</Link>
            </>
          )}
        </EmptyState>
      )}
    </Panel>
  )
}

function Announcements({ snap }: { snap: Snapshot }) {
  return (
    <Panel title="Announcements">
      {snap.announcements.length === 0 ? (
        <EmptyState icon={Megaphone} title="Nothing posted">Current school announcements appear here.</EmptyState>
      ) : (
        <ul className="divide-y divide-border">
          {snap.announcements.map((a) => (
            <li key={a.id} className="py-3 first:pt-0 last:pb-0">
              <p className="font-medium">{a.title || 'Announcement'}</p>
              {a.start && <p className="text-xs text-muted-foreground">{formatDate(a.start, { day: 'numeric', month: 'short', year: 'numeric' })}</p>}
              {/* Text only: the school's markup is stripped on the server and never rendered as HTML. */}
              {a.body && <p className="mt-1 text-sm whitespace-pre-line text-muted-foreground">{a.body}</p>}
            </li>
          ))}
        </ul>
      )}
    </Panel>
  )
}

function OtherLinks({ snap }: { snap: Snapshot }) {
  const links = snap.links.filter((l) => !l.conference)
  if (links.length === 0) return null
  return (
    <Panel title="School links">
      <ul>
        {links.map((l) => (
          <li key={l.url}>
            <External href={l.url} className={row}>
              <School className="size-5 shrink-0 text-muted-foreground" strokeWidth={1.8} />
              <span className="min-w-0 flex-1 font-medium">{l.title}</span>
              <ExternalLink className="size-4 shrink-0 text-faint" />
            </External>
          </li>
        ))}
      </ul>
    </Panel>
  )
}

interface Contact {
  name: string
  email: string
  phone: string
  courses: string[]
}

/** One entry per teacher, with every course they teach this student. */
function contacts(snap: Snapshot): Contact[] {
  const by = new Map<string, Contact>()
  for (const c of snap.courses) {
    const t = c.teacher
    if (!t?.name) continue
    const key = t.email || t.name
    const entry = by.get(key) ?? { name: t.name, email: t.email, phone: t.phone, courses: [] }
    if (c.academic) entry.courses.push(c.title)
    by.set(key, entry)
  }
  return [...by.values()].filter((c) => c.courses.length > 0).sort((a, b) => a.name.localeCompare(b.name))
}

function Teachers({ snap }: { snap: Snapshot }) {
  const list = useMemo(() => contacts(snap), [snap])
  return (
    <Panel title="Teachers">
      {list.length === 0 ? (
        <EmptyState icon={UsersRound} title="No teacher contacts">PowerSchool lists no teachers for this student.</EmptyState>
      ) : (
        <ul className="divide-y divide-border">
          {list.map((t) => (
            <li key={t.name} className="flex items-center gap-3 py-2.5 first:pt-0 last:pb-0">
              <div className="min-w-0 flex-1">
                <p className="font-medium">{t.name}</p>
                <p className="line-clamp-2 text-xs text-muted-foreground">{t.courses.join(' · ')}</p>
              </div>
              {t.email && (
                <a href={`mailto:${t.email}`} aria-label={`Email ${t.name}`} className="grid size-11 shrink-0 place-items-center rounded-full text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
                  <Mail className="size-5" strokeWidth={1.8} />
                </a>
              )}
              {t.phone && (
                <a href={`tel:${t.phone.replace(/[^\d+]/g, '')}`} aria-label={`Call ${t.name}`} className="grid size-11 shrink-0 place-items-center rounded-full text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
                  <Phone className="size-5" strokeWidth={1.8} />
                </a>
              )}
            </li>
          ))}
        </ul>
      )}
    </Panel>
  )
}
