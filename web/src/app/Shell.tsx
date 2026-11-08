import { useIsFetching, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Ellipsis, LogOut, RotateCw } from 'lucide-react'
import { useState } from 'react'
import { Link, NavLink, Outlet, ScrollRestoration, useLocation } from 'react-router'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { signOut } from '@/features/auth/session'
import { ago, clockTime, formatDate, schoolDate } from '@/lib/dates'
import { useSnapshot } from '@/lib/queries'
import { useNow } from '@/lib/useNow'
import { cn } from '@/lib/utils'
import { NAV, titleFor } from './nav'

export function Backdrop() {
  return (
    <div className="mesh" aria-hidden>
      <span />
      <span />
      <span />
      <span />
    </div>
  )
}

function isActive(to: string, pathname: string) {
  return to === '/' ? pathname === '/' : pathname.startsWith(to)
}

const SLIDE = 'transition-transform duration-[380ms] ease-[cubic-bezier(0.16,1,0.3,1)] motion-reduce:transition-none'

function Rail() {
  const { pathname } = useLocation()
  const active = NAV.findIndex((n) => isActive(n.to, pathname))
  return (
    <nav aria-label="Main" className="no-print sticky top-5 hidden self-start rail:block lg:top-8">
      <div className="glass relative flex flex-col items-center gap-1.5 rounded-full p-1.5">
        {/* The selection bead slides between items: 44px item + 6px gap. */}
        {active >= 0 && (
          <span aria-hidden className={cn('bead pointer-events-none absolute top-1.5 left-1.5 size-11 rounded-full bg-primary', SLIDE)} style={{ transform: `translateY(${active * 50}px)` }} />
        )}
        {NAV.map(({ to, label, icon: Icon }) => (
          <Tooltip key={to}>
            <TooltipTrigger asChild>
              <NavLink
                to={to}
                aria-label={label}
                className={cn(
                  'relative grid size-11 place-items-center rounded-full text-muted-foreground transition duration-200 hover:bg-accent hover:text-foreground active:scale-[0.94]',
                  isActive(to, pathname) && 'text-primary-foreground hover:bg-transparent hover:text-primary-foreground',
                )}
              >
                <Icon className="size-5" strokeWidth={1.8} />
              </NavLink>
            </TooltipTrigger>
            <TooltipContent side="right">{label}</TooltipContent>
          </Tooltip>
        ))}
        <span className="my-1 h-px w-6 bg-border" />
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              onClick={() => void signOut()}
              aria-label="Sign out"
              className="grid size-11 place-items-center rounded-full text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            >
              <LogOut className="size-5" strokeWidth={1.8} />
            </button>
          </TooltipTrigger>
          <TooltipContent side="right">Sign out</TooltipContent>
        </Tooltip>
      </div>
    </nav>
  )
}

function TabBar() {
  const { pathname } = useLocation()
  const [more, setMore] = useState(false)
  const moreActive = NAV.some((n) => !n.tab && isActive(n.to, pathname))
  const tabs = NAV.filter((n) => n.tab)
  const active = moreActive ? tabs.length : tabs.findIndex((n) => isActive(n.to, pathname))
  const item = 'relative flex h-14 flex-1 flex-col items-center justify-center gap-0.5 rounded-full text-[11px] font-medium transition duration-200 active:scale-[0.94]'
  return (
    <nav aria-label="Main" className="no-print fixed inset-x-3 bottom-[max(0.75rem,env(safe-area-inset-bottom))] z-40 rail:hidden">
      <div className="glass relative mx-auto flex max-w-lg rounded-full p-1">
        {active >= 0 && (
          <span
            aria-hidden
            className={cn('bead pointer-events-none absolute inset-y-1 left-1 rounded-full bg-primary', SLIDE)}
            style={{ width: `calc((100% - 0.5rem) / ${tabs.length + 1})`, transform: `translateX(${active * 100}%)` }}
          />
        )}
        {tabs.map(({ to, label, icon: Icon }) => (
          <NavLink key={to} to={to} className={cn(item, isActive(to, pathname) ? 'text-primary-foreground' : 'text-muted-foreground')}>
            <Icon className="size-5" strokeWidth={1.8} />
            {label}
          </NavLink>
        ))}
        <Sheet open={more} onOpenChange={setMore}>
          <SheetTrigger className={cn(item, moreActive ? 'text-primary-foreground' : 'text-muted-foreground')}>
            <Ellipsis className="size-5" strokeWidth={1.8} />
            More
          </SheetTrigger>
          <SheetContent side="bottom" className="glass rounded-t-2xl border-x-0 border-b-0 bg-popover pb-[max(1.5rem,env(safe-area-inset-bottom))]">
            <SheetHeader>
              <SheetTitle>More</SheetTitle>
              <SheetDescription className="sr-only">Other sections and sign out</SheetDescription>
            </SheetHeader>
            <div className="grid gap-1 px-4">
              {NAV.filter((n) => !n.tab).map(({ to, label, icon: Icon }) => (
                <Link
                  key={to}
                  to={to}
                  onClick={() => setMore(false)}
                  className={cn('flex h-12 items-center gap-3 rounded-md px-3 font-medium', isActive(to, pathname) ? 'bead bg-primary text-primary-foreground' : 'hover:bg-accent')}
                >
                  <Icon className="size-5" strokeWidth={1.8} /> {label}
                </Link>
              ))}
              <button onClick={() => void signOut()} className="flex h-12 items-center gap-3 rounded-md px-3 font-medium text-muted-foreground hover:bg-accent">
                <LogOut className="size-5" strokeWidth={1.8} /> Sign out
              </button>
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </nav>
  )
}

function Freshness() {
  const qc = useQueryClient()
  const fetching = useIsFetching() > 0
  const { dataUpdatedAt, error, data } = useSnapshot()
  const now = useNow(60_000)
  return (
    <div className="flex items-center gap-2 text-xs text-muted-foreground">
      {data && (
        <span className={cn('hidden sm:inline', error && 'text-warn')} title={error?.message}>
          {error ? 'Couldn’t refresh · saved ' : 'Updated '}
          {ago(dataUpdatedAt, now.getTime())}
        </span>
      )}
      <Button
        variant="ghost"
        size="icon"
        className="size-11 rounded-full md:pointer-fine:size-9"
        aria-label="Refresh"
        disabled={fetching}
        onClick={() => void qc.invalidateQueries()}
      >
        <RotateCw className={cn('size-4', fetching && 'animate-spin')} />
      </Button>
    </div>
  )
}

function greeting(hour: number) {
  return hour < 5 ? 'Good night' : hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening'
}

function Title({ pathname }: { pathname: string }) {
  const { data: snap } = useSnapshot()
  const now = useNow(60_000)
  if (/^\/classes\/[^/]+/.test(pathname)) {
    return (
      <Link to="/classes" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
        <ArrowLeft className="size-4" /> Classes
      </Link>
    )
  }
  if (pathname !== '/' || !snap) return <h1 className="text-2xl font-semibold tracking-tight">{titleFor(pathname) || 'Home'}</h1>
  const date = schoolDate(now, snap.timeZone)
  const hour = Number(clockTime(now.toISOString(), snap.timeZone).slice(0, 2))
  const cycle = snap.schedule.days.find((d) => d.date === date)
  return (
    <div className="min-w-0">
      <h1 className="truncate text-2xl font-semibold tracking-tight">
        {greeting(hour)}
        {snap.profile.firstName && `, ${snap.profile.firstName}`}
      </h1>
      <p className="text-sm text-muted-foreground">
        {formatDate(date, { weekday: 'long', day: 'numeric', month: 'long' })}
        {cycle?.day && ` · Day ${cycle.day}`}
      </p>
    </div>
  )
}

export function Shell() {
  const { pathname } = useLocation()
  return (
    <div className="mx-auto flex min-h-dvh max-w-[1500px] items-start gap-5 md:p-5 lg:gap-6 lg:p-8">
      <Rail />
      <div className="min-w-0 flex-1 md:glass md:min-h-[calc(100dvh-2.5rem)] md:rounded-2xl lg:min-h-[calc(100dvh-4rem)]">
        <header className="flex min-h-[4.75rem] items-center justify-between gap-3 px-4 pt-5 pb-3 sm:px-6 md:min-h-[5.5rem] md:pt-6">
          <Title pathname={pathname} />
          <Freshness />
        </header>
        <main key={pathname} className="settle px-3 pb-28 sm:px-6 rail:pb-6">
          <Outlet />
        </main>
      </div>
      <TabBar />
      {/* New pages open at the top; back/forward restores where you were. */}
      <ScrollRestoration />
    </div>
  )
}
