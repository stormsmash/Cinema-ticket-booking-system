import { createRouter, createWebHistory } from 'vue-router'

import { useAuthStore } from '@/features/auth/store'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
      meta: { title: 'Cinema Ticket Booking System — จองบัตรภาพยนตร์' },
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/AdminView.vue'),
      meta: { requiresAdmin: true, title: 'จัดการระบบ — Cinema Ticket Booking System' },
    },
  ],
})

router.afterEach((to) => {
  document.title =
    typeof to.meta.title === 'string' ? to.meta.title : 'Cinema Ticket Booking System'
})

router.beforeEach(async (to) => {
  if (!to.meta.requiresAdmin) return true

  const auth = useAuthStore()
  await auth.ensureLoaded()
  if (auth.user?.role === 'ADMIN') return true

  return { name: 'home' }
})

export default router
