<template>
  <div class="content-area">
    <!-- En-tête -->
    <PageHeader>
      <template #title>{{ $t('applications.title') }}</template>
      <template #description>
        {{ isGlobal ? $t('applications.subtitle') : $t('groupAdmin.applicationsManagement.subtitle') }}
      </template>
      <template #actions>
        <button @click="openCreateModal" class="btn btn-primary" :disabled="hasNoAppGroups">
          <Icon icon="mdi:plus" class="h-4 w-4 mr-2" />
          {{ $t('applications.new') }}
        </button>
      </template>
    </PageHeader>

    <!-- Aucun groupe de ressources : impossible d'ajouter une ressource -->
    <div v-if="hasNoAppGroups" class="mb-6 p-4 bg-yellow-900/20 border border-yellow-600 rounded-lg">
      <div class="flex items-start">
        <Icon icon="mdi:information" class="h-5 w-5 text-yellow-500 mr-3 mt-0.5" />
        <div>
          <h3 class="text-yellow-500 font-medium">{{ $t('groupAdmin.applicationsManagement.noAppGroups') }}</h3>
          <p class="text-sm text-yellow-200 mt-1">
            {{ $t('groupAdmin.applicationsManagement.noAppGroupsHelp') }}
          </p>
        </div>
      </div>
    </div>

    <!-- Filtres -->
    <div class="filters-container">
      <div class="search-container flex-1">
        <Icon icon="mdi:magnify" class="search-icon" />
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="$t('applications.searchPlaceholder')"
          class="form-input search-input"
        />
      </div>
      <select v-model="groupFilter" class="form-select w-full sm:w-auto">
        <option value="">{{ $t('applications.filter_all_groups') }}</option>
        <option v-for="group in appGroups" :key="group.id" :value="group.id">
          {{ group.name }}
        </option>
      </select>
      <select v-model="statusFilter" class="form-select w-full sm:w-auto">
        <option value="">{{ $t('common.status_all') }}</option>
        <option value="active">{{ $t('common.status_active') }}</option>
        <option value="inactive">{{ $t('common.status_inactive') }}</option>
      </select>
    </div>

    <!-- Contenu -->
    <div v-if="appStore.loading" class="flex items-center justify-center py-12">
      <Icon icon="mdi:loading" class="h-8 w-8 animate-spin text-gray-400" />
    </div>

    <div v-else-if="filteredApplications.length > 0" class="grid-container grid-cols-auto">
      <div
        v-for="app in filteredApplications"
        :key="app.id"
        class="card hover:border-gray-600 transition-all duration-200 fade-in"
      >
        <div class="card-header">
          <div class="flex items-center gap-3">
            <Icon
              :icon="app.icon || 'mdi:application'"
              class="h-8 w-8 flex-shrink-0"
              :style="{ color: app.color || '#6366f1' }"
            />
            <div class="flex-1">
              <h3 class="card-title">{{ app.name }}</h3>
              <p class="text-sm text-gray-400">{{ getAppGroupName(app.app_group_id) }}</p>
            </div>
          </div>

          <div class="flex gap-2">
            <button @click="editApplication(app)" class="btn-ghost" :title="$t('common.edit')">
              <Icon icon="mdi:pencil" class="h-4 w-4" />
            </button>
            <button @click="confirmDelete(app)" class="btn-ghost text-red-400 hover:text-red-300" :title="$t('common.delete')">
              <Icon icon="mdi:delete" class="h-4 w-4" />
            </button>
          </div>
        </div>

        <div class="card-content">
          <p v-if="app.description" class="text-sm mb-3">{{ app.description }}</p>

          <div class="mb-3">
            <a
              :href="app.url"
              :target="app.open_in_new_tab ? '_blank' : '_self'"
              class="text-sm text-green-400 hover:text-green-300 flex items-center space-x-1 group"
              :title="app.url"
            >
              <span class="truncate">{{ app.url }}</span>
              <Icon v-if="app.open_in_new_tab" icon="mdi:open-in-new" class="h-3 w-3 flex-shrink-0" />
            </a>
          </div>

          <div class="flex items-center justify-between text-sm">
            <StatusBadge :tone="app.is_active ? 'ok' : 'neutral'">{{ app.is_active ? $t('common.active') : $t('common.inactive') }}</StatusBadge>
            <span class="text-gray-500">{{ $t('applications.order') }}: {{ app.order || 0 }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- États vides -->
    <div v-else-if="applications.length === 0" class="empty-state">
      <Icon icon="mdi:application" class="empty-state-icon" />
      <h3 class="empty-state-title">{{ $t('applications.emptyTitle') }}</h3>
      <p class="empty-state-description">{{ $t('applications.emptyText') }}</p>
      <button @click="openCreateModal" class="btn btn-primary" :disabled="hasNoAppGroups">
        <Icon icon="mdi:plus" class="h-4 w-4 mr-2" />
        {{ $t('applications.add') }}
      </button>
    </div>

    <div v-else class="empty-state">
      <Icon icon="mdi:magnify" class="empty-state-icon" />
      <h3 class="empty-state-title">{{ $t('applications.noResultsTitle') }}</h3>
      <p class="empty-state-description">{{ $t('applications.noResultsText') }}</p>
    </div>

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

    <!-- Confirmation de suppression -->
    <div v-if="showDeleteModal" class="modal-overlay">
      <div class="modal-container">
        <div class="modal-panel sm:max-w-lg sm:w-full">
          <div class="modal-header">
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon icon="mdi:alert" class="h-6 w-6 text-red-500" />
              </div>
              <div>
                <h3 class="modal-title">{{ $t('applications.confirmDeleteTitle') }}</h3>
                <p class="modal-subtitle">
                  {{ $t('applications.confirmDeleteText', { name: applicationToDelete?.name }) }}
                </p>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeDeleteModal" class="btn btn-secondary w-full sm:w-auto">
              {{ $t('common.cancel') }}
            </button>
            <button
              @click="deleteApplication"
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
import StatusBadge from '@/components/ui/StatusBadge.vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { adminService, groupAdminService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { useManagementScope } from '@/composables/useManagementScope'
import { GROUP_ADMIN } from '@/config/access'
import ApplicationModal from '@/components/admin/ApplicationModal.vue'

const { t } = useI18n()
const appStore = useAppStore()

// Admin global : toutes les ressources (API admin).
// Admin de groupe : ressources des groupes de ressources privés de ses groupes (API group-admin).
const { isGlobal } = useManagementScope(GROUP_ADMIN)
const service = computed(() => (isGlobal.value ? adminService : groupAdminService))

// État
const applications = ref([])
const appGroups = ref([])
const appGroupsLoaded = ref(false)
const users = ref([])
const showModal = ref(false)
const selectedApplication = ref(null)
const showDeleteModal = ref(false)
const applicationToDelete = ref(null)
const deleteLoading = ref(false)
const searchQuery = ref('')
const groupFilter = ref('')
const statusFilter = ref('')

/** Les services renvoient soit un tableau, soit une enveloppe { data } */
const toArray = (data) => (Array.isArray(data) ? data : (data?.data || []))

// Calculs
const hasNoAppGroups = computed(() => appGroupsLoaded.value && appGroups.value.length === 0)

const filteredApplications = computed(() => {
  const query = searchQuery.value.toLowerCase()
  const groupId = groupFilter.value ? parseInt(groupFilter.value) : null

  return applications.value
    .filter(app => !query ||
      app.name.toLowerCase().includes(query) ||
      app.description?.toLowerCase().includes(query) ||
      app.url.toLowerCase().includes(query))
    .filter(app => groupId === null || app.app_group_id === groupId)
    .filter(app => {
      if (statusFilter.value === 'active') return app.is_active
      if (statusFilter.value === 'inactive') return !app.is_active
      return true
    })
    .sort((a, b) => (a.order || 0) - (b.order || 0))
})

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
  if (!isGlobal.value) return
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

const confirmDelete = (application) => {
  applicationToDelete.value = application
  showDeleteModal.value = true
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  applicationToDelete.value = null
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

const deleteApplication = async () => {
  if (!applicationToDelete.value?.id) return

  try {
    deleteLoading.value = true
    await service.value.deleteApplication(applicationToDelete.value.id)
    appStore.showSuccess(t('adminShared.applications.deleted'))
    closeDeleteModal()
    await loadApplications()
  } catch (error) {
    console.error('Error deleting application:', error)
    appStore.showError(t('groupAdmin.applicationsManagement.deleteError'))
  } finally {
    deleteLoading.value = false
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
