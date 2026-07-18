<script setup lang="ts">
import BrandMark from '@/features/movies/components/BrandMark.vue'
</script>

<template>
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
      <span>© 2026 Cinema Ticket Booking System</span>
      <span>Demo cinema booking project</span>
    </div>
  </footer>
</template>
