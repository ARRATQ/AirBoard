<template>
  <node-view-wrapper as="div" class="video-embed-nv">
    <div class="video-embed-wrapper" :class="{ 'video-selected': selected }">
      <!-- Service badge -->
      <div class="video-service-badge">
        <Icon
          :icon="node.attrs.service === 'vimeo' ? 'mdi:vimeo' : 'mdi:youtube'"
          class="h-4 w-4"
        />
        <span>{{ node.attrs.service === 'vimeo' ? 'Vimeo' : 'YouTube' }}</span>
      </div>

      <!-- Iframe preview -->
      <iframe
        :src="node.attrs.src"
        frameborder="0"
        allowfullscreen
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
        loading="lazy"
        class="video-embed-iframe"
      />

      <!-- Delete overlay (shown when selected) -->
      <div v-if="selected" class="video-delete-btn" @mousedown.prevent="deleteNode" title="Supprimer la vidéo">
        <Icon icon="mdi:close" class="h-4 w-4" />
      </div>
    </div>
  </node-view-wrapper>
</template>

<script setup>
import { NodeViewWrapper } from '@tiptap/vue-3'
import { Icon } from '@iconify/vue'

defineProps({
  editor: Object,
  node: Object,
  decorations: Array,
  selected: Boolean,
  extension: Object,
  getPos: Function,
  updateAttributes: Function,
  deleteNode: Function,
})
</script>

<style scoped>
.video-embed-nv {
  display: block;
  margin: 1.5rem 0;
}

.video-embed-wrapper {
  position: relative;
  border-radius: 10px;
  overflow: hidden;
  aspect-ratio: 16 / 9;
  background: #000;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.video-embed-wrapper.video-selected {
  outline: 2px solid #3B82F6;
  outline-offset: 2px;
}

.video-embed-iframe {
  width: 100%;
  height: 100%;
  display: block;
  border: none;
}

/* Service badge top-left */
.video-service-badge {
  position: absolute;
  top: 10px;
  left: 10px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  background: rgba(0, 0, 0, 0.65);
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  border-radius: 12px;
  z-index: 2;
  pointer-events: none;
}

/* Delete button top-right */
.video-delete-btn {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(239, 68, 68, 0.85);
  color: #fff;
  border-radius: 50%;
  cursor: pointer;
  z-index: 2;
  transition: background 0.15s;
}

.video-delete-btn:hover {
  background: #dc2626;
}
</style>
