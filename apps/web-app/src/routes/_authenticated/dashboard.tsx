import { createFileRoute } from '@tanstack/react-router'

import { Button } from '@/components/ui/button'
import { useAuth } from '@/lib/auth/auth-context'

export const Route = createFileRoute('/_authenticated/dashboard')({
  component: Dashboard,
})

function Dashboard() {
  const { user, logout } = useAuth()

  const displayName =
    user?.firstName || user?.username || user?.email || 'there'

  return (
    <main className="flex min-h-svh flex-col items-center justify-center gap-6 p-6 text-center">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">
          Welcome, {displayName}
        </h1>
        <p className="text-muted-foreground">
          This route is protected — you only see it while signed in.
        </p>
      </div>

      <Button variant="outline" onClick={() => logout()}>
        Sign out
      </Button>
    </main>
  )
}
