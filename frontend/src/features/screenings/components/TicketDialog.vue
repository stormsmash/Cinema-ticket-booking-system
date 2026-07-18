<script setup lang="ts">
import { ref, watch } from 'vue'
import type { MyTicket } from '../types'

const props = defineProps<{
  ticket: MyTicket | null
  qrCode?: string
}>()
const emit = defineEmits<{ close: [] }>()

const dialog = ref<HTMLDialogElement | null>(null)
const copied = ref(false)

const dateFormatter = new Intl.DateTimeFormat('th-TH', {
  weekday: 'short',
  day: 'numeric',
  month: 'short',
  year: 'numeric',
  timeZone: 'Asia/Bangkok',
})
const timeFormatter = new Intl.DateTimeFormat('th-TH', {
  hour: '2-digit',
  minute: '2-digit',
  timeZone: 'Asia/Bangkok',
})

watch(
  () => props.ticket?.id,
  () => {
    copied.value = false
  },
)

function open() {
  if (!dialog.value) return
  if (typeof dialog.value.showModal === 'function') dialog.value.showModal()
  else dialog.value.setAttribute('open', '')
}

function close() {
  if (!dialog.value?.open) return
  if (typeof dialog.value.close === 'function') dialog.value.close()
  else {
    dialog.value.removeAttribute('open')
    emit('close')
  }
}

async function copyTicketCode() {
  if (!props.ticket) return
  try {
    await navigator.clipboard.writeText(props.ticket.ticket_code)
    copied.value = true
  } catch {
    copied.value = false
  }
}

defineExpose({ open, close })
</script>

<template>
  <dialog ref="dialog" class="ticket-dialog" @cancel="emit('close')" @close="emit('close')">
    <template v-if="ticket">
      <div class="dialog-header">
        <span>E-TICKET</span>
        <button type="button" aria-label="ปิดตั๋ว" @click="close">×</button>
      </div>
      <div class="dialog-ticket">
        <p>ภาพยนตร์</p>
        <h3>{{ ticket.movie_title }}</h3>
        <div class="dialog-meta">
          <span>{{ dateFormatter.format(new Date(ticket.starts_at)) }}</span>
          <span>{{ timeFormatter.format(new Date(ticket.starts_at)) }} น.</span>
          <span>{{ ticket.auditorium_name }}</span>
        </div>
        <div class="dialog-seat">
          <span>SEAT</span>
          <strong>{{ ticket.seat_id }}</strong>
        </div>
        <img
          v-if="qrCode"
          class="ticket-qr"
          :src="qrCode"
          :alt="`QR ตั๋วที่นั่ง ${ticket.seat_id}`"
        />
        <code>{{ ticket.ticket_code }}</code>
        <button class="copy-button" type="button" @click="copyTicketCode">
          {{ copied ? 'คัดลอกแล้ว' : 'คัดลอกรหัสตั๋ว' }}
        </button>
        <small>ตั๋วตัวอย่างสำหรับระบบ Demo</small>
      </div>
    </template>
  </dialog>
</template>

<style scoped>
.ticket-dialog {
  width: min(25rem, calc(100% - 2rem));
  padding: 0;
  border: 0;
  color: #222226;
  background: #fff;
  box-shadow: 0 2rem 6rem rgb(0 0 0 / 50%);
}

.ticket-dialog::backdrop {
  background: rgb(0 0 0 / 75%);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  padding: 0.8rem 1rem;
  align-items: center;
  color: #fff;
  background: #111113;
  font-size: 0.68rem;
  font-weight: 850;
  letter-spacing: 0.1em;
}

.dialog-header button {
  padding: 0;
  border: 0;
  color: #fff;
  background: transparent;
  font-size: 1.5rem;
  cursor: pointer;
}

.dialog-ticket {
  display: grid;
  justify-items: center;
  padding: 1.5rem;
  text-align: center;
}

.dialog-ticket > p {
  margin: 0;
  color: #888890;
  font-size: 0.64rem;
}

.dialog-ticket h3 {
  margin: 0.3rem 0 0;
  font-size: 1.4rem;
}

.dialog-meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.35rem 0.9rem;
  margin-top: 0.55rem;
  color: #66666d;
  font-size: 0.68rem;
}

.dialog-seat {
  display: grid;
  margin-top: 1rem;
}

.dialog-seat span {
  color: #888890;
  font-size: 0.6rem;
  letter-spacing: 0.15em;
}

.dialog-seat strong {
  color: #d91920;
  font-size: 2.4rem;
}

.ticket-qr {
  width: 11rem;
  height: 11rem;
  margin-top: 0.5rem;
}

.dialog-ticket code {
  margin-top: 0.5rem;
  overflow-wrap: anywhere;
  color: #55555c;
  font-size: 0.62rem;
}

.copy-button {
  width: 100%;
  max-width: 13rem;
  min-height: 2.5rem;
  margin-top: 0.8rem;
  border: 1px solid #d91920;
  color: #fff;
  background: #d91920;
  font-size: 0.7rem;
  font-weight: 800;
  cursor: pointer;
}

.dialog-ticket small {
  margin-top: 0.75rem;
  color: #9999a0;
  font-size: 0.58rem;
}
</style>
