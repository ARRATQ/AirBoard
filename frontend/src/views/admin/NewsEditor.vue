<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ pageTitle }}</template>
      <template #description>{{ pageSubtitle }}</template>
      <template #actions>
        <router-link :to="LIST_ROUTE" class="btn btn-secondary">
          <Icon icon="lucide:arrow-left" class="rtl-flip h-4 w-4 me-2" />
          {{ $t('groupAdmin.newsEditor.backToList') }}
        </router-link>
      </template>
    </PageHeader>

    <form class="space-y-6" @submit.prevent="saveNews">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Contenu principal -->
        <div class="lg:col-span-2 space-y-6">
          <div class="card">
            <div class="mb-3">
              <label for="news-title" class="flex text-sm font-semibold text-gray-900 dark:text-white mb-2 items-center gap-2">
                <Icon icon="lucide:heading" class="h-5 w-5 text-primary-500" />
                {{ $t('groupAdmin.newsEditor.titleLabel') }} *
              </label>
              <input
                id="news-title"
                v-model="form.title"
                type="text"
                required
                :placeholder="$t('groupAdmin.newsEditor.titlePlaceholder')"
                class="w-full px-3 py-1.5 text-lg font-medium border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white transition-colors"
              />
            </div>
          </div>

          <div class="card">
            <div class="mb-3">
              <label for="news-summary" class="flex text-sm font-semibold text-gray-900 dark:text-white mb-2 items-center gap-2">
                <Icon icon="lucide:text-cursor-input" class="h-5 w-5 text-primary-500" />
                {{ $t('groupAdmin.newsEditor.summaryLabel') }}
              </label>
              <textarea
                id="news-summary"
                v-model="form.summary"
                rows="3"
                :maxlength="SUMMARY_MAX"
                :placeholder="$t('groupAdmin.newsEditor.summaryPlaceholder')"
                class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white resize-none transition-colors"
              ></textarea>
              <div class="flex justify-between items-center mt-2">
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ $t('groupAdmin.newsEditor.summaryHelp') }}
                </p>
                <p class="text-xs font-medium" :class="form.summary.length >= SUMMARY_MAX ? 'text-red-500' : 'text-gray-500 dark:text-gray-400'">
                  {{ form.summary.length }}/{{ SUMMARY_MAX }}
                </p>
              </div>
            </div>
          </div>

          <div class="card">
            <div class="mb-3">
              <p class="flex text-sm font-semibold text-gray-900 dark:text-white mb-3 items-center gap-2">
                <Icon icon="lucide:text" class="h-5 w-5 text-primary-500" />
                {{ $t('groupAdmin.newsEditor.contentLabel') }} *
              </p>
              <div class="border border-gray-300 dark:border-gray-600 rounded-lg">
                <!-- Limitée à ses groupes : médias envoyés par l'API group-admin -->
                <RichTextEditor
                  v-model="form.content"
                  :placeholder="$t('groupAdmin.newsEditor.contentPlaceholder')"
                  :as-group-admin="!isGlobal"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Colonne latérale -->
        <div class="space-y-6">
          <!-- Publication -->
          <div class="card bg-gradient-to-br from-white to-gray-50 dark:from-gray-800 dark:to-gray-900">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <Icon icon="lucide:upload" class="h-5 w-5 text-primary-500" />
              {{ $t('groupAdmin.newsEditor.publishSection') }}
            </h3>

            <div class="space-y-4">
              <div class="bg-white dark:bg-gray-800 rounded-lg p-3 border border-gray-300 dark:border-gray-600">
                <label class="flex items-center cursor-pointer">
                  <input
                    v-model="form.is_published"
                    type="checkbox"
                    class="form-checkbox h-5 w-5 text-primary-600 rounded focus:ring-2 focus:ring-primary-500"
                  />
                  <span class="ms-3 text-sm font-medium text-gray-900 dark:text-white">
                    {{ $t('groupAdmin.newsEditor.publishImmediately') }}
                  </span>
                </label>
              </div>

              <div v-if="form.is_published">
                <label for="news-published-at" class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="lucide:calendar-clock" class="h-4 w-4" />
                  {{ $t('groupAdmin.newsEditor.publishDate') }}
                </label>
                <input
                  id="news-published-at"
                  v-model="form.published_at"
                  type="datetime-local"
                  class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm"
                />
              </div>

              <div>
                <label for="news-expires-at" class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="lucide:calendar-x" class="h-4 w-4" />
                  {{ $t('groupAdmin.newsEditor.expiryDate') }}
                  <span class="text-xs text-gray-500 dark:text-gray-400 ms-1">{{ $t('groupAdmin.newsEditor.expiryDateOptional') }}</span>
                </label>
                <input
                  id="news-expires-at"
                  v-model="form.expires_at"
                  type="datetime-local"
                  class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm"
                />
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1.5 flex items-center gap-1">
                  <Icon icon="lucide:info" class="h-3 w-3" />
                  {{ $t('groupAdmin.newsEditor.expiryDateHelp') }}
                </p>
              </div>
            </div>
          </div>

          <!-- Groupes cibles : gestion limitée à ses groupes uniquement -->
          <NewsTargetGroupsCard
            v-if="!isGlobal"
            v-model="form.target_group_ids"
            :groups="managedGroups"
            :locked-groups="lockedGroups"
          />

          <!-- Classification -->
          <div class="card">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <Icon icon="lucide:tags" class="h-5 w-5 text-primary-500" />
              {{ $t('groupAdmin.newsEditor.classification') }}
            </h3>

            <div class="space-y-4">
              <div>
                <label for="news-category" class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="lucide:folder" class="h-4 w-4" />
                  {{ $t('groupAdmin.newsEditor.categoryLabel') }}
                </label>
                <div class="flex gap-2">
                  <select
                    id="news-category"
                    v-model="form.category_id"
                    class="flex-1 min-w-0 px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm"
                  >
                    <option :value="null">{{ $t('groupAdmin.newsEditor.noCategory') }}</option>
                    <option v-for="cat in categories" :key="cat.id" :value="cat.id">
                      {{ cat.name }}
                    </option>
                  </select>
                  <!-- Création de catégorie : endpoint réservé à l'admin global -->
                  <button
                    v-if="isAdmin"
                    type="button"
                    class="flex-shrink-0 px-3 py-2 bg-primary-500 text-white rounded-lg hover:bg-primary-600 transition-colors flex items-center gap-1"
                    :title="$t('events.editor.addCategory')"
                    :aria-label="$t('events.editor.addCategory')"
                    @click="showCategoryModal = true"
                  >
                    <Icon icon="lucide:plus" class="h-4 w-4" />
                  </button>
                </div>
              </div>

              <div>
                <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="lucide:shapes" class="h-4 w-4" />
                  {{ $t('groupAdmin.newsEditor.typeLabel') }}
                </p>
                <TypeSelect v-model="form.type" :types="typesStore.types" />
              </div>

              <div>
                <label for="news-priority" class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="lucide:flag" class="h-4 w-4" />
                  {{ $t('groupAdmin.newsEditor.priorityLabel') }}
                </label>
                <select
                  id="news-priority"
                  v-model="form.priority"
                  class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm"
                >
                  <option v-for="priority in PRIORITIES" :key="priority" :value="priority">
                    {{ $t(`groupAdmin.newsEditor.priorities.${priority}`) }}
                  </option>
                </select>
              </div>
            </div>
          </div>

          <NewsTagsCard v-model="form.tag_ids" :tags="tags" :create-tag="createTag" />

          <!-- Actions -->
          <div class="card bg-gradient-to-br from-primary-50 to-primary-100 dark:from-primary-900/20 dark:to-primary-800/20 border-2 border-primary-200 dark:border-primary-800">
            <div class="space-y-3">
              <button
                type="submit"
                :disabled="isSaving || missingTargetGroup"
                class="w-full px-4 py-1.5 bg-primary-600 hover:bg-primary-700 disabled:bg-gray-400 text-white font-semibold rounded-lg transition-all transform hover:scale-[1.02] active:scale-[0.98] disabled:scale-100 shadow-lg hover:shadow-xl flex items-center justify-center gap-2"
              >
                <Icon v-if="isSaving" icon="lucide:loader-circle" class="h-5 w-5 animate-spin" />
                <Icon v-else icon="lucide:save" class="h-5 w-5" />
                <span>{{ submitLabel }}</span>
              </button>

              <p v-if="missingTargetGroup" class="text-xs text-center text-red-500 dark:text-red-400">
                {{ $t('groupAdmin.newsEditor.validation.selectGroupRequired') }}
              </p>

              <router-link
                :to="LIST_ROUTE"
                class="w-full px-4 py-1.5 bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 font-medium rounded-lg transition-all flex items-center justify-center gap-2"
              >
                <Icon icon="lucide:x" class="h-5 w-5" />
                <span>{{ $t('groupAdmin.newsEditor.actions.cancel') }}</span>
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </form>

    <NewsCategoryModal
      v-if="isAdmin"
      :open="showCategoryModal"
      @close="showCategoryModal = false"
      @created="onCategoryCreated"
    />
  </div>
</template>

<script setup>
import PageHeader from '@/components/ui/PageHeader.vue'
import { toDateTimeInput, fromDateTimeInput } from '@/utils/dateInput'
import { isRichTextEmpty } from '@/utils/richText'
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useNewsTypesStore } from '@/stores/newsTypes'
import { newsService, groupAdminService } from '@/services/api'
import { NEWS, ACCESS_PERMISSION } from '@/config/access'
import { useManagementScope } from '@/composables/useManagementScope'
import RichTextEditor from '@/components/news/RichTextEditor.vue'
import TypeSelect from '@/components/news/TypeSelect.vue'
import NewsTargetGroupsCard from '@/components/admin/news/NewsTargetGroupsCard.vue'
import NewsTagsCard from '@/components/admin/news/NewsTagsCard.vue'
import NewsCategoryModal from '@/components/admin/news/NewsCategoryModal.vue'

const LIST_ROUTE = '/admin/news'
const SUMMARY_MAX = 300
const PRIORITIES = ['normal', 'important', 'urgent']

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
const typesStore = useNewsTypesStore()

// Sur tout le site : API editor ; limitée à ses groupes : API group-admin + ciblage obligatoire
const { isGlobal, isAdmin } = useManagementScope(NEWS)

const isEditMode = computed(() => Boolean(route.params.slug && route.params.slug !== 'new'))
const isSaving = ref(false)
const categories = ref([])
const tags = ref([])
const managedGroups = ref([])
// Groupes ciblés par l'article hors de ses groupes (lecture seule)
const targetGroups = ref([])
const lockedGroups = computed(() =>
  targetGroups.value.filter((group) => !managedGroups.value.some((g) => g.id === group.id))
)
const showCategoryModal = ref(false)

const form = ref({
  title: '',
  summary: '',
  content: '',
  type: 'article',
  priority: 'normal',
  is_published: false,
  published_at: null,
  expires_at: null,
  category_id: null,
  tag_ids: [],
  target_group_ids: []
})

const missingTargetGroup = computed(() => !isGlobal.value && form.value.target_group_ids.length === 0)

const pageTitle = computed(() => {
  if (isEditMode.value) return t('groupAdmin.newsEditor.title.edit')
  return isGlobal.value ? t('groupAdmin.newsManagement.newArticle') : t('groupAdmin.newsEditor.title.new')
})

const pageSubtitle = computed(() => {
  if (isGlobal.value) {
    return isEditMode.value ? t('adminShared.news.editorSubtitleEdit') : t('adminShared.news.editorSubtitleNew')
  }
  return isEditMode.value ? t('groupAdmin.newsEditor.subtitle.edit') : t('groupAdmin.newsEditor.subtitle.new')
})

const submitLabel = computed(() => {
  if (isSaving.value) return t('groupAdmin.newsEditor.actions.saving')
  return isEditMode.value ? t('groupAdmin.newsEditor.actions.update') : t('groupAdmin.newsEditor.actions.create')
})

// datetime-local attend une heure locale « AAAA-MM-JJTHH:mm »
const toInputDate = (value) => toDateTimeInput(value, null)
const toApiDate = fromDateTimeInput

// Le contenu Tiptap est stocké en JSON sérialisé
const parseContent = (raw) => {
  const content = raw || ''
  if (typeof content !== 'string' || !content.trim().startsWith('{')) return content
  try {
    return JSON.parse(content)
  } catch {
    return content
  }
}

const onCategoryCreated = (category) => {
  categories.value = [...categories.value, category]
  form.value = { ...form.value, category_id: category.id }
}

// Création d'un tag : réutilise un tag existant du même nom, sinon le crée via l'API du périmètre
const createTag = async ({ name, color }) => {
  const existing = tags.value.find((tag) => tag.name.toLowerCase() === name.toLowerCase())
  if (existing) {
    if (!form.value.tag_ids.includes(existing.id)) {
      form.value = { ...form.value, tag_ids: [...form.value.tag_ids, existing.id] }
    }
    appStore.showInfo(t('adminShared.news.tagExists'))
    return true
  }

  try {
    const payload = { name, color }
    const tag = isGlobal.value ? await newsService.createTag(payload) : await groupAdminService.createTag(payload)
    tags.value = [...tags.value, tag]
    form.value = { ...form.value, tag_ids: [...form.value.tag_ids, tag.id] }
    appStore.showSuccess(t('adminShared.news.tagCreated', { name }))
    return true
  } catch (error) {
    console.error('Erreur à la création du tag :', error)
    appStore.showError(error.response?.data?.error || t('admin.tagSaveError'))
    return false
  }
}

const loadManagedGroups = async () => {
  try {
    const data = await groupAdminService.getManagedGroups(ACCESS_PERMISSION[NEWS])
    managedGroups.value = Array.isArray(data) ? data : (data?.groups || data?.data || [])
  } catch (error) {
    console.error('Erreur au chargement de ses groupes :', error)
    managedGroups.value = []
  }
}

const loadNews = async () => {
  if (!isEditMode.value) return

  try {
    const news = await newsService.getNewsBySlug(route.params.slug)
    form.value = {
      title: news.title,
      summary: news.summary || '',
      content: parseContent(news.content),
      type: news.type,
      priority: news.priority,
      is_published: news.is_published,
      published_at: toInputDate(news.published_at),
      expires_at: toInputDate(news.expires_at),
      category_id: news.category_id,
      tag_ids: news.tags ? news.tags.map((tag) => tag.id) : [],
      target_group_ids: news.target_groups ? news.target_groups.map((group) => group.id) : []
    }
    targetGroups.value = news.target_groups || []
  } catch (error) {
    console.error('Erreur au chargement de l\'actualité :', error)
    appStore.showError(t('groupAdmin.newsEditor.messages.loadError'))
    router.push(LIST_ROUTE)
  }
}

const loadMetadata = async () => {
  try {
    const [categoryList, tagList] = await Promise.all([
      newsService.getCategories(),
      newsService.getTags()
    ])
    categories.value = categoryList || []
    tags.value = tagList || []
  } catch (error) {
    console.error('Erreur au chargement des catégories et tags :', error)
  }
}

// Charge utile envoyée à l'API. Les groupes cibles ne sont transmis qu'avec une gestion limitée à ses groupes :
// sans le champ, le backend conserve le ciblage existant (aucun écrasement côté admin/éditeur).
const buildPayload = () => {
  const { target_group_ids: targetGroupIds, ...rest } = form.value
  const payload = {
    ...rest,
    content: typeof rest.content === 'object' ? JSON.stringify(rest.content) : rest.content,
    published_at: toApiDate(rest.published_at),
    expires_at: toApiDate(rest.expires_at)
  }
  return isGlobal.value ? payload : { ...payload, target_group_ids: targetGroupIds }
}

const saveNews = async () => {
  if (missingTargetGroup.value) {
    appStore.showError(t('groupAdmin.newsEditor.messages.selectGroupError'))
    return
  }
  if (isRichTextEmpty(form.value.content)) {
    appStore.showError(t('common.validation.contentRequired'))
    return
  }

  isSaving.value = true
  try {
    const data = buildPayload()
    const service = isGlobal.value ? newsService : groupAdminService
    if (isEditMode.value) {
      await service.updateNews(route.params.slug, data)
      appStore.showSuccess(t('groupAdmin.newsEditor.messages.updateSuccess'))
    } else {
      await service.createNews(data)
      appStore.showSuccess(t('groupAdmin.newsEditor.messages.createSuccess'))
    }
    router.push(LIST_ROUTE)
  } catch (error) {
    console.error('Erreur à l\'enregistrement de l\'actualité :', error)
    appStore.showError(t('groupAdmin.newsEditor.messages.saveError'))
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  typesStore.fetchTypes()
  loadMetadata()
  if (!isGlobal.value) loadManagedGroups()
  loadNews()
})
</script>
