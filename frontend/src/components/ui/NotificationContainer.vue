<template>
  <!-- En haut, côté fin : loin du bouton flottant (bas) et du bouton de menu mobile (haut, début). -->
  <section class="toasts" :aria-label="t('ui.toast.region')">
    <!-- Deux régions live : les erreurs interrompent (assertive), le reste attend (polite). -->
    <div class="sr-only" aria-live="assertive" aria-atomic="true">{{ announce.assertive }}</div>
    <div class="sr-only" aria-live="polite" aria-atomic="true">{{ announce.polite }}</div>

    <TransitionGroup name="toast" tag="ul" class="toasts__list">
      <li
        v-for="notification in appStore.notifications"
        :key="notification.id"
        class="toast"
        :class="`toast--${toneOf(notification.type)}`"
      >
        <Icon :icon="ICONS[toneOf(notification.type)]" class="toast__icon" aria-hidden="true" />
        <div class="toast__body">
          <p v-if="notification.title" class="toast__title">{{ notification.title }}</p>
          <p v-if="notification.message" class="toast__message">{{ notification.message }}</p>
        </div>
        <button
          type="button"
          class="toast__close"
          :aria-label="t('ui.toast.close')"
          @click="appStore.removeNotification(notification.id)"
        >
          <Icon icon="mdi:close" aria-hidden="true" />
        </button>
      </li>
    </TransitionGroup>
  </section>
</template>

<script setup>
import { reactive, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const { t } = useI18n()

const ICONS = {
  success: 'mdi:check-circle-outline',
  error: 'mdi:alert-circle-outline',
  warning: 'mdi:alert-outline',
  info: 'mdi:information-outline',
}

const toneOf = (type) => (ICONS[type] ? type : 'info')

// Le texte de chaque nouveau toast est recopié dans la région live adaptée à sa gravité.
const announce = reactive({ assertive: '', polite: '' })
watch(
  () => appStore.notifications.map(n => n.id),
  (ids, previous = []) => {
    const added = appStore.notifications.filter(n => !previous.includes(n.id))
    for (const n of added) {
      const text = [n.title, n.message].filter(Boolean).join(' : ')
      if (n.type === 'error') announce.assertive = text
      else announce.polite = text
    }
  },
)
</script>

<style scoped>
.toasts {
  position: fixed;
  inset-block-start: 72px;
  inset-inline-end: 16px;
  z-index: 10002;
  width: min(380px, calc(100vw - 32px));
  pointer-events: none;
}

.toasts__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.toast {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding-block: 10px;
  padding-inline: 12px 10px;
  background: var(--surface);
  color: var(--text);
  border: 1px solid var(--line);
  border-inline-start: 3px solid var(--tone);
  border-radius: var(--r-m);
  box-shadow: var(--shadow-float);
  pointer-events: auto;
}

.toast--success { --tone: var(--ok); }
.toast--error { --tone: var(--bad); }
.toast--warning { --tone: var(--warn); }
.toast--info { --tone: var(--info); }

.toast__icon {
  flex: none;
  width: 18px;
  height: 18px;
  margin-block-start: 1px;
  color: var(--tone);
}

.toast__body {
  flex: 1;
  min-width: 0;
}

.toast__title {
  margin: 0;
  font-weight: 600;
}

.toast__message {
  margin: 2px 0 0;
  color: var(--muted);
  overflow-wrap: anywhere;
}

.toast__close {
  display: grid;
  place-items: center;
  flex: none;
  width: 24px;
  height: 24px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  color: var(--muted);
  cursor: pointer;
}

.toast__close:hover {
  background: var(--sunken);
  color: var(--text);
}

.toast__close:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 1px;
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.toast-move { transition: transform 0.18s ease; }
</style>
