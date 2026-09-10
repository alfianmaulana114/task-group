<script setup>
definePageMeta({ layout: 'dashboard' })

const auth = useAuth()
const user = auth.user
const orgId = auth.orgId
const orgRole = auth.orgRole
const token = auth.token
const { get, post, setOrgId } = useApi()
const { fetchUser } = auth
const router = useRouter()
const toast = useToast()

const organizations = ref([])
const selectedOrg = ref(null)
const showCreateForm = ref(false)
const newOrg = ref({ name: '', slug: '' })
const searchQuery = ref('')
const loading = ref(true)

const filteredOrgs = computed(() => {
  if (!searchQuery.value) return organizations.value
  return organizations.value.filter(o =>
    o.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    o.slug.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

function syncOrgId() {
  const oid = orgId.value || (import.meta.client ? localStorage.getItem('auth:org_id') : null)
  if (oid) setOrgId(oid)
  return oid
}

async function loadOrgs() {
  loading.value = true
  try {
    syncOrgId()
    const data = await get('/api/v1/organizations')
    organizations.value = data || []
  } catch (e) {
    console.error('Failed to fetch organizations:', e)
    toast.error(e?.message || 'Gagal memuat organisasi')
    organizations.value = []
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  syncOrgId()
  watch(orgId, syncOrgId)
  const hasToken = token.value || (import.meta.client ? localStorage.getItem('auth:token') : null)
  if (!hasToken) {
    loading.value = false
    router.push('/login')
    return
  }
  // ensure user is loaded if initAuth only set temp user
  if (!user.value?.email) {
    await fetchUser()
  }
  await loadOrgs()
  handleAutoSelection()
})

function handleAutoSelection() {
  const orgFromStorage = import.meta.client ? localStorage.getItem('auth:org_id') : null
  if (!orgFromStorage && organizations.value.length === 1) {
    setOrgSelection(organizations.value[0])
  } else if (!orgFromStorage && organizations.value.length === 0) {
    showCreateForm.value = true
  } else if (orgFromStorage) {
    const org = organizations.value.find(o => o.id === orgFromStorage)
    if (org) selectedOrg.value = org
  }
}

function setOrgSelection(org) {
  // pakai auth.setAuth yang benar (jangan buat useAuth baru)
  const userData = {
    id: user.value?.id || 'temp',
    email: user.value?.email || '',
    full_name: user.value?.full_name || '',
    org_id: org.id,
    org_role: org.role || orgRole.value || 'member',
  }
  // token dari state yang sudah ada
  const currentToken = token.value || (import.meta.client ? localStorage.getItem('auth:token') : '')
  if (!currentToken) {
    toast.error('Sesi berakhir, silakan login kembali')
    router.push('/login')
    return
  }
  auth.user.value = userData
  if (import.meta.client) {
    localStorage.setItem('auth:org_id', org.id)
    localStorage.setItem('auth:org_role', userData.org_role)
  }
  setOrgId(org.id)
  toast.success(`Organisasi ${org.name} dipilih`)
  router.push('/dashboard')
}

async function createOrganization() {
  if (!newOrg.value.name.trim()) return
  const currentToken = token.value || (import.meta.client ? localStorage.getItem('auth:token') : null)
  if (!currentToken) {
    toast.error('Sesi berakhir, silakan login')
    router.push('/login')
    return
  }
  try {
    syncOrgId()
    const org = await post('/api/v1/organizations', {
      name: newOrg.value.name,
      slug: newOrg.value.slug,
    })
    toast.success('Organisasi berhasil dibuat!')
    setOrgSelection(org)
  } catch (e) {
    toast.error(e?.message || 'Gagal membuat organisasi')
  }
}

function orgAvatarColor(org) {
  const colors = ['bg-indigo-500', 'bg-emerald-500', 'bg-amber-500', 'bg-rose-500', 'bg-sky-500']
  return colors[org.id.length % colors.length]
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex flex-col items-center py-12">
    <NuxtLink to="/dashboard" class="mb-6 text-indigo-600 hover:text-indigo-700 text-sm font-medium transition-colors">
      ← Kembali ke Dashboard
    </NuxtLink>
    
    <div class="bg-white rounded-xl shadow-lg w-full max-w-md p-6 border border-gray-200">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-8 h-8 rounded-lg flex items-center justify-center" :class="selectedOrg?.avatarColor || 'bg-indigo-500'">
          <span class="text-white font-bold text-lg">{{ selectedOrg?.name?.charAt(0) || 'T' }}</span>
        </div>
        <div>
          <h1 class="text-xl font-bold text-gray-900">{{ selectedOrg?.name || 'Pilih Organisasi' }}</h1>
          <p class="text-sm text-gray-500">{{ selectedOrg?.description || 'Silakan pilih organisasi untuk melanjutkan' }}</p>
        </div>
      </div>

      <div v-if="loading" class="space-y-2 animate-pulse">
        <div v-for="i in 3" :key="i" class="h-12 bg-gray-100 rounded-lg"></div>
      </div>

      <!-- Organizations List -->
      <div v-else-if="organizations.length > 0" class="space-y-3 mb-6">
        <p class="text-xs text-gray-400 uppercase tracking-wider mb-2">Organisasi Anda</p>
        <div class="space-y-2">
          <button
            v-for="org in filteredOrgs"
            :key="org.id"
            class="w-full group bg-gray-50 rounded-lg p-3 border border-gray-200 hover:border-indigo-300 hover:bg-indigo-50 transition-all cursor-pointer text-left flex items-center justify-between"
            @click="setOrgSelection(org)"
          >
            <div class="flex items-center gap-3">
              <div class="w-3 h-3 rounded-full" :class="orgAvatarColor(org)"></div>
              <div>
                <h3 class="font-medium text-gray-900 truncate">{{ org.name }}</h3>
                <p class="text-xs text-gray-500">{{ org.slug }}</p>
              </div>
            </div>
            <svg class="w-3.5 h-3.5 text-indigo-500 group-hover:text-indigo-600 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
          </button>
        </div>
      </div>

      <!-- Create Organization Form -->
      <div v-else-if="showCreateForm" class="mt-6">
        <h2 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-3">Buat Organisasi Baru</h2>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Nama Organisasi</label>
            <input
              v-model="newOrg.name"
              type="text"
              placeholder="Nama organisasi"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-300 focus:border-transparent"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Slug</label>
            <p class="text-xs text-gray-400">Slug akan digunakan untuk URL organisasi (contoh: taskgroup.app/{slug})</p>
            <input
              v-model="newOrg.slug"
              type="text"
              placeholder="taskgroup-app"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-300 focus:border-transparent"
            />
            <p v-if="!newOrg.slug" class="text-xs text-gray-300 mt-1">Minimal 3 karakter</p>
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-4">
          <button
            @click="showCreateForm = false"
            class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 rounded-lg transition-colors"
          >
            Batalkan
          </button>
          <button
            @click="createOrganization"
            :disabled="!newOrg.name.trim()"
            class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Buat Organisasi
          </button>
        </div>
      </div>

      <!-- No Organizations -->
      <div v-else class="mt-6 text-center text-gray-500">
        <p class="text-sm">Belum ada organisasi</p>
        <button @click="showCreateForm = true" class="mt-2 text-indigo-600 hover:text-indigo-700 underline">Buat organisasi pertama</button>
      </div>
    </div>
  </div>
</template>
