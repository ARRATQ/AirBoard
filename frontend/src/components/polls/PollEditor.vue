<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-container">
      <div
        v-focus-trap
        class="modal-panel sm:max-w-6xl sm:w-full"
        role="dialog"
        aria-modal="true"
        aria-labelledby="poll-editor-title"
      >
        <form class="modal-form" @submit.prevent="handleSubmit">
          <!-- Header -->
          <div class="modal-header">
            <div class="flex items-center justify-between">
              <div>
                <h3 id="poll-editor-title" class="modal-title">
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
                :aria-label="t('common.close')"
              >
                <Icon icon="lucide:x" class="h-5 w-5" aria-hidden="true" />
              </button>
            </div>
          </div>

          <!-- Content -->
          <!-- Content : contenu du sondage à gauche, réglages à droite -->
          <div class="modal-content grid grid-cols-1 lg:grid-cols-[3fr_2fr] gap-x-8 gap-y-6">
            <div class="space-y-6">
            <!-- Informations de base -->
            <div>
              <div class="section-header">
                <Icon icon="lucide:info" class="section-icon" />
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
                <Icon icon="lucide:list-checks" class="section-icon" />
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
                    :aria-label="t('polls.editor.optionPlaceholder', { index: index + 1 })"
                  />
                  <button
                    type="button"
                    :aria-pressed="!!option.is_other"
                    @click="toggleOtherOption(index)"
                    :class="[
                      'px-3 py-1.5 rounded-lg text-xs font-medium transition-colors whitespace-nowrap',
                      option.is_other
                        ? 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'
                    ]"
                  >
                    <Icon :icon="option.is_other ? 'lucide:circle-check' : 'lucide:circle-help'" class="inline h-4 w-4 me-1" aria-hidden="true" />
                    {{ option.is_other ? t('polls.editor.optionIsOther') : t('polls.editor.markAsOther') }}
                  </button>
                  <button
                    v-if="form.options.length > 2"
                    type="button"
                    @click="removeOption(index)"
                    class="btn-ghost p-2 text-red-600 hover:bg-red-50"
                    :aria-label="t('polls.editor.removeOption', { index: index + 1 })"
                    :title="t('polls.editor.removeOption', { index: index + 1 })"
                  >
                    <Icon icon="lucide:trash-2" class="h-5 w-5" aria-hidden="true" />
                  </button>
                </div>

                <button
                  v-if="form.options.length < 10"
                  type="button"
                  @click="addOption"
                  class="poll-add-option-btn"
                >
                  <Icon icon="lucide:plus" />
                  {{ t('polls.editor.addOption') }}
                </button>
                <p class="form-help">{{ t('polls.editor.otherHelp') }}</p>
                <p class="form-help">{{ t('polls.editor.optionsHelp') }}</p>
              </div>
            </div>

            </div>

            <div class="space-y-6 lg:border-s lg:ps-8 border-gray-200 dark:border-gray-700">
            <!-- Type et paramètres -->
            <div>
              <div class="section-header">
                <Icon icon="lucide:settings" class="section-icon" />
                <h4 class="section-title">{{ t('polls.editor.typeAndSettings') }}</h4>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
                <Icon icon="lucide:calendar-range" class="section-icon" />
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
                <label id="poll-editor-groups" for="poll-editor-groups-input" class="form-label">
                  {{ isScoped ? t('polls.editor.targetGroupsScoped') : t('polls.editor.targetGroups') }}
                </label>
                <p v-if="availableGroups.length === 0" class="text-gray-500 dark:text-gray-400 text-sm italic">
                  {{ t('polls.editor.noGroupsAvailable') }}
                </p>
                <ComboSelect
                  v-else
                  v-model="form.target_group_ids"
                  input-id="poll-editor-groups-input"
                  kind="group"
                  :options="groupOptions"
                  :labels="lockedGroupOptions"
                  :locked-ids="lockedGroupIds"
                  :label="isScoped ? t('polls.editor.targetGroupsScoped') : t('polls.editor.targetGroups')"
                  :placeholder="t('polls.editor.targetGroupsPlaceholder')"
                />
                <p class="form-help">
                  {{ isScoped ? t('polls.editor.targetGroupsScopedHelp') : t('polls.editor.targetGroupsHelp') }}
                </p>
                <p v-if="errors.target_group_ids" class="text-sm text-red-500 dark:text-red-400" role="alert">
                  {{ errors.target_group_ids }}
                </p>
              </div>
            </div>

            <!-- Liens optionnels -->
            <div v-if="showAdvanced">
              <div class="section-header">
                <Icon icon="lucide:link" class="section-icon" />
                <h4 class="section-title">{{ t('polls.editor.optionalLinks') }}</h4>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="form-group">
                  <label for="news_id" class="form-label">
                    {{ t('polls.editor.linkToArticle') }}
                  </label>
                  <select
                    id="news_id"
                    v-model="form.news_id"
                    class="form-select"
                  >
                    <option :value="null">{{ t('polls.editor.noArticleSelected') }}</option>
                    <option
                      v-for="article in availableNews"
                      :key="article.id"
                      :value="article.id"
                    >
                      {{ article.title }}
                    </option>
                  </select>
                  <p v-if="loadingNews" class="form-help text-gray-500">
                    <Icon icon="lucide:loader-circle" class="inline animate-spin" aria-hidden="true" /> {{ t('common.loading') }}
                  </p>
                </div>

                <div class="form-group">
                  <label for="announcement_id" class="form-label">
                    {{ t('polls.editor.linkToAnnouncement') }}
                  </label>
                  <select
                    id="announcement_id"
                    v-model="form.announcement_id"
                    class="form-select"
                  >
                    <option :value="null">{{ t('polls.editor.noAnnouncementSelected') }}</option>
                    <option
                      v-for="announcement in availableAnnouncements"
                      :key="announcement.id"
                      :value="announcement.id"
                    >
                      {{ announcement.title }}
                    </option>
                  </select>
                  <p v-if="loadingAnnouncements" class="form-help text-gray-500">
                    <Icon icon="lucide:loader-circle" class="inline animate-spin" aria-hidden="true" /> {{ t('common.loading') }}
                  </p>
                </div>
              </div>
            </div>

            <button
              type="button"
              :aria-expanded="showAdvanced"
              @click="showAdvanced = !showAdvanced"
              class="btn-ghost btn-sm"
            >
              <Icon :icon="showAdvanced ? 'lucide:chevron-up' : 'lucide:chevron-down'" />
              {{ showAdvanced ? t('polls.editor.hideAdvanced') : t('polls.editor.showAdvanced') }} {{ t('polls.editor.advancedOptions') }}
            </button>
            </div>
          </div>

          <!-- Footer -->
          <div class="modal-footer-bar">
            <span class="modal-footer-hint">{{ $t('common.requiredFieldsHint') }}</span>
            <div class="modal-footer-actions">
              <button type="button" @click="closeModal" class="poll-cancel-btn">
                {{ t('polls.editor.cancel') }}
              </button>
              <button type="submit" :disabled="loading" class="poll-submit-btn">
                <Icon v-if="loading" icon="lucide:loader-circle" class="spin" />
                <Icon v-else icon="lucide:check" />
                {{ isEdit ? t('polls.editor.save') : t('polls.editor.create') }}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
import { isEndBeforeStart } from '@/utils/dateInput'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { pollsService, newsService, announcementsService } from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { getRoles, canAccess, getScope, POLLS } from '@/config/access'
import ComboSelect from '@/components/ui/ComboSelect.vue'
import { groupOption } from '@/utils/comboFilter'

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
    { text: '', is_other: false, order: 0 },
    { text: '', is_other: false, order: 1 }
  ]
})

const errors = ref({})
const loading = ref(false)
const showAdvanced = ref(false)

// Listes pour les liens optionnels
const availableNews = ref([])
const availableAnnouncements = ref([])
const loadingNews = ref(false)
const loadingAnnouncements = ref(false)

const isEdit = computed(() => !!props.poll)
// Gestion limitée à ses groupes : au moins un groupe cible (pas de sondage global)
const isScoped = computed(() => getScope(POLLS, getRoles(authStore.user)) === 'group')

// Groupes disponibles : limités à ses groupes, ils sont déjà filtrés par le parent
const availableGroups = computed(() => {
  // Les groupes passés en props sont déjà les bons selon le rôle
  // (managedGroups pour group-admin, tous les groupes pour admin)
  return props.groups || []
})

// Charger les articles disponibles
const loadNews = async () => {
  try {
    loadingNews.value = true
    const response = await newsService.getNews({ page_size: 100, is_published: true })
    availableNews.value = response.news || []
  } catch (error) {
    console.error('Erreur lors du chargement des articles:', error)
    availableNews.value = []
  } finally {
    loadingNews.value = false
  }
}

// Charger les annonces disponibles
const loadAnnouncements = async () => {
  try {
    loadingAnnouncements.value = true
    // Utiliser getActiveAnnouncements pour les utilisateurs normaux
    const response = await announcementsService.getActiveAnnouncements()
    availableAnnouncements.value = response || []
  } catch (error) {
    console.error('Erreur lors du chargement des annonces:', error)
    availableAnnouncements.value = []
  } finally {
    loadingAnnouncements.value = false
  }
}

// Charger les articles et annonces quand on ouvre les options avancées
watch(showAdvanced, (newValue) => {
  if (newValue) {
    if (availableNews.value.length === 0) {
      loadNews()
    }
    if (availableAnnouncements.value.length === 0) {
      loadAnnouncements()
    }
  }
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
        is_other: !!opt.is_other,
        order: index
      })) || [
        { text: '', is_other: false, order: 0 },
        { text: '', is_other: false, order: 1 }
      ]
    }

    if (newPoll.news_id || newPoll.announcement_id) {
      showAdvanced.value = true
      // Charger les listes si elles ne sont pas encore chargées
      if (availableNews.value.length === 0) loadNews()
      if (availableAnnouncements.value.length === 0) loadAnnouncements()
    }
  }
}, { immediate: true })

// Ajouter une option
const addOption = () => {
  if (form.value.options.length < 10) {
    form.value.options.push({
      text: '',
      is_other: false,
      order: form.value.options.length
    })
  }
}

const toggleOtherOption = (index) => {
  const target = form.value.options[index]
  if (!target) return

  if (target.is_other) {
    target.is_other = false
    return
  }

  form.value.options.forEach((opt, idx) => {
    opt.is_other = idx === index
  })
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

// Groupes proposés, et groupes déjà ciblés hors de ceux-là : affichés, mais non retirables
const groupOptions = computed(() => availableGroups.value.map((group) => groupOption(group)))
const lockedGroupOptions = computed(() => {
  const available = new Set(availableGroups.value.map((g) => g.id))
  return (props.poll?.target_groups || []).filter((g) => !available.has(g.id)).map((group) => groupOption(group))
})
const lockedGroupIds = computed(() => lockedGroupOptions.value.map((g) => g.id))

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

  if (isEndBeforeStart(form.value.start_date, form.value.end_date)) {
    if (window.$toast) {
      window.$toast.error(t('common.validation.endBeforeStart'))
    }
    return
  }

  if (isScoped.value && form.value.target_group_ids.length === 0) {
    errors.value.target_group_ids = t('polls.editor.validation.groupRequired')
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
      // L'id d'une option existante la fait modifier sur place : ses votes sont conservés
      options: validOptions.map((opt, index) => ({
        ...(opt.id ? { id: opt.id } : {}),
        text: opt.text.trim(),
        is_other: !!opt.is_other,
        order: index
      }))
    }

    // Même règle que le menu et le routeur (config/access.js) ; sans droit, on s'arrête
    // au lieu d'annoncer un faux succès. L'admin passe par l'API admin, les autres par
    // l'API group-admin, qui applique la portée de leur permission.
    const roles = getRoles(authStore.user)
    if (!canAccess(POLLS, roles)) {
      throw new Error('Droits insuffisants pour enregistrer un sondage')
    }

    let result
    if (isEdit.value) {
      result = roles.isAdmin
        ? await pollsService.updatePoll(props.poll.id, payload)
        : await pollsService.updatePollAsGroupAdmin(props.poll.id, payload)
    } else {
      result = roles.isAdmin
        ? await pollsService.createPoll(payload)
        : await pollsService.createPollAsGroupAdmin(payload)
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
      { text: '', is_other: false, order: 0 },
      { text: '', is_other: false, order: 1 }
    ]
  }
  errors.value = {}
  showAdvanced.value = false
  emit('close')
}

// Échap ferme la modale
useEscapeClose(() => props.show, () => closeModal())
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

.poll-submit-btn {
  @apply inline-flex items-center gap-2 px-4 py-2 rounded-lg font-semibold text-white;
  @apply bg-gradient-to-r from-amber-500 to-orange-500 shadow-sm;
  @apply hover:from-amber-600 hover:to-orange-600 hover:shadow-md;
  @apply focus:outline-none focus:ring-2 focus:ring-amber-400 focus:ring-offset-2;
  @apply dark:focus:ring-offset-gray-800 transition-all duration-200;
  @apply disabled:opacity-60 disabled:cursor-not-allowed;
}

.poll-cancel-btn {
  @apply inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold;
  @apply focus:outline-none focus:ring-2 focus:ring-offset-2 transition-colors duration-200;
  background-color: var(--bg-card);
  border: 1px solid var(--border);
  color: var(--text-primary);
}
.poll-cancel-btn:hover {
  background-color: var(--hover);
}

.poll-add-option-btn {
  @apply inline-flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm font-semibold;
  @apply border border-amber-200 text-amber-700 bg-amber-50 hover:bg-amber-100 hover:border-amber-300;
  @apply dark:border-amber-800 dark:text-amber-300 dark:bg-amber-900/20 dark:hover:bg-amber-900/35;
  @apply focus:outline-none focus:ring-2 focus:ring-amber-300 focus:ring-offset-2 dark:focus:ring-offset-gray-800;
  @apply transition-colors duration-200;
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
