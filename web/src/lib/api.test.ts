import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, configureAuth, request } from './api'

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })

afterEach(() => vi.unstubAllGlobals())

describe('request', () => {
  it('names a Wi-Fi sign-in page instead of failing to parse it', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('<html>Welcome to SchoolNet</html>', { headers: { 'Content-Type': 'text/html' } })))
    const err = await request('/api/student').catch((e: unknown) => e)
    expect(err).toBeInstanceOf(ApiError)
    expect((err as ApiError).message).toMatch(/Wi-Fi/)
  })

  it('signs in again once on an expired session, then retries', async () => {
    const relogin = vi.fn(async () => 'fresh')
    configureAuth({ token: () => 'stale', relogin })
    const fetch = vi.fn(async (_: string, init?: RequestInit) =>
      (init?.headers as Record<string, string>).Authorization === 'Bearer fresh'
        ? json(200, { ok: true })
        : json(401, { error: { code: 'session_expired', message: 'expired' } }),
    )
    vi.stubGlobal('fetch', fetch)
    await expect(request('/api/student')).resolves.toEqual({ ok: true })
    expect(relogin).toHaveBeenCalledOnce()
    expect(fetch).toHaveBeenCalledTimes(2)
  })
})
