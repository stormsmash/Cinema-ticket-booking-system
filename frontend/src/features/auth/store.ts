import { ref } from 'vue'
import { defineStore } from 'pinia'

import { fetchAuthConfig, fetchCurrentUser, logout as requestLogout } from './api'
import type { AuthUser } from './api'

const callbackMessages: Record<string, string> = {
  access_denied: 'Google sign-in was cancelled.',
  invalid_state: 'The sign-in request expired. Please try again.',
  missing_code: 'Google did not return a sign-in code.',
  login_failed: 'Google sign-in could not be completed.',
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null)
  const googleEnabled = ref(false)
  const isLoading = ref(false)
  const initialized = ref(false)
  const error = ref('')
  let loadPromise: Promise<void> | null = null

  function load() {
    if (initialized.value) return Promise.resolve()
    if (loadPromise) return loadPromise

    loadPromise = loadAuth().finally(() => {
      loadPromise = null
    })
    return loadPromise
  }

  async function loadAuth() {
    isLoading.value = true
    error.value = readCallbackError()

    try {
      const config = await fetchAuthConfig()
      googleEnabled.value = config.google_enabled
      user.value = await fetchCurrentUser()
    } catch {
      error.value = 'Unable to check the sign-in status.'
    } finally {
      initialized.value = true
      isLoading.value = false
    }
  }

  async function logout() {
    error.value = ''

    try {
      await requestLogout()
      user.value = null
    } catch {
      error.value = 'Unable to sign out. Please try again.'
    }
  }

  return {
    user,
    googleEnabled,
    isLoading,
    initialized,
    error,
    load,
    ensureLoaded: load,
    logout,
  }
})

function readCallbackError() {
  const url = new URL(window.location.href)
  const code = url.searchParams.get('auth_error')
  if (!code) return ''

  url.searchParams.delete('auth_error')
  window.history.replaceState({}, '', url)
  return callbackMessages[code] ?? 'Google sign-in could not be completed.'
}
