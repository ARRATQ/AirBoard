<template>
  <main class="err" :aria-labelledby="titleId">
    <Icon :icon="icon" class="err__icon" :class="{ 'err__icon--bad': tone === 'bad' }" aria-hidden="true" />
    <p class="err__code">{{ code }}</p>
    <h1 :id="titleId" class="err__title">{{ title }}</h1>
    <p class="err__msg">{{ message }}</p>
    <div class="err__actions">
      <BaseButton tag="router-link" to="/home" variant="primary" icon="lucide:house">{{ t('errors.backHome') }}</BaseButton>
      <BaseButton icon="lucide:arrow-left" @click="router.back()">{{ t('errors.previousPage') }}</BaseButton>
    </div>
  </main>
</template>

<script setup>
// Gabarit commun des pages 404 et 403.
import { useId } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import BaseButton from '@/components/ui/BaseButton.vue'

defineProps({
  code: { type: String, required: true },
  icon: { type: String, required: true },
  title: { type: String, required: true },
  message: { type: String, required: true },
  tone: { type: String, default: 'neutral' },
})

const { t } = useI18n()
const router = useRouter()
const titleId = useId()
</script>

<style scoped>
.err {
  min-height: 70vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 32px 16px;
  text-align: center;
  color: var(--text);
}
.err__icon {
  width: 56px;
  height: 56px;
  color: var(--muted);
}
.err__icon--bad {
  color: var(--bad);
}
.err__code {
  font-family: var(--f-num);
  font-size: 11px;
  letter-spacing: 0.1em;
  color: var(--muted);
}
.err__title {
  font-size: 22px;
  font-weight: 600;
}
.err__msg {
  max-width: 44ch;
  color: var(--muted);
}
.err__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
  margin-top: 16px;
}
</style>
