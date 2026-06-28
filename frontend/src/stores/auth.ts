import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { User } from "@/types";
import * as api from "@/api/client";

export const useAuthStore = defineStore("auth", () => {
  const user = ref<User | null>(null);
  const loading = ref(false);

  const isAuthenticated = computed(() => user.value !== null);
  const isAdmin = computed(() => user.value?.is_admin ?? false);

  async function fetchMe() {
    loading.value = true;
    try {
      user.value = await api.getMe();
    } catch {
      user.value = null;
    } finally {
      loading.value = false;
    }
  }

  async function login(username: string, password: string) {
    user.value = await api.login(username, password);
  }

  async function logout() {
    await api.logout();
    user.value = null;
  }

  return { user, loading, isAuthenticated, isAdmin, fetchMe, login, logout };
});
