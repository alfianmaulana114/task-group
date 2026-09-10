<script setup>
const { toasts, remove } = useToast()

const typeStyles = {
  success: { bg: 'bg-emerald-50 border-emerald-200', text: 'text-emerald-700', icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z' },
  error: { bg: 'bg-red-50 border-red-200', text: 'text-red-700', icon: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z' },
  info: { bg: 'bg-blue-50 border-blue-200', text: 'text-blue-700', icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
  warning: { bg: 'bg-amber-50 border-amber-200', text: 'text-amber-700', icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z' },
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none">
      <TransitionGroup
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="translate-x-8 opacity-0 scale-95"
        enter-to-class="translate-x-0 opacity-100 scale-100"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="translate-x-0 opacity-100 scale-100"
        leave-to-class="translate-x-8 opacity-0 scale-95"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="pointer-events-auto flex items-center gap-3 px-4 py-3 rounded-lg border shadow-lg max-w-sm backdrop-blur-sm"
          :class="typeStyles[toast.type].bg"
        >
          <svg class="w-5 h-5 shrink-0" :class="typeStyles[toast.type].text" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="typeStyles[toast.type].icon" />
          </svg>
          <p class="text-sm font-medium flex-1" :class="typeStyles[toast.type].text">{{ toast.message }}</p>
          <button @click="remove(toast.id)" class="shrink-0 p-0.5 rounded hover:bg-black/5 transition-colors" :class="typeStyles[toast.type].text">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
