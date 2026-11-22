<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ isGlobal ? $t('common.appGroups') : $t('groupAdmin.appGroupsManagement.title') }}</template>
      <template #description>{{ isGlobal ? $t('appGroups.subtitle') : $t('groupAdmin.appGroupsManagement.subtitle') }}</template>
      <template #actions>
        <BaseButton variant="primary" icon="lucide:plus" :disabled="!canCreate" @click="openCreateModal">
          {{ $t('appGroups.new') }}
        </BaseButton>
      </template>
    </PageHeader>

    <!-- Admin de groupe sans groupe administré : création impossible -->
    <div v-if="showNoManagedGroups" class="ag-callout" role="note">
      <Icon icon="lucide:info" class="ag-callout__icon" aria-hidden="true" />
      <div>
        <p class="ag-callout__title">{{ $t('groupAdmin.appGroupsManagement.noManagedGroups') }}</p>
        <p class="ag-callout__text">{{ $t('groupAdmin.appGroupsManagement.noManagedGroupsHelp') }}</p>
      </div>
    </div>

    <EmptyState
      v-if="!loading && appGroups.length === 0"
      icon="lucide:folders"
      :title="$t('groupAdmin.appGroupsManagement.empty.title')"
      :description="$t('groupAdmin.appGroupsManagement.empty.description')"
    >
      <BaseButton variant="primary" size="sm" icon="lucide:plus" :disabled="!canCreate" @click="openCreateModal">
        {{ $t('groupAdmin.appGroupsManagement.empty.createButton') }}
      </BaseButton>
    </EmptyState>

    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :loading="loading"
      :search-keys="['name', 'description']"
      :search-placeholder="$t('groupAdmin.appGroupsManagement.searchPlaceholder')"
      :empty-title="$t('groupAdmin.appGroupsManagement.noResults.title')"
      :page-size="25"
    >
      <template #toolbar>
        <label>
          <span class="sr-only">{{ $t('ui.table.columns.status') }}</span>
          <select v-model="statusFilter" class="ui-select">
            <option value="">{{ $t('groupAdmin.appGroupsManagement.status.all') }}</option>
            <option value="active">{{ $t('groupAdmin.appGroupsManagement.status.active') }}</option>
            <option value="inactive">{{ $t('groupAdmin.appGroupsManagement.status.inactive') }}</option>
          </select>
        </label>
      </template>

      <template #cell-name="{ row }">
        <span class="ag-name">
          <Icon :icon="row.raw.icon || 'lucide:folder'" class="ag-name__icon" :style="{ color: row.raw.color || undefined }" aria-hidden="true" />
          <span class="ag-name__text">
            <span class="ag-name__title">{{ row.name }}</span>
            <span v-if="row.description" class="ag-name__desc">{{ row.description }}</span>
          </span>
        </span>
      </template>

      <template #cell-visibility="{ row }">
        <span class="ag-vis">
          <StatusBadge :tone="row.raw.is_private ? 'info' : 'neutral'">
            {{ row.raw.is_private ? $t('adminShared.appGroups.private') : $t('adminShared.appGroups.public') }}
          </StatusBadge>
          <span v-if="row.raw.is_private && row.raw.owner_group" class="ag-vis__owner">
            {{ $t('groupAdmin.appGroupsManagement.ownerGroup') }} {{ row.raw.owner_group.name }}
          </span>
        </span>
      </template>

      <template #cell-count="{ row }">
        <span class="ag-mono">{{ row.count }}</span>
      </template>

      <template #cell-status="{ row }">
        <StatusBadge :tone="row.raw.is_active ? 'ok' : 'neutral'">{{ row.status }}</StatusBadge>
      </template>

      <template #cell-order="{ row }">
        <span class="ag-mono">{{ row.order }}</span>
      </template>

      <template #cell-actions="{ row }">
        <RowActions :label="row.name" @edit="editGroup(row.raw)" @delete="confirmDelete(row.raw)" />
      </template>
    </DataTable>

    <AppGroupModal
      :show="showModal"
      :app-group="selectedAppGroup"
      :managed-groups="managedGroups"
      @close="closeModal"
      @submit="handleSubmit"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import RowActions from '@/components/ui/RowActions.vue'
import { useConfirm } from '@/composables/useConfirm'
import { adminService, groupAdminService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { useManagementScope } from '@/composables/useManagementScope'
import { RESOURCES } from '@/config/access'
import AppGroupModal from '@/components/admin/AppGroupModal.vue'

const { t } = useI18n()
const appStore = useAppStore()
const { confirm: askConfirm } = useConfirm()

// Admin global : tous les groupes de ressources (API admin) ;
// admin de groupe : groupes de ressources privés de ses groupes (API group-admin)
const { isGlobal } = useManagementScope(RESOURCES)
const service = computed(() => (isGlobal.value ? adminService : groupAdminService))

// State
const appGroups = ref([])
const managedGroups = ref([])
const managedGroupsLoaded = ref(false)
const loading = ref(false)
const showModal = ref(false)
const selectedAppGroup = ref(null)
const statusFilter = ref('')

// Les réponses arrivent soit en tableau, soit dans { data: [...] }
const toList = (data) => {
  if (Array.isArray(data)) return data
  return Array.isArray(data?.data) ? data.data : []
}

// Computed
// L'admin de groupe ne peut créer que s'il administre au moins un groupe (groupe propriétaire obligatoire)
const canCreate = computed(() => isGlobal.value || managedGroups.value.length > 0)
const showNoManagedGroups = computed(() =>
  !isGlobal.value && managedGroupsLoaded.value && managedGroups.value.length === 0
)

const columns = computed(() => [
  { key: 'name', label: t('ui.table.columns.name'), sortable: true },
  { key: 'visibility', label: t('ui.table.columns.visibility'), sortable: true, hideSm: true },
  { key: 'count', label: t('groupAdmin.appGroupsManagement.applications'), sortable: true, align: 'end', hideSm: true },
  { key: 'status', label: t('ui.table.columns.status'), sortable: true },
  { key: 'order', label: t('ui.table.columns.order'), sortable: true, align: 'end', width: '80px', hideSm: true },
  { key: 'actions', label: '', align: 'end', width: '80px' },
])

// Lignes du tableau : filtre de statut ; la recherche et le tri sont dans DataTable.
const rows = computed(() =>
  appGroups.value
    .filter(group => {
      if (statusFilter.value === 'active') return group.is_active
      if (statusFilter.value === 'inactive') return !group.is_active
      return true
    })
    .sort((a, b) => (a.order || 0) - (b.order || 0))
    .map(group => ({
      id: group.id,
      name: group.name,
      description: group.description || '',
      visibility: group.is_private ? t('adminShared.appGroups.private') : t('adminShared.appGroups.public'),
      count: group.applications?.length || 0,
      status: group.is_active ? t('common.active') : t('common.inactive'),
      order: group.order || 0,
      raw: group,
    })),
)

// Methods
const loadAppGroups = async () => {
  try {
    loading.value = true
    appGroups.value = toList(await service.value.getAppGroups())
  } catch (error) {
    console.error('Error loading app groups:', error)
    appGroups.value = []
    appStore.showError(t('adminShared.appGroups.loadError'))
  } finally {
    loading.value = false
  }
}

const loadManagedGroups = async () => {
  try {
    managedGroups.value = toList(await groupAdminService.getManagedGroups())
  } catch (error) {
    console.error('Error loading managed groups:', error)
    managedGroups.value = []
  } finally {
    managedGroupsLoaded.value = true
  }
}

const openCreateModal = () => {
  if (!canCreate.value) return
  selectedAppGroup.value = null
  showModal.value = true
}

const editGroup = (group) => {
  selectedAppGroup.value = { ...group }
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  selectedAppGroup.value = null
}

const handleSubmit = async (formData) => {
  try {
    if (formData.id) {
      await service.value.updateAppGroup(formData.id, formData)
      appStore.showSuccess(t('adminShared.appGroups.updated'))
    } else {
      await service.value.createAppGroup(formData)
      appStore.showSuccess(t('adminShared.appGroups.created'))
    }

    closeModal()
    await loadAppGroups()
  } catch (error) {
    console.error('Error saving app group:', error)
    appStore.showError(t('adminShared.appGroups.saveError'))
  }
}

const resourceCount = (group) => group?.applications?.length || 0

const confirmDelete = async (group) => {
  const confirmed = await askConfirm({
    title: t('groupAdmin.appGroupsManagement.deleteConfirm.title'),
    // Les ressources de la catégorie sont supprimées définitivement avec elle : le dire
    message: resourceCount(group)
      ? t('groupAdmin.appGroupsManagement.deleteConfirm.messageWithResources', { name: group.name, count: resourceCount(group) }, resourceCount(group))
      : t('groupAdmin.appGroupsManagement.deleteConfirm.message', { name: group.name }),
    confirmLabel: t('common.delete'),
    tone: 'danger',
  })
  if (!confirmed || !group?.id) return

  try {
    await service.value.deleteAppGroup(group.id)
    appStore.showSuccess(t('adminShared.appGroups.deleted'))
    await loadAppGroups()
  } catch (error) {
    console.error('Error deleting app group:', error)
    appStore.showError(t('groupAdmin.appGroupsManagement.deleteError'))
  }
}

// Lifecycle
onMounted(async () => {
  if (!isGlobal.value) {
    await loadManagedGroups()
  }
  await loadAppGroups()
})
</script>

<style scoped>
.ag-callout {
  display: flex;
  gap: 10px;
  margin-block-end: 16px;
  padding: 12px 14px;
  border: 1px solid var(--line);
  border-inline-start: 3px solid var(--warn);
  border-radius: var(--r-s);
  background: var(--surface);
}

.ag-callout__icon { flex: none; width: 18px; height: 18px; margin-block-start: 1px; color: var(--warn); }
.ag-callout__title { margin: 0; font-size: 13.5px; font-weight: 600; color: var(--text); }
.ag-callout__text { margin: 2px 0 0; font-size: 12.5px; color: var(--muted); }

.ag-name { display: flex; align-items: flex-start; gap: 10px; min-width: 0; }
.ag-name__icon { flex: none; width: 20px; height: 20px; margin-block-start: 1px; color: var(--muted); }
.ag-name__text { display: flex; flex-direction: column; min-width: 0; }
.ag-name__title { font-weight: 600; color: var(--text); }

.ag-name__desc {
  max-width: 48ch;
  overflow: hidden;
  font-size: 12px;
  color: var(--muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 640px) {
  .ag-name__desc { max-width: 24ch; }
}

.ag-vis { display: flex; flex-direction: column; align-items: flex-start; gap: 4px; }
.ag-vis__owner { font-size: 12px; color: var(--muted); }

.ag-mono {
  font-family: var(--f-num);
  font-size: 11.5px;
  letter-spacing: 0.04em;
  color: var(--muted);
  white-space: nowrap;
}
</style>
