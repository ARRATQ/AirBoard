<template>
  <node-view-wrapper as="div" class="image-view-outer" :style="outerStyle">
    <div
      ref="containerRef"
      class="image-view-inner"
      :class="{ 'image-view-selected': selected }"
      :style="{ width: currentWidth }"
    >
      <!-- Image -->
      <img
        :src="node.attrs.src"
        :alt="node.attrs.alt || ''"
        draggable="false"
        class="image-view-img"
      />

      <!-- Floating toolbar (visible when selected) -->
      <div v-if="selected" class="image-size-bar">
        <!-- Size presets -->
        <button
          v-for="p in presets"
          :key="p.value"
          type="button"
          class="image-size-preset"
          :class="{ 'image-size-active': currentWidth === p.value }"
          @mousedown.prevent="setWidth(p.value)"
        >
          {{ p.label }}
        </button>
        <span class="image-size-indicator">{{ currentWidth }}</span>

        <!-- Divider -->
        <span class="image-bar-divider" />

        <!-- Alignment buttons -->
        <button
          v-for="a in alignOptions"
          :key="a.value"
          type="button"
          class="image-align-btn"
          :class="{ 'image-size-active': currentAlign === a.value }"
          :title="a.label"
          @mousedown.prevent="setAlign(a.value)"
        >
          <Icon :icon="a.icon" class="h-3.5 w-3.5" />
        </button>
      </div>

      <!-- Right resize handle -->
      <div
        v-if="selected"
        class="image-resize-handle"
        @mousedown.prevent.stop="startResize"
      />
    </div>
  </node-view-wrapper>
</template>

<script setup>
import { ref, computed, onBeforeUnmount } from 'vue'
import { NodeViewWrapper } from '@tiptap/vue-3'
import { Icon } from '@iconify/vue'

const props = defineProps({
  editor: Object,
  node: Object,
  decorations: Array,
  selected: Boolean,
  extension: Object,
  getPos: Function,
  updateAttributes: Function,
  deleteNode: Function,
})

const containerRef = ref(null)

const presets = [
  { label: '25%', value: '25%' },
  { label: '50%', value: '50%' },
  { label: '75%', value: '75%' },
  { label: '100%', value: '100%' },
]

const alignOptions = [
  { value: 'left', label: 'Aligner à gauche', icon: 'mdi:format-align-left' },
  { value: 'center', label: 'Centrer', icon: 'mdi:format-align-center' },
  { value: 'right', label: 'Aligner à droite', icon: 'mdi:format-align-right' },
]

const currentWidth = computed(() => props.node.attrs.width || '100%')
const currentAlign = computed(() => props.node.attrs.align || 'center')

// Outer wrapper style drives the alignment
const outerStyle = computed(() => {
  const align = currentAlign.value
  if (align === 'left') {
    return { float: 'left', marginTop: '0.5rem', marginRight: '1.5rem', marginBottom: '0.5rem', width: 'auto' }
  }
  if (align === 'right') {
    return { float: 'right', marginTop: '0.5rem', marginLeft: '1.5rem', marginBottom: '0.5rem', width: 'auto' }
  }
  // center (default)
  return { float: 'none', textAlign: 'center', width: '100%', margin: '1rem 0' }
})

const setWidth = (value) => {
  props.updateAttributes({ width: value })
}

const setAlign = (value) => {
  props.updateAttributes({ align: value })
}

// Drag resize
let isResizing = false
let resizeStartX = 0
let resizeStartWidth = 0

const startResize = (e) => {
  isResizing = true
  resizeStartX = e.clientX
  resizeStartWidth = containerRef.value
    ? containerRef.value.getBoundingClientRect().width
    : 0
  window.addEventListener('mousemove', onResizeMove)
  window.addEventListener('mouseup', stopResize)
}

const onResizeMove = (e) => {
  if (!isResizing || !containerRef.value) return
  const dx = e.clientX - resizeStartX
  const parentWidth =
    containerRef.value.parentElement?.getBoundingClientRect().width || resizeStartWidth
  const newPct = Math.round(((resizeStartWidth + dx) / parentWidth) * 100)
  props.updateAttributes({ width: `${Math.max(10, Math.min(100, newPct))}%` })
}

const stopResize = () => {
  isResizing = false
  window.removeEventListener('mousemove', onResizeMove)
  window.removeEventListener('mouseup', stopResize)
}

onBeforeUnmount(() => {
  window.removeEventListener('mousemove', onResizeMove)
  window.removeEventListener('mouseup', stopResize)
})
</script>

<style scoped>
.image-view-outer {
  display: block;
}

.image-view-inner {
  position: relative;
  display: inline-block;
  max-width: 100%;
  vertical-align: top;
  cursor: default;
  user-select: none;
  transition: width 0.1s ease;
}

.image-view-selected {
  outline: 2px solid #3B82F6;
  outline-offset: 2px;
  border-radius: 8px;
}

.image-view-img {
  display: block;
  width: 100%;
  height: auto;
  border-radius: 6px;
  pointer-events: none;
}

/* Floating toolbar */
.image-size-bar {
  position: absolute;
  bottom: 10px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 4px 8px;
  background: rgba(15, 23, 42, 0.88);
  backdrop-filter: blur(6px);
  border-radius: 20px;
  z-index: 10;
  white-space: nowrap;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
}

.image-size-preset {
  padding: 3px 9px;
  font-size: 11px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.65);
  background: transparent;
  border: none;
  border-radius: 14px;
  cursor: pointer;
  transition: color 0.15s, background 0.15s;
}

.image-size-preset:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.12);
}

.image-size-active {
  color: #fff !important;
  background: #3B82F6 !important;
}

.image-size-indicator {
  font-size: 10px;
  color: rgba(255, 255, 255, 0.4);
  padding-left: 6px;
  margin-left: 4px;
  border-left: 1px solid rgba(255, 255, 255, 0.15);
}

/* Divider between size and align controls */
.image-bar-divider {
  display: inline-block;
  width: 1px;
  height: 16px;
  background: rgba(255, 255, 255, 0.2);
  margin: 0 6px;
  flex-shrink: 0;
}

/* Alignment buttons */
.image-align-btn {
  padding: 4px 6px;
  color: rgba(255, 255, 255, 0.65);
  background: transparent;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.15s, background 0.15s;
}

.image-align-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.12);
}

/* Right drag handle */
.image-resize-handle {
  position: absolute;
  right: -6px;
  top: 50%;
  transform: translateY(-50%);
  width: 12px;
  height: 44px;
  background: #3B82F6;
  border: 2px solid #fff;
  border-radius: 6px;
  cursor: ew-resize;
  z-index: 10;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.25);
  transition: background 0.15s, transform 0.15s;
}

.image-resize-handle:hover {
  background: #2563EB;
  transform: translateY(-50%) scaleX(1.2);
}
</style>
