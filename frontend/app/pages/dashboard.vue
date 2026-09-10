<script setup>
definePageMeta({ layout: 'dashboard' })

const { user, orgId, token } = useAuth()
const { get, setOrgId } = useApi()
const toast = useToast()

const loading = ref(true)
const projects = ref([])
const allTasks = ref([])

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
  fetchData()
})

async function fetchData() {
  loading.value = true
  try {
    syncOrgId()
    projects.value = await get('/api/v1/projects')
    // Fetch tasks from all projects
    const taskPromises = projects.value.map(p => get(`/api/v1/projects/${p.id}/tasks`).catch(() => []))
    const taskArrays = await Promise.all(taskPromises)
    allTasks.value = taskArrays.flat()
  } catch (e) {
    toast.error(e?.message || 'Gagal memuat dashboard')
  } finally {
    loading.value = false
  }
}

const myTasks = computed(() => allTasks.value.filter(t => t.assignee_id === user.value?.id).slice(0, 5))

const stats = computed(() => {
  const total = allTasks.value.length
  const active = allTasks.value.filter(t => t.status !== 'done' && t.status !== 'backlog').length
  const inProgress = allTasks.value.filter(t => t.status === 'in_progress').length
  const overdue = allTasks.value.filter(t => t.due_date && new Date(t.due_date) < new Date() && t.status !== 'done').length
  return [
    { label: 'Total Proyek', value: projects.value.length, icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z', color: 'bg-indigo-500', textColor: 'text-indigo-600', bgColor: 'bg-indigo-50' },
    { label: 'Total Tugas', value: total, icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2', color: 'bg-emerald-500', textColor: 'text-emerald-600', bgColor: 'bg-emerald-50' },
    { label: 'In Progress', value: inProgress, icon: 'M13 10V3L4 14h7v7l9-11h-7z', color: 'bg-amber-500', textColor: 'text-amber-600', bgColor: 'bg-amber-50' },
    { label: 'Terlambat', value: overdue, icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z', color: 'bg-red-500', textColor: 'text-red-600', bgColor: 'bg-red-50' },
  ]
})

const statusConfig = {
  backlog: { label: 'Backlog', dot: 'bg-gray-400' },
  todo: { label: 'Todo', dot: 'bg-blue-500' },
  in_progress: { label: 'In Progress', dot: 'bg-amber-500' },
  in_review: { label: 'In Review', dot: 'bg-purple-500' },
  done: { label: 'Done', dot: 'bg-emerald-500' },
}

const priorityConfig = {
  urgent: 'bg-red-500',
  high: 'bg-orange-400',
  medium: 'bg-gray-400',
  low: 'bg-gray-300',
}

function getProjectColor(i) {
  return ['bg-indigo-500', 'bg-emerald-500', 'bg-amber-500', 'bg-rose-500', 'bg-sky-500'][i % 5]
}

function isOverdue(dateStr) {
  if (!dateStr) return false
  return new Date(dateStr) < new Date()
}

const avatarColors = ['bg-violet-500', 'bg-sky-500', 'bg-emerald-500', 'bg-rose-500', 'bg-amber-500']
</script>

<template>
  <div class="p-4 lg:p-6 max-w-7xl mx-auto space-y-6 w-full">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
        <p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">Selamat datang kembali, {{ user?.full_name || 'User' }}. Berikut ringkasan hari ini.</p>
      </div>
    </div>

    <!-- Stats Cards -->
    <Transition name="fade" mode="out-in">
      <div v-if="loading" key="stats-skeleton" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div v-for="i in 4" :key="i" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 animate-pulse">
          <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-20 mb-3"></div>
          <div class="h-8 bg-gray-200 dark:bg-gray-700 rounded w-12"></div>
        </div>
      </div>
      <div v-else key="stats-data" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div v-for="(stat, i) in stats" :key="i" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 hover:shadow-md transition-all group">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg flex items-center justify-center" :class="stat.bgColor">
              <svg class="w-5 h-5" :class="stat.textColor" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="stat.icon" /></svg>
            </div>
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400 font-medium">{{ stat.label }}</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ stat.value }}</p>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- My Tasks -->
    <section>
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-sm font-semibold text-gray-900 dark:text-white">Tugas Saya</h2>
        <NuxtLink to="/my-tasks" class="text-xs font-medium text-indigo-600 hover:text-indigo-700 transition-colors flex items-center gap-1">
          Lihat semua
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
        </NuxtLink>
      </div>

      <Transition name="fade" mode="out-in">
        <div v-if="loading" key="tasks-skeleton" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
          <div class="divide-y divide-gray-50 dark:divide-gray-700">
            <div v-for="i in 5" :key="i" class="px-4 py-3 flex items-center gap-4 animate-pulse">
              <div class="w-2 h-2 rounded-full bg-gray-200 dark:bg-gray-700"></div>
              <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-16"></div>
              <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded flex-1"></div>
              <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-20"></div>
            </div>
          </div>
        </div>

        <div v-else-if="myTasks.length === 0" key="tasks-empty" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-8 text-center">
          <p class="text-sm text-gray-400 dark:text-gray-500">Tidak ada tugas yang ditugaskan kepada Anda.</p>
        </div>

        <div v-else key="tasks-data" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-100 dark:border-gray-700">
                <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5">Tugas</th>
                <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5 hidden sm:table-cell">Proyek</th>
                <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5">Status</th>
                <th class="text-right text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5">Deadline</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-50 dark:divide-gray-700">
              <tr v-for="task in myTasks" :key="task.id" class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer group" @click="navigateTo(`/tasks/${task.id}`)">
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2.5">
                    <div class="w-2 h-2 rounded-full shrink-0 transition-transform group-hover:scale-150" :class="priorityConfig[task.priority] || 'bg-gray-400'"></div>
                    <span class="text-xs font-mono text-gray-400 dark:text-gray-500 shrink-0">{{ task.key || 'TG-' + String(task.id).slice(-2) }}</span>
                    <span class="text-sm text-gray-900 dark:text-white truncate group-hover:text-indigo-600 transition-colors">{{ task.title }}</span>
                  </div>
                </td>
                <td class="px-4 py-3 hidden sm:table-cell">
                  <span class="text-xs text-gray-500 dark:text-gray-400">{{ projects.find(p => p.id === task.project_id)?.name || '-' }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-xs text-gray-500 dark:text-gray-400 inline-flex items-center gap-1.5">
                    <span class="w-1.5 h-1.5 rounded-full shrink-0" :class="statusConfig[task.status]?.dot || 'bg-gray-400'"></span>
                    {{ statusConfig[task.status]?.label || task.status }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right">
                  <span v-if="task.due_date" class="text-xs" :class="isOverdue(task.due_date) ? 'text-red-500 dark:text-red-400 font-medium' : 'text-gray-400 dark:text-gray-500'">
                    {{ new Date(task.due_date).toLocaleDateString('id-ID', { month: 'short', day: 'numeric' }) }}
                  </span>
                  <span v-else class="text-xs text-gray-300 dark:text-gray-600">—</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </Transition>
    </section>

    <!-- Projects -->
    <section>
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-sm font-semibold text-gray-900 dark:text-white">Proyek</h2>
        <NuxtLink to="/projects" class="text-xs font-medium text-indigo-600 hover:text-indigo-700 transition-colors flex items-center gap-1">
          Lihat semua
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
        </NuxtLink>
      </div>

      <Transition name="fade" mode="out-in">
        <div v-if="loading" key="projects-skeleton" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div v-for="i in 3" :key="i" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 animate-pulse">
            <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-32 mb-2"></div>
            <div class="h-2.5 bg-gray-100 dark:bg-gray-700 rounded w-full mb-3"></div>
            <div class="h-1 bg-gray-100 dark:bg-gray-700 rounded-full w-full"></div>
          </div>
        </div>

        <div v-else-if="projects.length === 0" key="projects-empty" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-8 text-center">
          <p class="text-sm text-gray-400 dark:text-gray-500">Belum ada proyek. Buat proyek pertama Anda!</p>
        </div>

        <div v-else key="projects-data" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <NuxtLink
            v-for="(project, i) in projects"
            :key="project.id"
            :to="`/projects/${project.slug}`"
            class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 hover:shadow-md hover:border-indigo-200 dark:hover:border-indigo-500 transition-all group"
          >
            <div class="flex items-center gap-2.5">
              <div class="w-2.5 h-2.5 rounded-sm group-hover:scale-125 transition-transform" :class="getProjectColor(i)"></div>
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white group-hover:text-indigo-600 transition-colors">{{ project.name }}</h3>
            </div>
            <p v-if="project.description" class="mt-1.5 text-xs text-gray-500 dark:text-gray-400 line-clamp-2">{{ project.description }}</p>
            <div class="mt-3 flex items-center justify-between">
              <span class="text-[11px] text-gray-400 dark:text-gray-500">{{ new Date(project.created_at).toLocaleDateString('id-ID', { month: 'short', day: 'numeric', year: 'numeric' }) }}</span>
              <div class="w-5 h-5 rounded-full flex items-center justify-center" :class="avatarColors[i % avatarColors.length]">
                <span class="text-[7px] text-white font-bold">{{ user?.full_name?.charAt(0) || 'U' }}</span>
              </div>
            </div>
          </NuxtLink>
        </div>
      </Transition>
    </section>
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
