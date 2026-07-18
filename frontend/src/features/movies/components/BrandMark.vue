<template>
  <span class="brand-mark" aria-label="Lumina Cinema">
    <span class="brand-icon" aria-hidden="true"><i></i><i></i><i></i></span>
    <span class="brand-type"><strong>LUMINA</strong><small>CINEMA</small></span>
  </span>
</template>

<style scoped>
.brand-mark {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  color: #fff;
}

.brand-icon {
  display: flex;
  width: 2rem;
  height: 2rem;
  align-items: flex-end;
  justify-content: center;
  gap: 0.12rem;
  padding: 0.38rem;
  border-radius: 0.2rem;
  background: #d91920;
}

.brand-icon i {
  display: block;
  width: 0.2rem;
  border-radius: 999px;
  background: #fff;
  transform-origin: bottom;
}

.brand-icon i:nth-child(1) {
  height: 0.55rem;
  transform: rotate(-18deg);
}

.brand-icon i:nth-child(2) {
  height: 1.05rem;
}

.brand-icon i:nth-child(3) {
  height: 0.55rem;
  transform: rotate(18deg);
}

.brand-type {
  display: grid;
  line-height: 1;
}

.brand-type strong {
  font-size: 0.88rem;
  letter-spacing: 0.16em;
}

.brand-type small {
  margin-top: 0.22rem;
  color: #b9b9bf;
  font-size: 0.5rem;
  font-weight: 800;
  letter-spacing: 0.34em;
}
</style>
