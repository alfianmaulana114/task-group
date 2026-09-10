<script setup>
const { register } = useAuth()
const toast = useToast()
const router = useRouter()

const form = reactive({
  fullName: '',
  email: '',
  password: '',
  confirmPassword: '',
})

const loading = ref(false)
const error = ref('')
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const nameInput = ref(null)

const passwordStrength = computed(() => {
  const p = form.password
  if (!p) return { level: 0, label: '', color: '' }
  let score = 0
  if (p.length >= 8) score++
  if (p.length >= 12) score++
  if (/[a-z]/.test(p) && /[A-Z]/.test(p)) score++
  if (/\d/.test(p)) score++
  if (/[^a-zA-Z0-9]/.test(p)) score++

  if (score <= 1) return { level: 1, label: 'Lemah', color: 'bg-red-500' }
  if (score <= 2) return { level: 2, label: 'Sedang', color: 'bg-amber-500' }
  if (score <= 3) return { level: 3, label: 'Kuat', color: 'bg-blue-500' }
  return { level: 4, label: 'Sangat Kuat', color: 'bg-emerald-500' }
})

const passwordsMatch = computed(() => {
  if (!form.confirmPassword) return null
  return form.password === form.confirmPassword
})

onMounted(() => {
  nextTick(() => nameInput.value?.focus())
})

async function handleRegister() {
  error.value = ''

  if (!form.fullName || !form.email || !form.password) {
    error.value = 'Semua field wajib diisi'
    return
  }

  if (form.password.length < 8) {
    error.value = 'Password minimal 8 karakter'
    return
  }

  if (form.password !== form.confirmPassword) {
    error.value = 'Password tidak cocok'
    return
  }

  loading.value = true
  try {
    await register(form.email, form.password, form.fullName)
    toast.success('Akun berhasil dibuat!')
    router.push('/dashboard')
  } catch (e) {
    error.value = e.message || 'Pendaftaran gagal'
    toast.error(e.message || 'Pendaftaran gagal')
  } finally {
    loading.value = false
  }
}

definePageMeta({
  layout: false,
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
    <div class="sm:mx-auto sm:w-full sm:max-w-md">
      <NuxtLink to="/" class="flex justify-center items-center gap-2">
        <div class="w-10 h-10 bg-indigo-600 rounded-lg flex items-center justify-center">
          <span class="text-white font-bold text-lg">TG</span>
        </div>
        <span class="text-2xl font-bold text-gray-900">Task Group</span>
      </NuxtLink>
      <h2 class="mt-6 text-center text-3xl font-bold text-gray-900">
        Buat akun baru
      </h2>
      <p class="mt-2 text-center text-sm text-gray-600">
        sudah punya akun?
        <NuxtLink to="/login" class="font-medium text-indigo-600 hover:text-indigo-500 transition-colors">
          masuk sekarang
        </NuxtLink>
      </p>
    </div>

    <Transition name="page-enter" appear>
      <div class="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div class="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
        <form class="space-y-5" @submit.prevent="handleRegister">
          <Transition
            enter-active-class="transition-all duration-300 ease-out"
            enter-from-class="opacity-0 -translate-y-1"
            enter-to-class="opacity-100 translate-y-0"
            leave-active-class="transition-all duration-200 ease-in"
            leave-from-class="opacity-100 translate-y-0"
            leave-to-class="opacity-0 -translate-y-1"
          >
            <div v-if="error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm flex items-center gap-2">
              <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              {{ error }}
            </div>
          </Transition>

          <div>
            <label for="fullName" class="block text-sm font-medium text-gray-700">Nama Lengkap</label>
            <div class="mt-1 relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              </div>
              <input
                ref="nameInput"
                id="fullName"
                v-model="form.fullName"
                type="text"
                required
                autocomplete="name"
                class="block w-full pl-10 pr-3 py-2.5 border border-gray-300 rounded-lg shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm transition-all"
                placeholder="John Doe"
              />
            </div>
          </div>

          <div>
            <label for="email" class="block text-sm font-medium text-gray-700">Email</label>
            <div class="mt-1 relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9m4.5-1.206a8.959 8.959 0 01-4.5 1.207" />
                </svg>
              </div>
              <input
                id="email"
                v-model="form.email"
                type="email"
                required
                autocomplete="email"
                class="block w-full pl-10 pr-3 py-2.5 border border-gray-300 rounded-lg shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm transition-all"
                placeholder="anda@email.com"
              />
            </div>
          </div>

          <div>
            <label for="password" class="block text-sm font-medium text-gray-700">Password</label>
            <div class="mt-1 relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </div>
              <input
                id="password"
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                required
                autocomplete="new-password"
                class="block w-full pl-10 pr-10 py-2.5 border border-gray-300 rounded-lg shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm transition-all"
                placeholder="Minimal 8 karakter"
              />
              <button
                type="button"
                class="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600 transition-colors"
                @click="showPassword = !showPassword"
              >
                <svg v-if="!showPassword" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                </svg>
              </button>
            </div>
            <!-- Password strength indicator -->
            <Transition
              enter-active-class="transition-all duration-300"
              enter-from-class="opacity-0 h-0"
              enter-to-class="opacity-100 h-auto"
            >
              <div v-if="form.password" class="mt-2 space-y-1">
                <div class="flex gap-1">
                  <div v-for="i in 4" :key="i" class="h-1 flex-1 rounded-full bg-gray-200 overflow-hidden">
                    <div
                      class="h-full rounded-full transition-all duration-300"
                      :class="i <= passwordStrength.level ? passwordStrength.color : ''"
                    ></div>
                  </div>
                </div>
                <p class="text-[11px] font-medium" :class="{ 'text-red-500': passwordStrength.level <= 1, 'text-amber-500': passwordStrength.level === 2, 'text-blue-500': passwordStrength.level === 3, 'text-emerald-500': passwordStrength.level === 4 }">
                  Kekuatan: {{ passwordStrength.label }}
                </p>
              </div>
            </Transition>
          </div>

          <div>
            <label for="confirmPassword" class="block text-sm font-medium text-gray-700">Konfirmasi Password</label>
            <div class="mt-1 relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
              </div>
              <input
                id="confirmPassword"
                v-model="form.confirmPassword"
                :type="showConfirmPassword ? 'text' : 'password'"
                required
                autocomplete="new-password"
                class="block w-full pl-10 pr-10 py-2.5 border rounded-lg shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 sm:text-sm transition-all"
                :class="passwordsMatch === false ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : passwordsMatch === true ? 'border-emerald-300 focus:ring-emerald-500 focus:border-emerald-500' : 'border-gray-300 focus:ring-indigo-500 focus:border-indigo-500'"
                placeholder="Ulangi password"
              />
              <button
                type="button"
                class="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600 transition-colors"
                @click="showConfirmPassword = !showConfirmPassword"
              >
                <svg v-if="!showConfirmPassword" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                </svg>
              </button>
            </div>
            <Transition
              enter-active-class="transition-all duration-200"
              enter-from-class="opacity-0"
              enter-to-class="opacity-100"
            >
              <p v-if="passwordsMatch === false" class="mt-1 text-[11px] text-red-500 font-medium">Password tidak cocok</p>
              <p v-else-if="passwordsMatch === true" class="mt-1 text-[11px] text-emerald-500 font-medium">Password cocok</p>
            </Transition>
          </div>

          <div>
            <button
              type="submit"
              :disabled="loading"
              class="w-full flex justify-center items-center gap-2 py-2.5 px-4 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
            >
              <svg v-if="loading" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
              <span v-if="loading">Membuat akun...</span>
              <span v-else>Buat akun</span>
            </button>
          </div>
        </form>
      </div>
      </div>
    </Transition>
  </div>
</template>

<style>
.page-enter-active {
  transition: opacity 0.4s ease, transform 0.4s ease;
}
.page-enter-from {
  opacity: 0;
  transform: translateY(12px);
}
</style>
