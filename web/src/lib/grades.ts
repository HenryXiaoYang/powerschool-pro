import type { Course, Score } from '@/api/student.gen'

/** Step 7 (best) … 1 on the grade ramp, or null when a grade has no place on it. */
export function gradeStep(grade: string): number | null {
  const g = grade.trim().toUpperCase()
  if (/^[1-7]$/.test(g)) return Number(g)
  // IGCSE letters (older report cards and some courses).
  const letters: Record<string, number> = { 'A*': 7, A: 6, B: 5, C: 4, D: 3, E: 2, F: 1, G: 1 }
  return letters[g] ?? null
}

/** Ten hues around the OKLCH ring; academic courses take them in timetable order. */
const HUES = [130, 200, 25, 270, 165, 60, 305, 235, 95, 340]

export function courseHues(courses: Course[]): Map<number, number> {
  const hues = new Map<number, number>()
  courses
    .filter((c) => c.academic)
    .forEach((c, i) => hues.set(c.id, HUES[i % HUES.length]))
  return hues
}

/** CSS colour for a course; non-academic blocks are neutral. */
export function hueColor(hue: number | undefined): string {
  return hue === undefined ? 'var(--faint)' : `oklch(var(--course-l) var(--course-c) ${hue})`
}

export function formatPoints(n: number | null): string {
  if (n === null) return ''
  return Number.isInteger(n) ? String(n) : n.toFixed(2).replace(/0+$/, '').replace(/\.$/, '')
}

/** Badge fill hue per ramp step: lime (7) through teal and amber to rose (1). */
const RAMP: Record<number, number> = { 7: 128, 6: 150, 5: 176, 4: 88, 3: 58, 2: 30, 1: 18 }

export function rampHue(step: number): number {
  return RAMP[step]
}

export type Flag = 'late' | 'missing' | 'exempt' | 'collected' | 'incomplete' | 'not counted'

export function scoreFlags(score: Score | null, counted: boolean): Flag[] {
  const flags: Flag[] = []
  if (score?.missing) flags.push('missing')
  if (score?.late) flags.push('late')
  if (score?.incomplete) flags.push('incomplete')
  if (score?.exempt) flags.push('exempt')
  if (score?.collected && (score.value === '--' || score.value === '')) flags.push('collected')
  if (!counted) flags.push('not counted')
  return flags
}
