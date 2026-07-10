import { useEffect, useRef, useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { Bell, CheckCheck } from 'lucide-react'
import {
  useNotifications,
  useMarkNotificationRead,
  useMarkAllNotificationsRead,
} from '@plexus/api'

export function NotificationsPanel() {
  const [open, setOpen] = useState(false)
  const panelRef = useRef<HTMLDivElement>(null)

  const { data: notifications = [] } = useNotifications()
  const markRead = useMarkNotificationRead()
  const markAllRead = useMarkAllNotificationsRead()

  const unreadCount = notifications.filter((n) => !n.read).length

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  function handleNotificationClick(id: string, read: boolean) {
    if (!read) markRead.mutate(id)
  }

  return (
    <div ref={panelRef} className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="relative flex h-8 w-8 items-center justify-center rounded text-white/80 hover:bg-white/10 hover:text-white"
        title="Notifications"
      >
        <Bell className="h-4 w-4" />
        {unreadCount > 0 && (
          <span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-[#DE350B] px-1 text-[10px] font-bold text-white">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 top-full z-50 mt-1 w-80 overflow-hidden rounded border border-plexus-border bg-plexus-surface shadow-lg">
          <div className="flex items-center justify-between border-b border-plexus-border px-3 py-2.5">
            <span className="text-sm font-semibold text-plexus-text">Notifications</span>
            {unreadCount > 0 && (
              <button
                type="button"
                onClick={() => markAllRead.mutate()}
                disabled={markAllRead.isPending}
                className="flex items-center gap-1 text-xs text-plexus-brand hover:underline disabled:opacity-50"
              >
                <CheckCheck className="h-3.5 w-3.5" />
                Mark all read
              </button>
            )}
          </div>

          <div className="max-h-96 overflow-y-auto">
            {notifications.length === 0 ? (
              <p className="px-3 py-8 text-center text-sm text-plexus-text-subtle">
                No notifications
              </p>
            ) : (
              <ul>
                {notifications.map((n) => (
                  <li key={n.id}>
                    <button
                      type="button"
                      onClick={() => handleNotificationClick(n.id, n.read)}
                      className={[
                        'flex w-full flex-col gap-0.5 border-b border-plexus-border px-3 py-2.5 text-left transition-colors last:border-0 hover:bg-black/5 dark:hover:bg-white/5',
                        !n.read ? 'bg-plexus-brand/5' : '',
                      ].join(' ')}
                    >
                      <span className="text-sm font-medium text-plexus-text">{n.title}</span>
                      {n.body && (
                        <span className="line-clamp-2 text-xs text-plexus-text-subtle">{n.body}</span>
                      )}
                      <span className="text-[11px] text-plexus-text-muted">
                        {formatDistanceToNow(new Date(n.created_at), { addSuffix: true })}
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

/** @alias NotificationsPanel */
export const NotificationBell = NotificationsPanel
