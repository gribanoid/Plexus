import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiFetch } from '../api-fetch'

interface ListIssuesParams {
  orgSlug: string
  projectKey: string
  statusId?: string
  assigneeId?: string
  sprintId?: string
  priority?: string
}

export function useIssues(params: ListIssuesParams) {
  return useQuery({
    queryKey: ['issues', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams()
      if (params.statusId) searchParams.set('status_id', params.statusId)
      if (params.assigneeId) searchParams.set('assignee_id', params.assigneeId)
      if (params.sprintId) searchParams.set('sprint_id', params.sprintId)
      if (params.priority) searchParams.set('priority', params.priority)
      const qs = searchParams.toString()
      const path = `/orgs/${params.orgSlug}/projects/${params.projectKey}/issues${qs ? `?${qs}` : ''}`
      const json = await apiFetch<{ items: unknown[] }>(path)
      return json.items ?? []
    },
    staleTime: 30_000,
  })
}

export function useCreateIssue(orgSlug: string, projectKey: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (body: {
      title: string
      type_id: string
      priority?: string
      description?: string
      assignee_id?: string
      sprint_id?: string
    }) => {
      return apiFetch(`/orgs/${orgSlug}/projects/${projectKey}/issues`, {
        method: 'POST',
        body: JSON.stringify(body),
      })
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['issues'] })
    },
  })
}

export function useMoveIssue(orgSlug: string, projectKey: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      issueNumber,
      statusId,
      position,
      sprintId,
    }: {
      issueNumber: number
      statusId?: string
      position?: number
      sprintId?: string
    }) => {
      await apiFetch(
        `/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}/move`,
        {
          method: 'POST',
          body: JSON.stringify({ status_id: statusId, position, sprint_id: sprintId }),
        },
      )
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['issues'] })
    },
  })
}
