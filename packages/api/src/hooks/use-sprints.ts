import { useMutation, useQueryClient } from '@tanstack/react-query'
import { apiFetch } from '../api-fetch'

export function useCreateSprint(orgSlug: string, projectKey: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (body: {
      name: string
      goal?: string
      start_date?: string
      end_date?: string
    }) => {
      return apiFetch(`/orgs/${orgSlug}/projects/${projectKey}/sprints`, {
        method: 'POST',
        body: JSON.stringify(body),
      })
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['sprints', orgSlug, projectKey] })
    },
  })
}
