'use client'

import { useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { toast } from 'sonner'
import { completeAuth, routes } from '@plexus/api'

export default function AuthCallbackPage() {
  const router = useRouter()
  const searchParams = useSearchParams()

  useEffect(() => {
    const accessToken = searchParams.get('access_token')
    const refreshToken = searchParams.get('refresh_token')

    if (!accessToken || !refreshToken) {
      toast.error('SSO login failed')
      router.replace('/login')
      return
    }

    completeAuth({
      access_token: accessToken,
      refresh_token: refreshToken,
      expires_in: 3600,
    })
      .then(async () => {
        const { apiFetch } = await import('@plexus/api')
        const orgs = await apiFetch<{ items: { slug: string }[] }>('/orgs')
        const org = orgs.items[0]
        if (!org) {
          router.replace(routes.orgs())
          return
        }
        const projects = await apiFetch<{ items: { key: string }[] }>(`/orgs/${org.slug}/projects`)
        const project = projects.items[0]
        router.replace(
          project ? routes.projectBoard(org.slug, project.key) : routes.org(org.slug)
        )
      })
      .catch(() => {
        toast.error('SSO login failed')
        router.replace('/login')
      })
  }, [router, searchParams])

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#FAFBFC] text-sm text-[#5E6C84]">
      Completing sign in…
    </div>
  )
}
