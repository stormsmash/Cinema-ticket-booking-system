<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useAdminStore } from '@/features/admin/store'

const store = useAdminStore()
const { bookings, bookingMeta, bookingFilters, isLoadingBookings, bookingsError } =
  storeToRefs(store)

const dateFormatter = new Intl.DateTimeFormat('th-TH', {
  dateStyle: 'medium',
  timeStyle: 'short',
  timeZone: 'Asia/Bangkok',
})

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '—' : dateFormatter.format(date)
}

function statusLabel(status: string) {
  const labels: Record<string, string> = {
    BOOKED: 'จองสำเร็จ',
    HOLDING: 'กำลังล็อก',
    TIMED_OUT: 'หมดเวลา',
    CANCELLED: 'ยกเลิก',
  }
  return labels[status] ?? status
}
</script>

<template>
  <section class="admin-panel" aria-labelledby="bookings-title">
    <div class="panel-heading">
      <div>
        <p class="section-label">BOOKINGS</p>
        <h2 id="bookings-title">รายการจองทั้งหมด</h2>
      </div>
      <span class="record-count">{{ bookingMeta.total }} รายการ</span>
    </div>

    <form class="filter-form" @submit.prevent="store.applyBookingFilters">
      <label>
        ชื่อภาพยนตร์
        <span class="input-wrap">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <circle cx="11" cy="11" r="7" />
            <path d="m20 20-4-4" />
          </svg>
          <input
            v-model="bookingFilters.movie"
            name="movie"
            maxlength="120"
            placeholder="เช่น หลานม่า"
          />
        </span>
      </label>
      <label>
        สถานะ
        <select v-model="bookingFilters.status" name="status">
          <option value="">ทุกสถานะ</option>
          <option value="BOOKED">จองสำเร็จ</option>
          <option value="HOLDING">กำลังล็อก</option>
          <option value="TIMED_OUT">หมดเวลา</option>
          <option value="CANCELLED">ยกเลิก</option>
        </select>
      </label>
      <div class="filter-actions">
        <button type="submit" :disabled="isLoadingBookings">ค้นหา</button>
        <button
          type="button"
          class="secondary-button"
          :disabled="isLoadingBookings"
          @click="store.clearBookingFilters"
        >
          ล้างค่า
        </button>
      </div>
    </form>

    <div v-if="isLoadingBookings" class="loading-table" role="status" aria-label="Loading bookings">
      <span v-for="index in 4" :key="index"></span>
    </div>

    <div v-else-if="bookingsError" class="inline-message inline-message--error" role="alert">
      <strong>โหลดรายการจองไม่สำเร็จ</strong>
      <p>{{ bookingsError }}</p>
      <button type="button" @click="store.loadBookings">ลองอีกครั้ง</button>
    </div>

    <div v-else-if="bookings.length === 0" class="inline-message">
      <strong>ไม่พบรายการที่ตรงกับตัวกรอง</strong>
      <button type="button" @click="store.clearBookingFilters">แสดงทั้งหมด</button>
    </div>

    <div v-else class="table-scroll">
      <table>
        <caption class="sr-only">
          Cinema bookings
        </caption>
        <thead>
          <tr>
            <th scope="col">ภาพยนตร์และรอบฉาย</th>
            <th scope="col">ลูกค้า</th>
            <th scope="col">ที่นั่ง</th>
            <th scope="col">สถานะ</th>
            <th scope="col">เวลาจอง</th>
            <th scope="col">หมายเลขอ้างอิง</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="booking in bookings" :key="booking.id">
            <td>
              <strong>{{ booking.screening.movie_title || 'ไม่ทราบชื่อเรื่อง' }}</strong>
              <span>
                {{ booking.screening.auditorium_name || 'ไม่ทราบโรง' }} ·
                {{ formatDate(booking.screening.starts_at) }}
              </span>
            </td>
            <td>
              <strong>{{ booking.user.name || 'ไม่ทราบชื่อ' }}</strong>
              <span>{{ booking.user.email || booking.user.id }}</span>
            </td>
            <td class="data-value seat-value">{{ booking.seat_id }}</td>
            <td>
              <span class="status-badge" :class="`status-badge--${booking.status.toLowerCase()}`">
                {{ statusLabel(booking.status) }}
              </span>
            </td>
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
        ก่อนหน้า
      </button>
      <span class="data-value">หน้า {{ bookingMeta.page }} / {{ bookingMeta.total_pages }}</span>
      <button
        type="button"
        class="secondary-button"
        :disabled="bookingMeta.page >= bookingMeta.total_pages || isLoadingBookings"
        @click="store.setBookingPage(bookingMeta.page + 1)"
      >
        ถัดไป
      </button>
    </nav>
  </section>
</template>
