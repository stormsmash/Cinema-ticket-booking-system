<script setup lang="ts">
defineProps<{
  bookingTotal: number
  bookedOnPage: number
  auditTotal: number
  successEventsOnPage: number
}>()
</script>

<template>
  <header class="admin-hero">
    <div>
      <p class="section-label">ADMIN DASHBOARD</p>
      <h1>จัดการรายการจอง</h1>
      <p class="header-summary">
        ตรวจสอบรายการจอง ค้นหาข้อมูลลูกค้า และติดตามเหตุการณ์สำคัญของระบบ
      </p>
    </div>
    <div class="live-chip"><i></i> พร้อมใช้งาน</div>
  </header>

  <section class="metric-grid" aria-label="ภาพรวมระบบ">
    <article class="metric-card metric-card--accent">
      <span>รายการจองทั้งหมด</span>
      <strong>{{ bookingTotal }}</strong>
      <small>ทุกรายการในระบบ</small>
    </article>
    <article class="metric-card">
      <span>สำเร็จในหน้านี้</span>
      <strong>{{ bookedOnPage }}</strong>
      <small>จากหน้าปัจจุบัน</small>
    </article>
    <article class="metric-card">
      <span>เหตุการณ์ระบบ</span>
      <strong>{{ auditTotal }}</strong>
      <small>เหตุการณ์ที่บันทึกไว้</small>
    </article>
    <article class="metric-card">
      <span>จองสำเร็จในหน้านี้</span>
      <strong>{{ successEventsOnPage }}</strong>
      <small>จากหน้าปัจจุบัน</small>
    </article>
  </section>
</template>
