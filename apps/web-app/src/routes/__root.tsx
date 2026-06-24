import { createRootRouteWithContext, Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools'

import type { AuthContextValue } from '@/lib/auth/auth-context'

export interface RouterContext {
  auth: AuthContextValue
}

const RootLayout = () => (
  <>
    <Outlet />
    <TanStackRouterDevtools />
  </>
)

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootLayout,
})
