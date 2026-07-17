import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiFetch } from '../api-fetch'

export interface Attachment {
  id: string
  filename: string
  mime_type: string
  size: number
  uploader_id: string
  created_at: string
}

export function useAttachments(orgSlug: string, projectKey: string, issueNumber: string) {
  return useQuery({
    queryKey: ['attachments', orgSlug, projectKey, issueNumber],
    enabled: Boolean(orgSlug && projectKey && issueNumber),
    queryFn: async () => {
      const json = await apiFetch<{ items: Attachment[] }>(
        `/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}/attachments`
      )
      return json.items ?? []
    },
  })
}

export function useUploadAttachment(orgSlug: string, projectKey: string, issueNumber: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (file: File) => {
      const { upload_url, storage_key } = await apiFetch<{
        upload_url: string
        storage_key: string
      }>(`/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}/attachments/upload-url`, {
        method: 'POST',
        body: JSON.stringify({
          filename: file.name,
          mime_type: file.type || 'application/octet-stream',
          size: file.size,
        }),
      })

      const putRes = await fetch(upload_url, {
        method: 'PUT',
        body: file,
        headers: { 'Content-Type': file.type || 'application/octet-stream' },
      })
      if (!putRes.ok) throw new Error('Upload failed')

      await apiFetch(`/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}/attachments`, {
        method: 'POST',
        body: JSON.stringify({
          filename: file.name,
          mime_type: file.type || 'application/octet-stream',
          size: file.size,
          storage_key,
        }),
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['attachments', orgSlug, projectKey, issueNumber] })
    },
  })
}

export function useDeleteAttachment(orgSlug: string, projectKey: string, issueNumber: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (attachmentId: string) => {
      await apiFetch(
        `/orgs/${orgSlug}/projects/${projectKey}/issues/${issueNumber}/attachments/${attachmentId}`,
        { method: 'DELETE' }
      )
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['attachments', orgSlug, projectKey, issueNumber] })
    },
  })
}
