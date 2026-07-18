<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import type { Booking, SeatLock } from '../types'

const props = defineProps<{
  locks: SeatLock[]
  signedIn: boolean
  isUpdating: boolean
  error: string
  bookings: Booking[]
  unitPriceBaht: number
  isConfirming: boolean
  bookingError: string
}>()

const emit = defineEmits<{
  releaseAll: []
  expired: []
  confirm: []
}>()

const remainingSeconds = ref(0)
const confirmationDialog = ref<HTMLDialogElement | null>(null)
let timer: ReturnType<typeof setInterval> | undefined
let expirationReported = false

const seatLabel = computed(() => props.locks.map((lock) => lock.seat_id).join(', '))
const totalPrice = computed(() => props.unitPriceBaht * props.locks.length)
const priceFormatter = new Intl.NumberFormat('th-TH')
const timeLabel = computed(() => {
  const minutes = Math.floor(remainingSeconds.value / 60)
  const seconds = remainingSeconds.value % 60
  return `${minutes}:${seconds.toString().padStart(2, '0')}`
})

function updateRemainingTime() {
  if (!props.locks.length) {
    remainingSeconds.value = 0
    expirationReported = false
    return
  }

  const earliestExpiry = Math.min(...props.locks.map((lock) => new Date(lock.expires_at).getTime()))
  remainingSeconds.value = Math.max(0, Math.ceil((earliestExpiry - Date.now()) / 1000))

  if (remainingSeconds.value === 0 && !expirationReported) {
    expirationReported = true
    emit('expired')
  }
}

watch(
  () => props.locks.map((lock) => `${lock.seat_id}:${lock.expires_at}`).join('|'),
  () => {
    expirationReported = false
    updateRemainingTime()
  },
)

watch(
  () => props.locks.length,
  (length) => {
    if (length === 0) closeConfirmation()
  },
)

function openConfirmation() {
  const dialog = confirmationDialog.value
  if (!dialog) return
  if (typeof dialog.showModal === 'function') dialog.showModal()
  else dialog.setAttribute('open', '')
}

function closeConfirmation() {
  const dialog = confirmationDialog.value
  if (!dialog?.open) return
  if (typeof dialog.close === 'function') dialog.close()
  else dialog.removeAttribute('open')
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
    <div v-if="locks.length" class="active-lock">
      <div class="lock-summary">
        <span>ที่นั่งที่เลือก</span>
        <strong>{{ seatLabel }}</strong>
        <small>เหลือเวลา {{ timeLabel }}</small>
      </div>
      <div class="price-summary">
        <span>{{ locks.length }} ที่นั่ง × ฿{{ priceFormatter.format(unitPriceBaht) }}</span>
        <strong>฿{{ priceFormatter.format(totalPrice) }}</strong>
      </div>
      <div class="lock-actions">
        <button
          type="button"
          class="confirm-button"
          :disabled="isUpdating || isConfirming"
          @click="openConfirmation"
        >
          ยืนยันและชำระเงิน
        </button>
        <button type="button" :disabled="isUpdating || isConfirming" @click="emit('releaseAll')">
          {{ isUpdating ? 'กำลังยกเลิก...' : 'ยกเลิกทั้งหมด' }}
        </button>
      </div>
    </div>

    <div v-else-if="bookings.length" class="booking-success" role="status">
      <strong>จองสำเร็จ {{ bookings.length }} ที่นั่ง</strong>
      <span>ที่นั่ง {{ bookings.map((booking) => booking.seat_id).join(', ') }}</span>
      <a href="#my-tickets">เปิดตั๋วของฉัน</a>
    </div>

    <p v-else-if="signedIn">
      {{ isUpdating ? 'กำลังอัปเดตที่นั่ง...' : 'เลือกได้สูงสุด 6 ที่นั่งสำหรับคุณและเพื่อน' }}
    </p>
    <p v-else>เข้าสู่ระบบด้วย Google ก่อนเลือกที่นั่ง</p>

    <p v-if="error" class="lock-error" role="alert">{{ error }}</p>
    <p v-if="bookingError" class="lock-error" role="alert">{{ bookingError }}</p>

    <dialog
      ref="confirmationDialog"
      class="confirmation-dialog"
      aria-labelledby="confirmation-title"
      @cancel="closeConfirmation"
    >
      <form method="dialog" @submit.prevent>
        <p class="dialog-label">ระบบชำระเงินจำลอง</p>
        <h3 id="confirmation-title">ตรวจสอบรายการจอง</h3>
        <dl>
          <div>
            <dt>ที่นั่ง</dt>
            <dd>{{ seatLabel }}</dd>
          </div>
          <div>
            <dt>จำนวน</dt>
            <dd>{{ locks.length }} ที่นั่ง</dd>
          </div>
          <div>
            <dt>ราคาต่อใบ</dt>
            <dd>฿{{ priceFormatter.format(unitPriceBaht) }}</dd>
          </div>
          <div class="dialog-total">
            <dt>ยอดรวม</dt>
            <dd>฿{{ priceFormatter.format(totalPrice) }}</dd>
          </div>
        </dl>
        <p>กดยืนยันเพื่อจำลองการชำระเงินและออก E-Ticket แยกตามแต่ละที่นั่ง</p>
        <p v-if="bookingError" class="dialog-error" role="alert">{{ bookingError }}</p>
        <div class="dialog-actions">
          <button type="button" :disabled="isConfirming" @click="closeConfirmation">
            กลับไปตรวจสอบ
          </button>
          <button
            type="button"
            class="confirm-button"
            :disabled="isConfirming"
            @click="emit('confirm')"
          >
            {{ isConfirming ? 'กำลังออกตั๋ว...' : `ชำระ ฿${priceFormatter.format(totalPrice)}` }}
          </button>
        </div>
      </form>
    </dialog>
  </div>
</template>

<style scoped>
.lock-status {
  min-height: 3.75rem;
  margin-top: 1.4rem;
  padding: 1rem;
  border: 1px solid rgb(255 255 255 / 8%);
  border-radius: 0.25rem;
  color: #93939b;
  background: #111820;
  font-size: 0.75rem;
}

.lock-status p {
  margin: 0;
}

.active-lock {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  gap: 1rem;
  align-items: center;
}

.lock-summary,
.price-summary {
  display: grid;
}

.lock-summary strong {
  margin-top: 0.2rem;
  color: #fff;
  font-size: 1rem;
}

.lock-summary small {
  margin-top: 0.2rem;
  color: #ef4d52;
  font-size: 0.65rem;
  font-variant-numeric: tabular-nums;
}

.price-summary {
  min-width: 7rem;
  text-align: right;
}

.price-summary span {
  color: #7f7f87;
  font-size: 0.62rem;
}

.price-summary strong {
  margin-top: 0.2rem;
  color: #fff;
  font-size: 1.2rem;
}

.lock-actions,
.dialog-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  justify-content: flex-end;
}

button {
  min-height: 2.35rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid #41414a;
  border-radius: 0.25rem;
  color: #d9d9de;
  background: transparent;
  font-size: 0.68rem;
  font-weight: 750;
  cursor: pointer;
}

button:disabled {
  cursor: wait;
  opacity: 0.6;
}

.confirm-button {
  border-color: #d91920;
  color: #fff;
  background: #d91920;
}

.booking-success {
  display: grid;
  gap: 0.25rem;
  color: #cdd3dc;
}

.booking-success strong {
  color: #74d5b1;
}

.booking-success a {
  width: max-content;
  margin-top: 0.35rem;
  color: #ef4d52;
  font-weight: 800;
}

.confirmation-dialog {
  width: min(28rem, calc(100% - 2rem));
  padding: 1.5rem;
  border: 1px solid rgb(255 255 255 / 12%);
  border-radius: 0.35rem;
  color: #dce1e8;
  background: #17171b;
  box-shadow: 0 2rem 6rem rgb(0 0 0 / 50%);
}

.confirmation-dialog::backdrop {
  background: rgb(0 0 0 / 72%);
}

.confirmation-dialog h3 {
  margin: 0;
  color: #fff;
  font-size: 1.25rem;
}

.confirmation-dialog > form > p:not(.dialog-label, .dialog-error) {
  margin: 0.75rem 0 0;
  color: #8e8e96;
  line-height: 1.6;
}

.confirmation-dialog dl {
  margin: 1.2rem 0 0;
}

.confirmation-dialog dl div {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.55rem 0;
  border-bottom: 1px solid #303035;
}

.confirmation-dialog dt {
  color: #888890;
}

.confirmation-dialog dd {
  margin: 0;
  color: #fff;
  font-weight: 750;
}

.confirmation-dialog .dialog-total dd {
  color: #ef4d52;
  font-size: 1.15rem;
}

.dialog-label {
  margin: 0 0 0.4rem !important;
  color: #ef4d52;
  font-size: 0.7rem;
  font-weight: 800;
  text-transform: uppercase;
}

.dialog-actions {
  margin-top: 1.5rem;
}

.dialog-error,
.lock-error {
  margin-top: 0.75rem !important;
  color: #fecaca;
}

@media (max-width: 720px) {
  .active-lock {
    grid-template-columns: 1fr auto;
  }

  .lock-actions {
    grid-column: 1 / -1;
    justify-content: stretch;
  }

  .lock-actions button {
    flex: 1;
  }
}
</style>
