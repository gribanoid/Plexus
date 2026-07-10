'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Button } from '@plexus/ui'
import { apiFetch, completeAuth, registerRequest, routes } from '@plexus/api'

const schema = z.object({
  display_name: z.string().min(2, 'Name must be at least 2 characters'),
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
})

type FormData = z.infer<typeof schema>

export default function RegisterPage() {
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
      const tokens = await registerRequest(data)
      await completeAuth(tokens)
      const orgs = await apiFetch<{ items: { slug: string }[] }>('/orgs')
      const org = orgs.items[0]
      if (org) {
        const projects = await apiFetch<{ items: { key: string }[] }>(`/orgs/${org.slug}/projects`)
        const project = projects.items[0]
        router.push(
          project ? routes.projectBoard(org.slug, project.key) : routes.org(org.slug),
        )
      } else {
        router.push(routes.orgs())
      }
      toast.success('Account created!')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Registration failed')
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
          <h1 className="text-3xl font-bold">Get started in minutes</h1>
          <p className="mt-4 max-w-md text-white/80">
            We&apos;ll create your workspace and first project automatically.
          </p>
        </div>
        <p className="text-sm text-white/60">© Plexus</p>
      </div>

      <div className="flex flex-1 items-center justify-center bg-[#FAFBFC] px-4">
        <div className="w-full max-w-sm space-y-6">
          <div className="space-y-2 text-center lg:text-left">
            <h2 className="text-2xl font-semibold text-[#172B4D]">Sign up</h2>
            <p className="text-sm text-[#5E6C84]">Create your account and workspace</p>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-1">
              <label className="text-sm font-medium text-[#172B4D]" htmlFor="display_name">
                Full name
              </label>
              <input
                id="display_name"
                type="text"
                autoComplete="name"
                className="flex h-10 w-full rounded border border-[#DFE1E6] bg-white px-3 text-sm focus:border-[#4C9AFF] focus:outline-none focus:ring-2 focus:ring-[#4C9AFF]/30"
                placeholder="Alice Smith"
                {...register('display_name')}
              />
              {errors.display_name && (
                <p className="text-xs text-[#DE350B]">{errors.display_name.message}</p>
              )}
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium text-[#172B4D]" htmlFor="email">
                Work email
              </label>
              <input
                id="email"
                type="email"
                autoComplete="email"
                className="flex h-10 w-full rounded border border-[#DFE1E6] bg-white px-3 text-sm focus:border-[#4C9AFF] focus:outline-none focus:ring-2 focus:ring-[#4C9AFF]/30"
                placeholder="you@company.com"
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
                autoComplete="new-password"
                className="flex h-10 w-full rounded border border-[#DFE1E6] bg-white px-3 text-sm focus:border-[#4C9AFF] focus:outline-none focus:ring-2 focus:ring-[#4C9AFF]/30"
                placeholder="Min 8 characters"
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
              {loading ? 'Creating account…' : 'Sign up'}
            </Button>
          </form>

          <p className="text-center text-sm text-[#5E6C84]">
            Already have an account?{' '}
            <Link href="/login" className="font-medium text-[#0052CC] hover:underline">
              Log in
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
