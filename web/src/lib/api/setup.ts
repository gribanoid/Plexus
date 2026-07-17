import {
  configureApi,
  configureAuth,
  createLocalStorageTokenStorage,
} from '@plexus/api'
import { useAuthStore } from '@/lib/stores/auth.store'

const baseUrl =
  typeof window !== 'undefined'
    ? '/api/v1'
    : `${process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'}/api/v1`

configureApi({
  baseUrl,
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
    if (typeof window !== 'undefined' && window.location.pathname !== '/login') {
      window.location.href = '/login'
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
