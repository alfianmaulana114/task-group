<script setup>
definePageMeta({ layout: 'dashboard' })

const { get, post, patch, del, setOrgId } = useApi()
const { user, orgId, orgRole, token } = useAuth()
const toast = useToast()

const loading = ref(true)
const members = ref([])
const searchQuery = ref('')
const showInvite = ref(false)
const inviting = ref(false)
const inviteEmail = ref('')
const inviteRole = ref('member')

const showRoleModal = ref(false)
const roleMember = ref(null)
const newRole = ref('')
const updatingRole = ref(false)

const showRemoveModal = ref(false)
const removeMember = ref(null)
const removing = ref(false)

function syncOrgId() {
  const oid = orgId.value || (import.meta.client ? localStorage.getItem('auth:org_id') : null)
  if (oid) setOrgId(oid)
  return oid
}

onMounted(() => {
  syncOrgId()
  watch(orgId, syncOrgId)
  const hasToken = token.value || (import.meta.client ? localStorage.getItem('auth:token') : null)
  if (!hasToken) {
    loading.value = false
    return
  }
  fetchMembers()
})

async function fetchMembers() {
  loading.value = true
  try {
    const oid = syncOrgId()
    if (!oid) {
      toast.error('Organisasi belum dipilih')
      return
    }
    members.value = await get(`/api/v1/organizations/${oid}/members`)
  } catch (e) {
    toast.error(e?.message || 'Gagal memuat anggota')
  } finally {
    loading.value = false
  }
}

async function inviteMember() {
  if (!inviteEmail.value.trim()) return
  const oid = syncOrgId()
  if (!oid) {
    toast.error('Organisasi belum dipilih')
    return
  }
  inviting.value = true
  try {
    await post(`/api/v1/organizations/${oid}/members`, {
      email: inviteEmail.value,
      role: inviteRole.value,
    })
    showInvite.value = false
    inviteEmail.value = ''
    inviteRole.value = 'member'
    toast.success('Undangan berhasil dikirim!')
    await fetchMembers()
  } catch (e) {
    toast.error(e?.message || 'Gagal mengundang anggota')
  } finally {
    inviting.value = false
  }
}

function openRoleModal(member) {
  roleMember.value = member
  newRole.value = member.role
  showRoleModal.value = true
}

async function updateRole() {
  if (!roleMember.value || !newRole.value) return
  const oid = syncOrgId()
  if (!oid) return
  updatingRole.value = true
  try {
    await patch(`/api/v1/organizations/${oid}/members/${roleMember.value.id}`, { role: newRole.value })
    toast.success('Peran berhasil diubah')
    showRoleModal.value = false
    await fetchMembers()
  } catch (e) {
    toast.error(e?.message || 'Gagal mengubah peran')
  } finally {
    updatingRole.value = false
  }
}

function openRemoveModal(member) {
  removeMember.value = member
  showRemoveModal.value = true
}

async function confirmRemove() {
  if (!removeMember.value) return
  const oid = syncOrgId()
  if (!oid) return
  removing.value = true
  try {
    await del(`/api/v1/organizations/${oid}/members/${removeMember.value.id}`)
    toast.success('Anggota berhasil dihapus')
    showRemoveModal.value = false
    await fetchMembers()
  } catch (e) {
    toast.error(e?.message || 'Gagal menghapus anggota')
  } finally {
    removing.value = false
  }
}

const roleConfig = {
  owner: { label: 'Owner', color: 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600' },
  admin: { label: 'Admin', color: 'bg-amber-50 text-amber-700' },
  member: { label: 'Member', color: 'bg-gray-100 dark:bg-gray-700 text-gray-600' },
}

const avatarColors = ['bg-violet-500', 'bg-sky-500', 'bg-emerald-500', 'bg-rose-500', 'bg-amber-500']

const filteredMembers = computed(() => {
  if (!searchQuery.value) return members.value
  const q = searchQuery.value.toLowerCase()
  return members.value.filter(m => (m.user?.full_name || m.user?.email || '').toLowerCase().includes(q))
})

const canManage = computed(() => ['owner', 'admin'].includes(orgRole.value))
</script>

<template>
  <div class="p-4 lg:p-6 max-w-5xl mx-auto w-full">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white">Anggota</h1>
        <p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">Kelola anggota dan peran organisasi.</p>
      </div>
      <button v-if="canManage" @click="showInvite = true" class="bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-all flex items-center gap-2 hover:shadow-sm active:scale-[0.97]">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
        Undang Anggota
      </button>
    </div>

    <!-- Search -->
    <div class="mb-5 max-w-sm relative">
      <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
      <input v-model="searchQuery" type="text" placeholder="Cari anggota..." class="w-full pl-10 pr-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all" />
    </div>

    <!-- Skeleton / Member List -->
    <Transition name="fade" mode="out-in">
      <div v-if="loading" :key="'skeleton'" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
        <div v-for="i in 4" :key="i" class="flex items-center gap-3 px-4 py-3 border-b border-gray-50 dark:border-gray-700 animate-pulse">
          <div class="w-8 h-8 rounded-full bg-gray-200 dark:bg-gray-600"></div>
          <div class="flex-1"><div class="h-3 bg-gray-200 dark:bg-gray-600 rounded w-32 mb-1"></div><div class="h-2 bg-gray-100 dark:bg-gray-700 rounded w-24"></div></div>
          <div class="h-5 bg-gray-100 dark:bg-gray-700 rounded-full w-14"></div>
        </div>
      </div>

      <!-- Member List -->
      <div v-else :key="'data'" class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
      <table class="w-full">
        <thead>
          <tr class="border-b border-gray-100 dark:border-gray-700">
            <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5">Anggota</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5 hidden sm:table-cell">Peran</th>
            <th class="text-left text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5 hidden md:table-cell">Bergabung</th>
            <th v-if="canManage" class="text-right text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-4 py-2.5 w-20"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-50 dark:divide-gray-700">
          <tr v-for="(member, i) in filteredMembers" :key="member.id" class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
            <td class="px-4 py-3">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-full flex items-center justify-center shrink-0" :class="avatarColors[i % avatarColors.length]">
                  <span class="text-xs text-white font-bold">{{ (member.user?.full_name || member.user?.email || '?').charAt(0).toUpperCase() }}</span>
                </div>
                <div class="min-w-0">
                  <div class="text-sm font-medium text-gray-900 dark:text-white truncate">{{ member.user?.full_name || 'User' }}</div>
                  <div class="text-xs text-gray-400 dark:text-gray-500 truncate">{{ member.user?.email }}</div>
                </div>
              </div>
            </td>
            <td class="px-4 py-3 hidden sm:table-cell">
              <span class="text-[11px] px-2 py-0.5 rounded-full font-medium" :class="roleConfig[member.role]?.color || 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300'">{{ roleConfig[member.role]?.label || member.role }}</span>
            </td>
            <td class="px-4 py-3 hidden md:table-cell">
              <span v-if="member.joined_at" class="text-xs text-gray-400 dark:text-gray-500">{{ new Date(member.joined_at).toLocaleDateString('id-ID', { month: 'short', day: 'numeric', year: 'numeric' }) }}</span>
              <span v-else class="text-xs text-gray-400 dark:text-gray-500">-</span>
            </td>
            <td v-if="canManage" class="px-4 py-3 text-right">
              <div v-if="member.role !== 'owner' && member.user_id !== user?.id" class="flex items-center justify-end gap-1">
                <button @click="openRoleModal(member)" class="p-1.5 text-gray-400 dark:text-gray-500 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors" title="Ubah Peran">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                </button>
                <button @click="openRemoveModal(member)" class="p-1.5 text-gray-400 dark:text-gray-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors" title="Hapus Anggota">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="filteredMembers.length === 0" class="px-4 py-8 text-center">
        <p class="text-sm text-gray-400 dark:text-gray-500">Tidak ada anggota ditemukan.</p>
      </div>
    </div>
    </Transition>

    <!-- Invite Modal -->
    <Teleport to="body">
      <Transition enter-active-class="transition-all duration-200 ease-out" leave-active-class="transition-all duration-150 ease-in" enter-from-class="opacity-0 scale-95" leave-to-class="opacity-0 scale-95">
        <div v-if="showInvite" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showInvite = false"></div>
          <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-sm overflow-hidden">
            <div class="px-5 py-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
              <h3 class="text-sm font-bold text-gray-900 dark:text-white">Undang Anggota</h3>
              <button @click="showInvite = false" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <div class="px-5 py-4 space-y-3">
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Email <span class="text-red-500">*</span></label>
                <input v-model="inviteEmail" type="email" placeholder="email@contoh.com" class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all" @keydown.enter="inviteMember" />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Peran</label>
                <select v-model="inviteRole" class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all">
                  <option value="member">Member</option>
                  <option value="admin">Admin</option>
                </select>
              </div>
            </div>
            <div class="px-5 py-3 border-t border-gray-100 dark:border-gray-700 flex justify-end gap-2">
              <button @click="showInvite = false" class="px-3 py-1.5 text-xs font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg transition-colors">Batal</button>
              <button @click="inviteMember" :disabled="!inviteEmail.trim() || inviting" class="px-3 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-all disabled:opacity-50 flex items-center gap-1.5">
                <svg v-if="inviting" class="animate-spin w-3 h-3" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
                Kirim Undangan
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Change Role Modal -->
    <Teleport to="body">
      <Transition enter-active-class="transition-all duration-200 ease-out" leave-active-class="transition-all duration-150 ease-in" enter-from-class="opacity-0 scale-95" leave-to-class="opacity-0 scale-95">
        <div v-if="showRoleModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showRoleModal = false"></div>
          <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-sm overflow-hidden">
            <div class="px-5 py-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
              <h3 class="text-sm font-bold text-gray-900 dark:text-white">Ubah Peran</h3>
              <button @click="showRoleModal = false" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <div class="px-5 py-4 space-y-3">
              <p class="text-sm text-gray-600 dark:text-gray-400">Ubah peran <span class="font-medium text-gray-900 dark:text-white">{{ roleMember?.user?.full_name || roleMember?.user?.email }}</span></p>
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Peran Baru</label>
                <select v-model="newRole" class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all">
                  <option value="admin">Admin</option>
                  <option value="member">Member</option>
                </select>
              </div>
            </div>
            <div class="px-5 py-3 border-t border-gray-100 dark:border-gray-700 flex justify-end gap-2">
              <button @click="showRoleModal = false" class="px-3 py-1.5 text-xs font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg transition-colors">Batal</button>
              <button @click="updateRole" :disabled="updatingRole" class="px-3 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-all disabled:opacity-50 flex items-center gap-1.5">
                <svg v-if="updatingRole" class="animate-spin w-3 h-3" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
                Simpan
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Remove Confirmation Modal -->
    <Teleport to="body">
      <Transition enter-active-class="transition-all duration-200 ease-out" leave-active-class="transition-all duration-150 ease-in" enter-from-class="opacity-0 scale-95" leave-to-class="opacity-0 scale-95">
        <div v-if="showRemoveModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showRemoveModal = false"></div>
          <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-sm overflow-hidden">
            <div class="px-5 py-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
              <h3 class="text-sm font-bold text-gray-900 dark:text-white">Hapus Anggota</h3>
              <button @click="showRemoveModal = false" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <div class="px-5 py-4">
              <p class="text-sm text-gray-600 dark:text-gray-400">Yakin ingin menghapus <span class="font-medium text-gray-900 dark:text-white">{{ removeMember?.user?.full_name || removeMember?.user?.email }}</span> dari organisasi?</p>
            </div>
            <div class="px-5 py-3 border-t border-gray-100 dark:border-gray-700 flex justify-end gap-2">
              <button @click="showRemoveModal = false" class="px-3 py-1.5 text-xs font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg transition-colors">Batal</button>
              <button @click="confirmRemove" :disabled="removing" class="px-3 py-1.5 text-xs font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-all disabled:opacity-50 flex items-center gap-1.5">
                <svg v-if="removing" class="animate-spin w-3 h-3" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
                Hapus
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
