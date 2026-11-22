<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ $t('groups.title') }}</template>
      <template #description>{{ $t('groups.subtitle') }}</template>
      <template #actions>
        <BaseButton variant="primary" icon="lucide:plus" @click="openCreateModal">
          {{ $t('groups.newGroup') }}
        </BaseButton>
      </template>
    </PageHeader>

    <EmptyState
      v-if="!appStore.loading && groups.length === 0"
      icon="lucide:users"
      :title="$t('groups.emptyTitle')"
      :description="$t('groups.emptyText')"
    >
      <BaseButton variant="primary" size="sm" icon="lucide:plus" @click="openCreateModal">
        {{ $t('groups.createGroup') }}
      </BaseButton>
    </EmptyState>

    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :loading="appStore.loading"
      :search-keys="['name', 'description']"
      :search-placeholder="$t('groups.searchPlaceholder')"
      :empty-title="$t('groups.noResultsTitle')"
      :page-size="25"
    >
      <template #toolbar>
        <label>
          <span class="sr-only">{{ $t('ui.table.columns.status') }}</span>
          <select v-model="statusFilter" class="ui-select">
            <option value="">{{ $t('groups.status_all') }}</option>
            <option value="active">{{ $t('groups.status_active') }}</option>
            <option value="inactive">{{ $t('groups.status_inactive') }}</option>
          </select>
        </label>
      </template>

      <template #cell-name="{ row }">
        <span class="gm-name">
          <span class="gm-name__swatch" :style="{ backgroundColor: row.raw.color || undefined }">
            <Icon icon="lucide:users" aria-hidden="true" />
          </span>
          <span class="gm-name__text">
            <span class="gm-name__title">{{ row.name }}</span>
            <span v-if="row.description" class="gm-name__desc">{{ row.description }}</span>
          </span>
        </span>
      </template>

      <template #cell-members="{ row }">
        <span class="gm-mono">{{ $t('groups.membersCount', row.members) }}</span>
      </template>

      <template #cell-appGroups="{ row }">
        <span v-if="row.raw.app_groups && row.raw.app_groups.length > 0" class="gm-chips">
          <span
            v-for="appGroup in row.raw.app_groups"
            :key="appGroup.id"
            class="tag gm-chip"
            :style="appGroup.color ? { borderColor: appGroup.color, color: appGroup.color } : undefined"
          >
            <Icon v-if="appGroup.icon" :icon="appGroup.icon" class="gm-chip__icon" aria-hidden="true" />
            {{ appGroup.name }}
          </span>
        </span>
        <span v-else class="gm-mono">0 {{ $t('groups.appsLabel') }}</span>
      </template>

      <template #cell-status="{ row }">
        <StatusBadge :tone="row.raw.is_active ? 'ok' : 'bad'">{{ row.status }}</StatusBadge>
      </template>

      <template #cell-actions="{ row }">
        <RowActions :label="row.name" @edit="editGroup(row.raw)" @delete="confirmDelete(row.raw)" />
      </template>
    </DataTable>

    <!-- Group Modal -->
    <GroupModal
      :show="showModal"
      :group="selectedGroup"
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
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { adminService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import GroupModal from '@/components/admin/GroupModal.vue'

const appStore = useAppStore()
const { t } = useI18n()
const { confirm: askConfirm } = useConfirm()

// State
const groups = ref([])
const showModal = ref(false)
const selectedGroup = ref(null)
const statusFilter = ref('')

// Computed
const columns = computed(() => [
  { key: 'name', label: t('ui.table.columns.name'), sortable: true },
  { key: 'members', label: t('ui.table.columns.members'), sortable: true, align: 'end', hideSm: true },
  { key: 'appGroups', label: t('groups.appsAccess').replace(/\s*:\s*$/, ''), hideSm: true },
  { key: 'status', label: t('ui.table.columns.status'), sortable: true },
  { key: 'actions', label: '', align: 'end', width: '120px' },
])

// Lignes du tableau : filtre de statut ; la recherche et le tri sont dans DataTable.
const rows = computed(() =>
  groups.value
    .filter(group => {
      if (statusFilter.value === 'active') return group.is_active
      if (statusFilter.value === 'inactive') return !group.is_active
      return true
    })
    .sort((a, b) => a.name.localeCompare(b.name))
    .map(group => ({
      id: group.id,
      name: group.name,
      description: group.description || '',
      members: group.users?.length || 0,
      status: group.is_active ? t('common.active') : t('common.inactive'),
      raw: group,
    })),
)

// Methods
const loadGroups = async () => {
  try {
    appStore.setLoading(true)
    const data = await adminService.getGroups()
    groups.value = Array.isArray(data) ? data : []
  } catch (error) {
    console.error('Erreur lors du chargement des groupes:', error)
    groups.value = []
    appStore.showError(t('groups.loadError'))
  } finally {
    appStore.setLoading(false)
  }
}

const openCreateModal = () => {
  selectedGroup.value = null
  showModal.value = true
}

const editGroup = (group) => {
  selectedGroup.value = group
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  selectedGroup.value = null
}

const handleSubmit = async (formData) => {
  try {
    appStore.setLoading(true)
    
    if (formData.id) {
      // Update existing group - extract only the fields we need
      const { id, ...updateData } = formData
      await adminService.updateGroup(id, updateData)
      appStore.showSuccess(t('groups.updated'))
    } else {
      // Create new group - extract only the fields we need
      const { id, ...createData } = formData
      await adminService.createGroup(createData)
      appStore.showSuccess(t('groups.created'))
    }
    
    closeModal()
    await loadGroups()
  } catch (error) {
    console.error('Erreur lors de la sauvegarde:', error)
    appStore.showError(t('groups.saveError'))
  } finally {
    appStore.setLoading(false)
  }
}

const confirmDelete = async (group) => {
  const confirmed = await askConfirm({
    title: t('groups.confirmDeleteTitle'),
    message: t('groups.confirmDeleteText', { name: group.name }),
    confirmLabel: t('common.delete'),
    tone: 'danger',
  })
  if (!confirmed) return

  try {
    await adminService.deleteGroup(group.id)
    appStore.showSuccess(t('groups.deleted'))
    await loadGroups()
  } catch (error) {
    console.error('Erreur lors de la suppression:', error)
    appStore.showError(t('groups.deleteError'))
  }
}

// Lifecycle
onMounted(() => {
  loadGroups()
})
</script>
<style scoped>
.gm-name { display: flex; align-items: flex-start; gap: 10px; min-width: 0; }

.gm-name__swatch {
  flex: none;
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: var(--r-s);
  background: var(--acc);
  color: var(--acc-on, #fff);
  font-size: 16px;
}

.gm-name__text { display: flex; flex-direction: column; min-width: 0; }
.gm-name__title { font-weight: 600; color: var(--text); }

.gm-name__desc {
  max-width: 44ch;
  overflow: hidden;
  font-size: 12px;
  color: var(--muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 640px) {
  .gm-name__desc { max-width: 22ch; }
}

.gm-chips { display: flex; flex-wrap: wrap; gap: 4px; }


.gm-chip__icon { width: 12px; height: 12px; }

.gm-mono {
  font-family: var(--f-num);
  font-size: 11.5px;
  letter-spacing: 0.04em;
  color: var(--muted);
  white-space: nowrap;
}
</style>
