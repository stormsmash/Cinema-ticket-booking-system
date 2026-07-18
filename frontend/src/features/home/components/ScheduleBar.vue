<script setup lang="ts">
defineProps<{
  dates: { key: string; label: string }[]
  modelValue: string
  allDatesValue: string
}>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <section class="schedule-bar" aria-label="เลือกสาขาและวันที่">
    <div class="content-shell schedule-inner">
      <div class="branch-info">
        <span class="schedule-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24">
            <path d="M12 21s7-5.4 7-12a7 7 0 1 0-14 0c0 6.6 7 12 7 12Z" />
            <circle cx="12" cy="9" r="2.5" />
          </svg>
        </span>
        <div>
          <small>สาขา</small>
          <strong>รัชโยธิน</strong>
        </div>
      </div>

      <div class="date-filter" role="toolbar" aria-label="กรองตามวันที่">
        <button
          type="button"
          :class="{ active: modelValue === allDatesValue }"
          :aria-pressed="modelValue === allDatesValue"
          @click="emit('update:modelValue', allDatesValue)"
        >
          ทุกรอบ
        </button>
        <button
          v-for="date in dates"
          :key="date.key"
          type="button"
          :class="{ active: modelValue === date.key }"
          :aria-pressed="modelValue === date.key"
          @click="emit('update:modelValue', date.key)"
        >
          {{ date.label }}
        </button>
      </div>
    </div>
  </section>
</template>
