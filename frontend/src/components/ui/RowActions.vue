<template>
  <div class="row-actions">
    <slot />
    <button
      v-if="editable"
      type="button"
      class="row-actions__btn"
      :aria-label="`${t('common.edit')} : ${label}`"
      :title="t('common.edit')"
      @click="$emit('edit')"
    >
      <Icon icon="mdi:pencil-outline" aria-hidden="true" />
    </button>
    <button
      v-if="deletable"
      type="button"
      class="row-actions__btn row-actions__btn--danger"
      :aria-label="`${t('common.delete')} : ${label}`"
      :title="t('common.delete')"
      @click="$emit('delete')"
    >
      <Icon icon="mdi:trash-can-outline" aria-hidden="true" />
    </button>
  </div>
</template>

<script setup>
// Actions de fin de ligne d'un tableau d'admin : boutons icônes Modifier / Supprimer, nommés
// pour les lecteurs d'écran (« Modifier : <libellé> »). Le slot ajoute d'autres actions avant.
// Pour une action supplémentaire, réutiliser la classe row-actions__btn.
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'

defineProps({
  label: { type: String, required: true },
  editable: { type: Boolean, default: true },
  deletable: { type: Boolean, default: true },
})

defineEmits(['edit', 'delete'])

const { t } = useI18n()
</script>

<style>
/* Non scopé : les boutons ajoutés par le slot depuis une page reprennent ce style. */
.row-actions {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 2px;
}

.row-actions__btn {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  color: var(--muted);
  font-size: 17px;
  cursor: pointer;
}

.row-actions__btn:hover { background: var(--sunken); color: var(--text); }
.row-actions__btn--danger:hover { color: var(--bad); }
.row-actions__btn:disabled { opacity: 0.4; cursor: not-allowed; }

.row-actions__btn:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 1px;
}
</style>
