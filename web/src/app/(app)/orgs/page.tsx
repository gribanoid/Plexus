'use client'

import { useQuery } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { Plus, Building2 } from 'lucide-react'
import { Button } from '@plexus/ui'
import { apiFetch, routes } from '@plexus/api'

interface Org {
  id: string
  slug: string
  name: string
  logo_url?: string | null
  plan: string
  my_role: string
}

export default function OrgsPage() {
  const router = useRouter()

  const { data, isLoading } = useQuery<{ items: Org[] }>({
    queryKey: ['orgs'],
    queryFn: () => apiFetch('/orgs'),
  })

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center bg-plexus-surface-subtle">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-plexus-brand border-t-transparent" />
      </div>
    )
  }

  const orgs = data?.items ?? []

  return (
    <div className="h-full overflow-y-auto bg-plexus-surface-subtle">
      <div className="mx-auto max-w-2xl px-4 py-12">
        <div className="mb-8 flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold text-plexus-text">Workspaces</h1>
            <p className="mt-1 text-sm text-plexus-text-subtle">
              Select a workspace to open or create a new one.
            </p>
          </div>
          <button
            type="button"
            onClick={() => router.push(routes.orgNew())}
            className="inline-flex h-9 shrink-0 items-center gap-1.5 rounded bg-plexus-brand px-3 text-sm font-medium text-white transition-colors hover:bg-plexus-brand-hover"
          >
            <Plus className="h-4 w-4" />
            New workspace
          </button>
        </div>

        {orgs.length === 0 ? (
          <div className="flex flex-col items-center gap-3 rounded-lg border border-dashed border-plexus-border bg-plexus-surface py-16 text-center">
            <Building2 className="h-10 w-10 text-plexus-text-muted" />
            <p className="text-sm text-plexus-text-subtle">No workspaces yet. Create your first one.</p>
            <button
              type="button"
              onClick={() => router.push(routes.orgNew())}
              className="inline-flex h-9 items-center gap-1.5 rounded border border-plexus-border bg-plexus-surface px-3 text-sm font-medium text-plexus-text transition-colors hover:bg-black/[0.03] dark:hover:bg-white/[0.05]"
            >
              <Plus className="h-4 w-4" />
              Create workspace
            </button>
          </div>
        ) : (
          <ul className="divide-y divide-plexus-border overflow-hidden rounded-lg border border-plexus-border bg-plexus-surface shadow-sm">
            {orgs.map((org) => (
              <li key={org.id}>
                <button
                  type="button"
                  className="flex w-full items-center gap-4 px-4 py-3 text-left transition-colors hover:bg-black/[0.03] dark:hover:bg-white/[0.05]"
                  onClick={() => router.push(routes.org(org.slug))}
                >
                  <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded bg-plexus-brand/15 text-sm font-semibold text-plexus-brand">
                    {org.name[0].toUpperCase()}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium text-plexus-text">{org.name}</p>
                    <p className="text-xs capitalize text-plexus-text-subtle">{org.my_role}</p>
                  </div>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
