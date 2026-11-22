<template>
  <div v-if="show" class="modal-overlay" @click.self="closeModal">
    <div class="modal-container" @click.self="closeModal">
      <div v-focus-trap class="modal-panel w-full sm:max-w-3xl" role="dialog" aria-modal="true" aria-labelledby="user-modal-title">
        <form class="modal-form" novalidate @submit.prevent="handleSubmit">
          <!-- Header -->
          <div class="modal-header">
            <div class="flex items-center justify-between gap-4">
              <div class="flex items-center gap-3 min-w-0">
                <div
                  class="h-10 w-10 flex-shrink-0 rounded-full flex items-center justify-center text-sm font-semibold text-white shadow-sm"
                  style="background-color: var(--accent)"
                >
                  <span v-if="initials">{{ initials }}</span>
                  <Icon v-else icon="lucide:user" class="h-5 w-5" />
                </div>
                <div class="min-w-0">
                  <h3 id="user-modal-title" class="modal-title truncate">
                    {{ isEdit ? t('userModal.titleEdit') : t('users.new') }}
                  </h3>
                  <p class="modal-subtitle truncate">
                    {{ isEdit ? t('userModal.subtitleEdit') : t('userModal.subtitleCreate') }}
                  </p>
                </div>
              </div>
              <button type="button" class="btn btn-ghost p-2" :aria-label="t('common.close')" @click="closeModal">
                <Icon icon="lucide:x" class="h-5 w-5" />
              </button>
            </div>
          </div>

          <!-- Tabs -->
          <div class="modal-tabs" role="tablist">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              type="button"
              role="tab"
              :aria-selected="activeTab === tab.id"
              class="modal-tab"
              :class="{ 'modal-tab-active': activeTab === tab.id }"
              @click="activeTab = tab.id"
            >
              <Icon :icon="tab.icon" class="h-4 w-4" />
              {{ tab.label }}
              <span v-if="tab.count !== undefined" class="modal-tab-count">{{ tab.count }}</span>
              <span v-if="tab.hasError" class="modal-tab-error" :aria-label="t('userModal.tabError')"></span>
            </button>
          </div>

          <!-- Content -->
          <div class="modal-content modal-body-tabbed">
            <!-- Onglet : Compte -->
            <div v-show="activeTab === 'account'" class="space-y-5">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="form-group mb-0">
                  <label for="username" class="form-label form-label-required">{{ t('userModal.username') }}</label>
                  <input
                    id="username"
                    v-model="form.username"
                    type="text"
                    required
                    :disabled="isEdit"
                    class="form-input"
                    :class="{ 'opacity-60 cursor-not-allowed': isEdit }"
                    placeholder="johndoe"
                  />
                  <p v-if="errors.username" class="form-error">{{ errors.username }}</p>
                  <p v-else-if="isEdit" class="text-xs text-gray-400 mt-1">{{ t('userModal.usernameLocked') }}</p>
                </div>

                <div class="form-group mb-0">
                  <label for="email" class="form-label form-label-required">{{ t('userModal.email') }}</label>
                  <input
                    id="email"
                    v-model="form.email"
                    type="email"
                    required
                    class="form-input"
                    placeholder="john.doe@example.com"
                  />
                  <p v-if="errors.email" class="form-error">{{ errors.email }}</p>
                </div>

                <div class="form-group mb-0">
                  <label for="first_name" class="form-label">{{ t('userModal.firstName') }}</label>
                  <input id="first_name" v-model="form.first_name" type="text" class="form-input" placeholder="John" />
                </div>

                <div class="form-group mb-0">
                  <label for="last_name" class="form-label">{{ t('userModal.lastName') }}</label>
                  <input id="last_name" v-model="form.last_name" type="text" class="form-input" placeholder="Doe" />
                </div>

                <div v-if="!isEdit" class="form-group mb-0 md:col-span-2">
                  <label for="password" class="form-label form-label-required">{{ t('userModal.password') }}</label>
                  <div class="relative">
                    <input
                      id="password"
                      v-model="form.password"
                      :type="showPassword ? 'text' : 'password'"
                      required
                      class="form-input !pe-10"
                      placeholder="••••••••"
                      minlength="6"
                      autocomplete="new-password"
                    />
                    <button
                      type="button"
                      class="absolute inset-y-0 end-0 px-3 flex items-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                      :aria-label="showPassword ? t('userModal.hidePassword') : t('userModal.showPassword')"
                      @click="showPassword = !showPassword"
                    >
                      <Icon :icon="showPassword ? 'lucide:eye-off' : 'lucide:eye'" class="h-4 w-4" />
                    </button>
                  </div>
                  <p v-if="errors.password" class="form-error">{{ errors.password }}</p>
                  <p v-else class="text-xs text-gray-400 mt-1">{{ t('userModal.passwordHint', { min: PASSWORD_MIN_LENGTH }) }}</p>
                </div>
              </div>

              <label class="toggle-row">
                <span>
                  <span class="toggle-row-title">{{ t('userModal.activeAccount') }}</span>
                  <span class="toggle-row-desc">{{ t('userModal.activeAccountDesc') }}</span>
                </span>
                <input v-model="form.is_active" type="checkbox" class="switch-input" />
                <span class="switch-track" aria-hidden="true"></span>
              </label>
            </div>

            <!-- Onglet : Rôle & groupes -->
            <div v-show="activeTab === 'access'" class="space-y-5">
              <div>
                <span class="form-label form-label-required">{{ t('userModal.role') }}</span>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <label
                    v-for="role in roles"
                    :key="role.value"
                    class="choice-card items-start"
                    :class="{ 'choice-card-active': form.role === role.value }"
                  >
                    <input v-model="form.role" type="radio" name="role" :value="role.value" class="mt-0.5" />
                    <span class="min-w-0">
                      <span class="flex items-center gap-1.5 text-sm font-medium text-gray-900 dark:text-white">
                        <Icon :icon="role.icon" class="h-4 w-4 text-gray-500" />
                        {{ role.label }}
                      </span>
                      <span class="block text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ role.desc }}</span>
                    </span>
                  </label>
                </div>
                <p v-if="errors.role" class="form-error">{{ errors.role }}</p>
              </div>

              <div>
                <div class="flex items-center justify-between gap-3 mb-2">
                  <label class="form-label mb-0" for="user-groups">{{ t('userModal.groups') }}</label>
                  <span class="text-xs text-gray-400">{{ t('userModal.groupsHint') }}</span>
                </div>
                <p v-if="groups.length === 0" class="text-sm text-gray-500 py-2">{{ t('userModal.noGroups') }}</p>
                <ComboSelect
                  v-else
                  v-model="form.group_ids"
                  input-id="user-groups"
                  kind="group"
                  :options="groupOptions"
                  :label="t('userModal.groups')"
                  :placeholder="t('userModal.groupsPlaceholder')"
                />
              </div>

              <!-- Rôles attribués (permissions) : sans objet pour un administrateur -->
              <fieldset>
                <legend class="form-label">{{ t('userModal.accessRoles') }}</legend>
                <p v-if="form.role === 'admin'" class="text-sm text-gray-500 dark:text-gray-400">
                  {{ t('userModal.accessRolesAdmin') }}
                </p>
                <template v-else>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mb-2">{{ t('userModal.accessRolesHint') }}</p>
                  <p v-if="accessRolesError" class="form-error">{{ t('userModal.accessRolesLoadError') }}</p>
                  <p v-else-if="accessRoles.length === 0" class="text-sm text-gray-500 text-center py-4">
                    {{ t('userModal.noAccessRoles') }}
                  </p>
                  <ul v-else class="role-assign">
                    <li
                      v-for="accessRole in accessRoles"
                      :key="accessRole.id"
                      class="role-assign__item"
                      :class="{ 'role-assign__item--on': isAssigned(accessRole.id) }"
                    >
                      <label class="role-assign__head">
                        <input
                          type="checkbox"
                          :checked="isAssigned(accessRole.id)"
                          @change="toggleRole(accessRole.id, $event.target.checked)"
                        />
                        <span class="min-w-0">
                          <span class="block text-sm font-medium text-gray-900 dark:text-white truncate">{{ accessRole.name }}</span>
                          <span v-if="accessRole.description" class="block text-xs text-gray-500 dark:text-gray-400 mt-0.5 line-clamp-2">{{ accessRole.description }}</span>
                        </span>
                      </label>
                      <!-- Portée « ses groupes » : quels groupes ce rôle couvre pour ce compte -->
                      <div v-if="isAssigned(accessRole.id) && hasGroupScope(accessRole)" class="role-assign__groups">
                        <ComboSelect
                          :model-value="groupsOf(accessRole.id)"
                          kind="group"
                          :options="roleGroupOptions"
                          :label="t('userModal.roleGroups', { role: accessRole.name })"
                          :placeholder="t('userModal.roleGroupsPlaceholder')"
                          @update:model-value="setRoleGroups(accessRole.id, $event)"
                        />
                        <p class="role-assign__hint">
                          <Icon :icon="groupsOf(accessRole.id).length ? 'lucide:target' : 'lucide:users'" aria-hidden="true" />
                          {{ groupsOf(accessRole.id).length ? t('userModal.roleGroupsChosen') : t('userModal.roleGroupsDefault') }}
                        </p>
                      </div>
                    </li>
                  </ul>
                </template>
              </fieldset>
            </div>
          </div>

          <!-- Footer -->
          <div class="modal-footer-bar">
            <span class="modal-footer-hint">{{ t('common.requiredFieldsHint') }}</span>
            <div class="modal-footer-actions">
              <button type="button" class="btn btn-secondary" @click="closeModal">{{ t('common.cancel') }}</button>
              <button type="submit" :disabled="loading" class="btn btn-primary">
                <Icon v-if="loading" icon="lucide:loader-circle" class="animate-spin h-4 w-4 me-2" />
                <Icon v-else :icon="isEdit ? 'lucide:save' : 'lucide:user-plus'" class="h-4 w-4 me-2" />
                {{ isEdit ? t('common.save') : t('userModal.submitCreate') }}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { adminService } from '@/services/api'
import ComboSelect from '@/components/ui/ComboSelect.vue'
import { groupOption } from '@/utils/comboFilter'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  user: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['close', 'submit'])

const { t } = useI18n()

const USERNAME_MIN_LENGTH = 3
const PASSWORD_MIN_LENGTH = 6

const loading = ref(false)
const errors = ref({})
const groups = ref([])

const form = reactive({
  username: '',
  email: '',
  password: '',
  first_name: '',
  last_name: '',
  role: 'user',
  group_ids: [],
  // Rôles attribués : [{ id, group_ids }] (group_ids vide : les groupes dont le compte est membre)
  access_roles: [],
  is_active: true
})

// Rôles créés par l'administrateur (page Rôles)
const accessRoles = ref([])
const accessRolesError = ref(false)

const isEdit = computed(() => !!props.user)

const activeTab = ref('account')
const showPassword = ref(false)

const roles = computed(() => [
  { value: 'user', label: t('users.role_user'), icon: 'lucide:user', desc: t('userModal.roleUserDesc') },
  { value: 'admin', label: t('users.role_admin'), icon: 'lucide:shield-check', desc: t('userModal.roleAdminDesc') }
])

const initials = computed(() => {
  const f = form.first_name?.trim()
  const l = form.last_name?.trim()
  if (f || l) return `${f?.charAt(0) || ''}${l?.charAt(0) || ''}`.toUpperCase()
  return form.username?.trim().charAt(0).toUpperCase() || ''
})

const tabs = computed(() => [
  {
    id: 'account',
    label: t('userModal.tabAccount'),
    icon: 'lucide:id-card',
    hasError: !!(errors.value.username || errors.value.email || errors.value.password)
  },
  {
    id: 'access',
    label: t('userModal.tabAccess'),
    icon: 'lucide:shield-user',
    count: form.group_ids.length + (form.role === 'admin' ? 0 : form.access_roles.length),
    hasError: !!errors.value.role
  }
])

// Load groups when component mounts
onMounted(async () => {
  try {
    const data = await adminService.getGroups()
    groups.value = Array.isArray(data) ? data : []
  } catch (error) {
    console.error('Erreur lors du chargement des groupes:', error)
    groups.value = []
  }
  try {
    const list = await adminService.getAccessRoles()
    accessRoles.value = Array.isArray(list) ? list : []
  } catch (error) {
    console.error('Erreur lors du chargement des rôles:', error)
    accessRolesError.value = true
  }
})

// Reset form when modal opens/closes
watch(() => props.show, (newVal) => {
  if (newVal) {
    resetForm()
  }
})

watch(() => props.user, (newVal) => {
  if (newVal) {
    Object.assign(form, {
      username: newVal.username || '',
      email: newVal.email || '',
      password: '',
      first_name: newVal.first_name || '',
      last_name: newVal.last_name || '',
      role: accountRole(newVal.role),
      group_ids: newVal.groups?.map(g => g.id) || [],
      access_roles: assignmentsOf(newVal.access_roles),
      is_active: newVal.is_active ?? true
    })
  }
}, { immediate: true })

// Rôles d'un compte tels que le serveur les renvoie -> [{ id, group_ids }]
const assignmentsOf = (list) =>
  (Array.isArray(list) ? list : []).map((role) => ({ id: role.id, group_ids: Array.isArray(role.group_ids) ? [...role.group_ids] : [] }))

// Les groupes choisis ne sont pas dans la liste des comptes : on relit les rôles du compte
let assignmentsRequest = 0
const loadAssignments = async (userId) => {
  const request = ++assignmentsRequest
  try {
    const list = await adminService.getUserAccessRoles(userId)
    if (request === assignmentsRequest && props.user?.id === userId) form.access_roles = assignmentsOf(list)
  } catch (error) {
    console.error('Erreur lors du chargement des rôles du compte:', error)
  }
}
watch(() => (props.show ? props.user?.id : null), (id) => { if (id) loadAssignments(id) }, { immediate: true })

const groupOptions = computed(() =>
  groups.value.map((group) => groupOption(group, t('userModal.membersCount', group.users?.length || 0)))
)
// Le groupe par défaut réunit tout le monde : il ne se choisit pas pour un rôle limité
const roleGroupOptions = computed(() => groupOptions.value.filter((option) => !groups.value.find((g) => g.id === option.id)?.is_default))

const isAssigned = (roleId) => form.access_roles.some((a) => a.id === roleId)
const groupsOf = (roleId) => form.access_roles.find((a) => a.id === roleId)?.group_ids || []
const hasGroupScope = (role) => (role.permissions || []).some((p) => p.scope === 'groups')
const toggleRole = (roleId, checked) => {
  form.access_roles = checked
    ? [...form.access_roles.filter((a) => a.id !== roleId), { id: roleId, group_ids: [] }]
    : form.access_roles.filter((a) => a.id !== roleId)
}
const setRoleGroups = (roleId, groupIds) => {
  form.access_roles = form.access_roles.map((a) => (a.id === roleId ? { ...a, group_ids: groupIds } : a))
}

// Rôle de compte : « admin » ou « user » (l'ancien « editor » n'existe plus)
const accountRole = (role) => (role === 'admin' ? 'admin' : 'user')

const resetForm = () => {
  if (props.user) {
    Object.assign(form, {
      username: props.user.username || '',
      email: props.user.email || '',
      password: '',
      first_name: props.user.first_name || '',
      last_name: props.user.last_name || '',
      role: accountRole(props.user.role),
      group_ids: props.user.groups?.map(g => g.id) || [],
      access_roles: assignmentsOf(props.user.access_roles),
      is_active: props.user.is_active ?? true
    })
  } else {
    Object.assign(form, {
      username: '',
      email: '',
      password: '',
      first_name: '',
      last_name: '',
      role: 'user',
      group_ids: [],
      access_roles: [],
      is_active: true
    })
  }
  errors.value = {}
  activeTab.value = 'account'
  showPassword.value = false
}

const validateForm = () => {
  errors.value = {}
  
  if (!form.username.trim()) {
    errors.value.username = t('userModal.errors.usernameRequired')
  } else if (form.username.length < USERNAME_MIN_LENGTH) {
    errors.value.username = t('userModal.errors.usernameTooShort', { min: USERNAME_MIN_LENGTH })
  }
  
  if (!form.email.trim()) {
    errors.value.email = t('userModal.errors.emailRequired')
  } else if (!isValidEmail(form.email)) {
    errors.value.email = t('userModal.errors.emailInvalid')
  }
  
  if (!isEdit.value) {
    if (!form.password.trim()) {
      errors.value.password = t('userModal.errors.passwordRequired')
    } else if (form.password.length < PASSWORD_MIN_LENGTH) {
      errors.value.password = t('userModal.errors.passwordTooShort', { min: PASSWORD_MIN_LENGTH })
    }
  }
  
  if (!form.role) {
    errors.value.role = t('userModal.errors.roleRequired')
  }
  
  return Object.keys(errors.value).length === 0
}

const isValidEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

const handleSubmit = async () => {
  if (!validateForm()) {
    const e = errors.value
    activeTab.value = (e.username || e.email || e.password) ? 'account' : 'access'
    return
  }
  
  loading.value = true
  try {
    // Les rôles attribués ne concernent pas un administrateur (il a tout)
    const formData = {
      ...form,
      access_roles: form.role === 'admin' ? [] : form.access_roles.map((a) => ({ id: a.id, group_ids: [...a.group_ids] }))
    }
    if (isEdit.value) {
      formData.id = props.user.id
      // Don't send password if it's empty in edit mode
      if (!formData.password) {
        delete formData.password
      }
    }
    
    emit('submit', formData)
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  resetForm()
  emit('close')
}

// Échap ferme la modale
useEscapeClose(() => props.show, () => closeModal())
</script>

<style scoped>
.role-assign {
  display: grid;
  gap: 8px;
}

.role-assign__item {
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background-color: var(--surface);
  transition: border-color 120ms ease, background-color 120ms ease;
}

.role-assign__item:hover {
  border-color: color-mix(in oklab, var(--acc) 50%, var(--line));
}

.role-assign__item--on {
  border-color: var(--acc);
  background-color: color-mix(in oklab, var(--acc) 5%, var(--surface));
}

.role-assign__head {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 10px 12px;
  cursor: pointer;
}

.role-assign__head input {
  flex: none;
  width: 16px;
  height: 16px;
  margin-top: 2px;
  accent-color: var(--acc);
  cursor: pointer;
}

.role-assign__head:has(input:focus-visible) {
  outline: 2px solid var(--fill);
  outline-offset: -2px;
  border-radius: var(--r-s);
}

/* Groupes du rôle : rattachés à la carte, décalés sous son intitulé */
.role-assign__groups {
  margin-inline: 40px 12px;
  padding-block: 2px 12px;
}

.role-assign__hint {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  color: var(--muted);
  font-size: var(--fs-small);
}

.role-assign__hint :deep(svg) {
  width: 13px;
  height: 13px;
  flex: none;
}

@media (max-width: 640px) {
  .role-assign__groups {
    margin-inline: 12px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .role-assign__item {
    transition: none;
  }
}
</style>
