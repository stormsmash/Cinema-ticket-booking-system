export type SeatStatus = 'AVAILABLE' | 'LOCKED' | 'BOOKED'

export interface Movie {
  title: string
  duration_minutes: number
}

export interface Auditorium {
  name: string
  rows: number
  seats_per_row: number
}

export interface ScreeningSummary {
  id: string
  movie: Movie
  auditorium: Auditorium
  starts_at: string
  ticket_price_baht: number
}

export interface Seat {
  id: string
  row: string
  number: number
  status: SeatStatus
  locked_by_me: boolean
  lock_expires_at?: string
}

export interface SeatMap {
  screening_id: string
  movie: Movie
  auditorium: Auditorium
  starts_at: string
  ticket_price_baht: number
  seats: Seat[]
}

export interface SeatLock {
  screening_id: string
  seat_id: string
  status: 'LOCKED'
  expires_at: string
}

export interface Booking {
  id: string
  screening_id: string
  seat_id: string
  price_baht: number
  ticket_code: string
  status: 'BOOKED'
  created_at: string
}

export interface MyTicket extends Booking {
  movie_title: string
  auditorium_name: string
  starts_at: string
}

export interface SeatEvent {
  version: 1
  type: 'seat.locked' | 'seat.released' | 'seat.expired' | 'seat.booked'
  booking_id?: string
  screening_id: string
  seat_id: string
  status: 'LOCKED' | 'AVAILABLE' | 'BOOKED'
  expires_at?: string
  occurred_at: string
}
