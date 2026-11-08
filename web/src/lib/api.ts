import type { ErrorBody, LoginResponse } from '@/api/api.gen'

// Empty in dev (Vite proxies /api); set VITE_API_BASE when the API lives on another origin.
const BASE = (import.meta.env.VITE_API_BASE as string | undefined) ?? ''

export class ApiError extends Error {
  readonly status: number
  readonly code: string
  constructor(status: number, code: string, message: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

interface AuthHooks {
  token(): string | null
  /** Sign in again with the stored credentials; resolves to the new token. */
  relogin(): Promise<string>
}

let auth: AuthHooks = {
  token: () => null,
  relogin: () => Promise.reject(new ApiError(401, 'session_expired', 'Sign in to continue.')),
}

/** Called once by the auth module, which owns the session. */
export function configureAuth(hooks: AuthHooks) {
  auth = hooks
}

async function toError(res: Response): Promise<ApiError> {
  try {
    const body = (await res.json()) as ErrorBody
    return new ApiError(res.status, body.error.code, body.error.message)
  } catch {
    return new ApiError(res.status, 'internal', `Request failed (${res.status}).`)
  }
}

async function send(path: string, token: string | null, init?: RequestInit): Promise<Response> {
  try {
    return await fetch(BASE + path, {
      ...init,
      // The server gives up on PowerSchool at 45s; past that the network itself has stalled.
      signal: AbortSignal.timeout(60_000),
      headers: { ...init?.headers, ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    })
  } catch (err) {
    if (err instanceof DOMException && err.name === 'TimeoutError') {
      throw new ApiError(0, 'timeout', 'The connection stalled. Try again.')
    }
    throw new ApiError(0, 'offline', 'Can’t reach the server. Check your connection.')
  }
}

/** A 200 that isn't our JSON is usually a Wi-Fi sign-in page answering in our place. */
async function json<T>(res: Response): Promise<T> {
  if (!res.headers.get('Content-Type')?.includes('application/json')) {
    throw new ApiError(0, 'offline', 'The network answered instead of the server. If you’re on school Wi-Fi, sign in to it first.')
  }
  return res.json() as Promise<T>
}

/** GET an authenticated endpoint. An expired session signs in again once, then retries. */
export async function request<T>(path: string, as: 'json' | 'text' = 'json'): Promise<T> {
  let res = await send(path, auth.token())
  if (res.status === 401) {
    const err = await toError(res)
    if (err.code !== 'session_expired') throw err
    res = await send(path, await auth.relogin())
  }
  if (!res.ok) throw await toError(res)
  return as === 'text' ? (res.text() as Promise<T>) : json<T>(res)
}

export async function postSession(username: string, password: string): Promise<LoginResponse> {
  const res = await send('/api/session', null, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
  if (!res.ok) throw await toError(res)
  return json<LoginResponse>(res)
}
