import type { Assignment, Meeting, Schedule } from '@/api/student.gen'
import { schoolDate } from '@/lib/dates'

export interface Today {
  date: string
  meetings: Meeting[]
  /** The block running now; when blocks overlap, the latest-started one. */
  current: Meeting | null
  /** The first class starting after now (any block if no class follows), today or later this month. */
  next: Meeting | null
  /** When today has nothing, the next date that does, and its meetings. */
  upcomingDate: string | null
  upcoming: Meeting[]
}

export function today(schedule: Schedule, timeZone: string, now: Date, isClass: (m: Meeting) => boolean = () => true): Today {
  const date = schoolDate(now, timeZone)
  const t = now.getTime()
  const meetings = schedule.meetings.filter((m) => m.date === date)
  const running = meetings.filter((m) => Date.parse(m.start) <= t && t < Date.parse(m.end))
  const later = schedule.meetings.filter((m) => Date.parse(m.start) > t)
  const upcomingDate = meetings.length === 0 ? (later.find((m) => m.date > date)?.date ?? null) : null
  return {
    date,
    meetings,
    current: running.at(-1) ?? null,
    next: later.find(isClass) ?? later[0] ?? null,
    upcomingDate,
    upcoming: upcomingDate ? schedule.meetings.filter((m) => m.date === upcomingDate) : [],
  }
}

/** Newest scored work first; "--" and empty values are not scores yet. */
export function latestScores(assignments: Assignment[], limit = 5): Assignment[] {
  return assignments
    .filter((a) => a.score && a.score.value !== '' && a.score.value !== '--' && !a.score.exempt)
    .sort((a, b) => b.dueDate.localeCompare(a.dueDate))
    .slice(0, limit)
}
