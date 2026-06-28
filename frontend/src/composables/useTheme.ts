import { ref } from 'vue'

const saved = localStorage.getItem('theme')
const isDark = ref(saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches))

export function useTheme() {
  function toggle() {
    isDark.value = !isDark.value
    if (isDark.value) {
      document.documentElement.classList.add('dark')
      localStorage.setItem('theme', 'dark')
    } else {
      document.documentElement.classList.remove('dark')
      localStorage.setItem('theme', 'light')
    }
  }

  return { isDark, toggle }
}
