import { ref } from "vue";

const currentItemId = ref<number | null>(null);

export function useCurrentItem() {
  function set(id: number | null) {
    currentItemId.value = id;
  }

  return { currentItemId, set };
}
