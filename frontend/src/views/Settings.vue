<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useFeedsStore } from "@/stores/feeds";
import { useAuthStore } from "@/stores/auth";
import AddFeedDialog from "@/components/AddFeedDialog.vue";
import ImportFeedsDialog from "@/components/ImportFeedsDialog.vue";
import { useSidebar } from "@/composables/useSidebar";
import { startPasskeyRegistration } from "@/lib/webauthn";
import * as api from "@/api/client";
import type { Passkey } from "@/types";

const feedsStore = useFeedsStore();
const auth = useAuthStore();
const showAddFeed = ref(false);
const showImportDialog = ref(false);
const expandedErrors = ref(new Set<number>());
const { toggle } = useSidebar();

function toggleError(feedId: number) {
  const s = new Set(expandedErrors.value);
  if (s.has(feedId)) {
    s.delete(feedId);
  } else {
    s.add(feedId);
  }
  expandedErrors.value = s;
}

async function handleExport() {
  const blob = await api.exportOpml();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "feeds.opml";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

onMounted(() => {
  feedsStore.loadFeeds();
  feedsStore.loadFeedsLimit();
  loadPasskeys();
});

// --- Account / change password ---
const currentPassword = ref("");
const newPassword = ref("");
const confirmPassword = ref("");
const passwordError = ref("");
const passwordSuccess = ref(false);
const changingPassword = ref(false);
const showCurrentPassword = ref(false);
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);

async function changePassword() {
  passwordError.value = "";
  passwordSuccess.value = false;
  if (!currentPassword.value || !newPassword.value) {
    passwordError.value = "Both fields are required.";
    return;
  }
  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = "New passwords do not match.";
    return;
  }
  changingPassword.value = true;
  try {
    await api.changePassword(currentPassword.value, newPassword.value);
    passwordSuccess.value = true;
    currentPassword.value = "";
    newPassword.value = "";
    confirmPassword.value = "";
  } catch (e: any) {
    passwordError.value = e.response?.data || "Failed to change password.";
  } finally {
    changingPassword.value = false;
  }
}

// --- Passkeys ---
const passkeys = ref<Passkey[]>([]);
const passkeyError = ref("");
const passkeySuccess = ref(false);
const registeringPasskey = ref(false);

async function loadPasskeys() {
  try {
    passkeys.value = await api.listPasskeys();
  } catch {
    passkeys.value = [];
  }
}

async function registerPasskey() {
  passkeyError.value = "";
  passkeySuccess.value = false;
  const name = window.prompt("Name for this passkey:");
  if (!name) return;
  registeringPasskey.value = true;
  try {
    await startPasskeyRegistration(
      () => api.passkeyRegisterBegin(),
      (stateId, _name, credential) => api.passkeyRegisterFinish(stateId, _name, credential),
      name,
    );
    passkeySuccess.value = true;
    await loadPasskeys();
  } catch (e: any) {
    if (e.name === "NotAllowedError" || e.message?.includes("cancelled")) {
      return;
    }
    passkeyError.value = e.response?.data || e.message || "Failed to register passkey.";
  } finally {
    registeringPasskey.value = false;
  }
}

async function deletePasskey(pk: Passkey) {
  if (!confirm(`Delete passkey "${pk.name}"?`)) return;
  try {
    await api.deletePasskey(pk.id);
    passkeys.value = passkeys.value.filter((p) => p.id !== pk.id);
  } catch (e: any) {
    passkeyError.value = e.response?.data || "Failed to delete passkey.";
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
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Settings</h1>
    </div>

    <!-- Feed Subscriptions -->
    <section class="mt-8">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">
          Feed Subscriptions ({{ feedsStore.feeds.length }}/{{ feedsStore.feedsLimit }})
        </h2>
        <div class="flex items-center gap-2">
          <button
            @click="handleExport"
            class="px-3 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600"
          >
            Export
          </button>
          <button
            @click="showImportDialog = true"
            class="px-3 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600"
          >
            Import
          </button>
          <button
            @click="showAddFeed = true"
            class="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
          >
            + Add Feed
          </button>
        </div>
      </div>

      <div class="space-y-2">
        <div
          v-for="feed in feedsStore.feeds"
          :key="feed.id"
          class="flex items-center justify-between px-4 py-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700"
        >
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
              {{ feed.title || feed.url }}
              <span
                v-if="feed.last_fetch_error"
                @click.stop="toggleError(feed.id)"
                class="cursor-pointer mr-1.5 select-none"
                title="Show error"
                >⚠️</span
              >
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ feed.url }}</p>
            <p
              v-if="feed.last_fetch_error && expandedErrors.has(feed.id)"
              class="mt-1.5 text-xs text-red-600 dark:text-red-400 break-words"
            >
              <b>Last error:</b> {{ feed.last_fetch_error }}
            </p>
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
          >admin</span
        >
      </p>

      <div class="mt-4 max-w-md">
        <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Change password</h3>
        <div class="mt-2 space-y-3">
          <div class="relative">
            <input
              v-model="currentPassword"
              :type="showCurrentPassword ? 'text' : 'password'"
              placeholder="Current password"
              autocomplete="current-password"
              class="w-full px-3 py-2 pr-10 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
            />
            <button
              type="button"
              @click="showCurrentPassword = !showCurrentPassword"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              tabindex="-1"
            >
              <svg v-if="showCurrentPassword" class="w-4 h-4">
                <use href="#icon-eye-slash" />
              </svg>
              <svg v-else class="w-4 h-4">
                <use href="#icon-eye" />
              </svg>
            </button>
          </div>
          <div class="relative">
            <input
              v-model="newPassword"
              :type="showNewPassword ? 'text' : 'password'"
              placeholder="New password"
              autocomplete="new-password"
              class="w-full px-3 py-2 pr-10 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
            />
            <button
              type="button"
              @click="showNewPassword = !showNewPassword"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              tabindex="-1"
            >
              <svg v-if="showNewPassword" class="w-4 h-4">
                <use href="#icon-eye-slash" />
              </svg>
              <svg v-else class="w-4 h-4">
                <use href="#icon-eye" />
              </svg>
            </button>
          </div>
          <div class="relative">
            <input
              v-model="confirmPassword"
              :type="showConfirmPassword ? 'text' : 'password'"
              placeholder="Confirm new password"
              autocomplete="new-password"
              class="w-full px-3 py-2 pr-10 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
            />
            <button
              type="button"
              @click="showConfirmPassword = !showConfirmPassword"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              tabindex="-1"
            >
              <svg v-if="showConfirmPassword" class="w-4 h-4">
                <use href="#icon-eye-slash" />
              </svg>
              <svg v-else class="w-4 h-4">
                <use href="#icon-eye" />
              </svg>
            </button>
          </div>
          <p v-if="passwordError" class="text-sm text-red-600 dark:text-red-400">
            {{ passwordError }}
          </p>
          <p v-if="passwordSuccess" class="text-sm text-green-600 dark:text-green-400">
            Password updated.
          </p>
          <button
            @click="changePassword"
            :disabled="changingPassword"
            class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {{ changingPassword ? "Saving..." : "Change password" }}
          </button>
        </div>
      </div>
    </section>

    <!-- Passkeys -->
    <section class="mt-10">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Passkeys</h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        Register a passkey to sign in without a password.
      </p>

      <div class="mt-4 space-y-2">
        <div
          v-for="pk in passkeys"
          :key="pk.id"
          class="flex items-center justify-between px-4 py-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700"
        >
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="text-sm font-medium text-gray-900 dark:text-gray-100">{{
                pk.name
              }}</span>
              <span v-if="pk.backup_eligible" class="text-xs text-green-600 dark:text-green-400"
                >syncable</span
              >
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              Created {{ new Date(pk.created_at).toLocaleDateString() }}
              <span v-if="pk.transports.length">· {{ pk.transports.join(", ") }}</span>
            </p>
          </div>
          <button
            @click="deletePasskey(pk)"
            class="ml-4 text-sm text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
          >
            Delete
          </button>
        </div>

        <p v-if="passkeys.length === 0" class="text-sm text-gray-500 dark:text-gray-400 mt-2">
          No passkeys registered yet.
        </p>
      </div>

      <p v-if="passkeyError" class="mt-2 text-sm text-red-600 dark:text-red-400">
        {{ passkeyError }}
      </p>
      <p v-if="passkeySuccess" class="mt-2 text-sm text-green-600 dark:text-green-400">
        Passkey registered.
      </p>

      <button
        @click="registerPasskey"
        :disabled="registeringPasskey"
        class="mt-3 px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
      >
        {{ registeringPasskey ? "Registering..." : "Register new passkey" }}
      </button>
    </section>

    <AddFeedDialog v-if="showAddFeed" @close="showAddFeed = false" />
    <ImportFeedsDialog v-if="showImportDialog" @close="showImportDialog = false" />
  </div>
</template>
