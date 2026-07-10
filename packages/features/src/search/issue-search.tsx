import { useEffect, useRef, useState } from 'react'
import { Search } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { apiFetch, useSearch, type SearchIssueHit } from '@plexus/api'

interface Project {
  id: string
  key: string
  name: string
}

export interface IssueSearchProps {
  orgSlug: string
  projectKey?: string
  variant?: 'topbar' | 'panel'
  autoFocus?: boolean
  onIssueSelect?: (hit: SearchIssueHit, projectKey: string) => void
  onNavigate?: () => void
}

export function IssueSearch({
  orgSlug,
  projectKey,
  variant = 'topbar',
  autoFocus = false,
  onIssueSelect,
  onNavigate,
}: IssueSearchProps) {
  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const { data: projects = [] } = useQuery<Project[]>({
    queryKey: ['projects', orgSlug],
    enabled: Boolean(orgSlug) && !projectKey,
    queryFn: async () => {
      const json = await apiFetch<{ items: Project[] }>(`/orgs/${orgSlug}/projects`)
      return json.items ?? []
    },
  })

  const projectKeyById = Object.fromEntries(projects.map((p) => [p.id, p.key]))

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(query), 300)
    return () => clearTimeout(timer)
  }, [query])

  const { data, isFetching } = useSearch({
    orgSlug,
    query: debouncedQuery,
    projectKey,
    enabled: open,
  })

  const results = data?.items ?? []

  useEffect(() => {
    function handleFocusSearch() {
      inputRef.current?.focus()
      setOpen(true)
    }
    window.addEventListener('plexus:focus-search', handleFocusSearch)
    return () => window.removeEventListener('plexus:focus-search', handleFocusSearch)
  }, [])

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  function resolveProjectKey(hit: SearchIssueHit): string | null {
    if (projectKey) return projectKey
    return projectKeyById[hit.project_id] ?? null
  }

  function selectIssue(hit: SearchIssueHit) {
    const key = resolveProjectKey(hit)
    if (!key || !onIssueSelect) return
    onIssueSelect(hit, key)
    setQuery('')
    setDebouncedQuery('')
    setOpen(false)
    onNavigate?.()
  }

  const isTopbar = variant === 'topbar'

  // Topbar is always dark — VK-style pill: translucent bg, icon left, light text.
  const inputWrapperClass = isTopbar
    ? 'flex h-9 w-full items-center gap-3 rounded-full bg-white/10 px-4 transition-colors focus-within:bg-white/15'
    : 'flex h-10 w-full items-center gap-3 rounded-lg border border-plexus-border bg-plexus-surface px-4 focus-within:border-plexus-brand focus-within:ring-2 focus-within:ring-plexus-brand/30'

  const inputClass = isTopbar
    ? 'min-w-0 flex-1 border-0 bg-transparent text-sm !text-white caret-white placeholder:!text-white/55 focus:outline-none'
    : 'min-w-0 flex-1 border-0 bg-transparent text-sm text-plexus-text caret-plexus-brand placeholder:text-plexus-text-muted focus:outline-none'

  const iconClass = isTopbar
    ? 'h-3.5 w-3.5 shrink-0 text-white/55'
    : 'h-3.5 w-3.5 shrink-0 text-plexus-text-muted'

  return (
    <div ref={containerRef} className="relative w-full">
      <div className={inputWrapperClass}>
        <Search className={iconClass} aria-hidden />
        <input
          ref={inputRef}
          data-issue-search
          autoFocus={autoFocus}
          placeholder="Search issues..."
          value={query}
          onChange={(e) => {
            setQuery(e.target.value)
            setOpen(true)
          }}
          onFocus={() => setOpen(true)}
          className={inputClass}
        />
      </div>

      {open && debouncedQuery.trim().length >= 2 && (
        <div className="absolute left-0 right-0 top-full z-50 mt-1 max-h-80 overflow-y-auto rounded border border-plexus-border bg-plexus-surface shadow-lg">
          {isFetching && results.length === 0 ? (
            <p className="px-3 py-4 text-center text-sm text-plexus-text-subtle">Searching…</p>
          ) : results.length === 0 ? (
            <p className="px-3 py-4 text-center text-sm text-plexus-text-subtle">No issues found</p>
          ) : (
            <ul>
              {results.map((hit) => {
                const key = resolveProjectKey(hit)
                return (
                  <li key={hit.id}>
                    <button
                      type="button"
                      disabled={!key || !onIssueSelect}
                      onClick={() => selectIssue(hit)}
                      className="flex w-full flex-col gap-0.5 px-3 py-2.5 text-left transition-colors hover:bg-black/5 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-white/5"
                    >
                      <span className="truncate text-sm font-medium text-plexus-text">
                        {key ? `${key}-${hit.number}` : `#${hit.number}`} — {hit.title}
                      </span>
                      <span className="text-xs text-plexus-text-subtle">
                        {hit.status_name}
                        {hit.assignee_name ? ` · ${hit.assignee_name}` : ''}
                      </span>
                    </button>
                  </li>
                )
              })}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
