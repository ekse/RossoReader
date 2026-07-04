import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";

const mockListUsers = vi.fn();
const mockCreateUser = vi.fn();
const mockDeleteUser = vi.fn();
vi.mock("@/api/client", () => ({
  listUsers: (...args: any[]) => mockListUsers(...args),
  createUser: (...args: any[]) => mockCreateUser(...args),
  deleteUser: (...args: any[]) => mockDeleteUser(...args),
}));

vi.mock("@/composables/useSidebar", () => ({
  useSidebar: () => ({ toggle: vi.fn() }),
}));

import Administration from "@/views/Administration.vue";
import { useAuthStore } from "@/stores/auth";

function setAuthUser(user: { id: number; username: string; is_admin: boolean }) {
  const auth = useAuthStore();
  auth.user = user;
}

describe("Administration", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it("renders the page title", () => {
    setAuthUser({ id: 1, username: "admin", is_admin: true });
    mockListUsers.mockResolvedValue([]);
    const wrapper = mount(Administration);
    expect(wrapper.find("h1").text()).toBe("Administration");
  });

  it("renders user list", async () => {
    setAuthUser({ id: 1, username: "admin", is_admin: true });
    mockListUsers.mockResolvedValue([
      { id: 1, username: "admin", is_admin: true },
      { id: 2, username: "bob", is_admin: false },
    ]);
    const wrapper = mount(Administration);
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(wrapper.text()).toContain("admin");
    expect(wrapper.text()).toContain("bob");
  });

  it("shows admin badge for admin users", async () => {
    setAuthUser({ id: 1, username: "admin", is_admin: true });
    mockListUsers.mockResolvedValue([
      { id: 1, username: "admin", is_admin: true },
      { id: 2, username: "bob", is_admin: false },
    ]);
    const wrapper = mount(Administration);
    await new Promise((resolve) => setTimeout(resolve, 0));

    const userRows = wrapper.findAll(".flex.items-center.justify-between");
    const firstRowHtml = userRows[0].html();
    expect(firstRowHtml).toContain("rounded-full");
    expect(firstRowHtml).toContain("admin");
  });

  it("current user cannot delete themselves", async () => {
    setAuthUser({ id: 1, username: "admin", is_admin: true });
    mockListUsers.mockResolvedValue([
      { id: 1, username: "admin", is_admin: true },
      { id: 2, username: "bob", is_admin: false },
    ]);
    const wrapper = mount(Administration);
    await new Promise((resolve) => setTimeout(resolve, 0));

    const rows = wrapper.findAll(".flex.items-center.justify-between");
    expect(rows.length).toBe(2);

    const firstRowText = rows[0].find(".text-xs.text-gray-400").text();
    expect(firstRowText).toContain("—");
  });

  it("can delete another user", async () => {
    setAuthUser({ id: 1, username: "admin", is_admin: true });
    mockListUsers.mockResolvedValue([
      { id: 1, username: "admin", is_admin: true },
      { id: 2, username: "bob", is_admin: false },
    ]);
    const wrapper = mount(Administration);
    await new Promise((resolve) => setTimeout(resolve, 0));

    const deleteButtons = wrapper.findAll("button").filter((b) => b.text().trim() === "Delete");
    expect(deleteButtons.length).toBe(1);
  });

  it("creates a user", async () => {
    setAuthUser({ id: 1, username: "admin", is_admin: true });
    mockListUsers.mockResolvedValue([]);
    mockCreateUser.mockResolvedValue(undefined);

    const wrapper = mount(Administration);
    await new Promise((resolve) => setTimeout(resolve, 0));

    const usernameInput = wrapper.find('input[placeholder="Username"]');
    await usernameInput.setValue("carol");

    const passwordInput = wrapper.find('input[placeholder="Password"]');
    await passwordInput.setValue("secret123");

    const createBtn = wrapper.findAll("button").find((b) => b.text().trim() === "Create user");
    await createBtn!.trigger("click");
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(mockCreateUser).toHaveBeenCalledWith("carol", "secret123", false);
  });

  it("shows error when username or password is missing", async () => {
    setAuthUser({ id: 1, username: "admin", is_admin: true });
    mockListUsers.mockResolvedValue([]);

    const wrapper = mount(Administration);
    await new Promise((resolve) => setTimeout(resolve, 0));

    const createBtn = wrapper.findAll("button").find((b) => b.text().trim() === "Create user");
    await createBtn!.trigger("click");
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(wrapper.text()).toContain("Username and password are required.");
  });

  it("creates admin user when checkbox is checked", async () => {
    setAuthUser({ id: 1, username: "admin", is_admin: true });
    mockListUsers.mockResolvedValue([]);
    mockCreateUser.mockResolvedValue(undefined);

    const wrapper = mount(Administration);
    await new Promise((resolve) => setTimeout(resolve, 0));

    await wrapper.find('input[placeholder="Username"]').setValue("eve");
    await wrapper.find('input[placeholder="Password"]').setValue("pw");
    await wrapper.find('input[type="checkbox"]').setValue(true);

    const createBtn = wrapper.findAll("button").find((b) => b.text().trim() === "Create user");
    await createBtn!.trigger("click");
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(mockCreateUser).toHaveBeenCalledWith("eve", "pw", true);
  });
});
