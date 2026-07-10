import * as React from 'react'
import { cn } from '../lib/utils'

export type StatusCategory = 'todo' | 'in_progress' | 'done'

interface IssueStatusBadgeProps {
  name: string
  color: string
  category: StatusCategory
  className?: string
}

export function IssueStatusBadge({ name, color, category, className }: IssueStatusBadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded px-2 py-0.5 text-xs font-semibold uppercase tracking-wide text-plexus-text',
        category === 'done' && 'line-through opacity-70',
        className
      )}
      style={{ backgroundColor: `${color}28` }}
    >
      <span
        className="h-2 w-2 shrink-0 rounded-full"
        style={{ backgroundColor: color }}
        aria-hidden
      />
      {name}
    </span>
  )
}
