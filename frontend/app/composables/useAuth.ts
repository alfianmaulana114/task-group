interface User {
  id: string
  email: string
  full_name: string
}

export function useAuth() {
  const api = useApi()
  const router = useRouter()

  const user = useState<User | null>('auth:user', () => null)
  const token = useState<string | null>('auth:token', () => null)
  const isAuthenticated = computed(() => !!token.value && !!user.value)

  function setAuth(userData: User, tokenValue: string) {
    user.value = userData
    token.value = tokenValue
    if (import.meta.client) {
      localStorage.setItem('auth:token', tokenValue)
    }
  }

  function clearAuth() {
    user.value = null
    token.value = null
    if (import.meta.client) {
      localStorage.removeItem('auth:token')
    }
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
    const data = await api.post<{ user: User; token: string }>('/api/v1/auth/login', {
      email,
      password,
    })
    setAuth(data.user, data.token)
    return data
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

  async function fetchUser() {
    try {
      const data = await api.get<User>('/api/v1/auth/me')
      user.value = data
      return data
    } catch {
      clearAuth()
      return null
    }
  }

  function initAuth() {
    if (import.meta.client) {
      const savedToken = localStorage.getItem('auth:token')
      if (savedToken) {
        token.value = savedToken
        fetchUser()
      }
    }
  }

  return {
    user,
    token,
    isAuthenticated,
    register,
    login,
    logout,
    fetchUser,
    initAuth,
  }
}
