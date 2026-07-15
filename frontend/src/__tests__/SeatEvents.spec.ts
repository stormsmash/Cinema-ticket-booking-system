import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { subscribeToSeatEvents } from '@/features/screenings/realtime'
import type { SeatEvent } from '@/features/screenings/types'

class FakeWebSocket extends EventTarget {
  static instances: FakeWebSocket[] = []

  readonly url: string
  close = vi.fn<(code?: number, reason?: string) => void>(() =>
    this.dispatchEvent(new Event('close')),
  )

  constructor(url: string | URL) {
    super()
    this.url = String(url)
    FakeWebSocket.instances.push(this)
  }

  open() {
    this.dispatchEvent(new Event('open'))
  }

  message(data: string) {
    this.dispatchEvent(new MessageEvent('message', { data }))
  }
}

describe('seat event WebSocket', () => {
  beforeEach(() => {
    FakeWebSocket.instances = []
    vi.useFakeTimers()
    vi.stubGlobal('WebSocket', FakeWebSocket)
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('accepts valid events and ignores malformed or unrelated messages', () => {
    const onEvent = vi.fn<(event: SeatEvent) => void>()
    const onConnected = vi.fn<() => void>()
    const stop = subscribeToSeatEvents('screening-1', onEvent, onConnected)
    const socket = FakeWebSocket.instances[0]!

    expect(socket.url).toBe('ws://localhost:3000/api/v1/screenings/screening-1/seat-events')
    socket.open()
    expect(onConnected).toHaveBeenCalledOnce()

    socket.message('{not-json')
    socket.message(
      JSON.stringify({
        version: 1,
        type: 'seat.locked',
        screening_id: 'screening-2',
        seat_id: 'A1',
        status: 'LOCKED',
        occurred_at: '2026-07-15T12:00:00Z',
      }),
    )
    socket.message(
      JSON.stringify({
        version: 1,
        type: 'seat.locked',
        screening_id: 'screening-1',
        seat_id: 'A1',
        status: 'LOCKED',
        occurred_at: '2026-07-15T12:00:00Z',
      }),
    )
    socket.message(
      JSON.stringify({
        version: 1,
        type: 'seat.booked',
        screening_id: 'screening-1',
        seat_id: 'A2',
        status: 'BOOKED',
        occurred_at: '2026-07-15T12:01:00Z',
      }),
    )
    socket.message(
      JSON.stringify({
        version: 1,
        type: 'seat.booked',
        booking_id: 'booking-1',
        screening_id: 'screening-1',
        seat_id: 'A2',
        status: 'BOOKED',
        occurred_at: '2026-07-15T12:01:00Z',
      }),
    )

    expect(onEvent).toHaveBeenCalledTimes(2)
    expect(onEvent.mock.calls[0]?.[0].seat_id).toBe('A1')
    expect(onEvent.mock.calls[1]?.[0].type).toBe('seat.booked')

    stop()
    expect(socket.close).toHaveBeenCalledWith(1000, 'screening changed')
  })

  it('reconnects after an unexpected close', () => {
    const stop = subscribeToSeatEvents(
      'screening-1',
      vi.fn<(event: SeatEvent) => void>(),
      vi.fn<() => void>(),
    )
    const socket = FakeWebSocket.instances[0]!

    socket.dispatchEvent(new Event('close'))
    vi.advanceTimersByTime(1_000)

    expect(FakeWebSocket.instances).toHaveLength(2)
    stop()
  })
})
