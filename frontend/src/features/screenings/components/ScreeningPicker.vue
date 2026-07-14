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

const dateFormatter = new Intl.DateTimeFormat('en-GB', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

function optionLabel(screening: ScreeningSummary) {
  return `${screening.movie.title} — ${dateFormatter.format(new Date(screening.starts_at))}`
}
</script>

<template>
  <div class="field">
    <label for="screening">Showtime</label>
    <select
      id="screening"
      :value="modelValue"
      :disabled="disabled || screenings.length === 0"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option v-if="screenings.length === 0" value="">No showtimes available</option>
      <option v-for="screening in screenings" :key="screening.id" :value="screening.id">
        {{ optionLabel(screening) }}
      </option>
    </select>
  </div>
</template>

<style scoped>
.field {
  display: grid;
  gap: 0.5rem;
}

label {
  color: #d4d4d8;
  font-size: 0.85rem;
  font-weight: 700;
}

select {
  width: 100%;
  padding: 0.8rem 2.5rem 0.8rem 0.85rem;
  border: 1px solid #3f3f46;
  border-radius: 0.6rem;
  color: #fafafa;
  background: #18181b;
  cursor: pointer;
}

select:disabled {
  cursor: wait;
  opacity: 0.65;
}
</style>
