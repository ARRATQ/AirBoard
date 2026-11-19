<template>
  <div
    ref="viewport"
    class="ot-viewport"
    :class="{ 'ot-viewport--dragging': dragging }"
    @pointerdown="onPointerDown"
    @pointermove="onPointerMove"
    @pointerup="onPointerUp"
    @pointercancel="onPointerUp"
    @wheel.prevent="onWheel"
  >
    <div ref="canvas" class="ot-canvas" :style="canvasStyle">
      <slot />
    </div>

    <div class="ot-controls" @pointerdown.stop>
      <button type="button" class="ot-control" :title="$t('directory.zoomIn')" :aria-label="$t('directory.zoomIn')" @click="zoomBy(1.2)">
        <Icon icon="mdi:plus" class="h-4 w-4" />
      </button>
      <span class="ot-zoom-level">{{ Math.round(scale * 100) }}%</span>
      <button type="button" class="ot-control" :title="$t('directory.zoomOut')" :aria-label="$t('directory.zoomOut')" @click="zoomBy(1 / 1.2)">
        <Icon icon="mdi:minus" class="h-4 w-4" />
      </button>
      <span class="ot-controls__sep" />
      <button type="button" class="ot-control" :title="$t('directory.fitToScreen')" :aria-label="$t('directory.fitToScreen')" @click="fit">
        <Icon icon="mdi:fit-to-screen-outline" class="h-4 w-4" />
      </button>
      <button type="button" class="ot-control" :title="$t('directory.actualSize')" :aria-label="$t('directory.actualSize')" @click="resetTop">
        <Icon icon="mdi:magnify-scan" class="h-4 w-4" />
      </button>
    </div>
    <p class="ot-hint">{{ $t('directory.panHint') }}</p>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { Icon } from '@iconify/vue'

const MIN_SCALE = 0.25
const MAX_SCALE = 2
const PADDING = 32

const viewport = ref(null)
const canvas = ref(null)
const scale = ref(1)
const x = ref(0)
const y = ref(0)
const dragging = ref(false)
const animate = ref(false)

const canvasStyle = computed(() => ({
  transform: `translate(${x.value}px, ${y.value}px) scale(${scale.value})`,
  transition: animate.value ? 'transform 280ms cubic-bezier(0.16, 1, 0.3, 1)' : 'none',
}))

const clamp = (v, min, max) => Math.min(max, Math.max(min, v))

const withAnimation = (fn) => {
  animate.value = true
  fn()
  setTimeout(() => { animate.value = false }, 300)
}

// Taille « naturelle » du contenu (indépendante de l'échelle courante)
const contentSize = () => ({ w: canvas.value?.scrollWidth || 0, h: canvas.value?.scrollHeight || 0 })

const zoomAt = (factor, cx, cy) => {
  const next = clamp(scale.value * factor, MIN_SCALE, MAX_SCALE)
  const ratio = next / scale.value
  x.value = cx - (cx - x.value) * ratio
  y.value = cy - (cy - y.value) * ratio
  scale.value = next
}

const zoomBy = (factor) => {
  const rect = viewport.value.getBoundingClientRect()
  withAnimation(() => zoomAt(factor, rect.width / 2, rect.height / 2))
}

// Ajuste tout l'arbre dans la zone visible (sans dépasser 100 %)
const fit = () => {
  if (!viewport.value) return
  const rect = viewport.value.getBoundingClientRect()
  const { w, h } = contentSize()
  if (!w || !h) return
  const s = clamp(Math.min((rect.width - PADDING * 2) / w, (rect.height - PADDING * 2) / h, 1), MIN_SCALE, MAX_SCALE)
  withAnimation(() => {
    scale.value = s
    x.value = (rect.width - w * s) / 2
    y.value = Math.max(PADDING, (rect.height - h * s) / 2)
  })
}

// Taille réelle, centré horizontalement, calé en haut
const resetTop = () => {
  const rect = viewport.value.getBoundingClientRect()
  const { w } = contentSize()
  withAnimation(() => {
    scale.value = 1
    x.value = (rect.width - w) / 2
    y.value = PADDING
  })
}

// Centre un élément du canvas dans la zone visible
const centerOn = (el) => {
  if (!el || !viewport.value) return
  const vp = viewport.value.getBoundingClientRect()
  const r = el.getBoundingClientRect()
  withAnimation(() => {
    x.value += vp.left + vp.width / 2 - (r.left + r.width / 2)
    y.value += vp.top + PADDING * 2 - r.top
  })
}

// ---- Déplacement (souris / doigt) et pincement ----
const pointers = new Map()
let last = null
let pinchDistance = 0

const isInteractive = (target) => !!target.closest('button, a, input, select, textarea')

const onPointerDown = (e) => {
  if (e.pointerType === 'mouse' && e.button !== 0) return
  if (isInteractive(e.target) && e.pointerType === 'mouse') return
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY })
  viewport.value.setPointerCapture?.(e.pointerId)
  last = { x: e.clientX, y: e.clientY }
  dragging.value = true
  if (pointers.size === 2) {
    const [a, b] = [...pointers.values()]
    pinchDistance = Math.hypot(a.x - b.x, a.y - b.y)
  }
}

const onPointerMove = (e) => {
  if (!pointers.has(e.pointerId)) return
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY })
  if (pointers.size === 2) {
    const [a, b] = [...pointers.values()]
    const dist = Math.hypot(a.x - b.x, a.y - b.y)
    if (pinchDistance) {
      const rect = viewport.value.getBoundingClientRect()
      zoomAt(dist / pinchDistance, (a.x + b.x) / 2 - rect.left, (a.y + b.y) / 2 - rect.top)
    }
    pinchDistance = dist
    return
  }
  x.value += e.clientX - last.x
  y.value += e.clientY - last.y
  last = { x: e.clientX, y: e.clientY }
}

const onPointerUp = (e) => {
  pointers.delete(e.pointerId)
  if (pointers.size < 2) pinchDistance = 0
  if (pointers.size === 0) dragging.value = false
  else last = [...pointers.values()][0]
}

const onWheel = (e) => {
  const rect = viewport.value.getBoundingClientRect()
  // Pavé tactile sans Ctrl : défilement ; molette ou Ctrl+pavé : zoom
  if (!e.ctrlKey && Math.abs(e.deltaX) > 0 && e.deltaMode === 0 && Math.abs(e.deltaY) < 50) {
    x.value -= e.deltaX
    y.value -= e.deltaY
    return
  }
  zoomAt(e.deltaY < 0 ? 1.1 : 1 / 1.1, e.clientX - rect.left, e.clientY - rect.top)
}

const onKey = (e) => {
  if (e.target.closest?.('input, textarea, select')) return
  if (e.key === '+' || e.key === '=') zoomBy(1.2)
  else if (e.key === '-') zoomBy(1 / 1.2)
  else if (e.key === '0') fit()
}

onMounted(async () => {
  await nextTick()
  fit()
  window.addEventListener('keydown', onKey)
})
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))

defineExpose({ fit, centerOn, resetTop })
</script>

<style scoped>
.ot-viewport {
  position: relative;
  height: clamp(32rem, calc(100vh - 16rem), 60rem);
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 1.25rem;
  cursor: grab;
  touch-action: none;
  user-select: none;
  background-color: var(--bg-surface);
  background-image: radial-gradient(circle, color-mix(in srgb, var(--text-muted) 22%, transparent) 1px, transparent 1.2px);
  background-size: 22px 22px;
}
.ot-viewport--dragging { cursor: grabbing; }
.ot-canvas {
  position: absolute;
  top: 0;
  left: 0;
  transform-origin: 0 0;
  width: max-content;
  will-change: transform;
}
.ot-controls {
  position: absolute;
  right: 1rem;
  bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.15rem;
  padding: 0.3rem;
  border: 1px solid var(--border);
  border-radius: 0.85rem;
  background: color-mix(in srgb, var(--bg-card) 88%, transparent);
  backdrop-filter: blur(8px);
  box-shadow: 0 10px 25px -12px rgba(var(--shadow-rgb, 15, 23, 42), 0.35);
  cursor: default;
}
.ot-control {
  display: inline-flex;
  padding: 0.45rem;
  border-radius: 0.55rem;
  color: var(--text-primary);
  transition: background-color 150ms ease;
}
.ot-control:hover { background: var(--hover); }
.ot-control:focus-visible { outline: 2px solid var(--accent); }
.ot-zoom-level { min-width: 3rem; text-align: center; font-size: 0.75rem; font-variant-numeric: tabular-nums; color: var(--text-muted); }
.ot-controls__sep { width: 1px; height: 1.25rem; margin: 0 0.25rem; background: var(--border); }
.ot-hint {
  position: absolute;
  left: 1rem;
  bottom: 1rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  pointer-events: none;
}
@media (prefers-reduced-motion: reduce) {
  .ot-canvas { transition: none !important; }
}
</style>
