<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { fetchHealth } from '@/features/system/api'

type CheckState = 'checking' | 'online' | 'offline'

const state = ref<CheckState>('checking')
const detail = ref('กำลังเชื่อมต่อ API...')

const label = computed(() => {
  if (state.value === 'online') return 'Backend เชื่อมต่อแล้ว'
  if (state.value === 'offline') return 'Backend ไม่พร้อมใช้งาน'
  return 'กำลังตรวจสอบ Backend'
})

async function checkHealth() {
  state.value = 'checking'
  detail.value = 'กำลังเชื่อมต่อ API...'

  try {
    await fetchHealth()
    state.value = 'online'
    detail.value = 'Frontend เชื่อมต่อกับ Go API ได้ตามปกติ'
  } catch {
    state.value = 'offline'
    detail.value = 'กรุณาเปิด API แล้วลองตรวจสอบอีกครั้ง'
  }
}

onMounted(checkHealth)
</script>

<template>
  <section class="status-card" aria-live="polite">
    <div class="status-heading">
      <span class="status-dot" :class="`status-dot--${state}`" aria-hidden="true"></span>
      <div>
        <p>สถานะระบบ</p>
        <strong>{{ label }}</strong>
      </div>
    </div>

    <p class="status-detail">{{ detail }}</p>

    <button type="button" :disabled="state === 'checking'" @click="checkHealth">
      {{ state === 'checking' ? 'กำลังตรวจสอบ…' : 'ตรวจสอบอีกครั้ง' }}
    </button>
  </section>
</template>

<style scoped>
.status-card {
  padding: 1rem;
  border: 1px solid rgb(255 255 255 / 8%);
  border-radius: 0.25rem;
  background: #090f18;
}

.status-heading {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-heading p,
.status-detail {
  margin: 0;
  color: #6f7a8b;
  font-size: 0.66rem;
}

.status-heading strong {
  display: block;
  margin-top: 0.2rem;
  color: #dce1e8;
  font-size: 0.78rem;
}

.status-dot {
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 50%;
  background: #fbbf24;
  box-shadow: 0 0 0 0.3rem rgb(251 191 36 / 12%);
}

.status-dot--online {
  background: #22c55e;
  box-shadow: 0 0 0 0.3rem rgb(34 197 94 / 12%);
}

.status-dot--offline {
  background: #ef4444;
  box-shadow: 0 0 0 0.3rem rgb(239 68 68 / 12%);
}

.status-detail {
  margin-top: 0.7rem;
  line-height: 1.6;
}

button {
  margin-top: 0.7rem;
  padding: 0.5rem 0.7rem;
  border: 1px solid #303b4a;
  border-radius: 0.4rem;
  color: #b9c1cd;
  background: transparent;
  font-size: 0.65rem;
  font-weight: 750;
  cursor: pointer;
}

button:disabled {
  cursor: wait;
  opacity: 0.65;
}
</style>
