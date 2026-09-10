<script setup>
definePageMeta({ layout: 'dashboard' })

const { user, orgId, orgRole, token } = useAuth()
const { get, post, setOrgId } = useApi()
const toast = useToast()
const router = useRouter()

const loading = ref(true)
const projects = ref([])
const viewMode = ref('grid')
const searchQuery = ref('')
const showCreate = ref(false)
const creating = ref(false)
const newProject = ref({ name: '', description: '' })
const canCreateProject = computed(() => ['owner', 'admin'].includes(orgRole.value))

function syncOrgId() {
  const oid = orgId.value || (import.meta.client ? localStorage.getItem('auth:org_id') : null)
  if (oid) setOrgId(oid)
  return oid
}

onMounted(() => {
  syncOrgId()
  // watch orgId dari useAuth (computed dari user) -> sync ke useApi
  watch(orgId, syncOrgId)
  const hasToken = token.value || (import.meta.client ? localStorage.getItem('auth:token') : null)
  if (!hasToken) {
    loading.value = false
    // biarkan layout/dashboard handle redirect, tapi jangan fetch
    return
  }
  fetchProjects()
})

async function fetchProjects() {
  loading.value = true
  try {
    syncOrgId()
    projects.value = await get('/api/v1/projects')
  } catch (e) {
    const msg = e?.message || 'Gagal memuat proyek'
    toast.error(msg)
  } finally {
    loading.value = false
  }
}

async function createProject() {
  if (!newProject.value.name.trim()) return
  // Guard: pastikan token & orgId ada sebelum hit backend/internal/auth/middleware.go & tenant.go
  const hasToken = token.value || (import.meta.client ? localStorage.getItem('auth:token') : null)
  if (!hasToken) {
    toast.error('Sesi telah berakhir, silakan login kembali.')
    router.push('/login')
    return
  }
  const oid = syncOrgId()
  if (!oid) {
    toast.error('Organisasi belum dipilih. Silakan pilih organisasi dulu.')
    router.push('/organizations')
    return
  }
  creating.value = true
  try {
    const p = await post('/api/v1/projects', {
      name: newProject.value.name,
      description: newProject.value.description,
    })
    projects.value.unshift(p)
    showCreate.value = false
    newProject.value = { name: '', description: '' }
    toast.success('Proyek berhasil dibuat!')
  } catch (e) {
    const msg = e?.message || 'Gagal membuat proyek'
    toast.error(msg)
  } finally {
    creating.value = false
  }
}

const filteredProjects = computed(() => {
  if (!searchQuery.value) return projects.value
  const q = searchQuery.value.toLowerCase()
  return projects.value.filter(p => p.name.toLowerCase().includes(q))
})

function getProjectColor(i) {
  return ['bg-indigo-500', 'bg-emerald-500', 'bg-amber-500', 'bg-rose-500', 'bg-sky-500'][i % 5]
}

const avatarColors = ['bg-violet-500', 'bg-sky-500', 'bg-emerald-500', 'bg-rose-500', 'bg-amber-500']
</script>

<template>
  <div class="p-4 lg:p-6 max-w-7xl mx-auto w-full dark:bg-gray-900">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white">Proyek</h1>
        <p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">Kelola semua proyek tim Anda.</p>
      </div>
      <button v-if="canCreateProject" @click="showCreate = true" class="bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-all flex items-center gap-2 hover:shadow-sm active:scale-[0.97]">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
        Proyek Baru
      </button>
    </div>

    <!-- Filters -->
    <div class="flex items-center gap-3 mb-5">
      <div class="flex-1 max-w-sm relative">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
        <input v-model="searchQuery" type="text" placeholder="Cari proyek..." class="w-full pl-10 pr-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all" />
      </div>
      <div class="flex bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
        <button @click="viewMode = 'grid'" class="p-2 transition-colors" :class="viewMode === 'grid' ? 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-300' : 'text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300'">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" /></svg>
        </button>
        <button @click="viewMode = 'list'" class="p-2 transition-colors" :class="viewMode === 'list' ? 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-300' : 'text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300'">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" /></svg>
        </button>
      </div>
    </div>

    <Transition name="fade" mode="out-in">
      <!-- Skeleton -->
      <div v-if="loading" :class="viewMode === 'grid' ? 'grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4' : ''" :key="'skeleton'">
        <template v-if="viewMode === 'grid'">
          <div v-for="i in 6" :key="i" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 animate-pulse">
            <div class="flex items-center gap-2 mb-3"><div class="w-3 h-3 rounded-sm bg-gray-200 dark:bg-gray-600"></div><div class="h-4 bg-gray-200 dark:bg-gray-600 rounded w-32"></div></div>
            <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded w-full mb-2"></div>
            <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded w-2/3 mb-4"></div>
            <div class="h-1 bg-gray-100 dark:bg-gray-700 rounded-full w-full mb-3"></div>
            <div class="flex justify-between"><div class="h-2 bg-gray-100 dark:bg-gray-700 rounded w-16"></div><div class="flex gap-1"><div class="w-6 h-6 rounded-full bg-gray-200 dark:bg-gray-600"></div><div class="w-6 h-6 rounded-full bg-gray-200 dark:bg-gray-600"></div></div></div>
          </div>
        </template>
        <template v-else>
          <div v-for="i in 4" :key="i" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 animate-pulse flex items-center gap-4">
            <div class="w-3 h-3 rounded-sm bg-gray-200 dark:bg-gray-600"></div>
            <div class="h-4 bg-gray-200 dark:bg-gray-600 rounded w-40 flex-1"></div>
            <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded w-12"></div>
          </div>
        </template>
      </div>

      <div v-else :key="'data'">
        <!-- Empty State -->
        <div v-if="projects.length === 0" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-12 text-center">
          <svg class="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" /></svg>
          <p class="text-sm font-medium text-gray-900 dark:text-white mb-1">Belum ada proyek</p>
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-4">Buat proyek pertama untuk memulai.</p>
          <button v-if="canCreateProject" @click="showCreate = true" class="bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium px-4 py-2 rounded-lg transition-all">Buat Proyek</button>
        </div>

        <!-- Grid View -->
        <div v-else-if="viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      <NuxtLink
        v-for="(project, i) in filteredProjects"
        :key="project.id"
        :to="`/projects/${project.slug}`"
        class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 hover:shadow-md hover:border-indigo-200 dark:hover:border-indigo-800 transition-all duration-200 group"
      >
        <div class="flex items-start justify-between">
          <div class="flex items-center gap-2.5">
            <div class="w-3 h-3 rounded-sm group-hover:scale-125 transition-transform" :class="getProjectColor(i)"></div>
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">{{ project.name }}</h3>
          </div>
          <span v-if="project.status === 'archived'" class="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 font-medium">Archived</span>
        </div>
        <p v-if="project.description" class="mt-2 text-xs text-gray-500 dark:text-gray-400 line-clamp-2">{{ project.description }}</p>
        <div class="mt-4 flex items-center justify-between">
          <span class="text-[11px] text-gray-400 dark:text-gray-500">{{ new Date(project.created_at).toLocaleDateString('id-ID', { month: 'short', day: 'numeric', year: 'numeric' }) }}</span>
          <div class="w-5 h-5 rounded-full flex items-center justify-center" :class="avatarColors[i % avatarColors.length]">
            <span class="text-[7px] text-white font-bold">{{ user?.full_name?.charAt(0) || 'U' }}</span>
          </div>
        </div>
      </NuxtLink>
    </div>

    <!-- List View -->
    <div v-else class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
      <table class="w-full">
        <thead>
          <tr class="border-b border-gray-100 dark:border-gray-700">
            <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5">Proyek</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5 hidden sm:table-cell">Status</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5 hidden md:table-cell">Dibuat</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-50 dark:divide-gray-700">
          <NuxtLink v-for="(project, i) in filteredProjects" :key="project.id" :to="`/projects/${project.slug}`" custom v-slot="{ navigate }">
            <tr @click="navigate" class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer">
              <td class="px-4 py-3">
                <div class="flex items-center gap-2.5">
                  <div class="w-2.5 h-2.5 rounded-sm shrink-0" :class="getProjectColor(i)"></div>
                  <div>
                    <div class="text-sm font-medium text-gray-900 dark:text-white">{{ project.name }}</div>
                    <div v-if="project.description" class="text-xs text-gray-400 dark:text-gray-500 truncate max-w-xs">{{ project.description }}</div>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3 hidden sm:table-cell">
                <span class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="project.status === 'active' ? 'bg-emerald-50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400' : 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400'">{{ project.status }}</span>
              </td>
              <td class="px-4 py-3 hidden md:table-cell">
                <span class="text-xs text-gray-400 dark:text-gray-500">{{ new Date(project.created_at).toLocaleDateString('id-ID', { month: 'short', day: 'numeric', year: 'numeric' }) }}</span>
              </td>
            </tr>
          </NuxtLink>
        </tbody>
      </table>
    </div>
      </div>
    </Transition>

    <!-- Create Project Modal -->
    <Teleport to="body">
      <Transition enter-active-class="transition-all duration-200 ease-out" leave-active-class="transition-all duration-150 ease-in" enter-from-class="opacity-0 scale-95" leave-to-class="opacity-0 scale-95">
        <div v-if="showCreate" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showCreate = false"></div>
          <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-sm overflow-hidden">
            <div class="px-5 py-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
              <h3 class="text-sm font-bold text-gray-900 dark:text-white">Proyek Baru</h3>
              <button @click="showCreate = false" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <div class="px-5 py-4 space-y-3">
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Nama <span class="text-red-500 dark:text-red-400">*</span></label>
                <input v-model="newProject.name" type="text" placeholder="Nama proyek..." class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all" @keydown.enter="createProject" />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Deskripsi</label>
                <textarea v-model="newProject.description" rows="2" placeholder="Deskripsi singkat..." class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none transition-all"></textarea>
              </div>
            </div>
            <div class="px-5 py-3 border-t border-gray-100 dark:border-gray-700 flex justify-end gap-2">
              <button @click="showCreate = false" class="px-3 py-1.5 text-xs font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg transition-colors">Batal</button>
              <button @click="createProject" :disabled="!newProject.name.trim() || creating" class="px-3 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-all disabled:opacity-50 flex items-center gap-1.5">
                <svg v-if="creating" class="animate-spin w-3 h-3" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
                Buat
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
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
