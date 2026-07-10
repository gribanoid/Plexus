'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Button } from '@plexus/ui'
import { apiFetch, completeAuth, loginRequest, routes } from '@plexus/api'

const schema = z.object({
  email: z.string().min(1, 'Email or username is required'),
  password: z.string().min(1, 'Password is required'),
})

type FormData = z.infer<typeof schema>

interface Org {
  slug: string
}

interface Project {
  key: string
}

async function redirectAfterAuth(router: ReturnType<typeof useRouter>) {
  const orgs = await apiFetch<{ items: Org[] }>('/orgs')
  const org = orgs.items[0]
  if (!org) {
    router.push(routes.orgs())
    return
  }
  const projects = await apiFetch<{ items: Project[] }>(`/orgs/${org.slug}/projects`)
  const project = projects.items[0]
  router.push(
    project ? routes.projectBoard(org.slug, project.key) : routes.org(org.slug),
  )
}

export default function LoginPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({ resolver: zodResolver(schema) })

  async function onSubmit(data: FormData) {
    setLoading(true)
    try {
      const tokens = await loginRequest(data.email, data.password)
      await completeAuth(tokens)
      await redirectAfterAuth(router)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen">
      <div className="hidden w-1/2 flex-col justify-between bg-[#0747A6] p-12 text-white lg:flex">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded bg-white/15 text-lg font-bold">
            P
          </div>
          <span className="text-xl font-semibold">Plexus</span>
        </div>
        <div>
          <h1 className="text-3xl font-bold leading-tight">
            Plan, track, and ship work — together.
          </h1>
          <p className="mt-4 max-w-md text-white/80">
            Boards, backlogs, and sprints for your team in one place.
          </p>
        </div>
        <p className="text-sm text-white/60">© Plexus</p>
      </div>

      <div className="flex flex-1 items-center justify-center bg-[#FAFBFC] px-4">
        <div className="w-full max-w-sm space-y-6">
          <div className="space-y-2 text-center lg:text-left">
            <h2 className="text-2xl font-semibold text-[#172B4D]">Log in to continue</h2>
            <p className="text-sm text-[#5E6C84]">Use your workspace account</p>
          </div>

          <div className="rounded-md border border-[#DFE1E6] bg-[#DEEBFF] px-3 py-2 text-xs text-[#0747A6]">
            Dev account: <strong>admin</strong> / <strong>admin</strong>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-1">
              <label className="text-sm font-medium text-[#172B4D]" htmlFor="email">
                Email or username
              </label>
              <input
                id="email"
                type="text"
                autoComplete="username"
                className="flex h-10 w-full rounded border border-[#DFE1E6] bg-white px-3 text-sm text-[#172B4D] shadow-sm focus:border-[#4C9AFF] focus:outline-none focus:ring-2 focus:ring-[#4C9AFF]/30"
                placeholder="admin or you@company.com"
                {...register('email')}
              />
              {errors.email && <p className="text-xs text-[#DE350B]">{errors.email.message}</p>}
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium text-[#172B4D]" htmlFor="password">
                Password
              </label>
              <input
                id="password"
                type="password"
                autoComplete="current-password"
                className="flex h-10 w-full rounded border border-[#DFE1E6] bg-white px-3 text-sm text-[#172B4D] shadow-sm focus:border-[#4C9AFF] focus:outline-none focus:ring-2 focus:ring-[#4C9AFF]/30"
                placeholder="••••••••"
                {...register('password')}
              />
              {errors.password && (
                <p className="text-xs text-[#DE350B]">{errors.password.message}</p>
              )}
            </div>

            <Button
              type="submit"
              className="h-10 w-full bg-[#0052CC] hover:bg-[#0065FF] text-white"
              disabled={loading}
            >
              {loading ? 'Signing in…' : 'Log in'}
            </Button>
          </form>

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t border-[#DFE1E6]" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-[#FAFBFC] px-2 text-[#5E6C84]">or</span>
            </div>
          </div>

          <Button
            type="button"
            variant="outline"
            className="h-10 w-full"
            onClick={() => {
              window.location.href = '/api/v1/auth/oidc/login'
            }}
          >
            Continue with SSO
          </Button>

          <p className="text-center text-sm text-[#5E6C84]">
            Don&apos;t have an account?{' '}
            <Link href="/register" className="font-medium text-[#0052CC] hover:underline">
              Sign up
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
