import { useState, type ReactNode } from 'react'
import { useParams, Link } from 'react-router-dom'
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

interface Sprint {
  id: string
  name: string
  state: 'active' | 'closed' | 'future'
}

export function BacklogPage() {
  const { orgSlug, projectKey } = useParams<{ orgSlug: string; projectKey: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())
  const [createOpen, setCreateOpen] = useState(false)
  const [createSprintOpen, setCreateSprintOpen] = useState(false)
  const [createSprintId, setCreateSprintId] = useState<string | null>(null)

  const { data: sprints = [] } = useQuery<Sprint[]>({
    queryKey: ['sprints', orgSlug, projectKey],
    enabled: !!orgSlug && !!projectKey,
    queryFn: async () => {
      const json = await apiFetch<{ items: Sprint[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/sprints`,
      )
      return json.items ?? []
    },
  })

  const { data: statuses = [] } = useQuery<SprintStatus[]>({
    queryKey: ['statuses', orgSlug, projectKey],
    enabled: !!orgSlug && !!projectKey,
    queryFn: async () => {
      const json = await apiFetch<{ items: SprintStatus[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/statuses`,
      )
      return json.items ?? []
    },
  })

  const { data: allIssues = [] } = useQuery<SprintIssue[]>({
    queryKey: ['issues', orgSlug, projectKey],
    enabled: !!orgSlug && !!projectKey,
    queryFn: async () => {
      const json = await apiFetch<{ items: SprintIssue[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/issues`,
      )
      return json.items ?? []
    },
  })

  if (!orgSlug || !projectKey) return null

  const statusMap = Object.fromEntries(statuses.map((s) => [s.id, s]))
  const activeSprint = sprints.find((s) => s.state === 'active')
  const futureSprints = sprints.filter((s) => s.state === 'future')
  const backlogIssues = allIssues.filter((i) => !i.sprint_id)

  function toggleSprint(sprintId: string) {
    setCollapsed((prev) => {
      const next = new Set(prev)
      if (next.has(sprintId)) next.delete(sprintId)
      else next.add(sprintId)
      return next
    })
  }

  function openCreate(sprintId: string | null) {
    setCreateSprintId(sprintId)
    setCreateOpen(true)
  }

  const sprintAction = useMutation({
    mutationFn: async ({ sprintId, action }: { sprintId: string; action: 'start' | 'complete' }) => {
      const path =
        action === 'start'
          ? `/orgs/${orgSlug}/projects/${projectKey}/sprints/${sprintId}/start`
          : `/orgs/${orgSlug}/projects/${projectKey}/sprints/${sprintId}/complete`
      await apiFetch(path, { method: 'POST' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sprints', orgSlug, projectKey] })
      queryClient.invalidateQueries({ queryKey: ['issues', orgSlug, projectKey] })
      toast.success('Sprint updated')
    },
    onError: () => toast.error('Failed to update sprint'),
  })

  const sprintSectionProps = {
    statusMap,
    projectKey,
    orgSlug,
    renderIssueRow: (issue: SprintIssue, children: ReactNode) => (
      <Link
        to={routes.issue(orgSlug, projectKey, issue.number)}
        className="block"
      >
        {children}
      </Link>
    ),
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
            collapsed={collapsed.has(activeSprint.id)}
            onToggle={() => toggleSprint(activeSprint.id)}
            onAddIssue={() => openCreate(activeSprint.id)}
            onSprintAction={() =>
              sprintAction.mutate({ sprintId: activeSprint.id, action: 'complete' })
            }
            sprintActionLabel="Complete sprint"
          />
        )}
        {futureSprints.map((s) => (
          <SprintSection
            key={s.id}
            {...sprintSectionProps}
            sprint={s}
            issues={allIssues.filter((i) => i.sprint_id === s.id)}
            collapsed={collapsed.has(s.id)}
            onToggle={() => toggleSprint(s.id)}
            onAddIssue={() => openCreate(s.id)}
            onSprintAction={() => sprintAction.mutate({ sprintId: s.id, action: 'start' })}
            sprintActionLabel="Start sprint"
          />
        ))}
        <SprintSection
          {...sprintSectionProps}
          sprint={{ id: 'backlog', name: 'Backlog', state: 'future' }}
          issues={backlogIssues}
          collapsed={collapsed.has('backlog')}
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
