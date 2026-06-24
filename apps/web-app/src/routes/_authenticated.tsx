import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'

/**
 * Pathless layout route that guards every route nested beneath it
 * (`src/routes/_authenticated/*`). Anonymous visitors are bounced to the
 * landing page, with their intended destination preserved in `redirect` so we
 * can return them there after a successful login.
 */
export const Route = createFileRoute('/_authenticated')({
  beforeLoad: ({ context, location }) => {
    if (!context.auth.isAuthenticated) {
      throw redirect({
        to: '/',
        search: { redirect: location.href },
      })
    }
  },
  component: AuthenticatedLayout,
})

function AuthenticatedLayout() {
  return <Outlet />
}
