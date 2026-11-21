<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ $t('media.title') }}</template>
      <template #description>{{ $t('media.subtitle') }}</template>
      <template #actions>
        <BaseButton variant="primary" icon="mdi:upload" @click="openUploadModal">{{ $t('media.upload') }}</BaseButton>
      </template>
    </PageHeader>

    <!-- Chiffres clés -->
    <dl class="mm-stats">
      <div class="mm-stat">
        <dt>{{ $t('media.total_files') }}</dt>
        <dd>{{ stats.totalFiles }}</dd>
      </div>
      <div class="mm-stat">
        <dt>{{ $t('media.total_images') }}</dt>
        <dd>{{ stats.totalImages }}</dd>
      </div>
      <div class="mm-stat">
        <dt>{{ $t('media.total_documents') }}</dt>
        <dd>{{ stats.totalDocuments }}</dd>
      </div>
      <div class="mm-stat">
        <dt>{{ $t('media.total_size') }}</dt>
        <dd>{{ formatSize(stats.totalSize) }}</dd>
      </div>
    </dl>

    <!-- Filtres -->
    <div class="mm-filters">
      <label class="mm-search">
        <Icon icon="mdi:magnify" class="mm-search__icon" aria-hidden="true" />
        <input
          v-model="searchQuery"
          type="search"
          :placeholder="$t('media.searchPlaceholder')"
          :aria-label="$t('media.searchPlaceholder')"
          class="mm-search__input"
        />
      </label>
      <label>
        <span class="sr-only">{{ $t('ui.table.columns.type') }}</span>
        <select v-model="typeFilter" class="ui-select">
          <option value="">{{ $t('media.filter_all_types') }}</option>
          <option value="image">{{ $t('media.type_image') }}</option>
          <option value="video">{{ $t('media.type_video') }}</option>
          <option value="pdf">{{ $t('media.type_pdf') }}</option>
          <option value="document">{{ $t('media.type_document') }}</option>
          <option value="spreadsheet">{{ $t('media.type_spreadsheet') }}</option>
          <option value="presentation">{{ $t('media.type_presentation') }}</option>
        </select>
      </label>
    </div>

    <!-- Contenu -->
    <div v-if="loading" class="mm-loading">
      <Icon icon="mdi:loading" class="mm-spin" aria-hidden="true" />
      <span class="sr-only">{{ $t('common.loading') }}</span>
    </div>

    <ul v-else-if="filteredMedia.length > 0" class="mm-grid">
      <li v-for="media in filteredMedia" :key="media.id" class="mm-card">
        <button type="button" class="mm-card__preview" :aria-label="`${$t('common.view')} : ${media.filename}`" @click="viewMedia(media)">
          <img
            v-if="isImage(media.mime_type) && !brokenImages.has(media.id)"
            :src="getMediaUrl(media.url)"
            :alt="media.filename"
            loading="lazy"
            @error="markBroken(media.id)"
          />
          <Icon v-else :icon="getFileIcon(media.mime_type)" class="mm-card__file-icon" aria-hidden="true" />
        </button>

        <div class="mm-card__body">
          <h3 class="mm-card__name" :title="media.filename">{{ media.filename }}</h3>
          <p class="mm-card__meta">
            <span>{{ formatSize(media.file_size) }}</span>
            <span v-if="media.width && media.height">{{ media.width }} × {{ media.height }}</span>
          </p>
          <p class="mm-card__meta">{{ $t('media.uploaded_by') }} : {{ media.uploader?.username || 'N/A' }}</p>
          <p class="mm-card__meta">{{ formatDate(media.created_at) }}</p>
        </div>

        <div class="mm-card__actions">
          <button type="button" class="row-actions__btn" :title="$t('common.view')" :aria-label="`${$t('common.view')} : ${media.filename}`" @click="viewMedia(media)">
            <Icon icon="mdi:eye-outline" aria-hidden="true" />
          </button>
          <button type="button" class="row-actions__btn" :title="$t('media.copy_url')" :aria-label="`${$t('media.copy_url')} : ${media.filename}`" @click="copyUrl(media)">
            <Icon icon="mdi:content-copy" aria-hidden="true" />
          </button>
          <button type="button" class="row-actions__btn" :title="$t('common.download')" :aria-label="`${$t('common.download')} : ${media.filename}`" @click="downloadMedia(media)">
            <Icon icon="mdi:download-outline" aria-hidden="true" />
          </button>
          <button type="button" class="row-actions__btn" :title="$t('common.edit')" :aria-label="`${$t('common.edit')} : ${media.filename}`" @click="editMedia(media)">
            <Icon icon="mdi:pencil-outline" aria-hidden="true" />
          </button>
          <button type="button" class="row-actions__btn row-actions__btn--danger" :title="$t('common.delete')" :aria-label="`${$t('common.delete')} : ${media.filename}`" @click="confirmDelete(media)">
            <Icon icon="mdi:trash-can-outline" aria-hidden="true" />
          </button>
        </div>
      </li>
    </ul>

    <EmptyState
      v-else-if="mediaList.length === 0"
      icon="mdi:image-multiple"
      :title="$t('media.emptyTitle')"
      :description="$t('media.emptyText')"
    >
      <BaseButton variant="primary" size="sm" icon="mdi:upload" @click="openUploadModal">{{ $t('media.upload') }}</BaseButton>
    </EmptyState>

    <EmptyState
      v-else
      icon="mdi:magnify"
      :title="$t('media.noResultsTitle')"
      :description="$t('media.noResultsText')"
    />

    <!-- Pagination -->
    <nav v-if="totalPages > 1" class="mm-pager">
      <button type="button" class="mm-pager__btn" :disabled="currentPage === 1" :aria-label="$t('ui.table.prev')" @click="changePage(currentPage - 1)">
        <Icon icon="mdi:chevron-left" aria-hidden="true" />
      </button>
      <span class="mm-pager__page">{{ $t('common.page') }} {{ currentPage }} {{ $t('common.of') }} {{ totalPages }}</span>
      <button type="button" class="mm-pager__btn" :disabled="currentPage === totalPages" :aria-label="$t('ui.table.next')" @click="changePage(currentPage + 1)">
        <Icon icon="mdi:chevron-right" aria-hidden="true" />
      </button>
    </nav>

    <!-- Upload Modal -->
    <div v-if="showUploadModal" class="modal-overlay" @click="closeUploadModal">
      <div class="modal-container" @click.stop>
        <div class="modal-panel sm:max-w-lg sm:w-full">
          <div class="modal-header">
            <h3 class="modal-title">{{ $t('media.upload_title') }}</h3>
            <button @click="closeUploadModal" class="row-actions__btn">
              <Icon icon="mdi:close" class="h-5 w-5" />
            </button>
          </div>

          <div class="modal-content">
            <div
              @dragover.prevent="isDragging = true"
              @dragleave.prevent="isDragging = false"
              @drop.prevent="handleDrop"
              :class="['mm-drop', { 'mm-drop--on': isDragging }]"
            >
              <Icon icon="mdi:cloud-upload" class="mm-drop__icon" />
              <p class="mm-drop__text">{{ $t('media.drag_drop') }}</p>
              <p class="mm-drop__or">{{ $t('media.or') }}</p>
              <input
                ref="fileInput"
                type="file"
                @change="handleFileSelect"
                class="hidden"
                accept="image/*,video/*,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.zip,.rar,.7z,.tar,.gz,.bz2"
              />
              <button @click="$refs.fileInput.click()" class="btn btn-primary">
                {{ $t('media.select_file') }}
              </button>
            </div>

            <div v-if="selectedFile" class="mm-selected">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <Icon :icon="getFileIcon(selectedFile.type)" class="h-8 w-8 mm-icon" />
                  <div>
                    <p class="text-sm font-medium">{{ selectedFile.name }}</p>
                    <p class="mm-dim">{{ formatSize(selectedFile.size) }}</p>
                  </div>
                </div>
                <button @click="selectedFile = null" class="row-actions__btn row-actions__btn--danger">
                  <Icon icon="mdi:close" class="h-5 w-5" />
                </button>
              </div>
            </div>

            <div v-if="uploadProgress > 0 && uploadProgress < 100" class="mt-4">
              <div class="mm-bar">
                <div
                  class="mm-bar__fill"
                  :style="{ width: uploadProgress + '%' }"
                ></div>
              </div>
              <p class="mm-dim mm-center mt-2">{{ uploadProgress }}%</p>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeUploadModal" class="btn btn-secondary w-full sm:w-auto">
              {{ $t('common.cancel') }}
            </button>
            <button
              @click="uploadFile"
              :disabled="!selectedFile || uploading"
              class="btn btn-primary w-full sm:w-auto"
            >
              <Icon v-if="uploading" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
              {{ uploading ? $t('media.uploading') : $t('media.upload') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="modal-overlay" @click="closeEditModal">
      <div class="modal-container" @click.stop>
        <div class="modal-panel sm:max-w-lg sm:w-full">
          <div class="modal-header">
            <h3 class="modal-title">{{ $t('media.edit_title') }}</h3>
            <button @click="closeEditModal" class="row-actions__btn">
              <Icon icon="mdi:close" class="h-5 w-5" />
            </button>
          </div>

          <div class="modal-content">
            <div class="form-group">
              <label class="form-label">{{ $t('media.filename') }}</label>
              <input
                v-model="editForm.filename"
                type="text"
                class="form-input"
                :placeholder="$t('media.filename_placeholder')"
              />
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeEditModal" class="btn btn-secondary w-full sm:w-auto">
              {{ $t('common.cancel') }}
            </button>
            <button
              @click="updateMedia"
              :disabled="updating"
              class="btn btn-primary w-full sm:w-auto"
            >
              <Icon v-if="updating" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
              {{ updating ? $t('common.updating') : $t('common.update') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- View Modal -->
    <div v-if="showViewModal" class="modal-overlay" @click="closeViewModal">
      <div class="modal-container" @click.stop>
        <div class="modal-panel sm:max-w-4xl sm:w-full">
          <div class="modal-header">
            <h3 class="modal-title">{{ viewingMedia?.filename }}</h3>
            <button @click="closeViewModal" class="row-actions__btn">
              <Icon icon="mdi:close" class="h-5 w-5" />
            </button>
          </div>

          <div class="modal-content">
            <img
              v-if="isImage(viewingMedia?.mime_type)"
              :src="getMediaUrl(viewingMedia?.url)"
              :alt="viewingMedia?.filename"
              class="mm-view-img"
            />
            <div v-else class="flex items-center justify-center h-64">
              <Icon :icon="getFileIcon(viewingMedia?.mime_type)" class="h-32 w-32 mm-icon-muted" />
            </div>

            <div class="mt-4 space-y-2 text-sm">
              <p><strong>{{ $t('media.filename') }}:</strong> {{ viewingMedia?.filename }}</p>
              <p><strong>{{ $t('media.file_size') }}:</strong> {{ formatSize(viewingMedia?.file_size) }}</p>
              <p><strong>{{ $t('media.mime_type') }}:</strong> {{ viewingMedia?.mime_type }}</p>
              <p v-if="viewingMedia?.width && viewingMedia?.height">
                <strong>{{ $t('media.dimensions') }}:</strong> {{ viewingMedia.width }} × {{ viewingMedia.height }}
              </p>
              <p><strong>{{ $t('media.uploaded_by') }}:</strong> {{ viewingMedia?.uploader?.username || 'N/A' }}</p>
              <p><strong>{{ $t('media.uploaded_at') }}:</strong> {{ formatDate(viewingMedia?.created_at) }}</p>
              <p><strong>{{ $t('media.url') }}:</strong>
                <a :href="getMediaUrl(viewingMedia?.url)" target="_blank" class="mm-link">
                  {{ getMediaUrl(viewingMedia?.url) }}
                </a>
              </p>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeViewModal" class="btn btn-secondary w-full sm:w-auto">
              {{ $t('common.close') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import PageHeader from '@/components/ui/PageHeader.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import { useConfirm } from '@/composables/useConfirm'
import { dateLocaleFor } from '@/utils/homeFormat'
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { getAdminMediaList, getAdminMediaStats, uploadAdminMedia, updateAdminMedia, deleteAdminMedia } from '@/services/api'

const { t, locale } = useI18n()
const { confirm: askConfirm } = useConfirm()
const appStore = useAppStore()

// State
const mediaList = ref([])
const loading = ref(false)
const searchQuery = ref('')
const typeFilter = ref('')
const currentPage = ref(1)
const totalPages = ref(1)
const pageSize = ref(24)

// Stats
const stats = ref({
  totalFiles: 0,
  totalImages: 0,
  totalDocuments: 0,
  totalSize: 0
})

// Upload
const showUploadModal = ref(false)
const selectedFile = ref(null)
const uploading = ref(false)
const uploadProgress = ref(0)
const isDragging = ref(false)

// Edit
const showEditModal = ref(false)
const editingMedia = ref(null)
const editForm = ref({ filename: '' })
const updating = ref(false)

// View
const showViewModal = ref(false)
const viewingMedia = ref(null)

// Computed
// La recherche est faite côté serveur, sur toute la médiathèque
const filteredMedia = computed(() => mediaList.value)

// Methods
const fetchMedia = async () => {
  try {
    loading.value = true
    const params = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    if (typeFilter.value) {
      params.type = typeFilter.value
    }
    const search = searchQuery.value.trim()
    if (search) {
      params.search = search
    }
    const response = await getAdminMediaList(params)
    mediaList.value = response.data.data || []
    totalPages.value = response.data.total_pages || 1
  } catch (error) {
    console.error('Failed to fetch media:', error)
    appStore.showError(t('common.feedback.loadError'))
  } finally {
    loading.value = false
  }
}

// Totaux de toute la médiathèque (et non de la page affichée)
const fetchStats = async () => {
  try {
    const { data } = await getAdminMediaStats()
    stats.value = {
      totalFiles: data.total_files || 0,
      totalImages: data.total_images || 0,
      totalDocuments: data.total_documents || 0,
      totalSize: data.total_size || 0
    }
  } catch (error) {
    console.error('Failed to fetch media stats:', error)
  }
}

// Le filtre de type est appliqué côté serveur : recharger depuis la première page
watch(typeFilter, () => {
  currentPage.value = 1
  fetchMedia()
})

// Recherche côté serveur, déclenchée après une courte pause de frappe
const SEARCH_DEBOUNCE_MS = 300
let searchTimer = null
watch(searchQuery, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    fetchMedia()
  }, SEARCH_DEBOUNCE_MS)
})
onBeforeUnmount(() => clearTimeout(searchTimer))

const changePage = (page) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    fetchMedia()
  }
}

const openUploadModal = () => {
  showUploadModal.value = true
  selectedFile.value = null
  uploadProgress.value = 0
}

const closeUploadModal = () => {
  showUploadModal.value = false
  selectedFile.value = null
  uploadProgress.value = 0
}

const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (file) {
    selectedFile.value = file
  }
}

const handleDrop = (event) => {
  isDragging.value = false
  const file = event.dataTransfer.files[0]
  if (file) {
    selectedFile.value = file
  }
}

const uploadFile = async () => {
  if (!selectedFile.value) return

  try {
    uploading.value = true
    const formData = new FormData()
    formData.append('file', selectedFile.value)

    const response = await uploadAdminMedia(formData, (progressEvent) => {
      uploadProgress.value = Math.round((progressEvent.loaded * 100) / progressEvent.total)
    })

    if (response.data.media) {
      mediaList.value.unshift(response.data.media)
    }
    fetchStats()
    closeUploadModal()
    appStore.showSuccess(t('common.feedback.created'))
  } catch (error) {
    console.error('Failed to upload file:', error)
    appStore.showError(t('media.upload_error'))
  } finally {
    uploading.value = false
  }
}

const editMedia = (media) => {
  editingMedia.value = media
  editForm.value = { filename: media.filename }
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
  editingMedia.value = null
  editForm.value = { filename: '' }
}

const updateMedia = async () => {
  if (!editingMedia.value || !editForm.value.filename) return

  try {
    updating.value = true
    const response = await updateAdminMedia(editingMedia.value.id, editForm.value)
    if (response.data.media) {
      const index = mediaList.value.findIndex(m => m.id === editingMedia.value.id)
      if (index !== -1) {
        mediaList.value[index] = response.data.media
      }
    }
    closeEditModal()
    appStore.showSuccess(t('common.feedback.saved'))
  } catch (error) {
    console.error('Failed to update media:', error)
    appStore.showError(t('media.update_error'))
  } finally {
    updating.value = false
  }
}

const viewMedia = (media) => {
  viewingMedia.value = media
  showViewModal.value = true
}

const closeViewModal = () => {
  showViewModal.value = false
  viewingMedia.value = null
}

const copyUrl = (media) => {
  const url = getMediaUrl(media.url)
  navigator.clipboard.writeText(url)
  appStore.showSuccess(t('media.url_copied'))
}

const downloadMedia = (media) => {
  const link = document.createElement('a')
  link.href = getMediaUrl(media.url)
  link.download = media.filename
  link.click()
}

const confirmDelete = async (media) => {
  const confirmed = await askConfirm({
    title: t('media.confirmDeleteTitle'),
    message: t('media.confirmDeleteText', { name: media.filename }),
    confirmLabel: t('common.delete'),
    tone: 'danger',
  })
  if (!confirmed) return

  try {
    await deleteAdminMedia(media.id)
    appStore.showSuccess(t('common.feedback.deleted'))
    mediaList.value = mediaList.value.filter(m => m.id !== media.id)
    fetchStats()
  } catch (error) {
    console.error('Failed to delete media:', error)
    appStore.showError(t('media.delete_error'))
  }
}

const isImage = (mimeType) => {
  return mimeType?.startsWith('image/')
}

const getFileIcon = (mimeType) => {
  if (!mimeType) return 'mdi:file'
  if (mimeType.startsWith('image/')) return 'mdi:image'
  if (mimeType.startsWith('video/')) return 'mdi:video'
  if (mimeType === 'application/pdf') return 'mdi:file-pdf-box'
  if (mimeType.includes('word')) return 'mdi:file-word'
  if (mimeType.includes('sheet') || mimeType.includes('excel')) return 'mdi:file-excel'
  if (mimeType.includes('presentation') || mimeType.includes('powerpoint')) return 'mdi:file-powerpoint'
  if (mimeType.includes('zip') || mimeType.includes('rar') || mimeType.includes('7z') || mimeType.includes('tar') || mimeType.includes('gz')) return 'mdi:folder-zip'
  return 'mdi:file'
}

const getMediaUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `${window.location.origin}${url}`
}

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const lang = dateLocaleFor(locale.value)
  return date.toLocaleDateString(lang) + ' ' + date.toLocaleTimeString(lang)
}

// Vignette introuvable : on affiche l'icône du type de fichier (l'ancien repli vers
// /placeholder.png, absent du projet, relançait l'erreur en boucle).
const brokenImages = ref(new Set())
const markBroken = (id) => {
  brokenImages.value = new Set(brokenImages.value).add(id)
}

// Lifecycle
onMounted(() => {
  fetchMedia()
  fetchStats()
})

// Échap ferme la modale
useEscapeClose(showUploadModal, () => closeUploadModal())
useEscapeClose(showEditModal, () => closeEditModal())
useEscapeClose(showViewModal, () => closeViewModal())
</script>

<style scoped>
.mm-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1px;
  margin: 0 0 16px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--line);
  overflow: hidden;
}

.mm-stat { padding: 12px 14px; background: var(--surface); }
.mm-stat dt { font-size: 12px; color: var(--muted); }
.mm-stat dd { margin: 2px 0 0; font-family: var(--f-num); font-size: 20px; font-weight: 600; color: var(--text); }

.mm-filters { display: flex; flex-wrap: wrap; gap: 8px; margin-block-end: 16px; }
.mm-search { position: relative; flex: 1 1 220px; min-width: 160px; }

.mm-search__icon {
  position: absolute;
  inset-inline-start: 9px;
  inset-block-start: 50%;
  width: 16px;
  height: 16px;
  transform: translateY(-50%);
  color: var(--muted);
  pointer-events: none;
}

.mm-search__input {
  width: 100%;
  padding-block: 7px;
  padding-inline: 30px 10px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
  color: var(--text);
  font-size: 13px;
}

.mm-search__input:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

.mm-loading { display: grid; place-items: center; padding-block: 48px; color: var(--muted); }
.mm-spin { width: 28px; height: 28px; animation: mm-rot 1s linear infinite; }
@keyframes mm-rot { to { transform: rotate(360deg); } }

.mm-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.mm-card {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
  overflow: hidden;
}

.mm-card:hover { border-color: var(--muted); }

.mm-card__preview {
  display: grid;
  place-items: center;
  height: 150px;
  padding: 0;
  border: 0;
  border-block-end: 1px solid var(--line);
  background: var(--sunken);
  cursor: pointer;
  overflow: hidden;
}

.mm-card__preview:focus-visible { outline: 2px solid var(--acc); outline-offset: -2px; }
.mm-card__preview img { width: 100%; height: 100%; object-fit: cover; }
.mm-card__file-icon { width: 48px; height: 48px; color: var(--muted); }

.mm-card__body { flex: 1; padding: 10px 12px 6px; min-width: 0; }

.mm-card__name {
  margin: 0 0 4px;
  overflow: hidden;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--text);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mm-card__meta {
  display: flex;
  gap: 10px;
  margin: 0;
  font-family: var(--f-num);
  font-size: 11px;
  letter-spacing: 0.03em;
  color: var(--muted);
}

.mm-card__actions { display: flex; justify-content: flex-end; gap: 2px; padding: 4px 6px 6px; }

.mm-pager { display: flex; align-items: center; justify-content: center; gap: 12px; margin-block-start: 18px; }
.mm-pager__page { font-family: var(--f-num); font-size: 12px; color: var(--muted); }

.mm-pager__btn {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid var(--line);
  border-radius: 3px;
  background: var(--surface);
  color: var(--text);
  cursor: pointer;
}

.mm-pager__btn:hover:not(:disabled) { background: var(--sunken); }
.mm-pager__btn:disabled { opacity: 0.4; cursor: not-allowed; }
.mm-pager__btn:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }
[dir='rtl'] .mm-pager__btn .iconify { transform: scaleX(-1); }

/* Modales : zone de dépôt et fichier choisi */
.mm-drop {
  padding: 32px;
  border: 2px dashed var(--line);
  border-radius: 4px;
  text-align: center;
}

.mm-drop--on { border-color: var(--acc); background: var(--sunken); }
.mm-drop__icon { width: 56px; height: 56px; margin: 0 auto 12px; color: var(--muted); }
.mm-drop__text { margin: 0 0 8px; color: var(--text); }
.mm-drop__or { margin: 0 0 12px; font-size: 13px; color: var(--muted); }
.mm-selected { margin-block-start: 16px; padding: 12px; border: 1px solid var(--line); border-radius: 4px; background: var(--sunken); }
.mm-icon { color: var(--acc-ink); }
.mm-icon-muted { color: var(--muted); }
.mm-dim { font-size: 12px; color: var(--muted); }
.mm-center { text-align: center; }
.mm-bar { height: 6px; border-radius: 3px; background: var(--sunken); overflow: hidden; }
.mm-bar__fill { height: 100%; background: var(--acc); transition: width 0.3s; }
.mm-view-img { width: 100%; height: auto; border-radius: 4px; }
.mm-link { color: var(--acc-ink); text-decoration: underline; overflow-wrap: anywhere; }

@media (prefers-reduced-motion: reduce) {
  .mm-spin { animation: none; }
  .mm-bar__fill { transition: none; }
}
</style>
