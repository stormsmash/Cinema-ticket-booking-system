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
      event.type === 'seat.expired'
    const validStatus =
      (event.type === 'seat.locked' && event.status === 'LOCKED') ||
      ((event.type === 'seat.released' || event.type === 'seat.expired') &&
        event.status === 'AVAILABLE')

    if (
      event.version !== 1 ||
      !validType ||
      !validStatus ||
      event.screening_id !== screeningID ||
      typeof event.seat_id !== 'string' ||
      event.seat_id === '' ||
      typeof event.occurred_at !== 'string'
    ) {
      return null
    }

    return event as SeatEvent
  } catch {
    return null
  }
}
