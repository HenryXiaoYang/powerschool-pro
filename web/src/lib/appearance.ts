import { useSyncExternalStore } from 'react'

/**
 * Appearance settings, kept on this device. They only drive CSS variables and one data
 * attribute on <html>, so changing them never re-renders the app.
 */
export interface Appearance {
  /** 20–200 (%): scales every glass surface's opacity. 100 is the designed look. */
  opacity: number
  /** 0–40 px of backdrop blur on glass surfaces. */
  blur: number
  /** Specular rim + sheen on glass ("liquid"), or plain frosted glass. */
  liquid: boolean
  /** Stop the drifting backdrop and the soft-focus settle: cheaper on older phones. */
  calm: boolean
}

// Phones composite every blurred pane each frame, so they start with a lighter blur.
const phone = typeof matchMedia === 'function' && matchMedia('(max-width: 767px)').matches
export const DEFAULTS: Appearance = { opacity: 100, blur: phone ? 16 : 24, liquid: true, calm: false }
const KEY = 'psm.appearance'

export function clampAppearance(raw: unknown): Appearance {
  const r = (raw && typeof raw === 'object' ? raw : {}) as Partial<Appearance>
  const num = (v: unknown, lo: number, hi: number, d: number) => (typeof v === 'number' && Number.isFinite(v) ? Math.min(hi, Math.max(lo, Math.round(v))) : d)
  return {
    opacity: num(r.opacity, 20, 200, DEFAULTS.opacity),
    blur: num(r.blur, 0, 40, DEFAULTS.blur),
    liquid: typeof r.liquid === 'boolean' ? r.liquid : DEFAULTS.liquid,
    calm: typeof r.calm === 'boolean' ? r.calm : DEFAULTS.calm,
  }
}

function load(): Appearance {
  try {
    return clampAppearance(JSON.parse(localStorage.getItem(KEY) ?? 'null'))
  } catch {
    return DEFAULTS
  }
}

let current = load()
const listeners = new Set<() => void>()

/** Write the settings onto <html>. Called on load and after every change. */
export function applyAppearance(a: Appearance = current) {
  const s = document.documentElement.style
  s.setProperty('--glass-k', String(a.opacity / 100))
  // Blur 0 drops the filter entirely: a backdrop-filter, even of 0px, still costs a compositing layer.
  s.setProperty('--glass-filter', a.blur === 0 ? 'none' : `blur(${a.blur}px) saturate(${a.liquid ? 180 : 150}%)`)
  document.documentElement.dataset.glass = a.liquid ? 'liquid' : 'frost'
  if (a.calm) document.documentElement.dataset.calm = ''
  else delete document.documentElement.dataset.calm
}

export function setAppearance(patch: Partial<Appearance>) {
  current = clampAppearance({ ...current, ...patch })
  applyAppearance(current)
  try {
    localStorage.setItem(KEY, JSON.stringify(current))
  } catch {
    // Private mode or full storage: the change still applies for this visit.
  }
  listeners.forEach((l) => l())
}

export function useAppearance(): Appearance {
  return useSyncExternalStore(
    (cb) => {
      listeners.add(cb)
      return () => listeners.delete(cb)
    },
    () => current,
  )
}
