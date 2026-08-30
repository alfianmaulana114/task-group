export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBaseUrl as string

  const token = useState<string | null>('auth:token', () => null)

  async function request<T = any>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...((options.headers as Record<string, string>) || {}),
    }

    if (token.value) {
      headers['Authorization'] = `Bearer ${token.value}`
    }

    const res = await fetch(`${baseURL}${path}`, {
      ...options,
      headers,
      credentials: 'include',
    })

    const body = await res.json()

    if (!res.ok) {
      throw new Error(body.error || `Request failed with status ${res.status}`)
    }

    return body as T
  }

  return {
    get: <T = any>(path: string) => request<T>(path),
    post: <T = any>(path: string, data?: any) =>
      request<T>(path, { method: 'POST', body: data ? JSON.stringify(data) : undefined }),
    patch: <T = any>(path: string, data?: any) =>
      request<T>(path, { method: 'PATCH', body: data ? JSON.stringify(data) : undefined }),
    del: <T = any>(path: string) => request<T>(path, { method: 'DELETE' }),
  }
}
