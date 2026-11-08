import { describe, expect, it } from 'vitest'
import type { Assignment, Meeting, Schedule } from '@/api/student.gen'
import { applyFilter, groupByWeek } from '@/features/assignments/filter'
import { latestScores, today } from '@/features/home/today'
import { placeDay, weekPlan } from '@/features/schedule/week'
import { addDays, clockTime, mondayOf, schoolDate, weekday } from './dates'
import { gradeStep, scoreFlags } from './grades'

const TZ = 'Asia/Shanghai'

const meeting = (date: string, start: string, end: string, courseId = 1): Meeting => ({
  courseId,
  date,
  start: new Date(`${date}T${start}:00+08:00`).toISOString(),
  end: new Date(`${date}T${end}:00+08:00`).toISOString(),
})

const assignment = (over: Partial<Assignment>): Assignment => ({
  id: 1, courseId: 1, courseTitle: 'Maths', name: 'Quiz', description: '', category: 'Formative',
  dueAt: null, dueDate: '2026-11-04', points: 20, weight: 1, counted: true, score: null, ...over,
})

describe('dates', () => {
  it('uses the school zone, not the viewer’s', () => {
    // 16:00Z is midnight in Shanghai: the next calendar day there.
    expect(schoolDate('2026-11-06T16:00:00Z', TZ)).toBe('2026-11-07')
    expect(clockTime('2026-11-02T00:00:00Z', TZ)).toBe('08:00')
  })
  it('does calendar math on local dates', () => {
    expect(addDays('2026-10-31', 1)).toBe('2026-11-01')
    expect(weekday('2026-11-02')).toBe(0)
    expect(mondayOf('2026-11-08')).toBe('2026-11-02')
  })
})

describe('today', () => {
  const schedule: Schedule = {
    year: 2026, month: 11, days: [],
    meetings: [meeting('2026-11-04', '08:00', '08:50', 1), meeting('2026-11-04', '08:30', '09:30', 2), meeting('2026-11-04', '10:00', '10:50', 3), meeting('2026-11-05', '08:00', '08:50', 4)],
  }
  it('picks the latest-started running block as current', () => {
    const t = today(schedule, TZ, new Date('2026-11-04T08:40:00+08:00'))
    expect(t.current?.courseId).toBe(2)
    expect(t.next?.courseId).toBe(3)
    expect(t.meetings).toHaveLength(3)
  })
  it('looks past today for the next block', () => {
    const t = today(schedule, TZ, new Date('2026-11-04T18:00:00+08:00'))
    expect(t.current).toBeNull()
    expect(t.next?.courseId).toBe(4)
  })
  it('prefers the next class over boarding blocks and fills empty days', () => {
    const t = today(schedule, TZ, new Date('2026-11-04T09:40:00+08:00'), (m) => m.courseId !== 3)
    expect(t.next?.courseId).toBe(4)
    const sat = today(schedule, TZ, new Date('2026-11-03T20:00:00+08:00'))
    expect(sat.upcomingDate).toBe('2026-11-04')
    expect(sat.upcoming).toHaveLength(3)
  })
})

describe('week', () => {
  it('lays overlapping blocks side by side', () => {
    const placed = placeDay([meeting('2026-11-04', '08:00', '09:00', 1), meeting('2026-11-04', '08:30', '09:30', 2), meeting('2026-11-04', '10:00', '11:00', 3)], TZ, 8 * 60)
    const byCourse = new Map(placed.map((p) => [p.meeting.courseId, p]))
    expect(byCourse.get(1)).toMatchObject({ lane: 0, lanes: 2, top: 0, height: 60 })
    expect(byCourse.get(2)).toMatchObject({ lane: 1, lanes: 2, top: 30 })
    expect(byCourse.get(3)).toMatchObject({ lane: 0, lanes: 1 })
  })
  it('includes weekends only when something meets', () => {
    const plan = weekPlan('2026-11-02', [{ year: 2026, month: 11, days: [{ date: '2026-11-02', day: 'MA', week: 'A' }], meetings: [meeting('2026-11-08', '19:00', '21:00')] }], [])
    expect(plan.map((p) => p.date)).toEqual(['2026-11-02', '2026-11-03', '2026-11-04', '2026-11-05', '2026-11-06', '2026-11-08'])
    expect(plan[0].code).toBe('MA')
  })
})

describe('grades', () => {
  it('maps IB and IGCSE grades onto one ramp', () => {
    expect(gradeStep('7')).toBe(7)
    expect(gradeStep('A*')).toBe(7)
    expect(gradeStep('pass')).toBeNull()
  })
  it('flags collected-but-unscored work and uncounted work', () => {
    const s = { value: '--', percent: null, grade: '', comment: '', late: true, missing: false, exempt: false, collected: true, incomplete: false }
    expect(scoreFlags(s, false)).toEqual(['late', 'collected', 'not counted'])
  })
})

describe('assignments', () => {
  const scored = assignment({ id: 1, score: { value: '0', percent: 0, grade: '1', comment: '', late: false, missing: false, exempt: false, collected: false, incomplete: false } })
  const dash = assignment({ id: 2, score: { value: '--', percent: null, grade: '', comment: '', late: false, missing: true, exempt: false, collected: false, incomplete: false } })
  const upcoming = assignment({ id: 3, dueDate: '2026-11-20' })
  const all = [scored, dash, upcoming]
  const f = { courseId: null, category: null, query: '' }

  it('treats a zero as a score and "--" as not scored', () => {
    expect(applyFilter(all, { ...f, status: 'graded' }, '2026-11-07').map((a) => a.id)).toEqual([1])
    expect(latestScores(all).map((a) => a.id)).toEqual([1])
  })
  it('filters upcoming and flagged work', () => {
    expect(applyFilter(all, { ...f, status: 'upcoming' }, '2026-11-07').map((a) => a.id)).toEqual([3])
    expect(applyFilter(all, { ...f, status: 'flagged' }, '2026-11-07').map((a) => a.id)).toEqual([2])
  })
  it('labels weeks relative to today', () => {
    expect(groupByWeek(all, '2026-11-07').map((g) => g.label)).toEqual(['Week of 16 Nov', 'This week'])
  })
})
