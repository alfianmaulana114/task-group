<script setup>
const backlog = ref([
  { id: 1, key: 'TG-12', title: 'Setup CI/CD pipeline', tag: 'DevOps', tagColor: 'bg-gray-100 text-gray-500', priority: 'bg-blue-400', avatar: null },
  { id: 2, key: 'TG-13', title: 'Write API documentation', tag: 'Docs', tagColor: 'bg-gray-100 text-gray-500', priority: 'bg-gray-300', avatar: null },
  { id: 3, key: 'TG-14', title: 'Database indexing review', tag: 'Backend', tagColor: 'bg-gray-100 text-gray-500', priority: 'bg-gray-300', avatar: { letter: 'A', color: 'bg-violet-500' } },
])

const todo = ref([
  { id: 4, key: 'TG-08', title: 'Implement user authentication', tag: 'Auth', tagColor: 'bg-blue-50 text-blue-600', priority: 'bg-orange-400', avatar: { letter: 'R', color: 'bg-sky-500' } },
  { id: 5, key: 'TG-09', title: 'Organization CRUD endpoints', tag: 'Backend', tagColor: 'bg-purple-50 text-purple-600', priority: 'bg-orange-400', avatar: null },
])

const inProgress = ref([
  { id: 6, key: 'TG-05', title: 'Build Kanban board UI', tag: 'Frontend', tagColor: 'bg-indigo-50 text-indigo-600', priority: 'bg-red-400', avatar: { letter: 'M', color: 'bg-emerald-500' }, highlight: true },
  { id: 7, key: 'TG-06', title: 'Task comments system', tags: ['Backend', 'Frontend'], tagColors: ['bg-purple-50 text-purple-600', 'bg-indigo-50 text-indigo-600'], priority: 'bg-orange-400', avatar: null },
])

const inReview = ref([
  { id: 8, key: 'TG-03', title: 'Database schema migrations', tag: 'Database', tagColor: 'bg-emerald-50 text-emerald-600', priority: 'bg-gray-300', avatar: { letter: 'S', color: 'bg-rose-500' } },
])

const done = ref([
  { id: 9, key: 'TG-01', title: 'Project setup & scaffolding', tag: 'Setup', tagColor: 'bg-gray-100 text-gray-500', priority: 'bg-gray-300', avatar: null, completed: true },
  { id: 10, key: 'TG-02', title: 'Docker Compose configuration', tag: 'DevOps', tagColor: 'bg-gray-100 text-gray-500', priority: 'bg-gray-300', avatar: null, completed: true },
])

const draggedItem = ref(null)
const dragSourceColumn = ref(null)
const dragOverColumn = ref(null)

function onDragStart(item, column) {
  draggedItem.value = item
  dragSourceColumn.value = column
}

function onDragEnd() {
  draggedItem.value = null
  dragSourceColumn.value = null
  dragOverColumn.value = null
}

function onDragOver(e, column) {
  e.preventDefault()
  if (dragSourceColumn.value !== column) {
    dragOverColumn.value = column
  }
}

function onDragLeave() {
  dragOverColumn.value = null
}

function onDrop(e, targetColumn) {
  e.preventDefault()
  dragOverColumn.value = null

  if (!draggedItem.value || !dragSourceColumn.value || dragSourceColumn.value === targetColumn) return

  const sourceList = getList(dragSourceColumn.value)
  const targetList = getList(targetColumn)
  const item = draggedItem.value

  sourceList.value = sourceList.value.filter(i => i.id !== item.id)
  targetList.value.push(item)

  draggedItem.value = null
  dragSourceColumn.value = null
}

function getList(column) {
  const map = { backlog, todo, inProgress, inReview, done }
  return map[column]
}

function getColumnCount(column) {
  return getList(column).value.length
}
</script>

<template>
  <div class="min-h-screen bg-white">
    <!-- Navbar -->
    <nav class="border-b border-gray-100">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16 items-center">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center">
              <span class="text-white font-bold text-sm">TG</span>
            </div>
            <span class="text-xl font-bold text-gray-900">Task Group</span>
          </div>
          <div class="flex items-center gap-3">
            <NuxtLink to="/login" class="text-gray-600 hover:text-gray-900 text-sm font-medium">Login</NuxtLink>
            <NuxtLink to="/register" class="bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors">Get Started</NuxtLink>
          </div>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-20 pb-24">
      <div class="text-center max-w-3xl mx-auto">
        <h1 class="text-5xl font-bold text-gray-900 tracking-tight">
          Manage projects together,
          <span class="text-indigo-600">in real time</span>
        </h1>
        <p class="mt-6 text-lg text-gray-600 leading-relaxed">
          Task Group is a collaborative project management platform for small teams.
          Plan work, assign tasks, discuss in context, and see updates instantly —
          all in one place.
        </p>
        <div class="mt-10 flex justify-center gap-4">
          <NuxtLink to="/register" class="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold px-8 py-3 rounded-lg text-base transition-colors">Start for Free</NuxtLink>
          <NuxtLink to="/login" class="border border-gray-300 hover:border-gray-400 text-gray-700 font-semibold px-8 py-3 rounded-lg text-base transition-colors">Sign In</NuxtLink>
        </div>
      </div>

      <!-- Kanban Board Preview -->
      <div class="mt-16 max-w-5xl mx-auto">
        <div class="bg-white rounded-2xl border border-gray-200 shadow-2xl overflow-hidden">
          <!-- Window Header -->
          <div class="bg-gray-50 border-b border-gray-200 px-4 py-3 flex items-center gap-3">
            <div class="flex gap-1.5">
              <div class="w-3 h-3 rounded-full bg-red-400"></div>
              <div class="w-3 h-3 rounded-full bg-amber-400"></div>
              <div class="w-3 h-3 rounded-full bg-green-400"></div>
            </div>
            <div class="flex-1 text-center">
              <span class="text-xs font-medium text-gray-500">Task Group — Website Redesign</span>
            </div>
          </div>

          <!-- Board -->
          <div class="p-4 flex gap-4 overflow-x-auto">
            <!-- Backlog -->
            <div
              class="flex-1 min-w-[180px] rounded-lg transition-colors"
              :class="{ 'bg-indigo-50': dragOverColumn === 'backlog' }"
              @dragover="onDragOver($event, 'backlog')"
              @dragleave="onDragLeave"
              @drop="onDrop($event, 'backlog')"
            >
              <div class="flex items-center gap-2 mb-3">
                <div class="w-2 h-2 rounded-full bg-gray-400"></div>
                <span class="text-xs font-semibold text-gray-600 uppercase tracking-wide">Backlog</span>
                <span class="text-xs text-gray-400 ml-auto">{{ getColumnCount('backlog') }}</span>
              </div>
              <div class="space-y-2 min-h-[60px]">
                <div
                  v-for="item in backlog"
                  :key="item.id"
                  draggable="true"
                  class="bg-white rounded-lg border border-gray-200 p-3 cursor-grab active:cursor-grabbing hover:shadow-md transition-all select-none"
                  :class="{ 'opacity-40': draggedItem?.id === item.id }"
                  @dragstart="onDragStart(item, 'backlog')"
                  @dragend="onDragEnd"
                >
                  <div class="flex items-start justify-between gap-2">
                    <span class="text-xs font-mono text-gray-400">{{ item.key }}</span>
                    <span class="w-1.5 h-1.5 rounded-full mt-1 shrink-0" :class="item.priority"></span>
                  </div>
                  <p class="text-sm font-medium text-gray-800 mt-1">{{ item.title }}</p>
                  <div class="flex items-center justify-between mt-2">
                    <span class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="item.tagColor">{{ item.tag }}</span>
                    <div v-if="item.avatar" class="w-5 h-5 rounded-full flex items-center justify-center" :class="item.avatar.color">
                      <span class="text-[9px] text-white font-bold">{{ item.avatar.letter }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Todo -->
            <div
              class="flex-1 min-w-[180px] rounded-lg transition-colors"
              :class="{ 'bg-indigo-50': dragOverColumn === 'todo' }"
              @dragover="onDragOver($event, 'todo')"
              @dragleave="onDragLeave"
              @drop="onDrop($event, 'todo')"
            >
              <div class="flex items-center gap-2 mb-3">
                <div class="w-2 h-2 rounded-full bg-blue-400"></div>
                <span class="text-xs font-semibold text-gray-600 uppercase tracking-wide">Todo</span>
                <span class="text-xs text-gray-400 ml-auto">{{ getColumnCount('todo') }}</span>
              </div>
              <div class="space-y-2 min-h-[60px]">
                <div
                  v-for="item in todo"
                  :key="item.id"
                  draggable="true"
                  class="bg-white rounded-lg border border-gray-200 p-3 cursor-grab active:cursor-grabbing hover:shadow-md transition-all select-none"
                  :class="{ 'opacity-40': draggedItem?.id === item.id }"
                  @dragstart="onDragStart(item, 'todo')"
                  @dragend="onDragEnd"
                >
                  <div class="flex items-start justify-between gap-2">
                    <span class="text-xs font-mono text-gray-400">{{ item.key }}</span>
                    <span class="w-1.5 h-1.5 rounded-full mt-1 shrink-0" :class="item.priority"></span>
                  </div>
                  <p class="text-sm font-medium text-gray-800 mt-1">{{ item.title }}</p>
                  <div class="flex items-center justify-between mt-2">
                    <span class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="item.tagColor">{{ item.tag }}</span>
                    <div v-if="item.avatar" class="w-5 h-5 rounded-full flex items-center justify-center" :class="item.avatar.color">
                      <span class="text-[9px] text-white font-bold">{{ item.avatar.letter }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- In Progress -->
            <div
              class="flex-1 min-w-[180px] rounded-lg transition-colors"
              :class="{ 'bg-indigo-50': dragOverColumn === 'inProgress' }"
              @dragover="onDragOver($event, 'inProgress')"
              @dragleave="onDragLeave"
              @drop="onDrop($event, 'inProgress')"
            >
              <div class="flex items-center gap-2 mb-3">
                <div class="w-2 h-2 rounded-full bg-amber-400"></div>
                <span class="text-xs font-semibold text-gray-600 uppercase tracking-wide">In Progress</span>
                <span class="text-xs text-gray-400 ml-auto">{{ getColumnCount('inProgress') }}</span>
              </div>
              <div class="space-y-2 min-h-[60px]">
                <div
                  v-for="item in inProgress"
                  :key="item.id"
                  draggable="true"
                  class="bg-white rounded-lg border p-3 cursor-grab active:cursor-grabbing hover:shadow-md transition-all select-none"
                  :class="[
                    item.highlight ? 'border-2 border-amber-200 shadow-sm' : 'border-gray-200',
                    draggedItem?.id === item.id ? 'opacity-40' : ''
                  ]"
                  @dragstart="onDragStart(item, 'inProgress')"
                  @dragend="onDragEnd"
                >
                  <div class="flex items-start justify-between gap-2">
                    <span class="text-xs font-mono text-gray-400">{{ item.key }}</span>
                    <span class="w-1.5 h-1.5 rounded-full mt-1 shrink-0" :class="item.priority"></span>
                  </div>
                  <p class="text-sm font-medium text-gray-800 mt-1">{{ item.title }}</p>
                  <div class="flex items-center justify-between mt-2">
                    <div class="flex items-center gap-1">
                      <template v-if="item.tags">
                        <span v-for="(tag, i) in item.tags" :key="i" class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="item.tagColors[i]">{{ tag }}</span>
                      </template>
                      <span v-else class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="item.tagColor">{{ item.tag }}</span>
                    </div>
                    <div v-if="item.avatar" class="w-5 h-5 rounded-full flex items-center justify-center" :class="item.avatar.color">
                      <span class="text-[9px] text-white font-bold">{{ item.avatar.letter }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- In Review -->
            <div
              class="flex-1 min-w-[180px] rounded-lg transition-colors"
              :class="{ 'bg-indigo-50': dragOverColumn === 'inReview' }"
              @dragover="onDragOver($event, 'inReview')"
              @dragleave="onDragLeave"
              @drop="onDrop($event, 'inReview')"
            >
              <div class="flex items-center gap-2 mb-3">
                <div class="w-2 h-2 rounded-full bg-purple-400"></div>
                <span class="text-xs font-semibold text-gray-600 uppercase tracking-wide">In Review</span>
                <span class="text-xs text-gray-400 ml-auto">{{ getColumnCount('inReview') }}</span>
              </div>
              <div class="space-y-2 min-h-[60px]">
                <div
                  v-for="item in inReview"
                  :key="item.id"
                  draggable="true"
                  class="bg-white rounded-lg border border-gray-200 p-3 cursor-grab active:cursor-grabbing hover:shadow-md transition-all select-none"
                  :class="{ 'opacity-40': draggedItem?.id === item.id }"
                  @dragstart="onDragStart(item, 'inReview')"
                  @dragend="onDragEnd"
                >
                  <div class="flex items-start justify-between gap-2">
                    <span class="text-xs font-mono text-gray-400">{{ item.key }}</span>
                    <span class="w-1.5 h-1.5 rounded-full mt-1 shrink-0" :class="item.priority"></span>
                  </div>
                  <p class="text-sm font-medium text-gray-800 mt-1">{{ item.title }}</p>
                  <div class="flex items-center justify-between mt-2">
                    <span class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="item.tagColor">{{ item.tag }}</span>
                    <div v-if="item.avatar" class="w-5 h-5 rounded-full flex items-center justify-center" :class="item.avatar.color">
                      <span class="text-[9px] text-white font-bold">{{ item.avatar.letter }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Done -->
            <div
              class="flex-1 min-w-[180px] rounded-lg transition-colors"
              :class="{ 'bg-indigo-50': dragOverColumn === 'done' }"
              @dragover="onDragOver($event, 'done')"
              @dragleave="onDragLeave"
              @drop="onDrop($event, 'done')"
            >
              <div class="flex items-center gap-2 mb-3">
                <div class="w-2 h-2 rounded-full bg-green-500"></div>
                <span class="text-xs font-semibold text-gray-600 uppercase tracking-wide">Done</span>
                <span class="text-xs text-gray-400 ml-auto">{{ getColumnCount('done') }}</span>
              </div>
              <div class="space-y-2 min-h-[60px]">
                <div
                  v-for="item in done"
                  :key="item.id"
                  draggable="true"
                  class="bg-white rounded-lg border border-gray-200 p-3 opacity-70 cursor-grab active:cursor-grabbing hover:shadow-md transition-all select-none"
                  :class="{ 'opacity-30': draggedItem?.id === item.id }"
                  @dragstart="onDragStart(item, 'done')"
                  @dragend="onDragEnd"
                >
                  <div class="flex items-start justify-between gap-2">
                    <span class="text-xs font-mono text-gray-400">{{ item.key }}</span>
                    <span class="w-1.5 h-1.5 rounded-full mt-1 shrink-0" :class="item.priority"></span>
                  </div>
                  <p class="text-sm font-medium text-gray-800 mt-1 line-through">{{ item.title }}</p>
                  <div class="flex items-center justify-between mt-2">
                    <span class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="item.tagColor">{{ item.tag }}</span>
                    <div v-if="item.avatar" class="w-5 h-5 rounded-full flex items-center justify-center" :class="item.avatar.color">
                      <span class="text-[9px] text-white font-bold">{{ item.avatar.letter }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Hint -->
          <div class="px-4 pb-4 pt-1 text-center">
            <span class="text-[11px] text-gray-400">Drag cards between columns to change status</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="bg-gray-50 py-24">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="text-center max-w-2xl mx-auto mb-16">
          <h2 class="text-3xl font-bold text-gray-900">Everything your team needs</h2>
          <p class="mt-4 text-gray-600">From planning to delivery, Task Group keeps your team aligned with powerful features built for collaboration.</p>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          <div class="bg-white rounded-xl p-6 border border-gray-200 hover:shadow-lg transition-shadow">
            <div class="w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900">Organizations & Teams</h3>
            <p class="mt-2 text-sm text-gray-600">Create workspaces, invite members, and manage roles. Each organization keeps its data fully isolated.</p>
          </div>

          <div class="bg-white rounded-xl p-6 border border-gray-200 hover:shadow-lg transition-shadow">
            <div class="w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" /></svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900">Project Management</h3>
            <p class="mt-2 text-sm text-gray-600">Organize work into projects. Add members, set descriptions, and track progress across your team.</p>
          </div>

          <div class="bg-white rounded-xl p-6 border border-gray-200 hover:shadow-lg transition-shadow">
            <div class="w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" /></svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900">Kanban Board</h3>
            <p class="mt-2 text-sm text-gray-600">Visualize your workflow with a drag-and-drop board. Track tasks from Backlog to Done with five built-in statuses.</p>
          </div>

          <div class="bg-white rounded-xl p-6 border border-gray-200 hover:shadow-lg transition-shadow">
            <div class="w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900">Real-time Updates</h3>
            <p class="mt-2 text-sm text-gray-600">See changes instantly. When someone moves a task or posts a comment, everyone sees it without refreshing.</p>
          </div>

          <div class="bg-white rounded-xl p-6 border border-gray-200 hover:shadow-lg transition-shadow">
            <div class="w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" /></svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900">Task Comments</h3>
            <p class="mt-2 text-sm text-gray-600">Discuss work right where it happens. Comment on tasks to keep context and decisions attached to the work.</p>
          </div>

          <div class="bg-white rounded-xl p-6 border border-gray-200 hover:shadow-lg transition-shadow">
            <div class="w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8h2a2 2 0 012 2v6a2 2 0 01-2 2h-2v4l-4-4H9a1.994 1.994 0 01-1.414-.586m0 0L11 14h4a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2v4l.586-.586z" /></svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900">Group Chat</h3>
            <p class="mt-2 text-sm text-gray-600">Every project gets its own chat channel. Talk with your team in context without switching tools.</p>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="py-24">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="bg-indigo-600 rounded-2xl p-12 text-center">
          <h2 class="text-3xl font-bold text-white">Ready to get started?</h2>
          <p class="mt-4 text-indigo-100 max-w-xl mx-auto">Set up your team workspace in minutes. No credit card required.</p>
          <div class="mt-8">
            <NuxtLink to="/register" class="bg-white hover:bg-gray-100 text-indigo-600 font-semibold px-8 py-3 rounded-lg text-base transition-colors inline-block">Create Free Account</NuxtLink>
          </div>
        </div>
      </div>
    </section>

    <!-- Footer -->
    <footer class="border-t border-gray-100 py-8">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex flex-col md:flex-row justify-between items-center gap-4">
          <div class="flex items-center gap-2">
            <div class="w-6 h-6 bg-indigo-600 rounded flex items-center justify-center">
              <span class="text-white font-bold text-xs">TG</span>
            </div>
            <span class="text-sm text-gray-500">Task Group</span>
          </div>
          <p class="text-sm text-gray-400">&copy; {{ new Date().getFullYear() }} Task Group. All rights reserved.</p>
        </div>
      </div>
    </footer>
  </div>
</template>
