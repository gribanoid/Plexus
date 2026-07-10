import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { X } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@plexus/ui'
import { useCreateSprint } from '@plexus/api'

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  goal: z.string().optional(),
  start_date: z.string().optional(),
  end_date: z.string().optional(),
})

type FormData = z.infer<typeof schema>

interface CreateSprintDialogProps {
  open: boolean
  onClose: () => void
  orgSlug: string
  projectKey: string
}

export function CreateSprintDialog({
  open,
  onClose,
  orgSlug,
  projectKey,
}: CreateSprintDialogProps) {
  const createSprint = useCreateSprint(orgSlug, projectKey)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', goal: '', start_date: '', end_date: '' },
  })

  useEffect(() => {
    if (!open) return
    reset({ name: '', goal: '', start_date: '', end_date: '' })
  }, [open, reset])

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
      await createSprint.mutateAsync({
        name: data.name,
        ...(data.goal ? { goal: data.goal } : {}),
        ...(data.start_date ? { start_date: new Date(data.start_date).toISOString() } : {}),
        ...(data.end_date ? { end_date: new Date(data.end_date).toISOString() } : {}),
      })
      toast.success('Sprint created')
      onClose()
    } catch {
      toast.error('Failed to create sprint')
    }
  }

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
        aria-labelledby="create-sprint-title"
        className="relative z-10 w-full max-w-md rounded-lg border border-plexus-border bg-plexus-surface shadow-xl"
      >
        <div className="flex items-center justify-between border-b border-plexus-border px-5 py-4">
          <h2 id="create-sprint-title" className="text-base font-semibold text-plexus-text">
            Create sprint
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
            <label htmlFor="sprint-name" className="text-sm font-medium text-plexus-text">
              Name
            </label>
            <input
              id="sprint-name"
              autoFocus
              className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text focus:border-plexus-brand focus:outline-none focus:ring-2 focus:ring-plexus-brand/30"
              placeholder="Sprint 1"
              {...register('name')}
            />
            {errors.name && (
              <p className="text-xs text-[#DE350B]">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-1">
            <label htmlFor="sprint-goal" className="text-sm font-medium text-plexus-text">
              Goal
            </label>
            <textarea
              id="sprint-goal"
              rows={2}
              className="flex w-full resize-none rounded border border-plexus-border bg-plexus-surface px-3 py-2 text-sm text-plexus-text focus:border-plexus-brand focus:outline-none focus:ring-2 focus:ring-plexus-brand/30"
              placeholder="What should this sprint achieve?"
              {...register('goal')}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label htmlFor="sprint-start" className="text-sm font-medium text-plexus-text">
                Start date
              </label>
              <input
                id="sprint-start"
                type="date"
                className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text focus:border-plexus-brand focus:outline-none focus:ring-2 focus:ring-plexus-brand/30"
                {...register('start_date')}
              />
            </div>
            <div className="space-y-1">
              <label htmlFor="sprint-end" className="text-sm font-medium text-plexus-text">
                End date
              </label>
              <input
                id="sprint-end"
                type="date"
                className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text focus:border-plexus-brand focus:outline-none focus:ring-2 focus:ring-plexus-brand/30"
                {...register('end_date')}
              />
            </div>
          </div>

          <div className="flex justify-end gap-2 border-t border-plexus-border pt-4">
            <Button type="button" variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={createSprint.isPending}
              className="border-0 bg-plexus-brand text-white shadow-none hover:bg-plexus-brand-hover"
            >
              {createSprint.isPending ? 'Creating…' : 'Create sprint'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
