<script setup lang="ts">
import type { ScreeningSummary } from '@/features/screenings/types'
import type { MoviePresentation } from '../catalog'

defineProps<{
  screening: ScreeningSummary
  movie: MoviePresentation
  selected: boolean
}>()

const emit = defineEmits<{
  select: [screeningID: string]
}>()

const timeFormatter = new Intl.DateTimeFormat('th-TH', {
  hour: '2-digit',
  minute: '2-digit',
  timeZone: 'Asia/Bangkok',
})
</script>

<template>
  <article class="movie-card" :class="{ 'movie-card--selected': selected }">
    <button
      type="button"
      class="movie-card__button"
      :aria-pressed="selected"
      :aria-label="`เลือก ${movie.title} รอบ ${timeFormatter.format(new Date(screening.starts_at))} น.`"
      @click="emit('select', screening.id)"
    >
      <span class="poster-wrap">
        <img :src="movie.poster" :alt="`โปสเตอร์ภาพยนตร์ ${movie.title}`" loading="lazy" />
        <span class="certificate">{{ movie.certificate }}</span>
      </span>

      <span class="movie-card__body">
        <span class="movie-card__genres">{{ movie.genres.join(' · ') }}</span>
        <strong>{{ movie.title }}</strong>
        <span class="movie-card__original">{{ movie.englishTitle }} · {{ movie.year }}</span>
        <span class="movie-card__meta">
          <span>{{ screening.movie.duration_minutes }} นาที</span>
          <span>รอบ {{ timeFormatter.format(new Date(screening.starts_at)) }} น.</span>
        </span>
        <span class="movie-card__action">
          {{
            selected
              ? 'เลือกเรื่องนี้แล้ว'
              : `฿${screening.ticket_price_baht.toLocaleString('th-TH')} · ดูรอบฉาย`
          }}
          <svg viewBox="0 0 20 20" aria-hidden="true"><path d="m7 4 6 6-6 6" /></svg>
        </span>
      </span>
    </button>
  </article>
</template>

<style scoped>
.movie-card {
  min-width: 0;
  border: 1px solid #dedee1;
  border-radius: 0.3rem;
  background: #fff;
}

.movie-card__button {
  display: grid;
  width: 100%;
  padding: 0;
  border: 0;
  color: inherit;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.poster-wrap {
  position: relative;
  display: block;
  overflow: hidden;
  aspect-ratio: 2 / 3;
  background: #e2e2e5;
}

.poster-wrap img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.certificate {
  position: absolute;
  top: 0.6rem;
  left: 0.6rem;
  padding: 0.25rem 0.42rem;
  border-radius: 0.15rem;
  color: #fff;
  background: rgb(17 17 19 / 88%);
  font-size: 0.62rem;
  font-weight: 800;
}

.movie-card__body {
  display: grid;
  min-height: 12rem;
  padding: 0.85rem;
}

.movie-card__genres {
  color: #d91920;
  font-size: 0.66rem;
  font-weight: 750;
}

.movie-card__body > strong {
  margin-top: 0.35rem;
  overflow: hidden;
  color: #242428;
  font-size: 1rem;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.movie-card__original {
  margin-top: 0.18rem;
  overflow: hidden;
  color: #8a8a91;
  font-size: 0.66rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.movie-card__meta {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
  margin-top: 0.85rem;
  padding-top: 0.7rem;
  border-top: 1px solid #ececef;
  color: #66666d;
  font-size: 0.67rem;
  font-variant-numeric: tabular-nums;
}

.movie-card__action {
  display: flex;
  justify-content: space-between;
  margin-top: auto;
  padding-top: 0.85rem;
  align-items: center;
  color: #d91920;
  font-size: 0.72rem;
  font-weight: 800;
}

.movie-card__action svg {
  width: 1rem;
  fill: none;
  stroke: currentcolor;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 1.8;
}

.movie-card:hover,
.movie-card:focus-within,
.movie-card--selected {
  border-color: #d91920;
}

.movie-card__button:focus-visible {
  outline: 3px solid #d91920;
  outline-offset: 3px;
}

.movie-card--selected .movie-card__action {
  color: #fff;
}

.movie-card--selected .movie-card__action {
  margin: 0.65rem -0.85rem -0.85rem;
  padding: 0.7rem 0.85rem;
  background: #d91920;
}

@media (max-width: 560px) {
  .movie-card__body {
    min-height: 10.8rem;
    padding: 0.7rem;
  }

  .movie-card__meta {
    display: grid;
    gap: 0.15rem;
  }

  .movie-card--selected .movie-card__action {
    margin-right: -0.7rem;
    margin-bottom: -0.7rem;
    margin-left: -0.7rem;
    padding-right: 0.7rem;
    padding-left: 0.7rem;
  }
}
</style>
