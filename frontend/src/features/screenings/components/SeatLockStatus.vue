<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import type { Booking, SeatLock } from '../types'

const props = defineProps<{
  lock: SeatLock | null
  signedIn: boolean
  isUpdating: boolean
  error: string
  booking: Booking | null
  isConfirming: boolean
  bookingError: string
}>()

const emit = defineEmits<{
  release: []
  expired: []
  confirm: []
}>()

const remainingSeconds = ref(0)
const confirmationDialog = ref<HTMLDialogElement | null>(null)
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

watch(
  () => props.lock,
  (lock) => {
    if (!lock) closeConfirmation()
  },
)

function openConfirmation() {
  const dialog = confirmationDialog.value
  if (!dialog) return
  if (typeof dialog.showModal === 'function') {
    dialog.showModal()
  } else {
    dialog.setAttribute('open', '')
  }
}

function closeConfirmation() {
  const dialog = confirmationDialog.value
  if (!dialog?.open) return
  if (typeof dialog.close === 'function') {
    dialog.close()
  } else {
    dialog.removeAttribute('open')
  }
}

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
      <div class="lock-actions">
        <button
          type="button"
          class="confirm-button"
          :disabled="isUpdating || isConfirming"
          @click="openConfirmation"
        >
          Confirm booking
        </button>
        <button type="button" :disabled="isUpdating || isConfirming" @click="emit('release')">
          {{ isUpdating ? 'Releasing...' : 'Release seat' }}
        </button>
      </div>
    </div>

    <div v-else-if="booking" class="booking-success" role="status">
      <strong>Seat {{ booking.seat_id }} is booked.</strong>
      <span>Booking reference {{ booking.id }}</span>
    </div>

    <p v-else-if="signedIn">
      {{
        isUpdating ? 'Locking your seat...' : 'Choose an available seat to hold it for 5 minutes.'
      }}
    </p>
    <p v-else>Sign in with Google before locking a seat.</p>

    <p v-if="error" class="lock-error" role="alert">{{ error }}</p>
    <p v-if="bookingError && !lock" class="lock-error" role="alert">{{ bookingError }}</p>

    <dialog
      ref="confirmationDialog"
      class="confirmation-dialog"
      aria-labelledby="confirmation-title"
      @cancel="closeConfirmation"
    >
      <form method="dialog" @submit.prevent>
        <p class="dialog-label">Mock payment</p>
        <h3 id="confirmation-title">Book seat {{ lock?.seat_id }}?</h3>
        <p>
          This simulates a successful payment. The seat will become permanently booked and cannot be
          selected again.
        </p>
        <p v-if="bookingError" class="dialog-error" role="alert">{{ bookingError }}</p>
        <div class="dialog-actions">
          <button type="button" :disabled="isConfirming" @click="closeConfirmation">Go back</button>
          <button
            type="button"
            class="confirm-button"
            :disabled="isConfirming"
            @click="emit('confirm')"
          >
            {{ isConfirming ? 'Confirming...' : `Book seat ${lock?.seat_id ?? ''}` }}
          </button>
        </div>
      </form>
    </dialog>
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

.lock-actions,
.dialog-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  justify-content: flex-end;
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

.confirm-button {
  border-color: #fbbf24;
  color: #18181b;
  background: #fbbf24;
}

.booking-success {
  display: grid;
  gap: 0.3rem;
  color: #d4d4d8;
}

.booking-success strong {
  color: #fbbf24;
}

.booking-success span {
  overflow-wrap: anywhere;
  font-size: 0.8rem;
}

.confirmation-dialog {
  width: min(28rem, calc(100% - 2rem));
  padding: 1.5rem;
  border: 1px solid #3f3f46;
  border-radius: 0.75rem;
  color: #e4e4e7;
  background: #18181b;
}

.confirmation-dialog::backdrop {
  background: rgb(0 0 0 / 0.72);
}

.confirmation-dialog h3 {
  margin: 0;
  color: #fafafa;
  font-size: 1.25rem;
}

.confirmation-dialog p {
  margin: 0.75rem 0 0;
  line-height: 1.6;
}

.dialog-label {
  margin: 0 0 0.4rem !important;
  color: #fbbf24;
  font-size: 0.75rem;
  font-weight: 800;
  text-transform: uppercase;
}

.dialog-actions {
  margin-top: 1.5rem;
}

.dialog-error {
  color: #fecaca;
}

.lock-error {
  margin-top: 0.75rem !important;
  color: #fecaca;
}
</style>
