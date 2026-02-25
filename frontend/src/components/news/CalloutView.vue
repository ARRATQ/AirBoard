<template>
  <node-view-wrapper as="div" class="callout-nv-outer">
    <div class="callout-block" :data-callout-type="node.attrs.type">
      <!-- Header -->
      <div class="callout-header">
        <Icon :icon="config.icon" class="callout-header-icon" :style="{ color: config.color }" />
        <span class="callout-header-label" :style="{ color: config.color }">{{ config.label }}</span>

        <!-- Type switcher (visible when selected) -->
        <div v-if="selected" class="callout-switcher">
          <button
            v-for="t in types"
            :key="t.value"
            type="button"
            class="callout-switch-btn"
            :class="{ 'callout-switch-active': node.attrs.type === t.value }"
            :title="t.label"
            :style="node.attrs.type === t.value ? { background: t.color, borderColor: t.color } : { borderColor: t.color }"
            @mousedown.prevent="updateAttributes({ type: t.value })"
          >
            <Icon :icon="t.icon" class="h-3.5 w-3.5" :style="{ color: node.attrs.type === t.value ? '#fff' : t.color }" />
          </button>
        </div>
      </div>

      <!-- Editable content -->
      <node-view-content class="callout-content" />
    </div>
  </node-view-wrapper>
</template>

<script setup>
import { computed } from 'vue'
import { NodeViewWrapper, NodeViewContent } from '@tiptap/vue-3'
import { Icon } from '@iconify/vue'
import { CALLOUT_TYPES } from './CalloutExtension.js'

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

const types = CALLOUT_TYPES

const config = computed(() => types.find(t => t.value === props.node.attrs.type) || types[0])
</script>

<style scoped>
.callout-nv-outer {
  display: block;
  margin: 1rem 0;
}

.callout-block {
  border-radius: 8px;
  border-left: 4px solid;
  padding: 12px 16px;
  margin: 0;
}

.callout-block[data-callout-type="info"] {
  background: #EFF6FF;
  border-left-color: #3B82F6;
}
.callout-block[data-callout-type="warning"] {
  background: #FFFBEB;
  border-left-color: #F59E0B;
}
.callout-block[data-callout-type="success"] {
  background: #F0FDF4;
  border-left-color: #22C55E;
}
.callout-block[data-callout-type="danger"] {
  background: #FEF2F2;
  border-left-color: #EF4444;
}

/* Dark mode */
:global(.dark) .callout-block[data-callout-type="info"]    { background: #1E3A5F; }
:global(.dark) .callout-block[data-callout-type="warning"] { background: #3D2C00; }
:global(.dark) .callout-block[data-callout-type="success"] { background: #052E16; }
:global(.dark) .callout-block[data-callout-type="danger"]  { background: #3B0A0A; }

.callout-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}

.callout-header-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.callout-header-label {
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.025em;
  text-transform: uppercase;
  flex: 1;
}

/* Type switcher */
.callout-switcher {
  display: flex;
  gap: 4px;
  margin-left: auto;
}

.callout-switch-btn {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 2px solid;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s, border-color 0.15s;
  padding: 0;
}

.callout-switch-btn:hover {
  opacity: 0.8;
}

.callout-content {
  font-size: 0.95em;
  line-height: 1.6;
}

.callout-content :deep(p:last-child) {
  margin-bottom: 0;
}
</style>
