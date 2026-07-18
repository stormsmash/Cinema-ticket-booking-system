<script setup lang="ts">
import MovieCard from '@/features/movies/components/MovieCard.vue'
import type { MoviePresentation } from '@/features/movies/catalog'
import type { ScreeningSummary } from '@/features/screenings/types'

defineProps<{
  items: { movie: MoviePresentation; screening: ScreeningSummary }[]
  genres: string[]
  activeGenre: string
  selectedScreeningId: string
  loading: boolean
  error: string
}>()
const emit = defineEmits<{
  select: [screeningID: string]
  'update:activeGenre': [genre: string]
  retry: []
  clearFilters: []
}>()
</script>

<template>
  <section id="movies" class="movie-section">
    <div class="content-shell">
      <div class="section-heading">
        <div>
          <p class="eyebrow">THAI MOVIE SELECTION</p>
          <h2>เลือกภาพยนตร์</h2>
          <p>รายการหนังไทยสำหรับทดลองระบบจอง เลือกเรื่องแล้วดูรอบฉายและที่นั่งได้ทันที</p>
        </div>
        <span class="result-count">{{ items.length }} เรื่อง</span>
      </div>

      <div class="genre-tabs" role="toolbar" aria-label="กรองตามประเภทภาพยนตร์">
        <button
          v-for="genre in genres"
          :key="genre"
          type="button"
          :class="{ active: activeGenre === genre }"
          :aria-pressed="activeGenre === genre"
          @click="emit('update:activeGenre', genre)"
        >
          {{ genre }}
        </button>
      </div>

      <div v-if="loading" class="movie-grid movie-grid--loading" role="status">
        <span class="sr-only">กำลังโหลดภาพยนตร์</span>
        <i v-for="index in 10" :key="index"></i>
      </div>

      <div v-else-if="error" class="catalog-message" role="alert">
        <strong>โหลดรอบฉายไม่สำเร็จ</strong>
        <p>{{ error }}</p>
        <button type="button" class="button button--primary" @click="emit('retry')">
          ลองอีกครั้ง
        </button>
      </div>

      <div v-else-if="items.length" class="movie-grid">
        <MovieCard
          v-for="item in items"
          :key="item.screening.id"
          :screening="item.screening"
          :movie="item.movie"
          :selected="selectedScreeningId === item.screening.id"
          @select="emit('select', $event)"
        />
      </div>

      <div v-else class="catalog-message">
        <strong>ไม่มีรอบฉายตรงกับตัวกรองนี้</strong>
        <p>ลองเลือกประเภทหรือวันที่อื่น</p>
        <button type="button" class="button button--secondary-dark" @click="emit('clearFilters')">
          ล้างตัวกรอง
        </button>
      </div>
    </div>
  </section>
</template>
