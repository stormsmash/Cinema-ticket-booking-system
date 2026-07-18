export type BookingStatus = 'HOLDING' | 'BOOKED' | 'TIMED_OUT' | 'CANCELLED'
export type AuditEvent = 'BOOKING_SUCCESS' | 'BOOKING_TIMEOUT' | 'SEAT_RELEASED' | 'SYSTEM_ERROR'

export interface AdminBooking {
  id: string
  seat_id: string
  status: BookingStatus
  created_at: string
  user: {
    id: string
    name: string
    email: string
  }
  screening: {
    id: string
    movie_title: string
    auditorium_name: string
    starts_at: string
  }
}

export interface AdminAuditLog {
  id: string
  event: AuditEvent
  booking_id?: string
  user_id?: string
  screening_id?: string
  seat_id?: string
  message?: string
  created_at: string
}

export interface PageMeta {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface AdminBookingFilters {
  movie: string
  status: '' | BookingStatus
  page: number
}

export interface AdminAuditFilters {
  event: '' | AuditEvent
  page: number
}
