import { MutationCache, QueryClient } from '@tanstack/react-query'

/**
 * Shared React Query client. Exported so non-component code (e.g. route
 * loaders, the router context, or imperative helpers) can reuse the same cache.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60_000,
      retry: 1,
    },
  },
  // Invalidate every query after any successful mutation so the UI always
  // reflects the latest server state. (`queryClient` is referenced lazily here,
  // so it is fully initialised by the time a mutation resolves.)
  mutationCache: new MutationCache({
    onSuccess: () => {
      void queryClient.invalidateQueries()
    },
  }),
})
