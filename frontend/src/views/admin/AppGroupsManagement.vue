<template>
  <div class="content-area">
    <!-- Header -->
    <PageHeader>
      <template #title>{{ isGlobal ? $t('common.appGroups') : $t('groupAdmin.appGroupsManagement.title') }}</template>
      <template #description>{{ isGlobal ? $t('appGroups.subtitle') : $t('groupAdmin.appGroupsManagement.subtitle') }}</template>
      <template #actions>
        <button @click="openCreateModal" class="btn btn-primary" :disabled="!canCreate">
          <Icon icon="mdi:plus" class="h-4 w-4 mr-2" />
          {{ $t('appGroups.new') }}
        </button>
      </template>
    </PageHeader>

    <!-- Admin de groupe sans groupe administré : création impossible -->
    <div v-if="showNoManagedGroups" class="mb-6 p-4 bg-yellow-900/20 border border-yellow-600 rounded-lg">
      <div class="flex items-start">
        <Icon icon="mdi:information" class="h-5 w-5 text-yellow-500 mr-3 mt-0.5" />
        <div>
          <h3 class="text-yellow-500 font-medium">{{ $t('groupAdmin.appGroupsManagement.noManagedGroups') }}</h3>
          <p class="text-sm text-yellow-200 mt-1">
            {{ $t('groupAdmin.appGroupsManagement.noManagedGroupsHelp') }}
          </p>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="filters-container">
      <div class="search-container flex-1">
        <Icon icon="mdi:magnify" class="search-icon" />
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="$t('groupAdmin.appGroupsManagement.searchPlaceholder')"
          class="form-input search-input"
        />
      </div>
      <select v-model="statusFilter" class="form-select w-full sm:w-auto">
        <option value="">{{ $t('groupAdmin.appGroupsManagement.status.all') }}</option>
        <option value="active">{{ $t('groupAdmin.appGroupsManagement.status.active') }}</option>
        <option value="inactive">{{ $t('groupAdmin.appGroupsManagement.status.inactive') }}</option>
      </select>
    </div>

    <!-- Content -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <Icon icon="mdi:loading" class="h-8 w-8 animate-spin text-gray-400" />
    </div>

    <div v-else-if="filteredAppGroups.length > 0" class="grid-container grid-cols-auto">
      <div
        v-for="group in filteredAppGroups"
        :key="group.id"
        class="card hover:border-gray-600 transition-all duration-200 fade-in"
      >
        <div class="card-header">
          <div class="flex items-center space-x-3">
            <Icon
              :icon="group.icon || 'mdi:folder'"
              class="h-8 w-8 flex-shrink-0"
              :style="{ color: group.color || '#10b981' }"
            />
            <div>
              <h3 class="card-title">{{ group.name }}</h3>
              <p class="text-sm text-gray-400">{{ group.applications?.length || 0 }} {{ $t('groupAdmin.appGroupsManagement.applications') }}</p>
            </div>
          </div>

          <div class="flex space-x-2">
            <button @click="editGroup(group)" class="btn-ghost" :title="$t('groupAdmin.appGroupsManagement.actions.edit')">
              <Icon icon="mdi:pencil" class="h-4 w-4" />
            </button>
            <button @click="confirmDelete(group)" class="btn-ghost text-red-400 hover:text-red-300" :title="$t('groupAdmin.appGroupsManagement.actions.delete')">
              <Icon icon="mdi:delete" class="h-4 w-4" />
            </button>
          </div>
        </div>

        <div class="card-content">
          <p v-if="group.description" class="text-sm mb-4">{{ group.description }}</p>

          <div class="flex items-center justify-between text-sm mb-3">
            <StatusBadge :tone="group.is_active ? 'ok' : 'neutral'">{{ group.is_active ? $t('common.active') : $t('common.inactive') }}</StatusBadge>
            <span class="text-gray-500">{{ $t('groupAdmin.appGroupsManagement.order') }} {{ group.order || 0 }}</span>
          </div>

          <div v-if="group.is_private" class="flex items-center space-x-2 text-xs">
            <span class="badge badge-info flex items-center space-x-1">
              <Icon icon="mdi:lock" class="h-3 w-3" />
              <span>{{ $t('adminShared.appGroups.private') }}</span>
            </span>
            <span v-if="group.owner_group" class="text-gray-400">
              {{ $t('groupAdmin.appGroupsManagement.ownerGroup') }} <span class="text-gray-300">{{ group.owner_group.name }}</span>
            </span>
          </div>
          <div v-else class="text-xs">
            <span class="badge badge-default flex items-center space-x-1 w-fit">
              <Icon icon="mdi:earth" class="h-3 w-3" />
              <span>{{ $t('adminShared.appGroups.public') }}</span>
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty States -->
    <div v-else-if="appGroups.length === 0" class="empty-state">
      <Icon icon="mdi:folder-multiple" class="empty-state-icon" />
      <h3 class="empty-state-title">{{ $t('groupAdmin.appGroupsManagement.empty.title') }}</h3>
      <p class="empty-state-description">{{ $t('groupAdmin.appGroupsManagement.empty.description') }}</p>
      <button @click="openCreateModal" class="btn btn-primary" :disabled="!canCreate">
        <Icon icon="mdi:plus" class="h-4 w-4 mr-2" />
        {{ $t('groupAdmin.appGroupsManagement.empty.createButton') }}
      </button>
    </div>

    <div v-else class="empty-state">
      <Icon icon="mdi:magnify" class="empty-state-icon" />
      <h3 class="empty-state-title">{{ $t('groupAdmin.appGroupsManagement.noResults.title') }}</h3>
      <p class="empty-state-description">{{ $t('groupAdmin.appGroupsManagement.noResults.description') }}</p>
    </div>

    <!-- App Group Modal -->
    <AppGroupModal
      :show="showModal"
      :app-group="selectedAppGroup"
      :managed-groups="managedGroups"
      @close="closeModal"
      @submit="handleSubmit"
    />

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay">
      <div class="modal-container">
        <div class="modal-panel sm:max-w-lg sm:w-full">
          <div class="modal-header">
            <div class="flex items-start space-x-3">
              <div class="flex-shrink-0">
                <Icon icon="mdi:alert" class="h-6 w-6 text-red-500" />
              </div>
              <div>
                <h3 class="modal-title">{{ $t('groupAdmin.appGroupsManagement.deleteConfirm.title') }}</h3>
                <p class="modal-subtitle">
                  {{ $t('groupAdmin.appGroupsManagement.deleteConfirm.message', { name: appGroupToDelete?.name }) }}
                </p>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeDeleteModal" class="btn btn-secondary w-full sm:w-auto">
              {{ $t('common.cancel') }}
            </button>
            <button
              @click="deleteGroup"
              :disabled="deleteLoading"
              class="btn btn-danger w-full sm:w-auto"
            >
              <Icon v-if="deleteLoading" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
              {{ $t('common.delete') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import { adminService, groupAdminService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { useManagementScope } from '@/composables/useManagementScope'
import { GROUP_ADMIN } from '@/config/access'
import AppGroupModal from '@/components/admin/AppGroupModal.vue'

const { t } = useI18n()
const appStore = useAppStore()

// Admin global : tous les groupes de ressources (API admin) ;
// admin de groupe : groupes de ressources privés de ses groupes (API group-admin)
const { isGlobal } = useManagementScope(GROUP_ADMIN)
const service = computed(() => (isGlobal.value ? adminService : groupAdminService))

// State
const appGroups = ref([])
const managedGroups = ref([])
const managedGroupsLoaded = ref(false)
const loading = ref(false)
const showModal = ref(false)
const selectedAppGroup = ref(null)
const showDeleteModal = ref(false)
const appGroupToDelete = ref(null)
const deleteLoading = ref(false)
const searchQuery = ref('')
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

const filteredAppGroups = computed(() => {
  let filtered = appGroups.value

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(group =>
      group.name.toLowerCase().includes(query) ||
      group.description?.toLowerCase().includes(query)
    )
  }

  if (statusFilter.value === 'active') {
    filtered = filtered.filter(group => group.is_active)
  } else if (statusFilter.value === 'inactive') {
    filtered = filtered.filter(group => !group.is_active)
  }

  return [...filtered].sort((a, b) => (a.order || 0) - (b.order || 0))
})

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

const confirmDelete = (group) => {
  appGroupToDelete.value = group
  showDeleteModal.value = true
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  appGroupToDelete.value = null
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

const deleteGroup = async () => {
  if (!appGroupToDelete.value?.id) return

  try {
    deleteLoading.value = true
    await service.value.deleteAppGroup(appGroupToDelete.value.id)
    appStore.showSuccess(t('adminShared.appGroups.deleted'))
    closeDeleteModal()
    await loadAppGroups()
  } catch (error) {
    console.error('Error deleting app group:', error)
    appStore.showError(t('groupAdmin.appGroupsManagement.deleteError'))
  } finally {
    deleteLoading.value = false
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
