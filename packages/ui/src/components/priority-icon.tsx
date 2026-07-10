import * as React from 'react'
import { AlertCircle, ArrowUp, ArrowRight, ArrowDown, Minus } from 'lucide-react'
import { cn } from '../lib/utils'

export type Priority = 'urgent' | 'high' | 'medium' | 'low' | 'no_priority'

interface PriorityIconProps {
  priority: Priority
  className?: string
}

const priorityConfig: Record<
  Priority,
  { icon: React.ComponentType<{ className?: string }>; label: string; color: string }
> = {
  urgent: { icon: AlertCircle, label: 'Urgent', color: 'text-red-500' },
  high: { icon: ArrowUp, label: 'High', color: 'text-orange-500' },
  medium: { icon: ArrowRight, label: 'Medium', color: 'text-yellow-500' },
  low: { icon: ArrowDown, label: 'Low', color: 'text-blue-500' },
  no_priority: { icon: Minus, label: 'No Priority', color: 'text-muted-foreground' },
}

export function PriorityIcon({ priority, className }: PriorityIconProps) {
  const config = priorityConfig[priority] ?? priorityConfig.no_priority
  const Icon = config.icon
  return (
    <Icon
      className={cn('h-4 w-4', config.color, className)}
      aria-label={config.label}
    />
  )
}

export function getPriorityLabel(priority: Priority): string {
  return priorityConfig[priority]?.label ?? 'No Priority'
}
