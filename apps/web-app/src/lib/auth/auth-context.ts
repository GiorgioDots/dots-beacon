import { createContext, use } from 'react'

export interface UserProfile {
  /** Subject (stable user id) from the token. */
  id?: string
  username?: string
  email?: string
  firstName?: string
  lastName?: string
  /** Full display name, if provided by the IdP. */
  fullName?: string
}

export interface AuthContextValue {
  /** Whether a valid Keycloak session is currently active. */
  isAuthenticated: boolean
  /** The authenticated user's profile, derived from the access token claims. */
  user: UserProfile | null
  /** Current access token (JWT), useful for authorizing API requests. */
  token: string | undefined
  /** Redirect to the Keycloak login page. */
  login: (options?: { redirectUri?: string }) => void
  /** End the Keycloak session and redirect back to the app. */
  logout: (options?: { redirectUri?: string }) => void
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth(): AuthContextValue {
  const context = use(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an <AuthProvider>')
  }
  return context
}
