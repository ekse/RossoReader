import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";

vi.mock("@/stores/items", () => ({
  useItemsStore: () => ({
    items: [],
    loading: false,
    hasMore: false,
    setFilterRead: vi.fn(),
    loadItems: vi.fn(),
    toggleRead: vi.fn(),
    toggleStarred: vi.fn(),
    loadMore: vi.fn(),
  }),
}));

vi.mock("@/stores/feeds", () => ({
  useFeedsStore: () => ({ feeds: [], feedNames: {}, feedIcons: {} }),
}));

vi.mock("@/api/client", () => ({
  markAllItemsRead: vi.fn(),
}));

vi.mock("@/composables/useSidebar", () => ({
  useSidebar: () => ({ toggle: vi.fn() }),
}));

vi.mock("@/composables/useHeader", () => ({
  useHeader: () => ({ isHeaderVisible: { value: true } }),
}));

import UnreadItems from "@/views/UnreadItems.vue";
import { useItemGrouping } from "@/composables/useItemGrouping";
import { ITEM_GROUPING_FEED, ITEM_GROUPING_NONE } from "@/types";

describe("UnreadItems grouping toggle", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    const { grouping } = useItemGrouping();
    grouping.value = ITEM_GROUPING_NONE;
  });

  function mountView() {
    return mount(UnreadItems, {
      global: {
        stubs: {
          TopBar: { template: '<div><slot name="actions" /></div>' },
          ItemList: true,
        },
      },
    });
  }

  it("toggles grouping by feed on and off", async () => {
    const { grouping } = useItemGrouping();
    const wrapper = mountView();
    const toggle = wrapper.get('[role="switch"]');

    expect(toggle.attributes("aria-checked")).toBe("false");

    await toggle.trigger("click");
    expect(grouping.value).toBe(ITEM_GROUPING_FEED);
    expect(toggle.attributes("aria-checked")).toBe("true");

    await toggle.trigger("click");
    expect(grouping.value).toBe(ITEM_GROUPING_NONE);
    expect(toggle.attributes("aria-checked")).toBe("false");
  });
});
