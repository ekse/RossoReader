import { ref } from 'vue'

const isHeaderVisible = ref(true)
const lastScrollTop = ref(0)

export function useHeader() {
  function handleScroll(e: Event) {
    if (window.innerWidth >= 768) {
      isHeaderVisible.value = true
      return
    }

    const target = e.target as HTMLElement
    const scrollTop = target.scrollTop

    if (scrollTop < 0) return

    if (scrollTop > lastScrollTop.value && scrollTop > 50) {
      isHeaderVisible.value = false
    } else {
      isHeaderVisible.value = true
    }

    lastScrollTop.value = scrollTop
  }

  return {
    isHeaderVisible,
    lastScrollTop,
    handleScroll,
  }
}
