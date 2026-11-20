<template>
  <div class="empty" :class="{ 'empty--compact': compact }">
    <Icon v-if="icon" :icon="icon" class="empty__icon" aria-hidden="true" />
    <p class="empty__title">{{ title }}</p>
    <p v-if="description" class="empty__desc">{{ description }}</p>
    <div v-if="$slots.default" class="empty__action">
      <slot />
    </div>
  </div>
</template>

<script setup>
import { Icon } from '@iconify/vue'

defineProps({
  title: { type: String, required: true },
  description: { type: String, default: '' },
  icon: { type: String, default: '' },
  compact: { type: Boolean, default: false },
})
</script>

<style scoped>
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 40px 16px;
  text-align: center;
  color: var(--muted);
}

.empty--compact { padding-block: 20px; }

.empty__icon {
  width: 28px;
  height: 28px;
  margin-block-end: 4px;
}

.empty__title {
  margin: 0;
  font-weight: 600;
  color: var(--text);
}

.empty__desc {
  margin: 0;
  max-width: 48ch;
}

.empty__action { margin-block-start: 8px; }
</style>
