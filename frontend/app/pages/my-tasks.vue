<script setup>
definePageMeta({ layout: 'dashboard' })

const tasks = ref([
  { id: '1', key: 'TG-45', title: 'Implement dark mode toggle', project: 'Website Redesign', projectColor: 'bg-indigo-500', status: 'in_progress', priority: 'high', assignee: 'Alice', assigneeColor: 'bg-violet-500', dueDate: '2026-09-05', labels: ['Frontend'] },
  { id: '2', key: 'TG-38', title: 'Design onboarding flow wireframes', project: 'Mobile App', projectColor: 'bg-emerald-500', status: 'todo', priority: 'medium', assignee: 'Alice', assigneeColor: 'bg-violet-500', dueDate: '2026-09-08', labels: ['Design'] },
  { id: '3', key: 'TG-52', title: 'Fix responsive layout on tablet', project: 'Website Redesign', projectColor: 'bg-indigo-500', status: 'review', priority: 'urgent', assignee: 'Alice', assigneeColor: 'bg-violet-500', dueDate: '2026-09-02', labels: ['Frontend', 'Bug'] },
  { id: '4', key: 'TG-29', title: 'Write unit tests for auth module', project: 'API Backend', projectColor: 'bg-amber-500', status: 'todo', priority: 'medium', assignee: 'Alice', assigneeColor: 'bg-violet-500', dueDate: '2026-09-10', labels: ['Backend', 'Testing'] },
  { id: '5', key: 'TG-61', title: 'Update API documentation for v2', project: 'API Backend', projectColor: 'bg-amber-500', status: 'backlog', priority: 'low', assignee: 'Alice', assigneeColor: 'bg-violet-500', dueDate: null, labels: ['Docs'] },
  { id: '6', key: 'TG-33', title: 'Setup CI/CD pipeline', project: 'API Backend', projectColor: 'bg-amber-500', status: 'done', priority: 'high', assignee: 'Bob', assigneeColor: 'bg-sky-500', dueDate: '2026-08-28', labels: ['DevOps'] },
  { id: '7', key: 'TG-41', title: 'Create user avatar component', project: 'Design System', projectColor: 'bg-rose-500', status: 'in_progress', priority: 'medium', assignee: 'Diana', assigneeColor: 'bg-rose-500', dueDate: '2026-09-06', labels: ['Frontend', 'Design'] },
  { id: '8', key: 'TG-55', title: 'Database schema migration script', project: 'API Backend', projectColor: 'bg-amber-500', status: 'done', priority: 'high', assignee: 'Charlie', assigneeColor: 'bg-emerald-500', dueDate: '2026-08-25', labels: ['Backend', 'Database'] },
])

const statusConfig = {
  backlog: { label: 'Backlog', dot: 'bg-gray-400' },
  todo: { label: 'Todo', dot: 'bg-blue-500' },
  in_progress: { label: 'In Progress', dot: 'bg-amber-500' },
  review: { label: 'In Review', dot: 'bg-purple-500' },
  done: { label: 'Done', dot: 'bg-emerald-500' },
}

const priorityConfig = {
  urgent: { label: 'Urgent', dot: 'bg-red-500' },
  high: { label: 'High', dot: 'bg-orange-400' },
  medium: { label: 'Medium', dot: 'bg-gray-400' },
  low: { label: 'Low', dot: 'bg-gray-300' },
}

const filterStatus = ref('all')
const filterPriority = ref('all')

const filteredTasks = computed(() => {
  return tasks.value.filter(t => {
    if (filterStatus.value !== 'all' && t.status !== filterStatus.value) return false
    if (filterPriority.value !== 'all' && t.priority !== filterPriority.value) return false
    return true
  })
})

function isOverdue(dateStr) {
  if (!dateStr) return false
  return new Date(dateStr) < new Date()
}
</script>

<template>
  <div class="p-4 lg:p-6 max-w-7xl mx-auto w-full">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-bold text-gray-900">My Tasks</h1>
        <p class="mt-0.5 text-sm text-gray-500">{{ filteredTasks.length }} tasks assigned to you.</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex items-center gap-3 mb-5 flex-wrap">
      <select v-model="filterStatus" class="text-sm border border-gray-200 rounded-lg px-3 py-1.5 bg-white text-gray-700 focus:outline-none focus:ring-2 focus:ring-indigo-500">
        <option value="all">All Status</option>
        <option value="backlog">Backlog</option>
        <option value="todo">Todo</option>
        <option value="in_progress">In Progress</option>
        <option value="review">In Review</option>
        <option value="done">Done</option>
      </select>
      <select v-model="filterPriority" class="text-sm border border-gray-200 rounded-lg px-3 py-1.5 bg-white text-gray-700 focus:outline-none focus:ring-2 focus:ring-indigo-500">
        <option value="all">All Priority</option>
        <option value="urgent">Urgent</option>
        <option value="high">High</option>
        <option value="medium">Medium</option>
        <option value="low">Low</option>
      </select>
    </div>

    <!-- Task List -->
    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
      <table class="w-full">
        <thead>
          <tr class="border-b border-gray-100">
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 w-8"></th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5">Task</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 hidden sm:table-cell">Project</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 hidden md:table-cell">Labels</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5">Status</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 hidden lg:table-cell">Assignee</th>
            <th class="text-right text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5">Due</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-50">
          <tr v-for="task in filteredTasks" :key="task.id" class="hover:bg-gray-50 transition-colors cursor-pointer">
            <td class="px-4 py-3">
              <div class="w-2 h-2 rounded-full" :class="priorityConfig[task.priority].dot"></div>
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <span class="text-xs font-mono text-gray-400 shrink-0">{{ task.key }}</span>
                <span class="text-sm text-gray-900 truncate">{{ task.title }}</span>
              </div>
            </td>
            <td class="px-4 py-3 hidden sm:table-cell">
              <div class="flex items-center gap-1.5">
                <div class="w-2 h-2 rounded-sm" :class="task.projectColor"></div>
                <span class="text-xs text-gray-500">{{ task.project }}</span>
              </div>
            </td>
            <td class="px-4 py-3 hidden md:table-cell">
              <div class="flex gap-1">
                <span v-for="label in task.labels" :key="label" class="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 text-gray-500 font-medium">{{ label }}</span>
              </div>
            </td>
            <td class="px-4 py-3">
              <span class="text-xs text-gray-500 inline-flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full shrink-0" :class="statusConfig[task.status].dot"></span>
                {{ statusConfig[task.status].label }}
              </span>
            </td>
            <td class="px-4 py-3 hidden lg:table-cell">
              <div class="flex items-center gap-1.5">
                <div class="w-5 h-5 rounded-full flex items-center justify-center" :class="task.assigneeColor">
                  <span class="text-[7px] text-white font-bold">{{ task.assignee.charAt(0) }}</span>
                </div>
                <span class="text-xs text-gray-500">{{ task.assignee }}</span>
              </div>
            </td>
            <td class="px-4 py-3 text-right">
              <span v-if="task.dueDate" class="text-xs" :class="isOverdue(task.dueDate) && task.status !== 'done' ? 'text-red-500 font-medium' : 'text-gray-400'">
                {{ new Date(task.dueDate).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) }}
              </span>
              <span v-else class="text-xs text-gray-300">—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
