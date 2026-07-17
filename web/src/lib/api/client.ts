import createClient from 'openapi-fetch'
import type { paths } from './schema'

const BASE_URL =
  typeof window !== 'undefined'
    ? '/api/v1'
    : `${process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'}/api/v1`

export const apiClient = createClient<paths>({ baseUrl: BASE_URL })

// Attach the access token from localStorage on every request
if (typeof window !== 'undefined') {
  apiClient.use({
    onRequest({ request }) {
      const token = localStorage.getItem('access_token')
      if (token) {
        request.headers.set('Authorization', `Bearer ${token}`)
      }
      return request
    },
    onResponse({ response }) {
      if (response.status === 401) {
        // Clear stale tokens and redirect to login
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        if (window.location.pathname !== '/login') {
          window.location.href = '/login'
        }
      }
      return response
    },
  })
}
