import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '../api-fetch'

export interface SearchIssueHit {
  id: string
  project_id: string
  number: number
  title: string
  description?: string
  priority: string
  assignee_name?: string
  status_name: string
  created_at: number
}

interface UseSearchParams {
  orgSlug: string
  query: string
  projectKey?: string
  enabled?: boolean
}

export function useSearch({ orgSlug, query, projectKey, enabled = true }: UseSearchParams) {
  const trimmed = query.trim()

  return useQuery({
    queryKey: ['search', orgSlug, trimmed, projectKey],
    queryFn: async () => {
      const params = new URLSearchParams({ q: trimmed })
      if (projectKey) params.set('project', projectKey)
      const json = await apiFetch<{ items: SearchIssueHit[]; total: number }>(
        `/orgs/${orgSlug}/search?${params}`,
      )
      return json
    },
    enabled: enabled && trimmed.length >= 2 && Boolean(orgSlug),
    staleTime: 10_000,
  })
}
