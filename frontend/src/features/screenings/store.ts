import { ref } from 'vue'
import { defineStore } from 'pinia'

import { fetchScreenings, fetchSeatMap } from './api'
import type { ScreeningSummary, SeatMap } from './types'

export const useScreeningStore = defineStore('screenings', () => {
  const screenings = ref<ScreeningSummary[]>([])
  const selectedScreeningID = ref('')
  const seatMap = ref<SeatMap | null>(null)
  const screeningsError = ref('')
  const seatsError = ref('')
  const isLoadingScreenings = ref(false)
  const isLoadingSeats = ref(false)

  let seatRequestNumber = 0

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
    selectedScreeningID.value = screeningID
    seatMap.value = null
    seatsError.value = ''
    isLoadingSeats.value = true

    const requestNumber = ++seatRequestNumber

    try {
      const result = await fetchSeatMap(screeningID)
      if (requestNumber === seatRequestNumber) {
        seatMap.value = result
      }
    } catch {
      if (requestNumber === seatRequestNumber) {
        seatsError.value = 'Unable to load the seat map. Please try again.'
      }
    } finally {
      if (requestNumber === seatRequestNumber) {
        isLoadingSeats.value = false
      }
    }
  }

  function reloadSeatMap() {
    if (selectedScreeningID.value) {
      return selectScreening(selectedScreeningID.value)
    }
  }

  return {
    screenings,
    selectedScreeningID,
    seatMap,
    screeningsError,
    seatsError,
    isLoadingScreenings,
    isLoadingSeats,
    loadScreenings,
    selectScreening,
    reloadSeatMap,
  }
})
