import './lib/api/setup'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Toaster } from 'sonner'
import { App } from './App'
import './styles/globals.css'

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 60_000, retry: 1 } },
})

const root = document.getElementById('root')!
createRoot(root).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
      <Toaster richColors closeButton position="bottom-right" />
    </QueryClientProvider>
  </StrictMode>
)
