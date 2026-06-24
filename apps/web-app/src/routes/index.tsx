import { createFileRoute, Link } from '@tanstack/react-router'

import { Button, buttonVariants } from '@/components/ui/button'
import { useAuth } from '@/lib/auth/auth-context'

export const Route = createFileRoute('/')({
  validateSearch: (search: Record<string, unknown>): { redirect?: string } => ({
    redirect: typeof search.redirect === 'string' ? search.redirect : undefined,
  }),
  component: Index,
})

function Index() {
  const { isAuthenticated, login } = useAuth()
  const { redirect } = Route.useSearch()

  return (
    <main className="flex min-h-svh flex-col items-center justify-center gap-6 p-6 text-center">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">dots-beacon</h1>
        <p className="max-w-md text-muted-foreground">
          Sign in to access your dashboard.
        </p>
      </div>

      {isAuthenticated ? (
        <Link to="/dashboard" className={buttonVariants({ size: 'lg' })}>
          Go to dashboard
        </Link>
      ) : (
        <Button
          size="lg"
          onClick={() =>
            login({
              redirectUri: window.location.origin + (redirect ?? '/dashboard'),
            })
          }
        >
          Sign in
        </Button>
      )}
    </main>
  )
}
