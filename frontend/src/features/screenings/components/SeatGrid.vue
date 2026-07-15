<script setup lang="ts">
import { computed } from 'vue'

import type { Seat, SeatMap } from '../types'

const props = defineProps<{
  seatMap: SeatMap
  canLock: boolean
  activeSeatId?: string
  isUpdatingLock: boolean
}>()

const emit = defineEmits<{
  lock: [seatID: string]
}>()

const seatRows = computed(() => {
  const rows = new Map<string, SeatMap['seats']>()

  for (const seat of props.seatMap.seats) {
    const row = rows.get(seat.row) ?? []
    row.push(seat)
    rows.set(seat.row, row)
  }

  return Array.from(rows.entries())
})

function isSelected(seat: Seat) {
  return seat.locked_by_me || seat.id === props.activeSeatId
}

function isDisabled(seat: Seat) {
  if (props.isUpdatingLock || !props.canLock) return true
  if (seat.status !== 'AVAILABLE') return true
  return Boolean(props.activeSeatId)
}

function seatLabel(seat: Seat) {
  if (seat.locked_by_me) return `Seat ${seat.id}, held by you`
  return `Seat ${seat.id}, ${seat.status.toLowerCase()}`
}
</script>

<template>
  <div class="seat-map">
    <div class="screen" aria-hidden="true">SCREEN</div>

    <div class="seat-rows" aria-label="Cinema seats">
      <div v-for="[rowName, seats] in seatRows" :key="rowName" class="seat-row">
        <span class="row-name" aria-hidden="true">{{ rowName }}</span>
        <div
          class="seats"
          :style="{
            gridTemplateColumns: `repeat(${seatMap.auditorium.seats_per_row}, minmax(2.25rem, 1fr))`,
          }"
        >
          <button
            v-for="seat in seats"
            :key="seat.id"
            type="button"
            class="seat"
            :class="[`seat--${seat.status.toLowerCase()}`, { 'seat--selected': isSelected(seat) }]"
            :disabled="isDisabled(seat)"
            :aria-label="seatLabel(seat)"
            :aria-pressed="isSelected(seat)"
            @click="emit('lock', seat.id)"
          >
            {{ seat.number }}
          </button>
        </div>
      </div>
    </div>

    <div class="legend" aria-label="Seat status legend">
      <span><i class="legend-swatch legend-swatch--available"></i>Available</span>
      <span><i class="legend-swatch legend-swatch--selected"></i>Your hold</span>
      <span><i class="legend-swatch legend-swatch--locked"></i>Locked</span>
      <span><i class="legend-swatch legend-swatch--booked"></i>Booked</span>
    </div>
  </div>
</template>

<style scoped>
.seat-map {
  min-width: 0;
}

.screen {
  width: min(34rem, 82%);
  margin: 0 auto 3rem;
  padding-top: 0.65rem;
  border-top: 3px solid #fbbf24;
  color: #71717a;
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.22em;
  text-align: center;
}

.seat-rows {
  display: grid;
  gap: 0.65rem;
  overflow-x: auto;
  padding: 0.25rem 0.25rem 1rem;
}

.seat-row {
  display: grid;
  grid-template-columns: 1.5rem minmax(24rem, 1fr);
  gap: 0.6rem;
  align-items: center;
}

.row-name {
  color: #71717a;
  font-size: 0.75rem;
  font-weight: 800;
  text-align: center;
}

.seats {
  display: grid;
  gap: 0.45rem;
}

.seat {
  min-height: 2.25rem;
  border: 1px solid #52525b;
  border-radius: 0.45rem 0.45rem 0.7rem 0.7rem;
  color: #e4e4e7;
  background: #27272a;
  font-size: 0.75rem;
  font-weight: 750;
  cursor: pointer;
}

.seat:hover:not(:disabled),
.seat--selected {
  border-color: #fbbf24;
  color: #18181b;
  background: #fbbf24;
}

.seat--locked {
  border-color: #92400e;
  color: #fcd34d;
  background: #451a03;
}

.seat--locked.seat--selected {
  border-color: #fbbf24;
  color: #18181b;
  background: #fbbf24;
}

.seat--booked {
  border-color: #3f3f46;
  color: #71717a;
  background: #18181b;
}

.seat:disabled {
  cursor: not-allowed;
}

.legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1.25rem;
  margin-top: 1.5rem;
  color: #a1a1aa;
  font-size: 0.78rem;
}

.legend span {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.legend-swatch {
  width: 0.8rem;
  height: 0.8rem;
  border: 1px solid #52525b;
  border-radius: 0.2rem;
  background: #27272a;
}

.legend-swatch--selected {
  border-color: #fbbf24;
  background: #fbbf24;
}

.legend-swatch--locked {
  border-color: #92400e;
  background: #451a03;
}

.legend-swatch--booked {
  border-color: #3f3f46;
  background: #18181b;
}
</style>
