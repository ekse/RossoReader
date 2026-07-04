import { ref } from "vue";

const currentItemId = ref<number | null>(null);
const expandedItems = ref<Record<number, boolean>>({});

export function useCurrentItem() {
  function set(id: number | null) {
    currentItemId.value = id;
  }

  function expandItem(id: number) {
    expandedItems.value[id] = true;
    currentItemId.value = id;
  }

  function collapseItem(id: number) {
    expandedItems.value[id] = false;
    if (currentItemId.value === id) {
      currentItemId.value = null;
    }
  }

  function toggleExpandItem(id: number) {
    if (expandedItems.value[id]) {
      collapseItem(id);
    } else {
      expandItem(id);
    }
  }

  function isExpanded(id: number): boolean {
    return !!expandedItems.value[id];
  }

  function clearExpanded() {
    expandedItems.value = {};
  }

  return {
    currentItemId,
    expandedItems,
    set,
    expandItem,
    collapseItem,
    toggleExpandItem,
    isExpanded,
    clearExpanded,
  };
}
