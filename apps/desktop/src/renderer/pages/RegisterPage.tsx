import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { apiFetch, completeAuth, registerRequest, routes } from '@plexus/api'

export function RegisterPage() {
  const navigate = useNavigate()
  const [displayName, setDisplayName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleRegister(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    try {
      const tokens = await registerRequest({ display_name: displayName, email, password })
      await completeAuth(tokens)
      const orgs = await apiFetch<{ items: { slug: string }[] }>('/orgs')
      navigate(orgs.items[0] ? routes.org(orgs.items[0].slug) : routes.orgs())
      toast.success('Account created')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex h-screen items-center justify-center bg-plexus-surface-subtle px-4">
      <div className="w-full max-w-sm space-y-6">
        <h1 className="text-center text-2xl font-bold text-plexus-text">Create account</h1>
        <form onSubmit={handleRegister} className="space-y-4">
          <input placeholder="Display name" value={displayName} onChange={(e) => setDisplayName(e.target.value)} className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text" required />
          <input type="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text" required />
          <input type="password" placeholder="Password" value={password} onChange={(e) => setPassword(e.target.value)} className="flex h-10 w-full rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text" required />
          <button type="submit" disabled={loading} className="h-10 w-full rounded bg-plexus-brand text-sm font-medium text-white hover:bg-plexus-brand-hover">
            {loading ? 'Creating…' : 'Register'}
          </button>
        </form>
        <p className="text-center text-sm text-plexus-text-subtle">
          Have an account? <Link to="/login" className="text-plexus-brand hover:underline">Sign in</Link>
        </p>
      </div>
    </div>
  )
}
