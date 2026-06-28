import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import ItemList from "../ItemList.vue";

describe("ItemList", () => {
  it("renders a list of items", () => {
    const items = [
      {
        id: 1,
        feed_id: 1,
        guid: "1",
        title: "Test Post",
        url: "https://ex.com/1",
        fetched_at: "2024-01-01T00:00:00Z",
        read: false,
        starred: false,
      },
    ];
    const wrapper = mount(ItemList, {
      props: { items },
    });
    expect(wrapper.text()).toContain("Test Post");
  });

  it("shows empty state when no items", () => {
    const wrapper = mount(ItemList, {
      props: { items: [] },
    });
    expect(wrapper.text()).toContain("No items to show.");
  });
});
