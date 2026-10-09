import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createRouter, createWebHistory } from "vue-router";
import ItemDetail from "../ItemDetail.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [{ path: "/", component: { template: "<div />" } }],
});

function makeItem(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    feed_id: 1,
    guid: "1",
    title: "Test Post",
    url: "https://ex.com/1",
    description: "<p>Body</p>",
    fetched_at: "2024-01-01T00:00:00Z",
    published_at: "2024-01-02T00:00:00Z",
    read: false,
    starred: false,
    ...overrides,
  };
}

function mountDetail(item = makeItem()) {
  return mount(ItemDetail, {
    props: { item },
    global: { plugins: [router] },
  });
}

describe("ItemDetail", () => {
  it("renders the read and save buttons", () => {
    const wrapper = mountDetail();
    expect(wrapper.find('button[title="Mark as read"]').exists()).toBe(true);
    expect(wrapper.find('button[title="Save"]').exists()).toBe(true);
  });

  it("renders the actions after the date", () => {
    const wrapper = mountDetail();
    const date = wrapper.find("article > div.mt-2");
    const actions = date.element.nextElementSibling;
    expect(date.text()).toContain("2024");
    expect(actions?.querySelector('button[title="Save"]')).not.toBeNull();
  });

  it("emits close when the X button next to the title is clicked", async () => {
    const wrapper = mountDetail();
    const closeButton = wrapper.find("[data-close-button]");
    expect(closeButton.find('use[href="#icon-x"]').exists()).toBe(true);
    await closeButton.trigger("click");
    expect(wrapper.emitted("close")).toHaveLength(1);
  });

  it("emits close when the invisible region above the title is clicked", async () => {
    const wrapper = mountDetail();
    await wrapper.find("[data-close-region]").trigger("click");
    expect(wrapper.emitted("close")).toHaveLength(1);
  });

  it("emits toggleRead when the read button is clicked", async () => {
    const item = makeItem();
    const wrapper = mountDetail(item);
    await wrapper.find('button[title="Mark as read"]').trigger("click");
    expect(wrapper.emitted("toggleRead")?.[0]).toEqual([item]);
  });

  it("emits toggleStarred when the save button is clicked", async () => {
    const item = makeItem();
    const wrapper = mountDetail(item);
    await wrapper.find('button[title="Save"]').trigger("click");
    expect(wrapper.emitted("toggleStarred")?.[0]).toEqual([item]);
  });
});
