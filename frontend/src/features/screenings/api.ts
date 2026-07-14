import type { ScreeningSummary, SeatMap } from './types'

interface ApiResponse<T> {
  data: T
}

export async function fetchScreenings(): Promise<ScreeningSummary[]> {
  return request<ScreeningSummary[]>('/api/v1/screenings')
}

export async function fetchSeatMap(screeningID: string): Promise<SeatMap> {
  return request<SeatMap>(`/api/v1/screenings/${screeningID}/seats`)
}

async function request<T>(url: string): Promise<T> {
  const response = await fetch(url, {
    headers: {
      Accept: 'application/json',
    },
  })

  if (!response.ok) {
    throw new Error(`Request failed with status ${response.status}`)
  }

  const payload = (await response.json()) as ApiResponse<T>
  if (!payload || !('data' in payload)) {
    throw new Error('API response has an unexpected format')
  }

  return payload.data
}
