import { useEffect, useState } from 'react'
import { apiFetch } from '../api-fetch'
import { getApiConfig } from '../auth'
import { connectProjectWs, type WsEvent } from '../ws'

export function useProjectId(orgSlug: string, projectKey: string): string | null {
  const [projectId, setProjectId] = useState<string | null>(null)

  useEffect(() => {
    if (!orgSlug || !projectKey) return
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

/** Subscribe to project WebSocket events using the configured token storage. */
export function useProjectWs(
  orgSlug: string,
  projectKey: string,
  onEvent: (event: WsEvent) => void,
): void {
  const projectId = useProjectId(orgSlug, projectKey)

  useEffect(() => {
    const token = getApiConfig().tokenStorage.getAccessToken()
    if (!token || !projectId) return

    return connectProjectWs({ token, projectId, onEvent })
  }, [orgSlug, projectKey, projectId, onEvent])
}
