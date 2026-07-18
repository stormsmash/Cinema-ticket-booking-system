<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'

import { useAuthStore } from '@/features/auth/store'
import { getMoviePresentation } from '@/features/movies/catalog'
import ScreeningPicker from '@/features/screenings/components/ScreeningPicker.vue'
import SeatGrid from '@/features/screenings/components/SeatGrid.vue'
import SeatLockStatus from '@/features/screenings/components/SeatLockStatus.vue'
import { useScreeningStore } from '@/features/screenings/store'

const store = useScreeningStore()
const authStore = useAuthStore()
const { user } = storeToRefs(authStore)
const {
  screenings,
  selectedScreeningID,
  seatMap,
  seatsError,
  isLoadingScreenings,
  isLoadingSeats,
  activeLocks,
  isUpdatingLock,
  lockError,
  confirmedBookings,
  isConfirmingBooking,
  bookingError,
} = storeToRefs(store)

const selectedScreening = computed(
  () => screenings.value.find((screening) => screening.id === selectedScreeningID.value) ?? null,
)
const selectedMovie = computed(() =>
  getMoviePresentation(selectedScreening.value?.movie.title ?? ''),
)
const availableSeatCount = computed(
  () => seatMap.value?.seats.filter((seat) => seat.status === 'AVAILABLE').length ?? 0,
)
const activeSeatIDs = computed(() => activeLocks.value.map((lock) => lock.seat_id))
const selectedTicketPrice = computed(
  () => selectedScreening.value?.ticket_price_baht ?? seatMap.value?.ticket_price_baht ?? 0,
)

const dateFormatter = new Intl.DateTimeFormat('th-TH', {
  weekday: 'long',
  day: 'numeric',
  month: 'long',
  timeZone: 'Asia/Bangkok',
})
const timeFormatter = new Intl.DateTimeFormat('th-TH', {
  hour: '2-digit',
  minute: '2-digit',
  timeZone: 'Asia/Bangkok',
})
</script>

<template>
  <section id="booking" class="booking-section" aria-labelledby="booking-title">
    <div class="content-shell">
      <div class="booking-heading">
        <div>
          <p class="eyebrow eyebrow--red">BOOKING</p>
          <h2 id="booking-title">เลือกรอบและที่นั่ง</h2>
        </div>
        <ol class="booking-steps" aria-label="ขั้นตอนการจอง">
          <li class="done"><span>1</span> เลือกหนัง</li>
          <li class="active"><span>2</span> เลือกที่นั่ง</li>
          <li><span>3</span> ยืนยัน</li>
        </ol>
      </div>

      <div class="booking-workspace">
        <aside class="booking-summary">
          <div v-if="selectedScreening" class="selected-film">
            <img :src="selectedMovie.poster" :alt="`โปสเตอร์ ${selectedMovie.title}`" />
            <div>
              <span>{{ selectedMovie.genres.join(' · ') }}</span>
              <h3>{{ selectedMovie.title }}</h3>
              <p>{{ selectedMovie.englishTitle }}</p>
            </div>
          </div>

          <ScreeningPicker
            :screenings="screenings"
            :model-value="selectedScreeningID"
            :disabled="isLoadingScreenings"
            @update:model-value="store.selectScreening"
          />

          <dl v-if="selectedScreening" class="ticket-details">
            <div>
              <dt>วันที่</dt>
              <dd>{{ dateFormatter.format(new Date(selectedScreening.starts_at)) }}</dd>
            </div>
            <div>
              <dt>เวลา</dt>
              <dd>{{ timeFormatter.format(new Date(selectedScreening.starts_at)) }} น.</dd>
            </div>
            <div>
              <dt>โรงภาพยนตร์</dt>
              <dd>{{ selectedScreening.auditorium.name }}</dd>
            </div>
            <div>
              <dt>ภาษา</dt>
              <dd>{{ selectedMovie.language }}</dd>
            </div>
            <div>
              <dt>ราคาต่อที่นั่ง</dt>
              <dd>฿{{ selectedTicketPrice.toLocaleString('th-TH') }}</dd>
            </div>
          </dl>

          <div class="seat-availability">
            <span><i></i> ที่นั่งว่าง</span>
            <strong>{{ availableSeatCount }}</strong>
          </div>

          <p class="booking-note">
            ที่นั่งจะถูกพักไว้ 5 นาทีหลังเลือก กรุณาตรวจสอบรอบฉายก่อนยืนยันการจอง
          </p>
        </aside>

        <section class="seat-panel" aria-labelledby="seats-title">
          <div class="seat-panel__heading">
            <div>
              <p>{{ selectedScreening?.auditorium.name ?? 'โรงภาพยนตร์' }}</p>
              <h3 id="seats-title">ผังที่นั่ง</h3>
            </div>
            <span v-if="seatMap">
              {{
                activeLocks.length
                  ? `เลือกแล้ว ${activeLocks.length} / 6`
                  : `${seatMap.seats.length} ที่นั่ง`
              }}
            </span>
          </div>

          <div v-if="isLoadingSeats" class="seat-skeleton" role="status">
            <span>กำลังโหลดผังที่นั่ง...</span>
            <i v-for="index in 40" :key="index"></i>
          </div>

          <div v-else-if="seatsError" class="seat-error" role="alert">
            <strong>เปิดผังที่นั่งไม่สำเร็จ</strong>
            <p>{{ seatsError }}</p>
            <button type="button" class="button button--primary" @click="store.reloadSeatMap">
              ลองอีกครั้ง
            </button>
          </div>

          <template v-else-if="seatMap">
            <SeatGrid
              :seat-map="seatMap"
              :can-lock="Boolean(user)"
              :active-seat-ids="activeSeatIDs"
              :max-selectable="store.maxSeatsPerBooking"
              :is-updating-lock="isUpdatingLock"
              @toggle="store.toggleSeatLock"
            />
            <SeatLockStatus
              :locks="activeLocks"
              :signed-in="Boolean(user)"
              :is-updating="isUpdatingLock"
              :error="lockError"
              :bookings="confirmedBookings"
              :unit-price-baht="selectedTicketPrice"
              :is-confirming="isConfirmingBooking"
              :booking-error="bookingError"
              @release-all="store.unlockAllSeats"
              @expired="store.handleLockExpired"
              @confirm="store.confirmBooking"
            />
          </template>

          <div v-else class="seat-error">
            <p>เลือกภาพยนตร์เพื่อดูผังที่นั่ง</p>
          </div>
        </section>
      </div>
    </div>
  </section>
</template>
