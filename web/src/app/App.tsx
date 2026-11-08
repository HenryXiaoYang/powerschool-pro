import { PersistQueryClientProvider } from '@tanstack/react-query-persist-client'
import { createBrowserRouter, Navigate, Outlet, RouterProvider } from 'react-router'
import { TooltipProvider } from '@/components/ui/tooltip'
import { LoginPage } from '@/features/auth/LoginPage'
import { useSession } from '@/features/auth/session'
import { AssignmentsPage } from '@/features/assignments/AssignmentsPage'
import { AttendancePage } from '@/features/attendance/AttendancePage'
import { ClassesPage } from '@/features/classes/ClassesPage'
import { CoursePage } from '@/features/classes/CoursePage'
import { GradesPage } from '@/features/grades/GradesPage'
import { HomePage } from '@/features/home/HomePage'
import { ReportCardPage } from '@/features/report/ReportCardPage'
import { ReportPage } from '@/features/report/ReportPage'
import { SchedulePage } from '@/features/schedule/SchedulePage'
import { SettingsPage } from '@/features/settings/SettingsPage'
import { persistOptions, queryClient } from '@/lib/persist'
import { RouteError } from './RouteError'
import { Backdrop, Shell } from './Shell'

function RequireSession() {
  const { signedIn } = useSession()
  return signedIn ? <Outlet /> : <Navigate to="/login" replace />
}

// A crash inside a page keeps the shell and nav; one outside it gets a bare screen.
const crashed = (
  <main className="grid min-h-dvh place-items-center p-4">
    <RouteError />
  </main>
)

const router = createBrowserRouter([
  { path: '/login', element: <LoginPage />, errorElement: crashed },
  {
    element: <RequireSession />,
    errorElement: crashed,
    children: [
      {
        element: <Shell />,
        children: [
          {
            errorElement: <RouteError />,
            children: [
              { index: true, element: <HomePage /> },
              { path: 'schedule', element: <SchedulePage /> },
              { path: 'classes', element: <ClassesPage /> },
              { path: 'classes/:id', element: <CoursePage /> },
              { path: 'grades', element: <GradesPage /> },
              { path: 'assignments', element: <AssignmentsPage /> },
              { path: 'attendance', element: <AttendancePage /> },
              { path: 'settings', element: <SettingsPage /> },
              { path: 'report', element: <ReportPage /> },
              { path: 'report/:id', element: <ReportCardPage /> },
              { path: '*', element: <Navigate to="/" replace /> },
            ],
          },
        ],
      },
    ],
  },
])

export function App() {
  return (
    <PersistQueryClientProvider client={queryClient} persistOptions={persistOptions}>
      <TooltipProvider delayDuration={300}>
        <Backdrop />
        <RouterProvider router={router} />
      </TooltipProvider>
    </PersistQueryClientProvider>
  )
}
