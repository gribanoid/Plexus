import { configureApi, configureAuth } from '@plexus/api'
import { useAuthStore } from '../stores/auth.store'

export const API_BASE = 'http://localhost:8080/api/v1'

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
