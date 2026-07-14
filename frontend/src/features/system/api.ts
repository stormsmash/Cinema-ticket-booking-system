export interface HealthResponse {
  status: 'ok'
}

export async function fetchHealth(): Promise<HealthResponse> {
  const response = await fetch('/api/v1/health', {
    headers: {
      Accept: 'application/json',
    },
  })

  if (!response.ok) {
    throw new Error(`Health request failed with status ${response.status}`)
  }

  const payload: unknown = await response.json()
  if (!isHealthResponse(payload)) {
    throw new Error('Health response has an unexpected format')
  }

  return payload
}

function isHealthResponse(value: unknown): value is HealthResponse {
  return typeof value === 'object' && value !== null && 'status' in value && value.status === 'ok'
}
