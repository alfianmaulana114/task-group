<script setup>
const { user } = useAuth()

definePageMeta({
  layout: 'dashboard',
})

const projects = ref([
  { id: '1', name: 'Website Redesign', slug: 'website-redesign', description: 'Complete overhaul of the marketing site with new brand guidelines', taskCounts: { backlog: 5, todo: 8, in_progress: 3, review: 2, done: 12 }, members: [{ name: 'Alice', color: 'bg-violet-500' }, { name: 'Bob', color: 'bg-sky-500' }, { name: 'Charlie', color: 'bg-emerald-500' }] },
  { id: '2', name: 'Mobile App', slug: 'mobile-app', description: 'Cross-platform mobile app for iOS and Android', taskCounts: { backlog: 3, todo: 4, in_progress: 6, review: 1, done: 8 }, members: [{ name: 'Diana', color: 'bg-rose-500' }, { name: 'Alice', color: 'bg-violet-500' }] },
  { id: '3', name: 'API Backend', slug: 'api-backend', description: 'Core API service and microservices infrastructure', taskCounts: { backlog: 2, todo: 3, in_progress: 4, review: 3, done: 15 }, members: [{ name: 'Bob', color: 'bg-sky-500' }, { name: 'Charlie', color: 'bg-emerald-500' }, { name: 'Diana', color: 'bg-rose-500' }] },
])

const myTasks = ref([
  { id: '1', key: 'TG-45', title: 'Implement dark mode toggle', project: 'Website Redesign', status: 'in_progress', priority: 'high', dueDate: '2026-09-05' },
  { id: '2', key: 'TG-38', title: 'Design onboarding flow wireframes', project: 'Mobile App', status: 'todo', priority: 'medium', dueDate: '2026-09-08' },
  { id: '3', key: 'TG-52', title: 'Fix responsive layout on tablet', project: 'Website Redesign', status: 'review', priority: 'urgent', dueDate: '2026-09-02' },
  { id: '4', key: 'TG-29', title: 'Write unit tests for auth module', project: 'API Backend', status: 'todo', priority: 'medium', dueDate: '2026-09-10' },
  { id: '5', key: 'TG-61', title: 'Update API documentation for v2', project: 'API Backend', status: 'backlog', priority: 'low', dueDate: null },
])

const activities = ref([
  { id: '1', user: 'Bob', avatar: 'bg-sky-500', action: 'moved', target: 'TG-45', detail: 'from Todo to In Progress', project: 'Website Redesign', time: '2 minutes ago' },
  { id: '2', user: 'Diana', avatar: 'bg-rose-500', action: 'created', target: 'TG-62', detail: '"Setup push notification service"', project: 'Mobile App', time: '15 minutes ago' },
  { id: '3', user: 'Charlie', avatar: 'bg-emerald-500', action: 'commented on', target: 'TG-38', detail: '"Looks good, just one minor adjustment needed"', project: 'Mobile App', time: '1 hour ago' },
  { id: '4', user: 'Alice', avatar: 'bg-violet-500', action: 'completed', target: 'TG-22', detail: '"Database migration script"', project: 'API Backend', time: '2 hours ago' },
  { id: '5', user: 'Bob', avatar: 'bg-sky-500', action: 'assigned', target: 'TG-52', detail: 'to Alice', project: 'Website Redesign', time: '3 hours ago' },
  { id: '6', user: 'Diana', avatar: 'bg-rose-500', action: 'created', target: 'TG-61', detail: '"Update API documentation for v2"', project: 'API Backend', time: '5 hours ago' },
])

const statusConfig = {
  backlog: { label: 'Backlog', dot: 'bg-gray-400' },
  todo: { label: 'Todo', dot: 'bg-blue-500' },
  in_progress: { label: 'In Progress', dot: 'bg-amber-500' },
  review: { label: 'In Review', dot: 'bg-purple-500' },
  done: { label: 'Done', dot: 'bg-emerald-500' },
}

const priorityConfig = {
  urgent: 'bg-red-500',
  high: 'bg-orange-400',
  medium: 'bg-gray-400',
  low: 'bg-gray-300',
}

function getProjectColor(i) {
  return ['bg-indigo-500', 'bg-emerald-500', 'bg-amber-500'][i % 3]
}

function isOverdue(dateStr) {
  if (!dateStr) return false
  return new Date(dateStr) < new Date()
}
</script>

<template>
  <div class="p-4 lg:p-6 max-w-7xl mx-auto space-y-6 w-full">
    <!-- Page Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-gray-900">Dashboard</h1>
        <p class="mt-0.5 text-sm text-gray-500">Welcome back, {{ user?.full_name || 'Alice' }}. Here's what's happening.</p>
      </div>
      <button class="bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
        New Task
      </button>
    </div>

    <!-- My Tasks -->
    <section>
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-sm font-semibold text-gray-900">My Tasks</h2>
        <NuxtLink to="/my-tasks" class="text-xs font-medium text-indigo-600 hover:text-indigo-700">View all</NuxtLink>
      </div>
      <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
        <table class="w-full">
          <thead>
            <tr class="border-b border-gray-100">
              <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5">Task</th>
              <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 hidden sm:table-cell">Project</th>
              <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5">Status</th>
              <th class="text-right text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5">Due</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="task in myTasks" :key="task.id" class="hover:bg-gray-50 transition-colors cursor-pointer">
              <td class="px-4 py-2.5">
                <div class="flex items-center gap-2.5">
                  <div class="w-2 h-2 rounded-full shrink-0" :class="priorityConfig[task.priority]"></div>
                  <span class="text-xs font-mono text-gray-400 shrink-0">{{ task.key }}</span>
                  <span class="text-sm text-gray-900 truncate">{{ task.title }}</span>
                </div>
              </td>
              <td class="px-4 py-2.5 hidden sm:table-cell">
                <span class="text-xs text-gray-500">{{ task.project }}</span>
              </td>
              <td class="px-4 py-2.5">
                <span class="text-xs text-gray-500 inline-flex items-center gap-1.5">
                  <span class="w-1.5 h-1.5 rounded-full shrink-0" :class="statusConfig[task.status].dot"></span>
                  {{ statusConfig[task.status].label }}
                </span>
              </td>
              <td class="px-4 py-2.5 text-right">
                <span v-if="task.dueDate" class="text-xs" :class="isOverdue(task.dueDate) ? 'text-red-500 font-medium' : 'text-gray-400'">
                  {{ new Date(task.dueDate).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) }}
                </span>
                <span v-else class="text-xs text-gray-300">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Recent Activity -->
      <section class="lg:col-span-2">
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-sm font-semibold text-gray-900">Recent Activity</h2>
        </div>
        <div class="bg-white rounded-xl border border-gray-200">
          <div v-for="(act, i) in activities" :key="act.id" class="flex items-start gap-3 px-4 py-3" :class="i < activities.length - 1 ? 'border-b border-gray-50' : ''">
            <div class="w-7 h-7 rounded-full shrink-0 flex items-center justify-center mt-0.5" :class="act.avatar">
              <span class="text-[9px] text-white font-bold">{{ act.user.charAt(0) }}</span>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm text-gray-700">
                <span class="font-medium">{{ act.user }}</span>
                {{ act.action }}
                <span class="font-medium">{{ act.target }}</span>
                <span class="text-gray-400"> — {{ act.detail }}</span>
              </p>
              <p class="text-xs text-gray-400 mt-0.5">{{ act.time }} in {{ act.project }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- Projects -->
      <section>
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-sm font-semibold text-gray-900">Projects</h2>
          <NuxtLink to="/projects" class="text-xs font-medium text-indigo-600 hover:text-indigo-700">View all</NuxtLink>
        </div>
        <div class="space-y-3">
          <NuxtLink
            v-for="(project, i) in projects"
            :key="project.id"
            :to="`/projects/${project.slug}`"
            class="block bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md transition-shadow"
          >
            <div class="flex items-center gap-2.5">
              <div class="w-2.5 h-2.5 rounded-sm" :class="getProjectColor(i)"></div>
              <h3 class="text-sm font-semibold text-gray-900">{{ project.name }}</h3>
            </div>
            <p class="mt-1.5 text-xs text-gray-500 line-clamp-2">{{ project.description }}</p>

            <div class="mt-3 h-1 rounded-full bg-gray-100 overflow-hidden">
              <div class="h-full bg-indigo-500 rounded-full" :style="{ width: (project.taskCounts.done / Object.values(project.taskCounts).reduce((a,b)=>a+b,0) * 100) + '%' }"></div>
            </div>

            <div class="mt-2.5 flex items-center justify-between">
              <div class="flex -space-x-1.5">
                <div v-for="(m, j) in project.members.slice(0, 3)" :key="j" class="w-5 h-5 rounded-full flex items-center justify-center border-2 border-white" :class="m.color">
                  <span class="text-[7px] text-white font-bold">{{ m.name.charAt(0) }}</span>
                </div>
              </div>
              <span class="text-[11px] text-gray-400">{{ Object.values(project.taskCounts).reduce((a,b)=>a+b,0) }} tasks</span>
            </div>
          </NuxtLink>
        </div>
      </section>
    </div>
  </div>
</template>
