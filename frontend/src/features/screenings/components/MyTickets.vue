<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import QRCode from 'qrcode'

import type { MyTicket } from '../types'
import TicketCard from './TicketCard.vue'
import TicketDialog from './TicketDialog.vue'

const props = defineProps<{
  tickets: MyTicket[]
  signedIn: boolean
  loading: boolean
  error: string
}>()

const qrCodes = ref<Record<string, string>>({})
const selectedTicket = ref<MyTicket | null>(null)
const ticketDialog = ref<InstanceType<typeof TicketDialog> | null>(null)

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

async function openTicket(ticket: MyTicket) {
  selectedTicket.value = ticket
  await nextTick()
  ticketDialog.value?.open()
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
        <TicketCard
          v-for="ticket in tickets"
          :key="ticket.id"
          :ticket="ticket"
          @open="openTicket"
        />
      </div>
    </div>

    <TicketDialog
      ref="ticketDialog"
      :ticket="selectedTicket"
      :qr-code="selectedTicket ? qrCodes[selectedTicket.id] : undefined"
      @close="selectedTicket = null"
    />
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
}
</style>
