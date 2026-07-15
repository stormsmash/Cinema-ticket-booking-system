<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import { storeToRefs } from 'pinia'

import AuthStatus from '@/features/auth/components/AuthStatus.vue'
import { useAuthStore } from '@/features/auth/store'
import ScreeningPicker from '@/features/screenings/components/ScreeningPicker.vue'
import SeatGrid from '@/features/screenings/components/SeatGrid.vue'
import SeatLockStatus from '@/features/screenings/components/SeatLockStatus.vue'
import { useScreeningStore } from '@/features/screenings/store'
import SystemStatus from '@/features/system/components/SystemStatus.vue'

const store = useScreeningStore()
const authStore = useAuthStore()
const { user } = storeToRefs(authStore)
const {
  screenings,
  selectedScreeningID,
  seatMap,
  screeningsError,
  seatsError,
  isLoadingScreenings,
  isLoadingSeats,
  activeLock,
  isUpdatingLock,
  lockError,
} = storeToRefs(store)

onMounted(store.loadScreenings)
onBeforeUnmount(store.stopRealtime)
</script>

<template>
  <main class="home-shell">
    <header class="page-header">
      <div>
        <p class="eyebrow">Cinema Ticket Booking</p>
        <h1>Pick a showtime and a seat.</h1>
        <p class="summary">
          Choose an available seat and Redis will hold it for 10 minutes while you finish the
          booking.
        </p>
      </div>

      <aside class="header-cards">
        <AuthStatus />
        <SystemStatus />
      </aside>
    </header>

    <section class="booking-layout" aria-labelledby="booking-title">
      <aside class="showtime-panel">
        <div>
          <p class="step-label">Step 1</p>
          <h2 id="booking-title">Choose a screening</h2>
        </div>

        <ScreeningPicker
          :screenings="screenings"
          :model-value="selectedScreeningID"
          :disabled="isLoadingScreenings"
          @update:model-value="store.selectScreening"
        />

        <p v-if="isLoadingScreenings" class="muted" role="status">Loading showtimes...</p>
        <div v-else-if="screeningsError" class="inline-error" role="alert">
          <p>{{ screeningsError }}</p>
          <button type="button" @click="store.loadScreenings">Try again</button>
        </div>
        <p v-else-if="screenings.length === 0" class="muted">No upcoming screenings found.</p>

        <dl v-if="seatMap" class="screening-details">
          <div>
            <dt>Movie</dt>
            <dd>{{ seatMap.movie.title }}</dd>
          </div>
          <div>
            <dt>Runtime</dt>
            <dd>{{ seatMap.movie.duration_minutes }} min</dd>
          </div>
          <div>
            <dt>Auditorium</dt>
            <dd>{{ seatMap.auditorium.name }}</dd>
          </div>
        </dl>
      </aside>

      <section class="seat-panel" aria-labelledby="seats-title">
        <div class="panel-heading">
          <div>
            <p class="step-label">Step 2</p>
            <h2 id="seats-title">Choose a seat</h2>
          </div>
          <span v-if="seatMap" class="seat-count">{{ seatMap.seats.length }} seats</span>
        </div>

        <div v-if="isLoadingSeats" class="seat-skeleton" role="status">
          <span>Loading seat map...</span>
          <i v-for="index in 30" :key="index"></i>
        </div>

        <div v-else-if="seatsError" class="seat-error" role="alert">
          <p>{{ seatsError }}</p>
          <button type="button" @click="store.reloadSeatMap">Try again</button>
        </div>

        <template v-else-if="seatMap">
          <SeatGrid
            :seat-map="seatMap"
            :can-lock="Boolean(user)"
            :active-seat-id="activeLock?.seat_id"
            :is-updating-lock="isUpdatingLock"
            @lock="store.lockSeat"
          />
          <SeatLockStatus
            :lock="activeLock"
            :signed-in="Boolean(user)"
            :is-updating="isUpdatingLock"
            :error="lockError"
            @release="store.unlockSeat"
            @expired="store.handleLockExpired"
          />
        </template>

        <p v-else class="empty-state">Choose a screening to see its available seats.</p>
      </section>
    </section>
  </main>
</template>

<style scoped>
.home-shell {
  width: min(76rem, 100%);
  min-height: 100vh;
  margin: 0 auto;
  padding: 3rem 2rem 4rem;
}

.page-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(17rem, 22rem);
  gap: 3rem;
  align-items: end;
}

.header-cards {
  display: grid;
  gap: 1rem;
}

.eyebrow,
.step-label {
  margin: 0 0 0.65rem;
  color: #fbbf24;
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

h1,
h2 {
  margin: 0;
  color: #fafafa;
}

h1 {
  max-width: 17ch;
  font-size: clamp(2.5rem, 6vw, 5rem);
  line-height: 1;
  letter-spacing: -0.055em;
}

h2 {
  font-size: 1.25rem;
  letter-spacing: -0.02em;
}

.summary {
  max-width: 42rem;
  margin: 1.25rem 0 0;
  color: #a1a1aa;
  line-height: 1.7;
}

.booking-layout {
  display: grid;
  grid-template-columns: minmax(15rem, 20rem) minmax(0, 1fr);
  margin-top: 3rem;
  border: 1px solid #27272a;
  border-radius: 1rem;
  overflow: hidden;
  background: #0f0f11;
}

.showtime-panel,
.seat-panel {
  padding: 1.5rem;
}

.showtime-panel {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  border-right: 1px solid #27272a;
  background: #18181b;
}

.panel-heading {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  margin-bottom: 2.5rem;
}

.seat-count {
  padding: 0.35rem 0.6rem;
  border: 1px solid #3f3f46;
  border-radius: 0.4rem;
  color: #a1a1aa;
  font-size: 0.78rem;
  font-weight: 700;
}

.muted,
.empty-state {
  margin: 0;
  color: #71717a;
  line-height: 1.6;
}

.empty-state {
  display: grid;
  min-height: 20rem;
  place-items: center;
  text-align: center;
}

.inline-error,
.seat-error {
  color: #fecaca;
}

.inline-error p,
.seat-error p {
  margin: 0 0 0.75rem;
}

.inline-error button,
.seat-error button {
  padding: 0.55rem 0.75rem;
  border: 0;
  border-radius: 0.45rem;
  color: #18181b;
  background: #fbbf24;
  font-weight: 750;
  cursor: pointer;
}

.seat-error {
  display: grid;
  min-height: 20rem;
  place-content: center;
  justify-items: center;
  text-align: center;
}

.screening-details {
  display: grid;
  gap: 0.9rem;
  margin: 0;
  padding-top: 1.25rem;
  border-top: 1px solid #3f3f46;
}

.screening-details div {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
}

.screening-details dt {
  color: #71717a;
  font-size: 0.8rem;
}

.screening-details dd {
  margin: 0;
  color: #e4e4e7;
  font-size: 0.85rem;
  font-weight: 650;
  text-align: right;
}

.seat-skeleton {
  display: grid;
  grid-template-columns: repeat(10, minmax(2.25rem, 1fr));
  gap: 0.5rem;
  min-height: 20rem;
  align-content: center;
}

.seat-skeleton span {
  grid-column: 1 / -1;
  margin-bottom: 1rem;
  color: #71717a;
  text-align: center;
}

.seat-skeleton i {
  height: 2.25rem;
  border-radius: 0.45rem;
  background: #27272a;
}

@media (max-width: 820px) {
  .page-header,
  .booking-layout {
    grid-template-columns: 1fr;
  }

  .showtime-panel {
    border-right: 0;
    border-bottom: 1px solid #27272a;
  }
}

@media (max-width: 520px) {
  .home-shell {
    padding: 2rem 1rem 3rem;
  }

  .page-header {
    gap: 2rem;
  }

  .booking-layout {
    margin-top: 2rem;
  }
}
</style>
