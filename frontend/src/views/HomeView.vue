<script setup lang="ts">
import SystemStatus from '@/features/system/components/SystemStatus.vue'
</script>

<template>
  <main class="home-shell">
    <div class="home-grid">
      <section class="intro">
        <p class="eyebrow">Cinema Ticket Booking</p>
        <h1>Choose your seat.<br />Keep the moment.</h1>
        <p class="summary">
          The system foundation is ready. Authentication, screenings, and real-time seat selection
          will be added one verified step at a time.
        </p>
      </section>

      <aside>
        <SystemStatus />
      </aside>
    </div>
  </main>
</template>

<style scoped>
.home-shell {
  display: grid;
  min-height: 100vh;
  place-items: center;
  padding: 2rem;
}

.home-grid {
  display: grid;
  width: min(68rem, 100%);
  grid-template-columns: minmax(0, 1.4fr) minmax(17rem, 0.6fr);
  gap: 4rem;
  align-items: end;
}

.eyebrow {
  margin: 0 0 1rem;
  color: #fbbf24;
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

h1 {
  max-width: 12ch;
  margin: 0;
  font-size: clamp(3rem, 8vw, 6.5rem);
  line-height: 0.94;
  letter-spacing: -0.065em;
}

.summary {
  max-width: 39rem;
  margin: 1.75rem 0 0;
  color: #a1a1aa;
  font-size: 1.05rem;
  line-height: 1.75;
}

@media (max-width: 760px) {
  .home-grid {
    grid-template-columns: 1fr;
    gap: 2.5rem;
  }
}
</style>
