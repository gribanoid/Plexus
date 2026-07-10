import { useState } from 'react'
import { Outlet, useNavigate, useParams, useLocation } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { LayoutGrid, ListTodo, Settings, Plus, LogOut, ChevronDown } from 'lucide-react'
import { Avatar, AvatarFallback } from '@plexus/ui'
import { useAuthStore } from '../lib/stores/auth.store'
import { apiFetch, logoutRequest, routes } from '@plexus/api'
import { ThemeToggle } from './layout/ThemeToggle'
import { CreateIssueDialog } from '@plexus/features'
import { CreateProjectDialog } from './projects/create-project-dialog'

interface Project {
  id: string
  key: string
  name: string
}

export function AppShell() {
  const navigate = useNavigate()
  const location = useLocation()
  const { orgSlug, projectKey } = useParams()
  const { user } = useAuthStore()
  const [createIssueOpen, setCreateIssueOpen] = useState(false)
  const [createProjectOpen, setCreateProjectOpen] = useState(false)

  const { data: projects = [] } = useQuery<Project[]>({
    queryKey: ['projects', orgSlug],
    enabled: !!orgSlug,
    queryFn: async () => {
      const json = await apiFetch<{ items: Project[] }>(`/orgs/${orgSlug}/projects`)
      return json.items ?? []
    },
  })

  async function handleLogout() {
    await logoutRequest()
    navigate('/login')
  }

  const canCreate = Boolean(orgSlug && projectKey)

  return (
    <>
      <div className="flex h-screen overflow-hidden bg-plexus-surface-subtle">
        <aside className="flex w-[240px] shrink-0 flex-col border-r border-plexus-border bg-plexus-sidebar">
          <div className="titlebar-drag flex h-11 items-center gap-2 border-b border-plexus-border px-4">
            <button
              type="button"
              className="flex items-center gap-1.5 text-sm font-semibold text-plexus-text"
              onClick={() => navigate(routes.orgs())}
            >
              <div className="flex h-5 w-5 items-center justify-center rounded bg-plexus-brand text-[10px] font-bold text-white">P</div>
              Plexus
            </button>
            <ChevronDown className="h-3.5 w-3.5 text-plexus-text-subtle" />
          </div>

          <div className="flex-1 space-y-0.5 overflow-y-auto p-2">
            {orgSlug && (
              <div className="mt-1">
                <div className="mb-1 flex items-center justify-between px-2">
                  <span className="text-[11px] font-semibold uppercase tracking-wider text-plexus-text-subtle">Projects</span>
                  <button
                    type="button"
                    onClick={() => setCreateProjectOpen(true)}
                    className="flex h-5 w-5 items-center justify-center rounded text-plexus-text-subtle hover:bg-black/5 hover:text-plexus-text dark:hover:bg-white/5"
                  >
                    <Plus className="h-3.5 w-3.5" />
                  </button>
                </div>
                {projects.map((p) => (
                  <div key={p.id}>
                    <NavItem
                      icon={<div className="flex h-4 w-4 items-center justify-center rounded bg-plexus-brand/15 text-[9px] font-bold text-plexus-brand">{p.key[0]}</div>}
                      label={p.name}
                      active={p.key === projectKey}
                      onClick={() => navigate(routes.projectBoard(orgSlug, p.key))}
                    />
                    {p.key === projectKey && (
                      <div className="ml-5 mt-0.5 space-y-0.5 border-l border-plexus-border pl-2">
                        <NavItem icon={<LayoutGrid className="h-3.5 w-3.5" />} label="Board" small active={location.pathname.includes('/board')} onClick={() => navigate(routes.projectBoard(orgSlug, p.key))} />
                        <NavItem icon={<ListTodo className="h-3.5 w-3.5" />} label="Backlog" small active={location.pathname.includes('/backlog')} onClick={() => navigate(routes.projectBacklog(orgSlug, p.key))} />
                        <NavItem icon={<Settings className="h-3.5 w-3.5" />} label="Settings" small active={location.pathname.includes('/settings')} onClick={() => navigate(routes.projectSettings(orgSlug, p.key))} />
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="border-t border-plexus-border p-2">
            <div className="flex items-center gap-2 px-2 py-1.5">
              <ThemeToggle />
              <Avatar className="h-7 w-7">
                <AvatarFallback className="bg-plexus-brand text-xs text-white">
                  {user?.display_name?.[0]?.toUpperCase() ?? '?'}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0 flex-1">
                <p className="truncate text-xs font-medium text-plexus-text">{user?.display_name}</p>
              </div>
              <button type="button" onClick={handleLogout} className="text-plexus-text-subtle hover:text-plexus-text">
                <LogOut className="h-4 w-4" />
              </button>
            </div>
          </div>
        </aside>

        <div className="flex min-w-0 flex-1 flex-col overflow-hidden">
          {canCreate && (
            <header className="flex shrink-0 items-center justify-between border-b border-plexus-border bg-plexus-surface px-4 py-2">
              <span className="text-sm text-plexus-text-subtle">{orgSlug} / {projectKey}</span>
              <button
                type="button"
                onClick={() => setCreateIssueOpen(true)}
                className="inline-flex h-8 items-center gap-1.5 rounded bg-plexus-brand px-3 text-sm font-medium text-white hover:bg-plexus-brand-hover"
              >
                <Plus className="h-4 w-4" />
                Create
              </button>
            </header>
          )}
          <main className="min-h-0 flex-1 overflow-hidden">
            <Outlet />
          </main>
        </div>
      </div>

      {orgSlug && (
        <CreateProjectDialog open={createProjectOpen} onClose={() => setCreateProjectOpen(false)} orgSlug={orgSlug} />
      )}
      {canCreate && (
        <CreateIssueDialog open={createIssueOpen} onClose={() => setCreateIssueOpen(false)} orgSlug={orgSlug!} projectKey={projectKey!} />
      )}
    </>
  )
}

function NavItem({ icon, label, active, small, onClick }: {
  icon: React.ReactNode; label: string; active?: boolean; small?: boolean; onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={[
        'flex w-full items-center gap-2 rounded px-2 text-left transition-colors',
        small ? 'py-1 text-xs' : 'py-1.5 text-sm',
        active ? 'bg-plexus-brand/15 font-medium text-plexus-brand' : 'text-plexus-text-subtle hover:bg-black/5 hover:text-plexus-text dark:hover:bg-white/5',
      ].join(' ')}
    >
      {icon}
      <span className="truncate">{label}</span>
    </button>
  )
}
