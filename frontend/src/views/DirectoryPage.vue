<template>
  <div class="directory-page">
    <PageHeader :eyebrow="$t('directory.eyebrow')" :title="$t('directory.title')" :description="$t('directory.subtitle')">
      <template v-if="canManage" #actions>
        <BaseButton tag="router-link" to="/admin/directory" size="sm" icon="mdi:cog-outline">{{ $t('directory.manage') }}</BaseButton>
      </template>
    </PageHeader>

    <dl v-if="!loading && !error" class="directory-figures">
      <div><dt>{{ $t('directory.statUnits') }}</dt><dd>{{ units.length }}</dd></div>
      <div><dt>{{ $t('directory.statPeople') }}</dt><dd>{{ people.length }}</dd></div>
      <div :class="{ 'directory-figures--warn': vacantCount > 0 }"><dt>{{ $t('directory.statVacant') }}</dt><dd>{{ vacantCount }}</dd></div>
    </dl>

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
          <Icon :icon="tab.icon" class="h-4 w-4" aria-hidden="true" /> {{ $t(tab.label) }}
        </button>
      </div>

      <div class="directory-filters">
        <label class="directory-search">
          <Icon icon="mdi:magnify" class="h-4 w-4 shrink-0" aria-hidden="true" />
          <input v-model="search" type="search" :aria-label="$t('directory.searchPlaceholder')" :placeholder="$t('directory.searchPlaceholder')" @input="onSearch" />
        </label>
        <select v-if="view === 'grid'" v-model="unitFilter" class="form-select directory-unit-select">
          <option :value="null">{{ $t('directory.allUnits') }}</option>
          <option v-for="u in flatUnits" :key="u.id" :value="u.id">{{ '\u00a0\u00a0'.repeat(u.depth) }}{{ u.name }}</option>
        </select>
        <template v-else>
          <BaseButton size="sm" icon="mdi:unfold-more-horizontal" @click="toggleAll(true)">{{ $t('directory.expandAll') }}</BaseButton>
          <BaseButton size="sm" icon="mdi:unfold-less-horizontal" @click="toggleAll(false)">{{ $t('directory.collapseAll') }}</BaseButton>
        </template>
      </div>
    </div>

    <div v-if="loading" class="directory-state" role="status">
      <Icon icon="mdi:loading" class="animate-spin h-8 w-8" />
    </div>
    <div v-else-if="error" class="directory-state">
      <Icon icon="mdi:alert-circle-outline" class="h-8 w-8 text-red-500" />
      <p>{{ error }}</p>
      <BaseButton variant="secondary" size="sm" icon="mdi:refresh" @click="load">{{ $t('common.retry') }}</BaseButton>
    </div>
    <div v-else-if="!units.length && !people.length" class="directory-state">
      <Icon icon="mdi:sitemap-outline" class="h-10 w-10" />
      <p>{{ $t('directory.empty') }}</p>
    </div>

    <!-- Organigramme -->
    <template v-else-if="view === 'org'">
      <div class="directory-orgbar">
        <div class="directory-layout" role="group" :aria-label="$t('directory.viewOrg')">
          <button
            v-for="l in layouts"
            :key="l.id"
            type="button"
            class="directory-layout__btn"
            :class="{ 'directory-layout__btn--active': layout === l.id }"
            :aria-pressed="layout === l.id"
            @click="setLayout(l.id)"
          >
            <Icon :icon="l.icon" class="h-4 w-4" /> {{ $t(l.label) }}
          </button>
        </div>

        <nav v-if="focusUnit" class="directory-crumbs" :aria-label="$t('directory.focusUnit')">
          <button type="button" class="directory-crumbs__item" @click="setFocus(null)">
            <Icon icon="mdi:home-outline" class="h-4 w-4" /> {{ $t('directory.backToRoot') }}
          </button>
          <template v-for="u in unitPath(focusUnit.id)" :key="u.id">
            <Icon icon="mdi:chevron-right" class="h-4 w-4 directory-crumbs__sep" />
            <button
              type="button"
              class="directory-crumbs__item"
              :class="{ 'directory-crumbs__item--current': u.id === focusUnit.id }"
              @click="setFocus(u.id)"
            >{{ u.name }}</button>
          </template>
        </nav>

        <select
          class="form-select directory-unit-select directory-goto"
          :value="focusUnit ? focusUnit.id : ''"
          :aria-label="$t('directory.goToUnit')"
          @change="setFocus($event.target.value ? Number($event.target.value) : null)"
        >
          <option value="">{{ $t('directory.goToUnit') }}</option>
          <option v-for="u in flatUnits" :key="u.id" :value="u.id">{{ '\u00a0\u00a0'.repeat(u.depth) }}{{ u.name }}</option>
        </select>
      </div>

      <OrgTreeViewport v-if="layout === 'tree'" ref="treeViewport" :key="focusUnit ? focusUnit.id : 'all'">
        <ul class="ot-root" dir="ltr">
          <OrgTreeNode
            v-for="unit in displayedRoots"
            :key="unit.id"
            :unit="unit"
            :children-of="childrenOf"
            :people-by-id="peopleById"
            :count-people="countPeopleInSubtree"
            :force-open="forceOpen"
            :text-dir="textDir"
            @open-person="openPerson"
            @focus="setFocus"
          />
        </ul>
      </OrgTreeViewport>

      <ul v-else class="directory-org">
        <OrgChartNode
          v-for="unit in displayedRoots"
          :key="unit.id"
          :unit="unit"
          :children-of="childrenOf"
          :people-by-id="peopleById"
          :count-people="countPeopleInSubtree"
          :force-open="forceOpen"
          @open-person="openPerson"
        />
      </ul>
    </template>

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
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import OrgChartNode from '@/components/directory/OrgChartNode.vue'
import OrgTreeNode from '@/components/directory/OrgTreeNode.vue'
import OrgTreeViewport from '@/components/directory/OrgTreeViewport.vue'
import PersonCard from '@/components/directory/PersonCard.vue'
import PersonDetailModal from '@/components/directory/PersonDetailModal.vue'
import { useDirectory, fullName } from '@/components/directory/useDirectory'

const route = useRoute()
const router = useRouter()

const {
  loading, error, units, people, fields, canManage, load,
  peopleById, unitsById, childrenOf, rootUnits, flatUnits, rolesByPerson,
  unitPath, descendantIds, countPeopleInSubtree,
} = useDirectory()

const tabs = [
  { id: 'org', icon: 'mdi:sitemap-outline', label: 'directory.viewOrg' },
  { id: 'grid', icon: 'mdi:view-grid-outline', label: 'directory.viewGrid' },
]

const layouts = [
  { id: 'tree', icon: 'mdi:family-tree', label: 'directory.layoutTree' },
  { id: 'list', icon: 'mdi:format-list-bulleted-type', label: 'directory.layoutList' },
]
// L'arbre graphique est peu lisible sur petit écran : liste par défaut sous 768 px
const MOBILE_BREAKPOINT = 768
const defaultLayout = () => (window.innerWidth < MOBILE_BREAKPOINT ? 'list' : 'tree')

const { locale } = useI18n()
const textDir = computed(() => (locale.value === 'ar' ? 'rtl' : 'ltr'))

const view = ref(route.query.view === 'grid' ? 'grid' : 'org')
const layout = ref(['tree', 'list'].includes(route.query.layout) ? route.query.layout : defaultLayout())
const focusId = ref(route.query.focus ? Number(route.query.focus) : null)
const treeViewport = ref(null)
const search = ref(typeof route.query.q === 'string' ? route.query.q : '')
const unitFilter = ref(route.query.unit ? Number(route.query.unit) : null)
const selectedPerson = ref(null)
const forceOpen = ref(null)

const syncQuery = () => {
  const query = {}
  if (view.value === 'grid') query.view = 'grid'
  if (search.value) query.q = search.value
  if (unitFilter.value) query.unit = String(unitFilter.value)
  if (view.value === 'org') {
    if (layout.value !== defaultLayout()) query.layout = layout.value
    if (focusUnit.value) query.focus = String(focusUnit.value.id)
  }
  router.replace({ query })
}

// Unité affichée comme racine (null = organisation complète)
const focusUnit = computed(() => (focusId.value ? unitsById.value.get(focusId.value) || null : null))
const displayedRoots = computed(() => (focusUnit.value ? [focusUnit.value] : rootUnits.value))

const setLayout = (l) => {
  layout.value = l
  syncQuery()
}

const setFocus = (id) => {
  focusId.value = id
  syncQuery()
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
  requestAnimationFrame(() => {
    forceOpen.value = value
    // Recadrer l'arbre une fois les branches dépliées ou repliées
    requestAnimationFrame(() => treeViewport.value?.fit())
  })
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
const editPerson = (person) => router.push({ path: '/admin/directory', query: { person: person.id } })

onMounted(load)
</script>

<style scoped>
.directory-page {
  min-height: 100%;
  padding: clamp(0.75rem, 2.5vw, 1.5rem);
  max-width: 88rem;
  margin: 0 auto;
}

/* Chiffres clés : bande sobre, les postes vacants signalés seulement s'il y en a. */
.directory-figures {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin: 0 0 4px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
}

.directory-figures > div { padding: 10px 14px; border-inline-start: 1px solid var(--line); }
.directory-figures > div:first-child { border-inline-start: 0; }

.directory-figures dt {
  font-family: var(--f-num);
  font-size: 10.5px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--muted);
}

.directory-figures dd {
  margin: 2px 0 0;
  font-family: var(--f-num);
  font-size: 24px;
  font-weight: 600;
  line-height: 1.1;
  color: var(--text);
  font-variant-numeric: tabular-nums;
}

.directory-figures--warn dd { color: var(--warn); }

.directory-toolbar {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px 12px;
  margin: 12px 0 16px;
  padding-block: 8px;
  border-block-end: 1px solid var(--line);
  background: var(--bg);
}

.directory-tabs {
  display: inline-flex;
  gap: 2px;
  padding: 2px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--sunken);
}

.directory-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding-inline: 12px;
  border-radius: var(--r-s);
  font-size: 13px;
  color: var(--muted);
}

.directory-tab:hover { color: var(--text); }
.directory-tab--active { background: var(--surface); color: var(--text); font-weight: 600; box-shadow: 0 0 0 1px var(--line); }
.directory-tab:focus-visible { outline: 2px solid var(--acc); outline-offset: -2px; }

.directory-filters { flex: 1 1 auto; display: flex; flex-wrap: wrap; align-items: center; justify-content: flex-end; gap: 6px; }

.directory-search {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: min(100%, 18rem);
  height: 34px;
  padding-inline: 10px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
  color: var(--muted);
}

.directory-search:focus-within { border-color: var(--acc); box-shadow: 0 0 0 1px var(--acc); }
.directory-search input { flex: 1; min-width: 0; background: transparent; border: 0; outline: none; color: var(--text); font-size: 13px; }
.directory-unit-select { width: auto; max-width: 16rem; height: 34px; padding-block: 0; padding-inline: 10px 32px; border-radius: var(--r-s); font-size: 13px; }
/* Chevron natif du select (plugin forms) côté fin de ligne en arabe. */
[dir='rtl'] .directory-unit-select { padding-inline: 32px 10px; background-position: left 0.5rem center; }

.directory-orgbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}
.directory-layout {
  display: inline-flex;
  gap: 2px;
  padding: 2px;
  border-radius: var(--r-s);
  border: 1px solid var(--line);
  background: var(--sunken);
}
.directory-layout__btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.7rem;
  border-radius: var(--r-s);
  font-size: 12.5px;
  color: var(--muted);
  transition: background-color 150ms ease, color 150ms ease;
}
.directory-layout__btn:hover { color: var(--text-primary); }
.directory-layout__btn--active { background: var(--surface); color: var(--text); font-weight: 600; box-shadow: 0 0 0 1px var(--line); }
.directory-layout__btn:focus-visible, .directory-crumbs__item:focus-visible { outline: 2px solid var(--accent); }
.directory-crumbs {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.15rem;
  min-width: 0;
  font-size: 0.85rem;
}
.directory-crumbs__item {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.25rem 0.45rem;
  border-radius: var(--r-s);
  color: var(--text-muted);
}
.directory-crumbs__item:hover { background: var(--hover); color: var(--text-primary); }
.directory-crumbs__item--current { color: var(--text-primary); font-weight: 700; }
.directory-crumbs__sep { color: var(--text-muted); opacity: 0.6; }
[dir='rtl'] .directory-crumbs__sep { transform: scaleX(-1); }
.directory-goto { margin-inline-start: auto; }
.ot-root { display: flex; justify-content: center; padding: 0; margin: 0; }
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
  .directory-toolbar { position: static; }
  .directory-filters { width: 100%; }
  .directory-search { flex: 1 1 100%; }
  .directory-unit-select { max-width: 100%; width: 100%; }
  .directory-figures > div { padding: 8px 10px; }
  .directory-figures dd { font-size: 20px; }
}
@media (prefers-reduced-motion: reduce) {
  .directory-page * { transition: none !important; }
}
</style>
