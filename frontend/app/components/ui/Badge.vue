<script setup>
const props = defineProps({
  variant: {
    type: String,
    default: 'gray',
    validator: (v) => ['gray', 'blue', 'green', 'red', 'amber', 'purple', 'indigo'].includes(v),
  },
  dot: { type: Boolean, default: false },
})

const variantStyles = {
  gray: { bg: 'bg-gray-100', text: 'text-gray-600', dot: 'bg-gray-400' },
  blue: { bg: 'bg-blue-50', text: 'text-blue-700', dot: 'bg-blue-500' },
  green: { bg: 'bg-emerald-50', text: 'text-emerald-700', dot: 'bg-emerald-500' },
  red: { bg: 'bg-red-50', text: 'text-red-700', dot: 'bg-red-500' },
  amber: { bg: 'bg-amber-50', text: 'text-amber-700', dot: 'bg-amber-500' },
  purple: { bg: 'bg-purple-50', text: 'text-purple-700', dot: 'bg-purple-500' },
  indigo: { bg: 'bg-indigo-50', text: 'text-indigo-700', dot: 'bg-indigo-500' },
}

const style = computed(() => variantStyles[props.variant])
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 text-[11px] font-medium px-2 py-0.5 rounded-full"
    :class="[style.bg, style.text]"
  >
    <span v-if="dot" class="w-1.5 h-1.5 rounded-full" :class="style.dot"></span>
    <slot />
  </span>
</template>
