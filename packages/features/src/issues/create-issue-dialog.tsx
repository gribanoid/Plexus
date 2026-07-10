import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { X } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@plexus/ui'
import { apiFetch, useCreateIssue } from '@plexus/api'

const schema = z.object({
  title: z.string().min(1, 'Title is required'),
  type_id: z.string().uuid('Select an issue type'),
  priority: z.enum(['urgent', 'high', 'medium', 'low', 'no_priority']),
})

type FormData = z.infer<typeof schema>

interface IssueType {
  id: string
  name: string
  color: string
}

export interface CreateIssueDialogProps {
  open: boolean
  onClose: () => void
  orgSlug: string
  projectKey: string
  sprintId?: string | null
  onCreated?: (issue: { id: string; number: number }) => void
}

export function CreateIssueDialog({
  open,
  onClose,
  orgSlug,
  projectKey,
  sprintId,
  onCreated,
}: CreateIssueDialogProps) {
  const createIssue = useCreateIssue(orgSlug, projectKey)
  const [issueTypes, setIssueTypes] = useState<IssueType[]>([])

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { priority: 'medium' },
  })

  useEffect(() => {
    if (!open) return

    async function loadTypes() {
      try {
        const json = await apiFetch<{ items: IssueType[] }>(
          `/orgs/${orgSlug}/projects/${projectKey}/issue-types`,
        )
        setIssueTypes(json.items ?? [])
      } catch {
        setIssueTypes([])
      }
    }

    loadTypes()
    reset({ title: '', type_id: '', priority: 'medium' })
  }, [open, orgSlug, projectKey, reset])

  useEffect(() => {
    if (issueTypes.length > 0) {
      setValue('type_id', issueTypes[0].id)
    }
  }, [issueTypes, setValue])

  useEffect(() => {
    if (!open) return
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [open, onClose])

  if (!open) return null

  async function onSubmit(data: FormData) {
    try {
      const issue = await createIssue.mutateAsync({
        title: data.title,
        type_id: data.type_id,
        priority: data.priority,
        ...(sprintId ? { sprint_id: sprintId } : {}),
      }) as { id: string; number: number }
      const issueKey = `${projectKey}-${issue.number}`
      toast.success(`${issueKey} created`, {
        action: onCreated
          ? {
              label: 'Open issue',
              onClick: () => onCreated(issue),
            }
          : undefined,
      })
      onClose()
    } catch {
      toast.error('Failed to create issue')
    }
  }

  const inputClass =
    'flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text focus:border-plexus-brand focus:outline-none focus:ring-2 focus:ring-plexus-brand/30'

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
        aria-labelledby="create-issue-title"
        className="relative z-10 w-full max-w-md rounded-lg border border-plexus-border bg-plexus-surface shadow-xl"
      >
        <div className="flex items-center justify-between border-b border-plexus-border px-5 py-4">
          <h2 id="create-issue-title" className="text-base font-semibold text-plexus-text">
            Create issue
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded p-1 text-plexus-text-subtle hover:bg-black/5 hover:text-plexus-text dark:hover:bg-white/5"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 px-5 py-4">
          <div className="space-y-1">
            <label htmlFor="issue-title" className="text-sm font-medium text-plexus-text">
              Title
            </label>
            <input
              id="issue-title"
              autoFocus
              className={inputClass}
              placeholder="What needs to be done?"
              {...register('title')}
            />
            {errors.title && (
              <p className="text-xs text-[#DE350B]">{errors.title.message}</p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label htmlFor="issue-type" className="text-sm font-medium text-plexus-text">
                Type
              </label>
              <select id="issue-type" className={inputClass} {...register('type_id')}>
                {issueTypes.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name}
                  </option>
                ))}
              </select>
              {errors.type_id && (
                <p className="text-xs text-[#DE350B]">{errors.type_id.message}</p>
              )}
            </div>

            <div className="space-y-1">
              <label htmlFor="issue-priority" className="text-sm font-medium text-plexus-text">
                Priority
              </label>
              <select id="issue-priority" className={inputClass} {...register('priority')}>
                <option value="urgent">Urgent</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
                <option value="no_priority">No priority</option>
              </select>
            </div>
          </div>

          <div className="flex justify-end gap-2 border-t border-plexus-border pt-4">
            <Button type="button" variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={createIssue.isPending}
              className="border-0 bg-plexus-brand text-white shadow-none hover:bg-plexus-brand-hover"
            >
              {createIssue.isPending ? 'Creating…' : 'Create issue'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
