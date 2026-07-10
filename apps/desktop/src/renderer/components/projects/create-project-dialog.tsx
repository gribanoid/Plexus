import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { X } from 'lucide-react'
import { toast } from 'sonner'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Button } from '@plexus/ui'
import { apiFetch, routes } from '@plexus/api'

const schema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
  key: z.string().optional().refine((val) => !val || /^[A-Z][A-Z0-9]{1,9}$/.test(val), 'Invalid key'),
  description: z.string().optional(),
})

type FormData = z.infer<typeof schema>

interface CreateProjectDialogProps {
  open: boolean
  onClose: () => void
  orgSlug: string
  onCreated?: (projectKey: string) => void
}

function suggestKey(name: string): string {
  const words = name.trim().split(/\s+/)
  let key = ''
  for (const word of words) {
    if (key.length >= 4) break
    const ch = word.match(/[a-zA-Z0-9]/)?.[0]
    if (ch) key += ch.toUpperCase()
  }
  return key || 'PRJ'
}

export function CreateProjectDialog({ open, onClose, orgSlug, onCreated }: CreateProjectDialogProps) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { register, handleSubmit, reset, watch, setValue, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
  })
  const name = watch('name')

  useEffect(() => {
    if (!open) return
    reset({ name: '', key: '', description: '' })
  }, [open, reset])

  useEffect(() => {
    if (name && name.length >= 2) setValue('key', suggestKey(name))
  }, [name, setValue])

  const createProject = useMutation({
    mutationFn: async (data: FormData) => {
      const body: { name: string; key?: string; description?: string } = { name: data.name }
      if (data.key) body.key = data.key
      if (data.description) body.description = data.description
      return apiFetch<{ key: string }>(`/orgs/${orgSlug}/projects`, {
        method: 'POST',
        body: JSON.stringify(body),
      })
    },
    onSuccess(project) {
      queryClient.invalidateQueries({ queryKey: ['projects', orgSlug] })
      toast.success('Project created')
      onClose()
      if (onCreated) onCreated(project.key)
      else navigate(routes.projectBoard(orgSlug, project.key))
    },
    onError(err) {
      toast.error(err instanceof Error ? err.message : 'Failed to create project')
    },
  })

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]">
      <button type="button" className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative z-10 w-full max-w-md rounded-lg border border-plexus-border bg-plexus-surface shadow-xl">
        <div className="flex items-center justify-between border-b border-plexus-border px-5 py-4">
          <h2 className="text-base font-semibold text-plexus-text">Create project</h2>
          <button type="button" onClick={onClose}><X className="h-4 w-4" /></button>
        </div>
        <form onSubmit={handleSubmit((d) => createProject.mutate(d))} className="space-y-4 px-5 py-4">
          <div className="space-y-1">
            <label className="text-sm font-medium text-plexus-text">Project name</label>
            <input autoFocus className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface-subtle px-3 text-sm text-plexus-text" {...register('name')} />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium text-plexus-text">Key</label>
            <input className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface-subtle px-3 font-mono text-sm uppercase text-plexus-text" {...register('key')} />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium text-plexus-text">Description</label>
            <textarea rows={2} className="flex w-full resize-none rounded border border-plexus-border bg-plexus-surface-subtle px-3 py-2 text-sm text-plexus-text" {...register('description')} />
          </div>
          <div className="flex justify-end gap-2 border-t border-plexus-border pt-4">
            <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
            <Button type="submit" disabled={createProject.isPending} className="border-0 bg-plexus-brand text-white hover:bg-plexus-brand-hover">
              {createProject.isPending ? 'Creating…' : 'Create project'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
