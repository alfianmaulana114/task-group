<script setup>
const { user, logout, isAuthenticated } = useAuth()
const { get, post } = useApi()
const router = useRouter()
const route = useRoute()
const config = useRuntimeConfig()
const { isDark, toggle: toggleTheme } = useTheme()

const sidebarOpen = ref(false)
const showUserMenu = ref(false)
const chatOpen = ref(false)
const chatInput = ref('')
const chatMessagesRef = ref(null)

// Search
const searchQuery = ref('')
const searchResults = ref([])
const searchLoading = ref(false)
const showSearchResults = ref(false)
let searchDebounce = null

async function doSearch() {
  const q = searchQuery.value.trim()
  if (!q || q.length < 2) { searchResults.value = []; showSearchResults.value = false; return }
  searchLoading.value = true
  try {
    searchResults.value = await get(`/api/v1/search?q=${encodeURIComponent(q)}`)
    showSearchResults.value = true
  } catch { searchResults.value = [] } finally { searchLoading.value = false }
}

function onSearchInput() {
  clearTimeout(searchDebounce)
  searchDebounce = setTimeout(doSearch, 300)
}

function goToSearchResult(item) {
  showSearchResults.value = false
  searchQuery.value = ''
  if (item.type === 'task') router.push(`/tasks/${item.id}`)
  else if (item.type === 'project') router.push(`/projects/${item.slug || item.id}`)
  else if (item.type === 'comment') router.push(`/tasks/${item.project_id}`)
}

function showSearch() {
  if (searchResults.value.length) showSearchResults.value = true
}

function hideSearchResults() {
  setTimeout(() => { showSearchResults.value = false }, 200)
}

// Notifications (mock for now)
const notifications = ref([])
const showNotifications = ref(false)
const unreadCount = computed(() => notifications.value.filter(n => !n.read).length)

// chat: project-aware
const currentProjectId = ref(null)
const currentProjectSlug = computed(() => route.params.slug || '')
const chatChannelName = ref('Chat')
const chatMessages = ref([])
const chatSending = ref(false)
const avatarColors = ['bg-violet-500','bg-sky-500','bg-emerald-500','bg-rose-500','bg-amber-500']

function avatarFor(name) {
  let h=0; for(let i=0;i<name.length;i++) h = (h*31 + name.charCodeAt(i))|0
  return avatarColors[Math.abs(h)%avatarColors.length]
}
function timeLabel(iso) {
  try { return new Date(iso).toLocaleTimeString('id-ID',{hour:'2-digit',minute:'2-digit'}) } catch { return '' }
}

let pollTimer = null
let ws = null
let resolveSeq = 0

async function fetchChat() {
  if (!currentProjectId.value) return
  try {
    const data = await get(`/api/v1/projects/${currentProjectId.value}/chat`)
    chatMessages.value = data.messages || []
    chatChannelName.value = data.channel?.name || 'Chat'
    nextTick(() => { if(chatMessagesRef.value) chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight })
  } catch {}
}
async function sendChat() {
  const text = chatInput.value.trim()
  if (!text || !currentProjectId.value) return
  chatSending.value = true
  try {
    const msg = await post(`/api/v1/projects/${currentProjectId.value}/chat`, { message: text, message_type: 'text' })
    chatMessages.value.push(msg)
    chatInput.value = ''
    nextTick(() => { if(chatMessagesRef.value) chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight })
  } catch (e) {
  } finally { chatSending.value = false }
}
function connectWS() {
  if (!import.meta.client || !currentProjectId.value) return
  if (!chatOpen.value) return
  const token = localStorage.getItem('auth:token')
  if (!token) return
  if (ws) try{ ws.close() }catch{}
  const base = config.public.apiBaseUrl.replace(/^http/, 'ws')
  const url = `${base}/api/v1/projects/${currentProjectId.value}/chat/ws?token=${encodeURIComponent(token)}`
  try {
    ws = new WebSocket(url)
    ws.onopen = () => { if (pollTimer) { clearInterval(pollTimer); pollTimer = null } }
    ws.onmessage = (ev) => {
      try {
        const parts = String(ev.data).split('\n').filter(Boolean)
        for (const p of parts) {
          const env = JSON.parse(p)
          const msg = env.data?.id ? env.data : env
          if (msg?.id && !chatMessages.value.find((m)=>m.id===msg.id)) {
            chatMessages.value.push(msg)
            nextTick(() => { if(chatMessagesRef.value) chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight })
          }
        }
      } catch {}
    }
    ws.onclose = () => {
      if (currentProjectId.value && chatOpen.value) {
        startChatPolling()
        setTimeout(() => { if(currentProjectId.value && chatOpen.value) connectWS() }, 3000)
      }
    }
  } catch {}
}
function startChatPolling() {
  if (!chatOpen.value || !currentProjectId.value) return
  if (pollTimer) clearInterval(pollTimer)
  pollTimer = setInterval(fetchChat, 3000)
}
function stopChat() {
  if (ws) try{ ws.close() }catch{}
  ws = null
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
}
async function resolveProjectId() {
  const seq = ++resolveSeq
  const slug = currentProjectSlug.value
  if (!slug) { currentProjectId.value = null; stopChat(); return }
  try {
    const projects = await get('/api/v1/projects')
    if (seq !== resolveSeq) return
    const found = projects.find((p) => p.slug === slug)
    currentProjectId.value = found?.id || null
    if (currentProjectId.value && chatOpen.value) {
      await fetchChat()
      connectWS()
      startChatPolling()
    } else if (!chatOpen.value) {
      chatMessages.value = []
    }
  } catch {}
}
watch(() => route.params.slug, () => { stopChat(); chatMessages.value = []; resolveProjectId() })
watch(chatOpen, (open) => {
  if (open && currentProjectId.value) { fetchChat(); connectWS(); startChatPolling() }
  else if (!open) { stopChat() }
})

const orgInfo = ref(null)
const projects = ref([])
const projectsLoading = ref(false)

const currentOrg = computed(() => {
  if (orgInfo.value) return { name: orgInfo.value.name, slug: orgInfo.value.slug, plan: orgInfo.value.plan }
  const orgIdVal = user.value?.org_id || (import.meta.client ? localStorage.getItem('auth:org_id') : null)
  if (orgIdVal) return { name: orgIdVal.slice(0, 8) + '…', slug: orgIdVal, plan: 'free' }
  return { name: 'Acme Corp', slug: 'acme-corp', plan: 'free' }
})

const projectColors = ['bg-indigo-500','bg-emerald-500','bg-amber-500','bg-rose-500','bg-sky-500','bg-violet-500']
function projectColor(i) { return projectColors[i % projectColors.length] }

async function fetchSidebarData() {
  const oid = user.value?.org_id || (import.meta.client ? localStorage.getItem('auth:org_id') : null)
  if (!oid) {
    if (import.meta.client && window.location.pathname !== '/organizations') { router.push('/organizations') }
    return
  }
  try {
    const orgs = await get('/api/v1/organizations')
    const found = orgs.find(o => o.id === oid)
    if (found) orgInfo.value = found
  } catch {}
  try {
    projectsLoading.value = true
    projects.value = await get('/api/v1/projects')
  } catch {} finally { projectsLoading.value = false }
}

const navItems = [
  { label: 'Dashboard', to: '/dashboard', icon: 'home' },
  { label: 'Projects', to: '/projects', icon: 'folder' },
  { label: 'My Tasks', to: '/my-tasks', icon: 'check-square' },
  { label: 'Members', to: '/members', icon: 'users' },
  { label: 'Calendar', to: '/calendar', icon: 'calendar' },
]
const bottomNavItems = [{ label: 'Settings', to: '/settings', icon: 'cog' }]

function isActive(to) {
  if (to === '/dashboard') return route.path === to
  if (to === '/projects') return route.path === '/projects' || route.path.startsWith('/projects/')
  return route.path.startsWith(to)
}
function isProjectActive(slug) { return route.params.slug === slug }
function handleLogout() { showUserMenu.value = false; logout() }

onMounted(() => {
  const hasToken = isAuthenticated.value || (import.meta.client ? !!localStorage.getItem('auth:token') : false)
  if (!hasToken) { router.push('/login'); return }
  watch(isAuthenticated, (val) => { if (!val) router.push('/login') })
  fetchSidebarData()
  watch(() => user.value?.org_id, () => fetchSidebarData())
  resolveProjectId()
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (ws) try{ ws.close() }catch{}
})
</script>

<template>
  <div class="h-screen bg-gray-100 dark:bg-gray-900 flex overflow-hidden transition-colors">
    <!-- Sidebar -->
    <aside
      class="fixed inset-y-0 left-0 z-50 w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col transform transition-transform duration-200 ease-in-out lg:translate-x-0 lg:static lg:inset-auto shrink-0"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <!-- Org Header -->
      <div class="h-14 flex items-center gap-3 px-5 border-b border-gray-200 dark:border-gray-700 shrink-0">
        <div class="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center shrink-0">
          <span class="text-white font-bold text-sm">{{ (currentOrg.name || 'A').charAt(0).toUpperCase() }}</span>
        </div>
        <div class="flex-1 min-w-0">
          <div class="text-sm font-semibold text-gray-900 dark:text-white truncate">{{ currentOrg.name }}</div>
          <div class="text-[11px] text-gray-400 dark:text-gray-500 capitalize">{{ currentOrg.plan || 'free' }} plan</div>
        </div>
        <NuxtLink to="/organizations" class="text-gray-300 dark:text-gray-600 hover:text-gray-500 p-1 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors shrink-0" title="Ganti organisasi">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l4-4 4 4m0 6l-4 4-4-4"/></svg>
        </NuxtLink>
      </div>

      <!-- Nav -->
      <nav class="flex-1 overflow-y-auto px-3 py-3 space-y-0.5">
        <NuxtLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="relative flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all duration-150"
          :class="isActive(item.to) ? 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-white'"
          @click="sidebarOpen = false"
        >
          <div v-if="isActive(item.to)" class="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 bg-indigo-600 rounded-r-full"></div>
          <svg v-if="item.icon === 'home'" class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" /></svg>
          <svg v-if="item.icon === 'folder'" class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" /></svg>
          <svg v-if="item.icon === 'check-square'" class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" /></svg>
          <svg v-if="item.icon === 'users'" class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
          <svg v-if="item.icon === 'calendar'" class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
          {{ item.label }}
        </NuxtLink>

        <!-- Projects Section -->
        <div class="pt-4">
          <div class="px-3 mb-1.5 flex items-center justify-between">
            <span class="text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Projects</span>
            <NuxtLink to="/projects" class="text-gray-400 hover:text-gray-600 p-0.5 transition-colors" title="Lihat semua">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
            </NuxtLink>
          </div>
          <div v-if="projectsLoading" class="px-3 py-2 space-y-2">
            <div v-for="i in 3" :key="i" class="h-6 bg-gray-100 dark:bg-gray-700 rounded-lg animate-pulse"></div>
          </div>
          <div v-else-if="projects.length===0" class="px-3 py-2">
            <p class="text-[11px] text-gray-400 mb-2">Belum ada proyek</p>
            <NuxtLink to="/projects" class="text-[11px] text-indigo-600 hover:text-indigo-700 font-medium">+ Buat proyek</NuxtLink>
          </div>
          <template v-else>
            <NuxtLink
              v-for="(project,i) in projects"
              :key="project.id"
              :to="`/projects/${project.slug}`"
              class="flex items-center gap-2.5 px-3 py-1.5 rounded-lg text-[13px] font-medium cursor-pointer transition-all duration-150"
              :class="isProjectActive(project.slug) ? 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-white'"
              @click="sidebarOpen = false"
            >
              <div class="w-2.5 h-2.5 rounded-sm shrink-0" :class="projectColor(i)"></div>
              <span class="truncate">{{ project.name }}</span>
            </NuxtLink>
          </template>
        </div>
      </nav>

      <!-- Bottom Nav -->
      <div class="border-t border-gray-200 dark:border-gray-700 px-3 py-2 space-y-0.5 shrink-0">
        <NuxtLink
          v-for="item in bottomNavItems"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
          :class="isActive(item.to) ? 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-white'"
          @click="sidebarOpen = false"
        >
          <svg v-if="item.icon === 'cog'" class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
          {{ item.label }}
        </NuxtLink>
      </div>
    </aside>

    <!-- Mobile overlay -->
    <Transition name="fade">
      <div v-if="sidebarOpen" class="fixed inset-0 z-40 bg-black/50 lg:hidden" @click="sidebarOpen = false"></div>
    </Transition>

    <!-- Main Content -->
    <div class="flex-1 flex flex-col min-w-0 h-screen">
      <!-- Top Bar -->
      <header class="h-12 bg-white dark:bg-gray-800 border-b border-gray-100 dark:border-gray-700 flex items-center px-4 lg:px-6 shrink-0 transition-colors">
        <button @click="sidebarOpen = true" class="lg:hidden text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-white p-1 -ml-1 transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" /></svg>
        </button>

        <!-- Search Bar -->
        <div class="flex-1 flex justify-center relative">
          <div class="relative w-full max-w-xs">
            <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
            <input v-model="searchQuery" type="text" placeholder="Cari tugas, proyek..." class="w-full pl-8 pr-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all" @input="onSearchInput" @focus="showSearch" @blur="hideSearchResults" />
          </div>
          <!-- Search Results Dropdown -->
          <Transition name="dropdown">
            <div v-if="showSearchResults && (searchResults.length > 0 || searchLoading)" class="absolute top-full mt-1 w-full max-w-xs bg-white dark:bg-gray-800 rounded-xl shadow-lg border border-gray-200 dark:border-gray-700 py-2 z-50 max-h-64 overflow-y-auto">
            <div v-if="searchLoading" class="px-4 py-2 text-xs text-gray-400">Mencari...</div>
            <template v-else>
              <button v-for="item in searchResults" :key="item.id" @mousedown.prevent="goToSearchResult(item)" class="w-full text-left px-4 py-2 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors flex items-center gap-3">
                <div class="w-6 h-6 rounded flex items-center justify-center shrink-0" :class="item.type === 'task' ? 'bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600' : 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600'">
                  <svg v-if="item.type === 'task'" class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2" /></svg>
                  <svg v-else class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" /></svg>
                </div>
                <div class="min-w-0">
                  <div class="text-xs font-medium text-gray-900 dark:text-white truncate">{{ item.title || item.name }}</div>
                  <div class="text-[10px] text-gray-400">{{ item.type === 'task' ? item.key : 'Proyek' }}</div>
                </div>
              </button>
            </template>
          </div>
          </Transition>
        </div>

        <div class="flex items-center gap-2">
          <!-- Notification Bell -->
          <button @click="showNotifications = !showNotifications" class="relative p-1.5 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /></svg>
            <span v-if="unreadCount > 0" class="absolute -top-0.5 -right-0.5 w-4 h-4 bg-red-500 text-white text-[8px] font-bold rounded-full flex items-center justify-center">{{ unreadCount > 9 ? '9+' : unreadCount }}</span>
          </button>

          <!-- Dark Mode Toggle -->
          <button @click="toggleTheme" class="p-1.5 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors" :title="isDark ? 'Mode terang' : 'Mode gelap'">
            <svg v-if="!isDark" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" /></svg>
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" /></svg>
          </button>

          <!-- User Menu -->
          <div class="relative">
            <button @click="showUserMenu = !showUserMenu" class="flex items-center gap-2 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-white transition-colors relative">
              <span class="text-xs font-medium hidden sm:block">{{ user?.full_name || 'User' }}</span>
              <div class="w-7 h-7 rounded-full bg-violet-500 flex items-center justify-center hover:ring-2 hover:ring-violet-200 dark:hover:ring-violet-500 transition-all">
                <span class="text-[10px] text-white font-medium">{{ user?.full_name?.charAt(0) || 'U' }}</span>
              </div>
            </button>
            <Transition enter-active-class="transition-all duration-150 ease-out" enter-from-class="opacity-0 scale-95 -translate-y-1" enter-to-class="opacity-100 scale-100 translate-y-0" leave-active-class="transition-all duration-100 ease-in" leave-from-class="opacity-100 scale-100 translate-y-0" leave-to-class="opacity-0 scale-95 -translate-y-1">
              <div v-if="showUserMenu" class="absolute right-0 top-full mt-2 w-48 bg-white dark:bg-gray-800 rounded-xl shadow-lg border border-gray-200 dark:border-gray-700 py-1 z-50">
                <div class="px-4 py-2 border-b border-gray-100 dark:border-gray-700">
                  <p class="text-sm font-medium text-gray-900 dark:text-white">{{ user?.full_name || 'User' }}</p>
                  <p class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ user?.email || '' }}</p>
                </div>
                <NuxtLink to="/settings" class="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors" @click="showUserMenu = false">
                  <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
                  Pengaturan
                </NuxtLink>
                <button class="w-full flex items-center gap-2 px-4 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors" @click="handleLogout">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
                  Keluar
                </button>
              </div>
            </Transition>
          </div>
        </div>
      </header>

      <!-- Notification Dropdown -->
      <Transition name="fade">
        <div v-if="showNotifications" class="fixed inset-0 z-40" @click="showNotifications = false"></div>
      </Transition>
      <Transition name="dropdown">
        <div v-if="showNotifications" class="fixed top-12 right-4 lg:right-6 w-80 bg-white dark:bg-gray-800 rounded-xl shadow-lg border border-gray-200 dark:border-gray-700 z-50 max-h-96 overflow-y-auto">
        <div class="px-4 py-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
          <h3 class="text-sm font-bold text-gray-900 dark:text-white">Notifikasi</h3>
          <span class="text-[11px] text-gray-400">{{ unreadCount }} belum dibaca</span>
        </div>
        <div v-if="notifications.length === 0" class="px-4 py-8 text-center">
          <svg class="w-8 h-8 mx-auto text-gray-300 dark:text-gray-600 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /></svg>
          <p class="text-xs text-gray-400">Belum ada notifikasi</p>
        </div>
        <div v-else class="divide-y divide-gray-50 dark:divide-gray-700">
          <div v-for="n in notifications" :key="n.id" class="px-4 py-3 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer" :class="{ 'bg-indigo-50/50 dark:bg-indigo-900/10': !n.read }">
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ n.title }}</p>
            <p class="text-[11px] text-gray-400 mt-0.5">{{ n.time }}</p>
          </div>
        </div>
      </div>
      </Transition>

      <!-- Click outside to close dropdown -->
      <div v-if="showUserMenu" class="fixed inset-0 z-40" @click="showUserMenu = false"></div>

      <!-- Page Content -->
      <main class="flex-1 overflow-y-auto">
        <slot />
      </main>
    </div>

    <!-- Chat Widget -->
    <div class="fixed bottom-5 right-5 z-50">
      <div v-if="chatOpen" class="w-80 bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 flex flex-col overflow-hidden transition-colors" style="height: 420px;">
        <div class="h-11 bg-indigo-600 flex items-center justify-between px-4 shrink-0">
          <div class="flex items-center gap-2 min-w-0">
            <svg class="w-4 h-4 text-indigo-200 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8h2a2 2 0 012 2v6a2 2 0 01-2 2h-2v4l-4-4H9a1.994 1.994 0 01-1.414-.586m0 0L11 14h4a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2v4l.586-.586z" /></svg>
            <span class="text-sm font-semibold text-white truncate">{{ chatChannelName }}</span>
          </div>
          <button @click="chatOpen = false" class="text-indigo-200 hover:text-white transition-colors shrink-0">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" /></svg>
          </button>
        </div>
        <div v-if="!currentProjectId" class="flex-1 flex items-center justify-center p-4 text-center">
          <p class="text-xs text-gray-400 dark:text-gray-500">Buka sebuah project untuk mulai chat.</p>
        </div>
        <template v-else>
          <div ref="chatMessagesRef" class="flex-1 overflow-y-auto px-4 py-3 space-y-3">
            <div v-if="chatMessages.length===0" class="text-center py-8">
              <p class="text-xs text-gray-400 dark:text-gray-500">Belum ada pesan.</p>
            </div>
            <div v-for="msg in chatMessages" :key="msg.id" class="flex items-start gap-2.5">
              <div class="w-6 h-6 rounded-full shrink-0 flex items-center justify-center" :class="avatarFor(msg.user?.full_name || msg.user_id?.slice(0,2) || 'U')">
                <span class="text-[8px] text-white font-bold">{{ (msg.user?.full_name || msg.user_id || 'U').charAt(0).toUpperCase() }}</span>
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex items-baseline gap-2">
                  <span class="text-xs font-semibold text-gray-900 dark:text-white truncate">{{ msg.user?.full_name || msg.user_id?.slice(0,8) || 'Unknown' }}</span>
                  <span class="text-[10px] text-gray-400">{{ timeLabel(msg.created_at) }}</span>
                </div>
                <p class="text-sm text-gray-600 dark:text-gray-300 mt-0.5 break-words">{{ msg.message }}</p>
              </div>
            </div>
          </div>
          <div class="border-t border-gray-100 dark:border-gray-700 px-3 py-2.5 shrink-0">
            <div class="flex items-center gap-2">
              <input v-model="chatInput" type="text" placeholder="Ketik pesan..." class="flex-1 px-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-xs text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-indigo-500" @keydown.enter="sendChat" />
              <button :disabled="chatSending || !chatInput.trim()" class="p-1.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 rounded-lg transition-colors" @click="sendChat">
                <svg v-if="!chatSending" class="w-3.5 h-3.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
                <svg v-else class="animate-spin w-3.5 h-3.5 text-white" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
              </button>
            </div>
          </div>
        </template>
      </div>
      <button v-if="!chatOpen" @click="chatOpen = true" class="w-12 h-12 bg-indigo-600 hover:bg-indigo-700 rounded-full shadow-lg flex items-center justify-center transition-colors">
        <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
      </button>
    </div>
  </div>
</template>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.dropdown-enter-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.dropdown-leave-active {
  transition: opacity 0.1s ease, transform 0.1s ease;
}
.dropdown-enter-from {
  opacity: 0;
  transform: translateY(-4px) scale(0.98);
}
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.98);
}
</style>
