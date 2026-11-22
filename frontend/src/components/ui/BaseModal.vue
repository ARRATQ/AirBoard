<template>
  <!-- Dialog de Headless UI : piège du focus, Échap, clic extérieur, role/aria-modal. -->
  <Dialog :open="open" :role="role" :initial-focus="initialFocus" class="base-modal" @close="onClose">
    <div class="base-modal__overlay" aria-hidden="true" />
    <div class="base-modal__wrap">
      <DialogPanel class="base-modal__panel" :class="`base-modal__panel--${size}`">
        <header v-if="title || $slots.title" class="base-modal__header">
          <div class="base-modal__heading">
            <DialogTitle class="base-modal__title">
              <slot name="title">{{ title }}</slot>
            </DialogTitle>
            <DialogDescription v-if="description" class="base-modal__desc">
              {{ description }}
            </DialogDescription>
          </div>
          <button
            v-if="closable"
            type="button"
            class="base-modal__close"
            :aria-label="t('ui.close')"
            @click="onClose"
          >
            <Icon icon="lucide:x" aria-hidden="true" />
          </button>
        </header>
        <div class="base-modal__body">
          <slot />
        </div>
        <footer v-if="$slots.footer" class="base-modal__footer">
          <slot name="footer" />
        </footer>
      </DialogPanel>
    </div>
  </Dialog>
</template>

<script setup>
import { Dialog, DialogPanel, DialogTitle, DialogDescription } from '@headlessui/vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '' },
  description: { type: String, default: '' },
  size: { type: String, default: 'md', validator: (v) => ['sm', 'md', 'lg', 'xl'].includes(v) },
  role: { type: String, default: 'dialog', validator: (v) => ['dialog', 'alertdialog'].includes(v) },
  closable: { type: Boolean, default: true },
  // Élément qui reçoit le focus à l'ouverture (ref de template, déballée par le template) ;
  // sinon le premier élément focusable. Headless UI refocalise quand la ref se remplit.
  initialFocus: { default: null },
})

const emit = defineEmits(['close'])
const { t } = useI18n()

const onClose = () => {
  if (props.closable) emit('close')
}
</script>

<style scoped>
.base-modal {
  position: fixed;
  inset: 0;
  z-index: 60;
}

.base-modal__overlay {
  position: fixed;
  inset: 0;
  background: rgb(5 6 7 / 0.55);
}

.base-modal__wrap {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  overflow-y: auto;
}

.base-modal__panel {
  position: relative;
  display: flex;
  flex-direction: column;
  width: 100%;
  max-height: calc(100vh - 32px);
  background: var(--surface);
  color: var(--text);
  border: 1px solid var(--line);
  border-radius: var(--r-l);
  box-shadow: var(--shadow-float);
  animation: base-modal-in 0.14s ease-out;
}

.base-modal__panel--sm { max-width: 420px; }
.base-modal__panel--md { max-width: 560px; }
.base-modal__panel--lg { max-width: 720px; }
.base-modal__panel--xl { max-width: 960px; }

.base-modal__header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px 18px 12px;
  border-bottom: 1px solid var(--line);
}

.base-modal__heading {
  flex: 1;
  min-width: 0;
}

.base-modal__title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.3;
}

.base-modal__desc {
  margin: 4px 0 0;
  color: var(--muted);
  font-size: 13px;
}

.base-modal__close {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  margin-inline-end: -6px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  color: var(--muted);
  font-size: 18px;
  cursor: pointer;
}

.base-modal__close:hover {
  background: var(--sunken);
  color: var(--text);
}

.base-modal__close:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 2px;
}

.base-modal__body {
  flex: 1;
  min-height: 0;
  padding: 16px 18px;
  overflow-y: auto;
}

.base-modal__footer {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 18px;
  border-top: 1px solid var(--line);
  background: var(--sunken);
  border-end-start-radius: var(--r-l);
  border-end-end-radius: var(--r-l);
}

@keyframes base-modal-in {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: none; }
}
</style>
