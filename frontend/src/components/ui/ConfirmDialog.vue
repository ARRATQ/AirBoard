<template>
  <BaseModal
    :open="confirmState.open"
    :title="confirmState.title || t('ui.confirm.title')"
    role="alertdialog"
    size="sm"
    :initial-focus="cancelButton"
    @close="settleConfirm(false)"
  >
    <p class="confirm-dialog__message">{{ confirmState.message }}</p>
    <template #footer>
      <button ref="cancelButton" type="button" class="confirm-dialog__btn confirm-dialog__btn--cancel" @click="settleConfirm(false)">
        {{ confirmState.cancelLabel || t('ui.confirm.cancel') }}
      </button>
      <button
        type="button"
        class="confirm-dialog__btn"
        :class="confirmState.tone === 'danger' ? 'confirm-dialog__btn--danger' : 'confirm-dialog__btn--primary'"
        @click="settleConfirm(true)"
      >
        {{ confirmState.confirmLabel || defaultConfirmLabel }}
      </button>
    </template>
  </BaseModal>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseModal from './BaseModal.vue'
import { confirmState, settleConfirm } from '@/composables/useConfirm'

const { t } = useI18n()
// Le focus va sur « Annuler » : une validation par Entrée ne déclenche pas une suppression.
const cancelButton = ref(null)

const defaultConfirmLabel = computed(() =>
  confirmState.tone === 'danger' ? t('ui.confirm.delete') : t('ui.confirm.ok'),
)
</script>

<style scoped>
.confirm-dialog__message {
  margin: 0;
  line-height: 1.5;
  white-space: pre-line;
}

/* Boutons natifs : la ref d'initialFocus pointe directement un élément focusable. */
.confirm-dialog__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 32px;
  padding-inline: 12px;
  border: 1px solid transparent;
  border-radius: var(--r-m);
  font: inherit;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.confirm-dialog__btn:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 2px;
}

.confirm-dialog__btn--cancel {
  background: var(--surface);
  color: var(--text);
  border-color: var(--line);
}

.confirm-dialog__btn--cancel:hover { background: var(--sunken); }

.confirm-dialog__btn--primary {
  background: var(--fill);
  color: var(--on-fill);
}

.confirm-dialog__btn--danger {
  background: var(--bad);
  color: #fff;
}

:global(html.dark) .confirm-dialog__btn--danger { color: #141210; }

.confirm-dialog__btn--primary:hover,
.confirm-dialog__btn--danger:hover { filter: brightness(1.08); }
</style>
