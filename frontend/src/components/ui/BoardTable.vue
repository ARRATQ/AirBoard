<template>
  <!-- Tableau d'affichage (panneau des départs) : fond sombre dans les deux modes. -->
  <div class="board">
    <div class="board__scroll">
      <table class="board__table" :style="{ minWidth }">
        <caption v-if="caption" class="sr-only">{{ caption }}</caption>
        <thead>
          <tr>
            <th v-for="col in columns" :key="col.key" scope="col" :class="`board__cell--${col.align || 'start'}`">
              {{ col.label }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rows" :key="row[rowKey]">
            <td
              v-for="col in columns"
              :key="col.key"
              :class="[`board__cell--${col.align || 'start'}`, col.tone ? `board__cell--${col.tone}` : '']"
            >
              <slot :name="`cell-${col.key}`" :row="row" :value="row[col.key]">{{ row[col.key] }}</slot>
            </td>
          </tr>
          <tr v-if="rows.length === 0">
            <td :colspan="columns.length" class="board__empty">{{ emptyText }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
defineProps({
  // [{ key, label, align?: 'start'|'end'|'center', tone?: 'hi'|'dim'|'text' }]
  columns: { type: Array, required: true },
  rows: { type: Array, default: () => [] },
  rowKey: { type: String, default: 'id' },
  caption: { type: String, default: '' },
  emptyText: { type: String, default: '—' },
  minWidth: { type: String, default: '640px' },
})
</script>

<style scoped>
.board {
  background: var(--board);
  color: var(--board-text);
  border-radius: var(--r-l);
  font-family: var(--f-num);
  font-size: 12.5px;
  overflow: hidden;
}

.board__scroll { overflow-x: auto; }

.board__table {
  width: 100%;
  border-collapse: collapse;
}

.board__table th {
  padding: 10px 12px 8px;
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--board-dim);
  border-block-end: 1px solid rgb(255 255 255 / 0.08);
  white-space: nowrap;
}

.board__table td {
  padding: 10px 12px;
  border-block-end: 1px solid rgb(255 255 255 / 0.06);
  vertical-align: middle;
  white-space: nowrap;
}

.board__table tbody tr:last-child td { border-block-end: 0; }
.board__table tbody tr:hover { background: rgb(255 255 255 / 0.03); }

.board__cell--start { text-align: start; }
.board__cell--end { text-align: end; }
.board__cell--center { text-align: center; }

.board__cell--hi { color: var(--board-hi); font-weight: 600; }
.board__cell--dim { color: var(--board-dim); }

.board__cell--text {
  font-family: var(--f-body);
  font-size: 13px;
  font-weight: 500;
  white-space: normal;
  min-width: 180px;
}

.board__empty {
  text-align: center;
  color: var(--board-dim);
}

/* Statuts dans le tableau : classes utilisables dans les slots. */
.board :deep(.board-ok) { color: var(--board-ok); }
.board :deep(.board-warn) { color: var(--board-hi); }
.board :deep(.board-bad) { color: var(--board-bad); }
</style>
