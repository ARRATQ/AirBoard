<template>
  <div class="content-area">
    <PageHeader
      :eyebrow="$t(isAdmin ? 'nav.admin.title' : 'groupAdmin.title')"
      :title="$t('adminHome.title')"
      :description="$t(isAdmin ? 'adminHome.subtitleAdmin' : 'adminHome.subtitleGroup')"
    />

    <!-- Chiffres clés -->
    <section class="overview-kpis" :aria-label="$t('adminHome.keyFigures')">
      <router-link
        v-for="kpi in kpis"
        :key="kpi.key"
        :to="kpi.to"
        class="overview-kpi"
      >
        <span class="overview-kpi__label">{{ $t(kpi.labelKey) }}</span>
        <span class="overview-kpi__value">
          <FlapDigits v-if="kpi.value !== null" :value="kpi.value" />
          <span v-else aria-hidden="true">—</span>
        </span>
      </router-link>
    </section>

    <div class="overview-grid">
      <!-- À traiter (admin global) -->
      <section v-if="isAdmin" class="overview-panel overview-panel--wide" aria-labelledby="overview-todo">
        <h2 id="overview-todo" class="overview-panel__title">{{ $t('adminHome.todo') }}</h2>
        <ul class="overview-rows">
          <li v-for="item in todo" :key="item.key">
            <router-link :to="item.to" class="overview-row">
              <span>{{ $t(item.labelKey) }}</span>
              <StatusBadge v-if="item.value > 0" tone="warn">{{ item.value }}</StatusBadge>
              <span v-else-if="item.value === 0" class="overview-row__meta">{{ $t('adminHome.nothingPending') }}</span>
              <span v-else class="overview-row__meta">—</span>
            </router-link>
          </li>
        </ul>
      </section>

      <!-- Groupes administrés (admin de groupe) -->
      <section v-else class="overview-panel overview-panel--wide" aria-labelledby="overview-groups">
        <h2 id="overview-groups" class="overview-panel__title">{{ $t('groupAdmin.yourGroups') }}</h2>
        <p v-if="!loading && managedGroups.length === 0" class="overview-panel__empty">
          {{ $t('groupAdmin.noManagedGroups') }}
        </p>
        <ul v-else class="overview-rows">
          <li v-for="group in managedGroups" :key="group.id" class="overview-row">
            <span>{{ group.name }}</span>
            <span class="overview-row__meta">
              {{ group.user_count ?? 0 }} {{ $t('common.users') }}
            </span>
          </li>
        </ul>
      </section>

      <!-- Raccourcis vers chaque domaine du menu -->
      <section
        v-for="section in shortcutSections"
        :key="section.id"
        class="overview-panel"
        :aria-labelledby="`overview-${section.id}`"
      >
        <h2 :id="`overview-${section.id}`" class="overview-panel__title">{{ $t(section.labelKey) }}</h2>
        <ul>
          <li v-for="item in section.items" :key="item.to">
            <router-link :to="item.to" class="overview-link">
              <Icon :icon="item.icon" class="h-4 w-4 flex-shrink-0" aria-hidden="true" />
              <span>{{ $t(item.labelKey) }}</span>
            </router-link>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>

<script setup>
// Accueil de l'administration : chiffres clés, ce qui attend une action, et un
// accès direct à chaque domaine du menu. Même page pour l'admin global et l'admin
// de groupe ; le contenu suit le périmètre (config/access.js).
import { computed, onMounted, ref } from 'vue'
import { Icon } from '@iconify/vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import FlapDigits from '@/components/ui/FlapDigits.vue'
import { GROUP_ADMIN } from '@/config/access'
import { buildAdminNav } from '@/config/navigation'
import { useManagementScope } from '@/composables/useManagementScope'
import {
  dashboardService,
  commentsService,
  suggestionsService,
  groupAdminService
} from '@/services/api'

const { roles, isAdmin } = useManagementScope(GROUP_ADMIN)

const loading = ref(true)
const counts = ref({})
const managedGroups = ref([])

const shortcutSections = computed(() => {
  const nav = buildAdminNav(roles.value)
  if (!nav) return []
  if (nav.grouped) return nav.sections
  return [{ id: 'manage', labelKey: 'adminHome.manage', items: nav.items }]
})

const kpis = computed(() => {
  const c = counts.value
  const shared = [
    { key: 'resources', labelKey: 'nav.admin.resources', to: '/admin/resources', value: c.resources ?? null },
    { key: 'categories', labelKey: 'adminHome.resourceCategories', to: '/admin/resources/categories', value: c.categories ?? null }
  ]
  if (isAdmin.value) {
    return [
      { key: 'users', labelKey: 'nav.admin.users', to: '/admin/users', value: c.users ?? null },
      { key: 'groups', labelKey: 'nav.admin.groups', to: '/admin/groups', value: c.groups ?? null },
      ...shared
    ]
  }
  return [
    { key: 'groups', labelKey: 'adminHome.managedGroups', to: '/admin', value: c.groups ?? null },
    ...shared
  ]
})

const todo = computed(() => [
  { key: 'comments', labelKey: 'adminHome.pendingComments', to: '/admin/comments', value: counts.value.pendingComments ?? null },
  { key: 'suggestions', labelKey: 'adminHome.newSuggestions', to: '/admin/suggestions', value: counts.value.newSuggestions ?? null }
])

// Nombre d'éléments d'une réponse de liste, quelle que soit sa forme
const countOf = (data, key) => {
  if (Array.isArray(data)) return data.length
  if (typeof data?.total === 'number') return data.total
  const list = key ? data?.[key] : null
  return Array.isArray(list) ? list.length : null
}

// Chaque chiffre est chargé à part : une source en panne n'efface pas les autres
const settle = async (loaders) => {
  const entries = Object.entries(loaders)
  const results = await Promise.allSettled(entries.map(([, load]) => load()))
  return Object.fromEntries(
    entries.map(([name], i) => [name, results[i].status === 'fulfilled' ? results[i].value : null])
  )
}

const loadAdmin = async () => {
  const data = await settle({
    stats: () => dashboardService.getDashboard().then((d) => d.stats),
    pendingComments: () => commentsService.getPendingComments({ page_size: 1 }).then((d) => countOf(d, 'comments')),
    newSuggestions: () => suggestionsService
      .getAdminSuggestions({ status: 'new', page: 1, page_size: 1 })
      .then((d) => countOf(d, 'suggestions'))
  })
  counts.value = {
    users: data.stats?.total_users ?? null,
    groups: data.stats?.total_groups ?? null,
    resources: data.stats?.total_applications ?? null,
    categories: data.stats?.total_app_groups ?? null,
    pendingComments: data.pendingComments,
    newSuggestions: data.newSuggestions
  }
}

const loadGroupAdmin = async () => {
  const data = await settle({
    groups: () => groupAdminService.getManagedGroups(),
    categories: () => groupAdminService.getAppGroups().then((d) => countOf(d, 'app_groups')),
    resources: () => groupAdminService.getApplications().then((d) => countOf(d, 'applications'))
  })
  managedGroups.value = Array.isArray(data.groups) ? data.groups : []
  counts.value = {
    groups: Array.isArray(data.groups) ? data.groups.length : null,
    categories: data.categories,
    resources: data.resources
  }
}

onMounted(async () => {
  try {
    await (isAdmin.value ? loadAdmin() : loadGroupAdmin())
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.overview-kpis {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
  gap: 12px;
  margin-block-end: 24px;
}

.overview-kpi {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px 16px;
  border-radius: 6px;
  background-color: var(--board);
  color: var(--board-text);
  transition: transform 120ms ease;
}

.overview-kpi:hover {
  transform: translateY(-1px);
}

.overview-kpi__label {
  font-family: var(--f-num);
  font-size: 11px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--board-dim);
}

.overview-kpi__value {
  font-family: var(--f-num);
  font-size: 26px;
  font-weight: 500;
  line-height: 1;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
  align-items: start;
}

.overview-panel {
  padding: 16px;
  border: 1px solid var(--line);
  border-radius: 6px;
  background-color: var(--surface);
}

/* « À traiter » ou « Vos groupes » prend deux colonnes quand il y a la place */
.overview-panel--wide {
  grid-column: span 2;
}

@media (max-width: 640px) {
  .overview-panel--wide {
    grid-column: auto;
  }
}

.overview-panel__title {
  margin-block-end: 10px;
  font-family: var(--f-num);
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
}

.overview-panel__empty {
  font-size: 13.5px;
  color: var(--text-muted);
}

.overview-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 40px;
  padding-inline: 8px;
  border-block-end: 1px solid var(--line);
  font-size: 13.5px;
  color: var(--text);
}

.overview-rows > li:last-child,
.overview-rows > li:last-child > .overview-row {
  border-block-end: 0;
}

a.overview-row:hover,
.overview-link:hover {
  background-color: var(--hover);
}

.overview-row__meta {
  font-family: var(--f-num);
  font-size: 12px;
  color: var(--text-muted);
}

.overview-link {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 34px;
  padding-inline: 8px;
  border-radius: 4px;
  font-size: 13.5px;
  color: var(--text);
}

.overview-kpi:focus-visible,
.overview-link:focus-visible,
a.overview-row:focus-visible {
  outline: 2px solid var(--fill);
  outline-offset: 2px;
}

@media (prefers-reduced-motion: reduce) {
  .overview-kpi {
    transition: none;
  }

  .overview-kpi:hover {
    transform: none;
  }
}
</style>
