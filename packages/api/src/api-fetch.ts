import { getApiConfig, refreshRequest, type TokenPair } from './auth'

let refreshPromise: Promise<TokenPair | null> | null = null

function isAuthPath(path: string): boolean {
  return path.startsWith('/auth/')
}

async function trySilentRefresh(): Promise<TokenPair | null> {
  const { tokenStorage } = getApiConfig()
  const refreshToken = tokenStorage.getRefreshToken()
  if (!refreshToken) return null

  if (!refreshPromise) {
    refreshPromise = (async () => {
      try {
        const tokens = await refreshRequest(refreshToken)
        tokenStorage.setTokens(tokens.access_token, tokens.refresh_token)
        return tokens
      } catch {
        return null
      } finally {
        refreshPromise = null
      }
    })()
  }

  return refreshPromise
}

export async function apiFetch<T = unknown>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const { baseUrl, tokenStorage, onUnauthorized } = getApiConfig()

  const makeRequest = async () => {
    const token = tokenStorage.getAccessToken()
    return fetch(`${baseUrl}${path}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      },
    })
  }

  let res = await makeRequest()

  if (res.status === 401 && !isAuthPath(path)) {
    const refreshed = await trySilentRefresh()
    if (refreshed) {
      res = await makeRequest()
    } else {
      tokenStorage.clearTokens()
      onUnauthorized?.()
    }
  }

  const json = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error((json as { error?: string }).error ?? `Request failed (${res.status})`)
  }
  return json as T
}
