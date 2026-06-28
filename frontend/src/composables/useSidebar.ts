import { ref } from "vue";

const isOpen = ref(true);

export function useSidebar() {
  function toggle() {
    isOpen.value = !isOpen.value;
  }
  function close() {
    isOpen.value = false;
  }
  function open() {
    isOpen.value = true;
  }

  return {
    isOpen,
    toggle,
    close,
    open,
  };
}
