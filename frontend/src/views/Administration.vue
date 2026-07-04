<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useAuthStore } from "@/stores/auth";
import { useSidebar } from "@/composables/useSidebar";
import * as api from "@/api/client";
import type { User } from "@/types";

const auth = useAuthStore();
const { toggle } = useSidebar();

const users = ref<User[]>([]);
const newUserUsername = ref("");
const newUserPassword = ref("");
const newUserAdmin = ref(false);
const userError = ref("");
const creatingUser = ref(false);
const showNewUserPassword = ref(false);

const itemsLimit = ref(150);
const savingLimit = ref(false);
const limitError = ref("");
const limitSaved = ref(false);

async function loadUsers() {
  try {
    users.value = await api.listUsers();
  } catch {
    users.value = [];
  }
}

const canDeleteUser = (u: User) => {
  if (!auth.user) return false;
  if (u.id === auth.user.id) return false;
  return users.value.length > 1;
};

async function createUser() {
  userError.value = "";
  if (!newUserUsername.value || !newUserPassword.value) {
    userError.value = "Username and password are required.";
    return;
  }
  creatingUser.value = true;
  try {
    await api.createUser(newUserUsername.value, newUserPassword.value, newUserAdmin.value);
    newUserUsername.value = "";
    newUserPassword.value = "";
    newUserAdmin.value = false;
    await loadUsers();
  } catch (e: any) {
    userError.value = e.response?.data || "Failed to create user.";
  } finally {
    creatingUser.value = false;
  }
}

async function deleteUser(u: User) {
  if (!confirm(`Delete user ${u.username}?`)) return;
  try {
    await api.deleteUser(u.id);
    await loadUsers();
  } catch (e: any) {
    userError.value = e.response?.data || "Failed to delete user.";
  }
}

onMounted(() => {
  loadUsers();
  loadAdminSettings();
});

async function loadAdminSettings() {
  try {
    const settings = await api.fetchAdminSettings();
    itemsLimit.value = settings.items_limit;
  } catch {
    // defaults to 150
  }
}

async function saveLimit() {
  limitError.value = "";
  limitSaved.value = false;
  if (itemsLimit.value < 1) {
    limitError.value = "Limit must be at least 1.";
    return;
  }
  savingLimit.value = true;
  try {
    await api.updateAdminSettings({ items_limit: itemsLimit.value });
    limitSaved.value = true;
  } catch (e: any) {
    limitError.value = e.response?.data || "Failed to save setting.";
  } finally {
    savingLimit.value = false;
  }
}
</script>

<template>
  <div class="max-w-3xl mx-auto px-4 py-6 md:px-6 md:py-8">
    <div class="flex items-center gap-3 mb-6 md:mb-8">
      <button
        @click="toggle"
        class="p-1.5 -ml-1 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:bg-gray-700 md:hidden transition-colors"
        title="Toggle Sidebar"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 6h16M4 12h16M4 18h16"
          />
        </svg>
      </button>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Administration</h1>
    </div>

    <section>
      <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Users</h2>
      <div class="space-y-2">
        <div
          v-for="u in users"
          :key="u.id"
          class="flex items-center justify-between px-4 py-2 bg-white dark:bg-gray-800 rounded-md border border-gray-200 dark:border-gray-700"
        >
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-gray-900 dark:text-gray-100">{{
              u.username
            }}</span>
            <span
              v-if="u.is_admin"
              class="px-2 py-0.5 text-xs font-medium rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
              >admin</span
            >
          </div>
          <button
            v-if="canDeleteUser(u)"
            @click="deleteUser(u)"
            class="text-sm text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
          >
            Delete
          </button>
          <span v-else class="text-xs text-gray-400 dark:text-gray-500">&mdash;</span>
        </div>
      </div>
    </section>

    <section class="mt-6 max-w-md">
      <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Create new user</h3>
      <div class="mt-2 space-y-3">
        <input
          v-model="newUserUsername"
          type="text"
          placeholder="Username"
          autocomplete="username"
          class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
        />
        <div class="relative">
          <input
            v-model="newUserPassword"
            :type="showNewUserPassword ? 'text' : 'password'"
            placeholder="Password"
            autocomplete="new-password"
            class="w-full px-3 py-2 pr-10 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          />
          <button
            type="button"
            @click="showNewUserPassword = !showNewUserPassword"
            class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            tabindex="-1"
          >
            <svg v-if="showNewUserPassword" class="w-4 h-4">
              <use href="#icon-eye-slash" />
            </svg>
            <svg v-else class="w-4 h-4">
              <use href="#icon-eye" />
            </svg>
          </button>
        </div>
        <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
          <input v-model="newUserAdmin" type="checkbox" class="rounded" />
          Admin
        </label>
        <p v-if="userError" class="text-sm text-red-600 dark:text-red-400">{{ userError }}</p>
        <button
          @click="createUser"
          :disabled="creatingUser"
          class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
        >
          {{ creatingUser ? "Creating..." : "Create user" }}
        </button>
      </div>
    </section>

    <section class="mt-8 max-w-md">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Application Settings</h2>
      <div class="mt-4">
        <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300"
          >Items per feed limit</label
        >
        <div class="mt-2 flex items-center gap-2">
          <input
            v-model.number="itemsLimit"
            type="number"
            min="1"
            class="w-24 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          />
          <button
            @click="saveLimit"
            :disabled="savingLimit"
            class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {{ savingLimit ? "Saving..." : "Save" }}
          </button>
        </div>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          Items beyond this limit will be purged every 6 hours. Starred items are never deleted.
        </p>
        <p v-if="limitError" class="mt-1 text-sm text-red-600 dark:text-red-400">
          {{ limitError }}
        </p>
        <p v-if="limitSaved" class="mt-1 text-sm text-green-600 dark:text-green-400">Saved.</p>
      </div>
    </section>
  </div>
</template>
