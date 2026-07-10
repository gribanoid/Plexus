import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { X } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@plexus/ui'
import { apiFetch } from '@plexus/api'

export interface EditIssueDialogProps {
  open: boolean
  onClose: () => void
  orgSlug: string
  projectKey: string
  issueNumber: string
  initial: {
    title: string
    description?: string | null
    status_id: string
    priority: string
    assignee_id?: string | null
  }
}

const inputClass =
  'flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text focus:border-plexus-brand focus:outline-none focus:ring-2 focus:ring-plexus-brand/30'

export function EditIssueDialog({
  open,
  onClose,
  orgSlug,
  projectKey,
  issueNumber,
  initial,
}: EditIssueDialogProps) {
  const queryClient = useQueryClient()
  const [title, setTitle] = useState(initial.title)
  const [description, setDescription] = useState(initial.description ?? '')
  const [statusId, setStatusId] = useState(initial.status_id)
  const [priority, setPriority] = useState(initial.priority)
  const [assigneeId, setAssigneeId] = useState(initial.assignee_id ?? '')

  useEffect(() => {
    if (open) {
      setTitle(initial.title)
      setDescription(initial.description ?? '')
      setStatusId(initial.status_id)
      setPriority(initial.priority)
      setAssigneeId(initial.assignee_id ?? '')
    }
  }, [open, initial])

  const { data: statuses = [] } = useQuery<{ id: string; name: string }[]>({
    queryKey: ['statuses', orgSlug, projectKey],
    queryFn: async () => {
      const json = await apiFetch<{ items: { id: string; name: string }[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/statuses`,
      )
      return json.items ?? []
    },
    enabled: open,
  })

  const { data: members = [] } = useQuery<{ id: string; display_name: string }[]>({
    queryKey: ['org-members', orgSlug],
    queryFn: async () => {
      const json = await apiFetch<{ items: { id: string; display_name: string }[] }>(
        `/orgs/${orgSlug}/members`,
      )
      return json.items ?? []
    },
    enabled: open,
  })

  const save = useMutation({
    mutationFn: async () => {
      await apiFetch(`/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}`, {
        method: 'PATCH',
        body: JSON.stringify({
          title,
          description: description || null,
          status_id: statusId,
          priority,
          assignee_id: assigneeId || null,
        }),
      })
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['issue', orgSlug, projectKey, issueNumber] })
      queryClient.invalidateQueries({ queryKey: ['issues', orgSlug, projectKey] })
      toast.success('Issue updated')
      onClose()
    },
    onError() {
      toast.error('Failed to update issue')
    },
  })

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]">
      <button
        type="button"
        className="absolute inset-0 bg-black/40"
        aria-label="Close dialog"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        className="relative z-10 w-full max-w-lg rounded-lg border border-plexus-border bg-plexus-surface shadow-xl"
      >
        <div className="flex items-center justify-between border-b border-plexus-border px-5 py-4">
          <h2 className="text-base font-semibold text-plexus-text">Edit issue</h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded p-1 text-plexus-text-subtle hover:bg-black/5 hover:text-plexus-text dark:hover:bg-white/5"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="space-y-4 px-5 py-4">
          <div className="space-y-1">
            <label htmlFor="edit-title" className="text-sm font-medium text-plexus-text">
              Title
            </label>
            <input
              id="edit-title"
              className={inputClass}
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <label htmlFor="edit-description" className="text-sm font-medium text-plexus-text">
              Description
            </label>
            <textarea
              id="edit-description"
              rows={4}
              className="flex w-full rounded border border-plexus-border bg-plexus-surface px-3 py-2 text-sm text-plexus-text focus:border-plexus-brand focus:outline-none focus:ring-2 focus:ring-plexus-brand/30"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label htmlFor="edit-status" className="text-sm font-medium text-plexus-text">
                Status
              </label>
              <select
                id="edit-status"
                className={inputClass}
                value={statusId}
                onChange={(e) => setStatusId(e.target.value)}
              >
                {statuses.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-1">
              <label htmlFor="edit-priority" className="text-sm font-medium text-plexus-text">
                Priority
              </label>
              <select
                id="edit-priority"
                className={inputClass}
                value={priority}
                onChange={(e) => setPriority(e.target.value)}
              >
                <option value="urgent">Urgent</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
                <option value="no_priority">No priority</option>
              </select>
            </div>
          </div>
          <div className="space-y-1">
            <label htmlFor="edit-assignee" className="text-sm font-medium text-plexus-text">
              Assignee
            </label>
            <select
              id="edit-assignee"
              className={inputClass}
              value={assigneeId}
              onChange={(e) => setAssigneeId(e.target.value)}
            >
              <option value="">Unassigned</option>
              {members.map((m) => (
                <option key={m.id} value={m.id}>
                  {m.display_name}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="flex justify-end gap-2 border-t border-plexus-border px-5 py-4">
          <Button type="button" variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            type="button"
            disabled={!title.trim() || save.isPending}
            className="border-0 bg-plexus-brand text-white shadow-none hover:bg-plexus-brand-hover"
            onClick={() => save.mutate()}
          >
            {save.isPending ? 'Saving…' : 'Save'}
          </Button>
        </div>
      </div>
    </div>
  )
}
