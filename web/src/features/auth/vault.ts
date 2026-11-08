import { del, get, set } from 'idb-keyval'

// The password is encrypted with a non-extractable AES-GCM key kept in IndexedDB. Copying
// the browser's storage elsewhere does not reveal it; script running on this page still
// could, which is why the tab-scoped default exists.

export interface Credentials {
  username: string
  password: string
}

const KEY_ID = 'psm.vault-key'
const BLOB = 'psm.credentials'

function stores(): Storage[] {
  try {
    return [sessionStorage, localStorage]
  } catch {
    return []
  }
}

async function vaultKey(): Promise<CryptoKey> {
  const existing = await get<CryptoKey>(KEY_ID)
  if (existing) return existing
  const key = await crypto.subtle.generateKey({ name: 'AES-GCM', length: 256 }, false, ['encrypt', 'decrypt'])
  await set(KEY_ID, key)
  return key
}

const b64 = (b: Uint8Array) => btoa(String.fromCharCode(...b))
const unb64 = (s: string) => Uint8Array.from(atob(s), (c) => c.charCodeAt(0))

/** Encrypt and store; `remember` persists across browser restarts, otherwise tab-scoped. */
export async function saveCredentials(creds: Credentials, remember: boolean): Promise<void> {
  const iv = crypto.getRandomValues(new Uint8Array(12))
  const data = new TextEncoder().encode(JSON.stringify(creds))
  const ct = new Uint8Array(await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, await vaultKey(), data))
  const [session, local] = stores()
  for (const s of stores()) s.removeItem(BLOB)
  ;(remember ? local : session)?.setItem(BLOB, `${b64(iv)}.${b64(ct)}`)
}

export async function loadCredentials(): Promise<Credentials | null> {
  const blob = stores().map((s) => s.getItem(BLOB)).find(Boolean)
  if (!blob) return null
  try {
    const [iv, ct] = blob.split('.').map(unb64)
    const plain = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, await vaultKey(), ct)
    return JSON.parse(new TextDecoder().decode(plain)) as Credentials
  } catch {
    await clearCredentials() // key gone or blob tampered: unusable
    return null
  }
}

export function hasCredentials(): boolean {
  return stores().some((s) => s.getItem(BLOB) !== null)
}

export async function clearCredentials(): Promise<void> {
  for (const s of stores()) s.removeItem(BLOB)
  await del(KEY_ID)
}
