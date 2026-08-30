<script setup>
const { user, logout, isAuthenticated, initAuth } = useAuth()
const router = useRouter()
const route = useRoute()

const sidebarOpen = ref(false)

const currentOrg = { name: 'Acme Corp', slug: 'acme-corp' }

const projects = ref([
  { id: '1', name: 'Website Redesign', slug: 'website-redesign', color: 'bg-indigo-500' },
  { id: '2', name: 'Mobile App', slug: 'mobile-app', color: 'bg-emerald-500' },
  { id: '3', name: 'API Backend', slug: 'api-backend', color: 'bg-amber-500' },
])

const navItems = [
  { label: 'Dashboard', to: '/dashboard', icon: 'home' },
  { label: 'Projects', to: '/projects', icon: 'folder' },
  { label: 'My Tasks', to: '/my-tasks', icon: 'check-square' },
]

const bottomNavItems = [
  { label: 'Settings', to: '/settings', icon: 'cog' },
  { label: 'Members', to: '/members', icon: 'users' },
]

function isActive(to) {
  return route.path === to
}

const chatOpen = ref(false)
const chatMessage = ref('')
const chatMessages = ref([
  { id: 1, user: 'Bob', avatar: 'bg-sky-500', text: 'Hey, the new landing page looks great!', time: '10:32 AM' },
  { id: 2, user: 'Diana', avatar: 'bg-rose-500', text: 'Thanks! Just pushed the final changes.', time: '10:34 AM' },
  { id: 3, user: 'Alice', avatar: 'bg-violet-500', text: 'Nice work team. Let\'s review it in the standup.', time: '10:36 AM' },
])

onMounted(() => {
  initAuth()
  const unwatch = watch(isAuthenticated, (val) => {
    if (!val) {
      router.push('/login')
    }
    unwatch()
  })
})
</script>

<template>
  <div class="h-screen bg-gray-100 flex overflow-hidden">
    <!-- Sidebar -->
    <aside
      class="fixed inset-y-0 left-0 z-50 w-64 bg-white border-r border-gray-200 flex flex-col transform transition-transform duration-200 ease-in-out lg:translate-x-0 lg:static lg:inset-auto shrink-0"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <!-- Org Header (fixed) -->
      <div class="h-14 flex items-center gap-3 px-5 border-b border-gray-200 shrink-0">
        <div class="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center shrink-0">
          <span class="text-white font-bold text-sm">A</span>
        </div>
        <div class="flex-1 min-w-0">
          <div class="text-sm font-semibold text-gray-900 truncate">{{ currentOrg.name }}</div>
          <div class="text-[11px] text-gray-400">Free plan</div>
        </div>
      </div>

      <!-- Nav (scrollable) -->
      <nav class="flex-1 overflow-y-auto px-3 py-3 space-y-0.5">
        <NuxtLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
          :class="isActive(item.to) ? 'bg-indigo-50 text-indigo-700' : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
          @click="sidebarOpen = false"
        >
          <!-- Home -->
          <svg v-if="item.icon === 'home'" class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" /></svg>
          <!-- Folder -->
          <svg v-if="item.icon === 'folder'" class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" /></svg>
          <!-- Check Square -->
          <svg v-if="item.icon === 'check-square'" class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" /></svg>
          {{ item.label }}
          <span v-if="item.label === 'My Tasks'" class="ml-auto bg-indigo-100 text-indigo-700 text-[11px] font-semibold px-2 py-0.5 rounded-full">5</span>
        </NuxtLink>

        <!-- Projects Section -->
        <div class="pt-4">
          <div class="px-3 mb-1.5 flex items-center justify-between">
            <span class="text-[11px] font-semibold text-gray-400 uppercase tracking-wider">Projects</span>
            <button class="text-gray-400 hover:text-gray-600 p-0.5">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
            </button>
          </div>
          <NuxtLink
            v-for="project in projects"
            :key="project.id"
            :to="`/projects/${project.slug}`"
            class="flex items-center gap-2.5 px-3 py-1.5 rounded-lg text-[13px] font-medium text-gray-600 hover:bg-gray-50 hover:text-gray-900 cursor-pointer transition-colors"
            @click="sidebarOpen = false"
          >
            <div class="w-2.5 h-2.5 rounded-sm shrink-0" :class="project.color"></div>
            <span class="truncate">{{ project.name }}</span>
          </NuxtLink>
        </div>
      </nav>

      <!-- Bottom Nav (fixed) -->
      <div class="border-t border-gray-200 px-3 py-2 space-y-0.5 shrink-0">
        <NuxtLink
          v-for="item in bottomNavItems"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
          :class="isActive(item.to) ? 'bg-indigo-50 text-indigo-700' : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
          @click="sidebarOpen = false"
        >
          <!-- Cog -->
          <svg v-if="item.icon === 'cog'" class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
          <!-- Users -->
          <svg v-if="item.icon === 'users'" class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
          {{ item.label }}
        </NuxtLink>
      </div>
    </aside>

    <!-- Mobile overlay -->
    <div v-if="sidebarOpen" class="fixed inset-0 z-40 bg-black/50 lg:hidden" @click="sidebarOpen = false"></div>

    <!-- Main Content -->
    <div class="flex-1 flex flex-col min-w-0 h-screen">
      <!-- Top Bar -->
      <header class="h-12 bg-white border-b border-gray-100 flex items-center px-4 lg:px-6 shrink-0">
        <button @click="sidebarOpen = true" class="lg:hidden text-gray-400 hover:text-gray-600 p-1 -ml-1">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" /></svg>
        </button>

        <div class="flex-1 flex justify-center">
          <div class="relative w-full max-w-xs">
            <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
            <input type="text" placeholder="Search..." class="w-full pl-8 pr-3 py-1.5 bg-gray-50 border border-gray-200 rounded-md text-xs text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500" />
          </div>
        </div>

        <button @click="logout" class="flex items-center gap-2 text-gray-500 hover:text-gray-700 transition-colors">
          <span class="text-xs font-medium hidden sm:block">{{ user?.full_name || 'Alice' }}</span>
          <div class="w-7 h-7 rounded-full bg-violet-500 flex items-center justify-center">
            <span class="text-[10px] text-white font-medium">{{ user?.full_name?.charAt(0) || 'A' }}</span>
          </div>
        </button>
      </header>

      <!-- Page Content -->
      <main class="flex-1 overflow-y-auto">
        <slot />
      </main>
    </div>

    <!-- Chat Widget -->
    <div class="fixed bottom-5 right-5 z-50">
      <!-- Expanded Chat -->
      <div v-if="chatOpen" class="w-80 bg-white rounded-xl shadow-2xl border border-gray-200 flex flex-col overflow-hidden" style="height: 420px;">
        <!-- Chat Header -->
        <div class="h-11 bg-indigo-600 flex items-center justify-between px-4 shrink-0">
          <div class="flex items-center gap-2">
            <svg class="w-4 h-4 text-indigo-200" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8h2a2 2 0 012 2v6a2 2 0 01-2 2h-2v4l-4-4H9a1.994 1.994 0 01-1.414-.586m0 0L11 14h4a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2v4l.586-.586z" /></svg>
            <span class="text-sm font-semibold text-white">Website Redesign</span>
          </div>
          <button @click="chatOpen = false" class="text-indigo-200 hover:text-white transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" /></svg>
          </button>
        </div>

        <!-- Chat Messages -->
        <div class="flex-1 overflow-y-auto px-4 py-3 space-y-3">
          <div v-for="msg in chatMessages" :key="msg.id" class="flex items-start gap-2.5">
            <div class="w-6 h-6 rounded-full shrink-0 flex items-center justify-center" :class="msg.avatar">
              <span class="text-[8px] text-white font-bold">{{ msg.user.charAt(0) }}</span>
            </div>
            <div class="min-w-0">
              <div class="flex items-baseline gap-2">
                <span class="text-xs font-semibold text-gray-900">{{ msg.user }}</span>
                <span class="text-[10px] text-gray-400">{{ msg.time }}</span>
              </div>
              <p class="text-sm text-gray-600 mt-0.5">{{ msg.text }}</p>
            </div>
          </div>
        </div>

        <!-- Chat Input -->
        <div class="border-t border-gray-100 px-3 py-2.5 shrink-0">
          <div class="flex items-center gap-2">
            <input
              v-model="chatMessage"
              type="text"
              placeholder="Type a message..."
              class="flex-1 px-3 py-1.5 bg-gray-50 border border-gray-200 rounded-lg text-xs text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              @keydown.enter="if(chatMessage.trim()) { chatMessages.push({ id: Date.now(), user: 'Alice', avatar: 'bg-violet-500', text: chatMessage, time: 'Now' }); chatMessage = '' }"
            />
            <button
              class="p-1.5 bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-colors"
              @click="if(chatMessage.trim()) { chatMessages.push({ id: Date.now(), user: 'Alice', avatar: 'bg-violet-500', text: chatMessage, time: 'Now' }); chatMessage = '' }"
            >
              <svg class="w-3.5 h-3.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
            </button>
          </div>
        </div>
      </div>

      <!-- Collapsed FAB -->
      <button
        v-if="!chatOpen"
        @click="chatOpen = true"
        class="w-12 h-12 bg-indigo-600 hover:bg-indigo-700 rounded-full shadow-lg flex items-center justify-center transition-colors"
      >
        <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
      </button>
    </div>
  </div>
</template>
