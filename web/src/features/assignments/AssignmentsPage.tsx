import { ClipboardCheck, Search } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useSearchParams } from 'react-router'
import { Panel } from '@/components/glass/Panel'
import { EmptyState, ErrorState, LoadingRows } from '@/components/glass/states'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { schoolDate } from '@/lib/dates'
import { Pill } from '@/components/glass/Pill'
import { markTransition } from '@/lib/motion'
import { useCourses, useSnapshot } from '@/lib/queries'
import { useNow } from '@/lib/useNow'
import { cn } from '@/lib/utils'
import { AssignmentRow } from './AssignmentRow'
import { applyFilter, groupByWeek, type Status } from './filter'

const STATUSES: { value: Status; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'graded', label: 'Graded' },
  { value: 'upcoming', label: 'Upcoming' },
  { value: 'flagged', label: 'Late/missing' },
]

const ALL = 'all'

export function AssignmentsPage() {
  const { data: snap, error, refetch, isPending } = useSnapshot()
  const courses = useCourses(snap)
  const [params, setParams] = useSearchParams()
  const [query, setQuery] = useState('')
  const now = useNow(60_000)
  const courseParam = params.get('course')
  const status = (params.get('status') as Status | null) ?? 'all'
  const category = params.get('category')

  const set = (key: string, value: string | null) => {
    markTransition('local')
    setParams((p) => {
      if (value === null || value === ALL) p.delete(key)
      else p.set(key, value)
      return p
    }, { replace: true, viewTransition: true })
  }

  const categories = useMemo(() => [...new Set((snap?.assignments ?? []).map((a) => a.category).filter(Boolean))].sort(), [snap])
  const withWork = useMemo(() => {
    const ids = new Set((snap?.assignments ?? []).map((a) => a.courseId))
    return (snap?.courses ?? []).filter((c) => ids.has(c.id))
  }, [snap])

  if (isPending) return <Panel><LoadingRows rows={8} /></Panel>
  if (!snap) return <ErrorState error={error!} onRetry={() => void refetch()} />

  const today = schoolDate(now, snap.timeZone)
  const list = applyFilter(snap.assignments, { courseId: courseParam ? Number(courseParam) : null, category, status, query }, today)
  const groups = groupByWeek(list, today)

  return (
    <div className="space-y-3 sm:space-y-4">
      <div className="flex flex-col gap-2 lg:flex-row lg:items-center">
        <div className="glass grid grid-cols-4 gap-1 rounded-full p-1 lg:flex" role="group" aria-label="Status">
          {STATUSES.map((s) => (
            <button
              key={s.value}
              onClick={() => set('status', s.value)}
              aria-pressed={status === s.value}
              className={cn(
                'relative h-11 min-w-0 rounded-full px-2 sm:pointer-fine:h-9 text-[13px] font-medium whitespace-nowrap transition duration-200 active:scale-[0.96] sm:px-4 sm:text-sm',
                status === s.value ? 'text-primary-foreground' : 'text-muted-foreground hover:text-foreground',
              )}
            >
              {status === s.value && <Pill name="status-pill" />}
              <span className="relative" style={{ viewTransitionName: `status-${s.value}` }}>{s.label}</span>
            </button>
          ))}
        </div>
        <div className="grid grid-cols-2 gap-2 lg:ml-auto lg:flex">
          <Select value={courseParam ?? ALL} onValueChange={(v) => set('course', v)}>
            <SelectTrigger className="glass w-full rounded-full data-[size=default]:h-11 sm:pointer-fine:data-[size=default]:h-10 lg:w-56" aria-label="Course">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL}>All courses</SelectItem>
              {withWork.map((c) => (
                <SelectItem key={c.id} value={String(c.id)}>{c.title}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select value={category ?? ALL} onValueChange={(v) => set('category', v)}>
            <SelectTrigger className="glass w-full rounded-full data-[size=default]:h-11 sm:pointer-fine:data-[size=default]:h-10 lg:w-44" aria-label="Category">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL}>All categories</SelectItem>
              {categories.map((c) => (
                <SelectItem key={c} value={c}>{c}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <label className="glass col-span-2 flex h-11 items-center sm:pointer-fine:h-10 gap-2 rounded-full px-3 lg:w-56">
            <Search className="size-4 text-faint" aria-hidden />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search assignments"
              aria-label="Search assignments"
              className="h-8 border-0 bg-transparent px-0 shadow-none focus-visible:ring-0 dark:bg-transparent"
            />
          </label>
        </div>
      </div>

      {groups.length === 0 ? (
        <Panel>
          <EmptyState icon={ClipboardCheck} title="No assignments match">
            {snap.assignments.length === 0 ? 'Teachers haven’t published any assignments yet.' : 'Try another status, course or category.'}
          </EmptyState>
        </Panel>
      ) : (
        groups.map((g) => (
          <Panel key={g.label} title={g.label} action={<span className="text-xs text-muted-foreground tabular">{g.items.length}</span>}>
            <ul className="-mx-2">
              {g.items.map((a) => (
                <AssignmentRow key={a.id} a={a} hue={courses.hues.get(a.courseId)} />
              ))}
            </ul>
          </Panel>
        ))
      )}
    </div>
  )
}
