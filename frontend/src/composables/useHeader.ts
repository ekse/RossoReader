import { ref } from 'vue'

const isHeaderVisible = ref(true)
const lastScrollTop = ref(0)

export function useHeader() {
  return {
    isHeaderVisible,
    lastScrollTop,
  }
}
