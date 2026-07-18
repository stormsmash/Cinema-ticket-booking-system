<script setup lang="ts">
import type { ScreeningSummary } from '../types'

defineProps<{
  screenings: ScreeningSummary[]
  modelValue: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [screeningID: string]
}>()

const dateFormatter = new Intl.DateTimeFormat('th-TH', {
  weekday: 'short',
  day: 'numeric',
  month: 'short',
  hour: '2-digit',
  minute: '2-digit',
  timeZone: 'Asia/Bangkok',
})

function optionLabel(screening: ScreeningSummary) {
  return `${screening.movie.title} · ${dateFormatter.format(new Date(screening.starts_at))}`
}
</script>

<template>
  <div class="field">
    <label for="screening">เปลี่ยนภาพยนตร์และรอบฉาย</label>
    <select
      id="screening"
      :value="modelValue"
      :disabled="disabled || screenings.length === 0"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option v-if="screenings.length === 0" value="">ยังไม่มีรอบฉาย</option>
      <option v-for="screening in screenings" :key="screening.id" :value="screening.id">
        {{ optionLabel(screening) }}
      </option>
    </select>
  </div>
</template>

<style scoped>
.field {
  display: grid;
  gap: 0.55rem;
}

label {
  color: #8e8e96;
  font-size: 0.68rem;
  font-weight: 700;
}

select {
  width: 100%;
  min-height: 2.8rem;
  padding: 0.7rem 2.5rem 0.7rem 0.75rem;
  border: 1px solid #3f3f45;
  border-radius: 0.25rem;
  color: #ededf0;
  background: #242429;
  color-scheme: dark;
  font-size: 0.72rem;
  cursor: pointer;
}

select:disabled {
  cursor: wait;
  opacity: 0.65;
}
</style>
