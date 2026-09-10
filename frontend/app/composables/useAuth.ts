interface User {
  id: string
  email: string
  full_name: string
  org_id?: string
  org_role?: string
}

export interface Organization {
  id: string
  name: string
  slug: string
  plan: string
  status: string
  role?: string
}

import { useApi } from '@/composables/useApi'

export function useAuth() {
  const api = useApi()
  const router = useRouter()

  const user = useState<User | null>('auth:user', () => null)
  const token = useState<string | null>('auth:token', () => null)
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const orgId = computed(() => user.value?.org_id)
  const orgRole = computed(() => user.value?.org_role)

  function setAuth(userData: User, tokenValue: string) {
    user.value = userData
    token.value = tokenValue
    if (import.meta.client) {
      localStorage.setItem('auth:token', tokenValue)
      localStorage.setItem('auth:org_id', userData.org_id || '')
      localStorage.setItem('auth:org_role', userData.org_role || '')
      localStorage.setItem('auth:email', userData.email || '')
      localStorage.setItem('auth:full_name', userData.full_name || '')
      // Set org ID in API client for tenant scoping
      api.setOrgId(userData.org_id || '')
    }
  }

  function clearAuth() {
    user.value = null
    token.value = null
    api.setOrgId(null as any)
    if (import.meta.client) {
      localStorage.removeItem('auth:token')
      localStorage.removeItem('auth:org_id')
      localStorage.removeItem('auth:org_role')
      localStorage.removeItem('auth:email')
      localStorage.removeItem('auth:full_name')
    }
  }

  // Set default organization header for API calls
  function setOrgHeader(xOrgId: string) {
    // This will be handled by the api composable's request interceptor
    // We just store it for reference
  }

  async function register(email: string, password: string, fullName: string) {
    const data = await api.post<{ user: User; token: string }>('/api/v1/auth/register', {
      email,
      password,
      full_name: fullName,
    })
    setAuth(data.user, data.token)
    return data
  }

  async function login(email: string, password: string) {
    try {
      const data = await api.post<{ user: User; token: string }>('/api/v1/auth/login', {
        email,
        password,
      })
  
      // Verify token is present in response
      if (!data?.token) {
        throw new Error('Token tidak ada dalam respons login')
      }
      
      // Ensure token is a valid string
      const validToken: string = typeof data.token === 'string' ? data.token : String(data.token)
      
      setAuth(data.user, validToken)
      
      // After login, fetch organizations and redirect
      await fetchUserOrgs()
      return data
    } catch (error: any) {
      // Re-throw with user-friendly message
      throw new Error(error.message || 'Login gagal - periksa kredensial dan coba lagi')
    }
  }

  async function fetchUserOrgs() {
    try {
      if (!token.value) throw new Error('Token tidak tersedia - silakan login kembali')
      const orgs = await api.get<Organization[]>('/api/v1/organizations')
      if (orgs.length === 1) {
        setAuth({ ...user.value!, org_id: orgs[0].id, org_role: (orgs[0] as any).role }, token.value as string)
        router.push('/dashboard')
      } else if (orgs.length > 1) {
        router.push('/organizations')
      } else {
        router.push('/organizations')
      }
    } catch (e) {
      console.error('Failed to fetch organizations', e)
      clearAuth()
      router.push('/login')
    }
  }

  async function fetchUser() {
    try {
      const data = await api.get<User>('/api/v1/auth/me')
      user.value = { ...user.value, ...data }
      return data
    } catch {
      clearAuth()
      return null
    }
  }

  function initAuth() {
    if (import.meta.client) {
      const savedToken = localStorage.getItem('auth:token')
      const savedOrgId = localStorage.getItem('auth:org_id')
      const savedOrgRole = localStorage.getItem('auth:org_role')
      if (savedToken) {
        token.value = savedToken
        if (savedOrgId) {
          api.setOrgId(savedOrgId)
          // decode JWT to get real user id for temp user (avoid 'temp' filter bug)
          let jwtUserId: string | null = null
          try {
            const payload = savedToken.split('.')[1]
            const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')))
            jwtUserId = decoded.user_id || decoded.sub || null
          } catch {}
          if (!user.value) {
            user.value = {
              id: jwtUserId || 'temp',
              email: localStorage.getItem('auth:email') || '',
              full_name: localStorage.getItem('auth:full_name') || '',
              org_id: savedOrgId,
              org_role: savedOrgRole || 'member',
            }
          } else {
            user.value = { ...user.value, org_id: savedOrgId, org_role: savedOrgRole || user.value.org_role }
          }
          fetchUser().catch(() => {})
        } else {
          fetchUser()
        }
      }
    }
  }

  async function logout() {
    try {
      await api.post('/api/v1/auth/logout')
    } catch {
      // ignore logout errors
    }
    clearAuth()
    router.push('/login')
  }

  return {
    user,
    token,
    isAuthenticated,
    orgId,
    orgRole,
    register,
    login,
    logout,
    fetchUser,
    initAuth,
    fetchUserOrgs,
  }
}
