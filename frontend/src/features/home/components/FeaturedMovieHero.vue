<script setup lang="ts">
import type { MoviePresentation } from '@/features/movies/catalog'
import type { ScreeningSummary } from '@/features/screenings/types'

defineProps<{
  movie: MoviePresentation
  screening: ScreeningSummary | null
  loading: boolean
}>()
const emit = defineEmits<{ select: [screeningID: string] }>()

const dateFormatter = new Intl.DateTimeFormat('th-TH', {
  weekday: 'long',
  day: 'numeric',
  month: 'long',
  timeZone: 'Asia/Bangkok',
})
const timeFormatter = new Intl.DateTimeFormat('th-TH', {
  hour: '2-digit',
  minute: '2-digit',
  timeZone: 'Asia/Bangkok',
})
</script>

<template>
  <section class="featured-banner" aria-labelledby="featured-title">
    <div class="content-shell featured-grid">
      <div class="featured-copy">
        <span class="program-label">โปรแกรมพิเศษ · หนังไทยคัดสรร</span>
        <p class="featured-genre">{{ movie.genres.join(' / ') }}</p>
        <h1 id="featured-title">{{ movie.title }}</h1>
        <p class="featured-english">{{ movie.englishTitle }}</p>
        <p class="featured-description">{{ movie.description }}</p>

        <div class="featured-meta" aria-label="ข้อมูลภาพยนตร์">
          <span>{{ movie.year }}</span>
          <span>{{ screening?.movie.duration_minutes ?? 0 }} นาที</span>
          <span>{{ movie.certificate }}</span>
          <span>{{ movie.language }}</span>
        </div>

        <div class="featured-actions">
          <button
            type="button"
            class="button button--primary"
            :disabled="!screening || loading"
            @click="screening && emit('select', screening.id)"
          >
            เลือกรอบฉาย
          </button>
          <a class="button button--secondary" href="#movies">ดูหนังทั้ง 10 เรื่อง</a>
        </div>
      </div>

      <div class="featured-poster">
        <img :src="movie.poster" :alt="`โปสเตอร์ภาพยนตร์ ${movie.title}`" />
        <div v-if="screening" class="featured-showtime">
          <span>รอบถัดไป</span>
          <strong>{{ timeFormatter.format(new Date(screening.starts_at)) }} น.</strong>
          <small>{{ dateFormatter.format(new Date(screening.starts_at)) }}</small>
          <button type="button" @click="emit('select', screening.id)">
            เลือกที่นั่ง
            <span aria-hidden="true">→</span>
          </button>
        </div>
      </div>
    </div>
  </section>
</template>
