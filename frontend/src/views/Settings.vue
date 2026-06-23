<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useFeedsStore } from '@/stores/feeds'
import { useAuthStore } from '@/stores/auth'
import AddFeedDialog from '@/components/AddFeedDialog.vue'
import { useSidebar } from '@/composables/useSidebar'
import * as api from '@/api/client'
import type { User } from '@/types'

const feedsStore = useFeedsStore()
const auth = useAuthStore()
const showAddFeed = ref(false)
const { toggle } = useSidebar()

onMounted(() => {
  feedsStore.loadFeeds()
  if (auth.isAdmin) {
    loadUsers()
  }
})

// --- Account / change password ---
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordError = ref('')
const passwordSuccess = ref(false)
const changingPassword = ref(false)

async function changePassword() {
  passwordError.value = ''
  passwordSuccess.value = false
  if (!currentPassword.value || !newPassword.value) {
    passwordError.value = 'Both fields are required.'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = 'New passwords do not match.'
    return
  }
  changingPassword.value = true
  try {
    await api.changePassword(currentPassword.value, newPassword.value)
    passwordSuccess.value = true
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (e: any) {
    passwordError.value = e.response?.data || 'Failed to change password.'
  } finally {
    changingPassword.value = false
  }
}

// --- Administration / users ---
const users = ref<User[]>([])
const newUserUsername = ref('')
const newUserPassword = ref('')
const newUserAdmin = ref(false)
const userError = ref('')
const creatingUser = ref(false)

async function loadUsers() {
  try {
    users.value = await api.listUsers()
  } catch {
    users.value = []
  }
}

const canDeleteUser = (u: User) => {
  if (!auth.user) return false
  if (u.id === auth.user.id) return false
  return users.value.length > 1
}

async function createUser() {
  userError.value = ''
  if (!newUserUsername.value || !newUserPassword.value) {
    userError.value = 'Username and password are required.'
    return
  }
  creatingUser.value = true
  try {
    await api.createUser(newUserUsername.value, newUserPassword.value, newUserAdmin.value)
    newUserUsername.value = ''
    newUserPassword.value = ''
    newUserAdmin.value = false
    await loadUsers()
  } catch (e: any) {
    userError.value = e.response?.data || 'Failed to create user.'
  } finally {
    creatingUser.value = false
  }
}

async function deleteUser(u: User) {
  if (!confirm(`Delete user ${u.username}?`)) return
  try {
    await api.deleteUser(u.id)
    await loadUsers()
  } catch (e: any) {
    userError.value = e.response?.data || 'Failed to delete user.'
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
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Settings</h1>
    </div>


    <!-- Feed Subscriptions -->
    <section class="mt-8">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Feed Subscriptions</h2>
        <button
          @click="showAddFeed = true"
          class="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
        >
          + Add Feed
        </button>
      </div>

      <div class="space-y-2">
        <div
          v-for="feed in feedsStore.feeds"
          :key="feed.id"
          class="flex items-center justify-between px-4 py-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700"
        >
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{{ feed.title || feed.url }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ feed.url }}</p>
          </div>
          <button
            @click="feedsStore.removeFeed(feed.id)"
            class="ml-4 text-sm text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
          >
            Remove
          </button>
        </div>
      </div>

      <p v-if="feedsStore.feeds.length === 0" class="text-sm text-gray-500 dark:text-gray-400 mt-4">
        No feeds subscribed yet. Add one to get started!
      </p>
    </section>

    <!-- Account -->
    <section class="mt-10">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Account</h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        Signed in as
        <span class="font-medium text-gray-700 dark:text-gray-300">{{ auth.user?.username }}</span>
        <span
          v-if="auth.isAdmin"
          class="ml-2 px-2 py-0.5 text-xs font-medium rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
        >admin</span>
      </p>

      <div class="mt-4 max-w-md">
        <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Change password</h3>
        <div class="mt-2 space-y-3">
          <input
            v-model="currentPassword"
            type="password"
            placeholder="Current password"
            autocomplete="current-password"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          />
          <input
            v-model="newPassword"
            type="password"
            placeholder="New password"
            autocomplete="new-password"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          />
          <input
            v-model="confirmPassword"
            type="password"
            placeholder="Confirm new password"
            autocomplete="new-password"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          />
          <p v-if="passwordError" class="text-sm text-red-600 dark:text-red-400">{{ passwordError }}</p>
          <p v-if="passwordSuccess" class="text-sm text-green-600 dark:text-green-400">Password updated.</p>
          <button
            @click="changePassword"
            :disabled="changingPassword"
            class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {{ changingPassword ? 'Saving...' : 'Change password' }}
          </button>
        </div>
      </div>
    </section>

    <!-- Administration -->
    <section v-if="auth.isAdmin" class="mt-10">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Administration</h2>

      <div class="mt-4">
        <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Users</h3>
        <div class="mt-2 space-y-2">
          <div
            v-for="u in users"
            :key="u.id"
            class="flex items-center justify-between px-4 py-2 bg-white dark:bg-gray-800 rounded-md border border-gray-200 dark:border-gray-700"
          >
            <div class="flex items-center gap-2">
              <span class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ u.username }}</span>
              <span
                v-if="u.is_admin"
                class="px-2 py-0.5 text-xs font-medium rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
              >admin</span>
            </div>
            <button
              v-if="canDeleteUser(u)"
              @click="deleteUser(u)"
              class="text-sm text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
            >
              Delete
            </button>
            <span v-else class="text-xs text-gray-400 dark:text-gray-500">—</span>
          </div>
        </div>
      </div>

      <div class="mt-6 max-w-md">
        <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Create new user</h3>
        <div class="mt-2 space-y-3">
          <input
            v-model="newUserUsername"
            type="text"
            placeholder="Username"
            autocomplete="username"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          />
          <input
            v-model="newUserPassword"
            type="password"
            placeholder="Password"
            autocomplete="new-password"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          />
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
            {{ creatingUser ? 'Creating...' : 'Create user' }}
          </button>
        </div>
      </div>
    </section>

    <AddFeedDialog v-if="showAddFeed" @close="showAddFeed = false" />
  </div>
</template>