import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import TopBar from "../TopBar.vue";

describe("TopBar", () => {
  it("does not render the mark all as read button by default", () => {
    const wrapper = mount(TopBar);
    expect(wrapper.find('button[aria-label="Mark all as read"]').exists()).toBe(false);
  });

  it("renders a mobile icon and a desktop label when showMarkAllRead is set", () => {
    const wrapper = mount(TopBar, { props: { showMarkAllRead: true } });
    const button = wrapper.find('button[aria-label="Mark all as read"]');
    expect(button.exists()).toBe(true);

    const icon = button.find("svg");
    expect(icon.find('use[href="#icon-envelope-open"]').exists()).toBe(true);
    expect(icon.classes()).toContain("md:hidden");

    const label = button.find("span");
    expect(label.text()).toBe("Mark all as read");
    expect(label.classes()).toContain("hidden");
    expect(label.classes()).toContain("md:inline");
  });

  it("emits markAllRead when clicked", async () => {
    const wrapper = mount(TopBar, { props: { showMarkAllRead: true } });
    await wrapper.find('button[aria-label="Mark all as read"]').trigger("click");
    expect(wrapper.emitted("markAllRead")).toHaveLength(1);
  });
});
