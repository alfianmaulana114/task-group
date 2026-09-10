const THEME_KEY = 'taskgroup:theme'

export function useTheme() {
  const isDark = useState<boolean>(THEME_KEY, () => false)

  function setDark(val: boolean) {
    isDark.value = val
    if (import.meta.client) {
      localStorage.setItem(THEME_KEY, String(val))
    }
  }

  function toggle() {
    setDark(!isDark.value)
  }

  function initTheme() {
    if (!import.meta.client) return
    const saved = localStorage.getItem(THEME_KEY)
    let val: boolean
    if (saved === 'true') val = true
    else if (saved === 'false') val = false
    else val = window.matchMedia('(prefers-color-scheme: dark)').matches
    isDark.value = val
    localStorage.setItem(THEME_KEY, String(val))
  }

  return { isDark, setDark, toggle, initTheme }
}
