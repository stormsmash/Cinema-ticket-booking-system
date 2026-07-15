import { reactive, ref } from 'vue'
import { defineStore } from 'pinia'

import { AdminApiError, fetchAdminAuditLogs, fetchAdminBookings } from './api'
import type {
  AdminAuditFilters,
  AdminAuditLog,
  AdminBooking,
  AdminBookingFilters,
  PageMeta,
} from './types'

const emptyMeta = (): PageMeta => ({ page: 1, page_size: 20, total: 0, total_pages: 0 })

export const useAdminStore = defineStore('admin', () => {
  const bookings = ref<AdminBooking[]>([])
  const bookingMeta = ref<PageMeta>(emptyMeta())
  const bookingFilters = reactive<AdminBookingFilters>({ movie: '', status: '', page: 1 })
  const isLoadingBookings = ref(false)
  const bookingsError = ref('')

  const auditLogs = ref<AdminAuditLog[]>([])
  const auditMeta = ref<PageMeta>(emptyMeta())
  const auditFilters = reactive<AdminAuditFilters>({ event: '', page: 1 })
  const isLoadingAudits = ref(false)
  const auditsError = ref('')

  async function loadBookings() {
    isLoadingBookings.value = true
    bookingsError.value = ''
    try {
      const result = await fetchAdminBookings(bookingFilters)
      bookings.value = result.data
      bookingMeta.value = result.meta
    } catch (error) {
      bookingsError.value = adminErrorMessage(error, 'bookings')
    } finally {
      isLoadingBookings.value = false
    }
  }

  async function loadAuditLogs() {
    isLoadingAudits.value = true
    auditsError.value = ''
    try {
      const result = await fetchAdminAuditLogs(auditFilters)
      auditLogs.value = result.data
      auditMeta.value = result.meta
    } catch (error) {
      auditsError.value = adminErrorMessage(error, 'audit logs')
    } finally {
      isLoadingAudits.value = false
    }
  }

  function applyBookingFilters() {
    bookingFilters.page = 1
    return loadBookings()
  }

  function clearBookingFilters() {
    bookingFilters.movie = ''
    bookingFilters.status = ''
    bookingFilters.page = 1
    return loadBookings()
  }

  function setBookingPage(page: number) {
    if (page < 1 || page > bookingMeta.value.total_pages || page === bookingFilters.page) return
    bookingFilters.page = page
    return loadBookings()
  }

  function applyAuditFilter() {
    auditFilters.page = 1
    return loadAuditLogs()
  }

  function clearAuditFilter() {
    auditFilters.event = ''
    auditFilters.page = 1
    return loadAuditLogs()
  }

  function setAuditPage(page: number) {
    if (page < 1 || page > auditMeta.value.total_pages || page === auditFilters.page) return
    auditFilters.page = page
    return loadAuditLogs()
  }

  function loadAll() {
    return Promise.all([loadBookings(), loadAuditLogs()])
  }

  return {
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
    loadAll,
    loadBookings,
    loadAuditLogs,
    applyBookingFilters,
    clearBookingFilters,
    setBookingPage,
    applyAuditFilter,
    clearAuditFilter,
    setAuditPage,
  }
})

function adminErrorMessage(error: unknown, resource: string) {
  if (error instanceof AdminApiError && [401, 403].includes(error.status)) {
    return 'Your account no longer has administrator access.'
  }
  return `Unable to load ${resource}. Please try again.`
}
