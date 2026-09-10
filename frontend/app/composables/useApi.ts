export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBaseUrl as string
  const token = useState<string | null>('auth:token', () => null)
  const orgId = useState<string | null>('org:org_id', () => null)
  const loading = useState<number>('api:loadingCount', () => 0)
  const isLoading = computed(() => loading.value > 0)

  // dedup refresh: single promise for concurrent 401s
  let refreshPromise: Promise<string | null> | null = null

  async function request<T = any>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...((options.headers as Record<string, string>) || {}),
    }

    // Fallback ke localStorage jika useState belum ter-hydrate (race initAuth)
    let effectiveToken = token.value
    if (!effectiveToken && import.meta.client) {
      effectiveToken = localStorage.getItem('auth:token')
      if (effectiveToken) token.value = effectiveToken
    }
    if (effectiveToken) {
      headers['Authorization'] = `Bearer ${effectiveToken}`
    }

    let effectiveOrgId = orgId.value
    if (!effectiveOrgId && import.meta.client) {
      effectiveOrgId = localStorage.getItem('auth:org_id')
      if (effectiveOrgId) orgId.value = effectiveOrgId
    }
    if (effectiveOrgId) {
      headers['X-Organization-Id'] = effectiveOrgId
    }

    loading.value++

    try {
      const res = await fetch(`${baseURL}${path}`, {
        ...options,
        headers,
        credentials: 'include',
      })

      const contentType = res.headers.get('content-type') || ''
      let body: any = null
      if (contentType.includes('application/json')) {
        body = await res.json()
      } else {
        const text = await res.text()
        try { body = JSON.parse(text) } catch { body = { error: text || 'Unexpected response' } }
      }

      if (!res.ok) {
        if (res.status === 0) {
          throw new Error('Tidak dapat terhubung ke server. Periksa koneksi internet Anda.')
        }
        if (res.status === 503) {
          throw new Error('Layanan sementara tidak tersedia. Coba lagi sebentar.')
        }
        if (res.status === 401) {
          const isAuthPath = path.includes('/auth/refresh') || path.includes('/auth/login') || path.includes('/auth/register')
          if (!isAuthPath && import.meta.client) {
            // dedup refresh for concurrent 401s
            if (!refreshPromise) {
              refreshPromise = (async () => {
                try {
                  const refreshRes = await fetch(`${baseURL}/api/v1/auth/refresh`, {
                    method: 'POST',
                    credentials: 'include',
                    headers: { 'Content-Type': 'application/json' },
                  })
                  if (refreshRes.ok) {
                    const refreshBody = await refreshRes.json()
                    const newToken = refreshBody.token || refreshBody.access_token
                    if (newToken) {
                      token.value = newToken
                      localStorage.setItem('auth:token', newToken)
                      return newToken as string
                    }
                  }
                  return null
                } catch {
                  return null
                } finally {
                  // reset after short delay to avoid infinite loop; keep promise until awaiters done
                  setTimeout(() => { refreshPromise = null }, 0)
                }
              })()
            }
            const newToken = await refreshPromise
            refreshPromise = null
            if (newToken) {
              // retry once with fresh token (avoid recursion loop: add header directly)
              const retryHeaders: Record<string, string> = { ...headers, 'Authorization': `Bearer ${newToken}` }
              const retryRes = await fetch(`${baseURL}${path}`, { ...options, headers: retryHeaders, credentials: 'include' })
              const ct2 = retryRes.headers.get('content-type') || ''
              let retryBody: any = null
              if (ct2.includes('application/json')) retryBody = await retryRes.json()
              else {
                const t = await retryRes.text()
                try { retryBody = JSON.parse(t) } catch { retryBody = { error: t || 'Unexpected response' } }
              }
              if (retryRes.ok) return retryBody as T
              // retry also 401 -> fall through to logout
              body = retryBody
            }
          }
          // refresh gagal -> sesi habis, paksa ke login
          if (import.meta.client) {
            localStorage.removeItem('auth:token')
            localStorage.removeItem('auth:org_id')
            localStorage.removeItem('auth:org_role')
            token.value = null
            orgId.value = null
            if (!path.includes('/auth/')) {
              const currentPath = window.location.pathname
              if (currentPath !== '/login' && currentPath !== '/register') {
                window.location.href = '/login'
              }
            }
          }
          throw new Error(body.error || 'Sesi telah berakhir, silakan login kembali.')
        }
        if (res.status === 403) {
          throw new Error(body.error || 'Anda tidak memiliki akses ke sumber daya ini.')
        }
        if (res.status === 404) {
          throw new Error(body.error || 'Sumber daya tidak ditemukan.')
        }
        if (res.status === 422) {
          throw new Error(body.error || 'Data yang dikirim tidak valid.')
        }
        throw new Error(body.error || `Permintaan gagal (status ${res.status})`)
      }

      return body as T
    } finally {
      loading.value = Math.max(0, loading.value - 1)
    }
  }

  return {
    loading: readonly(isLoading),
    get: <T = any>(path: string) => request<T>(path),
    post: <T = any>(path: string, data?: any) =>
      request<T>(path, { method: 'POST', body: data ? JSON.stringify(data) : undefined }),
    patch: <T = any>(path: string, data?: any) =>
      request<T>(path, { method: 'PATCH', body: data ? JSON.stringify(data) : undefined }),
    del: <T = any>(path: string) => request<T>(path, { method: 'DELETE' }),
    setOrgId: (id: string | null) => {
      orgId.value = id
      if (import.meta.client) {
        if (id) localStorage.setItem('auth:org_id', id)
        else localStorage.removeItem('auth:org_id')
      }
    },
  }
}
