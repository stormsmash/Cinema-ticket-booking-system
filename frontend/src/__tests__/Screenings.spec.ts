import { describe, expect, it } from 'vitest'

import { mount } from '@vue/test-utils'

import ScreeningPicker from '@/features/screenings/components/ScreeningPicker.vue'
import SeatGrid from '@/features/screenings/components/SeatGrid.vue'
import type { ScreeningSummary, SeatMap } from '@/features/screenings/types'

const screening: ScreeningSummary = {
  id: 'screening-1',
  movie: { title: 'Midnight Signal', duration_minutes: 112 },
  auditorium: { name: 'Hall 1', rows: 1, seats_per_row: 2 },
  starts_at: '2026-07-15T12:00:00Z',
}

const seatMap: SeatMap = {
  screening_id: screening.id,
  movie: screening.movie,
  auditorium: screening.auditorium,
  starts_at: screening.starts_at,
  seats: [
    { id: 'A1', row: 'A', number: 1, status: 'AVAILABLE' },
    { id: 'A2', row: 'A', number: 2, status: 'BOOKED' },
  ],
}

describe('ScreeningPicker', () => {
  it('emits the selected screening ID', async () => {
    const wrapper = mount(ScreeningPicker, {
      props: { screenings: [screening], modelValue: '' },
    })

    await wrapper.get('select').setValue(screening.id)

    expect(wrapper.emitted('update:modelValue')).toEqual([[screening.id]])
  })
})

describe('SeatGrid', () => {
  it('allows an available seat to be selected and disables a booked seat', async () => {
    const wrapper = mount(SeatGrid, { props: { seatMap } })
    const availableSeat = wrapper.get('button[aria-label="Seat A1, available"]')
    const bookedSeat = wrapper.get('button[aria-label="Seat A2, booked"]')

    expect(bookedSeat.attributes('disabled')).toBeDefined()

    await availableSeat.trigger('click')

    expect(wrapper.text()).toContain('Seat A1 selected')
    expect(availableSeat.attributes('aria-pressed')).toBe('true')
  })
})
