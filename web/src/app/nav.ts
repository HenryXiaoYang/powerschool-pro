import type { LucideIcon } from 'lucide-react'
import { BookOpen, CalendarDays, ClipboardCheck, FileText, GraduationCap, House, UserCheck } from 'lucide-react'

export interface NavItem {
  to: string
  label: string
  icon: LucideIcon
  /** Shown in the phone tab bar; the rest live under More. */
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
]

export function titleFor(pathname: string): string {
  if (pathname === '/') return 'Home'
  return NAV.find((n) => n.to !== '/' && pathname.startsWith(n.to))?.label ?? ''
}
