<script setup>
definePageMeta({ layout: 'dashboard' })

const toast = useToast()
const route = useRoute()
const { get, post, patch, del, setOrgId } = useApi()
const { user, orgId, orgRole, token } = useAuth()
const slug = route.params.slug

const loading = ref(true)
const project = ref(null)
const tasks = ref([])
const projectId = ref(null)

function syncOrgId() {
  const oid = orgId.value || (import.meta.client ? localStorage.getItem('auth:org_id') : null)
  if (oid) setOrgId(oid)
  return oid
}

onMounted(async () => {
  syncOrgId()
  watch(orgId, syncOrgId)
  const hasToken = token.value || (import.meta.client ? localStorage.getItem('auth:token') : null)
  if (!hasToken) {
    loading.value = false
    return
  }
  await fetchProject()
})

async function fetchProject() {
  loading.value = true
  try {
    syncOrgId()
    const projects = await get('/api/v1/projects')
    const found = projects.find(p => p.slug === slug)
    if (!found) {
      toast.error('Proyek tidak ditemukan')
      return
    }
    project.value = found
    projectId.value = found.id
    await fetchTasks()
  } catch (e) {
    toast.error(e?.message || 'Gagal memuat proyek')
  } finally {
    loading.value = false
  }
}

async function fetchTasks() {
  if (!projectId.value) return
  try {
    syncOrgId()
    const data = await get(`/api/v1/projects/${projectId.value}/tasks`)
    tasks.value = data
  } catch (e) {
    toast.error(e?.message || 'Gagal memuat tugas')
  }
}

const columns = [
  { key: 'backlog', label: 'Backlog', color: 'bg-gray-400' },
  { key: 'todo', label: 'Todo', color: 'bg-blue-500' },
  { key: 'in_progress', label: 'In Progress', color: 'bg-amber-500' },
  { key: 'in_review', label: 'In Review', color: 'bg-purple-500' },
  { key: 'done', label: 'Done', color: 'bg-emerald-500' },
]

const priorityConfig = {
  urgent: { label: 'Urgent', dot: 'bg-red-500', text: 'text-red-600' },
  high: { label: 'High', dot: 'bg-orange-400', text: 'text-orange-600' },
  medium: { label: 'Medium', dot: 'bg-gray-400', text: 'text-gray-600' },
  low: { label: 'Low', dot: 'bg-gray-300', text: 'text-gray-400' },
}

const labelColors = {
  Frontend: 'bg-indigo-50 text-indigo-600',
  Backend: 'bg-purple-50 text-purple-600',
  Design: 'bg-pink-50 text-pink-600',
  DevOps: 'bg-emerald-50 text-emerald-600',
  Testing: 'bg-amber-50 text-amber-700',
  Planning: 'bg-gray-100 text-gray-600',
}

const getTasksByStatus = (status) => tasks.value.filter(t => t.status === status).sort((a, b) => (a.position || 0) - (b.position || 0))

const selectedTask = ref(null)
const showTaskDetail = ref(false)

function openTask(task) {
  selectedTask.value = JSON.parse(JSON.stringify(task))
  showTaskDetail.value = true
}

async function closeTaskDetail() {
  if (selectedTask.value) {
    try {
      await patch(`/api/v1/tasks/${selectedTask.value.id}`, {
        status: selectedTask.value.status,
        priority: selectedTask.value.priority,
        description: selectedTask.value.description,
      })
      const idx = tasks.value.findIndex(t => t.id === selectedTask.value.id)
      if (idx !== -1) {
        tasks.value[idx].status = selectedTask.value.status
        tasks.value[idx].priority = selectedTask.value.priority
        tasks.value[idx].description = selectedTask.value.description
      }
      toast.success('Tugas diperbarui')
    } catch (e) {
      toast.error('Gagal memperbarui tugas')
    }
  }
  showTaskDetail.value = false
  selectedTask.value = null
}

const showAddTask = ref(false)
const newTask = ref({ title: '', description: '', priority: 'medium' })
const creating = ref(false)

function openAddTask(columnStatus) {
  newTask.value = { title: '', description: '', priority: 'medium', status: columnStatus || 'todo' }
  showAddTask.value = true
}

async function addTask() {
  if (!newTask.value.title.trim() || !projectId.value) return
  creating.value = true
  try {
    const t = await post(`/api/v1/projects/${projectId.value}/tasks`, {
      title: newTask.value.title,
      description: newTask.value.description,
      priority: newTask.value.priority,
      status: newTask.value.status || 'todo',
    })
    tasks.value.push(t)
    showAddTask.value = false
    toast.success('Tugas berhasil dibuat!')
  } catch (e) {
    toast.error(e.message || 'Gagal membuat tugas')
  } finally {
    creating.value = false
  }
}

async function deleteTask() {
  if (!selectedTask.value) return
  try {
    await del(`/api/v1/tasks/${selectedTask.value.id}`)
    tasks.value = tasks.value.filter(t => t.id !== selectedTask.value.id)
    showTaskDetail.value = false
    selectedTask.value = null
    toast.success('Tugas dihapus')
  } catch (e) {
    toast.error('Gagal menghapus tugas')
  }
}

const draggedItem = ref(null)
const dragSourceColumn = ref(null)
const dragOverColumn = ref(null)
const dragOverIndex = ref(-1)

function onDragStart(task, column) {
  draggedItem.value = task
  dragSourceColumn.value = column
}

function onDragEnd() {
  draggedItem.value = null
  dragSourceColumn.value = null
  dragOverColumn.value = null
  dragOverIndex.value = -1
}

function onDragOver(e, column, index) {
  e.preventDefault()
  if (dragSourceColumn.value !== column) {
    dragOverColumn.value = column
  }
  dragOverIndex.value = index
}

function onDragLeave() {
  dragOverColumn.value = null
  dragOverIndex.value = -1
}

function calculatePosition(tasksInColumn, targetIndex, sourceId) {
  const filtered = tasksInColumn.filter(t => t.id !== sourceId)
  if (filtered.length === 0) return 1.0
  if (targetIndex <= 0) return (filtered[0].position || 1) - 1
  if (targetIndex >= filtered.length) return (filtered[filtered.length - 1].position || filtered.length) + 1
  const before = filtered[targetIndex - 1].position || targetIndex
  const after = filtered[targetIndex].position || (targetIndex + 1)
  return (before + after) / 2
}

async function onDrop(e, targetStatus) {
  e.preventDefault()
  dragOverColumn.value = null
  dragOverIndex.value = -1
  if (!draggedItem.value) return

  const task = tasks.value.find(t => t.id === draggedItem.value.id)
  if (!task) { draggedItem.value = null; dragSourceColumn.value = null; return }

  const sameColumn = dragSourceColumn.value === targetStatus
  const targetTasks = getTasksByStatus(targetStatus)
  const targetIndex = dragOverIndex.value >= 0 ? dragOverIndex.value : targetTasks.length

  const newPosition = calculatePosition(targetTasks, targetIndex, task.id)

  try {
    const payload = { position: newPosition }
    if (!sameColumn) payload.status = targetStatus
    await patch(`/api/v1/tasks/${task.id}`, payload)
    task.position = newPosition
    if (!sameColumn) task.status = targetStatus
    if (!sameColumn) toast.info(`Tugas dipindah ke ${columns.find(c => c.key === targetStatus)?.label}`)
  } catch {
    toast.error('Gagal memindahkan tugas')
  }

  draggedItem.value = null
  dragSourceColumn.value = null
}

function moveTask(direction) {
  if (!selectedTask.value) return
  const statusOrder = ['backlog', 'todo', 'in_progress', 'in_review', 'done']
  const currentIndex = statusOrder.indexOf(selectedTask.value.status)
  const newIndex = currentIndex + direction
  if (newIndex < 0 || newIndex >= statusOrder.length) return
  selectedTask.value.status = statusOrder[newIndex]
}

function isOverdue(dateStr) {
  if (!dateStr) return false
  return new Date(dateStr) < new Date()
}

const taskCount = computed(() => {
  const counts = {}
  for (const col of columns) {
    counts[col.key] = getTasksByStatus(col.key).length
  }
  return counts
})

const completionPercent = computed(() => {
  const done = getTasksByStatus('done').length
  return tasks.value.length > 0 ? Math.round((done / tasks.value.length) * 100) : 0
})

const canEditProject = computed(() => ['owner', 'admin'].includes(orgRole.value))
const showEditProject = ref(false)
const editProjectForm = reactive({ name: '', description: '' })
const savingProject = ref(false)

function openEditProject() {
  editProjectForm.name = project.value?.name || ''
  editProjectForm.description = project.value?.description || ''
  showEditProject.value = true
}

async function saveProject() {
  if (!editProjectForm.name.trim() || !projectId.value) return
  savingProject.value = true
  try {
    await patch(`/api/v1/projects/${projectId.value}`, {
      name: editProjectForm.name,
      description: editProjectForm.description,
    })
    project.value.name = editProjectForm.name
    project.value.description = editProjectForm.description
    showEditProject.value = false
    toast.success('Proyek berhasil diperbarui')
  } catch (e) {
    toast.error(e?.message || 'Gagal memperbarui proyek')
  } finally {
    savingProject.value = false
  }
}
</script>

<template>
  <div class="h-full flex flex-col dark:bg-gray-900">
    <Transition name="fade" mode="out-in">
      <!-- Skeleton -->
      <div v-if="loading" class="h-full flex flex-col" :key="'skeleton'">
        <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-3 py-2 animate-pulse">
          <div class="flex items-center gap-2"><div class="h-3 bg-gray-200 dark:bg-gray-600 rounded w-24"></div><div class="h-3 bg-gray-200 dark:bg-gray-600 rounded w-32"></div></div>
        </div>
        <div class="flex-1 p-1.5 flex gap-1.5">
          <div v-for="i in 5" :key="i" class="flex-1 bg-gray-50/80 dark:bg-gray-700/50 rounded-xl p-2 space-y-2">
            <div class="h-3 bg-gray-200 dark:bg-gray-600 rounded w-16 mb-2"></div>
            <div v-for="j in 2" :key="j" class="h-16 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 animate-pulse"></div>
          </div>
        </div>
      </div>

      <div v-else-if="project" :key="'data'" class="h-full flex flex-col">
        <!-- Project Header -->
        <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-3 py-2 shrink-0">
          <div class="flex items-center justify-between gap-3">
            <div class="flex items-center gap-2 min-w-0">
              <NuxtLink to="/projects" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors shrink-0">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
              </NuxtLink>
              <div class="w-2 h-2 rounded-sm bg-indigo-500 shrink-0"></div>
              <div class="relative group min-w-0">
                <h1 class="text-[13px] font-bold text-gray-900 dark:text-white truncate cursor-default">{{ project.name }}</h1>
                <div v-if="project.description" class="absolute left-0 top-full mt-1 w-64 px-2.5 py-1.5 bg-gray-900 dark:bg-gray-700 text-white text-[11px] rounded-lg shadow-lg opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity duration-150 z-50">
                  {{ project.description }}
                </div>
              </div>
              <span class="text-[9px] px-1 py-px rounded bg-emerald-50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 font-medium shrink-0">Active</span>
            </div>

            <div class="flex items-center gap-2 shrink-0">
              <div class="hidden sm:flex items-center gap-1">
                <div class="w-14 h-1 rounded-full bg-gray-100 dark:bg-gray-700 overflow-hidden">
                  <div class="h-full bg-indigo-500 rounded-full transition-all duration-500" :style="{ width: completionPercent + '%' }"></div>
                </div>
                <span class="text-[10px] font-medium text-gray-500 dark:text-gray-400">{{ completionPercent }}%</span>
              </div>
              <button v-if="canEditProject" @click="openEditProject" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 text-[10px] font-medium px-2 py-1 rounded-md transition-all flex items-center gap-0.5" title="Edit Proyek">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                <span class="hidden sm:inline">Edit</span>
              </button>
              <button @click="openAddTask()" class="bg-indigo-600 hover:bg-indigo-700 text-white text-[10px] font-medium px-2 py-1 rounded-md transition-all flex items-center gap-0.5 hover:shadow-sm active:scale-[0.97]">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
                <span class="hidden sm:inline">Task</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Kanban Board -->
        <div class="flex-1 overflow-hidden p-1.5">
          <div class="flex gap-1.5 h-full">
            <div
              v-for="col in columns"
              :key="col.key"
              class="flex-1 min-w-0 flex flex-col rounded-xl bg-gray-50/80 dark:bg-gray-700/50 transition-all duration-200"
              :class="{ 'bg-indigo-50 dark:bg-indigo-900/30 ring-2 ring-indigo-200 dark:ring-indigo-800 shadow-lg': dragOverColumn === col.key }"
              @dragover="onDragOver($event, col.key)"
              @dragleave="onDragLeave"
              @drop="onDrop($event, col.key)"
            >
              <div class="flex items-center justify-between px-2 py-1.5 shrink-0">
                <div class="flex items-center gap-1 min-w-0">
                  <div class="w-1.5 h-1.5 rounded-full shrink-0" :class="col.color"></div>
                  <span class="text-[10px] font-semibold text-gray-600 dark:text-gray-400 uppercase tracking-wide truncate">{{ col.label }}</span>
                  <span class="text-[9px] text-gray-400 dark:text-gray-500 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 px-1 rounded-full font-medium shrink-0">{{ taskCount[col.key] }}</span>
                </div>
                <button class="text-gray-300 dark:text-gray-600 hover:text-gray-500 dark:hover:text-gray-400 p-0.5 rounded transition-colors shrink-0" @click="openAddTask(col.key)">
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
                </button>
              </div>

              <div class="flex-1 overflow-y-auto px-1.5 pb-1.5 space-y-1 min-h-0">
                <div
                  v-for="(task, taskIndex) in getTasksByStatus(col.key)"
                  :key="task.id"
                  draggable="true"
                  class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200/80 dark:border-gray-700/80 p-1.5 cursor-grab active:cursor-grabbing hover:shadow-md hover:border-gray-300 dark:hover:border-gray-600 transition-all select-none group"
                  :class="{ 'opacity-40 ring-2 ring-indigo-300 scale-[0.98]': draggedItem?.id === task.id }"
                  @dragstart="onDragStart(task, col.key)"
                  @dragend="onDragEnd"
                  @dragover.stop="onDragOver($event, col.key, taskIndex)"
                  @click="openTask(task)"
                >
                  <div class="flex items-center justify-between">
                    <span class="text-[10px] font-mono text-gray-400 dark:text-gray-500">{{ task.key || 'TG-' + String(task.id).slice(-2) }}</span>
                    <div class="w-1.5 h-1.5 rounded-full" :class="priorityConfig[task.priority]?.dot || 'bg-gray-400'"></div>
                  </div>
                  <p class="text-[12px] font-medium text-gray-800 dark:text-gray-200 leading-tight mt-0.5 truncate">{{ task.title }}</p>
                  <div v-if="task.labels && task.labels.length" class="flex items-center gap-0.5 mt-1">
                    <span class="text-[9px] px-1 py-px rounded font-medium truncate max-w-full" :class="labelColors[task.labels[0]] || 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'">{{ task.labels[0] }}</span>
                    <span v-if="task.labels.length > 1" class="text-[9px] text-gray-400 dark:text-gray-500 shrink-0">+{{ task.labels.length - 1 }}</span>
                  </div>
                  <div class="flex items-center justify-between mt-1 pt-1 border-t border-gray-100 dark:border-gray-700">
                    <span v-if="task.assignee_name" class="text-[9px] text-gray-500 dark:text-gray-400 truncate">{{ task.assignee_name }}</span>
                    <span v-else class="text-[9px] text-gray-300 dark:text-gray-600">—</span>
                    <span v-if="task.due_date" class="text-[9px] shrink-0 ml-1" :class="isOverdue(task.due_date) && task.status !== 'done' ? 'text-red-500 dark:text-red-400 font-medium' : 'text-gray-400 dark:text-gray-500'">
                      {{ new Date(task.due_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' }) }}
                    </span>
                  </div>
                </div>

                <button class="w-full py-1 rounded-lg border border-dashed border-gray-200 dark:border-gray-700 text-[10px] text-gray-400 dark:text-gray-500 hover:border-indigo-300 dark:hover:border-indigo-700 hover:text-indigo-500 dark:hover:text-indigo-400 hover:bg-white/80 dark:hover:bg-gray-800/80 transition-all text-center" @click="openAddTask(col.key)">
                  + Tambah
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Task Detail Slide Panel -->
    <Teleport to="body">
      <Transition name="slide-panel">
        <div v-if="showTaskDetail && selectedTask" class="fixed inset-0 z-50 flex justify-end">
          <div class="absolute inset-0 bg-black/20 backdrop-blur-sm" @click="closeTaskDetail"></div>
          <div class="relative w-full max-w-md bg-white dark:bg-gray-800 shadow-2xl overflow-y-auto" style="will-change: transform">
            <div class="sticky top-0 bg-white/95 dark:bg-gray-800/95 backdrop-blur-sm border-b border-gray-200 dark:border-gray-700 px-5 py-3 flex items-center justify-between z-10">
              <div class="flex items-center gap-2">
                <span class="text-xs font-mono text-gray-400 dark:text-gray-500">{{ selectedTask.key || 'TG-' + String(selectedTask.id).slice(-2) }}</span>
                <span class="text-[10px] font-medium px-1.5 py-0.5 rounded-full" :class="priorityConfig[selectedTask.priority]?.text?.replace('text-', 'bg-') + '/10 ' + priorityConfig[selectedTask.priority]?.text">
                  {{ priorityConfig[selectedTask.priority]?.label }}
                </span>
              </div>
              <button @click="closeTaskDetail" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>

            <div class="px-5 py-4 space-y-4">
              <h2 class="text-base font-bold text-gray-900 dark:text-white leading-snug">{{ selectedTask.title }}</h2>

              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="text-[10px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Status</label>
                  <select v-model="selectedTask.status" class="mt-1 text-xs border border-gray-200 dark:border-gray-700 rounded-lg px-2.5 py-1.5 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 w-full transition-all">
                    <option v-for="col in columns" :key="col.key" :value="col.key">{{ col.label }}</option>
                  </select>
                </div>
                <div>
                  <label class="text-[10px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Prioritas</label>
                  <select v-model="selectedTask.priority" class="mt-1 text-xs border border-gray-200 dark:border-gray-700 rounded-lg px-2.5 py-1.5 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 w-full transition-all">
                    <option value="urgent">Urgent</option>
                    <option value="high">High</option>
                    <option value="medium">Medium</option>
                    <option value="low">Low</option>
                  </select>
                </div>
              </div>

              <div>
                <label class="text-[10px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Deskripsi</label>
                <textarea v-model="selectedTask.description" rows="3" class="mt-1 text-xs border border-gray-200 dark:border-gray-700 rounded-lg px-2.5 py-2 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 w-full resize-none transition-all" placeholder="Tambahkan deskripsi..."></textarea>
              </div>
            </div>

            <div class="sticky bottom-0 bg-white/95 dark:bg-gray-800/95 backdrop-blur-sm border-t border-gray-200 dark:border-gray-700 px-5 py-2.5 flex items-center justify-between">
              <div class="flex items-center gap-1.5">
                <button @click="moveTask(-1)" :disabled="selectedTask.status === 'backlog'" class="text-[11px] text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white px-2 py-1 rounded-md border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-all disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-1">
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
                  Prev
                </button>
                <button @click="moveTask(1)" :disabled="selectedTask.status === 'done'" class="text-[11px] text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white px-2 py-1 rounded-md border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-all disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-1">
                  Next
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
                </button>
              </div>
              <button class="text-[11px] text-red-500 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 px-2 py-1 rounded-md hover:bg-red-50 dark:hover:bg-red-900/20 transition-all flex items-center gap-1" @click="deleteTask">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                Hapus
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Add Task Modal -->
    <Teleport to="body">
      <Transition enter-active-class="transition-all duration-200 ease-out" leave-active-class="transition-all duration-150 ease-in" enter-from-class="opacity-0 scale-95" leave-to-class="opacity-0 scale-95">
        <div v-if="showAddTask" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showAddTask = false"></div>
          <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-sm overflow-hidden">
            <div class="px-5 py-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
              <h3 class="text-sm font-bold text-gray-900 dark:text-white">Tugas Baru</h3>
              <button @click="showAddTask = false" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <div class="px-5 py-4 space-y-3">
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Judul <span class="text-red-500 dark:text-red-400">*</span></label>
                <input v-model="newTask.title" type="text" placeholder="Judul tugas..." class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all" @keydown.enter="addTask" />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Deskripsi</label>
                <textarea v-model="newTask.description" rows="2" placeholder="Deskripsi singkat..." class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none transition-all"></textarea>
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Prioritas</label>
                <select v-model="newTask.priority" class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all">
                  <option value="urgent">Urgent</option>
                  <option value="high">High</option>
                  <option value="medium">Medium</option>
                  <option value="low">Low</option>
                </select>
              </div>
            </div>
            <div class="px-5 py-3 border-t border-gray-100 dark:border-gray-700 flex justify-end gap-2">
              <button @click="showAddTask = false" class="px-3 py-1.5 text-xs font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg transition-colors">Batal</button>
              <button @click="addTask" :disabled="!newTask.title.trim() || creating" class="px-3 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-all disabled:opacity-50 flex items-center gap-1.5">
                <svg v-if="creating" class="animate-spin w-3 h-3" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
                Buat
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Edit Project Modal -->
    <Teleport to="body">
      <Transition enter-active-class="transition-all duration-200 ease-out" leave-active-class="transition-all duration-150 ease-in" enter-from-class="opacity-0 scale-95" leave-to-class="opacity-0 scale-95">
        <div v-if="showEditProject" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showEditProject = false"></div>
          <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-sm overflow-hidden">
            <div class="px-5 py-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
              <h3 class="text-sm font-bold text-gray-900 dark:text-white">Edit Proyek</h3>
              <button @click="showEditProject = false" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <div class="px-5 py-4 space-y-3">
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Nama Proyek <span class="text-red-500 dark:text-red-400">*</span></label>
                <input v-model="editProjectForm.name" type="text" placeholder="Nama proyek..." class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all" />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Deskripsi</label>
                <textarea v-model="editProjectForm.description" rows="3" placeholder="Deskripsi proyek..." class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none transition-all"></textarea>
              </div>
            </div>
            <div class="px-5 py-3 border-t border-gray-100 dark:border-gray-700 flex justify-end gap-2">
              <button @click="showEditProject = false" class="px-3 py-1.5 text-xs font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg transition-colors">Batal</button>
              <button @click="saveProject" :disabled="!editProjectForm.name.trim() || savingProject" class="px-3 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-all disabled:opacity-50 flex items-center gap-1.5">
                <svg v-if="savingProject" class="animate-spin w-3 h-3" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
                Simpan
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
.slide-panel-enter-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}
.slide-panel-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.slide-panel-enter-from {
  opacity: 0;
  transform: translateX(100%);
}
.slide-panel-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>
