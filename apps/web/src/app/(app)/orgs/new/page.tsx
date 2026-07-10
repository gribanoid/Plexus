'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Button } from '@plexus/ui'
import { apiFetch, routes } from '@plexus/api'

const schema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
  slug: z
    .string()
    .optional()
    .refine(
      (val) => !val || /^[a-z0-9][a-z0-9-]{1,38}[a-z0-9]$/.test(val),
      'Lowercase letters, numbers, hyphens only (3–40 chars)',
    ),
})

type FormData = z.infer<typeof schema>

const inputClass =
  'flex h-10 w-full rounded border border-plexus-border bg-plexus-surface-subtle px-3 text-sm text-plexus-text focus:border-plexus-brand focus:outline-none focus:ring-2 focus:ring-plexus-brand/30'

export default function NewOrgPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<FormData>({ resolver: zodResolver(schema) })

  const name = watch('name')

  async function onSubmit(data: FormData) {
    setLoading(true)
    try {
      const body: { name: string; slug?: string } = { name: data.name }
      if (data.slug) body.slug = data.slug

      const org = await apiFetch<{ slug: string }>('/orgs', {
        method: 'POST',
        body: JSON.stringify(body),
      })
      toast.success('Workspace created')
      router.push(routes.org(org.slug))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to create workspace')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="h-full overflow-y-auto bg-plexus-surface-subtle">
      <div className="mx-auto max-w-lg px-4 py-10">
        <h1 className="text-2xl font-semibold text-plexus-text">Create workspace</h1>
        <p className="mt-1 text-sm text-plexus-text-subtle">
          A workspace is where your team manages projects.
        </p>

        <form onSubmit={handleSubmit(onSubmit)} className="mt-8 space-y-4">
          <div className="space-y-1">
            <label className="text-sm font-medium text-plexus-text" htmlFor="name">
              Workspace name
            </label>
            <input
              id="name"
              className={inputClass}
              placeholder="Acme Inc"
              {...register('name')}
            />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium text-plexus-text" htmlFor="slug">
              URL slug <span className="text-plexus-text-subtle">(optional)</span>
            </label>
            <input
              id="slug"
              className={inputClass}
              placeholder={name ? name.toLowerCase().replace(/\s+/g, '-') : 'acme-inc'}
              {...register('slug')}
            />
            {errors.slug && <p className="text-xs text-destructive">{errors.slug.message}</p>}
          </div>

          <div className="flex gap-2 pt-2">
            <Button
              type="submit"
              className="border-0 bg-plexus-brand text-white shadow-none hover:bg-plexus-brand-hover"
              disabled={loading}
            >
              {loading ? 'Creating…' : 'Create workspace'}
            </Button>
            <Button type="button" variant="outline" onClick={() => router.back()}>
              Cancel
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
