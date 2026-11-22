<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ t('roles.title') }}</template>
      <template #description>{{ t('roles.subtitle') }}</template>
      <template #actions>
        <BaseButton variant="primary" icon="lucide:plus" @click="openCreate">{{ t('roles.create') }}</BaseButton>
      </template>
    </PageHeader>

    <!-- Le rôle administrateur n'est pas une ligne modifiable : il a tout, partout -->
    <section class="rm-admin" :aria-label="t('roles.adminRow.name')">
      <Icon icon="lucide:shield-check" class="rm-admin__icon" aria-hidden="true" />
      <div class="rm-admin__text">
        <h2 class="rm-admin__name">{{ t('roles.adminRow.name') }}</h2>
        <p class="rm-admin__help">{{ t('roles.adminRow.text') }}</p>
      </div>
      <span class="tag tag--acc">{{ t('roles.scope.all') }}</span>
    </section>

    <EmptyState
      v-if="!loading && roles.length === 0"
      icon="lucide:shield-plus"
      :title="t('roles.empty')"
      :description="t('roles.emptyHelp')"
    >
      <BaseButton variant="primary" size="sm" icon="lucide:plus" @click="openCreate">{{ t('roles.create') }}</BaseButton>
    </EmptyState>

    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :loading="loading"
      :search-keys="['name', 'description']"
      :page-size="25"
    >
      <template #cell-name="{ row }">
        <span class="rm-name">{{ row.name }}</span>
        <span v-if="row.description" class="rm-desc">{{ row.description }}</span>
      </template>

      <template #cell-permissions="{ row }">
        <span v-if="row.raw.permissions.length === 0" class="rm-muted">{{ t('roles.noPermission') }}</span>
        <span v-else class="rm-tags">
          <span v-for="p in row.raw.permissions" :key="p.permission" class="tag">
            {{ permissionLabel(p.permission) }} · {{ t(`roles.scope.${p.scope}`) }}
          </span>
        </span>
      </template>

      <template #cell-members="{ row }">
        <span class="rm-num">{{ row.members }}</span>
      </template>

      <template #cell-actions="{ row }">
        <RowActions :label="row.name" @edit="openEdit(row.raw)" @delete="removeRole(row.raw)">
          <button
            type="button"
            class="row-actions__btn"
            :aria-label="`${t('roles.members.manage')} : ${row.name}`"
            :title="t('roles.members.manage')"
            @click="openMembers(row.raw)"
          >
            <Icon icon="lucide:user-plus" aria-hidden="true" />
          </button>
        </RowActions>
      </template>
    </DataTable>

    <!-- Création / modification d'un rôle -->
    <Teleport to="body">
      <div v-if="showForm" class="rm-overlay" @click.self="closeForm">
        <div v-focus-trap class="rm-modal" role="dialog" aria-modal="true" aria-labelledby="role-form-title">
          <div class="rm-modal__head">
            <h2 id="role-form-title" class="rm-modal__title">{{ editingId ? t('common.edit') : t('roles.create') }}</h2>
            <button type="button" class="row-actions__btn" :aria-label="t('common.close')" @click="closeForm">
              <Icon icon="lucide:x" aria-hidden="true" />
            </button>
          </div>

          <form class="rm-form" @submit.prevent="saveRole">
            <div>
              <label for="role-name" class="form-label form-label-required">{{ t('ui.table.columns.name') }}</label>
              <input id="role-name" v-model="form.name" class="form-input" maxlength="80" required />
            </div>
            <div>
              <label for="role-description" class="form-label">{{ t('common.description') }}</label>
              <textarea id="role-description" v-model="form.description" class="form-input rm-textarea" maxlength="300" rows="2" />
            </div>

            <fieldset class="rm-perms">
              <legend class="form-label">{{ t('roles.form.permissions') }}</legend>
              <p class="rm-muted">{{ t('roles.form.permissionsHelp') }}</p>

              <div v-for="def in catalog" :key="def.key" class="rm-perm" :class="{ 'is-on': form.scopes[def.key] }">
                <label class="rm-perm__check">
                  <input type="checkbox" :checked="!!form.scopes[def.key]" @change="togglePermission(def, $event.target.checked)" />
                  <span>
                    <span class="rm-perm__label">{{ permissionLabel(def.key) }}</span>
                    <span class="rm-perm__help">{{ permissionHelp(def.key) }}</span>
                  </span>
                </label>

                <div v-if="form.scopes[def.key]" class="rm-scope" role="radiogroup" :aria-label="t('roles.form.scope')">
                  <label v-for="scope in def.scopes" :key="scope" class="rm-scope__opt">
                    <input v-model="form.scopes[def.key]" type="radio" :name="`scope-${def.key}`" :value="scope" />
                    <span>{{ t(`roles.scope.${scope}`) }}</span>
                  </label>
                  <p v-if="form.scopes[def.key] === 'groups'" class="rm-scope__help">{{ t('roles.scope.groupsHelp') }}</p>
                </div>
              </div>
            </fieldset>

            <div class="rm-form__actions">
              <BaseButton variant="secondary" @click="closeForm">{{ t('common.cancel') }}</BaseButton>
              <BaseButton variant="primary" type="submit" :loading="saving">{{ t('common.save') }}</BaseButton>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Membres d'un rôle -->
    <Teleport to="body">
      <div v-if="membersRole" class="rm-overlay" @click.self="closeMembers">
        <div v-focus-trap class="rm-modal" role="dialog" aria-modal="true" aria-labelledby="role-members-title">
          <div class="rm-modal__head">
            <div class="min-w-0">
              <h2 id="role-members-title" class="rm-modal__title">{{ t('roles.members.title') }}</h2>
              <p class="rm-muted truncate">{{ membersRole.name }}</p>
            </div>
            <button type="button" class="row-actions__btn" :aria-label="t('common.close')" @click="closeMembers">
              <Icon icon="lucide:x" aria-hidden="true" />
            </button>
          </div>

          <div class="rm-members">
            <p class="rm-muted">{{ t('roles.members.intro') }}</p>
            <input
              v-model="memberQuery"
              type="search"
              class="form-input"
              :aria-label="t('users.searchPlaceholder')"
              :placeholder="t('users.searchPlaceholder')"
            />

            <p v-if="loadingMembers" class="rm-muted" role="status">{{ t('roles.members.loading') }}</p>
            <ul v-else class="rm-members__list">
              <li v-if="memberCandidates.length === 0" class="rm-muted">{{ t('roles.members.noUsers') }}</li>
              <li v-for="user in memberCandidates" :key="user.id">
                <label class="choice-card py-2" :class="{ 'choice-card-active': selectedMemberIds.includes(user.id) }">
                  <input v-model="selectedMemberIds" type="checkbox" :value="user.id" />
                  <span class="min-w-0 flex-1">
                    <span class="rm-name block truncate">{{ displayName(user) }}</span>
                    <span class="rm-desc block truncate" dir="ltr">@{{ user.username }} · {{ user.email }}</span>
                  </span>
                </label>
              </li>
            </ul>
            <p v-if="searchTotal > searchResults.length" class="rm-muted">
              {{ t('ui.table.range', { from: 1, to: searchResults.length, total: searchTotal }) }}
            </p>
          </div>

          <div class="rm-modal__foot">
            <span class="rm-muted">{{ t('roles.members.selected', selectedMemberIds.length) }}</span>
            <div class="rm-form__actions">
              <BaseButton variant="secondary" @click="closeMembers">{{ t('common.cancel') }}</BaseButton>
              <BaseButton variant="primary" :loading="saving" @click="saveMembers">{{ t('common.save') }}</BaseButton>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
// Rôles créés par l'administrateur : permissions du catalogue du serveur (chacune sur tout
// le site ou sur les groupes de la personne) et membres. L'administrateur, lui, a tout.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { adminService } from '@/services/api'
import { useConfirm } from '@/composables/useConfirm'
import { useEscapeClose } from '@/composables/useEscapeClose'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import RowActions from '@/components/ui/RowActions.vue'

const { t, te } = useI18n()
const { confirm: askConfirm } = useConfirm()

const USER_SEARCH_PAGE_SIZE = 50
const SEARCH_DEBOUNCE_MS = 300

const loading = ref(false)
const saving = ref(false)
const roles = ref([])
const catalog = ref([])

const showForm = ref(false)
const editingId = ref(null)
const form = ref({ name: '', description: '', scopes: {} })

const membersRole = ref(null)
const currentMembers = ref([])
const searchResults = ref([])
const searchTotal = ref(0)
const selectedMemberIds = ref([])
const memberQuery = ref('')
const loadingMembers = ref(false)

const toast = (kind, message) => window.$toast?.[kind]?.(message)

// Libellés : clé i18n dérivée de la permission (« resources.manage » → resources_manage)
const permissionKey = (permission) => permission.replace(/\./g, '_')
const permissionLabel = (permission) => {
  const key = `roles.permissions.${permissionKey(permission)}.label`
  return te(key) ? t(key) : permission
}
const permissionHelp = (permission) => {
  const key = `roles.permissions.${permissionKey(permission)}.help`
  return te(key) ? t(key) : ''
}

const columns = computed(() => [
  { key: 'name', label: t('ui.table.columns.name'), sortable: true },
  { key: 'permissions', label: t('roles.form.permissions'), hideSm: true },
  { key: 'members', label: t('ui.table.columns.members'), sortable: true, align: 'end', width: '100px' },
  { key: 'actions', label: '', align: 'end', width: '120px' },
])

const rows = computed(() =>
  roles.value.map(role => ({
    id: role.id,
    name: role.name,
    description: role.description || '',
    members: role.user_count || 0,
    raw: role,
  })),
)

const loadRoles = async () => {
  loading.value = true
  try {
    const [list, defs] = await Promise.all([adminService.getAccessRoles(), adminService.getPermissionCatalog()])
    roles.value = Array.isArray(list) ? list : []
    catalog.value = Array.isArray(defs) ? defs : []
  } catch (error) {
    console.error('Error loading roles:', error)
    toast('error', t('roles.errors.load'))
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  editingId.value = null
  form.value = { name: '', description: '', scopes: {} }
  showForm.value = true
}

const openEdit = (role) => {
  editingId.value = role.id
  form.value = {
    name: role.name,
    description: role.description || '',
    scopes: Object.fromEntries(role.permissions.map(p => [p.permission, p.scope])),
  }
  showForm.value = true
}

const closeForm = () => { showForm.value = false }

// Cocher une permission : portée par défaut « ses groupes » quand elle est proposée
const togglePermission = (def, checked) => {
  const scopes = { ...form.value.scopes }
  if (checked) {
    scopes[def.key] = def.scopes.includes('groups') ? 'groups' : def.scopes[0]
  } else {
    delete scopes[def.key]
  }
  form.value = { ...form.value, scopes }
}

const saveRole = async () => {
  const payload = {
    name: form.value.name.trim(),
    description: form.value.description.trim(),
    permissions: Object.entries(form.value.scopes).map(([permission, scope]) => ({ permission, scope })),
  }
  saving.value = true
  try {
    if (editingId.value) {
      await adminService.updateAccessRole(editingId.value, payload)
    } else {
      await adminService.createAccessRole(payload)
    }
    toast('success', t('common.feedback.saved'))
    showForm.value = false
    await loadRoles()
  } catch (error) {
    const status = error.response?.status
    toast('error', status === 409 ? t('common.validation.duplicateName') : error.response?.data?.error || t('roles.errors.save'))
  } finally {
    saving.value = false
  }
}

const removeRole = async (role) => {
  const confirmed = await askConfirm({
    title: t('common.delete'),
    message: t('roles.deleteConfirm', { name: role.name, count: role.user_count || 0 }),
    confirmLabel: t('common.delete'),
    tone: 'danger',
  })
  if (!confirmed) return
  try {
    await adminService.deleteAccessRole(role.id)
    toast('success', t('common.feedback.deleted'))
    await loadRoles()
  } catch (error) {
    console.error('Error deleting role:', error)
    toast('error', t('roles.errors.delete'))
  }
}

// --- Membres ---

const displayName = (user) =>
  [user.first_name, user.last_name].filter(Boolean).join(' ') || user.username

const matchesQuery = (user) => {
  const q = memberQuery.value.trim().toLowerCase()
  if (!q) return true
  return [user.username, user.email, user.first_name, user.last_name].some(v => v?.toLowerCase().includes(q))
}

// Membres actuels (filtrés localement) puis résultats de la recherche, sans doublon
const memberCandidates = computed(() => {
  const members = currentMembers.value.filter(matchesQuery)
  const ids = new Set(members.map(u => u.id))
  return [...members, ...searchResults.value.filter(u => !ids.has(u.id))]
})

let searchRequest = 0
const searchUsers = async () => {
  const request = ++searchRequest
  try {
    const data = await adminService.getUsers({ search: memberQuery.value.trim() || undefined, page_size: USER_SEARCH_PAGE_SIZE })
    if (request !== searchRequest) return
    searchResults.value = data.data || []
    searchTotal.value = data.total || 0
  } catch (error) {
    if (request !== searchRequest) return
    console.error('Error searching users:', error)
    toast('error', t('roles.errors.members'))
  }
}

const openMembers = async (role) => {
  membersRole.value = role
  memberQuery.value = ''
  loadingMembers.value = true
  try {
    const [members] = await Promise.all([adminService.getAccessRoleMembers(role.id), searchUsers()])
    currentMembers.value = Array.isArray(members) ? members : []
    selectedMemberIds.value = currentMembers.value.map(u => u.id)
  } catch (error) {
    console.error('Error loading role members:', error)
    toast('error', t('roles.errors.members'))
    membersRole.value = null
  } finally {
    loadingMembers.value = false
  }
}

const closeMembers = () => {
  membersRole.value = null
  currentMembers.value = []
  searchResults.value = []
}

const saveMembers = async () => {
  saving.value = true
  try {
    await adminService.setAccessRoleMembers(membersRole.value.id, selectedMemberIds.value)
    toast('success', t('common.feedback.saved'))
    closeMembers()
    await loadRoles()
  } catch (error) {
    console.error('Error saving role members:', error)
    toast('error', t('roles.errors.members'))
  } finally {
    saving.value = false
  }
}

let searchTimer = null
watch(memberQuery, () => {
  if (!membersRole.value) return
  clearTimeout(searchTimer)
  searchTimer = setTimeout(searchUsers, SEARCH_DEBOUNCE_MS)
})
onBeforeUnmount(() => clearTimeout(searchTimer))

onMounted(loadRoles)

// Échap ferme la modale ouverte
useEscapeClose(showForm, () => closeForm())
useEscapeClose(computed(() => !!membersRole.value), () => closeMembers())
</script>

<style scoped>
.rm-admin {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-block-end: 16px;
  padding: 12px 16px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--sunken);
}
.rm-admin__icon { flex: none; width: 22px; height: 22px; color: var(--acc-ink); }
.rm-admin__text { flex: 1; min-width: 0; }
.rm-admin__name { margin: 0; font-size: 14px; font-weight: 600; color: var(--text); }
.rm-admin__help { margin: 2px 0 0; font-size: var(--fs-small); color: var(--muted); }

.rm-name { font-weight: 600; color: var(--text); }
.rm-desc { display: block; font-size: var(--fs-small); color: var(--muted); }
.rm-muted { margin: 0; font-size: var(--fs-small); color: var(--muted); }
.rm-num { font-family: var(--f-num); }
.rm-tags { display: flex; flex-wrap: wrap; gap: 6px; }

.rm-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: rgb(0 0 0 / 0.5);
}

.rm-modal {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 640px;
  max-height: calc(100vh - 32px);
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
  color: var(--text);
}

.rm-modal__head,
.rm-modal__foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 16px;
}
.rm-modal__head { border-block-end: 1px solid var(--line); }
.rm-modal__foot { border-block-start: 1px solid var(--line); flex-wrap: wrap; }
.rm-modal__title { margin: 0; font-size: 16px; font-weight: 600; }

.rm-form { display: grid; gap: 14px; padding: 16px; overflow-y: auto; }
.rm-form__actions { display: flex; justify-content: flex-end; gap: 8px; }
.rm-textarea { min-height: 0; resize: vertical; }

.rm-perms { display: grid; gap: 8px; margin: 0; padding: 0; border: 0; min-width: 0; }

.rm-perm {
  padding: 10px 12px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
}
.rm-perm.is-on { border-color: var(--acc-ink); }
.rm-perm__check { display: flex; align-items: flex-start; gap: 10px; cursor: pointer; }
.rm-perm__check input { margin-block-start: 3px; }
.rm-perm__label { display: block; font-weight: 600; font-size: var(--fs-body); }
.rm-perm__help { display: block; font-size: var(--fs-small); color: var(--muted); }

.rm-scope {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 16px;
  margin-block-start: 10px;
  padding-inline-start: 26px;
}
.rm-scope__opt { display: inline-flex; align-items: center; gap: 6px; font-size: var(--fs-small); cursor: pointer; }
.rm-scope__help { flex-basis: 100%; margin: 0; font-size: var(--fs-small); color: var(--muted); }

.rm-members { display: grid; gap: 10px; padding: 16px; overflow-y: auto; min-height: 0; }
.rm-members__list { display: grid; gap: 6px; margin: 0; padding: 0; list-style: none; }
</style>
