<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-container">
      <div class="modal-panel sm:max-w-4xl sm:w-full">
        <form @submit.prevent="handleSubmit">
          <!-- Header -->
          <div class="modal-header">
            <div class="flex items-center justify-between">
              <div>
                <h3 class="modal-title">
                  {{ isEdit ? t('polls.editor.title.edit') : t('polls.editor.title.new') }}
                </h3>
                <p class="modal-subtitle">
                  {{ isEdit ? t('polls.editor.subtitle.edit') : t('polls.editor.subtitle.new') }}
                </p>
              </div>
              <button
                type="button"
                @click="closeModal"
                class="btn-ghost p-2"
              >
                <Icon icon="mdi:close" class="h-5 w-5" />
              </button>
            </div>
          </div>

          <!-- Content -->
          <div class="modal-content space-y-6">
            <!-- Informations de base -->
            <div>
              <div class="section-header">
                <Icon icon="mdi:information-outline" class="section-icon" />
                <h4 class="section-title">{{ t('polls.editor.basicInfo') }}</h4>
              </div>

              <div class="space-y-4">
                <div class="form-group">
                  <label for="title" class="form-label form-label-required">
                    {{ t('polls.editor.pollQuestion') }}
                  </label>
                  <input
                    id="title"
                    v-model="form.title"
                    type="text"
                    required
                    class="form-input"
                    :placeholder="t('polls.editor.pollQuestionPlaceholder')"
                  />
                  <p v-if="errors.title" class="form-error">{{ errors.title }}</p>
                </div>

                <div class="form-group">
                  <label for="description" class="form-label">
                    {{ t('polls.editor.description') }}
                  </label>
                  <textarea
                    id="description"
                    v-model="form.description"
                    rows="2"
                    class="form-textarea"
                    :placeholder="t('polls.editor.descriptionPlaceholder')"
                  ></textarea>
                </div>
              </div>
            </div>

            <!-- Options du sondage -->
            <div>
              <div class="section-header">
                <Icon icon="mdi:format-list-checkbox" class="section-icon" />
                <h4 class="section-title">{{ t('polls.editor.responseOptions') }}</h4>
              </div>

              <div class="space-y-3">
                <div
                  v-for="(option, index) in form.options"
                  :key="index"
                  class="flex items-center gap-2"
                >
                  <span class="option-number">{{ index + 1 }}.</span>
                  <input
                    v-model="option.text"
                    type="text"
                    required
                    class="form-input flex-1"
                    :placeholder="t('polls.editor.optionPlaceholder', { index: index + 1 })"
                  />
                  <button
                    v-if="form.options.length > 2"
                    type="button"
                    @click="removeOption(index)"
                    class="btn-ghost p-2 text-red-600 hover:bg-red-50"
                  >
                    <Icon icon="mdi:delete" class="h-5 w-5" />
                  </button>
                </div>

                <button
                  v-if="form.options.length < 10"
                  type="button"
                  @click="addOption"
                  class="btn-secondary btn-sm"
                >
                  <Icon icon="mdi:plus" />
                  {{ t('polls.editor.addOption') }}
                </button>
                <p class="form-help">{{ t('polls.editor.optionsHelp') }}</p>
              </div>
            </div>

            <!-- Type et paramètres -->
            <div>
              <div class="section-header">
                <Icon icon="mdi:cog-outline" class="section-icon" />
                <h4 class="section-title">{{ t('polls.editor.typeAndSettings') }}</h4>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="form-group">
                  <label for="poll_type" class="form-label form-label-required">
                    {{ t('polls.editor.pollType') }}
                  </label>
                  <select
                    id="poll_type"
                    v-model="form.poll_type"
                    required
                    class="form-select"
                  >
                    <option value="single">{{ t('polls.editor.pollType_single') }}</option>
                    <option value="multiple">{{ t('polls.editor.pollType_multiple') }}</option>
                  </select>
                </div>

                <div class="form-group">
                  <label for="show_results" class="form-label form-label-required">
                    {{ t('polls.editor.resultsVisibility') }}
                  </label>
                  <select
                    id="show_results"
                    v-model="form.show_results"
                    required
                    class="form-select"
                  >
                    <option value="always">{{ t('polls.editor.resultsVisibility_always') }}</option>
                    <option value="after">{{ t('polls.editor.resultsVisibility_after') }}</option>
                    <option value="closed">{{ t('polls.editor.resultsVisibility_closed') }}</option>
                  </select>
                </div>
              </div>

              <div class="form-group">
                <div class="flex items-center gap-2">
                  <input
                    id="is_anonymous"
                    v-model="form.is_anonymous"
                    type="checkbox"
                    class="form-checkbox"
                  />
                  <label for="is_anonymous" class="form-label mb-0 cursor-pointer">
                    {{ t('polls.editor.anonymousPoll') }}
                  </label>
                </div>
              </div>

              <div class="form-group">
                <div class="flex items-center gap-2">
                  <input
                    id="is_active"
                    v-model="form.is_active"
                    type="checkbox"
                    class="form-checkbox"
                  />
                  <label for="is_active" class="form-label mb-0 cursor-pointer">
                    {{ t('polls.editor.activePoll') }}
                  </label>
                </div>
              </div>
            </div>

            <!-- Période et ciblage -->
            <div>
              <div class="section-header">
                <Icon icon="mdi:calendar-range" class="section-icon" />
                <h4 class="section-title">{{ t('polls.editor.periodAndTargeting') }}</h4>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="form-group">
                  <label for="start_date" class="form-label">
                    {{ t('polls.editor.startDate') }}
                  </label>
                  <input
                    id="start_date"
                    v-model="form.start_date"
                    type="datetime-local"
                    class="form-input"
                  />
                </div>

                <div class="form-group">
                  <label for="end_date" class="form-label">
                    {{ t('polls.editor.endDate') }}
                  </label>
                  <input
                    id="end_date"
                    v-model="form.end_date"
                    type="datetime-local"
                    class="form-input"
                  />
                </div>
              </div>

              <div class="form-group">
                <label for="target_groups" class="form-label">
                  {{ t('polls.editor.targetGroups') }}
                </label>
                <select
                  id="target_groups"
                  v-model="form.target_group_ids"
                  multiple
                  class="form-select"
                  size="5"
                >
                  <option
                    v-for="group in availableGroups"
                    :key="group.id"
                    :value="group.id"
                  >
                    {{ group.name }}
                  </option>
                </select>
                <p class="form-help">
                  {{ t('polls.editor.targetGroupsHelp') }}
                </p>
              </div>
            </div>

            <!-- Liens optionnels -->
            <div v-if="showAdvanced">
              <div class="section-header">
                <Icon icon="mdi:link-variant" class="section-icon" />
                <h4 class="section-title">{{ t('polls.editor.optionalLinks') }}</h4>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="form-group">
                  <label for="news_id" class="form-label">
                    {{ t('polls.editor.linkToArticle') }}
                  </label>
                  <input
                    id="news_id"
                    v-model.number="form.news_id"
                    type="number"
                    class="form-input"
                    :placeholder="t('polls.editor.linkToArticlePlaceholder')"
                  />
                </div>

                <div class="form-group">
                  <label for="announcement_id" class="form-label">
                    {{ t('polls.editor.linkToAnnouncement') }}
                  </label>
                  <input
                    id="announcement_id"
                    v-model.number="form.announcement_id"
                    type="number"
                    class="form-input"
                    :placeholder="t('polls.editor.linkToAnnouncementPlaceholder')"
                  />
                </div>
              </div>
            </div>

            <button
              type="button"
              @click="showAdvanced = !showAdvanced"
              class="btn-ghost btn-sm"
            >
              <Icon :icon="showAdvanced ? 'mdi:chevron-up' : 'mdi:chevron-down'" />
              {{ showAdvanced ? t('polls.editor.hideAdvanced') : t('polls.editor.showAdvanced') }} {{ t('polls.editor.advancedOptions') }}
            </button>
          </div>

          <!-- Footer -->
          <div class="modal-footer">
            <button
              type="button"
              @click="closeModal"
              class="btn-secondary"
            >
              {{ t('polls.editor.cancel') }}
            </button>
            <button
              type="submit"
              :disabled="loading"
              class="btn-primary"
            >
              <Icon v-if="loading" icon="mdi:loading" class="spin" />
              <Icon v-else icon="mdi:check" />
              {{ isEdit ? t('polls.editor.save') : t('polls.editor.create') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { pollsService } from '@/services/api'
import { useAuthStore } from '@/stores/auth'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  poll: {
    type: Object,
    default: null
  },
  groups: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['close', 'success'])

const authStore = useAuthStore()
const { t } = useI18n()

const form = ref({
  title: '',
  description: '',
  poll_type: 'single',
  is_anonymous: false,
  is_active: true,
  show_results: 'after',
  start_date: null,
  end_date: null,
  news_id: null,
  announcement_id: null,
  target_group_ids: [],
  options: [
    { text: '', order: 0 },
    { text: '', order: 1 }
  ]
})

const errors = ref({})
const loading = ref(false)
const showAdvanced = ref(false)

const isEdit = computed(() => !!props.poll)

// Groupes disponibles selon le rôle
const availableGroups = computed(() => {
  if (authStore.user.role === 'admin') {
    return props.groups
  } else if (authStore.user.role === 'group_admin') {
    // Filtrer uniquement les groupes administrés
    const managedGroupIds = authStore.user.managed_group_ids || []
    return props.groups.filter(g => managedGroupIds.includes(g.id))
  }
  return props.groups
})

// Charger les données du sondage en édition
watch(() => props.poll, (newPoll) => {
  if (newPoll) {
    form.value = {
      title: newPoll.title || '',
      description: newPoll.description || '',
      poll_type: newPoll.poll_type || 'single',
      is_anonymous: newPoll.is_anonymous || false,
      is_active: newPoll.is_active !== undefined ? newPoll.is_active : true,
      show_results: newPoll.show_results || 'after',
      start_date: newPoll.start_date ? formatDateTimeLocal(newPoll.start_date) : null,
      end_date: newPoll.end_date ? formatDateTimeLocal(newPoll.end_date) : null,
      news_id: newPoll.news_id || null,
      announcement_id: newPoll.announcement_id || null,
      target_group_ids: newPoll.target_groups?.map(g => g.id) || [],
      options: newPoll.options?.map((opt, index) => ({
        id: opt.id,
        text: opt.text,
        order: index
      })) || [
        { text: '', order: 0 },
        { text: '', order: 1 }
      ]
    }

    if (newPoll.news_id || newPoll.announcement_id) {
      showAdvanced.value = true
    }
  }
}, { immediate: true })

// Ajouter une option
const addOption = () => {
  if (form.value.options.length < 10) {
    form.value.options.push({
      text: '',
      order: form.value.options.length
    })
  }
}

// Supprimer une option
const removeOption = (index) => {
  if (form.value.options.length > 2) {
    form.value.options.splice(index, 1)
    // Réorganiser les order
    form.value.options.forEach((opt, idx) => {
      opt.order = idx
    })
  }
}

// Formater la date pour datetime-local input
const formatDateTimeLocal = (dateString) => {
  if (!dateString) return null
  const date = new Date(dateString)
  const offset = date.getTimezoneOffset()
  const localDate = new Date(date.getTime() - (offset * 60 * 1000))
  return localDate.toISOString().slice(0, 16)
}

// Convertir datetime-local en ISO string
const toISOString = (dateTimeLocal) => {
  if (!dateTimeLocal) return null
  return new Date(dateTimeLocal).toISOString()
}

// Soumettre le formulaire
const handleSubmit = async () => {
  errors.value = {}

  // Validation
  if (!form.value.title.trim()) {
    errors.value.title = t('polls.editor.validation.titleRequired')
    return
  }

  const validOptions = form.value.options.filter(opt => opt.text.trim())
  if (validOptions.length < 2) {
    if (window.$toast) {
      window.$toast.error(t('polls.editor.validation.minOptionsRequired'))
    }
    return
  }

  try {
    loading.value = true

    const payload = {
      title: form.value.title,
      description: form.value.description || '',
      poll_type: form.value.poll_type,
      is_anonymous: form.value.is_anonymous,
      is_active: form.value.is_active,
      show_results: form.value.show_results,
      start_date: toISOString(form.value.start_date),
      end_date: toISOString(form.value.end_date),
      news_id: form.value.news_id || null,
      announcement_id: form.value.announcement_id || null,
      target_group_ids: form.value.target_group_ids,
      options: validOptions.map((opt, index) => ({
        text: opt.text.trim(),
        order: index
      }))
    }

    let result
    const userRole = authStore.user.role

    if (isEdit.value) {
      // Mise à jour
      if (userRole === 'admin') {
        result = await pollsService.updatePoll(props.poll.id, payload)
      } else if (userRole === 'editor') {
        result = await pollsService.updatePollAsEditor(props.poll.id, payload)
      } else if (userRole === 'group_admin') {
        result = await pollsService.updatePollAsGroupAdmin(props.poll.id, payload)
      }
    } else {
      // Création
      if (userRole === 'admin') {
        result = await pollsService.createPoll(payload)
      } else if (userRole === 'editor') {
        result = await pollsService.createPollAsEditor(payload)
      } else if (userRole === 'group_admin') {
        result = await pollsService.createPollAsGroupAdmin(payload)
      }
    }

    if (window.$toast) {
      window.$toast.success(isEdit.value ? t('polls.editor.messages.updateSuccess') : t('polls.editor.messages.createSuccess'))
    }

    emit('success', result)
    closeModal()
  } catch (error) {
    console.error(t('polls.editor.messages.saveError'), error)
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('polls.editor.messages.saveError'))
    }
  } finally {
    loading.value = false
  }
}

// Fermer le modal
const closeModal = () => {
  // Réinitialiser le formulaire
  form.value = {
    title: '',
    description: '',
    poll_type: 'single',
    is_anonymous: false,
    is_active: true,
    show_results: 'after',
    start_date: null,
    end_date: null,
    news_id: null,
    announcement_id: null,
    target_group_ids: [],
    options: [
      { text: '', order: 0 },
      { text: '', order: 1 }
    ]
  }
  errors.value = {}
  showAdvanced.value = false
  emit('close')
}
</script>

<style scoped>

.option-number {
  font-weight: 600;
  color: #6b7280;
  min-width: 1.5rem;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
