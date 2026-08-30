<script setup>
definePageMeta({ layout: 'dashboard' })

const members = ref([
  { id: '1', name: 'Alice Johnson', email: 'alice@acme.com', role: 'owner', avatar: 'bg-violet-500', status: 'active', joinedAt: '2026-05-01', tasksCount: 12 },
  { id: '2', name: 'Bob Smith', email: 'bob@acme.com', role: 'admin', avatar: 'bg-sky-500', status: 'active', joinedAt: '2026-05-15', tasksCount: 8 },
  { id: '3', name: 'Charlie Brown', email: 'charlie@acme.com', role: 'member', avatar: 'bg-emerald-500', status: 'active', joinedAt: '2026-06-01', tasksCount: 15 },
  { id: '4', name: 'Diana Prince', email: 'diana@acme.com', role: 'member', avatar: 'bg-rose-500', status: 'active', joinedAt: '2026-06-20', tasksCount: 10 },
  { id: '5', name: 'Eve Williams', email: 'eve@acme.com', role: 'member', avatar: 'bg-amber-500', status: 'pending', joinedAt: null, tasksCount: 0 },
])

const roleConfig = {
  owner: { label: 'Owner', color: 'bg-indigo-50 text-indigo-600' },
  admin: { label: 'Admin', color: 'bg-amber-50 text-amber-700' },
  member: { label: 'Member', color: 'bg-gray-100 text-gray-600' },
}

const searchQuery = ref('')

const filteredMembers = computed(() => {
  if (!searchQuery.value) return members.value
  const q = searchQuery.value.toLowerCase()
  return members.value.filter(m => m.name.toLowerCase().includes(q) || m.email.toLowerCase().includes(q))
})
</script>

<template>
  <div class="p-4 lg:p-6 max-w-5xl mx-auto w-full">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-bold text-gray-900">Members</h1>
        <p class="mt-0.5 text-sm text-gray-500">Manage your organization members and roles.</p>
      </div>
      <button class="bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
        Invite Member
      </button>
    </div>

    <!-- Search -->
    <div class="mb-5 max-w-sm relative">
      <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
      <input v-model="searchQuery" type="text" placeholder="Search members..." class="w-full pl-10 pr-4 py-2 bg-white border border-gray-200 rounded-lg text-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
    </div>

    <!-- Member List -->
    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
      <table class="w-full">
        <thead>
          <tr class="border-b border-gray-100">
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5">Member</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 hidden sm:table-cell">Role</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 hidden md:table-cell">Status</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 hidden lg:table-cell">Tasks</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5 hidden lg:table-cell">Joined</th>
            <th class="text-right text-[11px] font-semibold text-gray-400 uppercase tracking-wider px-4 py-2.5">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-50">
          <tr v-for="member in filteredMembers" :key="member.id" class="hover:bg-gray-50 transition-colors">
            <td class="px-4 py-3">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-full flex items-center justify-center shrink-0" :class="member.avatar">
                  <span class="text-xs text-white font-bold">{{ member.name.charAt(0) }}</span>
                </div>
                <div class="min-w-0">
                  <div class="text-sm font-medium text-gray-900 truncate">{{ member.name }}</div>
                  <div class="text-xs text-gray-400 truncate">{{ member.email }}</div>
                </div>
              </div>
            </td>
            <td class="px-4 py-3 hidden sm:table-cell">
              <span class="text-[11px] px-2 py-0.5 rounded-full font-medium" :class="roleConfig[member.role].color">{{ roleConfig[member.role].label }}</span>
            </td>
            <td class="px-4 py-3 hidden md:table-cell">
              <div class="flex items-center gap-1.5">
                <div class="w-1.5 h-1.5 rounded-full" :class="member.status === 'active' ? 'bg-emerald-500' : 'bg-amber-400'"></div>
                <span class="text-xs text-gray-500 capitalize">{{ member.status }}</span>
              </div>
            </td>
            <td class="px-4 py-3 hidden lg:table-cell">
              <span class="text-sm text-gray-600">{{ member.tasksCount }}</span>
            </td>
            <td class="px-4 py-3 hidden lg:table-cell">
              <span v-if="member.joinedAt" class="text-xs text-gray-400">{{ new Date(member.joinedAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) }}</span>
              <span v-else class="text-xs text-amber-500 font-medium">Pending</span>
            </td>
            <td class="px-4 py-3 text-right">
              <button class="text-gray-400 hover:text-gray-600 p-1">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" /></svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
