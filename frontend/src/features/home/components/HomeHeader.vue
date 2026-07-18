<script setup lang="ts">
import AuthStatus from '@/features/auth/components/AuthStatus.vue'
import BrandMark from '@/features/movies/components/BrandMark.vue'
</script>

<template>
  <header class="site-header">
    <div class="content-shell header-inner">
      <a class="brand-link" href="#top" aria-label="Cinema Ticket Booking หน้าแรก">
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
</template>
