'use client'

import { useEffect, useRef } from 'react'

export const FOCUS_SEARCH_EVENT = 'plexus:focus-search'

export function focusIssueSearch(): void {
  window.dispatchEvent(new CustomEvent(FOCUS_SEARCH_EVENT))
}

function isEditableTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false
  const tag = target.tagName
  return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable
}

export interface BoardShortcutsOptions {
  mode: 'board'
  issues: { id: string; number: number }[]
  selectedIndex: number
  onSelectIndex: (index: number) => void
  onOpenSelected?: () => void
  onCreate?: () => void
}

export interface IssueShortcutsOptions {
  mode: 'issue'
  onEdit?: () => void
}

export interface BacklogShortcutsOptions {
  mode: 'backlog'
  onCreate?: () => void
}

export type KeyboardShortcutsOptions =
  | BoardShortcutsOptions
  | IssueShortcutsOptions
  | BacklogShortcutsOptions

export function useKeyboardShortcuts(options: KeyboardShortcutsOptions): void {
  const optionsRef = useRef(options)
  optionsRef.current = options

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (isEditableTarget(e.target)) return

      if (e.key === '/') {
        e.preventDefault()
        focusIssueSearch()
        return
      }

      const opts = optionsRef.current

      if (opts.mode === 'board') {
        if (e.key === 'j' || e.key === 'J') {
          e.preventDefault()
          if (opts.issues.length === 0) return
          const next = Math.min(opts.selectedIndex + 1, opts.issues.length - 1)
          opts.onSelectIndex(next)
          return
        }
        if (e.key === 'k' || e.key === 'K') {
          e.preventDefault()
          if (opts.issues.length === 0) return
          const prev = Math.max(opts.selectedIndex - 1, 0)
          opts.onSelectIndex(prev)
          return
        }
        if (e.key === 'Enter' && opts.selectedIndex >= 0) {
          e.preventDefault()
          opts.onOpenSelected?.()
          return
        }
        if (e.key === 'c' || e.key === 'C') {
          e.preventDefault()
          opts.onCreate?.()
        }
        return
      }

      if (opts.mode === 'issue' && (e.key === 'e' || e.key === 'E')) {
        e.preventDefault()
        opts.onEdit?.()
        return
      }

      if (opts.mode === 'backlog' && (e.key === 'c' || e.key === 'C')) {
        e.preventDefault()
        opts.onCreate?.()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])
}
