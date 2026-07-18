<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia'

import AdminAuditPanel from '@/features/admin/components/AdminAuditPanel.vue'
import AdminBookingsPanel from '@/features/admin/components/AdminBookingsPanel.vue'
import AdminOverview from '@/features/admin/components/AdminOverview.vue'
import AdminTopbar from '@/features/admin/components/AdminTopbar.vue'
import { useAdminStore } from '@/features/admin/store'
import './admin-view.css'

const store = useAdminStore()
const { bookings, bookingMeta, auditLogs, auditMeta } = storeToRefs(store)

const bookedOnPage = computed(
  () => bookings.value.filter((booking) => booking.status === 'BOOKED').length,
)
const successEventsOnPage = computed(
  () => auditLogs.value.filter((entry) => entry.event === 'BOOKING_SUCCESS').length,
)

onMounted(store.loadAll)
</script>

<template>
  <div class="admin-app">
    <AdminTopbar />

    <main class="admin-shell">
      <AdminOverview
        :booking-total="bookingMeta.total"
        :booked-on-page="bookedOnPage"
        :audit-total="auditMeta.total"
        :success-events-on-page="successEventsOnPage"
      />
      <AdminBookingsPanel />
      <AdminAuditPanel />
    </main>
  </div>
</template>
