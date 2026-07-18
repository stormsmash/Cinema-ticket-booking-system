<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useAdminStore } from '@/features/admin/store'

const store = useAdminStore()
const { auditLogs, auditMeta, auditFilters, isLoadingAudits, auditsError } = storeToRefs(store)

const dateFormatter = new Intl.DateTimeFormat('th-TH', {
  dateStyle: 'medium',
  timeStyle: 'short',
  timeZone: 'Asia/Bangkok',
})

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '—' : dateFormatter.format(date)
}
</script>

<template>
  <section class="admin-panel" aria-labelledby="audit-title">
    <div class="panel-heading">
      <div>
        <p class="section-label">AUDIT TRAIL</p>
        <h2 id="audit-title">เหตุการณ์สำคัญของระบบ</h2>
      </div>
      <span class="record-count">{{ auditMeta.total }} เหตุการณ์</span>
    </div>

    <form class="filter-form audit-filter" @submit.prevent="store.applyAuditFilter">
      <label>
        ประเภทเหตุการณ์
        <select v-model="auditFilters.event" name="event">
          <option value="">ทุกเหตุการณ์</option>
          <option value="BOOKING_SUCCESS">Booking success</option>
          <option value="BOOKING_TIMEOUT">Booking timeout</option>
          <option value="SEAT_RELEASED">Seat released</option>
          <option value="SYSTEM_ERROR">System error</option>
        </select>
      </label>
      <button type="submit" :disabled="isLoadingAudits">ใช้ตัวกรอง</button>
    </form>

    <div v-if="isLoadingAudits" class="loading-table" role="status" aria-label="Loading audit logs">
      <span v-for="index in 3" :key="index"></span>
    </div>

    <div v-else-if="auditsError" class="inline-message inline-message--error" role="alert">
      <strong>โหลด audit log ไม่สำเร็จ</strong>
      <p>{{ auditsError }}</p>
      <button type="button" @click="store.loadAuditLogs">ลองอีกครั้ง</button>
    </div>

    <div v-else-if="auditLogs.length === 0" class="inline-message">
      <strong>ไม่พบเหตุการณ์ที่ตรงกับตัวกรอง</strong>
      <button type="button" @click="store.clearAuditFilter">แสดงทุกเหตุการณ์</button>
    </div>

    <div v-else class="table-scroll">
      <table>
        <caption class="sr-only">
          System audit events
        </caption>
        <thead>
          <tr>
            <th scope="col">เหตุการณ์</th>
            <th scope="col">รายละเอียด</th>
            <th scope="col">เวลา</th>
            <th scope="col">Event ID</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="entry in auditLogs" :key="entry.id">
            <td>
              <span class="event-badge">{{ entry.event }}</span>
            </td>
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
        ก่อนหน้า
      </button>
      <span class="data-value">หน้า {{ auditMeta.page }} / {{ auditMeta.total_pages }}</span>
      <button
        type="button"
        class="secondary-button"
        :disabled="auditMeta.page >= auditMeta.total_pages || isLoadingAudits"
        @click="store.setAuditPage(auditMeta.page + 1)"
      >
        ถัดไป
      </button>
    </nav>
  </section>
</template>
