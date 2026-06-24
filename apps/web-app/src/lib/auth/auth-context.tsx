import {
  createContext,
  use,
  useEffect,
  useState,
  type ReactNode,
} from 'react'
import type { KeycloakProfile } from 'keycloak-js'

import { initKeycloak, keycloak } from './keycloak'

export interface AuthContextValue {
  /** Whether a valid Keycloak session is currently active. */
  isAuthenticated: boolean
  /** The authenticated user's profile, once loaded. */
  user: KeycloakProfile | null
  /** Current access token (JWT), useful for authorizing API requests. */
  token: string | undefined
  /** Redirect to the Keycloak login page. */
  login: (options?: { redirectUri?: string }) => void
  /** End the Keycloak session and redirect back to the app. */
  logout: (options?: { redirectUri?: string }) => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isInitialized, setIsInitialized] = useState(false)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [user, setUser] = useState<KeycloakProfile | null>(null)
  const [token, setToken] = useState<string | undefined>(undefined)

  useEffect(() => {
    let active = true

    // Keep the access token fresh and reflect lifecycle changes in React state.
    keycloak.onTokenExpired = () => {
      void keycloak.updateToken(30).catch(() => keycloak.logout())
    }
    keycloak.onAuthRefreshSuccess = () => {
      if (active) setToken(keycloak.token)
    }
    keycloak.onAuthLogout = () => {
      if (!active) return
      setIsAuthenticated(false)
      setUser(null)
      setToken(undefined)
    }

    initKeycloak()
      .then(async (authenticated) => {
        if (!active) return
        setIsAuthenticated(authenticated)
        setToken(keycloak.token)
        if (authenticated) {
          const profile = await keycloak.loadUserProfile().catch(() => null)
          if (active) setUser(profile)
        }
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

export function useAuth(): AuthContextValue {
  const context = use(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an <AuthProvider>')
  }
  return context
}

function AuthLoadingScreen() {
  return (
    <div className="flex min-h-svh items-center justify-center text-sm text-muted-foreground">
      Loading…
    </div>
  )
}
