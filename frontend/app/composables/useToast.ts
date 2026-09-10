interface Toast {
  id: number
  message: string
  type: 'success' | 'error' | 'info' | 'warning'
  duration: number
}

const toasts = ref<Toast[]>([])
let toastId = 0

export function useToast() {
  function add(message: string, type: Toast['type'] = 'info', duration = 3000) {
    const id = ++toastId
    toasts.value.push({ id, message, type, duration })

    if (duration > 0) {
      setTimeout(() => remove(id), duration)
    }
    return id
  }

  function remove(id: number) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  function success(message: string, duration = 3000) {
    return add(message, 'success', duration)
  }

  function error(message: string, duration = 4000) {
    return add(message, 'error', duration)
  }

  function info(message: string, duration = 3000) {
    return add(message, 'info', duration)
  }

  function warning(message: string, duration = 3500) {
    return add(message, 'warning', duration)
  }

  return {
    toasts: readonly(toasts),
    add,
    remove,
    success,
    error,
    info,
    warning,
  }
}
