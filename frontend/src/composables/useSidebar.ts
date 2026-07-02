import { ref } from "vue";

const isOpen = ref(true);

const STORAGE_KEY = "sidebarWidth";
export const MIN_WIDTH = 200;
export const MAX_WIDTH = 600;
export const DEFAULT_WIDTH = 288;

function clamp(val: number, min: number, max: number) {
  return Math.min(max, Math.max(min, val));
}

const saved = localStorage.getItem(STORAGE_KEY);
const sidebarWidth = ref(
  saved !== null ? clamp(Number(saved), MIN_WIDTH, MAX_WIDTH) : DEFAULT_WIDTH,
);

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

  function startResize(e: MouseEvent) {
    e.preventDefault();
    const startX = e.clientX;
    const startWidth = sidebarWidth.value;

    function onMouseMove(e: MouseEvent) {
      sidebarWidth.value = clamp(
        startWidth + e.clientX - startX,
        MIN_WIDTH,
        MAX_WIDTH,
      );
    }

    function onMouseUp() {
      localStorage.setItem(STORAGE_KEY, String(sidebarWidth.value));
      document.removeEventListener("mousemove", onMouseMove);
      document.removeEventListener("mouseup", onMouseUp);
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
    }

    document.addEventListener("mousemove", onMouseMove);
    document.addEventListener("mouseup", onMouseUp);
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
  }

  return { isOpen, sidebarWidth, toggle, close, open, startResize };
}
