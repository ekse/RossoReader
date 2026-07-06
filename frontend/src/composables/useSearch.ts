import { ref } from "vue";

const isSearchOpen = ref(false);

export function useSearch() {
  function openSearch() {
    isSearchOpen.value = true;
  }
  function closeSearch() {
    isSearchOpen.value = false;
  }
  return { isSearchOpen, openSearch, closeSearch };
}
