<template>
  <DataTable
    :columns="columns"
    :rows="rows"
    :search-keys="['name', 'description', 'categoryName', 'typeLabel']"
    :search-placeholder="$t('dashboard.searchApps')"
    :empty-title="$t('dashboard.noApplicationsFound')"
    :page-size="25"
  >
    <template #toolbar>
      <label class="rt__filter">
        <span class="sr-only">{{ $t('dashboard.category') }}</span>
        <select v-model="category" class="rt__select">
          <option value="">{{ $t('dashboard.allCategories') }}</option>
          <option v-for="group in appGroups" :key="group.id" :value="String(group.id)">{{ group.name }}</option>
        </select>
      </label>
    </template>

    <template #cell-favorite="{ row }">
      <button
        type="button"
        class="rt__star"
        :class="{ 'rt__star--on': favoritesStore.isFavorite(row.id) }"
        :aria-pressed="favoritesStore.isFavorite(row.id)"
        :aria-label="`${$t(favoritesStore.isFavorite(row.id) ? 'common.removeFromFavorites' : 'common.addToFavorites')} : ${row.name}`"
        @click="toggleFavorite(row)"
      >
        <Icon :icon="favoritesStore.isFavorite(row.id) ? 'mdi:star' : 'mdi:star-outline'" aria-hidden="true" />
      </button>
    </template>

    <template #cell-name="{ row }">
      <span class="rt__name">
        <a
          :href="row.url"
          :target="row.open_in_new_tab ? '_blank' : undefined"
          rel="noopener noreferrer"
          class="rt__link"
          @click="analyticsService.trackClick(row.id)"
        >{{ row.name }}</a>
        <StatusBadge v-if="row.isNew" tone="info">{{ $t('dashboard.newBadge') }}</StatusBadge>
      </span>
      <span v-if="row.description" class="rt__desc">{{ row.description }}</span>
      <span v-if="row.note" class="rt__note">{{ row.note }}</span>
    </template>

    <template #cell-typeLabel="{ row }">
      <span class="rt__type">{{ row.typeLabel }}</span>
    </template>

    <template #cell-createdAt="{ row }">
      <time class="rt__date" :datetime="row.createdAt">{{ formatDate(row.createdAt) }}</time>
    </template>

    <template #cell-details="{ row }">
      <button
        type="button"
        class="rt__icon-btn"
        :aria-label="`${$t('common.details')} : ${row.name}`"
        @click="$emit('show-details', row.raw)"
      >
        <Icon icon="mdi:information-outline" aria-hidden="true" />
      </button>
    </template>
  </DataTable>
</template>

<script setup>
// Vue Tableau du portail : recherche, tri (nom, type, catégorie, date d'ajout), filtre par
// catégorie, favoris, ouverture et détails. Le filtre par type est appliqué en amont (Dashboard).
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import DataTable from '@/components/ui/DataTable.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import { useFavoritesStore } from '@/stores/favorites'
import { useAppStore } from '@/stores/app'
import { analyticsService } from '@/services/api'
import { activeAdminNote, isNewResource, resourceTypeOf } from '@/utils/resourceType'
import { dateLocaleFor } from '@/utils/homeFormat'

const props = defineProps({
  appGroups: { type: Array, required: true },
})

defineEmits(['show-details'])

const { t, locale } = useI18n()
const favoritesStore = useFavoritesStore()
const appStore = useAppStore()
const category = ref('')

const columns = computed(() => [
  { key: 'favorite', label: '', width: '44px' },
  { key: 'name', label: t('dashboard.application'), sortable: true },
  { key: 'typeLabel', label: t('dashboard.type'), sortable: true },
  { key: 'categoryName', label: t('dashboard.category'), sortable: true },
  { key: 'createdAt', label: t('dashboard.addedOn'), sortable: true },
  { key: 'details', label: '', width: '44px', align: 'end' },
])

const rows = computed(() =>
  props.appGroups
    .filter((g) => !category.value || String(g.id) === category.value)
    .flatMap((g) =>
      (g.applications || []).map((app) => ({
        id: app.id,
        name: app.name,
        description: app.description || '',
        url: app.url,
        open_in_new_tab: app.open_in_new_tab,
        categoryName: g.name,
        typeLabel: t(`dashboard.types.${resourceTypeOf(app.url)}`),
        createdAt: app.created_at || '',
        isNew: isNewResource(app),
        note: activeAdminNote(app),
        raw: app,
      })),
    ),
)

const formatDate = (value) =>
  value ? new Date(value).toLocaleDateString(dateLocaleFor(locale.value), { day: '2-digit', month: '2-digit', year: 'numeric' }) : '—'

async function toggleFavorite(row) {
  try {
    await favoritesStore.toggleFavorite(row.id)
  } catch (error) {
    console.error('Error toggling favorite:', error)
    appStore.showError(t('ui.toast.error'))
  }
}
</script>

<style scoped>
.rt__select {
  height: 34px;
  padding-block: 0;
  padding-inline: 10px 32px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background-color: var(--surface);
  color: var(--text);
  font-size: 13px;
}

/* Chevron natif du select (plugin forms) côté fin de ligne en arabe. */
[dir='rtl'] .rt__select { padding-inline: 32px 10px; background-position: left 0.5rem center; }

.rt__name { display: inline-flex; flex-wrap: wrap; align-items: center; gap: 6px; }
.rt__link { font-weight: 600; color: var(--text); }
.rt__link:hover { color: var(--acc-ink); text-decoration: underline; }

.rt__link:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 2px;
  border-radius: 2px;
}

.rt__desc,
.rt__note {
  display: block;
  max-width: 60ch;
  font-size: 12px;
  color: var(--muted);
}

.rt__note {
  margin-block-start: 4px;
  padding-inline-start: 7px;
  border-inline-start: 2px solid var(--fill);
  color: var(--text);
}

.rt__type {
  font-family: var(--f-num);
  font-size: 10.5px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--muted);
  white-space: nowrap;
}

.rt__date { font-family: var(--f-num); font-size: 12px; white-space: nowrap; }

.rt__star,
.rt__icon-btn {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border: 0;
  border-radius: 3px;
  background: none;
  color: var(--muted);
  font-size: 17px;
  cursor: pointer;
}

.rt__star:hover,
.rt__icon-btn:hover { background: var(--sunken); color: var(--text); }
.rt__star--on { color: #d99a00; }

.rt__star:focus-visible,
.rt__icon-btn:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 1px;
}
</style>
