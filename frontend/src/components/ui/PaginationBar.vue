<template>
  <nav v-if="total > 0" class="dt__pager" :aria-label="t('ui.table.range', { from, to, total })">
    <span class="dt__range">{{ t('ui.table.range', { from, to, total }) }}</span>
    <template v-if="pageCount > 1">
      <button type="button" class="dt__page-btn" :disabled="page <= 1" :aria-label="t('ui.table.prev')" @click="go(page - 1)">
        <Icon icon="mdi:chevron-left" class="dt__chev" aria-hidden="true" />
      </button>
      <span class="dt__page">{{ page }} / {{ pageCount }}</span>
      <button type="button" class="dt__page-btn" :disabled="page >= pageCount" :aria-label="t('ui.table.next')" @click="go(page + 1)">
        <Icon icon="mdi:chevron-right" class="dt__chev" aria-hidden="true" />
      </button>
    </template>
  </nav>
</template>

<script setup>
// Barre de pagination partagée : pages calculées en mémoire (DataTable) ou côté serveur.
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  page: { type: Number, required: true },
  pageSize: { type: Number, required: true },
  total: { type: Number, required: true },
})
const emit = defineEmits(['update:page'])

const { t } = useI18n()

const pageCount = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const from = computed(() => (props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1))
const to = computed(() => Math.min(props.page * props.pageSize, props.total))

const go = (target) => {
  const next = Math.min(Math.max(1, target), pageCount.value)
  if (next !== props.page) emit('update:page', next)
}
</script>

<style scoped>
.dt__pager {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  padding: 8px 12px;
  border-block-start: 1px solid var(--line);
  font-family: var(--f-num);
  font-size: 12px;
  color: var(--muted);
}

.dt__range { margin-inline-end: auto; }

.dt__page-btn {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
  color: var(--text);
  cursor: pointer;
}

.dt__page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.dt__page-btn:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 2px;
}

.dt__chev {
  width: 16px;
  height: 16px;
}

:global([dir='rtl']) .dt__chev { transform: scaleX(-1); }
</style>
