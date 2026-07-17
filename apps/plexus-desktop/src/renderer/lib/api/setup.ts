import { configureApi, configureAuth } from '@plexus/api'
import { useAuthStore } from '../stores/auth.store'

export const API_BASE =
  (import.meta.env.VITE_API_URL as string | undefined)?.replace(/\/$/, '') ||
  'http://127.0.0.1:8080/api/v1'

configureApi({
  baseUrl: API_BASE,
  tokenStorage: {
    getAccessToken: () => useAuthStore.getState().accessToken,
    getRefreshToken: () => useAuthStore.getState().refreshToken,
    setTokens: (accessToken, refreshToken) => {
      useAuthStore.getState().setTokens(accessToken, refreshToken)
    },
    clearTokens: () => {
      useAuthStore.getState().logout()
    },
  },
  onUnauthorized: () => {
    useAuthStore.getState().logout()
    if (typeof window !== 'undefined' && !window.location.hash.includes('/login')) {
      window.location.hash = '#/login'
    }
  },
})

configureAuth({
  setTokens: (accessToken, refreshToken) => {
    useAuthStore.getState().setTokens(accessToken, refreshToken)
  },
  setUser: (user) => {
    useAuthStore.getState().setUser(user)
  },
  getRefreshToken: () => useAuthStore.getState().refreshToken,
  logout: () => {
    useAuthStore.getState().logout()
  },
})
