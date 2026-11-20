<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ $t('users.title') }}</template>
      <template #description>{{ $t('users.subtitle') }}</template>
      <template #actions>
        <BaseButton variant="primary" icon="mdi:plus" @click="openCreateModal">
          {{ $t('users.new') }}
        </BaseButton>
      </template>
    </PageHeader>

    <!-- Onglets : utilisateurs actifs / corbeille -->
    <div class="um-tabs">
      <SegmentedControl v-model="viewMode" :options="viewOptions" :label="$t('users.viewLabel')" />
    </div>

    <EmptyState
      v-if="!showDeleted && !appStore.loading && users.length === 0"
      icon="mdi:account-multiple-outline"
      :title="$t('users.emptyTitle')"
      :description="$t('users.emptyText')"
    >
      <BaseButton variant="primary" size="sm" icon="mdi:plus" @click="openCreateModal">
        {{ $t('users.create') }}
      </BaseButton>
    </EmptyState>

    <DataTable
      v-else
      :key="viewMode"
      :columns="columns"
      :rows="rows"
      :loading="!showDeleted && appStore.loading"
      :search-keys="['fullName', 'username', 'email']"
      :search-placeholder="$t('users.searchPlaceholder')"
      :empty-title="$t('users.noResultsTitle')"
      :page-size="25"
    >
      <template v-if="!showDeleted" #toolbar>
        <label>
          <span class="sr-only">{{ $t('ui.table.columns.role') }}</span>
          <select v-model="roleFilter" class="ui-select">
            <option value="">{{ $t('users.filter_all_roles') }}</option>
            <option value="admin">{{ $t('users.filter_admins') }}</option>
            <option value="editor">{{ $t('users.filter_editors') }}</option>
            <option value="user">{{ $t('users.filter_users') }}</option>
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

      <template #cell-fullName="{ row }">
        <span class="um-user">
          <span class="um-user__avatar" aria-hidden="true">{{ row.initials }}</span>
          <span class="um-user__text">
            <span class="um-user__name">{{ row.fullName }}</span>
            <span class="um-user__handle" dir="ltr">@{{ row.username }}</span>
          </span>
        </span>
      </template>

      <template #cell-email="{ row }">
        <span class="um-email" dir="ltr">{{ row.email }}</span>
      </template>

      <template #cell-role="{ row }">
        <StatusBadge :tone="roleTone(row.raw.role)">{{ row.role }}</StatusBadge>
      </template>

      <template #cell-groups="{ row }">
        <span v-if="row.raw.groups && row.raw.groups.length > 0" class="um-chips">
          <span
            v-for="group in row.raw.groups"
            :key="group.id"
            class="um-chip"
            :style="group.color ? { borderColor: group.color, color: group.color } : undefined"
          >{{ group.name }}</span>
        </span>
        <span v-else class="um-mono">{{ $t('users.no_group') }}</span>
      </template>

      <template #cell-lastLogin="{ row }">
        <span v-if="row.raw.last_login" class="um-mono">{{ formatLastConnection(row.raw.last_login) }}</span>
        <span v-else class="um-mono um-mono--dim">{{ $t('users.never_connected') }}</span>
      </template>

      <template #cell-status="{ row }">
        <StatusBadge :tone="row.raw.is_active ? 'ok' : 'bad'">{{ row.status }}</StatusBadge>
      </template>

      <template #cell-actions="{ row }">
        <!-- Utilisateurs supprimés : restaurer ou supprimer définitivement -->
        <div v-if="showDeleted" class="row-actions">
          <button
            type="button"
            class="row-actions__btn"
            :aria-label="`Restaurer l'utilisateur : ${row.username}`"
            title="Restaurer l'utilisateur"
            @click="restoreUser(row.raw)"
          >
            <Icon icon="mdi:restore" aria-hidden="true" />
          </button>
          <button
            type="button"
            class="row-actions__btn row-actions__btn--danger"
            :aria-label="`Supprimer définitivement : ${row.username}`"
            title="Supprimer définitivement (irréversible)"
            @click="confirmPermanentDelete(row.raw)"
          >
            <Icon icon="mdi:delete-forever-outline" aria-hidden="true" />
          </button>
        </div>
        <!-- Utilisateurs actifs -->
        <RowActions v-else :label="row.username" @edit="editUser(row.raw)" @delete="confirmDelete(row.raw)">
          <button
            type="button"
            class="row-actions__btn"
            :aria-label="`${row.raw.is_active ? $t('users.deactivate') : $t('users.activate')} : ${row.username}`"
            :title="row.raw.is_active ? $t('users.deactivate') : $t('users.activate')"
            @click="toggleUserStatus(row.raw)"
          >
            <Icon :icon="row.raw.is_active ? 'mdi:account-off-outline' : 'mdi:account-check-outline'" aria-hidden="true" />
          </button>
        </RowActions>
      </template>
    </DataTable>

    <!-- User Modal -->
    <UserModal
      :show="showModal"
      :user="selectedUser"
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
import SegmentedControl from '@/components/ui/SegmentedControl.vue'
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { adminService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { useConfirm } from '@/composables/useConfirm'
import UserModal from '@/components/admin/UserModal.vue'

const { t } = useI18n()
const { confirm: askConfirm } = useConfirm()
const appStore = useAppStore()

// State
const users = ref([])
const deletedUsers = ref([])
const viewMode = ref('active')
const showDeleted = computed(() => viewMode.value === 'deleted')
const showModal = ref(false)
const selectedUser = ref(null)
const roleFilter = ref('')
const statusFilter = ref('')

// Computed
const viewOptions = computed(() => [
  { value: 'active', label: t('users.viewActive'), icon: 'mdi:account-check-outline' },
  {
    value: 'deleted',
    label: t('users.viewTrash'),
    icon: 'mdi:trash-can-outline',
    count: deletedUsers.value.length > 0 ? deletedUsers.value.length : undefined,
  },
])

const columns = computed(() => [
  { key: 'fullName', label: t('ui.table.columns.name'), sortable: true },
  { key: 'email', label: t('ui.table.columns.email'), sortable: true, hideSm: true },
  { key: 'role', label: t('ui.table.columns.role'), sortable: true, hideSm: true },
  { key: 'groups', label: t('ui.table.columns.groups'), hideSm: true },
  { key: 'lastLogin', label: t('users.last_connection'), sortable: true, hideSm: true },
  { key: 'status', label: t('ui.table.columns.status'), sortable: true },
  { key: 'actions', label: '', align: 'end', width: showDeleted.value ? '90px' : '120px' },
])

const roleLabel = (role) =>
  role === 'admin' ? t('users.role_admin') : role === 'editor' ? t('users.role_editor') : t('users.role_user')

const roleTone = (role) => (role === 'admin' ? 'info' : role === 'editor' ? 'warn' : 'neutral')

const toRow = (user) => ({
  id: user.id,
  fullName: getUserDisplayName(user),
  initials: getUserInitials(user),
  username: user.username,
  email: user.email || '',
  role: roleLabel(user.role),
  lastLogin: user.last_login ? new Date(user.last_login).getTime() : 0,
  status: user.is_active ? t('common.active') : t('common.inactive'),
  raw: user,
})

// Lignes du tableau : filtres rôle et statut (actifs seulement) ; la recherche et le tri sont dans DataTable.
const rows = computed(() => {
  if (showDeleted.value) return deletedUsers.value.map(toRow)
  return users.value
    .filter(user => !roleFilter.value || user.role === roleFilter.value)
    .filter(user => {
      if (statusFilter.value === 'active') return user.is_active
      if (statusFilter.value === 'inactive') return !user.is_active
      return true
    })
    .map(toRow)
})

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

const formatLastConnection = (lastLogin) => {
  if (!lastLogin) return ''
  
  const now = new Date()
  const loginDate = new Date(lastLogin)
  const diffMs = now - loginDate
  const diffMins = Math.floor(diffMs / (1000 * 60))
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  
  if (diffMins < 1) {
    return "À l'instant"
  } else if (diffMins < 60) {
    return `il y a ${diffMins} minute${diffMins > 1 ? 's' : ''}`
  } else if (diffHours < 24) {
    return `il y a ${diffHours} heure${diffHours > 1 ? 's' : ''}`
  } else if (diffDays < 7) {
    return `il y a ${diffDays} jour${diffDays > 1 ? 's' : ''}`
  } else {
    return loginDate.toLocaleDateString('fr-FR', {
      day: 'numeric',
      month: 'short',
      year: loginDate.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
    })
  }
}

const loadUsers = async () => {
  try {
    appStore.setLoading(true)
    const data = await adminService.getUsers()
    users.value = Array.isArray(data) ? data : []
  } catch (error) {
    console.error('Erreur lors du chargement des utilisateurs:', error)
    users.value = []
    appStore.showError('Erreur lors du chargement des utilisateurs')
  } finally {
    appStore.setLoading(false)
  }
}

const loadDeletedUsers = async () => {
  try {
    const data = await adminService.getDeletedUsers()
    deletedUsers.value = Array.isArray(data.users) ? data.users : []
  } catch (error) {
    console.error('Erreur lors du chargement des utilisateurs supprimés:', error)
    deletedUsers.value = []
  }
}

const openCreateModal = () => {
  selectedUser.value = null
  showModal.value = true
}

const editUser = (user) => {
  selectedUser.value = user
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  selectedUser.value = null
}

const confirmDelete = async (user) => {
  const confirmed = await askConfirm({
    title: t('users.deleteTitle'),
    message: t('users.deleteMessage', { name: user.username }),
    confirmLabel: t('common.delete'),
    tone: 'danger',
  })
  if (!confirmed || !user?.id) return

  try {
    await adminService.deleteUser(user.id)
    appStore.showSuccess('Utilisateur supprimé avec succès')
    await loadUsers()
    await loadDeletedUsers()
  } catch (error) {
    console.error('Erreur lors de la suppression:', error)
    appStore.showError('Erreur lors de la suppression de l\'utilisateur')
  }
}

const handleSubmit = async (formData) => {
  try {
    appStore.setLoading(true)
    
    if (formData.id) {
      // Update existing user - only send relevant fields
      const { id, ...updateData } = formData
      await adminService.updateUser(id, updateData)
      appStore.showSuccess('Utilisateur modifié avec succès')
    } else {
      // Create new user - only send relevant fields  
      const { id, ...createData } = formData
      await adminService.createUser(createData)
      appStore.showSuccess('Utilisateur créé avec succès')
    }
    
    closeModal()
    await loadUsers()
  } catch (error) {
    console.error('Erreur lors de la sauvegarde:', error)
    appStore.showError('Erreur lors de la sauvegarde de l\'utilisateur')
  } finally {
    appStore.setLoading(false)
  }
}

const toggleUserStatus = async (user) => {
  try {
    // Only send the fields that need to be updated
    const updatedData = {
      is_active: !user.is_active
    }
    
    await adminService.updateUser(user.id, updatedData)
    appStore.showSuccess(`Utilisateur ${!user.is_active ? 'activé' : 'désactivé'} avec succès`)
    await loadUsers()
  } catch (error) {
    console.error('Erreur lors de la modification du statut:', error)
    appStore.showError('Erreur lors de la modification du statut')
  }
}

const restoreUser = async (user) => {
  try {
    await adminService.restoreUser(user.id)
    appStore.showSuccess(`Utilisateur ${user.username} restauré avec succès`)
    await loadUsers()
    await loadDeletedUsers()
  } catch (error) {
    console.error('Erreur lors de la restauration:', error)
    appStore.showError('Erreur lors de la restauration de l\'utilisateur')
  }
}

const confirmPermanentDelete = async (user) => {
  if (
    await askConfirm({
      title: t('users.permanentDeleteTitle'),
      message: t('users.permanentDeleteMessage', { username: user.username }),
      tone: 'danger'
    })
  ) {
    try {
      await adminService.permanentlyDeleteUser(user.id)
      appStore.showSuccess('Utilisateur supprimé définitivement')
      await loadDeletedUsers()
    } catch (error) {
      console.error('Erreur lors de la suppression définitive:', error)
      appStore.showError('Erreur lors de la suppression définitive')
    }
  }
}

// Lifecycle
onMounted(() => {
  loadUsers()
  loadDeletedUsers()
})
</script>
<style scoped>
.um-tabs { margin-block-end: 14px; }

.um-user { display: flex; align-items: center; gap: 10px; min-width: 0; }

.um-user__avatar {
  flex: none;
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid var(--line);
  border-radius: 3px;
  background: var(--sunken);
  font-family: var(--f-num);
  font-size: 11px;
  letter-spacing: 0.04em;
  color: var(--acc-ink);
}

.um-user__text { display: flex; flex-direction: column; min-width: 0; }
.um-user__name { font-weight: 600; color: var(--text); }
.um-user__handle { font-size: 12px; color: var(--muted); }

.um-email { font-size: 12.5px; color: var(--text); }

.um-chips { display: flex; flex-wrap: wrap; gap: 4px; }

.um-chip {
  padding: 1px 7px;
  border: 1px solid var(--line);
  border-radius: 3px;
  font-size: 11.5px;
  color: var(--muted);
}

.um-mono {
  font-family: var(--f-num);
  font-size: 11.5px;
  letter-spacing: 0.04em;
  color: var(--muted);
  white-space: nowrap;
}

.um-mono--dim { font-style: italic; opacity: 0.8; }
</style>
