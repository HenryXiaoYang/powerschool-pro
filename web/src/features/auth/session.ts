import { useSyncExternalStore } from 'react'
import { ApiError, configureAuth, postSession } from '@/lib/api'
import { clearCache } from '@/lib/persist'
import { clearCredentials, hasCredentials, loadCredentials, saveCredentials } from './vault'

// The sealed API token lives for the tab (sessionStorage). With "Remember me", a new tab
// has no token yet but has credentials, and signs in silently on its first request.

const TOKEN = 'psm.token'
const PORTAL = 'psm.report-cards'

function read(key: string): string | null {
  try {
    return sessionStorage.getItem(key)
  } catch {
    return null
  }
}

function write(key: string, value: string | null) {
  try {
    if (value === null) sessionStorage.removeItem(key)
    else sessionStorage.setItem(key, value)
  } catch {
    /* private mode: memory only */
  }
}

let token = read(TOKEN)
let reportCards = read(PORTAL) !== 'false'
let signedIn = token !== null || hasCredentials()

const listeners = new Set<() => void>()
let snapshot = { signedIn, reportCards }
function notify() {
  snapshot = { signedIn, reportCards }
  listeners.forEach((l) => l())
}

function accept(res: { token: string; reportCards: boolean }) {
  token = res.token
  reportCards = res.reportCards
  signedIn = true
  write(TOKEN, token)
  write(PORTAL, String(reportCards))
  notify()
}

export async function signIn(username: string, password: string, remember: boolean) {
  accept(await postSession(username, password))
  await saveCredentials({ username, password }, remember)
}

export async function signOut() {
  token = null
  signedIn = false
  write(TOKEN, null)
  write(PORTAL, null)
  await Promise.all([clearCredentials(), clearCache()])
  notify()
}

let pending: Promise<string> | null = null

/** Single-flight silent sign-in; concurrent 401s share one login. */
function relogin(): Promise<string> {
  pending ??= (async () => {
    const creds = await loadCredentials()
    if (!creds) {
      await signOut()
      throw new ApiError(401, 'session_expired', 'Your session ended. Sign in again.')
    }
    try {
      accept(await postSession(creds.username, creds.password))
      return token as string
    } catch (err) {
      if (err instanceof ApiError && err.code === 'invalid_credentials') await signOut() // password changed
      throw err
    }
  })().finally(() => {
    pending = null
  })
  return pending
}

configureAuth({ token: () => token, relogin })

function subscribe(l: () => void) {
  listeners.add(l)
  return () => listeners.delete(l)
}

export function useSession() {
  return useSyncExternalStore(subscribe, () => snapshot)
}
