import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createRouter, createWebHistory } from "vue-router";
import ItemList from "../ItemList.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [{ path: "/", component: { template: "<div />" } }],
});

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
      global: { plugins: [router] },
    });
    expect(wrapper.text()).toContain("Test Post");
  });

  it("shows empty state when no items", () => {
    const wrapper = mount(ItemList, {
      props: { items: [] },
      global: { plugins: [router] },
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
      global: { plugins: [router] },
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
      global: { plugins: [router] },
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
      global: { plugins: [router] },
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

  describe("grouping", () => {
    function makeItem(id: number, feedId: number, title: string) {
      return {
        id,
        feed_id: feedId,
        guid: String(id),
        title,
        url: `https://ex.com/${id}`,
        fetched_at: "2024-01-01T00:00:00Z",
        read: false,
        starred: false,
      };
    }

    it("shows the feed name on each row when grouping is none", () => {
      const items = [makeItem(1, 1, "Post One"), makeItem(2, 1, "Post Two")];
      const wrapper = mount(ItemList, {
        props: { items, feedNames: { 1: "Feed One" } },
        global: { plugins: [router] },
      });
      expect(wrapper.findAll("[data-feed-group]")).toHaveLength(0);
      expect(wrapper.text()).toContain("Feed One");
    });

    it("renders a header for each feed and no feed name on rows when grouped by feed", () => {
      const items = [
        makeItem(1, 1, "Post One"),
        makeItem(2, 2, "Post Two"),
        makeItem(3, 1, "Post Three"),
      ];
      const wrapper = mount(ItemList, {
        props: {
          items,
          feedNames: { 1: "Feed One", 2: "Feed Two" },
          groupBy: "feed",
        },
        global: { plugins: [router] },
      });

      const headers = wrapper.findAll("[data-feed-group]");
      expect(headers).toHaveLength(2);
      expect(headers[0].text()).toBe("Feed One");
      expect(headers[1].text()).toBe("Feed Two");
      expect(wrapper.find('[data-feed-group="1"]').exists()).toBe(true);
      expect(wrapper.find('[data-feed-group="2"]').exists()).toBe(true);
    });

    it("does not duplicate the feed name on individual rows when grouped by feed", () => {
      const items = [makeItem(1, 1, "Post One"), makeItem(2, 1, "Post Two")];
      const wrapper = mount(ItemList, {
        props: { items, feedNames: { 1: "Feed One" }, groupBy: "feed" },
        global: { plugins: [router] },
      });
      const occurrences = wrapper.text().split("Feed One").length - 1;
      expect(occurrences).toBe(1);
    });

    it("falls back to Unknown feed when the feed name is unavailable", () => {
      const items = [makeItem(1, 99, "Post One")];
      const wrapper = mount(ItemList, {
        props: { items, feedNames: {}, groupBy: "feed" },
        global: { plugins: [router] },
      });
      expect(wrapper.find('[data-feed-group="99"]').text()).toBe("Unknown feed");
    });
  });
});
