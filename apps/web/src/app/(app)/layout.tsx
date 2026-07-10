import { AuthGuard } from '@/components/layout/auth-guard'
import { AppSidebar } from '@/components/layout/app-sidebar'
import { AppTopbar } from '@/components/layout/app-topbar'

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthGuard>
      <div className="flex h-screen flex-col overflow-hidden bg-plexus-surface-subtle">
        <AppTopbar />
        <div className="flex min-h-0 flex-1">
          <AppSidebar />
          <main className="flex min-h-0 flex-1 flex-col overflow-hidden">{children}</main>
        </div>
      </div>
    </AuthGuard>
  )
}
