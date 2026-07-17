import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import { ThemeProvider } from 'next-themes'
import { Toaster } from 'sonner'
import { Providers } from './providers'
import '@/styles/globals.css'

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' })

export const metadata: Metadata = {
  title: { default: 'Plexus', template: '%s | Plexus' },
  description: 'Enterprise project management — fast, flexible, powerful',
  icons: { icon: '/favicon.ico' },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning className={inter.variable}>
      <body className="min-h-screen bg-background font-sans antialiased">
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
          <Providers>{children}</Providers>
          <Toaster richColors closeButton position="bottom-right" />
        </ThemeProvider>
      </body>
    </html>
  )
}
