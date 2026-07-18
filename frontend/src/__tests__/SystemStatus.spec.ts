import { afterEach, describe, expect, it, vi } from 'vitest'

import { flushPromises, mount } from '@vue/test-utils'

import SystemStatus from '@/features/system/components/SystemStatus.vue'

describe('SystemStatus', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('shows that the backend is connected', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ status: 'ok' }),
      }),
    )

    const wrapper = mount(SystemStatus)
    await flushPromises()

    expect(wrapper.text()).toContain('Backend เชื่อมต่อแล้ว')
  })

  it('shows an unavailable state when the request fails', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('network error')))

    const wrapper = mount(SystemStatus)
    await flushPromises()

    expect(wrapper.text()).toContain('Backend ไม่พร้อมใช้งาน')
  })
})
