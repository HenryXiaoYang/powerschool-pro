import type { Holiday, Meeting, Schedule } from '@/api/student.gen'
import { addDays, minuteOfDay } from '@/lib/dates'

export interface DayPlan {
  date: string
  /** Rotation day such as "MA"; "" when PowerSchool's meetings do not pin it down. */
  code: string
  week: string
  holiday: Holiday | null
  meetings: Meeting[]
}

/** Monday–Friday of the week, plus weekend days that have meetings. */
export function weekPlan(monday: string, schedules: Schedule[], holidays: Holiday[]): DayPlan[] {
  const meetings = schedules.flatMap((s) => s.meetings)
  const cycle = new Map(schedules.flatMap((s) => s.days).map((d) => [d.date, d]))
  const off = new Map(holidays.map((h) => [h.date, h]))
  const plans: DayPlan[] = []
  for (let i = 0; i < 7; i++) {
    const date = addDays(monday, i)
    const day = meetings.filter((m) => m.date === date).sort((a, b) => a.start.localeCompare(b.start))
    if (i >= 5 && day.length === 0) continue
    plans.push({ date, code: cycle.get(date)?.day ?? '', week: cycle.get(date)?.week ?? '', holiday: off.get(date) ?? null, meetings: day })
  }
  return plans
}

export interface Placed {
  meeting: Meeting
  top: number // minutes after the grid start
  height: number // minutes
  lane: number
  lanes: number
}

/** Side-by-side lanes for overlapping blocks; each overlap cluster shares a lane count. */
export function placeDay(meetings: Meeting[], timeZone: string, gridStart: number): Placed[] {
  const items = meetings
    .map((meeting) => {
      const start = minuteOfDay(meeting.start, timeZone)
      const end = Math.max(start + 10, minuteOfDay(meeting.end, timeZone))
      return { meeting, start, end, lane: 0 }
    })
    .sort((a, b) => a.start - b.start || b.end - a.end)

  const out: Placed[] = []
  let cluster: typeof items = []
  let clusterEnd = -1
  const flush = () => {
    const laneEnds: number[] = []
    for (const it of cluster) {
      let lane = laneEnds.findIndex((end) => end <= it.start)
      if (lane === -1) lane = laneEnds.push(0) - 1
      laneEnds[lane] = it.end
      it.lane = lane
    }
    for (const it of cluster) {
      out.push({ meeting: it.meeting, top: it.start - gridStart, height: it.end - it.start, lane: it.lane, lanes: laneEnds.length })
    }
    cluster = []
  }
  for (const it of items) {
    if (it.start >= clusterEnd) flush()
    cluster.push(it)
    clusterEnd = Math.max(clusterEnd, it.end)
  }
  flush()
  return out
}

/** Whole hours covering every meeting in the week. */
export function gridHours(plans: DayPlan[], timeZone: string): { start: number; end: number } {
  let start = 24 * 60
  let end = 0
  for (const p of plans) {
    for (const m of p.meetings) {
      start = Math.min(start, minuteOfDay(m.start, timeZone))
      end = Math.max(end, minuteOfDay(m.end, timeZone))
    }
  }
  if (end === 0) return { start: 8 * 60, end: 16 * 60 }
  return { start: Math.floor(start / 60) * 60, end: Math.ceil(end / 60) * 60 }
}
