import type { ReactNode } from 'react'
import { ChevronDown, ChevronRight, Plus } from 'lucide-react'
import { PriorityIcon, IssueStatusBadge, Button } from '@plexus/ui'

export interface SprintIssue {
  id: string
  number: number
  title: string
  priority: 'urgent' | 'high' | 'medium' | 'low' | 'no_priority'
  status_id: string
  story_points?: number | null
  sprint_id?: string | null
}

export interface SprintStatus {
  id: string
  name: string
  color: string
  category: 'todo' | 'in_progress' | 'done'
}

export interface SprintSectionProps {
  sprint: { id: string; name: string; state: string }
  issues: SprintIssue[]
  statusMap: Record<string, SprintStatus>
  projectKey: string
  orgSlug: string
  collapsed: boolean
  onToggle: () => void
  onAddIssue: () => void
  onSprintAction?: () => void
  sprintActionLabel?: string
  isBacklog?: boolean
  onIssueClick?: (issueNumber: number) => void
  renderIssueRow?: (issue: SprintIssue, children: ReactNode) => ReactNode
}

export function SprintSection({
  sprint,
  issues,
  statusMap,
  projectKey,
  orgSlug: _orgSlug,
  collapsed,
  onToggle,
  onAddIssue,
  onSprintAction,
  sprintActionLabel,
  isBacklog = false,
  onIssueClick,
  renderIssueRow,
}: SprintSectionProps) {
  const emptyText = isBacklog ? 'No issues in backlog' : 'No issues in this sprint'

  function renderRow(issue: SprintIssue) {
    const status = statusMap[issue.status_id]
    const content = (
      <>
        <PriorityIcon priority={issue.priority} />
        {status && (
          <IssueStatusBadge
            name={status.name}
            color={status.color}
            category={status.category}
          />
        )}
        <span className="flex-1 truncate text-sm text-plexus-text">{issue.title}</span>
        <span className="text-xs text-muted-foreground">
          {projectKey}-{issue.number}
        </span>
        {issue.story_points != null && (
          <span className="rounded bg-secondary px-1.5 py-0.5 text-xs font-medium">
            {issue.story_points}
          </span>
        )}
      </>
    )

    const rowClass =
      'flex items-center gap-3 px-4 py-2.5 hover:bg-black/[0.03] dark:hover:bg-white/[0.03]'

    if (renderIssueRow) {
      return (
        <div key={issue.id}>
          {renderIssueRow(issue, <div className={rowClass}>{content}</div>)}
        </div>
      )
    }

    if (onIssueClick) {
      return (
        <button
          key={issue.id}
          type="button"
          onClick={() => onIssueClick(issue.number)}
          className={`w-full text-left ${rowClass}`}
        >
          {content}
        </button>
      )
    }

    return (
      <div key={issue.id} className={rowClass}>
        {content}
      </div>
    )
  }

  return (
    <div className="rounded border border-plexus-border bg-plexus-surface">
      <div className="flex items-center gap-2 px-4 py-3 hover:bg-black/[0.03] dark:hover:bg-white/[0.03]">
        <button
          type="button"
          className="flex min-w-0 flex-1 items-center gap-2 text-left"
          onClick={onToggle}
        >
          {collapsed ? (
            <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
          ) : (
            <ChevronDown className="h-4 w-4 shrink-0 text-muted-foreground" />
          )}
          <span className="text-sm font-medium text-plexus-text">{sprint.name}</span>
          {sprint.state === 'active' && (
            <span className="rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">
              Active
            </span>
          )}
          <span className="ml-auto text-xs text-muted-foreground">{issues.length} issues</span>
        </button>
        {!isBacklog && onSprintAction && sprintActionLabel && (
          <Button
            size="sm"
            variant="ghost"
            className="h-7 shrink-0 px-2 text-xs"
            onClick={(e) => {
              e.stopPropagation()
              onSprintAction()
            }}
          >
            {sprintActionLabel}
          </Button>
        )}
      </div>

      {!collapsed && (
        <div className="divide-y border-t">
          {issues.length === 0 ? (
            <p className="px-4 py-6 text-center text-sm text-plexus-text-subtle">{emptyText}</p>
          ) : (
            issues.map((issue) => renderRow(issue))
          )}
          <div className="px-4 py-2">
            <button
              type="button"
              className="flex items-center gap-1.5 text-xs text-plexus-text-subtle hover:text-plexus-text"
              onClick={onAddIssue}
            >
              <Plus className="h-3.5 w-3.5" />
              Add issue
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
