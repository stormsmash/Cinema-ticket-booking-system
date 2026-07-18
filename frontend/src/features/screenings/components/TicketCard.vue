<script setup lang="ts">
import { getMoviePresentation } from '@/features/movies/catalog'
import type { MyTicket } from '../types'

const props = defineProps<{ ticket: MyTicket }>()
const emit = defineEmits<{ open: [ticket: MyTicket] }>()

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
</script>

<template>
  <article class="ticket-card">
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
          <dd>{{ dateFormatter.format(new Date(ticket.starts_at)) }}</dd>
        </div>
        <div>
          <dt>เวลา</dt>
          <dd>{{ timeFormatter.format(new Date(ticket.starts_at)) }} น.</dd>
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
      <button type="button" @click="emit('open', props.ticket)">เปิดตั๋วและ QR</button>
    </div>
  </article>
</template>

<style scoped>
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

.ticket-content button {
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

@media (max-width: 560px) {
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
