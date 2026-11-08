// Local dates are "YYYY-MM-DD" strings in the school's time zone (the API sends them so).
// Calendar math runs on them at UTC noon, so the viewer's own zone never shifts a day.

const fmtCache = new Map<string, Intl.DateTimeFormat>()
function fmt(opts: Intl.DateTimeFormatOptions): Intl.DateTimeFormat {
  const key = JSON.stringify(opts)
  let f = fmtCache.get(key)
  if (!f) fmtCache.set(key, (f = new Intl.DateTimeFormat('en-GB', opts)))
  return f
}

/** The school-local calendar date of an instant. */
export function schoolDate(instant: Date | string, timeZone: string): string {
  const parts = fmt({ timeZone, year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(new Date(instant))
  const get = (t: string) => parts.find((p) => p.type === t)?.value
  return `${get('year')}-${get('month')}-${get('day')}`
}

/** "08:05" in school time. */
export function clockTime(instant: string, timeZone: string): string {
  return fmt({ timeZone, hour: '2-digit', minute: '2-digit', hourCycle: 'h23' }).format(new Date(instant))
}

/** Minutes since school-local midnight. */
export function minuteOfDay(instant: string, timeZone: string): number {
  const [h, m] = clockTime(instant, timeZone).split(':').map(Number)
  return h * 60 + m
}

const noon = (date: string) => new Date(`${date}T12:00:00Z`)

export function addDays(date: string, n: number): string {
  const d = noon(date)
  d.setUTCDate(d.getUTCDate() + n)
  return d.toISOString().slice(0, 10)
}

/** 0 = Monday … 6 = Sunday. */
export function weekday(date: string): number {
  return (noon(date).getUTCDay() + 6) % 7
}

export function mondayOf(date: string): string {
  return addDays(date, -weekday(date))
}

export function formatDate(date: string, opts: Intl.DateTimeFormatOptions = { weekday: 'short', day: 'numeric', month: 'short' }): string {
  return date ? fmt({ ...opts, timeZone: 'UTC' }).format(noon(date)) : ''
}

export function daysBetween(from: string, to: string): number {
  return Math.round((noon(to).getTime() - noon(from).getTime()) / 86_400_000)
}

/** "just now", "4 min ago", "2 h ago", "3 days ago". */
export function ago(instant: string | number, now = Date.now()): string {
  const min = Math.round((now - new Date(instant).getTime()) / 60_000)
  if (min < 1) return 'just now'
  if (min < 60) return `${min} min ago`
  const h = Math.round(min / 60)
  if (h < 24) return `${h} h ago`
  const d = Math.round(h / 24)
  return `${d} day${d === 1 ? '' : 's'} ago`
}

/** "in 25 min", "in 1 h 10 min". */
export function inMinutes(min: number): string {
  if (min < 60) return `in ${min} min`
  const h = Math.floor(min / 60)
  const m = min % 60
  return m ? `in ${h} h ${m} min` : `in ${h} h`
}
