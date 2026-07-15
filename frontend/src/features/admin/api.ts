import type {
  AdminAuditFilters,
  AdminAuditLog,
  AdminBooking,
  AdminBookingFilters,
  PageMeta,
} from './types'

interface ListResponse<T> {
  data: T[]
  meta: PageMeta
}

export class AdminApiError extends Error {
  constructor(readonly status: number) {
    super(`Admin request failed with status ${status}`)
  }
}

export async function fetchAdminBookings(
  filters: AdminBookingFilters,
): Promise<ListResponse<AdminBooking>> {
  const query = new URLSearchParams({ page: String(filters.page), page_size: '20' })
  if (filters.movie.trim()) query.set('movie', filters.movie.trim())
  if (filters.status) query.set('status', filters.status)

  return requestList<AdminBooking>(`/api/v1/admin/bookings?${query}`)
}

export async function fetchAdminAuditLogs(
  filters: AdminAuditFilters,
): Promise<ListResponse<AdminAuditLog>> {
  const query = new URLSearchParams({ page: String(filters.page), page_size: '20' })
  if (filters.event) query.set('event', filters.event)

  return requestList<AdminAuditLog>(`/api/v1/admin/audit-logs?${query}`)
}

async function requestList<T>(url: string): Promise<ListResponse<T>> {
  const response = await fetch(url, {
    credentials: 'same-origin',
    headers: { Accept: 'application/json' },
  })
  if (!response.ok) throw new AdminApiError(response.status)
  return (await response.json()) as ListResponse<T>
}
