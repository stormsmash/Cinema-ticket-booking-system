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
  seats: Seat[]
}

export interface SeatLock {
  screening_id: string
  seat_id: string
  status: 'LOCKED'
  expires_at: string
}
