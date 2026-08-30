<script setup>
definePageMeta({ layout: 'dashboard' })

const projects = ref([
  { id: '1', name: 'Website Redesign', slug: 'website-redesign', description: 'Complete overhaul of the marketing site with new brand guidelines', status: 'active', taskCounts: { backlog: 5, todo: 8, in_progress: 3, review: 2, done: 12 }, members: [{ name: 'Alice', color: 'bg-violet-500' }, { name: 'Bob', color: 'bg-sky-500' }, { name: 'Charlie', color: 'bg-emerald-500' }], createdAt: '2026-07-15' },
  { id: '2', name: 'Mobile App', slug: 'mobile-app', description: 'Cross-platform mobile app for iOS and Android', status: 'active', taskCounts: { backlog: 3, todo: 4, in_progress: 6, review: 1, done: 8 }, members: [{ name: 'Diana', color: 'bg-rose-500' }, { name: 'Alice', color: 'bg-violet-500' }], createdAt: '2026-08-01' },
  { id: '3', name: 'API Backend', slug: 'api-backend', description: 'Core API service and microservices infrastructure', status: 'active', taskCounts: { backlog: 2, todo: 3, in_progress: 4, review: 3, done: 15 }, members: [{ name: 'Bob', color: 'bg-sky-500' }, { name: 'Charlie', color: 'bg-emerald-500' }, { name: 'Diana', color: 'bg-rose-500' }], createdAt: '2026-06-20' },
  { id: '4', name: 'Design System', slug: 'design-system', description: 'Reusable component library and design tokens', status: 'active', taskCounts: { backlog: 1, todo: 2, in_progress: 1, review: 0, done: 6 }, members: [{ name: 'Alice', color: 'bg-violet-500' }, { name: 'Diana', color: 'bg-rose-500' }], createdAt: '2026-08-10' },
  { id: '5', name: 'Legacy Migration', slug: 'legacy-migration', description: 'Migrate old monolith services to new architecture', status: 'archived', taskCounts: { backlog: 0, todo: 0, in_progress: 0, review: 0, done: 24 }, members: [{ name: 'Bob', color: 'bg-sky-500' }], createdAt: '2026-05-01' },
])

const viewMode = ref('grid')
const searchQuery = ref('')

const filteredProjects = computed(() => {
  if (!searchQuery.value) return projects.value
  return projects.value.filter(p => p.name.toLowerCase().includes(searchQuery.value.toLowerCase()))
})

function getProjectColor(i) {
  return ['bg-indigo-500', 'bg-emerald-500', 'bg-amber-500', 'bg-rose-500', 'bg-sky-500'][i % 5]
}

function getTotalTasks(counts) {
  return Object.values(counts).reduce((a, b) => a + b, 0)
}
</script>

<template>
  <div class="p-4 lg:p-6 max-w-7xl mx-auto w-full">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-bold text-gray-900">Projects</h1>
        <p class="mt-0.5 text-sm text-gray-500">Manage and track all team projects.</p>
      </div>
      <button class="bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
        New Project
      </button>
    </div>

    <!-- Filters -->
    <div class="flex items-center gap-3 mb-5">
      <div class="flex-1 max-w-sm relative">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
        <input v-model="searchQuery" type="text" placeholder="Search projects..." class="w-full pl-10 pr-4 py-2 bg-white border border-gray-200 rounded-lg text-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
      </div>
      <div class="flex bg-white border border-gray-200 rounded-lg overflow-hidden">
        <button @click="viewMode = 'grid'" class="p-2" :class="viewMode === 'grid' ? 'bg-indigo-50 text-indigo-600' : 'text-gray-400 hover:text-gray-600'">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" /></svg>
        </button>
        <button @click="viewMode = 'list'" class="p-2" :class="viewMode === 'list' ? 'bg-indigo-50 text-indigo-600' : 'text-gray-400 hover:text-gray-600'">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" /></svg>
        </button>
      </div>
    </div>

    <!-- Grid View -->
    <div v-if="viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      <NuxtLink
        v-for="(project, i) in filteredProjects"
        :key="project.id"
        :to="`/projects/${project.slug}`"
        class="bg-white rounded-xl border border-gray-200 p-5 hover:shadow-md transition-shadow"
      >
        <div class="flex items-start justify-between">
          <div class="flex items-center gap-2.5">
            <div class="w-3 h-3 rounded-sm" :class="getProjectColor(i)"></div>
            <h3 class="text-sm font-semibold text-gray-900">{{ project.name }}</h3>
          </div>
          <span v-if="project.status === 'archived'" class="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 text-gray-500 font-medium">Archived</span>
        </div>
        <p class="mt-2 text-xs text-gray-500 line-clamp-2">{{ project.description }}</p>

        <div class="mt-4 h-1 rounded-full bg-gray-100 overflow-hidden">
          <div class="h-full bg-indigo-500 rounded-full" :style="{ width: (project.taskCounts.done / getTotalTasks(project.taskCounts) * 100) + '%' }"></div>
        </div>

        <div class="mt-2 flex items-center justify-between">
          <span class="text-[11px] text-gray-400">{{ getTotalTasks(project.taskCounts) }} tasks</span>
          <div class="flex -space-x-1.5">
            <div v-for="(m, j) in project.members.slice(0, 4)" :key="j" class="w-6 h-6 rounded-full flex items-center justify-center border-2 border-white" :class="m.color">
              <span class="text-[8px] text-white font-bold">{{ m.name.charAt(0) }}</span>
            </div>
          </div>
        </div>
      </NuxtLink>
    </div>

    <!-- List View -->
    <div v-else class="bg-white rounded-xl border border-gray-200 overflow-hidden">
      <table class="w-full">
        <thead>
          <tr class="border-b border-gray-100">
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5">Project</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 hidden md:table-cell">Tasks</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 hidden lg:table-cell">Members</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 hidden sm:table-cell">Created</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-50">
          <tr v-for="(project, i) in filteredProjects" :key="project.id" class="hover:bg-gray-50 transition-colors cursor-pointer">
            <td class="px-4 py-3">
              <div class="flex items-center gap-2.5">
                <div class="w-2.5 h-2.5 rounded-sm shrink-0" :class="getProjectColor(i)"></div>
                <div>
                  <div class="text-sm font-medium text-gray-900">{{ project.name }}</div>
                  <div class="text-xs text-gray-400 truncate max-w-xs">{{ project.description }}</div>
                </div>
              </div>
            </td>
            <td class="px-4 py-3 hidden md:table-cell">
              <span class="text-sm text-gray-600">{{ getTotalTasks(project.taskCounts) }}</span>
            </td>
            <td class="px-4 py-3 hidden lg:table-cell">
              <div class="flex -space-x-1.5">
                <div v-for="(m, j) in project.members.slice(0, 3)" :key="j" class="w-5 h-5 rounded-full flex items-center justify-center border-2 border-white" :class="m.color">
                  <span class="text-[7px] text-white font-bold">{{ m.name.charAt(0) }}</span>
                </div>
              </div>
            </td>
            <td class="px-4 py-3 hidden sm:table-cell">
              <span class="text-xs text-gray-400">{{ new Date(project.createdAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
