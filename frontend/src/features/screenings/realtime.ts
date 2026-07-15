import type { SeatEvent } from './types'

const reconnectDelayMilliseconds = 1_000
const maximumReconnectDelayMilliseconds = 10_000

export function subscribeToSeatEvents(
  screeningID: string,
  onEvent: (event: SeatEvent) => void,
  onConnected: () => void,
) {
  let socket: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectAttempts = 0
  let stopped = false

  function connect() {
    if (stopped) return

    socket = new WebSocket(seatEventsURL(screeningID))
    socket.addEventListener('open', () => {
      if (!stopped) {
        reconnectAttempts = 0
        onConnected()
      }
    })
    socket.addEventListener('message', (message) => {
      if (stopped || typeof message.data !== 'string') return

      const event = parseSeatEvent(message.data, screeningID)
      if (event) onEvent(event)
    })
    socket.addEventListener('close', () => {
      socket = null
      if (!stopped) {
        const delay = Math.min(
          reconnectDelayMilliseconds * 2 ** reconnectAttempts,
          maximumReconnectDelayMilliseconds,
        )
        reconnectAttempts += 1
        reconnectTimer = setTimeout(connect, delay)
      }
    })
  }

  connect()

  return () => {
    stopped = true
    if (reconnectTimer) clearTimeout(reconnectTimer)
    socket?.close(1000, 'screening changed')
    socket = null
  }
}

function seatEventsURL(screeningID: string) {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}/api/v1/screenings/${encodeURIComponent(screeningID)}/seat-events`
}

function parseSeatEvent(value: string, screeningID: string): SeatEvent | null {
  try {
    const event = JSON.parse(value) as Partial<SeatEvent>
    const validType =
      event.type === 'seat.locked' ||
      event.type === 'seat.released' ||
      event.type === 'seat.expired' ||
      event.type === 'seat.booked'
    const validStatus =
      (event.type === 'seat.locked' && event.status === 'LOCKED') ||
      ((event.type === 'seat.released' || event.type === 'seat.expired') &&
        event.status === 'AVAILABLE') ||
      (event.type === 'seat.booked' && event.status === 'BOOKED')
    const validBookingID =
      event.type !== 'seat.booked' ||
      (typeof event.booking_id === 'string' && event.booking_id.trim().length > 0)
    const validSeatID =
      typeof event.seat_id === 'string' &&
      event.seat_id.length > 0 &&
      event.seat_id === event.seat_id.trim() &&
      event.seat_id.length <= 16
    const validOccurredAt =
      typeof event.occurred_at === 'string' && !Number.isNaN(Date.parse(event.occurred_at))

    if (
      event.version !== 1 ||
      !validType ||
      !validStatus ||
      !validBookingID ||
      event.screening_id !== screeningID ||
      !validSeatID ||
      !validOccurredAt
    ) {
      return null
    }

    return event as SeatEvent
  } catch {
    return null
  }
}
