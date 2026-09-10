<script setup>
definePageMeta({ layout: 'dashboard' })

const route = useRoute()
const { get, post, patch, del, setOrgId } = useApi()
const { user, orgId, token } = useAuth()
const toast = useToast()

const taskId = route.params.id
const loading = ref(true)
const task = ref(null)

const comments = ref([])
const loadingComments = ref(false)
const newComment = ref('')
const submittingComment = ref(false)

const activityLogs = ref([])
const loadingActivity = ref(false)

const dependencies = ref([])
const loadingDeps = ref(false)
const showAddDep = ref(false)
const depSearchQuery = ref('')
const depSearchResults = ref([])

const attachments = ref([])
const loadingAttachments = ref(false)
const uploading = ref(false)
const fileInput = ref(null)

const columns = [
  { key: 'backlog', label: 'Backlog', color: 'bg-gray-400' },
  { key: 'todo', label: 'Todo', color: 'bg-blue-500' },
  { key: 'in_progress', label: 'In Progress', color: 'bg-amber-500' },
  { key: 'in_review', label: 'In Review', color: 'bg-purple-500' },
  { key: 'done', label: 'Done', color: 'bg-emerald-500' },
]

const priorityConfig = {
  urgent: { label: 'Urgent', dot: 'bg-red-500', bg: 'bg-red-50', text: 'text-red-600' },
  high: { label: 'High', dot: 'bg-orange-400', bg: 'bg-orange-50', text: 'text-orange-600' },
  medium: { label: 'Medium', dot: 'bg-gray-400', bg: 'bg-gray-100', text: 'text-gray-600' },
  low: { label: 'Low', dot: 'bg-gray-300', bg: 'bg-gray-50', text: 'text-gray-400' },
}

const statusConfig = {
  backlog: { label: 'Backlog', dot: 'bg-gray-400', bg: 'bg-gray-100', text: 'text-gray-600' },
  todo: { label: 'Todo', dot: 'bg-blue-500', bg: 'bg-blue-50', text: 'text-blue-600' },
  in_progress: { label: 'In Progress', dot: 'bg-amber-500', bg: 'bg-amber-50', text: 'text-amber-700' },
  in_review: { label: 'In Review', dot: 'bg-purple-500', bg: 'bg-purple-50', text: 'text-purple-600' },
  done: { label: 'Done', dot: 'bg-emerald-500', bg: 'bg-emerald-50', text: 'text-emerald-600' },
}

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
  await fetchTask()
  await fetchComments()
  await fetchActivityLogs()
  await fetchDependencies()
  await fetchAttachments()
})

async function fetchTask() {
  loading.value = true
  try {
    syncOrgId()
    task.value = await get(`/api/v1/tasks/${taskId}`)
  } catch (e) {
    toast.error(e?.message || 'Gagal memuat tugas')
  } finally {
    loading.value = false
  }
}

async function updateTask() {
  if (!task.value) return
  try {
    await patch(`/api/v1/tasks/${task.value.id}`, {
      status: task.value.status,
      priority: task.value.priority,
      description: task.value.description,
    })
    toast.success('Tugas diperbarui')
  } catch (e) {
    toast.error('Gagal memperbarui tugas')
  }
}

async function deleteTask() {
  if (!task.value) return
  try {
    await del(`/api/v1/tasks/${task.value.id}`)
    toast.success('Tugas dihapus')
    navigateTo('/dashboard')
  } catch (e) {
    toast.error('Gagal menghapus tugas')
  }
}

async function fetchComments() {
  loadingComments.value = true
  try {
    syncOrgId()
    comments.value = await get(`/api/v1/tasks/${taskId}/comments`)
  } catch (e) {
    console.error('Failed to load comments', e)
  } finally {
    loadingComments.value = false
  }
}

async function fetchActivityLogs() {
  loadingActivity.value = true
  try {
    syncOrgId()
    activityLogs.value = await get(`/api/v1/activity-logs?entity_type=task&entity_id=${taskId}`)
  } catch (e) {
    console.error('Failed to load activity logs', e)
  } finally {
    loadingActivity.value = false
  }
}

async function fetchDependencies() {
  loadingDeps.value = true
  try {
    syncOrgId()
    dependencies.value = await get(`/api/v1/tasks/${taskId}/dependencies`)
  } catch { dependencies.value = [] } finally { loadingDeps.value = false }
}

async function searchDependencies() {
  const q = depSearchQuery.value.trim()
  if (!q || q.length < 2) { depSearchResults.value = []; return }
  try {
    syncOrgId()
    const all = await get(`/api/v1/tasks/${task.value.project_id}/tasks`)
    depSearchResults.value = all.filter(t => t.id !== taskId && (t.key?.toLowerCase().includes(q.toLowerCase()) || t.title?.toLowerCase().includes(q.toLowerCase()))).slice(0, 5)
  } catch { depSearchResults.value = [] }
}

async function addDependency(depTask) {
  try {
    syncOrgId()
    await post(`/api/v1/tasks/${taskId}/dependencies`, { depends_on_task_id: depTask.id })
    dependencies.value.push({ id: Date.now().toString(), depends_on_task: depTask })
    showAddDep.value = false
    depSearchQuery.value = ''
    depSearchResults.value = []
    toast.success(`Tugas ini bergantung pada ${depTask.key || depTask.title}`)
  } catch (e) { toast.error(e?.message || 'Gagal menambah dependency') }
}

async function removeDependency(depId) {
  try {
    syncOrgId()
    await del(`/api/v1/dependencies/${depId}`)
    dependencies.value = dependencies.value.filter(d => d.id !== depId)
    toast.success('Dependency dihapus')
  } catch (e) { toast.error(e?.message || 'Gagal menghapus dependency') }
}

async function fetchAttachments() {
  loadingAttachments.value = true
  try {
    syncOrgId()
    attachments.value = await get(`/api/v1/tasks/${taskId}/attachments`)
  } catch { attachments.value = [] } finally { loadingAttachments.value = false }
}

async function uploadFile(e) {
  const file = e.target.files?.[0]
  if (!file) return
  uploading.value = true
  try {
    syncOrgId()
    const form = new FormData()
    form.append('file', file)
    const res = await fetch(`${useRuntimeConfig().public.apiBaseUrl}/api/v1/tasks/${taskId}/attachments`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token.value}`, 'X-Organization-Id': orgId.value },
      body: form,
    })
    if (!res.ok) throw new Error('Upload gagal')
    const att = await res.json()
    attachments.value.push(att)
    toast.success('File berhasil diupload')
  } catch (e) { toast.error(e?.message || 'Gagal upload file') } finally { uploading.value = false; if (fileInput.value) fileInput.value.value = '' }
}

async function deleteAttachment(attId) {
  try {
    syncOrgId()
    await del(`/api/v1/attachments/${attId}`)
    attachments.value = attachments.value.filter(a => a.id !== attId)
    toast.success('File dihapus')
  } catch (e) { toast.error(e?.message || 'Gagal menghapus file') }
}

function formatFileSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let size = bytes
  while (size >= 1024 && i < units.length - 1) { size /= 1024; i++ }
  return `${size.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

async function submitComment() {
  const body = newComment.value.trim()
  if (!body) return
  submittingComment.value = true
  try {
    syncOrgId()
    const c = await post(`/api/v1/tasks/${taskId}/comments`, { body })
    comments.value.push(c)
    newComment.value = ''
  } catch (e) {
    toast.error(e?.message || 'Gagal mengirim komentar')
  } finally {
    submittingComment.value = false
  }
}

async function deleteComment(commentId) {
  try {
    syncOrgId()
    await del(`/api/v1/comments/${commentId}`)
    comments.value = comments.value.filter(c => c.id !== commentId)
    toast.success('Komentar dihapus')
  } catch (e) {
    toast.error(e?.message || 'Gagal menghapus komentar')
  }
}

function canDeleteComment(comment) {
  if (!user.value) return false
  return comment.user_id === user.value.id || ['owner', 'admin'].includes(user.value.org_role)
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', { month: 'short', day: 'numeric', year: 'numeric' })
}

function formatDateTime(dateStr) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('id-ID', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function isOverdue(dateStr) {
  if (!dateStr) return false
  return new Date(dateStr) < new Date()
}

function commentInitial(name, email) {
  return (name || email || '?').charAt(0).toUpperCase()
}
</script>

<template>
  <div class="p-4 lg:p-6 max-w-5xl mx-auto w-full">
    <Transition name="fade" mode="out-in">
      <!-- Skeleton -->
      <div v-if="loading" key="skeleton" class="space-y-4 animate-pulse">
        <div class="h-3 bg-gray-200 rounded w-48"></div>
        <div class="h-5 bg-gray-200 rounded w-64"></div>
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div class="lg:col-span-2 space-y-4">
            <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 h-32"></div>
          </div>
          <div class="space-y-4">
            <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 h-20"></div>
            <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 h-20"></div>
          </div>
        </div>
      </div>

      <div v-else-if="task" :key="'data'">
      <!-- Breadcrumb -->
      <div class="flex items-center gap-2 text-xs text-gray-400 dark:text-gray-500 mb-4">
        <NuxtLink to="/dashboard" class="hover:text-gray-600 transition-colors">Dashboard</NuxtLink>
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
        <span class="text-gray-600">{{ task.key || 'TG-' + String(task.id).slice(-2) }}</span>
      </div>

      <!-- Header -->
      <div class="flex items-start justify-between gap-4 mb-6">
        <div class="flex-1">
          <div class="flex items-center gap-2 mb-1">
            <span class="text-sm font-mono text-gray-400 dark:text-gray-500">{{ task.key || 'TG-' + String(task.id).slice(-2) }}</span>
            <span class="text-[11px] px-2 py-0.5 rounded-full font-medium" :class="statusConfig[task.status]?.bg + ' ' + statusConfig[task.status]?.text">
              {{ statusConfig[task.status]?.label }}
            </span>
            <span class="text-[11px] px-2 py-0.5 rounded-full font-medium" :class="priorityConfig[task.priority]?.bg + ' ' + priorityConfig[task.priority]?.text">
              {{ priorityConfig[task.priority]?.label }}
            </span>
          </div>
          <h1 class="text-xl font-bold text-gray-900 dark:text-white">{{ task.title }}</h1>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Main Content -->
        <div class="lg:col-span-2 space-y-6">
          <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
            <h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Deskripsi</h3>
            <p v-if="task.description" class="text-sm text-gray-700 dark:text-gray-300 leading-relaxed whitespace-pre-wrap">{{ task.description }}</p>
            <p v-else class="text-sm text-gray-400 dark:text-gray-500 italic">Belum ada deskripsi</p>
          </div>

          <!-- Comments Section -->
          <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
            <h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-4">Komentar</h3>

            <!-- Comment Input -->
            <div class="flex gap-3 mb-5">
              <div class="w-8 h-8 rounded-full bg-indigo-500 flex items-center justify-center shrink-0">
                <span class="text-xs text-white font-bold">{{ commentInitial(user?.full_name, user?.email) }}</span>
              </div>
              <div class="flex-1">
                <textarea
                  v-model="newComment"
                  placeholder="Tulis komentar..."
                  rows="2"
                  class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all resize-none"
                  @keydown.enter.meta="submitComment"
                  @keydown.enter.ctrl="submitComment"
                ></textarea>
                <div class="flex justify-end mt-2">
                  <button @click="submitComment" :disabled="!newComment.trim() || submittingComment" class="bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition-all disabled:opacity-50 flex items-center gap-1.5">
                    <svg v-if="submittingComment" class="animate-spin w-3 h-3" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
                    Kirim
                  </button>
                </div>
              </div>
            </div>

            <!-- Comments List -->
            <div v-if="loadingComments" class="space-y-3">
              <div v-for="i in 2" :key="i" class="flex gap-3 animate-pulse">
                <div class="w-8 h-8 rounded-full bg-gray-200"></div>
                <div class="flex-1 space-y-1"><div class="h-3 bg-gray-200 rounded w-24"></div><div class="h-3 bg-gray-100 dark:bg-gray-700 rounded w-full"></div></div>
              </div>
            </div>

            <div v-else-if="comments.length === 0" class="text-center py-4">
              <p class="text-sm text-gray-400 dark:text-gray-500">Belum ada komentar.</p>
            </div>

            <TransitionGroup v-else name="list" tag="div" class="space-y-4">
              <div v-for="comment in comments" :key="comment.id" class="flex gap-3 group">
                <div class="w-8 h-8 rounded-full bg-sky-500 flex items-center justify-center shrink-0">
                  <span class="text-xs text-white font-bold">{{ commentInitial(comment.user_name, comment.user_email) }}</span>
                </div>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2 mb-0.5">
                    <span class="text-sm font-medium text-gray-900 dark:text-white">{{ comment.user_name || comment.user_email }}</span>
                    <span class="text-[11px] text-gray-400 dark:text-gray-500">{{ formatDateTime(comment.created_at) }}</span>
                    <span v-if="comment.edited_at" class="text-[11px] text-gray-400 dark:text-gray-500">(diedit)</span>
                  </div>
                  <p class="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap">{{ comment.body }}</p>
                </div>
                <button v-if="canDeleteComment(comment)" @click="deleteComment(comment.id)" class="opacity-0 group-hover:opacity-100 p-1 text-gray-400 dark:text-gray-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-all shrink-0 self-start" title="Hapus komentar">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                </button>
              </div>
            </TransitionGroup>
          </div>

          <!-- Activity Logs Section -->
          <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
            <h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-4">Aktivitas</h3>

            <div v-if="loadingActivity" class="space-y-3">
              <div v-for="i in 2" :key="i" class="flex gap-3 animate-pulse">
                <div class="w-6 h-6 rounded-full bg-gray-200"></div>
                <div class="flex-1 space-y-1"><div class="h-2 bg-gray-200 rounded w-48"></div><div class="h-2 bg-gray-100 dark:bg-gray-700 rounded w-24"></div></div>
              </div>
            </div>

            <div v-else-if="activityLogs.length === 0" class="text-center py-4">
              <p class="text-sm text-gray-400 dark:text-gray-500">Belum ada aktivitas.</p>
            </div>

            <div v-else class="space-y-3">
              <div v-for="log in activityLogs" :key="log.id" class="flex gap-3">
                <div class="w-6 h-6 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center shrink-0 mt-0.5">
                  <svg v-if="log.action_type === 'created'" class="w-3 h-3 text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
                  <svg v-else-if="log.action_type === 'updated'" class="w-3 h-3 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                  <svg v-else-if="log.action_type === 'deleted'" class="w-3 h-3 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                  <svg v-else class="w-3 h-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                </div>
                <div class="flex-1 min-w-0">
                  <p class="text-sm text-gray-700 dark:text-gray-300">
                    <span class="font-medium text-gray-900 dark:text-white">{{ log.actor_name }}</span>
                    <span class="text-gray-500 dark:text-gray-400">
                      {{ log.action_type === 'created' ? 'membuat' : log.action_type === 'updated' ? 'memperbarui' : log.action_type === 'deleted' ? 'menghapus' : log.action_type }}
                      {{ log.entity_type === 'task' ? 'tugas' : log.entity_type === 'comment' ? 'komentar' : log.entity_type }}
                    </span>
                    <span v-if="log.entity_title" class="font-medium text-gray-900 dark:text-white"> "{{ log.entity_title }}"</span>
                  </p>
                  <p class="text-[11px] text-gray-400 dark:text-gray-500 mt-0.5">{{ formatDateTime(log.created_at) }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Dependencies Section -->
          <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Dependencies</h3>
              <button @click="showAddDep = !showAddDep" class="text-[11px] text-indigo-600 hover:text-indigo-700 font-medium flex items-center gap-1">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
                Tambah
              </button>
            </div>

            <!-- Add Dependency Search -->
            <div v-if="showAddDep" class="mb-4 p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <input v-model="depSearchQuery" type="text" placeholder="Cari task by key atau judul..." class="w-full px-3 py-1.5 border border-gray-200 dark:border-gray-700 rounded-lg text-xs focus:outline-none focus:ring-2 focus:ring-indigo-500" @input="searchDependencies" />
              <div v-if="depSearchResults.length > 0" class="mt-2 space-y-1">
                <button v-for="r in depSearchResults" :key="r.id" @click="addDependency(r)" class="w-full text-left px-2 py-1.5 rounded hover:bg-white transition-colors text-xs flex items-center gap-2">
                  <span class="font-mono text-gray-400 dark:text-gray-500">{{ r.key }}</span>
                  <span class="text-gray-700 dark:text-gray-300 truncate">{{ r.title }}</span>
                </button>
              </div>
            </div>

            <div v-if="loadingDeps" class="space-y-2">
              <div v-for="i in 2" :key="i" class="h-6 bg-gray-100 dark:bg-gray-700 rounded animate-pulse"></div>
            </div>
            <div v-else-if="dependencies.length === 0" class="text-center py-3">
              <p class="text-xs text-gray-400 dark:text-gray-500">Belum ada dependency.</p>
            </div>
            <TransitionGroup v-else name="list" tag="div" class="space-y-2">
              <div v-for="dep in dependencies" :key="dep.id" class="flex items-center gap-2 group">
                <svg class="w-3 h-3 text-amber-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" /></svg>
                <span class="text-xs font-mono text-gray-500 dark:text-gray-400">{{ dep.depends_on_task?.key || 'N/A' }}</span>
                <span class="text-xs text-gray-700 dark:text-gray-300 truncate flex-1">{{ dep.depends_on_task?.title || 'Unknown task' }}</span>
                <button @click="removeDependency(dep.id)" class="opacity-0 group-hover:opacity-100 text-gray-400 dark:text-gray-500 hover:text-red-500 p-0.5 rounded transition-all">
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
                </button>
              </div>
            </TransitionGroup>
          </div>

          <!-- Attachments Section -->
          <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Lampiran</h3>
              <label class="text-[11px] text-indigo-600 hover:text-indigo-700 font-medium flex items-center gap-1 cursor-pointer">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
                Upload
                <input ref="fileInput" type="file" class="hidden" @change="uploadFile" />
              </label>
            </div>

            <div v-if="uploading" class="mb-3 p-3 bg-indigo-50 dark:bg-indigo-900/30 rounded-lg flex items-center gap-2">
              <svg class="animate-spin w-4 h-4 text-indigo-600" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
              <span class="text-xs text-indigo-600">Mengupload file...</span>
            </div>

            <div v-if="loadingAttachments" class="space-y-2">
              <div v-for="i in 2" :key="i" class="h-10 bg-gray-100 dark:bg-gray-700 rounded animate-pulse"></div>
            </div>
            <div v-else-if="attachments.length === 0" class="text-center py-3">
              <p class="text-xs text-gray-400 dark:text-gray-500">Belum ada lampiran.</p>
            </div>
            <TransitionGroup v-else name="list" tag="div" class="space-y-2">
              <div v-for="att in attachments" :key="att.id" class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors group">
                <div class="w-8 h-8 rounded-lg bg-indigo-100 flex items-center justify-center shrink-0">
                  <svg class="w-4 h-4 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" /></svg>
                </div>
                <div class="flex-1 min-w-0">
                  <p class="text-xs font-medium text-gray-700 dark:text-gray-300 truncate">{{ att.original_name }}</p>
                  <p class="text-[10px] text-gray-400 dark:text-gray-500">{{ formatFileSize(att.size_bytes) }}</p>
                </div>
                <button @click="deleteAttachment(att.id)" class="opacity-0 group-hover:opacity-100 text-gray-400 dark:text-gray-500 hover:text-red-500 p-1 rounded transition-all">
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                </button>
              </div>
            </TransitionGroup>
          </div>
        </div>

        <!-- Sidebar -->
        <div class="space-y-4">
          <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4">
            <label class="text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Status</label>
            <div class="mt-2">
              <select v-model="task.status" @change="updateTask" class="w-full text-sm border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all">
                <option v-for="col in columns" :key="col.key" :value="col.key">{{ col.label }}</option>
              </select>
            </div>
          </div>

          <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4">
            <label class="text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Prioritas</label>
            <div class="mt-2">
              <select v-model="task.priority" @change="updateTask" class="w-full text-sm border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all">
                <option value="urgent">Urgent</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>
            </div>
          </div>

          <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4">
            <label class="text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Tanggal</label>
            <div class="mt-2 space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-xs text-gray-500 dark:text-gray-400">Dibuat</span>
                <span class="text-xs text-gray-700 dark:text-gray-300">{{ formatDate(task.created_at) }}</span>
              </div>
              <div v-if="task.due_date" class="flex items-center justify-between">
                <span class="text-xs text-gray-500 dark:text-gray-400">Deadline</span>
                <span class="text-xs" :class="isOverdue(task.due_date) && task.status !== 'done' ? 'text-red-500 dark:text-red-400 font-medium' : 'text-gray-700 dark:text-gray-300'">
                  {{ formatDate(task.due_date) }}
                </span>
              </div>
            </div>
          </div>

          <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4">
            <label class="text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Aksi</label>
            <div class="mt-2">
              <button @click="deleteTask" class="w-full text-left text-sm text-red-500 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 px-3 py-2 rounded-lg transition-colors flex items-center gap-2">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                Hapus Tugas
              </button>
            </div>
          </div>
        </div>
      </div>
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
.list-enter-active {
  transition: all 0.3s ease-out;
}
.list-leave-active {
  transition: all 0.2s ease-in;
}
.list-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.list-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
