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

  it("emits toggleRead when an unread item is expanded", async () => {
    const item = {
      id: 1,
      feed_id: 1,
      guid: "1",
      title: "Test Post",
      url: "https://ex.com/1",
      fetched_at: "2024-01-01T00:00:00Z",
      read: false,
      starred: false,
    };
    const wrapper = mount(ItemList, {
      props: { items: [item] },
    });
    await wrapper.find(".cursor-pointer").trigger("click");
    expect(wrapper.emitted("toggleRead")).toBeTruthy();
    expect(wrapper.emitted("toggleRead")?.[0]).toEqual([item]);
  });

  it("does not emit toggleRead when an already read item is expanded", async () => {
    const item = {
      id: 1,
      feed_id: 1,
      guid: "1",
      title: "Test Post",
      url: "https://ex.com/1",
      fetched_at: "2024-01-01T00:00:00Z",
      read: true,
      starred: false,
    };
    const wrapper = mount(ItemList, {
      props: { items: [item] },
    });
    await wrapper.find(".cursor-pointer").trigger("click");
    expect(wrapper.emitted("toggleRead")).toBeFalsy();
  });

  it("resets expanded items when items prop changes", async () => {
    const item1 = {
      id: 1,
      feed_id: 1,
      guid: "1",
      title: "Test Post 1",
      url: "https://ex.com/1",
      fetched_at: "2024-01-01T00:00:00Z",
      read: true,
      starred: false,
    };
    const wrapper = mount(ItemList, {
      props: { items: [item1] },
    });
    await wrapper.find(".cursor-pointer").trigger("click");
    expect(wrapper.find(".cursor-default").exists()).toBe(true);

    const item2 = {
      id: 2,
      feed_id: 1,
      guid: "2",
      title: "Test Post 2",
      url: "https://ex.com/2",
      fetched_at: "2024-01-01T00:00:00Z",
      read: true,
      starred: false,
    };
    await wrapper.setProps({ items: [item2] });
    expect(wrapper.find(".cursor-default").exists()).toBe(false);
  });
});
