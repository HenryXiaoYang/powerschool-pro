import 'fake-indexeddb/auto'
import { beforeEach, describe, expect, it } from 'vitest'

class MemoryStorage implements Storage {
  private m = new Map<string, string>()
  get length() { return this.m.size }
  clear() { this.m.clear() }
  getItem(k: string) { return this.m.get(k) ?? null }
  key(i: number) { return [...this.m.keys()][i] ?? null }
  removeItem(k: string) { this.m.delete(k) }
  setItem(k: string, v: string) { this.m.set(k, v) }
}
Object.assign(globalThis, { sessionStorage: new MemoryStorage(), localStorage: new MemoryStorage() })

const { clearCredentials, hasCredentials, loadCredentials, saveCredentials } = await import('./vault')

describe('vault', () => {
  beforeEach(async () => {
    await clearCredentials()
  })

  it('round-trips without storing the password in clear', async () => {
    await saveCredentials({ username: 'alex', password: 'p&ss<word' }, false)
    expect(sessionStorage.getItem('psm.credentials')).not.toContain('p&ss')
    expect(localStorage.getItem('psm.credentials')).toBeNull()
    expect(await loadCredentials()).toEqual({ username: 'alex', password: 'p&ss<word' })
  })

  it('persists only with remember me', async () => {
    await saveCredentials({ username: 'alex', password: 'pw' }, true)
    expect(localStorage.getItem('psm.credentials')).not.toBeNull()
    expect(sessionStorage.getItem('psm.credentials')).toBeNull()
    expect(hasCredentials()).toBe(true)
  })

  it('discards a blob it cannot decrypt', async () => {
    await saveCredentials({ username: 'alex', password: 'pw' }, false)
    const [iv, ct] = sessionStorage.getItem('psm.credentials')!.split('.')
    sessionStorage.setItem('psm.credentials', `${iv}.${ct.slice(0, -4)}AAAA`)
    expect(await loadCredentials()).toBeNull()
    expect(hasCredentials()).toBe(false)
  })
})
