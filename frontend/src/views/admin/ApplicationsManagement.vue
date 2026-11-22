<template>
  <div class="content-area">
    <!-- En-tête -->
    <PageHeader>
      <template #title>{{ $t('applications.title') }}</template>
      <template #description>
        {{ isGlobal ? $t('applications.subtitle') : $t('groupAdmin.applicationsManagement.subtitle') }}
      </template>
      <template #actions>
        <BaseButton variant="primary" icon="lucide:plus" :disabled="hasNoAppGroups" @click="openCreateModal">
          {{ $t('applications.new') }}
        </BaseButton>
      </template>
    </PageHeader>

    <!-- Aucun groupe de ressources : impossible d'ajouter une ressource -->
    <div v-if="hasNoAppGroups" class="am-callout" role="note">
      <Icon icon="lucide:info" class="am-callout__icon" aria-hidden="true" />
      <div>
        <p class="am-callout__title">{{ $t('groupAdmin.applicationsManagement.noAppGroups') }}</p>
        <p class="am-callout__text">{{ $t('groupAdmin.applicationsManagement.noAppGroupsHelp') }}</p>
      </div>
    </div>

    <!-- Aucune ressource : invitation à en créer une -->
    <EmptyState
      v-if="!appStore.loading && applications.length === 0"
      icon="lucide:app-window"
      :title="$t('applications.emptyTitle')"
      :description="$t('applications.emptyText')"
    >
      <BaseButton variant="primary" size="sm" icon="lucide:plus" :disabled="hasNoAppGroups" @click="openCreateModal">
        {{ $t('applications.add') }}
      </BaseButton>
    </EmptyState>

    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :loading="appStore.loading"
      :search-keys="['name', 'description', 'url', 'groupName']"
      :search-placeholder="$t('applications.searchPlaceholder')"
      :empty-title="$t('applications.noResultsTitle')"
      :page-size="25"
    >
      <template #toolbar>
        <label>
          <span class="sr-only">{{ $t('ui.table.columns.category') }}</span>
          <select v-model="groupFilter" class="ui-select">
            <option value="">{{ $t('applications.filter_all_groups') }}</option>
            <option v-for="group in appGroups" :key="group.id" :value="String(group.id)">{{ group.name }}</option>
          </select>
        </label>
        <label>
          <span class="sr-only">{{ $t('ui.table.columns.status') }}</span>
          <select v-model="statusFilter" class="ui-select">
            <option value="">{{ $t('common.status_all') }}</option>
            <option value="active">{{ $t('common.status_active') }}</option>
            <option value="inactive">{{ $t('common.status_inactive') }}</option>
          </select>
        </label>
      </template>

      <template #cell-name="{ row }">
        <span class="am-name">
          <Icon :icon="row.raw.icon || 'lucide:app-window'" class="am-name__icon" :style="{ color: row.raw.color || undefined }" aria-hidden="true" />
          <span class="am-name__text">
            <span class="am-name__title">{{ row.name }}</span>
            <a :href="row.url" target="_blank" rel="noopener noreferrer" class="am-name__url" dir="ltr">{{ row.url }}</a>
            <span v-if="row.note" class="am-name__note">{{ row.note }}</span>
          </span>
        </span>
      </template>

      <template #cell-typeLabel="{ row }">
        <span class="am-mono">{{ row.typeLabel }}</span>
      </template>

      <template #cell-status="{ row }">
        <StatusBadge :tone="row.raw.is_active ? 'ok' : 'neutral'">{{ row.status }}</StatusBadge>
      </template>

      <template #cell-order="{ row }">
        <span class="am-mono">{{ row.order }}</span>
      </template>

      <template #cell-actions="{ row }">
        <RowActions :label="row.name" @edit="editApplication(row.raw)" @delete="confirmDelete(row.raw)" />
      </template>
    </DataTable>

    <!-- Modale ressource -->
    <ApplicationModal
      :show="showModal"
      :application="selectedApplication"
      :app-groups="appGroups"
      :users="users"
      :private-groups-only="!isGlobal"
      @close="closeModal"
      @submit="handleSubmit"
    />

  </div>
</template>

<script setup>
import StatusBadge from '@/components/ui/StatusBadge.vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import RowActions from '@/components/ui/RowActions.vue'
import { useConfirm } from '@/composables/useConfirm'
import { activeAdminNote, resourceTypeOf } from '@/utils/resourceType'
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { adminService, groupAdminService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { useManagementScope } from '@/composables/useManagementScope'
import { RESOURCES } from '@/config/access'
import ApplicationModal from '@/components/admin/ApplicationModal.vue'

const { t } = useI18n()
const appStore = useAppStore()
const { confirm: askConfirm } = useConfirm()

// Admin global : toutes les ressources (API admin).
// Admin de groupe : ressources des groupes de ressources privés de ses groupes (API group-admin).
const { isGlobal, isAdmin } = useManagementScope(RESOURCES)
const service = computed(() => (isGlobal.value ? adminService : groupAdminService))

// État
const applications = ref([])
const appGroups = ref([])
const appGroupsLoaded = ref(false)
const users = ref([])
const showModal = ref(false)
const selectedApplication = ref(null)
const groupFilter = ref('')
const statusFilter = ref('')

/** Les services renvoient soit un tableau, soit une enveloppe { data } */
const toArray = (data) => (Array.isArray(data) ? data : (data?.data || []))

// Calculs
const hasNoAppGroups = computed(() => appGroupsLoaded.value && appGroups.value.length === 0)

const columns = computed(() => [
  { key: 'name', label: t('ui.table.columns.name'), sortable: true },
  { key: 'groupName', label: t('ui.table.columns.category'), sortable: true, hideSm: true },
  { key: 'typeLabel', label: t('ui.table.columns.type'), sortable: true, hideSm: true },
  { key: 'status', label: t('ui.table.columns.status'), sortable: true },
  { key: 'order', label: t('ui.table.columns.order'), sortable: true, align: 'end', width: '80px', hideSm: true },
  { key: 'actions', label: '', align: 'end', width: '80px' },
])

// Lignes du tableau : filtres catégorie et statut ; la recherche et le tri sont dans DataTable.
const rows = computed(() =>
  applications.value
    .filter(app => !groupFilter.value || String(app.app_group_id) === groupFilter.value)
    .filter(app => {
      if (statusFilter.value === 'active') return app.is_active
      if (statusFilter.value === 'inactive') return !app.is_active
      return true
    })
    .sort((a, b) => (a.order || 0) - (b.order || 0))
    .map(app => ({
      id: app.id,
      name: app.name,
      description: app.description || '',
      url: app.url,
      groupName: getAppGroupName(app.app_group_id),
      typeLabel: t(`dashboard.types.${resourceTypeOf(app.url)}`),
      status: app.is_active ? t('common.active') : t('common.inactive'),
      order: app.order || 0,
      note: activeAdminNote(app),
      raw: app,
    })),
)

// Chargements
const loadApplications = async () => {
  try {
    appStore.setLoading(true)
    const data = await service.value.getApplications({ limit: 1000 })
    applications.value = toArray(data)
  } catch (error) {
    console.error('Error loading applications:', error)
    applications.value = []
    appStore.showError(t('adminShared.applications.loadError'))
  } finally {
    appStore.setLoading(false)
  }
}

const loadAppGroups = async () => {
  try {
    const data = await service.value.getAppGroups()
    appGroups.value = toArray(data)
  } catch (error) {
    console.error('Error loading app groups:', error)
    appGroups.value = []
  } finally {
    appGroupsLoaded.value = true
  }
}

// Liste des utilisateurs pour le choix du créateur : endpoint réservé à l'admin global
const loadUsers = async () => {
  if (!isAdmin.value) return
  try {
    // Liste paginée côté serveur (25 par défaut) : la liste du créateur prend la page maximale.
    const data = await adminService.getUsers({ page_size: 500 })
    users.value = toArray(data)
  } catch (error) {
    console.error('Error loading users:', error)
    users.value = []
  }
}

const getAppGroupName = (appGroupId) => {
  const group = appGroups.value.find(g => g.id === appGroupId)
  return group ? group.name : t('common.noGroup')
}

// Modale
const openCreateModal = () => {
  if (hasNoAppGroups.value) return
  selectedApplication.value = null
  showModal.value = true
}

const editApplication = (application) => {
  selectedApplication.value = { ...application }
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  selectedApplication.value = null
}

const handleSubmit = async (formData) => {
  try {
    appStore.setLoading(true)

    if (formData.id) {
      await service.value.updateApplication(formData.id, formData)
      appStore.showSuccess(t('adminShared.applications.updated'))
    } else {
      await service.value.createApplication(formData)
      appStore.showSuccess(t('adminShared.applications.created'))
    }

    closeModal()
    await loadApplications()
  } catch (error) {
    console.error('Error saving application:', error)
    appStore.showError(t('adminShared.applications.saveError'))
  } finally {
    appStore.setLoading(false)
  }
}

const confirmDelete = async (application) => {
  const confirmed = await askConfirm({
    title: t('applications.confirmDeleteTitle'),
    message: t('applications.confirmDeleteText', { name: application.name }),
    confirmLabel: t('common.delete'),
    tone: 'danger',
  })
  if (!confirmed) return

  try {
    await service.value.deleteApplication(application.id)
    appStore.showSuccess(t('adminShared.applications.deleted'))
    await loadApplications()
  } catch (error) {
    console.error('Error deleting application:', error)
    appStore.showError(t('groupAdmin.applicationsManagement.deleteError'))
  }
}

onMounted(async () => {
  await Promise.all([
    loadApplications(),
    loadAppGroups(),
    loadUsers()
  ])
})
</script>

<style scoped>
.am-callout {
  display: flex;
  gap: 10px;
  margin-block-end: 16px;
  padding: 12px 14px;
  border: 1px solid var(--line);
  border-inline-start: 3px solid var(--warn);
  border-radius: var(--r-s);
  background: var(--surface);
}

.am-callout__icon { flex: none; width: 18px; height: 18px; margin-block-start: 1px; color: var(--warn); }
.am-callout__title { margin: 0; font-size: 13.5px; font-weight: 600; color: var(--text); }
.am-callout__text { margin: 2px 0 0; font-size: 12.5px; color: var(--muted); }

.am-name { display: flex; align-items: flex-start; gap: 10px; min-width: 0; }
.am-name__icon { flex: none; width: 20px; height: 20px; margin-block-start: 1px; color: var(--muted); }
.am-name__text { display: flex; flex-direction: column; min-width: 0; }
.am-name__title { font-weight: 600; color: var(--text); }

.am-name__url {
  max-width: 42ch;
  overflow: hidden;
  font-size: 12px;
  color: var(--muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 640px) {
  .am-name__url { max-width: 22ch; }
}

.am-name__url:hover { color: var(--acc-ink); text-decoration: underline; }

.am-name__note {
  margin-block-start: 4px;
  padding-inline-start: 7px;
  border-inline-start: 2px solid var(--fill);
  font-size: 12px;
  color: var(--text);
}

.am-mono {
  font-family: var(--f-num);
  font-size: 11.5px;
  letter-spacing: 0.04em;
  color: var(--muted);
  white-space: nowrap;
}
</style>
