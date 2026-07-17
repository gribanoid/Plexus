'use client'

import { useState } from 'react'
import { Plus, HelpCircle } from 'lucide-react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { Avatar, AvatarFallback } from '@plexus/ui'
import { useAuthStore } from '@/lib/stores/auth.store'
import { ThemeToggle } from '@/components/layout/theme-toggle'
import { CreateIssueDialog, IssueSearch, NotificationsPanel } from '@plexus/features'
import { routes } from '@plexus/api'
import { useRouter } from 'next/navigation'

export function AppTopbar() {
  const params = useParams() as { orgSlug?: string; projectKey?: string }
  const router = useRouter()
  const { user } = useAuthStore()
  const [createOpen, setCreateOpen] = useState(false)
  const canCreate = Boolean(params.orgSlug && params.projectKey)

  return (
    <>
      <header className="flex h-12 shrink-0 items-center gap-2 border-b border-plexus-border bg-plexus-topbar px-3 text-white sm:gap-3 sm:px-4">
        <Link href={routes.orgs()} className="flex shrink-0 items-center gap-2 font-semibold">
          <div className="flex h-7 w-7 items-center justify-center rounded bg-white/15 text-sm font-bold">
            P
          </div>
          <span className="hidden sm:inline">Plexus</span>
        </Link>

        {params.orgSlug && params.projectKey && (
          <div className="hidden min-w-0 items-center gap-1 truncate text-sm text-white/80 md:flex">
            <span>/</span>
            <span className="font-medium text-white">{params.orgSlug}</span>
            <span>/</span>
            <span className="font-medium text-white">{params.projectKey}</span>
          </div>
        )}

        {canCreate && (
          <button
            type="button"
            onClick={() => setCreateOpen(true)}
            className="flex h-8 shrink-0 items-center gap-1.5 rounded bg-plexus-brand px-3 text-sm font-medium text-white transition-colors hover:bg-plexus-brand-hover"
          >
            <Plus className="h-4 w-4 shrink-0" />
            <span className="hidden sm:inline">Create</span>
          </button>
        )}

        {params.orgSlug && (
          <div className="mx-auto flex min-w-0 max-w-lg flex-1 px-1 sm:px-2">
            <IssueSearch
              orgSlug={params.orgSlug}
              projectKey={params.projectKey}
              variant="topbar"
              onIssueSelect={(hit, key) =>
                router.push(routes.issue(params.orgSlug!, key, hit.number))
              }
            />
          </div>
        )}

        <div className="flex shrink-0 items-center gap-0.5">
          <ThemeToggle />
          <NotificationsPanel variant="topbar" />
          <button
            type="button"
            className="hidden h-8 w-8 items-center justify-center rounded text-white/80 hover:bg-white/10 hover:text-white sm:flex"
            title="Help"
          >
            <HelpCircle className="h-4 w-4" />
          </button>
          <Avatar className="h-7 w-7 border border-white/30">
            <AvatarFallback className="bg-plexus-brand text-xs text-white">
              {user?.display_name?.[0]?.toUpperCase() ?? '?'}
            </AvatarFallback>
          </Avatar>
        </div>
      </header>

      {canCreate && (
        <CreateIssueDialog
          open={createOpen}
          onClose={() => setCreateOpen(false)}
          orgSlug={params.orgSlug!}
          projectKey={params.projectKey!}
          onCreated={(issue) =>
            router.push(routes.issue(params.orgSlug!, params.projectKey!, issue.number))
          }
        />
      )}
    </>
  )
}
