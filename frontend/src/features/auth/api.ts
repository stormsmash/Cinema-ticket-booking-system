export interface AuthUser {
  id: string
  email: string
  name: string
  avatar_url?: string
  role: 'USER' | 'ADMIN'
}

export interface AuthConfig {
  google_enabled: boolean
}

interface ApiResponse<T> {
  data: T
}

export async function fetchAuthConfig(): Promise<AuthConfig> {
  const response = await fetch('/api/v1/auth/config', {
    headers: { Accept: 'application/json' },
  })

  if (!response.ok) {
    throw new Error(`Auth config request failed with status ${response.status}`)
  }

  const payload = (await response.json()) as ApiResponse<AuthConfig>
  return payload.data
}

export async function fetchCurrentUser(): Promise<AuthUser | null> {
  const response = await fetch('/api/v1/auth/me', {
    credentials: 'same-origin',
    headers: { Accept: 'application/json' },
  })

  if (response.status === 401) {
    return null
  }
  if (!response.ok) {
    throw new Error(`Current user request failed with status ${response.status}`)
  }

  const payload = (await response.json()) as ApiResponse<AuthUser>
  return payload.data
}

export async function logout(): Promise<void> {
  const response = await fetch('/api/v1/auth/logout', {
    method: 'POST',
    credentials: 'same-origin',
  })

  if (!response.ok) {
    throw new Error(`Logout request failed with status ${response.status}`)
  }
}
