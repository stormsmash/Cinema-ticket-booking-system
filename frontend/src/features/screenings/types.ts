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
}

export interface SeatMap {
  screening_id: string
  movie: Movie
  auditorium: Auditorium
  starts_at: string
  seats: Seat[]
}
