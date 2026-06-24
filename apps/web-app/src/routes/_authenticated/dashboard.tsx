import { createFileRoute } from '@tanstack/react-router'

import { Button } from '@/components/ui/button'
import { useAuth } from '@/lib/auth/auth-context'
import { useListSites } from '@/lib/api/generated/sites/sites'

export const Route = createFileRoute('/_authenticated/dashboard')({
  component: Dashboard,
})

function Dashboard() {
  const { user, logout } = useAuth()
  const { data, isPending, isError, error } = useListSites()

  const displayName =
    user?.firstName || user?.username || user?.email || 'there'

  return (
    <main className="mx-auto flex min-h-svh w-full max-w-2xl flex-col gap-6 p-6">
      <header className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            Welcome, {displayName}
          </h1>
          <p className="text-sm text-muted-foreground">
            This route is protected — you only see it while signed in.
          </p>
        </div>
        <Button variant="outline" onClick={() => logout()}>
          Sign out
        </Button>
      </header>

      <section className="space-y-3">
        <h2 className="text-sm font-medium text-muted-foreground">Sites</h2>
        {isPending ? (
          <p className="text-sm text-muted-foreground">Loading sites…</p>
        ) : isError ? (
          <p className="text-sm text-destructive">
            Failed to load sites: {error.message}
          </p>
        ) : data?.sites && data.sites.length > 0 ? (
          <ul className="divide-y rounded-2xl border">
            {data.sites.map((site) => (
              <li
                key={site.ID}
                className="flex items-center justify-between px-4 py-3 text-sm"
              >
                <span>{site.name}</span>
                <span className="text-muted-foreground">
                  {site.is_on ? 'on' : 'off'}
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-muted-foreground">No sites yet.</p>
        )}
      </section>
    </main>
  )
}
