<script setup>
definePageMeta({ layout: 'dashboard' })

const { user, orgId, token } = useAuth()
const { get, setOrgId } = useApi()
const toast = useToast()

const loading = ref(true)
const tasks = ref([])
const projects = ref([])

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
  fetchAllTasks()
})

async function fetchAllTasks() {
  loading.value = true
  try {
    syncOrgId()
    const [allTasks, allProjects] = await Promise.all([
      get('/api/v1/my-tasks'),
      get('/api/v1/projects')
    ])
    projects.value = allProjects
    const projMap = Object.fromEntries(allProjects.map(p => [p.id, p]))
    tasks.value = allTasks.map(t => ({ ...t, project_name: projMap[t.project_id]?.name || '', project_slug: projMap[t.project_id]?.slug || '' }))
  } catch (e) {
    toast.error(e?.message || 'Gagal memuat tugas')
  } finally {
    loading.value = false
  }
}

const statusConfig = {
  backlog: { label: 'Backlog', dot: 'bg-gray-400 dark:bg-gray-500', bg: 'bg-gray-50 dark:bg-gray-700', text: 'text-gray-600 dark:text-gray-400' },
  todo: { label: 'Todo', dot: 'bg-blue-500', bg: 'bg-blue-50 dark:bg-blue-900/30', text: 'text-blue-600 dark:text-blue-400' },
  in_progress: { label: 'In Progress', dot: 'bg-amber-500', bg: 'bg-amber-50 dark:bg-amber-900/30', text: 'text-amber-600 dark:text-amber-400' },
  in_review: { label: 'In Review', dot: 'bg-purple-500', bg: 'bg-purple-50 dark:bg-purple-900/30', text: 'text-purple-600 dark:text-purple-400' },
  done: { label: 'Done', dot: 'bg-emerald-500', bg: 'bg-emerald-50 dark:bg-emerald-900/30', text: 'text-emerald-600 dark:text-emerald-400' },
}

const priorityConfig = {
  urgent: { label: 'Urgent', dot: 'bg-red-500', bg: 'bg-red-50 dark:bg-red-900/20', text: 'text-red-600 dark:text-red-400' },
  high: { label: 'High', dot: 'bg-orange-400', bg: 'bg-orange-50 dark:bg-orange-900/20', text: 'text-orange-600 dark:text-orange-400' },
  medium: { label: 'Medium', dot: 'bg-gray-400', bg: 'bg-gray-50 dark:bg-gray-700', text: 'text-gray-600 dark:text-gray-400' },
  low: { label: 'Low', dot: 'bg-gray-300', bg: 'bg-gray-50 dark:bg-gray-700', text: 'text-gray-500 dark:text-gray-400' },
}

const filterStatus = ref('all')
const filterPriority = ref('all')
const searchQuery = ref('')

const filteredTasks = computed(() => {
  return tasks.value.filter(t => {
    if (filterStatus.value !== 'all' && t.status !== filterStatus.value) return false
    if (filterPriority.value !== 'all' && t.priority !== filterPriority.value) return false
    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase()
      return (t.title || '').toLowerCase().includes(q) || (t.key || '').toLowerCase().includes(q)
    }
    return true
  })
})

const taskStats = computed(() => ({
  total: tasks.value.length,
  active: tasks.value.filter(t => t.status !== 'done' && t.status !== 'backlog').length,
  completed: tasks.value.filter(t => t.status === 'done').length,
}))

function isOverdue(dateStr) {
  if (!dateStr) return false
  return new Date(dateStr) < new Date()
}

function getProjectName(projectId) {
  return projects.value.find(p => p.id === projectId)?.name || '-'
}
</script>

<template>
  <div class="p-4 lg:p-6 max-w-7xl mx-auto w-full">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white">Tugas Saya</h1>
        <p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">{{ filteredTasks.length }} dari {{ tasks.length }} tugas ditugaskan kepada Anda.</p>
      </div>
    </div>

    <!-- Quick Stats -->
    <div class="grid grid-cols-3 gap-4 mb-6">
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-3 text-center">
        <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ taskStats.total }}</p>
        <p class="text-xs text-gray-500 dark:text-gray-400">Total</p>
      </div>
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-3 text-center">
        <p class="text-2xl font-bold text-amber-600">{{ taskStats.active }}</p>
        <p class="text-xs text-gray-500 dark:text-gray-400">Aktif</p>
      </div>
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-3 text-center">
        <p class="text-2xl font-bold text-emerald-600">{{ taskStats.completed }}</p>
        <p class="text-xs text-gray-500 dark:text-gray-400">Selesai</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex items-center gap-3 mb-5 flex-wrap">
      <div class="relative flex-1 min-w-[200px] max-w-xs">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
        <input v-model="searchQuery" type="text" placeholder="Cari tugas..." class="w-full pl-9 pr-3 py-2 text-sm border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all" />
      </div>
      <select v-model="filterStatus" class="text-sm border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all cursor-pointer">
        <option value="all">Semua Status</option>
        <option value="backlog">Backlog</option>
        <option value="todo">Todo</option>
        <option value="in_progress">In Progress</option>
        <option value="in_review">In Review</option>
        <option value="done">Done</option>
      </select>
      <select v-model="filterPriority" class="text-sm border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all cursor-pointer">
        <option value="all">Semua Prioritas</option>
        <option value="urgent">Urgent</option>
        <option value="high">High</option>
        <option value="medium">Medium</option>
        <option value="low">Low</option>
      </select>
    </div>

    <!-- Skeleton / Task List / Empty State -->
    <Transition name="fade" mode="out-in">
      <!-- Skeleton -->
      <div v-if="loading" key="tasks-skeleton" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
        <div class="divide-y divide-gray-50 dark:divide-gray-700">
          <div v-for="i in 8" :key="i" class="px-4 py-3 flex items-center gap-4 animate-pulse">
            <div class="w-2 h-2 rounded-full bg-gray-200 dark:bg-gray-700"></div>
            <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-16"></div>
            <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded flex-1"></div>
            <div class="h-5 bg-gray-200 dark:bg-gray-700 rounded w-20"></div>
          </div>
        </div>
      </div>

      <!-- Task List -->
      <div v-else-if="filteredTasks.length > 0" key="tasks-data" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
        <table class="w-full">
          <thead>
            <tr class="border-b border-gray-100 dark:border-gray-700">
              <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5 w-8"></th>
              <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5">Tugas</th>
              <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5 hidden sm:table-cell">Proyek</th>
              <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5">Status</th>
              <th class="text-right text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5">Deadline</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50 dark:divide-gray-700">
            <tr v-for="task in filteredTasks" :key="task.id" class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-all cursor-pointer group" @click="navigateTo(`/tasks/${task.id}`)">
              <td class="px-4 py-3">
                <div class="w-2 h-2 rounded-full transition-transform group-hover:scale-150" :class="priorityConfig[task.priority]?.dot || 'bg-gray-400 dark:text-gray-500'"></div>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <span class="text-xs font-mono text-gray-400 dark:text-gray-500 shrink-0">{{ task.key || 'TG-' + String(task.id).slice(-2) }}</span>
                  <span class="text-sm text-gray-900 dark:text-white truncate group-hover:text-indigo-600 transition-colors">{{ task.title }}</span>
                </div>
              </td>
              <td class="px-4 py-3 hidden sm:table-cell">
                <span class="text-xs text-gray-500 dark:text-gray-400">{{ getProjectName(task.project_id) }}</span>
              </td>
              <td class="px-4 py-3">
                <span class="text-[11px] font-medium inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full" :class="[statusConfig[task.status]?.bg || 'bg-gray-50 dark:bg-gray-700', statusConfig[task.status]?.text || 'text-gray-600 dark:text-gray-400']">
                  <span class="w-1.5 h-1.5 rounded-full shrink-0" :class="statusConfig[task.status]?.dot || 'bg-gray-400 dark:text-gray-500'"></span>
                  {{ statusConfig[task.status]?.label || task.status }}
                </span>
              </td>
              <td class="px-4 py-3 text-right">
                <span v-if="task.due_date" class="text-xs" :class="isOverdue(task.due_date) && task.status !== 'done' ? 'text-red-500 dark:text-red-400 font-medium' : 'text-gray-400 dark:text-gray-500'">
                  {{ new Date(task.due_date).toLocaleDateString('id-ID', { month: 'short', day: 'numeric' }) }}
                </span>
                <span v-else class="text-xs text-gray-300 dark:text-gray-600">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Empty State -->
      <div v-else key="tasks-empty" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-12 text-center">
        <svg class="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" /></svg>
        <p class="text-sm font-medium text-gray-900 dark:text-white mb-1">Tidak ada tugas ditemukan</p>
        <p class="text-xs text-gray-500 dark:text-gray-400">Coba ubah filter atau kata kunci pencarian Anda.</p>
      </div>
    </Transition>
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
