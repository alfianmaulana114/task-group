<script setup>
definePageMeta({ layout: 'dashboard' })

const { get, post, patch, setOrgId } = useApi()
const { orgId, token } = useAuth()
const toast = useToast()

const loading = ref(true)
const tasks = ref([])
const projects = ref([])
const currentDate = ref(new Date())
const viewMode = ref('month')
const selectedProjectId = ref('')

const showTaskDetail = ref(false)
const selectedTask = ref(null)
const showCreateModal = ref(false)
const createForm = ref({ title: '', description: '', priority: 'medium', project_id: '' })
const creating = ref(false)

function syncOrgId() {
  const oid = orgId.value || (import.meta.client ? localStorage.getItem('auth:org_id') : null)
  if (oid) setOrgId(oid)
  return oid
}

onMounted(() => {
  syncOrgId()
  watch(orgId, syncOrgId)
  const hasToken = token.value || (import.meta.client ? localStorage.getItem('auth:token') : null)
  if (!hasToken) { loading.value = false; return }
  fetchData()
})

async function fetchData() {
  loading.value = true
  try {
    syncOrgId()
    projects.value = await get('/api/v1/projects')
    const allTasks = await get('/api/v1/my-tasks')
    tasks.value = allTasks.map(t => {
      const proj = projects.value.find(p => p.id === t.project_id)
      return { ...t, project_name: proj?.name || '', project_slug: proj?.slug || '' }
    })
  } catch {
    toast.error('Gagal memuat data')
  } finally { loading.value = false }
}

const filteredTasks = computed(() => {
  if (!selectedProjectId.value) return tasks.value
  return tasks.value.filter(t => t.project_id === selectedProjectId.value)
})

const currentMonth = computed(() => currentDate.value.getMonth())
const currentYear = computed(() => currentDate.value.getFullYear())

const headerLabel = computed(() => {
  if (viewMode.value === 'week') {
    const { start, end } = weekRange.value
    const opts = { day: 'numeric', month: 'long', year: 'numeric' }
    return start.toLocaleDateString('id-ID', opts) + ' — ' + end.toLocaleDateString('id-ID', opts)
  }
  return new Date(currentYear.value, currentMonth.value).toLocaleDateString('id-ID', { month: 'long', year: 'numeric' })
})

const weekRange = computed(() => {
  const d = new Date(currentDate.value)
  const day = d.getDay()
  const start = new Date(d)
  start.setDate(d.getDate() - day)
  const end = new Date(start)
  end.setDate(start.getDate() + 6)
  return { start, end }
})

const calendarDays = computed(() => {
  if (viewMode.value === 'week') {
    const days = []
    const { start } = weekRange.value
    for (let i = 0; i < 7; i++) {
      const d = new Date(start)
      d.setDate(start.getDate() + i)
      days.push({ day: d.getDate(), month: d.getMonth(), year: d.getFullYear(), outside: false, date: d })
    }
    return days
  }
  const firstDay = new Date(currentYear.value, currentMonth.value, 1)
  const lastDay = new Date(currentYear.value, currentMonth.value + 1, 0)
  const startDay = firstDay.getDay()
  const totalDays = lastDay.getDate()

  const days = []
  const prevMonth = new Date(currentYear.value, currentMonth.value, 0)
  for (let i = startDay - 1; i >= 0; i--) {
    const d = prevMonth.getDate() - i
    days.push({ day: d, month: currentMonth.value - 1, year: currentYear.value, outside: true, date: new Date(currentYear.value, currentMonth.value - 1, d) })
  }
  for (let d = 1; d <= totalDays; d++) {
    days.push({ day: d, month: currentMonth.value, year: currentYear.value, outside: false, date: new Date(currentYear.value, currentMonth.value, d) })
  }
  const remaining = 42 - days.length
  for (let d = 1; d <= remaining; d++) {
    days.push({ day: d, month: currentMonth.value + 1, year: currentYear.value, outside: true, date: new Date(currentYear.value, currentMonth.value + 1, d) })
  }
  return days
})

function tasksForDate(date) {
  const dateStr = formatDate(date)
  return filteredTasks.value.filter(t => {
    if (!t.due_date) return false
    return t.due_date.split('T')[0] === dateStr
  })
}

function formatDate(d) {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${dd}`
}

function isToday(date) {
  const today = new Date()
  return date.getDate() === today.getDate() && date.getMonth() === today.getMonth() && date.getFullYear() === today.getFullYear()
}

function prevPeriod() {
  const d = new Date(currentDate.value)
  if (viewMode.value === 'week') {
    d.setDate(d.getDate() - 7)
  } else {
    d.setMonth(d.getMonth() - 1)
  }
  currentDate.value = d
}

function nextPeriod() {
  const d = new Date(currentDate.value)
  if (viewMode.value === 'week') {
    d.setDate(d.getDate() + 7)
  } else {
    d.setMonth(d.getMonth() + 1)
  }
  currentDate.value = d
}

function goToToday() {
  currentDate.value = new Date()
}

const priorityConfig = {
  urgent: { dot: 'bg-red-500', text: 'text-red-600 dark:text-red-400', bg: 'bg-red-100 dark:bg-red-900/30' },
  high: { dot: 'bg-orange-400', text: 'text-orange-600 dark:text-orange-400', bg: 'bg-orange-100 dark:bg-orange-900/30' },
  medium: { dot: 'bg-gray-400', text: 'text-gray-600 dark:text-gray-400', bg: 'bg-gray-100 dark:bg-gray-700' },
  low: { dot: 'bg-gray-300', text: 'text-gray-400 dark:text-gray-500', bg: 'bg-gray-50 dark:bg-gray-800' },
}

function openTask(task) {
  selectedTask.value = task
  showTaskDetail.value = true
}

function closeTaskDetail() {
  showTaskDetail.value = false
  selectedTask.value = null
}

// Drag & Drop
const dragTask = ref(null)

function onDragStart(e, task) {
  dragTask.value = task
  e.dataTransfer.effectAllowed = 'move'
  e.dataTransfer.setData('text/plain', task.id)
  e.target.classList.add('opacity-50')
}

function onDragEnd(e) {
  dragTask.value = null
  e.target.classList.remove('opacity-50')
}

function onDragOver(e) {
  e.preventDefault()
  e.dataTransfer.dropEffect = 'move'
  e.currentTarget.classList.add('ring-2', 'ring-indigo-400', 'ring-inset')
}

function onDragLeave(e) {
  e.currentTarget.classList.remove('ring-2', 'ring-indigo-400', 'ring-inset')
}

async function onDrop(e, date) {
  e.currentTarget.classList.remove('ring-2', 'ring-indigo-400', 'ring-inset')
  const taskId = e.dataTransfer.getData('text/plain')
  if (!taskId || !dragTask.value) return

  const newDate = formatDate(date)
  const oldDate = dragTask.value.due_date?.split('T')[0]
  if (newDate === oldDate) return

  try {
    await patch(`/api/v1/tasks/${taskId}`, { due_date: newDate })
    const idx = tasks.value.findIndex(t => t.id === taskId)
    if (idx !== -1) {
      tasks.value[idx] = { ...tasks.value[idx], due_date: newDate + 'T00:00:00Z' }
    }
    toast.success('Deadline diperbarui')
  } catch {
    toast.error('Gagal memperbarui deadline')
  }
}

// Quick Create
function openCreateModal(date) {
  createForm.value = {
    title: '',
    description: '',
    priority: 'medium',
    project_id: selectedProjectId.value || (projects.value[0]?.id || ''),
    due_date: formatDate(date)
  }
  showCreateModal.value = true
}

function closeCreateModal() {
  showCreateModal.value = false
  createForm.value = { title: '', description: '', priority: 'medium', project_id: '', due_date: '' }
}

async function createTask() {
  if (!createForm.value.title.trim()) return toast.error('Judul harus diisi')
  if (!createForm.value.project_id) return toast.error('Pilih proyek')

  creating.value = true
  try {
    const newTask = await post(`/api/v1/projects/${createForm.value.project_id}/tasks`, {
      title: createForm.value.title.trim(),
      description: createForm.value.description.trim(),
      priority: createForm.value.priority,
      due_date: createForm.value.due_date || null
    })
    const proj = projects.value.find(p => p.id === createForm.value.project_id)
    tasks.value.unshift({ ...newTask, project_name: proj?.name || '', project_slug: proj?.slug || '' })
    toast.success('Tugas dibuat')
    closeCreateModal()
  } catch {
    toast.error('Gagal membuat tugas')
  } finally { creating.value = false }
}
</script>

<template>
  <div class="p-4 lg:p-6 max-w-6xl mx-auto w-full">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white">Kalender</h1>
        <p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">Drag tugas untuk ubah deadline, klik tanggal kosong untuk buat tugas baru.</p>
      </div>
    </div>

    <!-- Controls -->
    <div class="flex flex-wrap items-center gap-3 mb-4">
      <!-- View Toggle -->
      <div class="flex items-center bg-gray-100 dark:bg-gray-800 rounded-lg p-0.5">
        <button @click="viewMode = 'month'" class="px-3 py-1.5 text-xs font-medium rounded-md transition-colors" :class="viewMode === 'month' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'">
          Bulanan
        </button>
        <button @click="viewMode = 'week'" class="px-3 py-1.5 text-xs font-medium rounded-md transition-colors" :class="viewMode === 'week' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'">
          Mingguan
        </button>
      </div>

      <!-- Project Filter -->
      <select v-model="selectedProjectId" class="text-xs border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-1.5 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-colors">
        <option value="">Semua Proyek</option>
        <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
      </select>
    </div>

    <!-- Calendar Navigation -->
    <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
      <div class="flex items-center justify-between px-5 py-3 border-b border-gray-100 dark:border-gray-700">
        <div class="flex items-center gap-3">
          <button @click="prevPeriod" class="p-1.5 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
          </button>
          <h2 class="text-sm font-bold text-gray-900 dark:text-white min-w-[200px] text-center">{{ headerLabel }}</h2>
          <button @click="nextPeriod" class="p-1.5 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
          </button>
        </div>
        <button @click="goToToday" class="text-xs text-indigo-600 hover:text-indigo-700 font-medium px-3 py-1.5 rounded-lg hover:bg-indigo-50 dark:hover:bg-indigo-900/20 transition-colors">Hari Ini</button>
      </div>

      <!-- Skeleton / Calendar Grid -->
      <Transition name="fade" mode="out-in">
        <div v-if="loading" :key="'skeleton'" class="p-5 animate-pulse">
          <div class="grid grid-cols-7 gap-1">
            <div v-for="i in 35" :key="i" class="h-20 bg-gray-50 dark:bg-gray-700 rounded-lg"></div>
          </div>
        </div>

        <div v-else :key="viewMode">
          <!-- Day Headers -->
          <div class="grid grid-cols-7 border-b border-gray-100 dark:border-gray-700">
            <div v-for="day in ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']" :key="day" class="px-2 py-2 text-center text-[10px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">
              {{ day }}
            </div>
          </div>

          <!-- Days -->
          <div class="grid grid-cols-7">
            <div
              v-for="(day, i) in calendarDays"
              :key="i"
              class="border-r border-b border-gray-100 dark:border-gray-700 p-1.5 transition-all duration-150"
              :class="[
                viewMode === 'week' ? 'min-h-[200px]' : 'min-h-[80px]',
                day.outside ? 'bg-gray-50/50 dark:bg-gray-700/50' : 'bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer'
              ]"
              @dragover="!day.outside && onDragOver($event)"
              @dragleave="onDragLeave($event)"
              @drop="!day.outside && onDrop($event, day.date)"
              @click="!day.outside && openCreateModal(day.date)"
            >
              <div class="flex items-center justify-between mb-1">
                <span
                  class="text-xs font-medium w-6 h-6 flex items-center justify-center rounded-full"
                  :class="[
                    isToday(day.date) ? 'bg-indigo-600 text-white' : day.outside ? 'text-gray-300 dark:text-gray-600' : 'text-gray-700 dark:text-gray-300'
                  ]"
                >{{ day.day }}</span>
                <span v-if="!day.outside && tasksForDate(day.date).length" class="text-[9px] text-gray-400 dark:text-gray-500">{{ tasksForDate(day.date).length }}</span>
              </div>
              <div class="space-y-0.5">
                <div
                  v-for="task in tasksForDate(day.date).slice(0, viewMode === 'week' ? 8 : 3)"
                  :key="task.id"
                  draggable="true"
                  @dragstart="onDragStart($event, task)"
                  @dragend="onDragEnd($event)"
                  @click.stop="openTask(task)"
                  class="w-full text-left px-1.5 py-0.5 rounded text-[10px] font-medium truncate cursor-grab active:cursor-grabbing transition-all hover:shadow-sm"
                  :class="[priorityConfig[task.priority]?.text || 'text-gray-600 dark:text-gray-400', priorityConfig[task.priority]?.bg || 'bg-gray-100 dark:bg-gray-700']"
                >
                  <span class="inline-block w-1.5 h-1.5 rounded-full mr-1 align-middle" :class="priorityConfig[task.priority]?.dot || 'bg-gray-400 dark:text-gray-500'"></span>
                  {{ task.title }}
                </div>
                <span v-if="tasksForDate(day.date).length > (viewMode === 'week' ? 8 : 3)" class="text-[9px] text-gray-400 dark:text-gray-500 px-1.5">+{{ tasksForDate(day.date).length - (viewMode === 'week' ? 8 : 3) }} lagi</span>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </div>

    <!-- Task Detail Modal -->
    <Teleport to="body">
      <Transition enter-active-class="transition-all duration-200 ease-out" leave-active-class="transition-all duration-150 ease-in" enter-from-class="opacity-0 scale-95" leave-to-class="opacity-0 scale-95">
        <div v-if="showTaskDetail && selectedTask" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="closeTaskDetail"></div>
          <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-md overflow-hidden">
            <div class="px-5 py-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="text-xs font-mono text-gray-400 dark:text-gray-500">{{ selectedTask.key }}</span>
                <span class="text-[11px] px-2 py-0.5 rounded-full font-medium" :class="priorityConfig[selectedTask.priority]?.text + ' ' + priorityConfig[selectedTask.priority]?.bg">
                  {{ selectedTask.priority }}
                </span>
              </div>
              <button @click="closeTaskDetail" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <div class="px-5 py-4 space-y-3">
              <h3 class="text-base font-bold text-gray-900 dark:text-white">{{ selectedTask.title }}</h3>
              <p v-if="selectedTask.description" class="text-sm text-gray-600 dark:text-gray-400">{{ selectedTask.description }}</p>
              <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
                <span v-if="selectedTask.project_name">Proyek: {{ selectedTask.project_name }}</span>
                <span v-if="selectedTask.due_date">Deadline: {{ new Date(selectedTask.due_date).toLocaleDateString('id-ID') }}</span>
              </div>
            </div>
            <div class="px-5 py-3 border-t border-gray-100 dark:border-gray-700 flex justify-end">
              <NuxtLink :to="`/tasks/${selectedTask.id}`" class="text-xs text-indigo-600 hover:text-indigo-700 font-medium" @click="closeTaskDetail">Lihat Detail</NuxtLink>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Quick Create Modal -->
    <Teleport to="body">
      <Transition enter-active-class="transition-all duration-200 ease-out" leave-active-class="transition-all duration-150 ease-in" enter-from-class="opacity-0 scale-95" leave-to-class="opacity-0 scale-95">
        <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="closeCreateModal"></div>
          <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-md overflow-hidden">
            <div class="px-5 py-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
              <h3 class="text-sm font-bold text-gray-900 dark:text-white">Buat Tugas Baru</h3>
              <button @click="closeCreateModal" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <div class="px-5 py-4 space-y-3">
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Judul *</label>
                <input v-model="createForm.title" type="text" placeholder="Judul tugas..." class="w-full text-sm border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 bg-white dark:bg-gray-900 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-colors" @keydown.enter="createTask" />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Deskripsi</label>
                <textarea v-model="createForm.description" rows="2" placeholder="Deskripsi singkat..." class="w-full text-sm border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 bg-white dark:bg-gray-900 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-colors resize-none"></textarea>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Prioritas</label>
                  <select v-model="createForm.priority" class="w-full text-sm border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 bg-white dark:bg-gray-900 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-colors">
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                    <option value="urgent">Urgent</option>
                  </select>
                </div>
                <div>
                  <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Deadline</label>
                  <input v-model="createForm.due_date" type="date" class="w-full text-sm border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 bg-white dark:bg-gray-900 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-colors" />
                </div>
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Proyek *</label>
                <select v-model="createForm.project_id" class="w-full text-sm border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 bg-white dark:bg-gray-900 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-colors">
                  <option value="" disabled>Pilih proyek</option>
                  <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
                </select>
              </div>
            </div>
            <div class="px-5 py-3 border-t border-gray-100 dark:border-gray-700 flex justify-end gap-2">
              <button @click="closeCreateModal" class="px-4 py-2 text-xs font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors">Batal</button>
              <button @click="createTask" :disabled="creating" class="px-4 py-2 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-colors disabled:opacity-50">
                {{ creating ? 'Membuat...' : 'Buat Tugas' }}
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
