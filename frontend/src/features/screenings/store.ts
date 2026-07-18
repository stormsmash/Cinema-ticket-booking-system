import { ref } from 'vue'
import { defineStore } from 'pinia'

import {
  acquireSeatLock,
  confirmSeatBooking,
  fetchMyTickets,
  fetchScreenings,
  fetchSeatMap,
  releaseSeatLock,
  ScreeningApiError,
} from './api'
import { subscribeToSeatEvents } from './realtime'
import type { Booking, MyTicket, ScreeningSummary, SeatEvent, SeatLock, SeatMap } from './types'

const maxSeatsPerBooking = 6

export const useScreeningStore = defineStore('screenings', () => {
  const screenings = ref<ScreeningSummary[]>([])
  const selectedScreeningID = ref('')
  const seatMap = ref<SeatMap | null>(null)
  const screeningsError = ref('')
  const seatsError = ref('')
  const isLoadingScreenings = ref(false)
  const isLoadingSeats = ref(false)
  const activeLocks = ref<SeatLock[]>([])
  const isUpdatingLock = ref(false)
  const lockError = ref('')
  const confirmedBookings = ref<Booking[]>([])
  const isConfirmingBooking = ref(false)
  const bookingError = ref('')
  const myTickets = ref<MyTicket[]>([])
  const isLoadingTickets = ref(false)
  const ticketsError = ref('')

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
      screeningsError.value = 'โหลดรอบฉายไม่สำเร็จ กรุณาตรวจสอบการเชื่อมต่อ API'
    } finally {
      isLoadingScreenings.value = false
    }
  }

  async function selectScreening(screeningID: string) {
    stopRealtime()
    selectedScreeningID.value = screeningID
    seatMap.value = null
    activeLocks.value = []
    lockError.value = ''
    confirmedBookings.value = []
    bookingError.value = ''
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
      activeLocks.value = locksFromSeatMap(result)
      seatsError.value = ''
      return true
    } catch {
      if (
        requestNumber === seatRequestNumber &&
        selectedScreeningID.value === screeningID &&
        !seatMap.value
      ) {
        seatsError.value = 'โหลดผังที่นั่งไม่สำเร็จ กรุณาลองอีกครั้ง'
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
      event.type === 'seat.expired' &&
      activeLocks.value.some((lock) => lock.seat_id === event.seat_id)

    if (realtimeRefreshTimer) clearTimeout(realtimeRefreshTimer)
    realtimeRefreshTimer = setTimeout(async () => {
      realtimeRefreshTimer = null
      await reloadSeatMap()
      if (ownLockExpired) {
        lockError.value = `หมดเวลาพักที่นั่ง ${event.seat_id} กรุณาเลือกใหม่`
      }
    }, 75)
  }

  function stopRealtime() {
    stopSeatEvents?.()
    stopSeatEvents = null
    if (realtimeRefreshTimer) clearTimeout(realtimeRefreshTimer)
    realtimeRefreshTimer = null
  }

  async function toggleSeatLock(seatID: string) {
    const existing = activeLocks.value.find((lock) => lock.seat_id === seatID)
    if (existing) {
      await unlockSeat(seatID)
      return
    }
    await lockSeat(seatID)
  }

  async function lockSeat(seatID: string) {
    if (!selectedScreeningID.value || isUpdatingLock.value) return
    if (activeLocks.value.length >= maxSeatsPerBooking) {
      lockError.value = `เลือกได้สูงสุด ${maxSeatsPerBooking} ที่นั่งต่อรายการ`
      return
    }

    isUpdatingLock.value = true
    lockError.value = ''
    confirmedBookings.value = []
    bookingError.value = ''

    try {
      const lock = await acquireSeatLock(selectedScreeningID.value, seatID)
      activeLocks.value = [...activeLocks.value, lock].sort((left, right) =>
        left.seat_id.localeCompare(right.seat_id, undefined, { numeric: true }),
      )

      const seat = seatMap.value?.seats.find((item) => item.id === lock.seat_id)
      if (seat) {
        seat.status = 'LOCKED'
        seat.locked_by_me = true
        seat.lock_expires_at = lock.expires_at
      }
    } catch (error) {
      if (error instanceof ScreeningApiError && error.status === 409) {
        await reloadSeatMap()
      }
      lockError.value = lockErrorMessage(error)
    } finally {
      isUpdatingLock.value = false
    }
  }

  async function unlockSeat(seatID: string) {
    const lock = activeLocks.value.find((item) => item.seat_id === seatID)
    if (!lock || isUpdatingLock.value) return

    isUpdatingLock.value = true
    lockError.value = ''
    bookingError.value = ''

    try {
      await releaseSeatLock(lock.screening_id, lock.seat_id)
      activeLocks.value = activeLocks.value.filter((item) => item.seat_id !== seatID)
      await reloadSeatMap()
    } catch (error) {
      lockError.value = lockErrorMessage(error)
    } finally {
      isUpdatingLock.value = false
    }
  }

  async function unlockAllSeats() {
    if (isUpdatingLock.value || activeLocks.value.length === 0) return
    isUpdatingLock.value = true
    lockError.value = ''

    const locks = [...activeLocks.value]
    try {
      const results = await Promise.allSettled(
        locks.map((lock) => releaseSeatLock(lock.screening_id, lock.seat_id)),
      )
      if (results.some((result) => result.status === 'rejected')) {
        lockError.value = 'ยกเลิกที่นั่งบางรายการไม่สำเร็จ กรุณาลองอีกครั้ง'
      }
      await reloadSeatMap()
    } finally {
      isUpdatingLock.value = false
    }
  }

  async function handleLockExpired() {
    await reloadSeatMap()
    lockError.value = 'มีที่นั่งหมดเวลาพัก กรุณาตรวจสอบรายการอีกครั้ง'
  }

  async function confirmBooking() {
    const locks = [...activeLocks.value]
    if (!locks.length || isConfirmingBooking.value || isUpdatingLock.value) return

    isConfirmingBooking.value = true
    bookingError.value = ''

    try {
      const results = await Promise.allSettled(
        locks.map((lock) => confirmSeatBooking(lock.screening_id, lock.seat_id)),
      )
      const successful = results.flatMap((result) =>
        result.status === 'fulfilled' ? [result.value] : [],
      )
      const failed = results.find((result) => result.status === 'rejected')

      confirmedBookings.value = successful
      await reloadSeatMap()
      if (successful.length) await loadMyTickets()

      if (failed?.status === 'rejected') {
        bookingError.value = successful.length
          ? `จองสำเร็จ ${successful.length} ที่นั่ง แต่มีบางที่นั่งไม่สำเร็จ กรุณาตรวจสอบตั๋วของฉัน`
          : bookingErrorMessage(failed.reason)
      }
    } finally {
      isConfirmingBooking.value = false
    }
  }

  async function loadMyTickets() {
    isLoadingTickets.value = true
    ticketsError.value = ''
    try {
      myTickets.value = await fetchMyTickets()
    } catch (error) {
      myTickets.value = []
      if (!(error instanceof ScreeningApiError && error.status === 401)) {
        ticketsError.value = 'โหลดตั๋วไม่สำเร็จ กรุณาลองอีกครั้ง'
      }
    } finally {
      isLoadingTickets.value = false
    }
  }

  function clearTickets() {
    myTickets.value = []
    ticketsError.value = ''
  }

  return {
    screenings,
    selectedScreeningID,
    seatMap,
    screeningsError,
    seatsError,
    isLoadingScreenings,
    isLoadingSeats,
    activeLocks,
    maxSeatsPerBooking,
    isUpdatingLock,
    lockError,
    confirmedBookings,
    isConfirmingBooking,
    bookingError,
    myTickets,
    isLoadingTickets,
    ticketsError,
    loadScreenings,
    selectScreening,
    reloadSeatMap,
    toggleSeatLock,
    unlockAllSeats,
    handleLockExpired,
    confirmBooking,
    loadMyTickets,
    clearTickets,
    stopRealtime,
  }
})

function locksFromSeatMap(seatMap: SeatMap): SeatLock[] {
  return seatMap.seats.flatMap((seat) => {
    if (!seat.locked_by_me || !seat.lock_expires_at) return []
    return [
      {
        screening_id: seatMap.screening_id,
        seat_id: seat.id,
        status: 'LOCKED' as const,
        expires_at: seat.lock_expires_at,
      },
    ]
  })
}

function bookingErrorMessage(error: unknown) {
  if (error instanceof ScreeningApiError) {
    if (error.status === 401) return 'Session หมดอายุ กรุณาเข้าสู่ระบบอีกครั้ง'
    if (error.code === 'SEAT_LOCK_EXPIRED') return 'หมดเวลาพักที่นั่ง กรุณาเลือกที่นั่งใหม่'
    if (error.code === 'SEAT_LOCK_NOT_OWNED') return 'ที่นั่งนี้ไม่ได้ถูกพักด้วยบัญชีของคุณ'
    if (error.code === 'SEAT_ALREADY_BOOKED') return 'ที่นั่งถูกจองไปแล้ว กรุณาเลือกที่นั่งอื่น'
    if (error.code === 'SCREENING_STARTED') return 'รอบฉายนี้เริ่มแล้ว ไม่สามารถจองได้'
  }

  return 'ยืนยันการจองไม่สำเร็จ กรุณาลองอีกครั้ง'
}

function lockErrorMessage(error: unknown) {
  if (error instanceof ScreeningApiError) {
    if (error.status === 401) return 'Session หมดอายุ กรุณาเข้าสู่ระบบอีกครั้ง'
    if (error.status === 409) return 'มีผู้ใช้อื่นเลือกที่นั่งนี้ก่อน กรุณาเลือกที่นั่งอื่น'
  }

  return 'เปลี่ยนสถานะที่นั่งไม่สำเร็จ กรุณาลองอีกครั้ง'
}
