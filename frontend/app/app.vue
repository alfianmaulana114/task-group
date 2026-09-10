<script setup>
const { initAuth } = useAuth()
const { initTheme, isDark } = useTheme()
initAuth()

useHead({
  htmlAttrs: {
    class: computed(() => isDark.value ? 'dark' : ''),
  },
})

onMounted(() => {
  initTheme()
})
</script>

<template>
  <div>
    <NuxtRouteAnnouncer />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <UiToast />
  </div>
</template>

<style>
.page-enter-active,
.page-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.page-enter-from {
  opacity: 0;
  transform: translateY(6px);
}
.page-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.layout-enter-active,
.layout-leave-active {
  transition: opacity 0.25s ease;
}
.layout-enter-from,
.layout-leave-to {
  opacity: 0;
}
</style>
