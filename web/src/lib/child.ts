import { useSyncExternalStore } from 'react'

// Which student the app is showing: 0 means "the first one". Kept on this device.
const KEY = 'psm.child'

function load(): number {
  try {
    const n = Number(localStorage.getItem(KEY))
    return Number.isSafeInteger(n) && n > 0 ? n : 0
  } catch {
    return 0
  }
}

let current = load()
const listeners = new Set<() => void>()

export function setChild(id: number) {
  if (id === current) return
  current = id
  try {
    if (id) localStorage.setItem(KEY, String(id))
    else localStorage.removeItem(KEY)
  } catch {
    // Private mode: the choice lasts for this visit.
  }
  listeners.forEach((l) => l())
}

export function useChild(): number {
  return useSyncExternalStore(
    (cb) => {
      listeners.add(cb)
      return () => listeners.delete(cb)
    },
    () => current,
  )
}
