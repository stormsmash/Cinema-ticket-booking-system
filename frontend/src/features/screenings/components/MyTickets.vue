<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import QRCode from 'qrcode'

import { getMoviePresentation } from '@/features/movies/catalog'
import type { MyTicket } from '../types'

const props = defineProps<{
  tickets: MyTicket[]
  signedIn: boolean
  loading: boolean
  error: string
}>()

const qrCodes = ref<Record<string, string>>({})
const selectedTicket = ref<MyTicket | null>(null)
const ticketDialog = ref<HTMLDialogElement | null>(null)
const copiedTicketID = ref('')

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
const priceFormatter = new Intl.NumberFormat('th-TH')

watch(
  () => props.tickets,
  async (tickets) => {
    const entries = await Promise.all(
      tickets.map(async (ticket) => [
        ticket.id,
        await QRCode.toDataURL(ticket.ticket_code, {
          width: 220,
          margin: 1,
          color: { dark: '#111113', light: '#ffffff' },
          errorCorrectionLevel: 'M',
        }),
      ]),
    )
    qrCodes.value = Object.fromEntries(entries)
  },
  { immediate: true },
)

function formatDate(value: string) {
  return dateFormatter.format(new Date(value))
}

function formatTime(value: string) {
  return timeFormatter.format(new Date(value))
}

async function openTicket(ticket: MyTicket) {
  selectedTicket.value = ticket
  copiedTicketID.value = ''
  await nextTick()
  const dialog = ticketDialog.value
  if (!dialog) return
  if (typeof dialog.showModal === 'function') dialog.showModal()
  else dialog.setAttribute('open', '')
}

function closeTicket() {
  const dialog = ticketDialog.value
  if (!dialog?.open) return
  if (typeof dialog.close === 'function') dialog.close()
  else dialog.removeAttribute('open')
}

async function copyTicketCode(ticket: MyTicket) {
  try {
    await navigator.clipboard.writeText(ticket.ticket_code)
    copiedTicketID.value = ticket.id
  } catch {
    copiedTicketID.value = ''
  }
}
</script>

<template>
  <section id="my-tickets" class="tickets-section" aria-labelledby="tickets-title">
    <div class="content-shell">
      <div class="tickets-heading">
        <div>
          <p>MY TICKETS</p>
          <h2 id="tickets-title">ตั๋วของฉัน</h2>
          <span>เปิด QR บนโทรศัพท์เพื่อแสดงให้พนักงานตรวจตั๋วก่อนเข้าโรง</span>
        </div>
        <strong v-if="signedIn && tickets.length">{{ tickets.length }} ใบ</strong>
      </div>

      <div v-if="!signedIn" class="ticket-message">
        <strong>เข้าสู่ระบบเพื่อดูตั๋วของคุณ</strong>
        <p>ตั๋วที่ซื้อแล้วจะบันทึกไว้กับบัญชี Google และเปิดดูได้จากหน้านี้</p>
      </div>

      <div v-else-if="loading" class="ticket-message" role="status">กำลังโหลดตั๋ว...</div>

      <div v-else-if="error" class="ticket-message ticket-message--error" role="alert">
        {{ error }}
      </div>

      <div v-else-if="tickets.length === 0" class="ticket-message">
        <strong>ยังไม่มีตั๋ว</strong>
        <p>เลือกหนัง รอบฉาย และที่นั่งด้านบนเพื่อเริ่มจอง</p>
      </div>

      <div v-else class="ticket-grid">
        <article v-for="ticket in tickets" :key="ticket.id" class="ticket-card">
          <img
            class="ticket-poster"
            :src="getMoviePresentation(ticket.movie_title).poster"
            :alt="`โปสเตอร์ ${ticket.movie_title}`"
          />
          <div class="ticket-content">
            <span class="ticket-status">จองแล้ว</span>
            <h3>{{ ticket.movie_title }}</h3>
            <dl>
              <div>
                <dt>วันฉาย</dt>
                <dd>{{ formatDate(ticket.starts_at) }}</dd>
              </div>
              <div>
                <dt>เวลา</dt>
                <dd>{{ formatTime(ticket.starts_at) }} น.</dd>
              </div>
              <div>
                <dt>โรง</dt>
                <dd>{{ ticket.auditorium_name }}</dd>
              </div>
              <div>
                <dt>ที่นั่ง</dt>
                <dd>{{ ticket.seat_id }}</dd>
              </div>
            </dl>
            <div class="ticket-price">
              <span>ราคาสุทธิ</span>
              <strong>฿{{ priceFormatter.format(ticket.price_baht) }}</strong>
            </div>
            <button type="button" @click="openTicket(ticket)">เปิดตั๋วและ QR</button>
          </div>
        </article>
      </div>
    </div>

    <dialog ref="ticketDialog" class="ticket-dialog" @cancel="closeTicket">
      <template v-if="selectedTicket">
        <div class="dialog-header">
          <span>LUMINA E-TICKET</span>
          <button type="button" aria-label="ปิดตั๋ว" @click="closeTicket">×</button>
        </div>
        <div class="dialog-ticket">
          <p>ภาพยนตร์</p>
          <h3>{{ selectedTicket.movie_title }}</h3>
          <div class="dialog-meta">
            <span>{{ formatDate(selectedTicket.starts_at) }}</span>
            <span>{{ formatTime(selectedTicket.starts_at) }} น.</span>
            <span>{{ selectedTicket.auditorium_name }}</span>
          </div>
          <div class="dialog-seat">
            <span>SEAT</span>
            <strong>{{ selectedTicket.seat_id }}</strong>
          </div>
          <img
            v-if="qrCodes[selectedTicket.id]"
            class="ticket-qr"
            :src="qrCodes[selectedTicket.id]"
            :alt="`QR ตั๋วที่นั่ง ${selectedTicket.seat_id}`"
          />
          <code>{{ selectedTicket.ticket_code }}</code>
          <button class="copy-button" type="button" @click="copyTicketCode(selectedTicket)">
            {{ copiedTicketID === selectedTicket.id ? 'คัดลอกแล้ว' : 'คัดลอกรหัสตั๋ว' }}
          </button>
          <small>ตั๋วตัวอย่างสำหรับระบบ Demo</small>
        </div>
      </template>
    </dialog>
  </section>
</template>

<style scoped>
.tickets-section {
  padding: 4.5rem 0 5rem;
  background: #ececee;
  scroll-margin-top: 4.5rem;
}

.content-shell {
  width: min(78rem, calc(100% - 3rem));
  margin: 0 auto;
}

.tickets-heading {
  display: flex;
  justify-content: space-between;
  gap: 2rem;
  align-items: end;
}

.tickets-heading p {
  margin: 0 0 0.45rem;
  color: #d91920;
  font-size: 0.7rem;
  font-weight: 850;
  letter-spacing: 0.1em;
}

.tickets-heading h2 {
  margin: 0;
  color: #222226;
  font-size: clamp(1.9rem, 4vw, 2.7rem);
}

.tickets-heading span {
  display: block;
  margin-top: 0.65rem;
  color: #696970;
  font-size: 0.82rem;
}

.tickets-heading > strong {
  color: #77777e;
  font-size: 0.8rem;
}

.ticket-message {
  display: grid;
  min-height: 9rem;
  margin-top: 1.5rem;
  padding: 1.5rem;
  place-content: center;
  text-align: center;
  background: #fff;
}

.ticket-message p {
  margin: 0.35rem 0 0;
  color: #77777e;
  font-size: 0.75rem;
}

.ticket-message--error {
  color: #a31e24;
}

.ticket-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  margin-top: 1.5rem;
}

.ticket-card {
  display: grid;
  grid-template-columns: 9rem minmax(0, 1fr);
  overflow: hidden;
  border: 1px solid #d7d7da;
  border-left: 0.28rem solid #d91920;
  background: #fff;
}

.ticket-poster {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.ticket-content {
  min-width: 0;
  padding: 1rem;
}

.ticket-status {
  color: #28704a;
  font-size: 0.65rem;
  font-weight: 800;
}

.ticket-content h3 {
  margin: 0.25rem 0 0;
  color: #222226;
  font-size: 1.15rem;
}

.ticket-content dl {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.7rem;
  margin: 1rem 0 0;
}

.ticket-content dl div {
  min-width: 0;
}

.ticket-content dt {
  color: #8b8b92;
  font-size: 0.6rem;
}

.ticket-content dd {
  margin: 0.18rem 0 0;
  overflow: hidden;
  color: #343439;
  font-size: 0.7rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ticket-price {
  display: flex;
  justify-content: space-between;
  margin-top: 0.9rem;
  padding-top: 0.75rem;
  border-top: 1px dashed #d2d2d6;
  align-items: center;
}

.ticket-price span {
  color: #77777e;
  font-size: 0.67rem;
}

.ticket-price strong {
  color: #d91920;
  font-size: 1rem;
}

.ticket-content button,
.copy-button {
  width: 100%;
  min-height: 2.5rem;
  margin-top: 0.8rem;
  border: 1px solid #d91920;
  color: #fff;
  background: #d91920;
  font-size: 0.7rem;
  font-weight: 800;
  cursor: pointer;
}

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
  max-width: 13rem;
}

.dialog-ticket small {
  margin-top: 0.75rem;
  color: #9999a0;
  font-size: 0.58rem;
}

@media (max-width: 780px) {
  .content-shell {
    width: min(100% - 2rem, 78rem);
  }

  .ticket-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .content-shell {
    width: min(100% - 1.25rem, 78rem);
  }

  .tickets-section {
    padding: 3.5rem 0;
  }

  .ticket-card {
    grid-template-columns: 6.5rem minmax(0, 1fr);
  }

  .ticket-content {
    padding: 0.8rem;
  }

  .ticket-content dl {
    grid-template-columns: 1fr;
    gap: 0.4rem;
  }
}
</style>
