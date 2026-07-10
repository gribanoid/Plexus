import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { apiFetch, routes } from '@plexus/api'

export function NewOrgPage() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    try {
      const body: { name: string; slug?: string } = { name }
      if (slug) body.slug = slug
      const org = await apiFetch<{ slug: string }>('/orgs', { method: 'POST', body: JSON.stringify(body) })
      toast.success('Workspace created')
      navigate(routes.org(org.slug))
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
        <form onSubmit={handleSubmit} className="mt-8 space-y-4">
          <input placeholder="Workspace name" value={name} onChange={(e) => setName(e.target.value)} className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text" required />
          <input placeholder="URL slug (optional)" value={slug} onChange={(e) => setSlug(e.target.value)} className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text" />
          <div className="flex gap-2">
            <button type="submit" disabled={loading} className="h-9 rounded bg-plexus-brand px-4 text-sm text-white hover:bg-plexus-brand-hover">{loading ? 'Creating…' : 'Create'}</button>
            <button type="button" onClick={() => navigate(-1)} className="h-9 rounded border border-plexus-border px-4 text-sm text-plexus-text">Cancel</button>
          </div>
        </form>
      </div>
    </div>
  )
}
