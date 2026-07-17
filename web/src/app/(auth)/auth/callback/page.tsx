'use client'

import { useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { toast } from 'sonner'
import { completeAuth, routes } from '@plexus/api'

export default function AuthCallbackPage() {
  const router = useRouter()
  const searchParams = useSearchParams()

  useEffect(() => {
    const code = searchParams.get('code')
    // Legacy query tokens are ignored for security (OIDC uses one-time code exchange).
    if (!code) {
      toast.error('SSO login failed')
      router.replace('/login')
      return
    }

    const base =
      typeof window !== 'undefined'
        ? '/api/v1'
        : `${process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'}/api/v1`

    fetch(`${base}/auth/oidc/exchange`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code }),
    })
      .then(async (res) => {
        const json = await res.json().catch(() => ({}))
        if (!res.ok) {
          throw new Error((json as { error?: string }).error ?? 'SSO login failed')
        }
        return json as { access_token: string; refresh_token: string; expires_in: number }
      })
      .then((tokens) => completeAuth(tokens))
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
          project ? routes.projectBoard(org.slug, project.key) : routes.org(org.slug),
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
