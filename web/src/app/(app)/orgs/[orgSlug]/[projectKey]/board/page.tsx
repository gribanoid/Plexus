'use client'

import { use, useCallback, useMemo, useState } from 'react'
import Link from 'next/link'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { DragDropContext, Droppable, Draggable, type DropResult } from '@hello-pangea/dnd'
import { toast } from 'sonner'
import { IssueStatusBadge, PriorityIcon, Avatar, AvatarFallback, Button } from '@plexus/ui'
import { routes, apiFetch } from '@plexus/api'
import { isIssueEvent } from '@/lib/ws'
import { useProjectWs } from '@/lib/use-project-ws'
import { useKeyboardShortcuts } from '@/lib/keyboard-shortcuts'
import { CreateIssueDialog } from '@plexus/features'

interface Status {
  id: string
  name: string
  color: string
  category: 'todo' | 'in_progress' | 'done'
  position: number
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

interface BoardPageProps {
  params: Promise<{ orgSlug: string; projectKey: string }>
}

export default function BoardPage({ params }: BoardPageProps) {
  const { orgSlug, projectKey } = use(params)
  const queryClient = useQueryClient()
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [createOpen, setCreateOpen] = useState(false)

  const assigneeFilter = searchParams.get('assignee') ?? ''
  const priorityFilter = searchParams.get('priority') ?? ''

  const handleWsEvent = useCallback(
    (event: { type: string }) => {
      if (isIssueEvent(event.type)) {
        queryClient.invalidateQueries({ queryKey: ['issues', orgSlug, projectKey] })
      }
    },
    [orgSlug, projectKey, queryClient]
  )

  useProjectWs(orgSlug, projectKey, handleWsEvent)

  const { data: statuses = [] } = useQuery<Status[]>({
    queryKey: ['statuses', orgSlug, projectKey],
    queryFn: async () => {
      const json = await apiFetch<{ items: Status[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/statuses`
      )
      return json.items ?? []
    },
  })

  const { data: issues = [] } = useQuery<Issue[]>({
    queryKey: ['issues', orgSlug, projectKey],
    queryFn: async () => {
      const json = await apiFetch<{ items: Issue[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/issues`
      )
      return json.items ?? []
    },
  })

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
    onError() {
      toast.error('Failed to move issue')
      queryClient.invalidateQueries({ queryKey: ['issues', orgSlug, projectKey] })
    },
  })

  const { data: members = [] } = useQuery<OrgMember[]>({
    queryKey: ['org-members', orgSlug],
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

  const columns = useMemo(() => statuses.filter((status) => status.id), [statuses])

  const issuesByStatus = useMemo(
    () =>
      columns.reduce<Record<string, Issue[]>>((acc, status) => {
        acc[status.id] = filteredIssues.filter((i) => i.status_id === status.id)
        return acc
      }, {}),
    [columns, filteredIssues]
  )

  const flatIssues = useMemo(() => {
    return columns.flatMap((status) => issuesByStatus[status.id] ?? [])
  }, [columns, issuesByStatus])

  useKeyboardShortcuts({
    mode: 'board',
    issues: flatIssues,
    selectedIndex,
    onSelectIndex: setSelectedIndex,
    onOpenSelected: () => {
      const issue = flatIssues[selectedIndex]
      if (issue) {
        router.push(routes.issue(orgSlug, projectKey, issue.number))
      }
    },
    onCreate: () => setCreateOpen(true),
  })

  function setFilter(key: string, value: string) {
    const params = new URLSearchParams(searchParams.toString())
    if (value) params.set(key, value)
    else params.delete(key)
    const qs = params.toString()
    router.replace(qs ? `${pathname}?${qs}` : pathname)
    setSelectedIndex(0)
  }

  function onDragEnd(result: DropResult) {
    const { destination, source, draggableId } = result
    if (!destination) return
    if (destination.droppableId === source.droppableId && destination.index === source.index) return

    const issueId = draggableId
    const issue = issues.find((i) => i.id === issueId)
    if (!issue) return

    const destIssues = (issuesByStatus[destination.droppableId] ?? []).filter((i) => i.id !== issueId)
    const prev = destIssues[destination.index - 1]?.number ?? 0
    const next = destIssues[destination.index]?.number ?? 0
    const newPosition = prev && next ? (prev + next) / 2 : (prev || next || 0) + 65536

    queryClient.setQueryData<Issue[]>(['issues', orgSlug, projectKey], (old = []) =>
      old.map((i) =>
        i.id === issueId ? { ...i, status_id: destination.droppableId } : i
      )
    )

    moveIssue.mutate({
      issueNumber: issue.number,
      statusId: destination.droppableId,
      position: newPosition,
    })
  }

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden bg-plexus-surface-subtle">
      <div className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-plexus-border bg-plexus-surface px-6 py-3">
        <h1 className="text-lg font-semibold text-plexus-text">{projectKey} — Board</h1>
        <div className="flex flex-wrap items-center gap-2">
          <select
            value={assigneeFilter}
            onChange={(e) => setFilter('assignee', e.target.value)}
            className="h-8 min-w-[9.5rem] rounded border border-plexus-border bg-plexus-surface pl-3 pr-8 text-sm text-plexus-text"
            aria-label="Filter by assignee"
          >
            <option value="">All assignees</option>
            <option value="unassigned">Unassigned</option>
            {members.map((member) => (
              <option key={member.id} value={member.id}>
                {member.display_name}
              </option>
            ))}
          </select>
          <select
            value={priorityFilter}
            onChange={(e) => setFilter('priority', e.target.value)}
            className="h-8 min-w-[9.5rem] rounded border border-plexus-border bg-plexus-surface pl-3 pr-8 text-sm text-plexus-text"
            aria-label="Filter by priority"
          >
            <option value="">All priorities</option>
            <option value="urgent">Urgent</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
            <option value="no_priority">No priority</option>
          </select>
          <Button
            size="sm"
            className="border-0 bg-plexus-brand text-white shadow-none hover:bg-plexus-brand-hover"
            onClick={() => setCreateOpen(true)}
          >
            Create issue
          </Button>
        </div>
      </div>

      <DragDropContext onDragEnd={onDragEnd}>
        <div className="flex min-h-0 flex-1 gap-3 overflow-x-auto p-4">
          {columns.map((status) => (
            <div key={status.id} className="flex h-full w-72 shrink-0 flex-col gap-2">
              <div className="flex shrink-0 items-center justify-between rounded-md bg-secondary px-3 py-2">
                <IssueStatusBadge
                  name={status.name}
                  color={status.color}
                  category={status.category}
                />
                <span className="text-xs font-medium tabular-nums text-plexus-text-subtle">
                  {issuesByStatus[status.id]?.length ?? 0}
                </span>
              </div>

              <Droppable droppableId={status.id}>
                {(provided, snapshot) => (
                  <div
                    ref={provided.innerRef}
                    {...provided.droppableProps}
                    className={`min-h-0 flex-1 overflow-y-auto rounded p-1 transition-colors ${
                      snapshot.isDraggingOver
                        ? 'bg-[#DEEBFF] dark:bg-plexus-brand/10'
                        : 'bg-secondary/60'
                    }`}
                  >
                    <div className="flex flex-col gap-1.5">
                      {(issuesByStatus[status.id] ?? []).map((issue, idx) => {
                        const globalIndex = flatIssues.findIndex((i) => i.id === issue.id)
                        const isSelected = globalIndex === selectedIndex
                        return (
                        <Draggable key={issue.id} draggableId={issue.id} index={idx}>
                          {(drag, snap) => (
                            <Link
                              href={routes.issue(orgSlug, projectKey, issue.number)}
                              className={`block cursor-grab rounded border bg-plexus-surface p-3 shadow-sm transition-shadow active:cursor-grabbing ${
                                isSelected
                                  ? 'border-plexus-brand ring-2 ring-plexus-brand/40'
                                  : 'border-plexus-border'
                              } ${
                                snap.isDragging
                                  ? 'shadow-lg ring-2 ring-plexus-brand'
                                  : 'hover:shadow-md'
                              }`}
                              ref={drag.innerRef}
                              {...drag.draggableProps}
                              {...drag.dragHandleProps}
                            >
                              <p className="mb-2 text-sm leading-snug text-plexus-text">
                                {issue.title}
                              </p>
                              <div className="flex items-center justify-between">
                                <div className="flex items-center gap-1.5">
                                  <PriorityIcon priority={issue.priority} />
                                  <span className="text-xs text-plexus-text-subtle">
                                    {projectKey}-{issue.number}
                                  </span>
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
                        )
                      })}
                    </div>
                    {provided.placeholder}
                  </div>
                )}
              </Droppable>
            </div>
          ))}
        </div>
      </DragDropContext>

      <CreateIssueDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        orgSlug={orgSlug}
        projectKey={projectKey}
        onCreated={(issue) => router.push(routes.issue(orgSlug, projectKey, issue.number))}
      />
    </div>
  )
}
