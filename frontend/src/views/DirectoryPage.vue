<template>
  <div class="directory-page">
    <header class="directory-hero">
      <div class="directory-hero__text">
        <p class="directory-hero__eyebrow">{{ $t('directory.eyebrow') }}</p>
        <h1 class="directory-hero__title">{{ $t('directory.title') }}</h1>
        <p class="directory-hero__subtitle">{{ $t('directory.subtitle') }}</p>
      </div>
      <dl v-if="!loading && !error" class="directory-hero__stats">
        <div><dt>{{ $t('directory.statUnits') }}</dt><dd>{{ units.length }}</dd></div>
        <div><dt>{{ $t('directory.statPeople') }}</dt><dd>{{ people.length }}</dd></div>
        <div><dt>{{ $t('directory.statVacant') }}</dt><dd>{{ vacantCount }}</dd></div>
      </dl>
    </header>

    <div class="directory-toolbar">
      <div class="directory-tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          role="tab"
          :aria-selected="view === tab.id"
          class="directory-tab"
          :class="{ 'directory-tab--active': view === tab.id }"
          @click="setView(tab.id)"
        >
          <Icon :icon="tab.icon" class="h-4 w-4" /> {{ $t(tab.label) }}
        </button>
      </div>

      <div class="directory-filters">
        <label class="directory-search">
          <Icon icon="mdi:magnify" class="h-4 w-4 shrink-0" />
          <input v-model="search" type="search" :placeholder="$t('directory.searchPlaceholder')" @input="onSearch" />
        </label>
        <select v-if="view === 'grid'" v-model="unitFilter" class="form-select directory-unit-select">
          <option :value="null">{{ $t('directory.allUnits') }}</option>
          <option v-for="u in flatUnits" :key="u.id" :value="u.id">{{ '  '.repeat(u.depth) }}{{ u.name }}</option>
        </select>
        <template v-else>
          <button type="button" class="btn btn-ghost" @click="toggleAll(true)">
            <Icon icon="mdi:unfold-more-horizontal" class="h-4 w-4 mr-1" /> {{ $t('directory.expandAll') }}
          </button>
          <button type="button" class="btn btn-ghost" @click="toggleAll(false)">
            <Icon icon="mdi:unfold-less-horizontal" class="h-4 w-4 mr-1" /> {{ $t('directory.collapseAll') }}
          </button>
        </template>
        <router-link v-if="canManage" to="/group-admin/directory" class="btn btn-secondary">
          <Icon icon="mdi:cog-outline" class="h-4 w-4 mr-1" /> {{ $t('directory.manage') }}
        </router-link>
      </div>
    </div>

    <div v-if="loading" class="directory-state">
      <Icon icon="mdi:loading" class="animate-spin h-8 w-8" />
    </div>
    <div v-else-if="error" class="directory-state">
      <Icon icon="mdi:alert-circle-outline" class="h-8 w-8 text-red-500" />
      <p>{{ error }}</p>
      <button type="button" class="btn btn-secondary" @click="load">{{ $t('common.retry') }}</button>
    </div>
    <div v-else-if="!units.length && !people.length" class="directory-state">
      <Icon icon="mdi:sitemap-outline" class="h-10 w-10" />
      <p>{{ $t('directory.empty') }}</p>
    </div>

    <!-- Organigramme -->
    <ul v-else-if="view === 'org'" class="directory-org">
      <OrgChartNode
        v-for="unit in rootUnits"
        :key="unit.id"
        :unit="unit"
        :children-of="childrenOf"
        :people-by-id="peopleById"
        :count-people="countPeopleInSubtree"
        :force-open="forceOpen"
        @open-person="openPerson"
      />
    </ul>

    <!-- Grille de cartes -->
    <template v-else>
      <p class="directory-count">{{ $t('directory.resultsCount', { count: filteredPeople.length }) }}</p>
      <div v-if="filteredPeople.length" class="directory-grid">
        <PersonCard
          v-for="person in filteredPeople"
          :key="person.id"
          :person="person"
          :roles="rolesByPerson.get(person.id) || []"
          :fields="fields"
          @open="openPerson"
        />
      </div>
      <div v-else class="directory-state">
        <Icon icon="mdi:account-search-outline" class="h-10 w-10" />
        <p>{{ $t('directory.noResults') }}</p>
      </div>
    </template>

    <PersonDetailModal
      v-if="selectedPerson"
      :person="selectedPerson"
      :roles="rolesByPerson.get(selectedPerson.id) || []"
      :fields="fields"
      :unit-path="unitPath"
      :can-manage="canManage"
      @close="selectedPerson = null"
      @edit="editPerson"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import OrgChartNode from '@/components/directory/OrgChartNode.vue'
import PersonCard from '@/components/directory/PersonCard.vue'
import PersonDetailModal from '@/components/directory/PersonDetailModal.vue'
import { useDirectory, fullName } from '@/components/directory/useDirectory'

const route = useRoute()
const router = useRouter()

const {
  loading, error, units, people, fields, canManage, load,
  peopleById, childrenOf, rootUnits, flatUnits, rolesByPerson,
  unitPath, descendantIds, countPeopleInSubtree,
} = useDirectory()

const tabs = [
  { id: 'org', icon: 'mdi:sitemap-outline', label: 'directory.viewOrg' },
  { id: 'grid', icon: 'mdi:view-grid-outline', label: 'directory.viewGrid' },
]

const view = ref(route.query.view === 'grid' ? 'grid' : 'org')
const search = ref(typeof route.query.q === 'string' ? route.query.q : '')
const unitFilter = ref(route.query.unit ? Number(route.query.unit) : null)
const selectedPerson = ref(null)
const forceOpen = ref(null)

const syncQuery = () => {
  const query = {}
  if (view.value === 'grid') query.view = 'grid'
  if (search.value) query.q = search.value
  if (unitFilter.value) query.unit = String(unitFilter.value)
  router.replace({ query })
}

const setView = (v) => {
  view.value = v
  syncQuery()
}

// Une recherche s'affiche toujours dans la grille
const onSearch = () => {
  if (search.value && view.value !== 'grid') view.value = 'grid'
  syncQuery()
}

const toggleAll = (value) => {
  forceOpen.value = null
  requestAnimationFrame(() => { forceOpen.value = value })
}

const normalize = (s) => (s || '').toString().normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase()

const searchIndex = computed(() => {
  const index = new Map()
  for (const p of people.value) {
    const roles = rolesByPerson.value.get(p.id) || []
    const parts = [
      fullName(p), p.email, p.phone,
      ...roles.flatMap((r) => [r.position.title, r.unit.name, r.unit.code]),
      ...Object.values(p.custom_fields || {}),
    ]
    index.set(p.id, normalize(parts.join(' ')))
  }
  return index
})

const filteredPeople = computed(() => {
  const terms = normalize(search.value).split(/\s+/).filter(Boolean)
  const allowedUnits = unitFilter.value ? descendantIds(unitFilter.value) : null
  return people.value.filter((p) => {
    if (allowedUnits) {
      const roles = rolesByPerson.value.get(p.id) || []
      if (!roles.some((r) => allowedUnits.has(r.unit.id))) return false
    }
    const haystack = searchIndex.value.get(p.id) || ''
    return terms.every((t) => haystack.includes(t))
  })
})

const vacantCount = computed(() =>
  units.value.reduce((n, u) => n + (u.positions || []).filter((p) => !(p.assignments || []).length).length, 0)
)

const openPerson = (person) => { selectedPerson.value = person }
const editPerson = (person) => router.push({ path: '/group-admin/directory', query: { person: person.id } })

onMounted(load)
</script>

<style scoped>
.directory-page {
  min-height: 100%;
  padding: clamp(1rem, 2.5vw, 2rem);
  max-width: 88rem;
  margin: 0 auto;
}
.directory-hero {
  position: relative;
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1.5rem;
  padding: clamp(1.5rem, 3vw, 2.5rem);
  border-radius: 1.25rem;
  border: 1px solid var(--border);
  overflow: hidden;
  background:
    radial-gradient(80% 120% at 100% 0%, color-mix(in srgb, var(--accent) 18%, transparent), transparent 60%),
    radial-gradient(60% 90% at 0% 100%, color-mix(in srgb, var(--gradient-to, var(--accent)) 12%, transparent), transparent 70%),
    var(--bg-card);
}
.directory-hero__eyebrow {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--accent);
}
.directory-hero__title {
  font-size: clamp(2rem, 1.2rem + 3vw, 3.25rem);
  font-weight: 800;
  line-height: 1.05;
  letter-spacing: -0.02em;
  color: var(--text-primary);
}
.directory-hero__subtitle { margin-top: 0.5rem; max-width: 36rem; color: var(--text-muted); }
.directory-hero__stats { display: flex; gap: 0.75rem; }
.directory-hero__stats > div {
  min-width: 6rem;
  padding: 0.75rem 1rem;
  border-radius: 0.85rem;
  border: 1px solid var(--border);
  background: color-mix(in srgb, var(--bg-card) 75%, transparent);
  backdrop-filter: blur(6px);
}
.directory-hero__stats dt { font-size: 0.7rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-muted); }
.directory-hero__stats dd { font-size: 1.75rem; font-weight: 800; color: var(--text-primary); font-variant-numeric: tabular-nums; }

.directory-toolbar {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin: 1.25rem 0;
  padding: 0.6rem 0;
  background: var(--bg-page);
}
.directory-tabs {
  display: inline-flex;
  padding: 0.25rem;
  border-radius: 0.75rem;
  border: 1px solid var(--border);
  background: var(--bg-card);
}
.directory-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.9rem;
  border-radius: 0.55rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-muted);
  transition: background-color 150ms ease, color 150ms ease;
}
.directory-tab:hover { color: var(--text-primary); }
.directory-tab--active { background: var(--accent); color: #fff; }
.directory-tab--active:hover { color: #fff; }
.directory-filters { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem; }
.directory-search {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0 0.75rem;
  min-width: min(100%, 18rem);
  height: 2.5rem;
  border-radius: 0.65rem;
  border: 1px solid var(--border);
  background: var(--bg-card);
  color: var(--text-muted);
}
.directory-search:focus-within { border-color: var(--accent); }
.directory-search input { flex: 1; min-width: 0; background: transparent; border: 0; outline: none; color: var(--text-primary); }
.directory-unit-select { width: auto; max-width: 16rem; height: 2.5rem; }

.directory-org { display: flex; flex-direction: column; gap: 1rem; padding: 0; }
.directory-count { font-size: 0.8rem; color: var(--text-muted); margin-bottom: 0.75rem; }
.directory-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(13.5rem, 1fr));
  gap: 1rem;
}
.directory-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 4rem 1rem;
  color: var(--text-muted);
  text-align: center;
}
@media (max-width: 640px) {
  .directory-hero__stats { width: 100%; }
  .directory-hero__stats > div { flex: 1; min-width: 0; }
  .directory-unit-select { max-width: 100%; width: 100%; }
}
@media (prefers-reduced-motion: reduce) {
  .directory-page * { transition: none !important; }
}
</style>
