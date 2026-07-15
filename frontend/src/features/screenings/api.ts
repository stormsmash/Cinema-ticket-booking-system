import type { Booking, ScreeningSummary, SeatLock, SeatMap } from './types'

interface ApiResponse<T> {
  data: T
}

export async function fetchScreenings(): Promise<ScreeningSummary[]> {
  return request<ScreeningSummary[]>('/api/v1/screenings')
}

export async function fetchSeatMap(screeningID: string): Promise<SeatMap> {
  return request<SeatMap>(`/api/v1/screenings/${screeningID}/seats`)
}

export async function acquireSeatLock(screeningID: string, seatID: string): Promise<SeatLock> {
  return request<SeatLock>(`/api/v1/screenings/${screeningID}/seats/${seatID}/lock`, {
    method: 'POST',
  })
}

export async function releaseSeatLock(screeningID: string, seatID: string): Promise<void> {
  const response = await fetch(`/api/v1/screenings/${screeningID}/seats/${seatID}/lock`, {
    method: 'DELETE',
    credentials: 'same-origin',
  })

  if (!response.ok) {
    throw await apiError(response)
  }
}

export async function confirmSeatBooking(screeningID: string, seatID: string): Promise<Booking> {
  return request<Booking>('/api/v1/bookings', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ screening_id: screeningID, seat_id: seatID }),
  })
}

export class ScreeningApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message)
  }
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...init,
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json',
      ...init?.headers,
    },
  })

  if (!response.ok) {
    throw await apiError(response)
  }

  const payload = (await response.json()) as ApiResponse<T>
  if (!payload || !('data' in payload)) {
    throw new Error('API response has an unexpected format')
  }

  return payload.data
}

async function apiError(response: Response) {
  let code = 'REQUEST_FAILED'
  let message = `Request failed with status ${response.status}`

  try {
    const payload = (await response.json()) as {
      error?: { code?: string; message?: string }
    }
    code = payload.error?.code ?? code
    message = payload.error?.message ?? message
  } catch {
    // The status code still gives the caller enough information to handle the failure.
  }

  return new ScreeningApiError(response.status, code, message)
}
