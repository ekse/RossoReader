import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";

const mockLogin = vi.fn();
vi.mock("@/api/client", () => ({
  login: (...args: any[]) => mockLogin(...args),
  passkeyLoginBegin: vi.fn(),
  passkeyLoginFinish: vi.fn(),
}));

const mockPush = vi.fn();
vi.mock("vue-router", () => ({
  useRouter: () => ({ push: mockPush }),
}));

import Login from "@/views/Login.vue";

describe("Login", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it("shows server error message when API responds with an error", async () => {
    mockLogin.mockRejectedValue({
      response: { status: 401, data: "invalid credentials" },
    });
    const wrapper = mount(Login);

    await wrapper.find('input[autocomplete="username"]').setValue("alice");
    await wrapper.find('input[autocomplete="current-password"]').setValue("wrong");
    await wrapper.find("form").trigger("submit");
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(wrapper.text()).toContain("invalid credentials");
  });

  it("shows generic error when server responds without a body", async () => {
    mockLogin.mockRejectedValue({
      response: { status: 500, data: null },
    });
    const wrapper = mount(Login);

    await wrapper.find('input[autocomplete="username"]').setValue("alice");
    await wrapper.find('input[autocomplete="current-password"]').setValue("pw");
    await wrapper.find("form").trigger("submit");
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(wrapper.text()).toContain("Invalid credentials");
  });

  it("shows network error when backend is unreachable", async () => {
    mockLogin.mockRejectedValue(new Error("Failed to fetch"));
    const wrapper = mount(Login);

    await wrapper.find('input[autocomplete="username"]').setValue("alice");
    await wrapper.find('input[autocomplete="current-password"]').setValue("pw");
    await wrapper.find("form").trigger("submit");
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(wrapper.text()).toContain("Unable to reach the server");
  });

  it("redirects to /unread on successful login", async () => {
    mockLogin.mockResolvedValue({ id: 1, username: "alice", is_admin: false });
    const wrapper = mount(Login);

    await wrapper.find('input[autocomplete="username"]').setValue("alice");
    await wrapper.find('input[autocomplete="current-password"]').setValue("pw");
    await wrapper.find("form").trigger("submit");
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(mockPush).toHaveBeenCalledWith("/unread");
  });

  it("does not submit when username is empty", async () => {
    const wrapper = mount(Login);

    await wrapper.find('input[autocomplete="current-password"]').setValue("pw");
    await wrapper.find("form").trigger("submit");
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(mockLogin).not.toHaveBeenCalled();
  });

  it("does not submit when password is empty", async () => {
    const wrapper = mount(Login);

    await wrapper.find('input[autocomplete="username"]').setValue("alice");
    await wrapper.find("form").trigger("submit");
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(mockLogin).not.toHaveBeenCalled();
  });
});
