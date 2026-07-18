import { afterEach, describe, expect, it, vi } from 'vitest'

import { createPinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import AuthStatus from '@/features/auth/components/AuthStatus.vue'

describe('AuthStatus', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('shows Google sign-in when OAuth is configured', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(response({ data: { google_enabled: true } }))
        .mockResolvedValueOnce(response({}, 401)),
    )

    const wrapper = mount(AuthStatus, {
      global: { plugins: [createPinia()] },
    })
    await flushPromises()

    expect(wrapper.get('a').attributes('href')).toBe('/api/v1/auth/google')
    expect(wrapper.text()).toContain('เข้าสู่ระบบด้วย Google')
  })

  it('shows the current user from the session', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(response({ data: { google_enabled: true } }))
        .mockResolvedValueOnce(
          response({
            data: {
              id: 'user-1',
              email: 'viewer@example.com',
              name: 'Cinema Viewer',
              role: 'USER',
            },
          }),
        ),
    )

    const wrapper = mount(AuthStatus, {
      global: { plugins: [createPinia()] },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Cinema Viewer')
    expect(wrapper.text()).toContain('viewer@example.com')
    expect(wrapper.text()).toContain('ออกจากระบบ')
  })

  it('shows the dashboard link only to an admin', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(response({ data: { google_enabled: true } }))
        .mockResolvedValueOnce(
          response({
            data: {
              id: 'admin-1',
              email: 'admin@example.com',
              name: 'Cinema Admin',
              role: 'ADMIN',
            },
          }),
        ),
    )

    const wrapper = mount(AuthStatus, {
      global: { plugins: [createPinia()] },
    })
    await flushPromises()

    const link = wrapper.get('a.admin-link')
    expect(link.attributes('href')).toBe('/admin')
    expect(link.text()).toBe('หน้าจัดการระบบ')
  })
})

function response(body: unknown, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}
