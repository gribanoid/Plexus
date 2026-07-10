import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiFetch } from '../api-fetch'

export interface Notification {
  id: string
  type: 'assigned' | 'mentioned' | 'commented' | 'status_changed'
  title: string
  body?: string | null
  read: boolean
  issue_id?: string | null
  created_at: string
}

export function useNotifications() {
  return useQuery({
    queryKey: ['notifications'],
    queryFn: async () => {
      const json = await apiFetch<{ items: Notification[] }>('/notifications')
      return json.items ?? []
    },
    staleTime: 15_000,
    refetchInterval: 30_000,
  })
}

export function useMarkNotificationRead() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (notificationId: string) => {
      await apiFetch(`/notifications/${notificationId}/read`, { method: 'POST' })
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
    },
  })
}

export function useMarkAllNotificationsRead() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      await apiFetch('/notifications/read-all', { method: 'POST' })
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
    },
  })
}
