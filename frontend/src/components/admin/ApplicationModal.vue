<template>
  <div v-if="show" class="modal-overlay" @click.self="closeModal">
    <div class="modal-container" @click.self="closeModal">
      <div class="modal-panel w-full sm:max-w-5xl">
        <form class="modal-form" @submit.prevent="handleSubmit">
          <!-- En-tête -->
          <div class="modal-header">
            <div class="flex items-center justify-between gap-4">
              <div class="flex items-center gap-3 min-w-0">
                <div
                  class="h-10 w-10 flex-shrink-0 rounded-lg flex items-center justify-center text-white shadow-sm"
                  :style="{ backgroundColor: form.color }"
                >
                  <Icon :icon="form.icon || 'mdi:application'" class="h-5 w-5" />
                </div>
                <div class="min-w-0">
                  <h3 class="modal-title truncate">
                    {{ isEdit ? $t('groupAdmin.applicationModal.title.edit') : $t('groupAdmin.applicationModal.title.new') }}
                  </h3>
                  <p class="modal-subtitle truncate">
                    {{ isEdit ? $t('groupAdmin.applicationModal.subtitle.edit') : $t('groupAdmin.applicationModal.subtitle.new') }}
                  </p>
                </div>
              </div>
              <button type="button" class="btn btn-ghost p-2" :aria-label="$t('common.close')" @click="closeModal">
                <Icon icon="mdi:close" class="h-5 w-5" />
              </button>
            </div>
          </div>

          <!-- Contenu : infos | apparence -->
          <div class="modal-content grid grid-cols-1 lg:grid-cols-2 gap-x-8 gap-y-6">
            <!-- Colonne gauche -->
            <div class="space-y-4">
              <div class="section-header mb-0">
                <Icon icon="mdi:information-outline" class="section-icon" />
                <h4 class="section-title">{{ $t('groupAdmin.applicationModal.basicInfo') }}</h4>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div class="form-group mb-0">
                  <label for="name" class="form-label form-label-required">{{ $t('groupAdmin.applicationModal.applicationName') }}</label>
                  <input id="name" v-model="form.name" type="text" required class="form-input" placeholder="GitLab" />
                  <p v-if="errors.name" class="form-error">{{ errors.name }}</p>
                </div>
                <div class="form-group mb-0">
                  <label for="app_group_id" class="form-label form-label-required">{{ $t('groupAdmin.applicationModal.appGroup') }}</label>
                  <select id="app_group_id" v-model="form.app_group_id" required class="form-select">
                    <option value="">{{ $t('groupAdmin.applicationModal.selectGroup') }}</option>
                    <option v-for="group in availableAppGroups" :key="group.id" :value="group.id">{{ group.name }}</option>
                  </select>
                  <p v-if="errors.app_group_id" class="form-error">{{ errors.app_group_id }}</p>
                  <!-- Admin de groupe : seuls les groupes de ressources privés de ses groupes sont proposés -->
                  <p v-if="privateGroupsOnly && availableAppGroups.length === 0" class="form-help text-yellow-500">
                    {{ $t('groupAdmin.applicationModal.noPrivateGroups') }}
                  </p>
                </div>
              </div>

              <div class="form-group mb-0">
                <label for="url" class="form-label form-label-required">{{ $t('groupAdmin.applicationModal.applicationUrl') }}</label>
                <div class="relative">
                  <Icon icon="mdi:link-variant" class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                  <input
                    id="url"
                    v-model="form.url"
                    type="url"
                    required
                    class="form-input !pl-9"
                    :placeholder="$t('groupAdmin.applicationModal.urlPlaceholder')"
                  />
                </div>
                <p v-if="errors.url" class="form-error">{{ errors.url }}</p>
                <p v-else class="text-xs text-gray-400 mt-1">{{ $t('groupAdmin.applicationModal.urlHelp') }}</p>
              </div>

              <div class="form-group mb-0">
                <label for="description" class="form-label">{{ $t('groupAdmin.applicationModal.description') }}</label>
                <textarea
                  id="description"
                  v-model="form.description"
                  rows="2"
                  class="form-textarea resize-none"
                  :placeholder="$t('groupAdmin.applicationModal.descriptionPlaceholder')"
                ></textarea>
              </div>

              <!-- Note courte affichée sur la tuile du portail, jusqu'à une date facultative -->
              <div class="grid grid-cols-1 sm:grid-cols-[1fr_auto] gap-3">
                <div class="form-group mb-0">
                  <label for="admin_note" class="form-label">{{ $t('adminShared.applications.note') }}</label>
                  <input
                    id="admin_note"
                    v-model="form.admin_note"
                    type="text"
                    maxlength="140"
                    class="form-input"
                    aria-describedby="admin_note_help"
                    :placeholder="$t('adminShared.applications.notePlaceholder')"
                  />
                  <p id="admin_note_help" class="text-xs text-gray-400 mt-1">
                    {{ $t('adminShared.applications.noteHelp', { count: form.admin_note.length }) }}
                  </p>
                </div>
                <div class="form-group mb-0">
                  <label for="admin_note_until" class="form-label">{{ $t('adminShared.applications.noteUntil') }}</label>
                  <input
                    id="admin_note_until"
                    v-model="form.admin_note_until"
                    type="date"
                    class="form-input"
                    :disabled="!form.admin_note.trim()"
                  />
                </div>
              </div>

              <!-- Créateur : seulement si la page fournit la liste des utilisateurs (admin global) -->
              <div v-if="users.length" class="form-group mb-0">
                <label for="created_by_id" class="form-label">{{ $t('adminShared.applications.creator') }}</label>
                <select id="created_by_id" v-model="form.created_by_id" class="form-select">
                  <option :value="null">{{ $t('adminShared.applications.creatorNone') }}</option>
                  <option v-for="u in users" :key="u.id" :value="u.id">
                    {{ u.first_name || u.last_name ? `${u.first_name} ${u.last_name}`.trim() : u.username }}
                    <template v-if="u.first_name || u.last_name"> ({{ u.username }})</template>
                  </option>
                </select>
                <p class="text-xs text-gray-400 mt-1">{{ $t('adminShared.applications.creatorHelp') }}</p>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                <label class="toggle-row">
                  <span>
                    <span class="toggle-row-title">{{ $t('groupAdmin.applicationModal.openInNewTab') }}</span>
                  </span>
                  <input v-model="form.open_in_new_tab" type="checkbox" class="switch-input" />
                  <span class="switch-track" aria-hidden="true"></span>
                </label>
                <label class="toggle-row">
                  <span>
                    <span class="toggle-row-title">{{ $t('common.active') }}</span>
                    <span class="toggle-row-desc">{{ $t('groupAdmin.applicationModal.activeAndVisible') }}</span>
                  </span>
                  <input v-model="form.is_active" type="checkbox" class="switch-input" />
                  <span class="switch-track" aria-hidden="true"></span>
                </label>
              </div>
            </div>

            <!-- Colonne droite -->
            <div class="space-y-4">
              <div class="section-header mb-0">
                <Icon icon="mdi:palette-outline" class="section-icon" />
                <h4 class="section-title">{{ $t('groupAdmin.applicationModal.appearance') }}</h4>
              </div>

              <div class="card p-3">
                <div class="flex items-center gap-3">
                  <Icon :icon="form.icon || 'mdi:application'" class="h-9 w-9 flex-shrink-0" :style="{ color: form.color }" />
                  <div class="min-w-0 flex-1">
                    <div class="font-medium text-gray-900 dark:text-white truncate">
                      {{ form.name || $t('groupAdmin.applicationModal.previewName') }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400 truncate">
                      {{ form.url || $t('groupAdmin.applicationModal.previewUrl') }}
                    </div>
                  </div>
                  <Icon v-if="form.open_in_new_tab" icon="mdi:open-in-new" class="h-4 w-4 text-gray-400" />
                </div>
              </div>

              <div class="grid grid-cols-[1fr_7rem] gap-4">
                <div class="form-group mb-0">
                  <label for="color" class="form-label">{{ $t('groupAdmin.applicationModal.color') }}</label>
                  <div class="flex items-center gap-2">
                    <input
                      id="color"
                      v-model="form.color"
                      type="color"
                      class="h-10 w-12 flex-shrink-0 rounded-lg border cursor-pointer p-1"
                      style="border-color: var(--border); background-color: var(--bg-card)"
                    />
                    <input v-model="form.color" type="text" class="form-input font-mono" placeholder="#6366f1" />
                  </div>
                </div>
                <div class="form-group mb-0">
                  <label for="order" class="form-label">{{ $t('groupAdmin.applicationModal.displayOrder') }}</label>
                  <input id="order" v-model.number="form.order" type="number" min="0" class="form-input" placeholder="0" />
                </div>
              </div>

              <IconInput
                v-model="form.icon"
                :label="$t('groupAdmin.applicationModal.icon')"
                placeholder="mdi:application, lucide:app-window, heroicons:squares-2x2"
                category="application"
                :show-suggestions="true"
              />
            </div>
          </div>

          <!-- Pied -->
          <div class="modal-footer-bar">
            <span class="modal-footer-hint">{{ $t('common.requiredFieldsHint') }}</span>
            <div class="modal-footer-actions">
              <button type="button" class="btn btn-secondary" @click="closeModal">{{ $t('common.cancel') }}</button>
              <button type="submit" :disabled="loading || availableAppGroups.length === 0" class="btn btn-primary">
                <Icon v-if="loading" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
                <Icon v-else :icon="isEdit ? 'mdi:content-save' : 'mdi:plus'" class="h-4 w-4 mr-2" />
                {{ isEdit ? $t('groupAdmin.applicationModal.update') : $t('groupAdmin.applicationModal.create') }}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
import { noteUntilFromInput, noteUntilToInput } from '@/utils/resourceType'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import IconInput from '@/components/ui/IconInput.vue'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  application: {
    type: Object,
    default: null
  },
  appGroups: {
    type: Array,
    default: () => []
  },
  // Liste des utilisateurs pour choisir le créateur (admin global uniquement)
  users: {
    type: Array,
    default: () => []
  },
  // Admin de groupe : ne proposer que les groupes de ressources privés
  privateGroupsOnly: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'submit'])
const { t } = useI18n()

const DEFAULT_COLOR = '#6366f1'
const DEFAULT_ICON = 'mdi:application'

const loading = ref(false)
const errors = ref({})

/** Valeurs du formulaire à partir d'une ressource existante (ou vides) */
const toFormValues = (app) => ({
  name: app?.name || '',
  description: app?.description || '',
  url: app?.url || '',
  app_group_id: app?.app_group_id || '',
  color: app?.color || DEFAULT_COLOR,
  icon: app?.icon || DEFAULT_ICON,
  order: app?.order || 0,
  open_in_new_tab: app?.open_in_new_tab ?? true,
  is_active: app?.is_active ?? true,
  created_by_id: app?.created_by_id || null,
  admin_note: app?.admin_note || '',
  admin_note_until: noteUntilToInput(app?.admin_note_until)
})

const form = reactive(toFormValues(null))

const isEdit = computed(() => !!props.application)

const availableAppGroups = computed(() =>
  props.privateGroupsOnly
    ? props.appGroups.filter(group => group.is_private === true)
    : props.appGroups
)

const resetForm = () => {
  Object.assign(form, toFormValues(props.application))
  errors.value = {}
}

watch(() => props.show, (isOpen) => {
  if (isOpen) resetForm()
})

watch(() => props.application, (app) => {
  if (app) Object.assign(form, toFormValues(app))
}, { immediate: true })

const isValidUrl = (url) => {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

const validateForm = () => {
  const nextErrors = {}

  if (!form.name.trim()) {
    nextErrors.name = t('groupAdmin.applicationModal.errors.nameRequired')
  }

  if (!form.url.trim()) {
    nextErrors.url = t('groupAdmin.applicationModal.errors.urlRequired')
  } else if (!isValidUrl(form.url)) {
    nextErrors.url = t('groupAdmin.applicationModal.errors.invalidUrl')
  }

  if (!form.app_group_id) {
    nextErrors.app_group_id = t('groupAdmin.applicationModal.errors.appGroupRequired')
  }

  errors.value = nextErrors
  return Object.keys(nextErrors).length === 0
}

const handleSubmit = () => {
  if (!validateForm()) return

  loading.value = true
  try {
    const note = form.admin_note.trim()
    const values = { ...form, admin_note: note, admin_note_until: note ? noteUntilFromInput(form.admin_note_until) : null }
    const formData = isEdit.value
      ? { ...values, id: props.application.id }
      : values
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
