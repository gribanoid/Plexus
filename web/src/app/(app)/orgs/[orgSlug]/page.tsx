'use client'

import { use, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { Plus, FolderKanban } from 'lucide-react'
import { apiFetch, routes } from '@plexus/api'
import { CreateProjectDialog } from '@plexus/features'

interface Project {
  id: string
  key: string
  name: string
  description?: string | null
}

interface OrgPageProps {
  params: Promise<{ orgSlug: string }>
}

export default function OrgProjectsPage({ params }: OrgPageProps) {
  const { orgSlug } = use(params)
  const router = useRouter()
  const [createOpen, setCreateOpen] = useState(false)

  const { data: org } = useQuery({
    queryKey: ['org', orgSlug],
    queryFn: () => apiFetch<{ name: string; slug: string }>(`/orgs/${orgSlug}`),
  })

  const { data: projects = [], isLoading } = useQuery<Project[]>({
    queryKey: ['projects', orgSlug],
    queryFn: async () => {
      const json = await apiFetch<{ items: Project[] }>(`/orgs/${orgSlug}/projects`)
      return json.items ?? []
    },
  })

  return (
    <div className="h-full overflow-y-auto bg-plexus-surface-subtle">
      <div className="mx-auto max-w-3xl px-6 py-8">
        <div className="mb-6 flex items-start justify-between gap-4">
          <div>
            <button
              type="button"
              onClick={() => router.push(routes.orgs())}
              className="mb-1 text-xs text-plexus-text-subtle hover:text-plexus-brand hover:underline"
            >
              ← All workspaces
            </button>
            <h1 className="text-2xl font-semibold text-plexus-text">{org?.name ?? orgSlug}</h1>
            <p className="mt-1 text-sm text-plexus-text-subtle">Select a project or create a new one</p>
          </div>
          <button
            type="button"
            onClick={() => setCreateOpen(true)}
            className="inline-flex h-9 shrink-0 items-center gap-1.5 rounded bg-plexus-brand px-3 text-sm font-medium text-white transition-colors hover:bg-plexus-brand-hover"
          >
            <Plus className="h-4 w-4" />
            New project
          </button>
        </div>

        {isLoading ? (
          <div className="flex justify-center py-16">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-plexus-brand border-t-transparent" />
          </div>
        ) : projects.length === 0 ? (
          <div className="flex flex-col items-center gap-3 rounded-lg border border-dashed border-plexus-border bg-plexus-surface py-16 text-center">
            <FolderKanban className="h-10 w-10 text-plexus-text-muted" />
            <p className="text-sm text-plexus-text-subtle">No projects yet.</p>
            <button
              type="button"
              onClick={() => setCreateOpen(true)}
              className="inline-flex h-9 items-center gap-1.5 rounded border border-plexus-border bg-plexus-surface px-3 text-sm font-medium text-plexus-text transition-colors hover:bg-black/[0.03] dark:hover:bg-white/[0.05]"
            >
              <Plus className="h-4 w-4" />
              Create project
            </button>
          </div>
        ) : (
          <ul className="divide-y divide-plexus-border overflow-hidden rounded-lg border border-plexus-border bg-plexus-surface shadow-sm">
            {projects.map((project) => (
              <li key={project.id}>
                <button
                  type="button"
                  className="flex w-full items-center gap-4 px-4 py-4 text-left transition-colors hover:bg-black/[0.03] dark:hover:bg-white/[0.05]"
                  onClick={() => router.push(routes.projectBoard(orgSlug, project.key))}
                >
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded bg-plexus-brand/15 text-xs font-bold text-plexus-brand">
                    {project.key}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="font-medium text-plexus-text">{project.name}</p>
                    <p className="truncate text-sm text-plexus-text-subtle">
                      {project.description ?? 'No description'}
                    </p>
                  </div>
                  <span className="shrink-0 text-xs text-plexus-brand">Open board →</span>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      <CreateProjectDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        orgSlug={orgSlug}
        onCreated={(key) => router.push(routes.projectBoard(orgSlug, key))}
      />
    </div>
  )
}
