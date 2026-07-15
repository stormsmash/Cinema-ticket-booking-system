import { describe, expect, it } from 'vitest'

import { mount } from '@vue/test-utils'

import ScreeningPicker from '@/features/screenings/components/ScreeningPicker.vue'
import SeatGrid from '@/features/screenings/components/SeatGrid.vue'
import SeatLockStatus from '@/features/screenings/components/SeatLockStatus.vue'
import type { Booking, ScreeningSummary, SeatLock, SeatMap } from '@/features/screenings/types'

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
    { id: 'A1', row: 'A', number: 1, status: 'AVAILABLE', locked_by_me: false },
    { id: 'A2', row: 'A', number: 2, status: 'BOOKED', locked_by_me: false },
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
  it('requests a lock for an available seat and disables a booked seat', async () => {
    const wrapper = mount(SeatGrid, {
      props: { seatMap, canLock: true, isUpdatingLock: false },
    })
    const availableSeat = wrapper.get('button[aria-label="Seat A1, available"]')
    const bookedSeat = wrapper.get('button[aria-label="Seat A2, booked"]')

    expect(bookedSeat.attributes('disabled')).toBeDefined()

    await availableSeat.trigger('click')

    expect(wrapper.emitted('lock')).toEqual([['A1']])
    expect(availableSeat.attributes('aria-pressed')).toBe('false')
  })
})

describe('SeatLockStatus', () => {
  it('shows the held seat and lets its owner release it', async () => {
    const lock: SeatLock = {
      screening_id: screening.id,
      seat_id: 'A1',
      status: 'LOCKED',
      expires_at: new Date(Date.now() + 5 * 60 * 1000).toISOString(),
    }
    const wrapper = mount(SeatLockStatus, {
      props: {
        lock,
        signedIn: true,
        isUpdating: false,
        error: '',
        booking: null,
        isConfirming: false,
        bookingError: '',
      },
    })

    expect(wrapper.text()).toContain('Seat A1 held for')
    await wrapper.get('.lock-actions button:last-child').trigger('click')
    expect(wrapper.emitted('release')).toHaveLength(1)

    wrapper.unmount()
  })

  it('asks for confirmation before emitting a booking request', async () => {
    const lock: SeatLock = {
      screening_id: screening.id,
      seat_id: 'A1',
      status: 'LOCKED',
      expires_at: new Date(Date.now() + 5 * 60 * 1000).toISOString(),
    }
    const wrapper = mount(SeatLockStatus, {
      props: {
        lock,
        signedIn: true,
        isUpdating: false,
        error: '',
        booking: null,
        isConfirming: false,
        bookingError: '',
      },
    })

    await wrapper.get('.lock-actions .confirm-button').trigger('click')

    expect(wrapper.get('dialog').attributes('open')).toBeDefined()
    await wrapper.get('.dialog-actions .confirm-button').trigger('click')
    expect(wrapper.emitted('confirm')).toHaveLength(1)

    wrapper.unmount()
  })

  it('shows the confirmed booking reference', () => {
    const booking: Booking = {
      id: 'booking-123',
      screening_id: screening.id,
      seat_id: 'A1',
      status: 'BOOKED',
      created_at: '2026-07-15T12:00:00Z',
    }
    const wrapper = mount(SeatLockStatus, {
      props: {
        lock: null,
        signedIn: true,
        isUpdating: false,
        error: '',
        booking,
        isConfirming: false,
        bookingError: '',
      },
    })

    expect(wrapper.text()).toContain('Seat A1 is booked')
    expect(wrapper.text()).toContain('booking-123')
    wrapper.unmount()
  })
})
