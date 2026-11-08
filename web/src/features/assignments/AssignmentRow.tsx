import { ChevronDown } from 'lucide-react'
import { useId, useState } from 'react'
import type { Assignment } from '@/api/student.gen'
import { CourseDot, FlagChip, GradeBadge } from '@/components/glass/badges'
import { formatDate } from '@/lib/dates'
import { formatPoints, scoreFlags } from '@/lib/grades'
import { cn } from '@/lib/utils'
import { isGraded } from './filter'

/** One assignment; expands in place to show the teacher's comment and details. */
export function AssignmentRow({ a, hue, showCourse = true }: { a: Assignment; hue: number | undefined; showCourse?: boolean }) {
  const [open, setOpen] = useState(false)
  const id = useId()
  const s = a.score
  const flags = scoreFlags(s, a.counted)
  const graded = isGraded(a)

  return (
    <li>
      <button
        onClick={() => setOpen((o) => !o)}
        aria-expanded={open}
        aria-controls={id}
        className="flex w-full items-center gap-3 rounded-md px-2 py-2.5 text-left transition-colors hover:bg-accent/60"
      >
        <CourseDot hue={hue} />
        <span className="block min-w-0 flex-1">
          <span className="block truncate font-medium">{a.name}</span>
          <span className="line-clamp-2 block text-xs text-muted-foreground">
            {[showCourse && a.courseTitle, a.category, a.dueDate && `Due ${formatDate(a.dueDate, { day: 'numeric', month: 'short' })}`].filter(Boolean).join(' · ')}
          </span>
          {flags.length > 0 && (
            <span className="mt-1.5 flex flex-wrap gap-1">
              {flags.map((f) => <FlagChip key={f} flag={f} />)}
            </span>
          )}
        </span>
        <span className="block tabular shrink-0 text-right text-sm">
          {graded ? (
            <>
              <span className="block">
                {s!.value}
                {a.points !== null && <span className="text-faint"> / {formatPoints(a.points)}</span>}
              </span>
              {s!.percent !== null && <span className="block text-xs text-muted-foreground">{formatPoints(Math.round(s!.percent * 10) / 10)}%</span>}
            </>
          ) : (
            <>
              <span className="hidden text-muted-foreground sm:block">{s ? 'Not scored' : 'Ungraded'}</span>
              <span className="text-faint sm:hidden" aria-hidden>—</span>
              <span className="sr-only sm:hidden">{s ? 'Not scored' : 'Ungraded'}</span>
            </>
          )}
        </span>
        {s?.grade ? <GradeBadge grade={s.grade} /> : <span className="hidden w-8 shrink-0 sm:block" />}
        <ChevronDown className={cn('size-4 shrink-0 text-faint transition-transform duration-200', open && 'rotate-180')} />
      </button>
      <div
        id={id}
        inert={!open}
        className={cn(
          'grid transition-[grid-template-rows,opacity] duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] motion-reduce:transition-none',
          open ? 'grid-rows-[1fr] opacity-100' : 'grid-rows-[0fr] opacity-0',
        )}
      >
        <div className="overflow-hidden">
        <div className="mx-2 mb-2 space-y-2 rounded-md glass-sunk px-3 py-2.5 text-sm wrap-anywhere">
          {s?.comment && (
            <p>
              <span className="text-muted-foreground">Teacher comment: </span>
              {s.comment}
            </p>
          )}
          {a.description && <p className="whitespace-pre-line text-muted-foreground">{a.description}</p>}
          <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-xs">
            <dt className="text-muted-foreground">Due</dt>
            <dd>{a.dueDate ? formatDate(a.dueDate, { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }) : '—'}</dd>
            <dt className="text-muted-foreground">Points possible</dt>
            <dd className="tabular">{a.points === null ? '—' : formatPoints(a.points)}</dd>
            <dt className="text-muted-foreground">Weight</dt>
            <dd className="tabular">{a.weight === null ? '—' : formatPoints(a.weight)}</dd>
            <dt className="text-muted-foreground">Counts toward grade</dt>
            <dd>{a.counted ? 'Yes' : 'No'}</dd>
          </dl>
        </div>
        </div>
      </div>
    </li>
  )
}
