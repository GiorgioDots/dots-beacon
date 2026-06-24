import Keycloak from 'keycloak-js'

export const keycloak = new Keycloak({
  url: import.meta.env.VITE_KEYCLOAK_URL,
  realm: import.meta.env.VITE_KEYCLOAK_REALM,
  clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID,
})

let initPromise: Promise<boolean> | undefined

/**
 * Initialise Keycloak exactly once and return the resulting auth state.
 *
 * `check-sso` verifies an existing session without forcing a login, which keeps
 * public routes (e.g. the landing page) reachable for anonymous visitors. The
 * module-level promise guards against React StrictMode's double-invoked effects,
 * since `keycloak.init()` may only be called a single time per instance.
 */
export function initKeycloak(): Promise<boolean> {
  if (!initPromise) {
    initPromise = keycloak.init({
      onLoad: 'check-sso',
      silentCheckSsoRedirectUri: `${window.location.origin}/silent-check-sso.html`,
      pkceMethod: 'S256',
    })
  }
  return initPromise
}
