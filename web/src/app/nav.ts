import type { LucideIcon } from 'lucide-react'
import { BookOpen, CalendarDays, ClipboardCheck, FileText, GraduationCap, House, Settings, UserCheck } from 'lucide-react'

export interface NavItem {
  to: string
  label: string
  icon: LucideIcon
}

export const NAV: NavItem[] = [
  { to: '/', label: 'Home', icon: House },
  { to: '/schedule', label: 'Schedule', icon: CalendarDays },
  { to: '/classes', label: 'Classes', icon: BookOpen },
  { to: '/assignments', label: 'Assignments', icon: ClipboardCheck },
  { to: '/grades', label: 'Grades', icon: GraduationCap },
  { to: '/attendance', label: 'Attendance', icon: UserCheck },
  { to: '/report', label: 'Report', icon: FileText },
  { to: '/settings', label: 'Settings', icon: Settings },
]

export function titleFor(pathname: string): string {
  if (pathname === '/') return 'Home'
  return NAV.find((n) => n.to !== '/' && pathname.startsWith(n.to))?.label ?? ''
}
