<template>
  <div class="dt">
    <div v-if="searchable || $slots.toolbar" class="dt__toolbar">
      <label v-if="searchable" class="dt__search">
        <Icon icon="mdi:magnify" class="dt__search-icon" aria-hidden="true" />
        <input
          v-model="query"
          type="search"
          class="dt__search-input"
          :placeholder="searchPlaceholder || t('ui.table.search')"
          :aria-label="searchPlaceholder || t('ui.table.search')"
        />
      </label>
      <slot name="toolbar" />
    </div>

    <div class="dt__scroll">
      <table class="dt__table">
        <thead>
          <tr>
            <th
              v-for="col in columns"
              :key="col.key"
              scope="col"
              :class="`dt__cell--${col.align || 'start'}`"
              :style="col.width ? { width: col.width } : null"
              :aria-sort="col.sortable ? ariaSort(col.key) : undefined"
            >
              <button
                v-if="col.sortable"
                type="button"
                class="dt__sort"
                :title="sortKey === col.key && sortDir === 'asc' ? t('ui.table.sortDesc') : t('ui.table.sortAsc')"
                @click="toggleSort(col.key)"
              >
                {{ col.label }}
                <Icon :icon="sortIcon(col.key)" class="dt__sort-icon" aria-hidden="true" />
              </button>
              <template v-else>{{ col.label }}</template>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="columns.length" class="dt__state">
              <Icon icon="mdi:loading" class="dt__spin" aria-hidden="true" />
              <span class="sr-only">{{ t('common.loading') }}</span>
            </td>
          </tr>
          <tr v-else-if="pageRows.length === 0">
            <td :colspan="columns.length" class="dt__state">
              <slot name="empty">
                <EmptyState compact :title="emptyTitle || t('ui.table.empty')" />
              </slot>
            </td>
          </tr>
          <template v-else>
            <tr v-for="row in pageRows" :key="row[rowKey]" class="dt__row">
              <td v-for="col in columns" :key="col.key" :class="`dt__cell--${col.align || 'start'}`">
                <slot :name="`cell-${col.key}`" :row="row" :value="row[col.key]">{{ row[col.key] }}</slot>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <nav v-if="!loading && total > 0" class="dt__pager" :aria-label="t('ui.table.range', { from, to, total })">
      <span class="dt__range">{{ t('ui.table.range', { from, to, total }) }}</span>
      <template v-if="pageCount > 1">
        <button type="button" class="dt__page-btn" :disabled="page === 1" :aria-label="t('ui.table.prev')" @click="page -= 1">
          <Icon icon="mdi:chevron-left" class="dt__chev" aria-hidden="true" />
        </button>
        <span class="dt__page">{{ page }} / {{ pageCount }}</span>
        <button type="button" class="dt__page-btn" :disabled="page === pageCount" :aria-label="t('ui.table.next')" @click="page += 1">
          <Icon icon="mdi:chevron-right" class="dt__chev" aria-hidden="true" />
        </button>
      </template>
    </nav>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import EmptyState from './EmptyState.vue'

const props = defineProps({
  // [{ key, label, sortable?, align?: 'start'|'end'|'center', width? }]
  columns: { type: Array, required: true },
  rows: { type: Array, default: () => [] },
  rowKey: { type: String, default: 'id' },
  searchable: { type: Boolean, default: true },
  // Clés sur lesquelles porte la recherche ; par défaut toutes les colonnes.
  searchKeys: { type: Array, default: null },
  searchPlaceholder: { type: String, default: '' },
  pageSize: { type: Number, default: 20 },
  loading: { type: Boolean, default: false },
  emptyTitle: { type: String, default: '' },
})

const { t } = useI18n()

const query = ref('')
const sortKey = ref('')
const sortDir = ref('asc')
const page = ref(1)

// Insensible à la casse et aux accents.
const normalize = (value) =>
  String(value ?? '').normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase()

const filtered = computed(() => {
  const q = normalize(query.value.trim())
  if (!q) return props.rows
  const keys = props.searchKeys || props.columns.map(c => c.key)
  return props.rows.filter(row => keys.some(key => normalize(row[key]).includes(q)))
})

const compare = (a, b) => {
  if (a == null && b == null) return 0
  if (a == null) return 1
  if (b == null) return -1
  if (typeof a === 'number' && typeof b === 'number') return a - b
  return String(a).localeCompare(String(b), undefined, { numeric: true, sensitivity: 'base' })
}

const sorted = computed(() => {
  if (!sortKey.value) return filtered.value
  const dir = sortDir.value === 'asc' ? 1 : -1
  return [...filtered.value].sort((x, y) => dir * compare(x[sortKey.value], y[sortKey.value]))
})

const total = computed(() => sorted.value.length)
const pageCount = computed(() => Math.max(1, Math.ceil(total.value / props.pageSize)))
const start = computed(() => (page.value - 1) * props.pageSize)
const pageRows = computed(() => sorted.value.slice(start.value, start.value + props.pageSize))
const from = computed(() => (total.value === 0 ? 0 : start.value + 1))
const to = computed(() => Math.min(start.value + props.pageSize, total.value))

watch(query, () => { page.value = 1 })
watch(pageCount, (count) => { if (page.value > count) page.value = count })

const toggleSort = (key) => {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'asc'
  }
}

const ariaSort = (key) => {
  if (sortKey.value !== key) return 'none'
  return sortDir.value === 'asc' ? 'ascending' : 'descending'
}

const sortIcon = (key) => {
  if (sortKey.value !== key) return 'mdi:unfold-more-horizontal'
  return sortDir.value === 'asc' ? 'mdi:chevron-up' : 'mdi:chevron-down'
}
</script>

<style scoped>
.dt {
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--r-l);
  min-width: 0;
}

.dt__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-block-end: 1px solid var(--line);
}

.dt__search {
  position: relative;
  display: flex;
  align-items: center;
  flex: 1 1 220px;
  max-width: 340px;
}

.dt__search-icon {
  position: absolute;
  inset-inline-start: 9px;
  width: 16px;
  height: 16px;
  color: var(--muted);
  pointer-events: none;
}

.dt__search-input {
  width: 100%;
  height: 32px;
  padding-inline: 32px 10px;
  border: 1px solid var(--line);
  border-radius: var(--r-m);
  background: var(--bg);
  color: var(--text);
  font: inherit;
  font-size: 13px;
}

.dt__search-input:focus {
  outline: 2px solid var(--acc);
  outline-offset: -1px;
  border-color: transparent;
}

.dt__scroll { overflow-x: auto; }

.dt__table {
  width: 100%;
  border-collapse: collapse;
}

.dt__table th {
  height: 36px;
  padding-inline: 12px;
  font-family: var(--f-num);
  font-size: 10.5px;
  font-weight: 500;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
  background: var(--sunken);
  border-block-end: 1px solid var(--line);
  white-space: nowrap;
}

.dt__table td {
  height: 40px;
  padding: 6px 12px;
  border-block-end: 1px solid var(--line);
  color: var(--text);
  vertical-align: middle;
}

.dt__row:last-child td { border-block-end: 0; }
.dt__row:hover { background: var(--sunken); }

.dt__cell--start { text-align: start; }
.dt__cell--end { text-align: end; }
.dt__cell--center { text-align: center; }

.dt__sort {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 0;
  border: 0;
  background: none;
  color: inherit;
  font: inherit;
  letter-spacing: inherit;
  text-transform: inherit;
  cursor: pointer;
}

.dt__sort:hover { color: var(--text); }

.dt__sort:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 2px;
  border-radius: var(--r-s);
}

.dt__sort-icon {
  width: 14px;
  height: 14px;
}

.dt__state { text-align: center; }

.dt__spin {
  width: 20px;
  height: 20px;
  color: var(--muted);
  animation: dt-spin 0.8s linear infinite;
}

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

@keyframes dt-spin {
  to { transform: rotate(360deg); }
}
</style>
