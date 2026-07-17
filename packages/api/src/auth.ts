export interface TokenStorage {
  getAccessToken(): string | null
  getRefreshToken(): string | null
  setTokens(accessToken: string, refreshToken: string): void
  clearTokens(): void
}

export interface ApiConfig {
  baseUrl: string
  tokenStorage: TokenStorage
  onUnauthorized?: () => void
}

export interface AuthCallbacks {
  setTokens: (accessToken: string, refreshToken: string) => void
  setUser: (user: MeResponse) => void
  getRefreshToken: () => string | null
  logout: () => void
}

export function createLocalStorageTokenStorage(): TokenStorage {
  return {
    getAccessToken() {
      return typeof localStorage !== 'undefined' ? localStorage.getItem('access_token') : null
    },
    getRefreshToken() {
      return typeof localStorage !== 'undefined' ? localStorage.getItem('refresh_token') : null
    },
    setTokens(accessToken, refreshToken) {
      localStorage.setItem('access_token', accessToken)
      localStorage.setItem('refresh_token', refreshToken)
    },
    clearTokens() {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
    },
  }
}

let apiConfig: ApiConfig | null = null
let authCallbacks: AuthCallbacks | null = null

export function configureApi(config: ApiConfig) {
  apiConfig = config
}

export function configureAuth(callbacks: AuthCallbacks) {
  authCallbacks = callbacks
}

export function getApiConfig(): ApiConfig {
  if (!apiConfig) {
    throw new Error('@plexus/api: call configureApi() before using the API')
  }
  return apiConfig
}

function getAuthCallbacks(): AuthCallbacks {
  if (!authCallbacks) {
    throw new Error('@plexus/api: call configureAuth() before using auth helpers')
  }
  return authCallbacks
}

async function publicFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const { baseUrl } = getApiConfig()
  const res = await fetch(`${baseUrl}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  })

  const json = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error((json as { error?: string }).error ?? `Request failed (${res.status})`)
  }
  return json as T
}

export function resolveLoginEmail(input: string): string {
  const trimmed = input.trim()
  if (trimmed.toLowerCase() === 'admin') return 'admin@plexus.local'
  return trimmed
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface MeResponse {
  id: string
  email: string
  display_name: string
  avatar_url?: string | null
  role: 'admin' | 'user'
  created_at: string
}

export async function loginRequest(email: string, password: string): Promise<TokenPair> {
  return publicFetch<TokenPair>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email: resolveLoginEmail(email), password }),
  })
}

export async function registerRequest(data: {
  email: string
  password: string
  display_name: string
}): Promise<TokenPair> {
  return publicFetch<TokenPair>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function refreshRequest(refreshToken: string): Promise<TokenPair> {
  return publicFetch<TokenPair>('/auth/refresh', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
}

export async function fetchMe(): Promise<MeResponse> {
  const { apiFetch } = await import('./api-fetch')
  return apiFetch<MeResponse>('/me')
}

export async function completeAuth(tokens: TokenPair): Promise<MeResponse> {
  const { setTokens, setUser } = getAuthCallbacks()
  setTokens(tokens.access_token, tokens.refresh_token)
  const user = await fetchMe()
  setUser(user)
  return user
}

export async function logoutRequest(): Promise<void> {
  const { getRefreshToken, logout } = getAuthCallbacks()
  const refreshToken = getRefreshToken()
  try {
    const { apiFetch } = await import('./api-fetch')
    await apiFetch('/auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
  } catch {
    // ignore logout API errors
  }
  logout()
}
