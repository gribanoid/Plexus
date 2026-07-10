'use client'

import { use, useCallback, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@plexus/ui'
import { apiFetch, routes } from '@plexus/api'
import {
  CreateIssueDialog,
  CreateSprintDialog,
  SprintSection,
  type SprintIssue,
  type SprintStatus,
} from '@plexus/features'
import { useProjectWs } from '@/lib/use-project-ws'
import { isIssueEvent, isSprintEvent } from '@/lib/ws'
import { useKeyboardShortcuts } from '@/lib/keyboard-shortcuts'

interface Sprint {
  id: string
  name: string
  state: 'active' | 'closed' | 'future'
}

interface BacklogPageProps {
  params: Promise<{ orgSlug: string; projectKey: string }>
}

export default function BacklogPage({ params }: BacklogPageProps) {
  const { orgSlug, projectKey } = use(params)
  const router = useRouter()
  const queryClient = useQueryClient()
  const [collapsedSprints, setCollapsedSprints] = useState<Set<string>>(new Set())
  const [createOpen, setCreateOpen] = useState(false)
  const [createSprintOpen, setCreateSprintOpen] = useState(false)
  const [createSprintId, setCreateSprintId] = useState<string | null>(null)

  const handleWsEvent = useCallback(
    (event: { type: string }) => {
      if (isIssueEvent(event.type)) {
        queryClient.invalidateQueries({ queryKey: ['issues', orgSlug, projectKey] })
      }
      if (isSprintEvent(event.type)) {
        queryClient.invalidateQueries({ queryKey: ['sprints', orgSlug, projectKey] })
      }
    },
    [orgSlug, projectKey, queryClient]
  )

  useProjectWs(orgSlug, projectKey, handleWsEvent)

  const { data: sprints = [] } = useQuery<Sprint[]>({
    queryKey: ['sprints', orgSlug, projectKey],
    queryFn: async () => {
      const json = await apiFetch<{ items: Sprint[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/sprints`,
      )
      return json.items ?? []
    },
  })

  const { data: statuses = [] } = useQuery<SprintStatus[]>({
    queryKey: ['statuses', orgSlug, projectKey],
    queryFn: async () => {
      const json = await apiFetch<{ items: SprintStatus[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/statuses`,
      )
      return json.items ?? []
    },
  })

  const { data: allIssues = [] } = useQuery<SprintIssue[]>({
    queryKey: ['issues', orgSlug, projectKey],
    queryFn: async () => {
      const json = await apiFetch<{ items: SprintIssue[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/issues`,
      )
      return json.items ?? []
    },
  })

  const statusMap = Object.fromEntries(statuses.map((s) => [s.id, s]))

  function toggleSprint(sprintId: string) {
    setCollapsedSprints((prev) => {
      const next = new Set(prev)
      if (next.has(sprintId)) next.delete(sprintId)
      else next.add(sprintId)
      return next
    })
  }

  function openCreate(sprintId: string | null = null) {
    setCreateSprintId(sprintId)
    setCreateOpen(true)
  }

  useKeyboardShortcuts({
    mode: 'backlog',
    onCreate: () => openCreate(null),
  })

  const activeSprint = sprints.find((s) => s.state === 'active')
  const futureSprints = sprints.filter((s) => s.state === 'future')
  const backlogIssues = allIssues.filter((i) => !i.sprint_id)

  const sprintAction = useMutation({
    mutationFn: async ({ sprintId, action }: { sprintId: string; action: 'start' | 'complete' }) => {
      const path =
        action === 'start'
          ? `/orgs/${orgSlug}/projects/${projectKey}/sprints/${sprintId}/start`
          : `/orgs/${orgSlug}/projects/${projectKey}/sprints/${sprintId}/complete`
      await apiFetch(path, { method: 'POST' })
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['sprints', orgSlug, projectKey] })
      queryClient.invalidateQueries({ queryKey: ['issues', orgSlug, projectKey] })
      toast.success('Sprint updated')
    },
    onError() {
      toast.error('Failed to update sprint')
    },
  })

  const sprintSectionProps = {
    statusMap,
    projectKey,
    orgSlug,
    onIssueClick: (issueNumber: number) =>
      router.push(routes.issue(orgSlug, projectKey, issueNumber)),
  }

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden bg-plexus-surface-subtle">
      <div className="flex shrink-0 items-center justify-between border-b border-plexus-border bg-plexus-surface px-6 py-3">
        <h1 className="text-lg font-semibold text-plexus-text">{projectKey} — Backlog</h1>
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            className="border-plexus-border text-plexus-text"
            onClick={() => setCreateSprintOpen(true)}
          >
            <Plus className="mr-1.5 h-4 w-4" />
            Create sprint
          </Button>
          <Button
            size="sm"
            className="border-0 bg-plexus-brand text-white shadow-none hover:bg-plexus-brand-hover"
            onClick={() => openCreate(null)}
          >
            <Plus className="mr-1.5 h-4 w-4" />
            Create issue
          </Button>
        </div>
      </div>

      <div className="flex-1 space-y-4 overflow-y-auto p-4">
        {activeSprint && (
          <SprintSection
            {...sprintSectionProps}
            sprint={activeSprint}
            issues={allIssues.filter((i) => i.sprint_id === activeSprint.id)}
            collapsed={collapsedSprints.has(activeSprint.id)}
            onToggle={() => toggleSprint(activeSprint.id)}
            onAddIssue={() => openCreate(activeSprint.id)}
            onSprintAction={() =>
              sprintAction.mutate({ sprintId: activeSprint.id, action: 'complete' })
            }
            sprintActionLabel="Complete sprint"
          />
        )}

        {futureSprints.map((sprint) => (
          <SprintSection
            key={sprint.id}
            {...sprintSectionProps}
            sprint={sprint}
            issues={allIssues.filter((i) => i.sprint_id === sprint.id)}
            collapsed={collapsedSprints.has(sprint.id)}
            onToggle={() => toggleSprint(sprint.id)}
            onAddIssue={() => openCreate(sprint.id)}
            onSprintAction={() => sprintAction.mutate({ sprintId: sprint.id, action: 'start' })}
            sprintActionLabel="Start sprint"
          />
        ))}

        <SprintSection
          {...sprintSectionProps}
          sprint={{ id: 'backlog', name: 'Backlog', state: 'future' }}
          issues={backlogIssues}
          collapsed={collapsedSprints.has('backlog')}
          onToggle={() => toggleSprint('backlog')}
          onAddIssue={() => openCreate(null)}
          isBacklog
        />
      </div>

      <CreateIssueDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        orgSlug={orgSlug}
        projectKey={projectKey}
        sprintId={createSprintId}
      />

      <CreateSprintDialog
        open={createSprintOpen}
        onClose={() => setCreateSprintOpen(false)}
        orgSlug={orgSlug}
        projectKey={projectKey}
      />
    </div>
  )
}
