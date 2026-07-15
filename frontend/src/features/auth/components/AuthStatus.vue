<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'

import { useAuthStore } from '../store'

const store = useAuthStore()
const { user, googleEnabled, isLoading, error } = storeToRefs(store)

onMounted(store.load)
</script>

<template>
  <section class="auth-card" aria-live="polite">
    <p class="card-label">Your account</p>

    <p v-if="isLoading" class="muted" role="status">Checking sign-in...</p>

    <template v-else-if="user">
      <div class="profile">
        <img v-if="user.avatar_url" :src="user.avatar_url" alt="" referrerpolicy="no-referrer" />
        <span v-else class="avatar-fallback" aria-hidden="true">{{ user.name.charAt(0) }}</span>
        <div>
          <strong>{{ user.name }}</strong>
          <span>{{ user.email }}</span>
        </div>
      </div>
      <button type="button" class="secondary-button" @click="store.logout">Sign out</button>
    </template>

    <template v-else>
      <p class="muted">Sign in before locking or booking a seat.</p>
      <a v-if="googleEnabled" class="google-button" href="/api/v1/auth/google">
        Continue with Google
      </a>
      <button v-else type="button" class="google-button" disabled>
        Google sign-in unavailable
      </button>
      <small v-if="!googleEnabled">Add Google credentials to the local .env file.</small>
    </template>

    <p v-if="error" class="auth-error" role="alert">{{ error }}</p>
  </section>
</template>

<style scoped>
.auth-card {
  padding: 1.25rem;
  border: 1px solid #3f3f46;
  border-radius: 1rem;
  background: #18181b;
}

.card-label,
.muted,
.auth-error {
  margin: 0;
}

.card-label {
  margin-bottom: 0.8rem;
  color: #a1a1aa;
  font-size: 0.82rem;
}

.muted,
small {
  color: #a1a1aa;
  line-height: 1.5;
}

.profile {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.profile img,
.avatar-fallback {
  display: grid;
  width: 2.5rem;
  height: 2.5rem;
  flex: 0 0 auto;
  border-radius: 50%;
  place-items: center;
  color: #18181b;
  background: #fbbf24;
  font-weight: 800;
}

.profile strong,
.profile span {
  display: block;
}

.profile strong {
  color: #fafafa;
}

.profile span {
  margin-top: 0.15rem;
  overflow-wrap: anywhere;
  color: #a1a1aa;
  font-size: 0.78rem;
}

.google-button,
.secondary-button {
  display: inline-block;
  margin-top: 0.9rem;
  padding: 0.65rem 0.85rem;
  border: 0;
  border-radius: 0.55rem;
  text-decoration: none;
  font-weight: 750;
  cursor: pointer;
}

.google-button {
  color: #18181b;
  background: #fbbf24;
}

.google-button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.secondary-button {
  border: 1px solid #52525b;
  color: #e4e4e7;
  background: transparent;
}

small {
  display: block;
  margin-top: 0.6rem;
  font-size: 0.74rem;
}

.auth-error {
  margin-top: 0.75rem;
  color: #fecaca;
  font-size: 0.8rem;
}
</style>
