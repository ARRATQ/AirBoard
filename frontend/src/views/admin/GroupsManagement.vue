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
        <RowActions :label="row.name" @edit="editGroup(row.raw)" @delete="confirmDelete(row.raw)">
          <button
            type="button"
            class="row-actions__btn"
            :aria-label="$t('groupAdminsModal.manageFor', { name: row.name })"
            :title="$t('groupAdminsModal.manage')"
            @click="manageGroupAdmins(row.raw)"
          >
            <Icon icon="lucide:shield-user" aria-hidden="true" />
          </button>
        </RowActions>
      </template>
    </DataTable>

    <!-- Group Modal -->
    <GroupModal
      :show="showModal"
      :group="selectedGroup"
      @close="closeModal"
      @submit="handleSubmit"
    />

    <!-- Group Admins Modal -->
    <div v-if="showGroupAdminsModal" class="modal-overlay" @click.self="closeGroupAdminsModal">
      <div class="modal-container" @click.self="closeGroupAdminsModal">
        <div v-focus-trap class="modal-panel w-full sm:max-w-2xl" role="dialog" aria-modal="true" aria-labelledby="group-admins-modal-title">
          <div class="modal-form">
            <div class="modal-header">
              <div class="flex items-center justify-between gap-4">
                <div class="flex items-center gap-3 min-w-0">
                  <div
                    class="h-10 w-10 flex-shrink-0 rounded-lg flex items-center justify-center text-white shadow-sm"
                    :style="{ backgroundColor: selectedGroupForAdmins?.color || '#6B7280' }"
                  >
                    <Icon icon="lucide:shield-user" class="h-5 w-5" />
                  </div>
                  <div class="min-w-0">
                    <h3 id="group-admins-modal-title" class="modal-title truncate">{{ $t('groupAdminsModal.title') }}</h3>
                    <p class="modal-subtitle truncate">{{ selectedGroupForAdmins?.name }}</p>
                  </div>
                </div>
                <button type="button" class="btn btn-ghost p-2" :aria-label="$t('common.close')" @click="closeGroupAdminsModal">
                  <Icon icon="lucide:x" class="h-5 w-5" />
                </button>
              </div>
            </div>

            <div class="modal-content modal-body-tabbed flex flex-col gap-3">
              <p class="text-sm text-gray-500 dark:text-gray-400">
                {{ $t('groupAdminsModal.intro') }}
              </p>

              <div class="search-container">
                <Icon icon="lucide:search" class="search-icon" />
                <input
                  v-model="adminSearchQuery"
                  type="search"
                  :placeholder="$t('users.searchPlaceholder')"
                  class="form-input search-input"
                />
              </div>

              <div v-if="loadingUsers" class="flex-1 flex justify-center items-center">
                <Icon icon="lucide:loader-circle" class="h-6 w-6 animate-spin text-gray-400" />
              </div>

              <div v-else class="flex-1 min-h-0 overflow-y-auto pe-1 space-y-2">
                <p v-if="filteredUsersForAdmins.length === 0" class="py-8 text-center text-sm text-gray-500">
                  {{ $t('groupAdminsModal.noUsers') }}
                </p>
                <!-- La liste est une page de résultats : au-delà, affiner la recherche -->
                <p v-if="userSearchTotal > userSearchResults.length" class="text-xs text-[var(--muted)]">
                  {{ t('ui.table.range', { from: 1, to: userSearchResults.length, total: userSearchTotal }) }}
                </p>
                <label
                  v-for="user in filteredUsersForAdmins"
                  :key="user.id"
                  class="choice-card py-2.5"
                  :class="{ 'choice-card-active': selectedAdminIds.includes(user.id) }"
                >
                  <input v-model="selectedAdminIds" :value="user.id" type="checkbox" />
                  <span
                    class="h-8 w-8 flex-shrink-0 rounded-full flex items-center justify-center text-xs font-semibold text-white"
                    style="background-color: var(--accent)"
                  >
                    {{ getUserInitials(user) }}
                  </span>
                  <span class="min-w-0 flex-1">
                    <span class="block text-sm font-medium text-gray-900 dark:text-white truncate">{{ getUserDisplayName(user) }}</span>
                    <span class="block text-xs text-gray-500 dark:text-gray-400 truncate">@{{ user.username }} • {{ user.email }}</span>
                  </span>
                  <span
                    :class="{
                      'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200': user.role === 'admin',
                      'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200': user.role === 'editor',
                      'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200': user.role !== 'admin' && user.role !== 'editor'
                    }"
                    class="badge text-xs flex-shrink-0"
                  >
                    {{ t(user.role === 'admin' ? 'users.role_admin' : user.role === 'editor' ? 'users.role_editor' : 'users.role_user') }}
                  </span>
                </label>
              </div>
            </div>

            <div class="modal-footer-bar">
              <span class="modal-footer-hint">{{ $t('groupAdminsModal.selectedCount', selectedAdminIds.length) }}</span>
              <div class="modal-footer-actions">
                <button type="button" class="btn btn-secondary" @click="closeGroupAdminsModal">{{ $t('common.cancel') }}</button>
                <button type="button" :disabled="savingAdmins" class="btn btn-primary" @click="saveGroupAdmins">
                  <Icon v-if="savingAdmins" icon="lucide:loader-circle" class="animate-spin h-4 w-4 me-2" />
                  <Icon v-else icon="lucide:save" class="h-4 w-4 me-2" />
                  {{ $t('common.save') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

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
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
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

// Group Admins Modal
const showGroupAdminsModal = ref(false)
const selectedGroupForAdmins = ref(null)
// Résultats de recherche (une page côté serveur) et admins actuels, toujours affichés.
const userSearchResults = ref([])
const userSearchTotal = ref(0)
const currentGroupAdmins = ref([])
const USER_SEARCH_PAGE_SIZE = 50
const SEARCH_DEBOUNCE_MS = 300
const selectedAdminIds = ref([])
const loadingUsers = ref(false)
const savingAdmins = ref(false)
const adminSearchQuery = ref('')

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

const matchesAdminSearch = (user) => {
  const query = adminSearchQuery.value.trim().toLowerCase()
  if (!query) return true
  return [user.username, user.email, user.first_name, user.last_name]
    .some(value => value?.toLowerCase().includes(query))
}

// Admins actuels (filtrés localement) puis résultats du serveur, sans doublon.
const filteredUsersForAdmins = computed(() => {
  const admins = currentGroupAdmins.value.filter(matchesAdminSearch)
  const adminIds = new Set(admins.map(user => user.id))
  return [...admins, ...userSearchResults.value.filter(user => !adminIds.has(user.id))]
})

let userSearchRequest = 0
const searchUsersForAdmins = async () => {
  const request = ++userSearchRequest
  try {
    const data = await adminService.getUsers({
      search: adminSearchQuery.value.trim() || undefined,
      page_size: USER_SEARCH_PAGE_SIZE,
    })
    if (request !== userSearchRequest) return
    userSearchResults.value = data.data || []
    userSearchTotal.value = data.total || 0
  } catch (error) {
    if (request !== userSearchRequest) return
    console.error('Error searching users:', error)
    appStore.showError(t('groups.loadError'))
  }
}

let adminSearchTimer = null
watch(adminSearchQuery, () => {
  if (!showGroupAdminsModal.value) return
  clearTimeout(adminSearchTimer)
  adminSearchTimer = setTimeout(searchUsersForAdmins, SEARCH_DEBOUNCE_MS)
})
onBeforeUnmount(() => clearTimeout(adminSearchTimer))

// Methods
const getUserInitials = (user) => {
  if (user.first_name && user.last_name) {
    return `${user.first_name.charAt(0)}${user.last_name.charAt(0)}`.toUpperCase()
  }
  return user.username.charAt(0).toUpperCase()
}

const getUserDisplayName = (user) => {
  if (user.first_name && user.last_name) {
    return `${user.first_name} ${user.last_name}`
  }
  return user.username
}

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

// Group Admins Management
const manageGroupAdmins = async (group) => {
  selectedGroupForAdmins.value = group
  showGroupAdminsModal.value = true
  adminSearchQuery.value = ''

  try {
    loadingUsers.value = true

    // Admins actuels et première page d'utilisateurs, en parallèle
    const [adminsData] = await Promise.all([
      adminService.getGroupAdmins(group.id),
      searchUsersForAdmins(),
    ])
    const currentAdmins = Array.isArray(adminsData) ? adminsData : (adminsData.data || [])
    currentGroupAdmins.value = currentAdmins
    selectedAdminIds.value = currentAdmins.map(admin => admin.id)
  } catch (error) {
    console.error('Error loading group admins:', error)
    appStore.showError(t('groups.adminsLoadError'))
  } finally {
    loadingUsers.value = false
  }
}

const closeGroupAdminsModal = () => {
  showGroupAdminsModal.value = false
  selectedGroupForAdmins.value = null
  clearTimeout(adminSearchTimer)
  userSearchResults.value = []
  userSearchTotal.value = 0
  currentGroupAdmins.value = []
  selectedAdminIds.value = []
  adminSearchQuery.value = ''
}

const saveGroupAdmins = async () => {
  if (!selectedGroupForAdmins.value) return

  try {
    savingAdmins.value = true
    await adminService.assignGroupAdmins(selectedGroupForAdmins.value.id, selectedAdminIds.value)
    appStore.showSuccess(t('groups.adminsSaved'))
    closeGroupAdminsModal()
  } catch (error) {
    console.error('Error saving group admins:', error)
    appStore.showError(t('groups.saveError'))
  } finally {
    savingAdmins.value = false
  }
}

// Lifecycle
onMounted(() => {
  loadGroups()
})

// Échap ferme la modale
useEscapeClose(showGroupAdminsModal, () => closeGroupAdminsModal())
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
