import { useSyncExternalStore } from 'react'
import { ApiError, configureAuth, deleteParent, postSession, putParent } from '@/lib/api'
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
let parentLinked = hasCredentials('parent')
let snapshot = { signedIn, reportCards, parentLinked }
function notify() {
  snapshot = { signedIn, reportCards, parentLinked }
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
  parentLinked = false
  write(TOKEN, null)
  write(PORTAL, null)
  await Promise.all([clearCredentials(), clearCache()])
  notify()
}

/** Attach the saved parent login to the current session. Never throws: a parent problem must not cost the main sign-in. */
async function attachParent() {
  const creds = await loadCredentials('parent')
  if (!creds) return
  try {
    accept(await putParent(token, creds.username, creds.password))
  } catch (err) {
    if (err instanceof ApiError && err.code === 'invalid_credentials') {
      await clearCredentials('parent') // the parent password changed: stop retrying it
      parentLinked = false
      notify()
    }
  }
}

/** Sign a parent account in beside the current session and remember it on this device. */
export async function linkParent(username: string, password: string, remember: boolean) {
  let res
  try {
    res = await putParent(token, username, password)
  } catch (err) {
    if (!(err instanceof ApiError) || err.code !== 'session_expired') throw err
    await relogin() // the main session lapsed: sign in again, then retry once
    res = await putParent(token, username, password)
  }
  accept(res)
  await saveCredentials({ username, password }, remember, 'parent')
  parentLinked = true
  notify()
  await clearCache() // every cached view was for the old session
}

export async function unlinkParent() {
  await clearCredentials('parent')
  parentLinked = false
  try {
    accept(await deleteParent(token))
  } catch {
    notify() // the session ends with the tab anyway; the parent login is already forgotten here
  }
  await clearCache()
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
      await attachParent() // a new student session has no parent attached yet
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
