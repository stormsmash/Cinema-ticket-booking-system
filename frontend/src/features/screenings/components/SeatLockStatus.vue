<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import type { SeatLock } from '../types'

const props = defineProps<{
  lock: SeatLock | null
  signedIn: boolean
  isUpdating: boolean
  error: string
}>()

const emit = defineEmits<{
  release: []
  expired: []
}>()

const remainingSeconds = ref(0)
let timer: ReturnType<typeof setInterval> | undefined
let expirationReported = false

const timeLabel = computed(() => {
  const minutes = Math.floor(remainingSeconds.value / 60)
  const seconds = remainingSeconds.value % 60
  return `${minutes}:${seconds.toString().padStart(2, '0')}`
})

function updateRemainingTime() {
  if (!props.lock) {
    remainingSeconds.value = 0
    expirationReported = false
    return
  }

  remainingSeconds.value = Math.max(
    0,
    Math.ceil((new Date(props.lock.expires_at).getTime() - Date.now()) / 1000),
  )

  if (remainingSeconds.value === 0 && !expirationReported) {
    expirationReported = true
    emit('expired')
  }
}

watch(
  () => props.lock?.expires_at,
  () => {
    expirationReported = false
    updateRemainingTime()
  },
)

onMounted(() => {
  updateRemainingTime()
  timer = setInterval(updateRemainingTime, 1000)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="lock-status" aria-live="polite">
    <div v-if="lock" class="active-lock">
      <div>
        <span>Seat {{ lock.seat_id }} held for</span>
        <strong>{{ timeLabel }}</strong>
      </div>
      <button type="button" :disabled="isUpdating" @click="emit('release')">
        {{ isUpdating ? 'Releasing...' : 'Release seat' }}
      </button>
    </div>

    <p v-else-if="signedIn">
      {{
        isUpdating ? 'Locking your seat...' : 'Choose an available seat to hold it for 10 minutes.'
      }}
    </p>
    <p v-else>Sign in with Google before locking a seat.</p>

    <p v-if="error" class="lock-error" role="alert">{{ error }}</p>
  </div>
</template>

<style scoped>
.lock-status {
  min-height: 3.75rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #27272a;
  color: #a1a1aa;
  font-size: 0.9rem;
}

.lock-status p {
  margin: 0;
}

.active-lock {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
}

.active-lock span,
.active-lock strong {
  display: block;
}

.active-lock strong {
  margin-top: 0.2rem;
  color: #fbbf24;
  font-size: 1.35rem;
  font-variant-numeric: tabular-nums;
}

button {
  padding: 0.55rem 0.75rem;
  border: 1px solid #52525b;
  border-radius: 0.45rem;
  color: #e4e4e7;
  background: transparent;
  font-weight: 700;
  cursor: pointer;
}

button:disabled {
  cursor: wait;
  opacity: 0.6;
}

.lock-error {
  margin-top: 0.75rem !important;
  color: #fecaca;
}
</style>
