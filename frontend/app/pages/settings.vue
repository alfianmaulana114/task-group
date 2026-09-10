<script setup>
definePageMeta({ layout: 'dashboard' })

const { user, orgId, token } = useAuth()
const { get, patch, setOrgId } = useApi()
const toast = useToast()

const loading = ref(true)
const saving = ref(false)
const org = ref(null)
const activeTab = ref('organization')

const orgForm = reactive({ name: '', slug: '' })
const profileForm = reactive({ full_name: '' })
const savingProfile = ref(false)

function syncOrgId() {
  const oid = orgId.value || (import.meta.client ? localStorage.getItem('auth:org_id') : null)
  if (oid) setOrgId(oid)
  return oid
}

onMounted(() => {
  syncOrgId()
  watch(orgId, syncOrgId)
  const hasToken = token.value || (import.meta.client ? localStorage.getItem('auth:token') : null)
  if (!hasToken) {
    loading.value = false
    return
  }
  fetchOrg()
  if (user.value) {
    profileForm.full_name = user.value.full_name || ''
  }
  watch(user, (u) => {
    if (u) profileForm.full_name = u.full_name || ''
  })
})

async function fetchOrg() {
  loading.value = true
  try {
    syncOrgId()
    const orgs = await get('/api/v1/organizations')
    const effectiveOrgId = orgId.value || localStorage.getItem('auth:org_id')
    org.value = orgs.find(o => o.id === effectiveOrgId)
    if (org.value) {
      orgForm.name = org.value.name
      orgForm.slug = org.value.slug
    }
  } catch (e) {
    toast.error(e?.message || 'Gagal memuat pengaturan')
  } finally {
    loading.value = false
  }
}

async function saveOrg() {
  if (!orgForm.name.trim()) return
  const effectiveOrgId = syncOrgId()
  if (!effectiveOrgId) {
    toast.error('Organisasi belum dipilih')
    return
  }
  saving.value = true
  try {
    await patch(`/api/v1/organizations/${effectiveOrgId}`, { name: orgForm.name })
    toast.success('Pengaturan tersimpan!')
  } catch (e) {
    toast.error(e?.message || 'Gagal menyimpan pengaturan')
  } finally {
    saving.value = false
  }
}

async function saveProfile() {
  if (!profileForm.full_name.trim()) return
  savingProfile.value = true
  try {
    await patch('/api/v1/auth/me', { full_name: profileForm.full_name })
    toast.success('Profil tersimpan!')
  } catch (e) {
    toast.error(e?.message || 'Gagal menyimpan profil')
  } finally {
    savingProfile.value = false
  }
}
</script>

<template>
  <div class="p-4 lg:p-6 max-w-3xl mx-auto w-full">
    <div class="mb-6">
      <h1 class="text-xl font-bold text-gray-900 dark:text-white">Pengaturan</h1>
      <p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">Kelola pengaturan organisasi dan akun Anda.</p>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 border-b border-gray-200 dark:border-gray-700 mb-6">
      <button @click="activeTab = 'organization'" class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors" :class="activeTab === 'organization' ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'">
        Organisasi
      </button>
      <button @click="activeTab = 'profile'" class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors" :class="activeTab === 'profile' ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'">
        Profil
      </button>
    </div>

    <!-- Organization Settings -->
    <div v-if="activeTab === 'organization'" class="space-y-6">
      <Transition name="fade" mode="out-in">
        <div v-if="loading" :key="'skeleton'" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6 space-y-4 animate-pulse">
          <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-32"></div>
          <div class="h-10 bg-gray-100 dark:bg-gray-700 rounded w-full"></div>
          <div class="h-10 bg-gray-100 dark:bg-gray-700 rounded w-full"></div>
        </div>

        <div v-else :key="'data'">
        <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6">
          <h2 class="text-sm font-semibold text-gray-900 dark:text-white mb-4">Detail Organisasi</h2>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Nama Organisasi</label>
              <input v-model="orgForm.name" type="text" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Slug</label>
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-400 dark:text-gray-500">taskgroup.app/</span>
                <input :value="orgForm.slug" type="text" disabled class="flex-1 px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm bg-gray-50 dark:bg-gray-700 text-gray-500 dark:text-gray-400" />
              </div>
            </div>
            <div class="flex justify-end">
              <button @click="saveOrg" :disabled="saving" class="bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-all disabled:opacity-50 flex items-center gap-2">
                <svg v-if="saving" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
                Simpan
              </button>
            </div>
          </div>
        </div>
      </div>
      </Transition>
    </div>

    <!-- Profile Settings -->
    <div v-if="activeTab === 'profile'" class="space-y-6">
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6">
        <h2 class="text-sm font-semibold text-gray-900 dark:text-white mb-4">Informasi Profil</h2>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Nama Lengkap</label>
            <input v-model="profileForm.full_name" type="text" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
            <input :value="user?.email || ''" type="email" disabled class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm bg-gray-50 dark:bg-gray-700 text-gray-500 dark:text-gray-400" />
          </div>
          <div class="flex justify-end">
            <button @click="saveProfile" :disabled="savingProfile || !profileForm.full_name.trim()" class="bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-all disabled:opacity-50 flex items-center gap-2">
              <svg v-if="savingProfile" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
              Simpan
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}
.fade-enter-from {
  opacity: 0;
  transform: translateY(4px);
}
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
