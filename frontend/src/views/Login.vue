<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { startPasskeyLogin } from "@/lib/webauthn";
import * as api from "@/api/client";

const auth = useAuthStore();
const router = useRouter();

const username = ref("");
const password = ref("");
const error = ref("");
const loading = ref(false);
const passkeyLoading = ref(false);

async function submit() {
  if (!username.value || !password.value) return;
  loading.value = true;
  error.value = "";
  try {
    await auth.login(username.value, password.value);
    router.push("/unread");
  } catch (e: any) {
    error.value = e.response?.data || "Invalid credentials";
  } finally {
    loading.value = false;
  }
}

async function loginWithPasskey() {
  passkeyLoading.value = true;
  error.value = "";
  try {
    const user = await startPasskeyLogin(
      () => api.passkeyLoginBegin(),
      (stateId, credential) => api.passkeyLoginFinish(stateId, credential),
    );
    auth.user = user;
    router.push("/unread");
  } catch (e: any) {
    if (e.name === "NotAllowedError" || e.message?.includes("cancelled")) {
      return;
    }
    error.value = e.response?.data || e.message || "Passkey login failed";
  } finally {
    passkeyLoading.value = false;
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen bg-gray-50 dark:bg-gray-900">
    <div class="w-full max-w-sm mx-auto p-8 bg-white dark:bg-gray-800 rounded-lg shadow-lg">
      <div class="text-center mb-6">
        <div class="text-2xl font-bold tracking-[0.2em] text-red-600">🌹Rosso</div>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Sign in to your account</p>
      </div>

      <form @submit.prevent="submit" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Username</label>
          <input
            v-model="username"
            type="text"
            autocomplete="username"
            autofocus
            class="w-full mt-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Password</label>
          <input
            v-model="password"
            type="password"
            autocomplete="current-password"
            class="w-full mt-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        <p v-if="error" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>

        <button
          type="submit"
          :disabled="loading || !username || !password"
          class="w-full px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
        >
          {{ loading ? "Signing in..." : "Sign in" }}
        </button>
      </form>

      <div class="mt-4 flex items-center gap-2">
        <div class="flex-1 border-t border-gray-300 dark:border-gray-600"></div>
        <span class="text-xs text-gray-400">or</span>
        <div class="flex-1 border-t border-gray-300 dark:border-gray-600"></div>
      </div>

      <button
        @click="loginWithPasskey"
        :disabled="passkeyLoading"
        class="mt-4 w-full px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-50 transition-colors"
      >
        {{ passkeyLoading ? "Checking..." : "Sign in with passkey" }}
      </button>
    </div>
  </div>
</template>
