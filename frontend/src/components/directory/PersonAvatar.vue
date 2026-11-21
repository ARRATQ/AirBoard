<template>
  <span
    class="person-avatar"
    :class="[`person-avatar--${size}`, { 'person-avatar--vacant': !person }]"
    :style="!person || showPhoto ? null : { backgroundColor: color }"
    :aria-label="person ? name : undefined"
  >
    <img
      v-if="person && showPhoto"
      :src="person.photo_url"
      :alt="name"
      loading="lazy"
      @error="photoFailed = true"
    />
    <Icon v-else-if="!person" icon="mdi:account-question-outline" class="person-avatar__vacant-icon" />
    <span v-else class="person-avatar__initials">{{ letters }}</span>
  </span>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { fullName, initials, avatarColor } from './useDirectory'

const props = defineProps({
  person: { type: Object, default: null },
  size: { type: String, default: 'md' }, // xs | sm | md | lg | xl
})

const photoFailed = ref(false)
watch(() => props.person?.photo_url, () => { photoFailed.value = false })

const showPhoto = computed(() => !!props.person?.photo_url && !photoFailed.value)
const name = computed(() => fullName(props.person))
const letters = computed(() => initials(props.person))
const color = computed(() => avatarColor(props.person))
</script>

<style scoped>
.person-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: 9999px;
  overflow: hidden;
  color: #fff;
  font-weight: 600;
  letter-spacing: 0.02em;
  box-shadow: 0 0 0 2px var(--surface);
}
.person-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.person-avatar--vacant {
  background: transparent;
  color: var(--muted);
  border: 1px dashed var(--muted);
  box-shadow: none;
}
.person-avatar--xs { width: 1.75rem; height: 1.75rem; font-size: 0.65rem; }
.person-avatar--sm { width: 2.25rem; height: 2.25rem; font-size: 0.75rem; }
.person-avatar--md { width: 3rem; height: 3rem; font-size: 0.95rem; }
.person-avatar--lg { width: 5rem; height: 5rem; font-size: 1.5rem; }
.person-avatar--xl { width: 7rem; height: 7rem; font-size: 2.1rem; }
.person-avatar__vacant-icon { width: 55%; height: 55%; }
</style>
