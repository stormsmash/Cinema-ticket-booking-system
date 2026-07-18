import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { confirmSeatBooking, fetchSeatMap } from '@/features/screenings/api'
import { subscribeToSeatEvents } from '@/features/screenings/realtime'
import { useScreeningStore } from '@/features/screenings/store'
import type {
  Booking,
  MyTicket,
  ScreeningSummary,
  SeatEvent,
  SeatLock,
  SeatMap,
} from '@/features/screenings/types'

vi.mock('@/features/screenings/api', () => ({
  acquireSeatLock: vi.fn<(screeningID: string, seatID: string) => Promise<SeatLock>>(),
  confirmSeatBooking: vi.fn<(screeningID: string, seatID: string) => Promise<Booking>>(),
  fetchScreenings: vi.fn<() => Promise<ScreeningSummary[]>>(),
  fetchSeatMap: vi.fn<(screeningID: string) => Promise<SeatMap>>(),
  fetchMyTickets: vi.fn<() => Promise<MyTicket[]>>().mockResolvedValue([]),
  releaseSeatLock: vi.fn<(screeningID: string, seatID: string) => Promise<void>>(),
  ScreeningApiError: class ScreeningApiError extends Error {},
}))

vi.mock('@/features/screenings/realtime', () => ({
  subscribeToSeatEvents:
    vi.fn<
      (
        screeningID: string,
        onEvent: (event: SeatEvent) => void,
        onConnected: () => void,
      ) => () => void
    >(),
}))

const availableMap: SeatMap = {
  screening_id: 'screening-1',
  movie: { title: 'Midnight Signal', duration_minutes: 112 },
  auditorium: { name: 'Hall 1', rows: 1, seats_per_row: 1 },
  starts_at: '2026-07-15T12:00:00Z',
  ticket_price_baht: 240,
  seats: [{ id: 'A1', row: 'A', number: 1, status: 'AVAILABLE', locked_by_me: false }],
}

describe('screening store realtime updates', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('reloads the seat map after an event and closes the old screening connection', async () => {
    let onEvent: ((event: SeatEvent) => void) | undefined
    const stop = vi.fn<() => void>()
    vi.mocked(subscribeToSeatEvents).mockImplementation((_id, eventHandler) => {
      onEvent = eventHandler
      return stop
    })
    vi.mocked(fetchSeatMap)
      .mockResolvedValueOnce(structuredClone(availableMap))
      .mockResolvedValueOnce({
        ...structuredClone(availableMap),
        seats: [
          {
            id: 'A1',
            row: 'A',
            number: 1,
            status: 'LOCKED',
            locked_by_me: false,
            lock_expires_at: '2026-07-15T12:10:00Z',
          },
        ],
      })

    const store = useScreeningStore()
    await store.selectScreening('screening-1')
    expect(store.seatMap?.seats[0]?.status).toBe('AVAILABLE')

    onEvent?.({
      version: 1,
      type: 'seat.locked',
      screening_id: 'screening-1',
      seat_id: 'A1',
      status: 'LOCKED',
      occurred_at: '2026-07-15T12:00:00Z',
    })
    await vi.advanceTimersByTimeAsync(75)

    expect(store.seatMap?.seats[0]?.status).toBe('LOCKED')
    expect(store.seatMap?.seats[0]?.locked_by_me).toBe(false)

    store.stopRealtime()
    expect(stop).toHaveBeenCalledOnce()
  })

  it('confirms the active hold and reloads it as booked', async () => {
    vi.mocked(subscribeToSeatEvents).mockReturnValue(vi.fn<() => void>())
    const lockedMap: SeatMap = {
      ...structuredClone(availableMap),
      seats: [
        {
          id: 'A1',
          row: 'A',
          number: 1,
          status: 'LOCKED',
          locked_by_me: true,
          lock_expires_at: '2026-07-15T12:05:00Z',
        },
      ],
    }
    const bookedMap: SeatMap = {
      ...structuredClone(availableMap),
      seats: [{ id: 'A1', row: 'A', number: 1, status: 'BOOKED', locked_by_me: false }],
    }
    const booking: Booking = {
      id: 'booking-123',
      screening_id: 'screening-1',
      seat_id: 'A1',
      price_baht: 240,
      ticket_code: 'LUMINA-booking-123',
      status: 'BOOKED',
      created_at: '2026-07-15T12:00:00Z',
    }
    vi.mocked(fetchSeatMap)
      .mockResolvedValueOnce(structuredClone(lockedMap))
      .mockResolvedValueOnce(structuredClone(bookedMap))
    vi.mocked(confirmSeatBooking).mockResolvedValue(booking)

    const store = useScreeningStore()
    await store.selectScreening('screening-1')
    await store.confirmBooking()

    expect(confirmSeatBooking).toHaveBeenCalledWith('screening-1', 'A1')
    expect(store.confirmedBookings[0]?.id).toBe('booking-123')
    expect(store.activeLocks).toHaveLength(0)
    expect(store.seatMap?.seats[0]?.status).toBe('BOOKED')
  })
})
