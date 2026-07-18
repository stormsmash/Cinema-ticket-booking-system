<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'

import AuthStatus from '@/features/auth/components/AuthStatus.vue'
import { useAuthStore } from '@/features/auth/store'
import BrandMark from '@/features/movies/components/BrandMark.vue'
import MovieCard from '@/features/movies/components/MovieCard.vue'
import { getMoviePresentation, movieCatalog } from '@/features/movies/catalog'
import MyTickets from '@/features/screenings/components/MyTickets.vue'
import ScreeningPicker from '@/features/screenings/components/ScreeningPicker.vue'
import SeatGrid from '@/features/screenings/components/SeatGrid.vue'
import SeatLockStatus from '@/features/screenings/components/SeatLockStatus.vue'
import { useScreeningStore } from '@/features/screenings/store'

const store = useScreeningStore()
const authStore = useAuthStore()
const { user } = storeToRefs(authStore)
const {
  screenings,
  selectedScreeningID,
  seatMap,
  screeningsError,
  seatsError,
  isLoadingScreenings,
  isLoadingSeats,
  activeLocks,
  isUpdatingLock,
  lockError,
  confirmedBookings,
  isConfirmingBooking,
  bookingError,
  myTickets,
  isLoadingTickets,
  ticketsError,
} = storeToRefs(store)

const allGenres = 'ทั้งหมด'
const allDates = 'all'
const activeGenre = ref(allGenres)
const activeDateKey = ref(allDates)
const bookingSection = ref<HTMLElement | null>(null)

const dateFormatter = new Intl.DateTimeFormat('th-TH', {
  weekday: 'long',
  day: 'numeric',
  month: 'long',
  timeZone: 'Asia/Bangkok',
})
const shortDateFormatter = new Intl.DateTimeFormat('th-TH', {
  weekday: 'short',
  day: 'numeric',
  month: 'short',
  timeZone: 'Asia/Bangkok',
})
const timeFormatter = new Intl.DateTimeFormat('th-TH', {
  hour: '2-digit',
  minute: '2-digit',
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

const selectedScreening = computed(
  () => screenings.value.find((screening) => screening.id === selectedScreeningID.value) ?? null,
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
const selectedMovie = computed(() =>
  getMoviePresentation(selectedScreening.value?.movie.title ?? featuredMovie.value.title),
)
const availableSeatCount = computed(
  () => seatMap.value?.seats.filter((seat) => seat.status === 'AVAILABLE').length ?? 0,
)
const activeSeatIDs = computed(() => activeLocks.value.map((lock) => lock.seat_id))
const selectedTicketPrice = computed(
  () => selectedScreening.value?.ticket_price_baht ?? seatMap.value?.ticket_price_baht ?? 0,
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

function formatDate(value?: string) {
  if (!value) return 'กำลังเตรียมรอบฉาย'
  return dateFormatter.format(new Date(value))
}

function formatTime(value?: string) {
  if (!value) return '--:--'
  return timeFormatter.format(new Date(value))
}

async function chooseScreening(screeningID: string, scrollToSeats = true) {
  if (!screeningID || screeningID === selectedScreeningID.value) {
    if (scrollToSeats) scrollToBooking()
    return
  }

  await store.selectScreening(screeningID)
  if (scrollToSeats) scrollToBooking()
}

async function chooseFeaturedScreening() {
  if (!featuredScreening.value) return
  await chooseScreening(featuredScreening.value.id)
}

async function scrollToBooking() {
  await nextTick()
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  bookingSection.value?.scrollIntoView({
    behavior: reduceMotion ? 'auto' : 'smooth',
    block: 'start',
  })
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
    <header class="site-header">
      <div class="content-shell header-inner">
        <a class="brand-link" href="#top" aria-label="Lumina Cinema หน้าแรก">
          <BrandMark />
        </a>

        <nav class="main-nav" aria-label="เมนูหลัก">
          <a href="#movies">ภาพยนตร์</a>
          <a href="#booking">รอบฉายและที่นั่ง</a>
          <a href="#my-tickets">ตั๋วของฉัน</a>
          <a href="#how-to-book">วิธีจอง</a>
        </nav>

        <AuthStatus />
      </div>
    </header>

    <main id="top">
      <section class="featured-banner" aria-labelledby="featured-title">
        <div class="content-shell featured-grid">
          <div class="featured-copy">
            <span class="program-label">โปรแกรมพิเศษ · หนังไทยคัดสรร</span>
            <p class="featured-genre">{{ featuredMovie.genres.join(' / ') }}</p>
            <h1 id="featured-title">{{ featuredMovie.title }}</h1>
            <p class="featured-english">{{ featuredMovie.englishTitle }}</p>
            <p class="featured-description">{{ featuredMovie.description }}</p>

            <div class="featured-meta" aria-label="ข้อมูลภาพยนตร์">
              <span>{{ featuredMovie.year }}</span>
              <span>{{ featuredScreening?.movie.duration_minutes ?? 0 }} นาที</span>
              <span>{{ featuredMovie.certificate }}</span>
              <span>{{ featuredMovie.language }}</span>
            </div>

            <div class="featured-actions">
              <button
                type="button"
                class="button button--primary"
                :disabled="!featuredScreening || isLoadingScreenings"
                @click="chooseFeaturedScreening"
              >
                เลือกรอบฉาย
              </button>
              <a class="button button--secondary" href="#movies">ดูหนังทั้ง 10 เรื่อง</a>
            </div>
          </div>

          <div class="featured-poster">
            <img :src="featuredMovie.poster" :alt="`โปสเตอร์ภาพยนตร์ ${featuredMovie.title}`" />
            <div v-if="featuredScreening" class="featured-showtime">
              <span>รอบถัดไป</span>
              <strong>{{ formatTime(featuredScreening.starts_at) }} น.</strong>
              <small>{{ formatDate(featuredScreening.starts_at) }}</small>
              <button type="button" @click="chooseFeaturedScreening">
                เลือกที่นั่ง
                <span aria-hidden="true">→</span>
              </button>
            </div>
          </div>
        </div>
      </section>

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
              <strong>LUMINA CINEMA รัชโยธิน</strong>
            </div>
          </div>

          <div class="date-filter" role="toolbar" aria-label="กรองตามวันที่">
            <button
              type="button"
              :class="{ active: activeDateKey === allDates }"
              :aria-pressed="activeDateKey === allDates"
              @click="activeDateKey = allDates"
            >
              ทุกรอบ
            </button>
            <button
              v-for="date in dateOptions"
              :key="date.key"
              type="button"
              :class="{ active: activeDateKey === date.key }"
              :aria-pressed="activeDateKey === date.key"
              @click="activeDateKey = date.key"
            >
              {{ date.label }}
            </button>
          </div>
        </div>
      </section>

      <section id="movies" class="movie-section">
        <div class="content-shell">
          <div class="section-heading">
            <div>
              <p class="eyebrow">THAI MOVIE SELECTION</p>
              <h2>เลือกภาพยนตร์</h2>
              <p>รายการหนังไทยสำหรับทดลองระบบจอง เลือกเรื่องแล้วดูรอบฉายและที่นั่งได้ทันที</p>
            </div>
            <span class="result-count">{{ filteredMovieCards.length }} เรื่อง</span>
          </div>

          <div class="genre-tabs" role="toolbar" aria-label="กรองตามประเภทภาพยนตร์">
            <button
              v-for="genre in genreFilters"
              :key="genre"
              type="button"
              :class="{ active: activeGenre === genre }"
              :aria-pressed="activeGenre === genre"
              @click="activeGenre = genre"
            >
              {{ genre }}
            </button>
          </div>

          <div v-if="isLoadingScreenings" class="movie-grid movie-grid--loading" role="status">
            <span class="sr-only">กำลังโหลดภาพยนตร์</span>
            <i v-for="index in 10" :key="index"></i>
          </div>

          <div v-else-if="screeningsError" class="catalog-message" role="alert">
            <strong>โหลดรอบฉายไม่สำเร็จ</strong>
            <p>{{ screeningsError }}</p>
            <button type="button" class="button button--primary" @click="store.loadScreenings">
              ลองอีกครั้ง
            </button>
          </div>

          <div v-else-if="filteredMovieCards.length" class="movie-grid">
            <MovieCard
              v-for="item in filteredMovieCards"
              :key="item.screening.id"
              :screening="item.screening"
              :movie="item.movie"
              :selected="selectedScreeningID === item.screening.id"
              @select="chooseScreening"
            />
          </div>

          <div v-else class="catalog-message">
            <strong>ไม่มีรอบฉายตรงกับตัวกรองนี้</strong>
            <p>ลองเลือกประเภทหรือวันที่อื่น</p>
            <button
              type="button"
              class="button button--secondary-dark"
              @click="((activeGenre = allGenres), (activeDateKey = allDates))"
            >
              ล้างตัวกรอง
            </button>
          </div>
        </div>
      </section>

      <section
        id="booking"
        ref="bookingSection"
        class="booking-section"
        aria-labelledby="booking-title"
      >
        <div class="content-shell">
          <div class="booking-heading">
            <div>
              <p class="eyebrow eyebrow--red">BOOKING</p>
              <h2 id="booking-title">เลือกรอบและที่นั่ง</h2>
            </div>
            <ol class="booking-steps" aria-label="ขั้นตอนการจอง">
              <li class="done"><span>1</span> เลือกหนัง</li>
              <li class="active"><span>2</span> เลือกที่นั่ง</li>
              <li><span>3</span> ยืนยัน</li>
            </ol>
          </div>

          <div class="booking-workspace">
            <aside class="booking-summary">
              <div v-if="selectedScreening" class="selected-film">
                <img :src="selectedMovie.poster" :alt="`โปสเตอร์ ${selectedMovie.title}`" />
                <div>
                  <span>{{ selectedMovie.genres.join(' · ') }}</span>
                  <h3>{{ selectedMovie.title }}</h3>
                  <p>{{ selectedMovie.englishTitle }}</p>
                </div>
              </div>

              <ScreeningPicker
                :screenings="screenings"
                :model-value="selectedScreeningID"
                :disabled="isLoadingScreenings"
                @update:model-value="(id) => chooseScreening(id, false)"
              />

              <dl v-if="selectedScreening" class="ticket-details">
                <div>
                  <dt>วันที่</dt>
                  <dd>{{ formatDate(selectedScreening.starts_at) }}</dd>
                </div>
                <div>
                  <dt>เวลา</dt>
                  <dd>{{ formatTime(selectedScreening.starts_at) }} น.</dd>
                </div>
                <div>
                  <dt>โรงภาพยนตร์</dt>
                  <dd>{{ selectedScreening.auditorium.name }}</dd>
                </div>
                <div>
                  <dt>ภาษา</dt>
                  <dd>{{ selectedMovie.language }}</dd>
                </div>
                <div>
                  <dt>ราคาต่อที่นั่ง</dt>
                  <dd>฿{{ selectedTicketPrice.toLocaleString('th-TH') }}</dd>
                </div>
              </dl>

              <div class="seat-availability">
                <span><i></i> ที่นั่งว่าง</span>
                <strong>{{ availableSeatCount }}</strong>
              </div>

              <p class="booking-note">
                ที่นั่งจะถูกพักไว้ 5 นาทีหลังเลือก กรุณาตรวจสอบรอบฉายก่อนยืนยันการจอง
              </p>
            </aside>

            <section class="seat-panel" aria-labelledby="seats-title">
              <div class="seat-panel__heading">
                <div>
                  <p>{{ selectedScreening?.auditorium.name ?? 'โรงภาพยนตร์' }}</p>
                  <h3 id="seats-title">ผังที่นั่ง</h3>
                </div>
                <span v-if="seatMap">
                  {{
                    activeLocks.length
                      ? `เลือกแล้ว ${activeLocks.length} / 6`
                      : `${seatMap.seats.length} ที่นั่ง`
                  }}
                </span>
              </div>

              <div v-if="isLoadingSeats" class="seat-skeleton" role="status">
                <span>กำลังโหลดผังที่นั่ง...</span>
                <i v-for="index in 40" :key="index"></i>
              </div>

              <div v-else-if="seatsError" class="seat-error" role="alert">
                <strong>เปิดผังที่นั่งไม่สำเร็จ</strong>
                <p>{{ seatsError }}</p>
                <button type="button" class="button button--primary" @click="store.reloadSeatMap">
                  ลองอีกครั้ง
                </button>
              </div>

              <template v-else-if="seatMap">
                <SeatGrid
                  :seat-map="seatMap"
                  :can-lock="Boolean(user)"
                  :active-seat-ids="activeSeatIDs"
                  :max-selectable="store.maxSeatsPerBooking"
                  :is-updating-lock="isUpdatingLock"
                  @toggle="store.toggleSeatLock"
                />
                <SeatLockStatus
                  :locks="activeLocks"
                  :signed-in="Boolean(user)"
                  :is-updating="isUpdatingLock"
                  :error="lockError"
                  :bookings="confirmedBookings"
                  :unit-price-baht="selectedTicketPrice"
                  :is-confirming="isConfirmingBooking"
                  :booking-error="bookingError"
                  @release-all="store.unlockAllSeats"
                  @expired="store.handleLockExpired"
                  @confirm="store.confirmBooking"
                />
              </template>

              <div v-else class="seat-error">
                <p>เลือกภาพยนตร์เพื่อดูผังที่นั่ง</p>
              </div>
            </section>
          </div>
        </div>
      </section>

      <MyTickets
        :tickets="myTickets"
        :signed-in="Boolean(user)"
        :loading="isLoadingTickets"
        :error="ticketsError"
      />

      <section id="how-to-book" class="how-to-book">
        <div class="content-shell how-to-grid">
          <div>
            <p class="eyebrow">HOW TO BOOK</p>
            <h2>จองได้ใน 3 ขั้นตอน</h2>
          </div>
          <ol>
            <li>
              <span>1</span><strong>เลือกหนังและรอบ</strong
              ><small>ตรวจสอบวันที่ เวลา และโรงฉาย</small>
            </li>
            <li>
              <span>2</span><strong>เข้าสู่ระบบและเลือกที่นั่ง</strong
              ><small>ระบบพักที่นั่งไว้ให้ 5 นาที</small>
            </li>
            <li>
              <span>3</span><strong>ยืนยันการจอง</strong
              ><small>รับหมายเลขอ้างอิงหลังทำรายการ</small>
            </li>
          </ol>
        </div>
      </section>
    </main>

    <footer class="site-footer">
      <div class="content-shell footer-grid">
        <div class="footer-about">
          <BrandMark />
          <p>ระบบสาธิตการจองตั๋วภาพยนตร์แบบเรียลไทม์</p>
        </div>
        <div class="footer-links">
          <strong>เมนู</strong>
          <a href="#movies">ภาพยนตร์</a>
          <a href="#booking">รอบฉายและที่นั่ง</a>
          <a href="#my-tickets">ตั๋วของฉัน</a>
          <a href="#how-to-book">วิธีจอง</a>
        </div>
      </div>
      <div class="content-shell footer-bottom">
        <span>© 2026 LUMINA CINEMA</span>
        <span>Demo cinema booking project</span>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.cinema-app {
  min-height: 100vh;
  color: #222226;
  background: #f5f5f6;
}

.content-shell {
  width: min(78rem, calc(100% - 3rem));
  margin: 0 auto;
}

.site-header {
  position: sticky;
  z-index: 40;
  top: 0;
  border-bottom: 1px solid #29292e;
  color: #fff;
  background: #111113;
}

.header-inner {
  display: grid;
  grid-template-columns: auto 1fr auto;
  min-height: 4.5rem;
  align-items: center;
  gap: 2rem;
}

.brand-link {
  display: inline-flex;
  text-decoration: none;
}

.main-nav {
  display: flex;
  justify-content: center;
  gap: 2rem;
}

.main-nav a {
  color: #d0d0d4;
  font-size: 0.83rem;
  font-weight: 650;
  text-decoration: none;
}

.main-nav a:hover {
  color: #fff;
}

.featured-banner {
  color: #fff;
  background: #1a1a1e;
}

.featured-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(19rem, 26rem);
  min-height: 33rem;
  gap: clamp(3rem, 8vw, 8rem);
  align-items: center;
  padding-top: 3.5rem;
  padding-bottom: 3.5rem;
}

.featured-copy {
  max-width: 42rem;
}

.program-label,
.eyebrow {
  color: #d91920;
  font-size: 0.73rem;
  font-weight: 800;
  letter-spacing: 0.1em;
}

.program-label {
  display: inline-flex;
  padding: 0.42rem 0.65rem;
  border-radius: 0.2rem;
  color: #fff;
  background: #d91920;
  letter-spacing: 0;
}

.featured-genre {
  margin: 1.7rem 0 0;
  color: #c9c9ce;
  font-size: 0.8rem;
  font-weight: 700;
}

.featured-copy h1 {
  margin: 0.35rem 0 0;
  font-size: clamp(3.4rem, 7vw, 6.6rem);
  line-height: 1.02;
}

.featured-english {
  margin: 0.65rem 0 0;
  color: #a8a8af;
  font-size: 1rem;
}

.featured-description {
  max-width: 38rem;
  margin: 1.4rem 0 0;
  color: #dedee1;
  font-size: 1rem;
  line-height: 1.75;
}

.featured-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem 1.2rem;
  margin-top: 1.5rem;
  color: #b9b9be;
  font-size: 0.78rem;
}

.featured-meta span + span::before {
  margin-right: 1.2rem;
  color: #55555d;
  content: '•';
}

.featured-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 2rem;
}

.button {
  display: inline-flex;
  min-height: 2.85rem;
  align-items: center;
  justify-content: center;
  padding: 0.68rem 1.15rem;
  border: 1px solid transparent;
  border-radius: 0.25rem;
  font-size: 0.8rem;
  font-weight: 750;
  text-decoration: none;
  cursor: pointer;
}

.button--primary {
  border-color: #d91920;
  color: #fff;
  background: #d91920;
}

.button--primary:hover {
  border-color: #b80e15;
  background: #b80e15;
}

.button--secondary {
  border-color: #56565d;
  color: #fff;
  background: transparent;
}

.button--secondary-dark {
  border-color: #c6c6ca;
  color: #2d2d31;
  background: #fff;
}

.button:disabled {
  cursor: wait;
  opacity: 0.55;
}

.featured-poster {
  width: min(100%, 23rem);
  overflow: hidden;
  border: 1px solid #303036;
  border-radius: 0.4rem;
  background: #111115;
  box-shadow: 0 1.5rem 3.5rem rgb(0 0 0 / 35%);
  justify-self: end;
}

.featured-poster img {
  display: block;
  width: 100%;
  aspect-ratio: 2 / 3;
  object-fit: cover;
}

.featured-showtime {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  grid-template-rows: repeat(3, auto);
  gap: 0.12rem 1rem;
  align-items: center;
  padding: 0.9rem 1rem 1rem;
  border-top: 2px solid #d91920;
  color: #fff;
  background: #111115;
}

.featured-showtime span,
.featured-showtime small {
  color: #a9a9b0;
  font-size: 0.66rem;
}

.featured-showtime strong {
  margin: 0.18rem 0;
  font-size: 1.5rem;
  font-variant-numeric: tabular-nums;
}

.featured-showtime > span,
.featured-showtime > strong,
.featured-showtime > small {
  grid-column: 1;
}

.featured-showtime button {
  display: inline-flex;
  grid-column: 2;
  grid-row: 1 / 4;
  min-height: 2.6rem;
  gap: 0.65rem;
  align-items: center;
  justify-content: center;
  padding: 0.65rem 0.8rem;
  border: 1px solid #d91920;
  border-radius: 0.15rem;
  color: #fff;
  background: #d91920;
  font: inherit;
  font-size: 0.72rem;
  font-weight: 750;
  cursor: pointer;
}

.featured-showtime button:hover {
  border-color: #ed2b32;
  background: #ed2b32;
}

.featured-showtime button span {
  color: inherit;
  font-size: 0.9rem;
}

.schedule-bar {
  border-bottom: 1px solid #dedee1;
  background: #fff;
}

.schedule-inner {
  display: grid;
  grid-template-columns: minmax(16rem, 0.8fr) minmax(0, 1.2fr);
  gap: 2rem;
  align-items: center;
  padding-top: 1.1rem;
  padding-bottom: 1.1rem;
}

.branch-info {
  display: flex;
  align-items: center;
  gap: 0.8rem;
}

.schedule-icon {
  display: grid;
  width: 2.4rem;
  height: 2.4rem;
  flex: 0 0 auto;
  border-radius: 50%;
  place-items: center;
  color: #d91920;
  background: #fce8e9;
}

.schedule-icon svg {
  width: 1.2rem;
  fill: none;
  stroke: currentcolor;
  stroke-width: 1.8;
}

.branch-info div {
  display: grid;
}

.branch-info small {
  color: #77777e;
  font-size: 0.68rem;
}

.branch-info strong {
  margin-top: 0.12rem;
  font-size: 0.85rem;
}

.date-filter {
  display: flex;
  justify-content: flex-end;
  overflow-x: auto;
}

.date-filter button,
.genre-tabs button {
  flex: 0 0 auto;
  border: 1px solid #d7d7da;
  color: #57575e;
  background: #fff;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
}

.date-filter button {
  min-height: 2.55rem;
  padding: 0.55rem 0.8rem;
  border-right: 0;
}

.date-filter button:first-child {
  border-radius: 0.25rem 0 0 0.25rem;
}

.date-filter button:last-child {
  border-right: 1px solid #d7d7da;
  border-radius: 0 0.25rem 0.25rem 0;
}

.date-filter button.active,
.genre-tabs button.active {
  border-color: #d91920;
  color: #fff;
  background: #d91920;
}

.movie-section {
  padding: 4.5rem 0 5rem;
}

.section-heading,
.booking-heading {
  display: flex;
  justify-content: space-between;
  gap: 2rem;
  align-items: flex-end;
}

.eyebrow {
  margin: 0 0 0.5rem;
  color: #77777e;
}

.eyebrow--red {
  color: #d91920;
}

.section-heading h2,
.booking-heading h2,
.how-to-grid h2 {
  margin: 0;
  color: #1e1e22;
  font-size: clamp(1.9rem, 4vw, 2.7rem);
  line-height: 1.15;
}

.section-heading > div > p:last-child {
  max-width: 40rem;
  margin: 0.75rem 0 0;
  color: #6e6e75;
  font-size: 0.88rem;
  line-height: 1.7;
}

.result-count {
  flex: 0 0 auto;
  color: #77777e;
  font-size: 0.82rem;
  font-weight: 700;
}

.genre-tabs {
  display: flex;
  gap: 0.5rem;
  overflow-x: auto;
  margin: 2rem 0 1.6rem;
  padding-bottom: 0.25rem;
}

.genre-tabs button {
  min-height: 2.35rem;
  padding: 0.5rem 0.8rem;
  border-radius: 999px;
}

.movie-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 2rem 1.1rem;
}

.movie-grid--loading i {
  display: block;
  aspect-ratio: 2 / 3.75;
  border-radius: 0.3rem;
  background: #e2e2e5;
  animation: pulse 1.2s ease-in-out infinite alternate;
}

.catalog-message,
.seat-error {
  padding: 2rem;
  border: 1px solid #dedee1;
  text-align: center;
  background: #fff;
}

.catalog-message p,
.seat-error p {
  margin: 0.45rem 0 1.2rem;
  color: #717178;
  font-size: 0.82rem;
}

.booking-section {
  padding: 4.5rem 0 5rem;
  color: #e9e9ec;
  background: #151518;
  scroll-margin-top: 4.5rem;
}

.booking-heading h2 {
  color: #fff;
}

.booking-steps {
  display: flex;
  gap: 1rem;
  margin: 0;
  padding: 0;
  list-style: none;
  color: #77777f;
  font-size: 0.72rem;
}

.booking-steps li {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.booking-steps span {
  display: grid;
  width: 1.6rem;
  height: 1.6rem;
  border: 1px solid #4e4e55;
  border-radius: 50%;
  place-items: center;
  font-weight: 800;
}

.booking-steps .done,
.booking-steps .active {
  color: #fff;
}

.booking-steps .done span,
.booking-steps .active span {
  border-color: #d91920;
  background: #d91920;
}

.booking-workspace {
  display: grid;
  grid-template-columns: 20rem minmax(0, 1fr);
  margin-top: 2rem;
  border: 1px solid #333338;
  background: #1d1d21;
}

.booking-summary {
  padding: 1.5rem;
  border-right: 1px solid #333338;
  background: #19191d;
}

.selected-film {
  display: grid;
  grid-template-columns: 5rem 1fr;
  gap: 0.9rem;
  margin-bottom: 1.5rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid #333338;
}

.selected-film img {
  width: 5rem;
  aspect-ratio: 2 / 3;
  object-fit: cover;
}

.selected-film div {
  min-width: 0;
}

.selected-film span {
  color: #d91920;
  font-size: 0.66rem;
  font-weight: 750;
}

.selected-film h3 {
  margin: 0.35rem 0 0;
  color: #fff;
  font-size: 1.1rem;
  line-height: 1.3;
}

.selected-film p {
  margin: 0.3rem 0 0;
  color: #888890;
  font-size: 0.68rem;
  line-height: 1.45;
}

.ticket-details {
  display: grid;
  gap: 0;
  margin: 1.3rem 0 0;
}

.ticket-details div {
  display: grid;
  grid-template-columns: 5.8rem 1fr;
  gap: 0.75rem;
  padding: 0.68rem 0;
  border-bottom: 1px solid #303035;
}

.ticket-details dt {
  color: #7f7f87;
  font-size: 0.68rem;
}

.ticket-details dd {
  margin: 0;
  color: #e3e3e6;
  font-size: 0.73rem;
  font-weight: 650;
}

.seat-availability {
  display: flex;
  justify-content: space-between;
  margin-top: 1.3rem;
  padding: 0.85rem;
  align-items: center;
  color: #a8a8af;
  background: #242429;
  font-size: 0.72rem;
}

.seat-availability span {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.seat-availability i {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 50%;
  background: #43a66e;
}

.seat-availability strong {
  color: #fff;
  font-size: 1.05rem;
}

.booking-note {
  margin: 1rem 0 0;
  color: #73737b;
  font-size: 0.68rem;
  line-height: 1.65;
}

.seat-panel {
  min-width: 0;
  padding: 1.8rem;
}

.seat-panel__heading {
  display: flex;
  justify-content: space-between;
  margin-bottom: 2rem;
  align-items: flex-end;
}

.seat-panel__heading p {
  margin: 0 0 0.25rem;
  color: #d91920;
  font-size: 0.67rem;
  font-weight: 800;
}

.seat-panel__heading h3 {
  margin: 0;
  color: #fff;
  font-size: 1.5rem;
}

.seat-panel__heading > span {
  color: #7f7f87;
  font-size: 0.72rem;
}

.seat-skeleton {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 0.5rem;
}

.seat-skeleton span {
  grid-column: 1 / -1;
  color: #888890;
  font-size: 0.75rem;
}

.seat-skeleton i {
  display: block;
  min-height: 2.2rem;
  background: #29292e;
  animation: pulse-dark 1.2s ease-in-out infinite alternate;
}

.seat-error {
  border-color: #3b3b40;
  color: #fff;
  background: #202025;
}

.how-to-book {
  padding: 4rem 0;
  border-bottom: 1px solid #dedee1;
  background: #fff;
}

.how-to-grid {
  display: grid;
  grid-template-columns: 0.65fr 1.35fr;
  gap: 4rem;
  align-items: start;
}

.how-to-grid ol {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.how-to-grid li {
  display: grid;
  padding-left: 1rem;
  border-left: 2px solid #d91920;
}

.how-to-grid li > span {
  color: #d91920;
  font-size: 0.72rem;
  font-weight: 850;
}

.how-to-grid strong {
  margin-top: 0.55rem;
  font-size: 0.86rem;
}

.how-to-grid small {
  margin-top: 0.3rem;
  color: #77777e;
  font-size: 0.7rem;
  line-height: 1.55;
}

.site-footer {
  padding: 3rem 0 1.5rem;
  color: #c7c7cc;
  background: #111113;
}

.footer-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 3rem;
}

.footer-about p {
  margin: 1rem 0 0.2rem;
  color: #d0d0d4;
  font-size: 0.78rem;
}

.footer-links {
  display: grid;
  align-content: start;
  gap: 0.65rem;
}

.footer-links strong {
  margin-bottom: 0.25rem;
  color: #fff;
  font-size: 0.75rem;
}

.footer-links a {
  color: #8e8e96;
  font-size: 0.72rem;
  text-decoration: none;
}

.footer-bottom {
  display: flex;
  justify-content: space-between;
  margin-top: 2.5rem;
  padding-top: 1rem;
  border-top: 1px solid #29292e;
  color: #62626a;
  font-size: 0.64rem;
}

@keyframes pulse {
  to {
    background: #d3d3d7;
  }
}

@keyframes pulse-dark {
  to {
    background: #34343a;
  }
}

@media (max-width: 1020px) {
  .content-shell {
    width: min(100% - 2rem, 78rem);
  }

  .featured-grid {
    grid-template-columns: minmax(0, 1fr) 19rem;
    gap: 3rem;
  }

  .movie-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .booking-workspace {
    grid-template-columns: 18rem minmax(0, 1fr);
  }

  .how-to-grid {
    grid-template-columns: 1fr;
    gap: 2rem;
  }
}

@media (max-width: 780px) {
  .header-inner {
    grid-template-columns: 1fr auto;
  }

  .main-nav {
    display: none;
  }

  .featured-grid {
    grid-template-columns: minmax(0, 1fr) 13rem;
    min-height: auto;
    gap: 2rem;
  }

  .featured-copy h1 {
    font-size: clamp(3rem, 9vw, 4.5rem);
  }

  .featured-description {
    font-size: 0.88rem;
  }

  .featured-showtime {
    grid-template-columns: 1fr;
  }

  .featured-showtime button {
    grid-column: 1;
    grid-row: auto;
    margin-top: 0.55rem;
  }

  .schedule-inner {
    grid-template-columns: 1fr;
    gap: 1rem;
  }

  .date-filter {
    justify-content: flex-start;
  }

  .movie-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .booking-heading {
    display: grid;
  }

  .booking-workspace {
    grid-template-columns: 1fr;
  }

  .booking-summary {
    border-right: 0;
    border-bottom: 1px solid #333338;
  }

  .footer-grid {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 560px) {
  .content-shell {
    width: min(100% - 1.25rem, 78rem);
  }

  .featured-grid {
    grid-template-columns: 1fr;
    padding-top: 2rem;
    padding-bottom: 2rem;
  }

  .featured-poster {
    display: grid;
    grid-template-columns: 8.5rem minmax(0, 1fr);
    width: 100%;
    gap: 0;
    justify-self: start;
  }

  .featured-poster img {
    height: 100%;
  }

  .featured-showtime {
    align-content: center;
    border-top: 0;
    border-left: 2px solid #d91920;
  }

  .featured-meta span + span::before {
    margin-right: 0.7rem;
  }

  .featured-meta {
    gap: 0.35rem 0.7rem;
  }

  .featured-actions .button {
    flex: 1 1 9rem;
  }

  .section-heading {
    align-items: start;
  }

  .movie-section,
  .booking-section {
    padding-top: 3.5rem;
    padding-bottom: 3.5rem;
  }

  .movie-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 1.7rem 0.75rem;
  }

  .booking-steps {
    gap: 0.5rem;
    font-size: 0.64rem;
  }

  .booking-steps li {
    gap: 0.3rem;
  }

  .booking-steps span {
    width: 1.4rem;
    height: 1.4rem;
  }

  .seat-panel,
  .booking-summary {
    padding: 1rem;
  }

  .how-to-grid ol {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }

  .footer-grid {
    grid-template-columns: 1fr;
    gap: 2rem;
  }

  .footer-bottom {
    display: grid;
    gap: 0.35rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .movie-grid--loading i,
  .seat-skeleton i {
    animation: none;
  }
}
</style>
