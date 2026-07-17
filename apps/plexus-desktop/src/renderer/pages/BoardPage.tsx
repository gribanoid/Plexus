import { useEffect, useMemo, useState } from 'react'
import { useParams, Link, useSearchParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { DragDropContext, Droppable, Draggable, type DropResult } from '@hello-pangea/dnd'
import { toast } from 'sonner'
import { IssueStatusBadge, PriorityIcon, Avatar, AvatarFallback } from '@plexus/ui'
import { apiFetch, routes, connectProjectWs, isIssueEvent } from '@plexus/api'
import { useAuthStore } from '../lib/stores/auth.store'

interface Status {
  id: string
  name: string
  color: string
  category: 'todo' | 'in_progress' | 'done'
}

interface Issue {
  id: string
  number: number
  title: string
  priority: 'urgent' | 'high' | 'medium' | 'low' | 'no_priority'
  status_id: string
  assignee_id?: string | null
  story_points?: number | null
}

interface OrgMember {
  id: string
  display_name: string
}

export function BoardPage() {
  const { orgSlug, projectKey } = useParams<{ orgSlug: string; projectKey: string }>()
  const queryClient = useQueryClient()
  const accessToken = useAuthStore((s) => s.accessToken)
  const [projectId, setProjectId] = useState<string | null>(null)
  const [searchParams, setSearchParams] = useSearchParams()

  const assigneeFilter = searchParams.get('assignee') ?? ''
  const priorityFilter = searchParams.get('priority') ?? ''

  useEffect(() => {
    if (!orgSlug || !projectKey) return
    let cancelled = false
    apiFetch<{ id: string }>(`/orgs/${orgSlug}/projects/${projectKey}`)
      .then((json) => {
        if (!cancelled && json?.id) setProjectId(json.id)
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [orgSlug, projectKey])

  useEffect(() => {
    if (!accessToken || !projectId || !orgSlug || !projectKey) return
    return connectProjectWs({
      token: accessToken,
      projectId,
      onEvent(event) {
        if (isIssueEvent(event.type)) {
          queryClient.invalidateQueries({ queryKey: ['issues', orgSlug, projectKey] })
        }
      },
    })
  }, [orgSlug, projectKey, projectId, accessToken, queryClient])

  const { data: statuses = [] } = useQuery<Status[]>({
    queryKey: ['statuses', orgSlug, projectKey],
    enabled: !!orgSlug && !!projectKey,
    queryFn: async () => {
      const json = await apiFetch<{ items: Status[] }>(`/orgs/${orgSlug}/projects/${projectKey}/statuses`)
      return json.items ?? []
    },
  })

  const { data: issues = [] } = useQuery<Issue[]>({
    queryKey: ['issues', orgSlug, projectKey],
    enabled: !!orgSlug && !!projectKey,
    queryFn: async () => {
      const json = await apiFetch<{ items: Issue[] }>(`/orgs/${orgSlug}/projects/${projectKey}/issues`)
      return json.items ?? []
    },
  })

  const { data: members = [] } = useQuery<OrgMember[]>({
    queryKey: ['org-members', orgSlug],
    enabled: !!orgSlug,
    queryFn: async () => {
      const json = await apiFetch<{ items: OrgMember[] }>(`/orgs/${orgSlug}/members`)
      return json.items ?? []
    },
  })

  const filteredIssues = useMemo(() => {
    return issues.filter((issue) => {
      if (assigneeFilter === 'unassigned' && issue.assignee_id) return false
      if (assigneeFilter && assigneeFilter !== 'unassigned' && issue.assignee_id !== assigneeFilter) {
        return false
      }
      if (priorityFilter && issue.priority !== priorityFilter) return false
      return true
    })
  }, [issues, assigneeFilter, priorityFilter])

  const moveIssue = useMutation({
    mutationFn: async ({
      issueNumber,
      statusId,
      position,
    }: {
      issueNumber: number
      statusId: string
      position: number
    }) => {
      await apiFetch(`/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}/move`, {
        method: 'POST',
        body: JSON.stringify({ status_id: statusId, position }),
      })
    },
    onError: () => {
      toast.error('Failed to move issue')
      queryClient.invalidateQueries({ queryKey: ['issues', orgSlug, projectKey] })
    },
  })

  const columns = statuses.filter((s) => s.id)
  const issuesByStatus = columns.reduce<Record<string, Issue[]>>((acc, s) => {
    acc[s.id] = filteredIssues.filter((i) => i.status_id === s.id)
    return acc
  }, {})

  function setFilter(key: 'assignee' | 'priority', value: string) {
    const next = new URLSearchParams(searchParams)
    if (!value) next.delete(key)
    else next.set(key, value)
    setSearchParams(next, { replace: true })
  }

  function onDragEnd(result: DropResult) {
    const { destination, source, draggableId } = result
    if (!destination || !orgSlug || !projectKey) return
    if (destination.droppableId === source.droppableId && destination.index === source.index) return
    const issue = filteredIssues.find((i) => i.id === draggableId)
    if (!issue) return
    const destIssues = issuesByStatus[destination.droppableId] ?? []
    const prev = destIssues[destination.index - 1]?.number ?? 0
    const next = destIssues[destination.index]?.number ?? 0
    const newPosition = prev && next ? (prev + next) / 2 : (prev || next || 0) + 65536
    moveIssue.mutate({ issueNumber: issue.number, statusId: destination.droppableId, position: newPosition })
    queryClient.setQueryData<Issue[]>(['issues', orgSlug, projectKey], (old = []) =>
      old.map((i) => (i.id === draggableId ? { ...i, status_id: destination.droppableId } : i)),
    )
  }

  if (!orgSlug || !projectKey) return null

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden bg-plexus-surface-subtle">
      <div className="flex shrink-0 flex-wrap items-center gap-3 border-b border-plexus-border bg-plexus-surface px-6 py-3">
        <h1 className="text-lg font-semibold text-plexus-text">{projectKey} — Board</h1>
        <select
          className="h-8 rounded border border-plexus-border bg-plexus-surface px-2 text-xs text-plexus-text"
          value={assigneeFilter}
          onChange={(e) => setFilter('assignee', e.target.value)}
        >
          <option value="">All assignees</option>
          <option value="unassigned">Unassigned</option>
          {members.map((m) => (
            <option key={m.id} value={m.id}>
              {m.display_name}
            </option>
          ))}
        </select>
        <select
          className="h-8 rounded border border-plexus-border bg-plexus-surface px-2 text-xs text-plexus-text"
          value={priorityFilter}
          onChange={(e) => setFilter('priority', e.target.value)}
        >
          <option value="">All priorities</option>
          <option value="urgent">Urgent</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
          <option value="no_priority">No priority</option>
        </select>
      </div>
      <DragDropContext onDragEnd={onDragEnd}>
        <div className="flex min-h-0 flex-1 gap-3 overflow-x-auto p-4">
          {columns.map((status) => (
            <div key={status.id} className="flex h-full w-72 shrink-0 flex-col gap-2">
              <div className="flex shrink-0 items-center justify-between rounded-md bg-secondary px-3 py-2">
                <IssueStatusBadge name={status.name} color={status.color} category={status.category} />
                <span className="text-xs text-plexus-text-subtle">
                  {issuesByStatus[status.id]?.length ?? 0}
                </span>
              </div>
              <Droppable droppableId={status.id}>
                {(provided, snapshot) => (
                  <div
                    ref={provided.innerRef}
                    {...provided.droppableProps}
                    className={`min-h-0 flex-1 overflow-y-auto rounded p-1 ${
                      snapshot.isDraggingOver ? 'bg-plexus-brand/10' : 'bg-secondary/60'
                    }`}
                  >
                    <div className="flex flex-col gap-1.5">
                      {(issuesByStatus[status.id] ?? []).map((issue, idx) => (
                        <Draggable key={issue.id} draggableId={issue.id} index={idx}>
                          {(drag, snap) => (
                            <Link
                              to={routes.issue(orgSlug, projectKey, issue.number)}
                              ref={drag.innerRef}
                              {...drag.draggableProps}
                              {...drag.dragHandleProps}
                              className={`block rounded border border-plexus-border bg-plexus-surface p-3 shadow-sm ${
                                snap.isDragging ? 'shadow-lg ring-2 ring-plexus-brand' : 'hover:shadow-md'
                              }`}
                            >
                              <p className="mb-2 text-sm text-plexus-text">{issue.title}</p>
                              <div className="flex items-center justify-between">
                                <div className="flex items-center gap-1.5">
                                  <PriorityIcon priority={issue.priority} />
                                  <span className="text-xs text-plexus-text-subtle">
                                    {projectKey}-{issue.number}
                                  </span>
                                  {issue.story_points != null && (
                                    <span className="rounded bg-secondary px-1 text-[10px] text-plexus-text-subtle">
                                      {issue.story_points}
                                    </span>
                                  )}
                                </div>
                                {issue.assignee_id && (
                                  <Avatar className="h-5 w-5">
                                    <AvatarFallback className="text-[9px]">?</AvatarFallback>
                                  </Avatar>
                                )}
                              </div>
                            </Link>
                          )}
                        </Draggable>
                      ))}
                    </div>
                    {provided.placeholder}
                  </div>
                )}
              </Droppable>
            </div>
          ))}
        </div>
      </DragDropContext>
    </div>
  )
}
