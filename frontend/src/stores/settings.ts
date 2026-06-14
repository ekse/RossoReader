import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as api from '@/api/client'

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<Record<string, string>>({})
  const loading = ref(false)

  async function loadSettings() {
    loading.value = true
    try {
      settings.value = await api.fetchSettings()
    } finally {
      loading.value = false
    }
  }

  async function updateSettings(data: Record<string, string>) {
    settings.value = await api.updateSettings(data)
  }

  return { settings, loading, loadSettings, updateSettings }
})
