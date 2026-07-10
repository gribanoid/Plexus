import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { apiFetch, completeAuth, loginRequest, routes } from '@plexus/api'

export function LoginPage() {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    try {
      const tokens = await loginRequest(email, password)
      await completeAuth(tokens)
      const orgs = await apiFetch<{ items: { slug: string }[] }>('/orgs')
      const org = orgs.items[0]
      if (!org) {
        navigate(routes.orgs())
        return
      }
      const projects = await apiFetch<{ items: { key: string }[] }>(`/orgs/${org.slug}/projects`)
      const project = projects.items[0]
      navigate(project ? routes.projectBoard(org.slug, project.key) : routes.org(org.slug))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex h-screen items-center justify-center bg-plexus-surface-subtle px-4">
      <div className="w-full max-w-sm space-y-6">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-plexus-text">Plexus</h1>
          <p className="mt-1 text-sm text-plexus-text-subtle">Sign in to continue</p>
        </div>
        <form onSubmit={handleLogin} className="space-y-4">
          <input type="text" placeholder="Email or admin" value={email} onChange={(e) => setEmail(e.target.value)} className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text" required />
          <input type="password" placeholder="Password" value={password} onChange={(e) => setPassword(e.target.value)} className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text" required />
          <button type="submit" disabled={loading} className="h-10 w-full rounded bg-plexus-brand text-sm font-medium text-white hover:bg-plexus-brand-hover disabled:opacity-50">
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
        <p className="text-center text-sm text-plexus-text-subtle">
          No account? <Link to="/register" className="text-plexus-brand hover:underline">Register</Link>
        </p>
      </div>
    </div>
  )
}
