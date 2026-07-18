<script setup lang="ts">
import BrandMark from '@/features/movies/components/BrandMark.vue'
</script>

<template>
  <header class="admin-topbar">
    <div class="topbar-shell">
      <a href="/" class="brand-link" aria-label="Cinema Ticket Booking หน้าแรก"><BrandMark /></a>
      <span class="admin-divider"></span>
      <span class="admin-product">ADMIN</span>
      <a class="back-link" href="/">
        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m15 18-6-6 6-6" /></svg>
        กลับไปหน้าจองบัตร
      </a>
    </div>
  </header>
</template>
