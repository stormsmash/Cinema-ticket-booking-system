import { afterEach, describe, expect, it, vi } from 'vitest'

import { createPinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import AdminView from '@/views/AdminView.vue'

describe('AdminView', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders admin records and sends movie filters to the API', async () => {
    const fetchMock = vi.fn<
      (input: string | URL | Request) => Promise<ReturnType<typeof response>>
    >((input) => {
      const url = String(input)
      if (url.includes('/admin/bookings')) {
        return Promise.resolve(
          response({
            data: [
              {
                id: 'booking-1',
                seat_id: 'A1',
                status: 'BOOKED',
                created_at: '2026-07-15T12:00:00Z',
                user: { id: 'user-1', name: 'Cinema Viewer', email: 'viewer@example.com' },
                screening: {
                  id: 'screening-1',
                  movie_title: 'Midnight Signal',
                  auditorium_name: 'Hall 1',
                  starts_at: '2026-07-16T19:00:00Z',
                },
              },
            ],
            meta: { page: 1, page_size: 20, total: 1, total_pages: 1 },
          }),
        )
      }
      return Promise.resolve(
        response({
          data: [
            {
              id: 'audit-1',
              event: 'BOOKING_SUCCESS',
              booking_id: 'booking-1',
              seat_id: 'A1',
              created_at: '2026-07-15T12:00:00Z',
            },
          ],
          meta: { page: 1, page_size: 20, total: 1, total_pages: 1 },
        }),
      )
    })
    vi.stubGlobal('fetch', fetchMock)

    const wrapper = mount(AdminView, {
      global: { plugins: [createPinia()] },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Midnight Signal')
    expect(wrapper.text()).toContain('viewer@example.com')
    expect(wrapper.text()).toContain('BOOKING_SUCCESS')

    await wrapper.get('input[name="movie"]').setValue('Orbit & Beyond')
    await wrapper.findAll('form')[0]!.trigger('submit')
    await flushPromises()

    const bookingURLs = fetchMock.mock.calls
      .map(([input]) => String(input))
      .filter((url) => url.includes('/admin/bookings'))
    expect(bookingURLs[bookingURLs.length - 1]).toContain('movie=Orbit+%26+Beyond')
  })
})

function response(body: unknown, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}
