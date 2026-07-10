'use client'

import { useEffect, useState } from 'react'
import { apiFetch } from '@plexus/api'
import { connectProjectWs, type WsEvent } from '@/lib/ws'

export function useProjectId(orgSlug: string, projectKey: string): string | null {
  const [projectId, setProjectId] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    apiFetch<{ id: string }>(`/orgs/${orgSlug}/projects/${projectKey}`)
      .then((json) => {
        if (!cancelled && json?.id) setProjectId(json.id)
      })
      .catch(() => {})

    return () => {
      cancelled = true
    }
  }, [orgSlug, projectKey])

  return projectId
}

export function useProjectWs(
  orgSlug: string,
  projectKey: string,
  onEvent: (event: WsEvent) => void
): void {
  const projectId = useProjectId(orgSlug, projectKey)

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    if (!token || !projectId) return

    return connectProjectWs({ token, projectId, onEvent })
  }, [orgSlug, projectKey, projectId, onEvent])
}
