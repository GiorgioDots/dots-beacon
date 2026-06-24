import { useEffect, useState, type ReactNode } from 'react'
import type { KeycloakTokenParsed } from 'keycloak-js'

import { initKeycloak, keycloak } from './keycloak'
import {
  AuthContext,
  type AuthContextValue,
  type UserProfile,
} from './auth-context'

// Standard OIDC profile claims that Keycloak includes in the access token.
type ProfileClaims = KeycloakTokenParsed & {
  preferred_username?: string
  email?: string
  given_name?: string
  family_name?: string
  name?: string
}

// Build the user profile straight from the token claims. This avoids a call to
// the Keycloak Account REST API (`loadUserProfile`), which requires the token
// to carry the `account` audience + `view-profile` role — neither of which this
// client issues (it runs with `fullScopeAllowed: false`).
function readUserProfile(): UserProfile | null {
  const claims = keycloak.tokenParsed as ProfileClaims | undefined
  if (!claims) return null
  return {
    id: claims.sub,
    username: claims.preferred_username,
    email: claims.email,
    firstName: claims.given_name,
    lastName: claims.family_name,
    fullName: claims.name,
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isInitialized, setIsInitialized] = useState(false)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [user, setUser] = useState<UserProfile | null>(null)
  const [token, setToken] = useState<string | undefined>(undefined)

  useEffect(() => {
    let active = true

    // Keep the access token fresh and reflect lifecycle changes in React state.
    keycloak.onTokenExpired = () => {
      void keycloak.updateToken(30).catch(() => keycloak.logout())
    }
    keycloak.onAuthRefreshSuccess = () => {
      if (!active) return
      setToken(keycloak.token)
      setUser(readUserProfile())
    }
    keycloak.onAuthLogout = () => {
      if (!active) return
      setIsAuthenticated(false)
      setUser(null)
      setToken(undefined)
    }

    initKeycloak()
      .then((authenticated) => {
        if (!active) return
        setIsAuthenticated(authenticated)
        setToken(keycloak.token)
        setUser(authenticated ? readUserProfile() : null)
      })
      .finally(() => {
        if (active) setIsInitialized(true)
      })

    return () => {
      active = false
    }
  }, [])

  const value: AuthContextValue = {
    isAuthenticated,
    user,
    token,
    login: (options) =>
      void keycloak.login({
        redirectUri: options?.redirectUri ?? window.location.origin,
      }),
    logout: (options) =>
      void keycloak.logout({
        redirectUri: options?.redirectUri ?? window.location.origin,
      }),
  }

  // Hold rendering until Keycloak has resolved the session. This prevents
  // protected routes from briefly seeing `isAuthenticated === false` and
  // bouncing the user to the landing page on a hard refresh.
  if (!isInitialized) {
    return <AuthLoadingScreen />
  }

  return <AuthContext value={value}>{children}</AuthContext>
}

function AuthLoadingScreen() {
  return (
    <div className="flex min-h-svh items-center justify-center text-sm text-muted-foreground">
      Loading…
    </div>
  )
}
