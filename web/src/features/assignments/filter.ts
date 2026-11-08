import type { Assignment } from '@/api/student.gen'
import { addDays, formatDate, mondayOf } from '@/lib/dates'

export type Status = 'all' | 'graded' | 'upcoming' | 'flagged'

export interface Filter {
  courseId: number | null
  category: string | null
  status: Status
  query: string
}

export const isGraded = (a: Assignment) => !!a.score && a.score.value !== '' && a.score.value !== '--'
export const isFlagged = (a: Assignment) => !!a.score && (a.score.missing || a.score.late || a.score.incomplete)

export function applyFilter(list: Assignment[], f: Filter, today: string): Assignment[] {
  const q = f.query.trim().toLowerCase()
  return list.filter((a) => {
    if (f.courseId !== null && a.courseId !== f.courseId) return false
    if (f.category !== null && a.category !== f.category) return false
    if (q && !`${a.name} ${a.courseTitle} ${a.category}`.toLowerCase().includes(q)) return false
    switch (f.status) {
      case 'graded':
        return isGraded(a)
      case 'upcoming':
        return !a.score && (a.dueDate === '' || a.dueDate >= today)
      case 'flagged':
        return isFlagged(a)
    }
    return true
  })
}

export interface Group {
  label: string
  items: Assignment[]
}

/** Groups by due week, newest first, with friendly labels near today. */
export function groupByWeek(list: Assignment[], today: string): Group[] {
  const thisWeek = mondayOf(today)
  const groups = new Map<string, Assignment[]>()
  for (const a of [...list].sort((x, y) => y.dueDate.localeCompare(x.dueDate))) {
    const key = a.dueDate ? mondayOf(a.dueDate) : ''
    groups.set(key, [...(groups.get(key) ?? []), a])
  }
  return [...groups].map(([monday, items]) => ({
    label:
      monday === ''
        ? 'No due date'
        : monday === thisWeek
          ? 'This week'
          : monday === addDays(thisWeek, 7)
            ? 'Next week'
            : monday === addDays(thisWeek, -7)
              ? 'Last week'
              : `Week of ${formatDate(monday, { day: 'numeric', month: 'short' })}`,
    items,
  }))
}
