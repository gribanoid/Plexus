import { useEffect, useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { fetchMe } from '@plexus/api'
import { useAuthStore } from '../lib/stores/auth.store'

export function RequireAuth({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate()
  const accessToken = useAuthStore((s) => s.accessToken)
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)
  const logout = useAuthStore((s) => s.logout)
  const [ready, setReady] = useState(() => Boolean(accessToken && user))

  useEffect(() => {
    if (!accessToken) {
      setReady(false)
      return
    }
    if (user) {
      setReady(true)
      return
    }

    let cancelled = false
    setReady(false)
    fetchMe()
      .then((me) => {
        if (cancelled) return
        setUser(me)
        setReady(true)
      })
      .catch(() => {
        if (cancelled) return
        logout()
        navigate('/login', { replace: true })
      })

    return () => {
      cancelled = true
    }
  }, [accessToken, user, setUser, logout, navigate])

  if (!accessToken) {
    return <Navigate to="/login" replace />
  }

  if (!ready) {
    return (
      <div className="flex h-screen items-center justify-center bg-plexus-surface-subtle">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-plexus-brand border-t-transparent" />
      </div>
    )
  }

  return <>{children}</>
}
