<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'

import { useAdminStore } from '@/features/admin/store'

const store = useAdminStore()
const {
  bookings,
  bookingMeta,
  bookingFilters,
  isLoadingBookings,
  bookingsError,
  auditLogs,
  auditMeta,
  auditFilters,
  isLoadingAudits,
  auditsError,
} = storeToRefs(store)

const dateFormatter = new Intl.DateTimeFormat('en-GB', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '—' : dateFormatter.format(date)
}

onMounted(store.loadAll)
</script>

<template>
  <main class="admin-shell">
    <header class="admin-header">
      <div>
        <p class="section-label">Cinema operations</p>
        <h1>Admin dashboard</h1>
        <p class="header-summary">
          Review confirmed bookings and the append-only audit trail. Filters run on the server.
        </p>
      </div>
      <a class="back-link" href="/">Back to seat booking</a>
    </header>

    <section class="admin-panel" aria-labelledby="bookings-title">
      <div class="panel-heading">
        <div>
          <p class="section-label">Bookings</p>
          <h2 id="bookings-title">All booking records</h2>
        </div>
        <span class="record-count">{{ bookingMeta.total }} records</span>
      </div>

      <form class="filter-form" @submit.prevent="store.applyBookingFilters">
        <label>
          Movie title
          <input
            v-model="bookingFilters.movie"
            name="movie"
            maxlength="120"
            placeholder="Midnight Signal"
          />
        </label>
        <label>
          Status
          <select v-model="bookingFilters.status" name="status">
            <option value="">All statuses</option>
            <option value="BOOKED">Booked</option>
            <option value="HOLDING">Holding</option>
            <option value="TIMED_OUT">Timed out</option>
            <option value="CANCELLED">Cancelled</option>
          </select>
        </label>
        <div class="filter-actions">
          <button type="submit" :disabled="isLoadingBookings">Apply filters</button>
          <button
            type="button"
            class="secondary-button"
            :disabled="isLoadingBookings"
            @click="store.clearBookingFilters"
          >
            Clear
          </button>
        </div>
      </form>

      <div v-if="isLoadingBookings" class="loading-table" role="status" aria-label="Loading bookings">
        <span v-for="index in 4" :key="index"></span>
      </div>

      <div v-else-if="bookingsError" class="inline-error" role="alert">
        <p>{{ bookingsError }}</p>
        <button type="button" @click="store.loadBookings">Try again</button>
      </div>

      <div v-else-if="bookings.length === 0" class="empty-state">
        <p>No bookings match these filters.</p>
        <button type="button" @click="store.clearBookingFilters">Clear filters</button>
      </div>

      <div v-else class="table-scroll">
        <table>
          <caption class="sr-only">
            Cinema bookings
          </caption>
          <thead>
            <tr>
              <th scope="col">Movie and showtime</th>
              <th scope="col">Customer</th>
              <th scope="col">Seat</th>
              <th scope="col">Status</th>
              <th scope="col">Booked at</th>
              <th scope="col">Reference</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="booking in bookings" :key="booking.id">
              <td>
                <strong>{{ booking.screening.movie_title || 'Unknown movie' }}</strong>
                <span>
                  {{ booking.screening.auditorium_name || 'Unknown hall' }} ·
                  {{ formatDate(booking.screening.starts_at) }}
                </span>
              </td>
              <td>
                <strong>{{ booking.user.name || 'Unknown user' }}</strong>
                <span>{{ booking.user.email || booking.user.id }}</span>
              </td>
              <td class="data-value">{{ booking.seat_id }}</td>
              <td><span class="status-badge">{{ booking.status }}</span></td>
              <td class="data-value">{{ formatDate(booking.created_at) }}</td>
              <td class="reference data-value">{{ booking.id }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <nav v-if="bookingMeta.total_pages > 1" class="pagination" aria-label="Booking pages">
        <button
          type="button"
          class="secondary-button"
          :disabled="bookingMeta.page <= 1 || isLoadingBookings"
          @click="store.setBookingPage(bookingMeta.page - 1)"
        >
          Previous
        </button>
        <span class="data-value">Page {{ bookingMeta.page }} of {{ bookingMeta.total_pages }}</span>
        <button
          type="button"
          class="secondary-button"
          :disabled="bookingMeta.page >= bookingMeta.total_pages || isLoadingBookings"
          @click="store.setBookingPage(bookingMeta.page + 1)"
        >
          Next
        </button>
      </nav>
    </section>

    <section class="admin-panel" aria-labelledby="audit-title">
      <div class="panel-heading">
        <div>
          <p class="section-label">Audit trail</p>
          <h2 id="audit-title">Important system events</h2>
        </div>
        <span class="record-count">{{ auditMeta.total }} events</span>
      </div>

      <form class="filter-form audit-filter" @submit.prevent="store.applyAuditFilter">
        <label>
          Event type
          <select v-model="auditFilters.event" name="event">
            <option value="">All events</option>
            <option value="BOOKING_SUCCESS">Booking success</option>
            <option value="BOOKING_TIMEOUT">Booking timeout</option>
            <option value="SEAT_RELEASED">Seat released</option>
            <option value="SYSTEM_ERROR">System error</option>
          </select>
        </label>
        <button type="submit" :disabled="isLoadingAudits">Apply filter</button>
      </form>

      <div v-if="isLoadingAudits" class="loading-table" role="status" aria-label="Loading audit logs">
        <span v-for="index in 3" :key="index"></span>
      </div>

      <div v-else-if="auditsError" class="inline-error" role="alert">
        <p>{{ auditsError }}</p>
        <button type="button" @click="store.loadAuditLogs">Try again</button>
      </div>

      <div v-else-if="auditLogs.length === 0" class="empty-state">
        <p>No audit events match this filter.</p>
        <button type="button" @click="store.clearAuditFilter">
          Show all events
        </button>
      </div>

      <div v-else class="table-scroll">
        <table>
          <caption class="sr-only">
            System audit events
          </caption>
          <thead>
            <tr>
              <th scope="col">Event</th>
              <th scope="col">Context</th>
              <th scope="col">Time</th>
              <th scope="col">Event ID</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="entry in auditLogs" :key="entry.id">
              <td><span class="status-badge">{{ entry.event }}</span></td>
              <td>
                <strong v-if="entry.seat_id">Seat {{ entry.seat_id }}</strong>
                <span v-if="entry.booking_id">Booking {{ entry.booking_id }}</span>
                <span v-else-if="entry.screening_id">Screening {{ entry.screening_id }}</span>
                <span v-if="entry.message">{{ entry.message }}</span>
              </td>
              <td class="data-value">{{ formatDate(entry.created_at) }}</td>
              <td class="reference data-value">{{ entry.id }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <nav v-if="auditMeta.total_pages > 1" class="pagination" aria-label="Audit log pages">
        <button
          type="button"
          class="secondary-button"
          :disabled="auditMeta.page <= 1 || isLoadingAudits"
          @click="store.setAuditPage(auditMeta.page - 1)"
        >
          Previous
        </button>
        <span class="data-value">Page {{ auditMeta.page }} of {{ auditMeta.total_pages }}</span>
        <button
          type="button"
          class="secondary-button"
          :disabled="auditMeta.page >= auditMeta.total_pages || isLoadingAudits"
          @click="store.setAuditPage(auditMeta.page + 1)"
        >
          Next
        </button>
      </nav>
    </section>
  </main>
</template>

<style scoped>
.admin-shell {
  width: min(90rem, 100%);
  min-height: 100vh;
  margin: 0 auto;
  padding: 3rem 2rem 4rem;
}

.admin-header,
.panel-heading,
.pagination {
  display: flex;
  justify-content: space-between;
  gap: 1.5rem;
  align-items: center;
}

.admin-header {
  align-items: end;
  margin-bottom: 2rem;
}

.section-label {
  margin: 0 0 0.55rem;
  color: #fbbf24;
  font-size: 0.76rem;
  font-weight: 800;
  text-transform: uppercase;
}

h1,
h2 {
  margin: 0;
  color: #fafafa;
  text-wrap: balance;
}

h1 {
  font-size: clamp(2.25rem, 5vw, 4rem);
  line-height: 1;
}

h2 {
  font-size: 1.35rem;
}

.header-summary {
  max-width: 44rem;
  margin: 1rem 0 0;
  color: #a1a1aa;
  line-height: 1.65;
  text-wrap: pretty;
}

.back-link {
  flex: 0 0 auto;
  color: #fbbf24;
  font-weight: 700;
}

.admin-panel {
  margin-top: 1.25rem;
  padding: 1.5rem;
  border: 1px solid #27272a;
  border-radius: 1rem;
  background: #111113;
}

.record-count {
  color: #a1a1aa;
  font-size: 0.82rem;
  font-variant-numeric: tabular-nums;
}

.filter-form {
  display: grid;
  grid-template-columns: minmax(14rem, 1fr) minmax(11rem, 0.45fr) auto;
  gap: 1rem;
  align-items: end;
  margin: 1.5rem 0;
  padding: 1rem;
  border: 1px solid #27272a;
  border-radius: 0.75rem;
  background: #18181b;
}

.filter-form.audit-filter {
  grid-template-columns: minmax(14rem, 22rem) auto;
  justify-content: start;
}

label {
  display: grid;
  gap: 0.45rem;
  color: #d4d4d8;
  font-size: 0.82rem;
  font-weight: 700;
}

input,
select {
  width: 100%;
  min-height: 2.65rem;
  padding: 0.65rem 0.75rem;
  border: 1px solid #52525b;
  border-radius: 0.5rem;
  color: #fafafa;
  background: #09090b;
}

.filter-actions {
  display: flex;
  gap: 0.6rem;
}

button {
  min-height: 2.65rem;
  padding: 0.65rem 0.9rem;
  border: 1px solid #fbbf24;
  border-radius: 0.5rem;
  color: #18181b;
  background: #fbbf24;
  font-weight: 750;
  cursor: pointer;
}

button.secondary-button {
  border-color: #52525b;
  color: #e4e4e7;
  background: transparent;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.table-scroll {
  overflow-x: auto;
}

table {
  width: 100%;
  min-width: 54rem;
  border-collapse: collapse;
}

th,
td {
  padding: 0.85rem 0.75rem;
  border-bottom: 1px solid #27272a;
  text-align: left;
  vertical-align: top;
}

th {
  color: #a1a1aa;
  font-size: 0.75rem;
  font-weight: 750;
}

td {
  color: #d4d4d8;
  font-size: 0.83rem;
}

td strong,
td span {
  display: block;
}

td strong {
  color: #fafafa;
}

td span:not(.status-badge) {
  margin-top: 0.25rem;
  color: #a1a1aa;
}

.status-badge {
  display: inline-block;
  width: max-content;
  padding: 0.25rem 0.45rem;
  border: 1px solid #52525b;
  border-radius: 0.35rem;
  color: #fde68a;
  font-size: 0.72rem;
  font-weight: 750;
}

.data-value {
  font-variant-numeric: tabular-nums;
}

.reference {
  max-width: 12rem;
  overflow-wrap: anywhere;
  color: #a1a1aa;
}

.pagination {
  justify-content: flex-end;
  margin-top: 1rem;
}

.pagination span {
  color: #a1a1aa;
  font-size: 0.82rem;
}

.loading-table {
  display: grid;
  gap: 0.65rem;
  min-height: 12rem;
  align-content: center;
}

.loading-table span {
  display: block;
  height: 2.8rem;
  border-radius: 0.45rem;
  background: #27272a;
}

.inline-error,
.empty-state {
  display: grid;
  min-height: 12rem;
  place-content: center;
  justify-items: center;
  text-align: center;
}

.inline-error {
  color: #fecaca;
}

.empty-state {
  color: #a1a1aa;
}

.inline-error p,
.empty-state p {
  margin: 0 0 0.9rem;
  text-wrap: pretty;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 760px) {
  .admin-shell {
    padding: 2rem 1rem 3rem;
  }

  .admin-header,
  .panel-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .filter-form,
  .filter-form.audit-filter {
    grid-template-columns: 1fr;
  }

  .filter-actions,
  .filter-actions button {
    width: 100%;
  }
}
</style>
