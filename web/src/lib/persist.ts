import { createAsyncStoragePersister } from '@tanstack/query-async-storage-persister'
import { QueryClient } from '@tanstack/react-query'
import { del, get, set } from 'idb-keyval'

export const STALE_MS = 10 * 60_000
const DAY_MS = 24 * 60 * 60_000

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: STALE_MS, // refetch on open/focus only when older than this
      gcTime: DAY_MS, // must outlive the persisted cache's maxAge
      retry: (count, err) => count < 2 && !(err instanceof Error && 'status' in err && err.status === 401),
    },
  },
})

/** Student data cached in IndexedDB so the app opens instantly; wiped on sign-out. */
export const persister = createAsyncStoragePersister({
  key: 'psm.query-cache',
  storage: {
    getItem: (k) => get<string>(k).then((v) => v ?? null),
    setItem: (k, v: string) => set(k, v),
    removeItem: (k) => del(k),
  },
})

export const persistOptions = {
  persister,
  maxAge: DAY_MS,
  // The bundle's URL carries its content hash, so a new build drops a cache whose shape it may not read.
  buster: import.meta.url,
  // Report card documents are large and opened rarely: keep them out of the persisted cache.
  dehydrateOptions: { shouldDehydrateQuery: (q: { queryKey: readonly unknown[]; state: { status: string } }) => q.state.status === 'success' && q.queryKey[0] !== 'report-card' },
}

export async function clearCache() {
  queryClient.clear()
  await persister.removeClient()
}
