'use client'

import { use, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@plexus/ui'
import { apiFetch } from '@plexus/api'

interface SettingsPageProps {
  params: Promise<{ orgSlug: string; projectKey: string }>
}

type Tab = 'workflow' | 'issue-types' | 'labels' | 'members'

interface Status {
  id: string
  name: string
  color: string
  category: 'todo' | 'in_progress' | 'done'
  position: number
}

interface IssueType {
  id: string
  name: string
  color: string
}

interface Label {
  id: string
  name: string
  color: string
}

interface OrgMember {
  id: string
  display_name: string
  email: string
  role: string
  joined_at: string
}

const TABS: { id: Tab; label: string }[] = [
  { id: 'workflow', label: 'Workflow' },
  { id: 'issue-types', label: 'Issue Types' },
  { id: 'labels', label: 'Labels' },
  { id: 'members', label: 'Members' },
]

const STATUS_CATEGORIES = [
  { value: 'todo', label: 'To do' },
  { value: 'in_progress', label: 'In progress' },
  { value: 'done', label: 'Done' },
] as const

export default function ProjectSettingsPage({ params }: SettingsPageProps) {
  const { orgSlug, projectKey } = use(params)
  const [tab, setTab] = useState<Tab>('workflow')
  const queryClient = useQueryClient()
  const base = `/orgs/${orgSlug}/projects/${projectKey}`

  const { data: statuses = [], isLoading: statusesLoading } = useQuery<Status[]>({
    queryKey: ['statuses', orgSlug, projectKey],
    queryFn: async () => {
      const json = await apiFetch<{ items: Status[] }>(`${base}/statuses`)
      return json.items ?? []
    },
  })

  const { data: issueTypes = [], isLoading: typesLoading } = useQuery<IssueType[]>({
    queryKey: ['issue-types', orgSlug, projectKey],
    queryFn: async () => {
      const json = await apiFetch<{ items: IssueType[] }>(`${base}/issue-types`)
      return json.items ?? []
    },
  })

  const { data: labels = [], isLoading: labelsLoading } = useQuery<Label[]>({
    queryKey: ['labels', orgSlug, projectKey],
    queryFn: async () => {
      const json = await apiFetch<{ items: Label[] }>(`${base}/labels`)
      return json.items ?? []
    },
  })

  const { data: members = [], isLoading: membersLoading } = useQuery<OrgMember[]>({
    queryKey: ['org-members', orgSlug],
    queryFn: async () => {
      const json = await apiFetch<{ items: OrgMember[] }>(`/orgs/${orgSlug}/members`)
      return json.items ?? []
    },
  })

  const createStatus = useMutation({
    mutationFn: async (body: { name: string; color: string; category: string }) => {
      await apiFetch(`${base}/statuses`, { method: 'POST', body: JSON.stringify(body) })
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['statuses', orgSlug, projectKey] })
      toast.success('Status created')
    },
    onError() {
      toast.error('Failed to create status')
    },
  })

  const deleteStatus = useMutation({
    mutationFn: async (statusId: string) => {
      await apiFetch(`${base}/statuses/${statusId}`, { method: 'DELETE' })
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['statuses', orgSlug, projectKey] })
      toast.success('Status deleted')
    },
    onError() {
      toast.error('Failed to delete status')
    },
  })

  const createIssueType = useMutation({
    mutationFn: async (body: { name: string; color: string }) => {
      await apiFetch(`${base}/issue-types`, { method: 'POST', body: JSON.stringify(body) })
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['issue-types', orgSlug, projectKey] })
      toast.success('Issue type created')
    },
    onError() {
      toast.error('Failed to create issue type')
    },
  })

  const createLabel = useMutation({
    mutationFn: async (body: { name: string; color: string }) => {
      await apiFetch(`${base}/labels`, { method: 'POST', body: JSON.stringify(body) })
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['labels', orgSlug, projectKey] })
      toast.success('Label created')
    },
    onError() {
      toast.error('Failed to create label')
    },
  })

  const inviteMember = useMutation({
    mutationFn: async (body: { email: string; role: string }) => {
      await apiFetch(`/orgs/${orgSlug}/members/invite`, {
        method: 'POST',
        body: JSON.stringify(body),
      })
    },
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ['org-members', orgSlug] })
      toast.success('Invitation sent')
    },
    onError() {
      toast.error('Failed to invite member')
    },
  })

  return (
    <div className="mx-auto max-w-3xl px-6 py-8">
      <h1 className="text-2xl font-semibold text-plexus-text">{projectKey} — Settings</h1>
      <p className="mt-1 text-sm text-plexus-text-subtle">
        Manage workflow, types, labels, and team members.
      </p>

      <div className="mt-6 flex gap-1 border-b border-plexus-border">
        {TABS.map((t) => (
          <button
            key={t.id}
            type="button"
            onClick={() => setTab(t.id)}
            className={[
              'px-4 py-2 text-sm font-medium transition-colors',
              tab === t.id
                ? 'border-b-2 border-plexus-brand text-plexus-brand'
                : 'text-plexus-text-subtle hover:text-plexus-text',
            ].join(' ')}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div className="mt-6">
        {tab === 'workflow' && (
          <WorkflowTab
            statuses={statuses}
            loading={statusesLoading}
            onCreate={(data) => createStatus.mutate(data)}
            onDelete={(id) => deleteStatus.mutate(id)}
            creating={createStatus.isPending}
          />
        )}
        {tab === 'issue-types' && (
          <IssueTypesTab
            issueTypes={issueTypes}
            loading={typesLoading}
            onCreate={(data) => createIssueType.mutate(data)}
            creating={createIssueType.isPending}
          />
        )}
        {tab === 'labels' && (
          <LabelsTab
            labels={labels}
            loading={labelsLoading}
            onCreate={(data) => createLabel.mutate(data)}
            creating={createLabel.isPending}
          />
        )}
        {tab === 'members' && (
          <MembersTab
            members={members}
            loading={membersLoading}
            onInvite={(data) => inviteMember.mutate(data)}
            inviting={inviteMember.isPending}
          />
        )}
      </div>
    </div>
  )
}

function WorkflowTab({
  statuses,
  loading,
  onCreate,
  onDelete,
  creating,
}: {
  statuses: Status[]
  loading: boolean
  onCreate: (data: { name: string; color: string; category: string }) => void
  onDelete: (id: string) => void
  creating: boolean
}) {
  const [name, setName] = useState('')
  const [color, setColor] = useState('#4C9AFF')
  const [category, setCategory] = useState('todo')

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    onCreate({ name: name.trim(), color, category })
    setName('')
  }

  if (loading) return <LoadingState />

  return (
    <div className="space-y-6">
      <form onSubmit={handleCreate} className="flex flex-wrap items-end gap-3">
        <Field label="Name">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="In Review"
            className={inputClass}
          />
        </Field>
        <Field label="Color">
          <input type="color" value={color} onChange={(e) => setColor(e.target.value)} className="h-10 w-14 cursor-pointer rounded border border-plexus-border" />
        </Field>
        <Field label="Category">
          <select value={category} onChange={(e) => setCategory(e.target.value)} className={inputClass}>
            {STATUS_CATEGORIES.map((c) => (
              <option key={c.value} value={c.value}>{c.label}</option>
            ))}
          </select>
        </Field>
        <Button type="submit" disabled={creating || !name.trim()} className="bg-plexus-brand text-white hover:bg-plexus-brand-hover">
          <Plus className="mr-1.5 h-4 w-4" />
          Add status
        </Button>
      </form>

      <div className="divide-y rounded border border-plexus-border bg-plexus-surface">
        {statuses.length === 0 ? (
          <p className="px-4 py-6 text-center text-sm text-plexus-text-subtle">No statuses yet</p>
        ) : (
          statuses.map((s) => (
            <div key={s.id} className="flex items-center gap-3 px-4 py-3">
              <span className="h-3 w-3 rounded-full" style={{ backgroundColor: s.color }} />
              <span className="flex-1 text-sm font-medium text-plexus-text">{s.name}</span>
              <span className="text-xs capitalize text-plexus-text-subtle">{s.category.replace('_', ' ')}</span>
              <button
                type="button"
                onClick={() => onDelete(s.id)}
                className="rounded p-1.5 text-plexus-text-subtle hover:bg-black/5 hover:text-[#DE350B] dark:hover:bg-white/5"
                title="Delete status"
              >
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

function IssueTypesTab({
  issueTypes,
  loading,
  onCreate,
  creating,
}: {
  issueTypes: IssueType[]
  loading: boolean
  onCreate: (data: { name: string; color: string }) => void
  creating: boolean
}) {
  const [name, setName] = useState('')
  const [color, setColor] = useState('#6554C0')

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    onCreate({ name: name.trim(), color })
    setName('')
  }

  if (loading) return <LoadingState />

  return (
    <div className="space-y-6">
      <form onSubmit={handleCreate} className="flex flex-wrap items-end gap-3">
        <Field label="Name">
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Bug" className={inputClass} />
        </Field>
        <Field label="Color">
          <input type="color" value={color} onChange={(e) => setColor(e.target.value)} className="h-10 w-14 cursor-pointer rounded border border-plexus-border" />
        </Field>
        <Button type="submit" disabled={creating || !name.trim()} className="bg-plexus-brand text-white hover:bg-plexus-brand-hover">
          <Plus className="mr-1.5 h-4 w-4" />
          Add type
        </Button>
      </form>

      <div className="divide-y rounded border border-plexus-border bg-plexus-surface">
        {issueTypes.length === 0 ? (
          <p className="px-4 py-6 text-center text-sm text-plexus-text-subtle">No issue types yet</p>
        ) : (
          issueTypes.map((t) => (
            <div key={t.id} className="flex items-center gap-3 px-4 py-3">
              <span className="h-3 w-3 rounded-sm" style={{ backgroundColor: t.color }} />
              <span className="text-sm font-medium text-plexus-text">{t.name}</span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

function LabelsTab({
  labels,
  loading,
  onCreate,
  creating,
}: {
  labels: Label[]
  loading: boolean
  onCreate: (data: { name: string; color: string }) => void
  creating: boolean
}) {
  const [name, setName] = useState('')
  const [color, setColor] = useState('#36B37E')

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    onCreate({ name: name.trim(), color })
    setName('')
  }

  if (loading) return <LoadingState />

  return (
    <div className="space-y-6">
      <form onSubmit={handleCreate} className="flex flex-wrap items-end gap-3">
        <Field label="Name">
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="frontend" className={inputClass} />
        </Field>
        <Field label="Color">
          <input type="color" value={color} onChange={(e) => setColor(e.target.value)} className="h-10 w-14 cursor-pointer rounded border border-plexus-border" />
        </Field>
        <Button type="submit" disabled={creating || !name.trim()} className="bg-plexus-brand text-white hover:bg-plexus-brand-hover">
          <Plus className="mr-1.5 h-4 w-4" />
          Add label
        </Button>
      </form>

      <div className="flex flex-wrap gap-2">
        {labels.length === 0 ? (
          <p className="text-sm text-plexus-text-subtle">No labels yet</p>
        ) : (
          labels.map((l) => (
            <span
              key={l.id}
              className="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium text-white"
              style={{ backgroundColor: l.color }}
            >
              {l.name}
            </span>
          ))
        )}
      </div>
    </div>
  )
}

function MembersTab({
  members,
  loading,
  onInvite,
  inviting,
}: {
  members: OrgMember[]
  loading: boolean
  onInvite: (data: { email: string; role: string }) => void
  inviting: boolean
}) {
  const [email, setEmail] = useState('')
  const [role, setRole] = useState('member')

  function handleInvite(e: React.FormEvent) {
    e.preventDefault()
    if (!email.trim()) return
    onInvite({ email: email.trim(), role })
    setEmail('')
  }

  if (loading) return <LoadingState />

  return (
    <div className="space-y-6">
      <form onSubmit={handleInvite} className="flex flex-wrap items-end gap-3">
        <Field label="Email">
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="colleague@company.com"
            className={inputClass}
          />
        </Field>
        <Field label="Role">
          <select value={role} onChange={(e) => setRole(e.target.value)} className={inputClass}>
            <option value="member">Member</option>
            <option value="admin">Admin</option>
            <option value="guest">Guest</option>
            <option value="owner">Owner</option>
          </select>
        </Field>
        <Button type="submit" disabled={inviting || !email.trim()} className="bg-plexus-brand text-white hover:bg-plexus-brand-hover">
          <Plus className="mr-1.5 h-4 w-4" />
          Invite
        </Button>
      </form>

      <div className="divide-y rounded border border-plexus-border bg-plexus-surface">
        {members.length === 0 ? (
          <p className="px-4 py-6 text-center text-sm text-plexus-text-subtle">No members yet</p>
        ) : (
          members.map((m) => (
            <div key={m.id} className="flex items-center gap-3 px-4 py-3">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-plexus-brand/15 text-xs font-semibold text-plexus-brand">
                {m.display_name[0]?.toUpperCase() ?? '?'}
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-plexus-text">{m.display_name}</p>
                <p className="truncate text-xs text-plexus-text-subtle">{m.email}</p>
              </div>
              <span className="rounded bg-secondary px-2 py-0.5 text-xs capitalize text-plexus-text-subtle">
                {m.role}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1">
      <label className="text-sm font-medium text-plexus-text">{label}</label>
      {children}
    </div>
  )
}

function LoadingState() {
  return (
    <div className="flex justify-center py-12">
      <div className="h-6 w-6 animate-spin rounded-full border-2 border-plexus-brand border-t-transparent" />
    </div>
  )
}

const inputClass =
  'flex h-10 min-w-[140px] rounded border border-plexus-border bg-plexus-surface px-3 text-sm text-plexus-text focus:border-plexus-brand focus:outline-none focus:ring-2 focus:ring-plexus-brand/30'
