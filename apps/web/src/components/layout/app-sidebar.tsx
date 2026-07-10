'use client'

import { useState } from 'react'
import { usePathname, useParams } from 'next/navigation'
import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import {
  LayoutGrid,
  ListTodo,
  Settings,
  Search,
  ChevronDown,
  Plus,
  LogOut,
  Building2,
  X,
} from 'lucide-react'
import { Avatar, AvatarFallback, Button } from '@plexus/ui'
import { useAuthStore } from '@/lib/stores/auth.store'
import { useRouter } from 'next/navigation'
import { apiFetch, logoutRequest, routes } from '@plexus/api'
import { IssueSearch } from '@plexus/features'

interface Project {
  id: string
  key: string
  name: string
  icon_url?: string | null
}

export function AppSidebar() {
  const pathname = usePathname()
  const params = useParams() as { orgSlug?: string; projectKey?: string }
  const { orgSlug, projectKey } = params
  const router = useRouter()
  const { user } = useAuthStore()
  const [searchOpen, setSearchOpen] = useState(false)

  const { data: projects = [] } = useQuery<Project[]>({
    queryKey: ['projects', orgSlug],
    enabled: !!orgSlug,
    queryFn: async () => {
      const json = await apiFetch<{ items: Project[] }>(`/orgs/${orgSlug}/projects`)
      return json.items ?? []
    },
  })

  function isActive(path: string) {
    return pathname === path || pathname.startsWith(path + '/')
  }

  async function handleLogout() {
    await logoutRequest()
    router.push('/login')
  }

  return (
    <>
      <aside className="flex w-[var(--sidebar-width)] shrink-0 flex-col border-r border-plexus-border bg-plexus-sidebar">
        {orgSlug && (
          <div className="border-b border-plexus-border px-3 py-3">
            <button
              type="button"
              className="flex w-full items-center gap-2 rounded px-1 py-1 text-left hover:bg-black/5 dark:hover:bg-white/5"
              onClick={() => router.push(routes.org(orgSlug!))}
            >
              <Building2 className="h-4 w-4 text-plexus-text-subtle" />
              <span className="flex-1 truncate text-sm font-semibold capitalize text-plexus-text">
                {orgSlug}
              </span>
              <ChevronDown className="h-4 w-4 text-plexus-text-subtle" />
            </button>
          </div>
        )}

        <div className="flex-1 overflow-y-auto p-2">
          <nav className="space-y-0.5">
            <SidebarItem
              href={routes.orgs()}
              icon={<Building2 className="h-4 w-4" />}
              label="Workspaces"
              active={isActive('/orgs')}
            />
            {orgSlug ? (
              <button
                type="button"
                onClick={() => setSearchOpen(true)}
                className="flex w-full items-center gap-2 rounded px-2 py-1.5 text-left text-sm text-plexus-text-subtle transition-colors hover:bg-black/5 hover:text-plexus-text dark:hover:bg-white/5"
              >
                <Search className="h-4 w-4" />
                <span className="truncate">Search</span>
              </button>
            ) : (
              <SidebarItem
                href="#"
                icon={<Search className="h-4 w-4" />}
                label="Search"
                active={false}
              />
            )}
          </nav>

          {orgSlug && (
            <div className="mt-4">
              <div className="mb-1 flex items-center justify-between px-2">
                <span className="text-[11px] font-semibold uppercase tracking-wider text-plexus-text-subtle">
                  Projects
                </span>
                <Button
                  size="icon"
                  variant="ghost"
                  className="h-5 w-5 text-plexus-text-subtle hover:text-plexus-text"
                >
                  <Plus className="h-3.5 w-3.5" />
                </Button>
              </div>

              <nav className="space-y-0.5">
                {projects.map((project) => (
                  <div key={project.id}>
                    <SidebarItem
                      href={routes.projectBoard(orgSlug!, project.key)}
                      icon={
                        <div className="flex h-4 w-4 items-center justify-center rounded bg-[#DEEBFF] text-[9px] font-bold text-plexus-brand dark:bg-plexus-brand/20">
                          {project.key[0]}
                        </div>
                      }
                      label={project.name}
                      active={projectKey === project.key}
                      bold={projectKey === project.key}
                    />

                    {projectKey === project.key && (
                      <div className="ml-5 mt-0.5 space-y-0.5 border-l border-plexus-border pl-2">
                        <SidebarItem
                          href={routes.projectBoard(orgSlug!, project.key)}
                          icon={<LayoutGrid className="h-3.5 w-3.5" />}
                          label="Board"
                          active={isActive(routes.projectBoard(orgSlug!, project.key))}
                          small
                        />
                        <SidebarItem
                          href={routes.projectBacklog(orgSlug!, project.key)}
                          icon={<ListTodo className="h-3.5 w-3.5" />}
                          label="Backlog"
                          active={isActive(routes.projectBacklog(orgSlug!, project.key))}
                          small
                        />
                        <SidebarItem
                          href={routes.projectSettings(orgSlug!, project.key)}
                          icon={<Settings className="h-3.5 w-3.5" />}
                          label="Settings"
                          active={isActive(routes.projectSettings(orgSlug!, project.key))}
                          small
                        />
                      </div>
                    )}
                  </div>
                ))}
              </nav>
            </div>
          )}
        </div>

        <div className="border-t border-plexus-border p-2">
          <div className="flex items-center gap-2 rounded px-2 py-1.5 hover:bg-black/5 dark:hover:bg-white/5">
            <Avatar className="h-7 w-7">
              <AvatarFallback className="bg-plexus-brand text-xs text-white">
                {user?.display_name?.[0]?.toUpperCase() ?? '?'}
              </AvatarFallback>
            </Avatar>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-medium text-plexus-text">
                {user?.display_name ?? 'Account'}
              </p>
              <p className="truncate text-[11px] text-plexus-text-subtle">{user?.email}</p>
            </div>
            <button
              type="button"
              onClick={handleLogout}
              className="text-plexus-text-subtle hover:text-plexus-text"
              title="Sign out"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </aside>

      {searchOpen && orgSlug && (
        <div className="fixed inset-0 z-50 flex items-start justify-center pt-[12vh]">
          <button
            type="button"
            className="absolute inset-0 bg-black/40"
            aria-label="Close search"
            onClick={() => setSearchOpen(false)}
          />
          <div className="relative z-10 w-full max-w-lg rounded-lg border border-plexus-border bg-plexus-surface p-4 shadow-xl">
            <div className="mb-3 flex items-center justify-between">
              <h2 className="text-base font-semibold text-plexus-text">Search issues</h2>
              <button
                type="button"
                onClick={() => setSearchOpen(false)}
                className="rounded p-1 text-plexus-text-subtle hover:bg-black/5 hover:text-plexus-text dark:hover:bg-white/5"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
            <IssueSearch
              orgSlug={orgSlug}
              projectKey={projectKey}
              variant="panel"
              autoFocus
              onIssueSelect={(hit, key) =>
                router.push(routes.issue(orgSlug, key, hit.number))
              }
              onNavigate={() => setSearchOpen(false)}
            />
          </div>
        </div>
      )}
    </>
  )
}

interface SidebarItemProps {
  href: string
  icon: React.ReactNode
  label: string
  active: boolean
  bold?: boolean
  small?: boolean
}

function SidebarItem({ href, icon, label, active, bold, small }: SidebarItemProps) {
  const cls = [
    'flex w-full items-center gap-2 rounded px-2 text-left transition-colors',
    small ? 'py-1 text-xs' : 'py-1.5 text-sm',
    active
      ? 'bg-[#DEEBFF] font-medium text-plexus-brand dark:bg-plexus-brand/15'
      : 'text-plexus-text-subtle hover:bg-black/5 hover:text-plexus-text dark:hover:bg-white/5',
    bold ? 'font-medium' : '',
  ].join(' ')

  return (
    <Link href={href} className={cls}>
      {icon}
      <span className="truncate">{label}</span>
    </Link>
  )
}
