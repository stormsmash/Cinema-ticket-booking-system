<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { fetchHealth } from '@/features/system/api'

type CheckState = 'checking' | 'online' | 'offline'

const state = ref<CheckState>('checking')
const detail = ref('Contacting the booking API...')

const label = computed(() => {
  if (state.value === 'online') return 'Backend connected'
  if (state.value === 'offline') return 'Backend unavailable'
  return 'Checking backend'
})

async function checkHealth() {
  state.value = 'checking'
  detail.value = 'Contacting the booking API...'

  try {
    await fetchHealth()
    state.value = 'online'
    detail.value = 'The Vue frontend can reach the Go API.'
  } catch {
    state.value = 'offline'
    detail.value = 'Start the API, then try the connection again.'
  }
}

onMounted(checkHealth)
</script>

<template>
  <section class="status-card" aria-live="polite">
    <div class="status-heading">
      <span class="status-dot" :class="`status-dot--${state}`" aria-hidden="true"></span>
      <div>
        <p>System status</p>
        <strong>{{ label }}</strong>
      </div>
    </div>

    <p class="status-detail">{{ detail }}</p>

    <button type="button" :disabled="state === 'checking'" @click="checkHealth">
      {{ state === 'checking' ? 'Checking…' : 'Check again' }}
    </button>
  </section>
</template>

<style scoped>
.status-card {
  padding: 1.25rem;
  border: 1px solid rgb(255 255 255 / 12%);
  border-radius: 1rem;
  background: rgb(24 24 27 / 78%);
  box-shadow: 0 1.5rem 4rem rgb(0 0 0 / 28%);
}

.status-heading {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-heading p,
.status-detail {
  margin: 0;
  color: #a1a1aa;
}

.status-heading strong {
  display: block;
  margin-top: 0.2rem;
  font-size: 1.05rem;
}

.status-dot {
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 50%;
  background: #fbbf24;
  box-shadow: 0 0 0 0.3rem rgb(251 191 36 / 12%);
}

.status-dot--online {
  background: #22c55e;
  box-shadow: 0 0 0 0.3rem rgb(34 197 94 / 12%);
}

.status-dot--offline {
  background: #ef4444;
  box-shadow: 0 0 0 0.3rem rgb(239 68 68 / 12%);
}

.status-detail {
  margin-top: 1rem;
  line-height: 1.6;
}

button {
  margin-top: 1rem;
  padding: 0.65rem 0.9rem;
  border: 0;
  border-radius: 0.65rem;
  color: #18181b;
  background: #fbbf24;
  font-weight: 700;
  cursor: pointer;
}

button:disabled {
  cursor: wait;
  opacity: 0.65;
}
</style>
