import type { LucideIcon } from 'lucide-react'
import { BookOpen, CalendarDays, ClipboardCheck, FileText, GraduationCap, House, Settings, UserCheck } from 'lucide-react'

export interface NavItem {
  to: string
  label: string
  icon: LucideIcon
  /** Kept for callers that still split primary from secondary destinations. */
  tab: boolean
}

export const NAV: NavItem[] = [
  { to: '/', label: 'Home', icon: House, tab: true },
  { to: '/schedule', label: 'Schedule', icon: CalendarDays, tab: true },
  { to: '/classes', label: 'Classes', icon: BookOpen, tab: true },
  { to: '/assignments', label: 'Assignments', icon: ClipboardCheck, tab: true },
  { to: '/grades', label: 'Grades', icon: GraduationCap, tab: false },
  { to: '/attendance', label: 'Attendance', icon: UserCheck, tab: false },
  { to: '/report', label: 'Report', icon: FileText, tab: false },
  { to: '/settings', label: 'Settings', icon: Settings, tab: false },
]

export function titleFor(pathname: string): string {
  if (pathname === '/') return 'Home'
  return NAV.find((n) => n.to !== '/' && pathname.startsWith(n.to))?.label ?? ''
}
