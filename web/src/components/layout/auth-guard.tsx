'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/lib/stores/auth.store'
import { fetchMe } from '@plexus/api'

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const { accessToken, user, setUser } = useAuthStore()

  useEffect(() => {
    if (!accessToken) {
      router.replace('/login')
      return
    }
    if (!user) {
      fetchMe()
        .then(setUser)
        .catch(() => router.replace('/login'))
    }
  }, [accessToken, user, router, setUser])

  if (!accessToken) {
    return (
      <div className="flex h-screen items-center justify-center bg-[#FAFBFC]">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-[#0052CC] border-t-transparent" />
      </div>
    )
  }

  return <>{children}</>
}
