<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'

import { useAuthStore } from '../store'

const store = useAuthStore()
const { user, googleEnabled, isLoading, error } = storeToRefs(store)

onMounted(store.load)
</script>

<template>
  <details class="auth-card">
    <summary>
      <img v-if="user?.avatar_url" :src="user.avatar_url" alt="" referrerpolicy="no-referrer" />
      <span v-else class="avatar-fallback" aria-hidden="true">
        {{ user?.name.charAt(0) || 'L' }}
      </span>
      <span class="account-label">
        <small>{{ user ? 'สมาชิก LUMINA' : 'บัญชีของฉัน' }}</small>
        <strong>{{ isLoading ? 'กำลังตรวจสอบ...' : user?.name || 'เข้าสู่ระบบ' }}</strong>
      </span>
      <svg viewBox="0 0 20 20" aria-hidden="true"><path d="m6 8 4 4 4-4" /></svg>
    </summary>

    <div class="account-menu" aria-live="polite">
      <template v-if="user">
        <div class="profile-detail">
          <span>เข้าสู่ระบบด้วย</span>
          <strong>{{ user.email }}</strong>
        </div>
        <a v-if="user.role === 'ADMIN'" class="admin-link" href="/admin">หน้าจัดการระบบ</a>
        <button type="button" class="secondary-button" @click="store.logout">ออกจากระบบ</button>
      </template>

      <template v-else-if="!isLoading">
        <p>เข้าสู่ระบบก่อนเลือกและยืนยันที่นั่ง</p>
        <a v-if="googleEnabled" class="google-button" href="/api/v1/auth/google">
          เข้าสู่ระบบด้วย Google
        </a>
        <button v-else type="button" class="google-button" disabled>ยังไม่เปิดใช้งาน Google</button>
        <small v-if="!googleEnabled">กรุณาตั้งค่า Google OAuth ในไฟล์ .env</small>
      </template>

      <p v-if="error" class="auth-error" role="alert">{{ error }}</p>
    </div>
  </details>
</template>

<style scoped>
.auth-card {
  position: relative;
}

summary {
  display: flex;
  min-width: 12.5rem;
  align-items: center;
  gap: 0.65rem;
  padding: 0.45rem 0;
  list-style: none;
  cursor: pointer;
}

summary::-webkit-details-marker {
  display: none;
}

summary img,
.avatar-fallback {
  display: grid;
  width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  border: 1px solid #e04247;
  border-radius: 50%;
  place-items: center;
  color: #fff;
  background: #d91920;
  object-fit: cover;
  font-size: 0.72rem;
  font-weight: 850;
}

.account-label {
  display: grid;
  min-width: 0;
}

.account-label small {
  color: #687485;
  font-size: 0.55rem;
  font-weight: 750;
}

.account-label strong {
  max-width: 8rem;
  margin-top: 0.12rem;
  overflow: hidden;
  color: #e8ecf2;
  font-size: 0.73rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

summary svg {
  width: 0.85rem;
  margin-left: auto;
  fill: none;
  stroke: #7e8999;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 1.7;
  transition: transform 160ms ease;
}

.auth-card[open] summary svg {
  transform: rotate(180deg);
}

.account-menu {
  position: absolute;
  top: calc(100% + 0.65rem);
  right: 0;
  display: grid;
  width: min(19rem, calc(100vw - 2rem));
  gap: 0.65rem;
  padding: 1rem;
  border: 1px solid rgb(255 255 255 / 12%);
  border-radius: 0.3rem;
  background: #1b1b1f;
  box-shadow: 0 1rem 2.5rem rgb(0 0 0 / 35%);
}

.profile-detail {
  display: grid;
  min-width: 0;
  padding-bottom: 0.7rem;
  border-bottom: 1px solid rgb(255 255 255 / 8%);
}

.profile-detail span,
.account-menu p,
.account-menu small {
  margin: 0;
  color: #7f8999;
  font-size: 0.68rem;
  line-height: 1.55;
}

.profile-detail strong {
  margin-top: 0.2rem;
  overflow-wrap: anywhere;
  color: #d8dde5;
  font-size: 0.72rem;
}

.google-button,
.secondary-button,
.admin-link {
  display: flex;
  min-height: 2.45rem;
  align-items: center;
  justify-content: center;
  padding: 0.55rem 0.75rem;
  border-radius: 0.25rem;
  font-size: 0.72rem;
  font-weight: 800;
  text-decoration: none;
  cursor: pointer;
}

.google-button {
  border: 1px solid #d91920;
  color: #fff;
  background: #d91920;
}

.secondary-button,
.admin-link {
  border: 1px solid rgb(255 255 255 / 13%);
  color: #d9dee6;
  background: transparent;
}

.google-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.auth-error {
  color: #ff9d9d !important;
}

@media (max-width: 620px) {
  summary {
    min-width: 0;
  }

  .account-label {
    display: none;
  }

  summary svg {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  summary svg {
    transition: none;
  }
}
</style>
