import { flushSync } from 'react-dom'

// View-transition choreography is chosen in index.css by html[data-vt]:
//   local      content swaps in place, only the selection pill travels
//   next/prev  the timetable slides the way you paged
export type TransitionKind = 'local' | 'next' | 'prev'

const reduced = () => matchMedia('(prefers-reduced-motion: reduce)').matches
let timer: ReturnType<typeof setTimeout> | undefined

/**
 * Tags the next React Router view transition (pass `{ viewTransition: true }` to the navigation).
 * shortcut: cleared on a timer because the router owns the transition object; fine while transitions stay under 700ms.
 */
export function markTransition(kind: TransitionKind) {
  const root = document.documentElement
  root.dataset.vt = kind
  clearTimeout(timer)
  timer = setTimeout(() => delete root.dataset.vt, 700)
}

/** Runs a local React state change inside a view transition. */
export function transition(kind: TransitionKind, update: () => void) {
  if (!('startViewTransition' in document) || reduced()) return update()
  const root = document.documentElement
  root.dataset.vt = kind
  document.startViewTransition(() => flushSync(update)).finished.finally(() => delete root.dataset.vt)
}
