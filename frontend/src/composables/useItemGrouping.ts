import { computed, ref } from "vue";
import { DEFAULT_ITEM_GROUPING, ITEM_GROUPING_FEED, type ItemGrouping } from "@/types";

const STORAGE_KEY = "itemGrouping";

function readStored(): ItemGrouping {
  const saved = localStorage.getItem(STORAGE_KEY);
  return saved === ITEM_GROUPING_FEED ? ITEM_GROUPING_FEED : DEFAULT_ITEM_GROUPING;
}

const groupingRef = ref<ItemGrouping>(readStored());

export function useItemGrouping() {
  const grouping = computed<ItemGrouping>({
    get: () => groupingRef.value,
    set: (value) => {
      groupingRef.value = value;
      localStorage.setItem(STORAGE_KEY, value);
    },
  });

  function setGrouping(value: ItemGrouping) {
    grouping.value = value;
  }

  return { grouping, setGrouping };
}
