import { API_BASE } from './api/setup'

export type WsEventType =
  | 'issue.created'
  | 'issue.updated'
  | 'issue.deleted'
  | 'comment.created'
  | 'sprint.updated'
  | 'notification'

export interface WsEvent {
  type: WsEventType | string
  project_id?: string
  user_id?: string
  payload: unknown
}

function wsBaseUrl(): string {
  const httpBase = API_BASE.replace(/\/api\/v1\/?$/, '')
  return httpBase.replace(/^http/, 'ws')
}

export interface ProjectWsOptions {
  token: string
  projectId: string
  onEvent: (event: WsEvent) => void
  onOpen?: () => void
  onClose?: () => void
  onError?: (error: Event) => void
}

/** Connect to the Plexus WebSocket for a project-scoped event stream. */
export function connectProjectWs({
  token,
  projectId,
  onEvent,
  onOpen,
  onClose,
  onError,
}: ProjectWsOptions): () => void {
  const url = `${wsBaseUrl()}/ws?token=${encodeURIComponent(token)}&project_id=${encodeURIComponent(projectId)}`
  const socket = new WebSocket(url)

  socket.addEventListener('open', () => onOpen?.())
  socket.addEventListener('close', () => onClose?.())
  socket.addEventListener('error', (e) => onError?.(e))
  socket.addEventListener('message', (msg) => {
    try {
      const event = JSON.parse(msg.data as string) as WsEvent
      onEvent(event)
    } catch {
      // ignore malformed frames
    }
  })

  return () => {
    socket.close()
  }
}

export function isIssueEvent(type: string): boolean {
  return type === 'issue.created' || type === 'issue.updated' || type === 'issue.deleted'
}
