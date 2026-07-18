<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'

import { useAuthStore } from '@/features/auth/store'
import BookingSection from '@/features/home/components/BookingSection.vue'
import FeaturedMovieHero from '@/features/home/components/FeaturedMovieHero.vue'
import HomeFooter from '@/features/home/components/HomeFooter.vue'
import HomeHeader from '@/features/home/components/HomeHeader.vue'
import HowToBookSection from '@/features/home/components/HowToBookSection.vue'
import MovieCatalogSection from '@/features/home/components/MovieCatalogSection.vue'
import ScheduleBar from '@/features/home/components/ScheduleBar.vue'
import { getMoviePresentation, movieCatalog } from '@/features/movies/catalog'
import MyTickets from '@/features/screenings/components/MyTickets.vue'
import { useScreeningStore } from '@/features/screenings/store'
import './home-view.css'

const store = useScreeningStore()
const authStore = useAuthStore()
const { user } = storeToRefs(authStore)
const {
  screenings,
  selectedScreeningID,
  screeningsError,
  isLoadingScreenings,
  myTickets,
  isLoadingTickets,
  ticketsError,
} = storeToRefs(store)

const allGenres = 'ทั้งหมด'
const allDates = 'all'
const activeGenre = ref(allGenres)
const activeDateKey = ref(allDates)

const shortDateFormatter = new Intl.DateTimeFormat('th-TH', {
  weekday: 'short',
  day: 'numeric',
  month: 'short',
  timeZone: 'Asia/Bangkok',
})

const movieCards = computed(() => {
  const screeningByTitle = new Map(
    screenings.value.map((screening) => [screening.movie.title, screening]),
  )
  return movieCatalog.flatMap((movie) => {
    const screening = screeningByTitle.get(movie.title)
    return screening ? [{ movie, screening }] : []
  })
})

const genreFilters = computed(() => {
  const genres = new Set(movieCards.value.flatMap(({ movie }) => movie.genres))
  return [allGenres, ...Array.from(genres)]
})

const dateOptions = computed(() => {
  const uniqueDates = new Map<string, Date>()
  for (const screening of screenings.value) {
    const date = new Date(screening.starts_at)
    uniqueDates.set(bangkokDateKey(date), date)
  }
  return Array.from(uniqueDates, ([key, date]) => ({
    key,
    label: shortDateFormatter.format(date),
  }))
})

const filteredMovieCards = computed(() =>
  movieCards.value.filter(({ movie, screening }) => {
    const matchesGenre = activeGenre.value === allGenres || movie.genres.includes(activeGenre.value)
    const matchesDate =
      activeDateKey.value === allDates ||
      bangkokDateKey(new Date(screening.starts_at)) === activeDateKey.value
    return matchesGenre && matchesDate
  }),
)

const featuredScreening = computed(
  () =>
    screenings.value.find((screening) => getMoviePresentation(screening.movie.title).featured) ??
    screenings.value[0] ??
    null,
)
const featuredMovie = computed(() =>
  getMoviePresentation(featuredScreening.value?.movie.title ?? movieCatalog[0]!.title),
)

function bangkokDateKey(date: Date) {
  const parts = new Intl.DateTimeFormat('en-CA', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    timeZone: 'Asia/Bangkok',
  }).formatToParts(date)
  const values = Object.fromEntries(parts.map((part) => [part.type, part.value]))
  return `${values.year}-${values.month}-${values.day}`
}

async function chooseScreening(screeningID: string) {
  if (screeningID && screeningID !== selectedScreeningID.value) {
    await store.selectScreening(screeningID)
  }
  await nextTick()
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  document.getElementById('booking')?.scrollIntoView({
    behavior: reduceMotion ? 'auto' : 'smooth',
    block: 'start',
  })
}

function clearFilters() {
  activeGenre.value = allGenres
  activeDateKey.value = allDates
}

onMounted(store.loadScreenings)
onBeforeUnmount(store.stopRealtime)
watch(
  user,
  (currentUser) => {
    if (currentUser) void store.loadMyTickets()
    else store.clearTickets()
  },
  { immediate: true },
)
</script>

<template>
  <div class="cinema-app">
    <HomeHeader />

    <main id="top">
      <FeaturedMovieHero
        :movie="featuredMovie"
        :screening="featuredScreening"
        :loading="isLoadingScreenings"
        @select="chooseScreening"
      />

      <ScheduleBar v-model="activeDateKey" :dates="dateOptions" :all-dates-value="allDates" />

      <MovieCatalogSection
        v-model:active-genre="activeGenre"
        :items="filteredMovieCards"
        :genres="genreFilters"
        :selected-screening-id="selectedScreeningID"
        :loading="isLoadingScreenings"
        :error="screeningsError"
        @select="chooseScreening"
        @retry="store.loadScreenings"
        @clear-filters="clearFilters"
      />

      <BookingSection />

      <MyTickets
        :tickets="myTickets"
        :signed-in="Boolean(user)"
        :loading="isLoadingTickets"
        :error="ticketsError"
      />

      <HowToBookSection />
    </main>

    <HomeFooter />
  </div>
</template>
