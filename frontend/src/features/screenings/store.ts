import { ref } from 'vue'
import { defineStore } from 'pinia'

import {
  acquireSeatLock,
  fetchScreenings,
  fetchSeatMap,
  releaseSeatLock,
  ScreeningApiError,
} from './api'
import { subscribeToSeatEvents } from './realtime'
import type { ScreeningSummary, SeatEvent, SeatLock, SeatMap } from './types'

export const useScreeningStore = defineStore('screenings', () => {
  const screenings = ref<ScreeningSummary[]>([])
  const selectedScreeningID = ref('')
  const seatMap = ref<SeatMap | null>(null)
  const screeningsError = ref('')
  const seatsError = ref('')
  const isLoadingScreenings = ref(false)
  const isLoadingSeats = ref(false)
  const activeLock = ref<SeatLock | null>(null)
  const isUpdatingLock = ref(false)
  const lockError = ref('')

  let seatRequestNumber = 0
  let stopSeatEvents: (() => void) | null = null
  let realtimeRefreshTimer: ReturnType<typeof setTimeout> | null = null

  async function loadScreenings() {
    isLoadingScreenings.value = true
    screeningsError.value = ''

    try {
      screenings.value = await fetchScreenings()

      const firstScreening = screenings.value[0]
      if (firstScreening) {
        await selectScreening(firstScreening.id)
      } else {
        selectedScreeningID.value = ''
        seatMap.value = null
      }
    } catch {
      screeningsError.value = 'Unable to load showtimes. Check that the API is running.'
    } finally {
      isLoadingScreenings.value = false
    }
  }

  async function selectScreening(screeningID: string) {
    stopRealtime()
    selectedScreeningID.value = screeningID
    seatMap.value = null
    activeLock.value = null
    lockError.value = ''
    seatsError.value = ''

    const loaded = await loadSeatMap(screeningID, true)
    if (loaded && selectedScreeningID.value === screeningID) {
      startRealtime(screeningID)
    }
  }

  function reloadSeatMap() {
    if (selectedScreeningID.value) {
      return loadSeatMap(selectedScreeningID.value, seatMap.value === null)
    }
  }

  async function loadSeatMap(screeningID: string, showLoadingState: boolean) {
    const requestNumber = ++seatRequestNumber
    if (showLoadingState) isLoadingSeats.value = true

    try {
      const result = await fetchSeatMap(screeningID)
      if (requestNumber !== seatRequestNumber || selectedScreeningID.value !== screeningID) {
        return false
      }

      seatMap.value = result
      activeLock.value = lockFromSeatMap(result)
      seatsError.value = ''
      return true
    } catch {
      if (
        requestNumber === seatRequestNumber &&
        selectedScreeningID.value === screeningID &&
        !seatMap.value
      ) {
        seatsError.value = 'Unable to load the seat map. Please try again.'
      }
      return false
    } finally {
      if (requestNumber === seatRequestNumber && showLoadingState) {
        isLoadingSeats.value = false
      }
    }
  }

  function startRealtime(screeningID: string) {
    stopSeatEvents = subscribeToSeatEvents(
      screeningID,
      scheduleRealtimeRefresh,
      () => void reloadSeatMap(),
    )
  }

  function scheduleRealtimeRefresh(event: SeatEvent) {
    const ownLockExpired =
      event.type === 'seat.expired' && activeLock.value?.seat_id === event.seat_id

    if (realtimeRefreshTimer) clearTimeout(realtimeRefreshTimer)
    realtimeRefreshTimer = setTimeout(async () => {
      realtimeRefreshTimer = null
      await reloadSeatMap()
      if (ownLockExpired && !activeLock.value) {
        lockError.value = 'Your seat hold expired. Choose an available seat to try again.'
      }
    }, 75)
  }

  function stopRealtime() {
    stopSeatEvents?.()
    stopSeatEvents = null
    if (realtimeRefreshTimer) clearTimeout(realtimeRefreshTimer)
    realtimeRefreshTimer = null
  }

  async function lockSeat(seatID: string) {
    if (!selectedScreeningID.value || activeLock.value || isUpdatingLock.value) return

    isUpdatingLock.value = true
    lockError.value = ''

    try {
      const lock = await acquireSeatLock(selectedScreeningID.value, seatID)
      activeLock.value = lock

      const seat = seatMap.value?.seats.find((item) => item.id === lock.seat_id)
      if (seat) {
        seat.status = 'LOCKED'
        seat.locked_by_me = true
        seat.lock_expires_at = lock.expires_at
      }
    } catch (error) {
      const message = lockErrorMessage(error)
      if (error instanceof ScreeningApiError && error.status === 409) {
        await reloadSeatMap()
      }
      lockError.value = message
    } finally {
      isUpdatingLock.value = false
    }
  }

  async function unlockSeat() {
    const lock = activeLock.value
    if (!lock || isUpdatingLock.value) return

    isUpdatingLock.value = true
    lockError.value = ''

    try {
      await releaseSeatLock(lock.screening_id, lock.seat_id)
      activeLock.value = null
      await reloadSeatMap()
    } catch (error) {
      lockError.value = lockErrorMessage(error)
    } finally {
      isUpdatingLock.value = false
    }
  }

  async function handleLockExpired() {
    activeLock.value = null
    await reloadSeatMap()
    lockError.value = 'Your seat hold expired. Choose an available seat to try again.'
  }

  return {
    screenings,
    selectedScreeningID,
    seatMap,
    screeningsError,
    seatsError,
    isLoadingScreenings,
    isLoadingSeats,
    activeLock,
    isUpdatingLock,
    lockError,
    loadScreenings,
    selectScreening,
    reloadSeatMap,
    lockSeat,
    unlockSeat,
    handleLockExpired,
    stopRealtime,
  }
})

function lockFromSeatMap(seatMap: SeatMap): SeatLock | null {
  const seat = seatMap.seats.find((item) => item.locked_by_me && item.lock_expires_at)
  if (!seat?.lock_expires_at) return null

  return {
    screening_id: seatMap.screening_id,
    seat_id: seat.id,
    status: 'LOCKED',
    expires_at: seat.lock_expires_at,
  }
}

function lockErrorMessage(error: unknown) {
  if (error instanceof ScreeningApiError) {
    if (error.status === 401) return 'Your session expired. Sign in again before locking a seat.'
    if (error.status === 409) return 'Another viewer locked that seat first. Choose another seat.'
  }

  return 'Unable to update the seat hold. Please try again.'
}
