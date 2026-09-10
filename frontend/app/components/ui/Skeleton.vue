<script setup>
const props = defineProps({
  lines: { type: Number, default: 1 },
  height: { type: String, default: 'h-4' },
  rounded: { type: String, default: 'rounded' },
})
</script>

<template>
  <div class="animate-pulse space-y-2">
    <div
      v-for="i in lines"
      :key="i"
      class="bg-gray-200"
      :class="[height, rounded]"
      :style="i === lines ? 'width: 60%' : 'width: 100%'"
    ></div>
  </div>
</template>
