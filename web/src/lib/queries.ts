import { queryOptions, useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import type { ReportCardList } from '@/api/api.gen'
import type { Course, Schedule, Snapshot } from '@/api/student.gen'
import { request } from './api'
import { useChild } from './child'
import { courseHues } from './grades'

/** Everything for one student; `child` 0 is the account's first student. */
export const snapshotQuery = (child: number) =>
  queryOptions({
    queryKey: ['student', child],
    queryFn: () => request<Snapshot>(child ? `/api/student?child=${child}` : '/api/student'),
  })

export function useSnapshot() {
  return useQuery(snapshotQuery(useChild()))
}

/** Courses by id plus their hue, derived once per snapshot. */
export function useCourses(snap: Snapshot | undefined) {
  return useMemo(() => {
    const byId = new Map<number, Course>()
    for (const c of snap?.courses ?? []) byId.set(c.id, c)
    return { byId, hues: courseHues(snap?.courses ?? []) }
  }, [snap])
}

/** A month of meetings. The snapshot already holds the current month. */
export function useSchedule(
  snap: Snapshot | undefined,
  year: number,
  month: number,
): { data: Schedule | undefined; isPending: boolean; error: Error | null; refetch: () => void } {
  const child = useChild()
  const inSnapshot = snap?.schedule.year === year && snap.schedule.month === month
  const q = useQuery({
    queryKey: ['schedule', child, year, month],
    queryFn: () => request<Schedule>(`/api/schedule?year=${year}&month=${month}${child ? `&child=${child}` : ''}`),
    enabled: !!snap && !inSnapshot,
  })
  if (inSnapshot) return { data: snap.schedule, isPending: false, error: null, refetch: () => {} }
  return { data: q.data, isPending: q.isPending, error: q.error, refetch: () => void q.refetch() }
}

export function useReportCards(enabled: boolean) {
  return useQuery({
    queryKey: ['report-cards'],
    queryFn: () => request<ReportCardList>('/api/report-cards').then((r) => r.reportCards),
    enabled,
  })
}

export function useReportCard(id: string) {
  return useQuery({
    queryKey: ['report-card', id],
    queryFn: () => request<string>(`/api/report-cards/${encodeURIComponent(id)}`, 'text'),
    staleTime: Infinity,
  })
}
