'use client'

import { use, useCallback, useState } from 'react'
import Link from 'next/link'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { format, formatDistanceToNow } from 'date-fns'
import { Send, Paperclip, MessageSquare, Pencil } from 'lucide-react'
import { toast } from 'sonner'
import { PriorityIcon, IssueStatusBadge, Avatar, AvatarFallback, Button } from '@plexus/ui'
import { routes, apiFetch, useAttachments, useUploadAttachment, useDeleteAttachment } from '@plexus/api'

import { EditIssueDialog } from '@plexus/features'
import { useProjectWs } from '@/lib/use-project-ws'
import { isIssueEvent, isCommentEvent } from '@/lib/ws'
import { useKeyboardShortcuts } from '@/lib/keyboard-shortcuts'

interface Issue {
  id: string
  number: number
  title: string
  description?: string | null
  priority: 'urgent' | 'high' | 'medium' | 'low' | 'no_priority'
  status_id: string
  type_id: string
  story_points?: number | null
  assignee_id?: string | null
  assignee_name?: string | null
  reporter_id: string
  reporter_name?: string | null
  due_date?: string | null
  created_at: string
  updated_at: string
}

interface Status {
  id: string
  name: string
  color: string
  category: 'todo' | 'in_progress' | 'done'
}

interface IssueType {
  id: string
  name: string
  color: string
}

interface OrgMember {
  id: string
  display_name: string
  email: string
}

interface IssuePageProps {
  params: Promise<{ orgSlug: string; projectKey: string; issueNumber: string }>
}

const PRIORITY_LABELS: Record<Issue['priority'], string> = {
  urgent: 'Urgent',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
  no_priority: 'No priority',
}

export default function IssuePage({ params }: IssuePageProps) {
  const { orgSlug, projectKey, issueNumber } = use(params)
  const queryClient = useQueryClient()
  const [commentBody, setCommentBody] = useState('')
  const [editOpen, setEditOpen] = useState(false)
  const [editingDescription, setEditingDescription] = useState(false)
  const [descriptionDraft, setDescriptionDraft] = useState('')

  const handleWsEvent = useCallback(
    (event: { type: string }) => {
      if (isIssueEvent(event.type)) {
        queryClient.invalidateQueries({ queryKey: ['issue', orgSlug, projectKey, issueNumber] })
        queryClient.invalidateQueries({ queryKey: ['issues', orgSlug, projectKey] })
        queryClient.invalidateQueries({ queryKey: ['history', orgSlug, projectKey, issueNumber] })
      }
      if (isCommentEvent(event.type)) {
        queryClient.invalidateQueries({ queryKey: ['comments', orgSlug, projectKey, issueNumber] })
      }
    },
    [orgSlug, projectKey, issueNumber, queryClient]
  )

  useProjectWs(orgSlug, projectKey, handleWsEvent)

  useKeyboardShortcuts({
    mode: 'issue',
    onEdit: () => setEditOpen(true),
  })

  const { data: issue, isLoading } = useQuery<Issue>({
    queryKey: ['issue', orgSlug, projectKey, issueNumber],
    queryFn: () =>
      apiFetch<Issue>(`/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}`),
  })

  const { data: statuses = [] } = useQuery<Status[]>({
    queryKey: ['statuses', orgSlug, projectKey],
    queryFn: async () => {
      const json = await apiFetch<{ items: Status[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/statuses`
      )
      return json.items ?? []
    },
  })

  const { data: issueTypes = [] } = useQuery<IssueType[]>({
    queryKey: ['issue-types', orgSlug, projectKey],
    queryFn: async () => {
      const json = await apiFetch<{ items: IssueType[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/issue-types`
      )
      return json.items ?? []
    },
  })

  const { data: comments = [] } = useQuery<
    { id: string; body: string; author_id: string; created_at: string }[]
  >({
    queryKey: ['comments', orgSlug, projectKey, issueNumber],
    queryFn: async () => {
      const json = await apiFetch<{ items: { id: string; body: string; author_id: string; created_at: string }[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}/comments`
      )
      return json.items ?? []
    },
  })

  const { data: members = [] } = useQuery<OrgMember[]>({
    queryKey: ['org-members', orgSlug],
    queryFn: async () => {
      const json = await apiFetch<{ items: OrgMember[] }>(`/orgs/${orgSlug}/members`)
      return json.items ?? []
    },
  })

  const memberNameById = Object.fromEntries(members.map((m) => [m.id, m.display_name]))

  const { data: history = [] } = useQuery<
    { id: string; field: string; old_value?: string; new_value?: string; created_at: string }[]
  >({
    queryKey: ['history', orgSlug, projectKey, issueNumber],
    queryFn: async () => {
      const json = await apiFetch<{
        items: { id: string; field: string; old_value?: string; new_value?: string; created_at: string }[]
      }>(`/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}/history`)
      return json.items ?? []
    },
  })

  const { data: attachments = [] } = useAttachments(orgSlug, projectKey, issueNumber)
  const uploadAttachment = useUploadAttachment(orgSlug, projectKey, issueNumber)
  const deleteAttachment = useDeleteAttachment(orgSlug, projectKey, issueNumber)

  const addComment = useMutation({
    mutationFn: async (body: string) => {
      await apiFetch(`/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}/comments`, {
        method: 'POST',
        body: JSON.stringify({ body }),
      })
    },
    onSuccess() {
      setCommentBody('')
      queryClient.invalidateQueries({ queryKey: ['comments', orgSlug, projectKey, issueNumber] })
    },
    onError() {
      toast.error('Failed to add comment')
    },
  })

  const updateDescription = useMutation({
    mutationFn: async (description: string) => {
      await apiFetch(`/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}`, {
        method: 'PATCH',
        body: JSON.stringify({ description }),
      })
    },
    onSuccess() {
      setEditingDescription(false)
      queryClient.invalidateQueries({ queryKey: ['issue', orgSlug, projectKey, issueNumber] })
      queryClient.invalidateQueries({ queryKey: ['history', orgSlug, projectKey, issueNumber] })
      toast.success('Description saved')
    },
    onError() {
      toast.error('Failed to save description')
    },
  })

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-plexus-brand border-t-transparent" />
      </div>
    )
  }

  if (!issue) {
    return (
      <div className="flex h-full items-center justify-center text-plexus-text-subtle">
        Issue not found
      </div>
    )
  }

  const status = statuses.find((s) => s.id === issue.status_id)
  const issueType = issueTypes.find((t) => t.id === issue.type_id)
  const issueKey = `${projectKey}-${issue.number}`

  return (
    <div className="flex h-full min-h-0 overflow-hidden bg-plexus-surface-subtle">
      {/* Main column */}
      <div className="min-w-0 flex-1 overflow-y-auto">
        <div className="border-b border-plexus-border bg-plexus-surface px-6 py-4">
          <div className="mb-2 flex items-center gap-2 text-sm text-plexus-text-subtle">
            <Link
              href={routes.projectBoard(orgSlug, projectKey)}
              className="hover:text-plexus-brand hover:underline"
            >
              {projectKey}
            </Link>
            <span>/</span>
            <span className="font-medium text-plexus-brand">{issueKey}</span>
          </div>

          <h1 className="text-xl font-semibold leading-snug text-plexus-text sm:text-2xl">
            {issue.title}
          </h1>

          <div className="mt-3 flex flex-wrap items-center gap-2">
            <Button
              size="sm"
              variant="outline"
              className="h-8 gap-1.5 border-plexus-border text-plexus-text"
              onClick={() => setEditOpen(true)}
            >
              <Pencil className="h-3.5 w-3.5" />
              Edit
            </Button>
            <Button size="sm" variant="outline" className="h-8 gap-1.5 border-plexus-border text-plexus-text">
              <MessageSquare className="h-3.5 w-3.5" />
              Comment
            </Button>
            {status && (
              <IssueStatusBadge
                name={status.name}
                color={status.color}
                category={status.category}
              />
            )}
          </div>
        </div>

        <div className="mx-auto max-w-3xl px-6 py-6">
          {/* Details */}
          <section className="mb-8">
            <h2 className="mb-4 text-sm font-semibold text-plexus-text">Details</h2>
            <dl className="grid grid-cols-1 gap-x-10 gap-y-4 sm:grid-cols-2">
              <DetailField label="Type">
                {issueType ? (
                  <span className="flex items-center gap-1.5">
                    <span
                      className="h-2.5 w-2.5 rounded-sm"
                      style={{ backgroundColor: issueType.color }}
                    />
                    {issueType.name}
                  </span>
                ) : (
                  '—'
                )}
              </DetailField>
              <DetailField label="Status">
                {status ? (
                  <IssueStatusBadge
                    name={status.name}
                    color={status.color}
                    category={status.category}
                  />
                ) : (
                  '—'
                )}
              </DetailField>
              <DetailField label="Priority">
                <span className="flex items-center gap-1.5">
                  <PriorityIcon priority={issue.priority} />
                  {PRIORITY_LABELS[issue.priority]}
                </span>
              </DetailField>
              <DetailField label="Assignee">
                {issue.assignee_id ? (
                  <span className="flex items-center gap-1.5">
                    <Avatar className="h-5 w-5">
                      <AvatarFallback className="text-[9px]">
                        {(issue.assignee_name ?? '?').slice(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    {issue.assignee_name ?? 'Assigned'}
                  </span>
                ) : (
                  <span className="text-plexus-text-subtle">Unassigned</span>
                )}
              </DetailField>
              <DetailField label="Story points">
                {issue.story_points != null ? issue.story_points : '—'}
              </DetailField>
              <DetailField label="Due date">
                {issue.due_date ? (
                  format(new Date(issue.due_date), 'dd MMM yyyy')
                ) : (
                  <span className="text-plexus-text-subtle">None</span>
                )}
              </DetailField>
            </dl>
          </section>

          {/* Description */}
          <section className="mb-8">
            <h2 className="mb-3 text-sm font-semibold text-plexus-text">Description</h2>
            {editingDescription ? (
              <div className="space-y-2">
                <textarea
                  autoFocus
                  rows={6}
                  value={descriptionDraft}
                  onChange={(e) => setDescriptionDraft(e.target.value)}
                  className="w-full rounded border border-plexus-border bg-plexus-surface p-3 text-sm leading-relaxed text-plexus-text focus:border-plexus-brand focus:outline-none focus:ring-2 focus:ring-plexus-brand/30"
                  placeholder="Add a description…"
                />
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    className="border-0 bg-plexus-brand text-white hover:bg-plexus-brand-hover"
                    disabled={updateDescription.isPending}
                    onClick={() => updateDescription.mutate(descriptionDraft)}
                  >
                    {updateDescription.isPending ? 'Saving…' : 'Save'}
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => setEditingDescription(false)}
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            ) : issue.description ? (
              <button
                type="button"
                onClick={() => {
                  setDescriptionDraft(issue.description ?? '')
                  setEditingDescription(true)
                }}
                className="w-full rounded border border-plexus-border bg-plexus-surface p-4 text-left text-sm leading-relaxed text-plexus-text transition-colors hover:border-plexus-brand/50"
              >
                {issue.description}
              </button>
            ) : (
              <button
                type="button"
                onClick={() => {
                  setDescriptionDraft('')
                  setEditingDescription(true)
                }}
                className="w-full rounded border border-dashed border-plexus-border bg-plexus-surface px-4 py-8 text-sm text-plexus-text-subtle transition-colors hover:border-plexus-brand hover:text-plexus-text"
              >
                Click to add description
              </button>
            )}
          </section>

          {/* Attachments */}
          <section className="mb-8">
            <div className="mb-3 flex items-center justify-between">
              <h2 className="text-sm font-semibold text-plexus-text">
                Attachments ({attachments.length})
              </h2>
              <label className="cursor-pointer">
                <input
                  type="file"
                  className="hidden"
                  onChange={async (e) => {
                    const file = e.target.files?.[0]
                    if (!file) return
                    try {
                      await uploadAttachment.mutateAsync(file)
                      toast.success('File uploaded')
                    } catch {
                      toast.error('Upload failed')
                    }
                    e.target.value = ''
                  }}
                />
                <Button
                  size="sm"
                  variant="ghost"
                  className="h-7 gap-1 text-xs text-plexus-text-subtle"
                  disabled={uploadAttachment.isPending}
                  asChild
                >
                  <span>
                    <Paperclip className="h-3.5 w-3.5" />
                    {uploadAttachment.isPending ? 'Uploading…' : 'Attach'}
                  </span>
                </Button>
              </label>
            </div>
            {attachments.length === 0 ? (
              <div className="rounded border border-dashed border-plexus-border bg-plexus-surface px-4 py-10 text-center text-sm text-plexus-text-subtle">
                Drop files here or click Attach
              </div>
            ) : (
              <ul className="divide-y rounded border border-plexus-border bg-plexus-surface">
                {attachments.map((a) => (
                  <li key={a.id} className="flex items-center justify-between px-4 py-2.5 text-sm">
                    <span className="truncate text-plexus-text">{a.filename}</span>
                    <div className="flex items-center gap-3 text-xs text-plexus-text-subtle">
                      <span>{(a.size / 1024).toFixed(1)} KB</span>
                      <button
                        type="button"
                        className="text-red-500 hover:underline"
                        onClick={() => deleteAttachment.mutate(a.id)}
                      >
                        Delete
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </section>

          {/* Comments */}
          <section>
            <h2 className="mb-4 text-sm font-semibold text-plexus-text">
              Comments ({comments.length})
            </h2>

            {comments.length > 0 && (
              <div className="mb-4 space-y-4">
                {comments.map((comment) => {
                  const authorName = memberNameById[comment.author_id] ?? 'Unknown user'
                  return (
                  <div key={comment.id} className="flex gap-3">
                    <Avatar className="h-8 w-8 shrink-0">
                      <AvatarFallback className="text-xs">
                        {authorName.slice(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="min-w-0 flex-1">
                      <div className="mb-1 flex items-center gap-2">
                        <span className="text-sm font-medium text-plexus-text">{authorName}</span>
                        <span className="text-xs text-plexus-text-subtle">
                          {formatDistanceToNow(new Date(comment.created_at), { addSuffix: true })}
                        </span>
                      </div>
                      <p className="rounded border border-plexus-border bg-plexus-surface px-3 py-2 text-sm text-plexus-text">
                        {comment.body}
                      </p>
                    </div>
                  </div>
                  )
                })}
              </div>
            )}

            <div className="flex gap-3">
              <Avatar className="h-8 w-8 shrink-0">
                <AvatarFallback className="text-xs">Me</AvatarFallback>
              </Avatar>
              <div className="flex min-w-0 flex-1 items-end gap-2 rounded border border-plexus-border bg-plexus-surface px-3 py-2">
                <textarea
                  className="min-w-0 flex-1 resize-none bg-transparent text-sm text-plexus-text outline-none placeholder:text-plexus-text-muted"
                  placeholder="Add a comment…"
                  rows={2}
                  value={commentBody}
                  onChange={(e) => setCommentBody(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey) && commentBody.trim()) {
                      addComment.mutate(commentBody.trim())
                    }
                  }}
                />
                <Button
                  size="icon"
                  className="h-8 w-8 shrink-0 border-0 bg-plexus-brand text-white hover:bg-plexus-brand-hover"
                  disabled={!commentBody.trim() || addComment.isPending}
                  onClick={() => addComment.mutate(commentBody.trim())}
                >
                  <Send className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
          </section>
        </div>
      </div>

      {/* Right sidebar */}
      <aside className="hidden w-72 shrink-0 overflow-y-auto border-l border-plexus-border bg-plexus-surface px-5 py-6 lg:block">
        <SidebarSection title="People">
          <SidebarRow label="Assignee">
            {issue.assignee_id ? (
              <span className="flex items-center gap-1.5 text-plexus-text">
                <Avatar className="h-5 w-5">
                  <AvatarFallback className="text-[9px]">
                    {(issue.assignee_name ?? '?').slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                {issue.assignee_name ?? 'Assigned'}
              </span>
            ) : (
              <span className="text-plexus-text-subtle">Unassigned</span>
            )}
          </SidebarRow>
          <SidebarRow label="Reporter">
            <span className="flex items-center gap-1.5 text-plexus-text">
              <Avatar className="h-5 w-5">
                <AvatarFallback className="text-[9px]">
                  {(issue.reporter_name ?? '?').slice(0, 2).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              {issue.reporter_name ?? 'Reporter'}
            </span>
          </SidebarRow>
        </SidebarSection>

        <SidebarSection title="Dates">
          <SidebarRow label="Created">
            {format(new Date(issue.created_at), 'dd MMM yyyy, HH:mm')}
          </SidebarRow>
          <SidebarRow label="Updated">
            {formatDistanceToNow(new Date(issue.updated_at), { addSuffix: true })}
          </SidebarRow>
          <SidebarRow label="Due date">
            {issue.due_date ? (
              format(new Date(issue.due_date), 'dd MMM yyyy')
            ) : (
              <span className="text-plexus-text-subtle">None</span>
            )}
          </SidebarRow>
        </SidebarSection>

        <SidebarSection title="Activity">
          {history.length === 0 ? (
            <p className="text-xs text-plexus-text-subtle">No history yet.</p>
          ) : (
            <div className="space-y-3">
              {history.slice(0, 10).map((h) => (
                <div key={h.id} className="text-xs">
                  <p className="text-plexus-text">
                    <span className="font-medium">{h.field}</span> changed
                    {h.new_value && (
                      <>
                        {' '}
                        to <span className="font-medium">{h.new_value}</span>
                      </>
                    )}
                  </p>
                  <p className="mt-0.5 text-plexus-text-subtle">
                    {formatDistanceToNow(new Date(h.created_at), { addSuffix: true })}
                  </p>
                </div>
              ))}
            </div>
          )}
        </SidebarSection>
      </aside>

      <EditIssueDialog
        open={editOpen}
        onClose={() => setEditOpen(false)}
        orgSlug={orgSlug}
        projectKey={projectKey}
        issueNumber={issueNumber}
        initial={{
          title: issue.title,
          description: issue.description,
          status_id: issue.status_id,
          priority: issue.priority,
          assignee_id: issue.assignee_id,
        }}
      />
    </div>
  )
}

function DetailField({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex gap-4 text-sm">
      <dt className="w-28 shrink-0 pt-0.5 text-plexus-text-subtle">{label}</dt>
      <dd className="min-w-0 flex-1 text-plexus-text">{children}</dd>
    </div>
  )
}

function SidebarSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="mb-6">
      <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-plexus-text-subtle">
        {title}
      </h3>
      <div className="space-y-3">{children}</div>
    </div>
  )
}

function SidebarRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <p className="mb-0.5 text-xs text-plexus-text-subtle">{label}</p>
      <div className="text-sm">{children}</div>
    </div>
  )
}
