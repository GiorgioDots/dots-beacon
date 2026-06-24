import Axios, { type AxiosError, type AxiosRequestConfig } from 'axios'

import { keycloak } from '@/lib/auth/keycloak'

export const AXIOS_INSTANCE = Axios.create({
  baseURL: import.meta.env.VITE_API_URL,
})

// Attach the current Keycloak access token to every request, refreshing it
// first if it is about to expire. Anonymous requests pass through untouched.
AXIOS_INSTANCE.interceptors.request.use(async (config) => {
  if (keycloak.authenticated) {
    try {
      await keycloak.updateToken(30)
    } catch {
      // Refresh failed (session expired) — bounce to login.
      keycloak.login()
    }
    if (keycloak.token) {
      config.headers.set('Authorization', `Bearer ${keycloak.token}`)
    }
  }
  return config
})

/**
 * Custom mutator used by the Orval-generated client. Returns the response body
 * directly and exposes `.cancel()` so React Query can abort in-flight requests.
 */
export const customInstance = <T>(
  config: AxiosRequestConfig,
  options?: AxiosRequestConfig,
): Promise<T> => {
  const source = Axios.CancelToken.source()
  const promise = AXIOS_INSTANCE({
    ...config,
    ...options,
    cancelToken: source.token,
  }).then(({ data }) => data)

  // @ts-expect-error -- `cancel` is read by React Query for query cancellation.
  promise.cancel = () => {
    source.cancel('Query was cancelled')
  }

  return promise
}

export type ErrorType<Error> = AxiosError<Error>
export type BodyType<BodyData> = BodyData
