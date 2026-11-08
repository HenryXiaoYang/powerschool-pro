import type { FinalGrade, ReportingTerm, Snapshot } from '@/api/student.gen'

/** courseId → reportingTermId → stored grade. */
export function gradeTable(snap: Snapshot): Map<number, Map<number, FinalGrade>> {
  const table = new Map<number, Map<number, FinalGrade>>()
  for (const g of snap.finalGrades) {
    const row = table.get(g.courseId) ?? new Map<number, FinalGrade>()
    row.set(g.reportingTermId, g)
    table.set(g.courseId, row)
  }
  return table
}

/** Terms that have at least one grade, in calendar order. */
export function gradedTerms(snap: Snapshot): ReportingTerm[] {
  const ids = new Set(snap.finalGrades.map((g) => g.reportingTermId))
  return snap.reportingTerms.filter((t) => ids.has(t.id))
}

/** The course's grade for the current term, else its most recent graded term. */
export function latestGrade(snap: Snapshot, table: ReturnType<typeof gradeTable>, courseId: number): { grade: FinalGrade; term: ReportingTerm } | null {
  const row = table.get(courseId)
  if (!row) return null
  const terms = [...snap.reportingTerms].sort((a, b) => Number(b.current) - Number(a.current) || b.end.localeCompare(a.end))
  for (const term of terms) {
    const grade = row.get(term.id)
    if (grade) return { grade, term }
  }
  return null
}
