<script setup lang="ts">
import { computed } from 'vue'

import type { Seat, SeatMap } from '../types'

const props = defineProps<{
  seatMap: SeatMap
  canLock: boolean
  activeSeatIds?: string[]
  maxSelectable?: number
  isUpdatingLock: boolean
}>()

const emit = defineEmits<{
  toggle: [seatID: string]
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
  return Boolean(seat.locked_by_me || props.activeSeatIds?.includes(seat.id))
}

function isDisabled(seat: Seat) {
  if (props.isUpdatingLock || !props.canLock) return true
  if (isSelected(seat)) return false
  if (seat.status !== 'AVAILABLE') return true
  return (props.activeSeatIds?.length ?? 0) >= (props.maxSelectable ?? 6)
}

function seatLabel(seat: Seat) {
  if (seat.locked_by_me) return `Seat ${seat.id}, held by you`
  return `Seat ${seat.id}, ${seat.status.toLowerCase()}`
}
</script>

<template>
  <div class="seat-map">
    <div class="screen" aria-hidden="true"><span>SCREEN</span></div>

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
            @click="emit('toggle', seat.id)"
          >
            {{ seat.number }}
          </button>
        </div>
      </div>
    </div>

    <div class="legend" aria-label="Seat status legend">
      <span><i class="legend-swatch legend-swatch--available"></i>ว่าง</span>
      <span><i class="legend-swatch legend-swatch--selected"></i>ที่นั่งของคุณ</span>
      <span><i class="legend-swatch legend-swatch--locked"></i>กำลังถูกเลือก</span>
      <span><i class="legend-swatch legend-swatch--booked"></i>จองแล้ว</span>
    </div>
  </div>
</template>

<style scoped>
.seat-map {
  min-width: 0;
}

.screen {
  position: relative;
  width: min(38rem, 88%);
  height: 3.5rem;
  margin: 0 auto 2.8rem;
  overflow: hidden;
  color: #697586;
  font-size: 0.6rem;
  font-weight: 850;
  letter-spacing: 0.28em;
  text-align: center;
}

.screen::before {
  position: absolute;
  top: 0;
  right: 5%;
  left: 5%;
  height: 2rem;
  border-top: 3px solid #d91920;
  border-radius: 50%;
  box-shadow: 0 -0.6rem 2.2rem rgb(217 25 32 / 18%);
  content: '';
}

.screen span {
  position: relative;
  top: 1rem;
}

.seat-rows {
  display: grid;
  gap: 0.62rem;
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
  color: #5f6a7b;
  font-size: 0.68rem;
  font-weight: 800;
  text-align: center;
}

.seats {
  display: grid;
  gap: 0.42rem;
}

.seat {
  min-height: 2.25rem;
  border: 1px solid #394555;
  border-radius: 0.25rem 0.25rem 0.5rem 0.5rem;
  color: #aeb7c5;
  background: #29292f;
  box-shadow: inset 0 -0.22rem 0 rgb(0 0 0 / 20%);
  font-size: 0.68rem;
  font-weight: 750;
  cursor: pointer;
  transition:
    transform 130ms ease,
    border-color 130ms ease,
    background 130ms ease;
}

.seat:hover:not(:disabled),
.seat--selected {
  border-color: #e63c42;
  color: #fff;
  background: #d91920;
  box-shadow: none;
  transform: translateY(-0.12rem);
}

.seat--locked {
  border-color: #78653f;
  color: #d4b873;
  background: #3b3425;
}

.seat--locked.seat--selected {
  border-color: #e63c42;
  color: #fff;
  background: #d91920;
}

.seat--booked {
  border-color: #242e3b;
  color: #414c5b;
  background: #101720;
}

.seat:disabled {
  cursor: not-allowed;
}

.legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1.25rem;
  margin-top: 1.5rem;
  justify-content: center;
  color: #748091;
  font-size: 0.68rem;
}

.legend span {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.legend-swatch {
  width: 0.8rem;
  height: 0.8rem;
  border: 1px solid #394555;
  border-radius: 0.2rem;
  background: #202b3a;
}

.legend-swatch--selected {
  border-color: #d91920;
  background: #d91920;
}

.legend-swatch--locked {
  border-color: #6c4b34;
  background: #33251f;
}

.legend-swatch--booked {
  border-color: #242e3b;
  background: #101720;
}

@media (prefers-reduced-motion: reduce) {
  .seat {
    transition: none;
  }
}

@media (max-width: 620px) {
  .seat-row {
    grid-template-columns: 1rem minmax(17.5rem, 1fr);
    gap: 0.3rem;
  }

  .seats {
    grid-template-columns: repeat(10, minmax(1.55rem, 1fr)) !important;
    gap: 0.24rem;
  }

  .seat {
    min-height: 1.95rem;
    font-size: 0.6rem;
  }

  .legend {
    justify-content: flex-start;
  }
}
</style>
