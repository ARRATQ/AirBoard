<template>
  <div class="content-area">
    <!-- Header -->
    <div class="page-header">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="page-title">{{ $t('media.title') }}</h1>
          <p class="page-subtitle">{{ $t('media.subtitle') }}</p>
        </div>
        <button @click="openUploadModal" class="btn btn-primary">
          <Icon icon="mdi:upload" class="h-4 w-4 mr-2" />
          {{ $t('media.upload') }}
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="filters-container">
      <div class="search-container flex-1">
        <Icon icon="mdi:magnify" class="search-icon" />
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="$t('media.searchPlaceholder')"
          class="form-input search-input"
        />
      </div>
      <select v-model="typeFilter" class="form-select w-full sm:w-auto">
        <option value="">{{ $t('media.filter_all_types') }}</option>
        <option value="image">{{ $t('media.type_image') }}</option>
        <option value="video">{{ $t('media.type_video') }}</option>
        <option value="pdf">{{ $t('media.type_pdf') }}</option>
        <option value="document">{{ $t('media.type_document') }}</option>
        <option value="spreadsheet">{{ $t('media.type_spreadsheet') }}</option>
        <option value="presentation">{{ $t('media.type_presentation') }}</option>
      </select>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4 mb-6">
      <div class="card">
        <div class="card-content">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-400">{{ $t('media.total_files') }}</p>
              <p class="text-2xl font-bold text-white">{{ stats.totalFiles }}</p>
            </div>
            <Icon icon="mdi:file-multiple" class="h-8 w-8 text-blue-500" />
          </div>
        </div>
      </div>
      <div class="card">
        <div class="card-content">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-400">{{ $t('media.total_images') }}</p>
              <p class="text-2xl font-bold text-white">{{ stats.totalImages }}</p>
            </div>
            <Icon icon="mdi:image" class="h-8 w-8 text-green-500" />
          </div>
        </div>
      </div>
      <div class="card">
        <div class="card-content">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-400">{{ $t('media.total_documents') }}</p>
              <p class="text-2xl font-bold text-white">{{ stats.totalDocuments }}</p>
            </div>
            <Icon icon="mdi:file-document" class="h-8 w-8 text-purple-500" />
          </div>
        </div>
      </div>
      <div class="card">
        <div class="card-content">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-400">{{ $t('media.total_size') }}</p>
              <p class="text-2xl font-bold text-white">{{ formatSize(stats.totalSize) }}</p>
            </div>
            <Icon icon="mdi:database" class="h-8 w-8 text-yellow-500" />
          </div>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <Icon icon="mdi:loading" class="h-8 w-8 animate-spin text-gray-400" />
    </div>

    <div v-else-if="filteredMedia.length > 0" class="grid-container grid-cols-auto">
      <div
        v-for="media in filteredMedia"
        :key="media.id"
        class="card hover:border-gray-600 transition-all duration-200 fade-in relative group"
      >
        <!-- Preview -->
        <div class="relative h-48 bg-gray-800 rounded-t-lg overflow-hidden">
          <img
            v-if="isImage(media.mime_type)"
            :src="getMediaUrl(media.url)"
            :alt="media.filename"
            class="w-full h-full object-cover"
            @error="handleImageError"
          />
          <div v-else class="flex items-center justify-center h-full">
            <Icon :icon="getFileIcon(media.mime_type)" class="h-16 w-16 text-gray-500" />
          </div>

          <!-- Actions overlay -->
          <div class="absolute inset-0 bg-black bg-opacity-50 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex items-center justify-center gap-2">
            <button @click="viewMedia(media)" class="btn-ghost" :title="$t('common.view')">
              <Icon icon="mdi:eye" class="h-5 w-5" />
            </button>
            <button @click="copyUrl(media)" class="btn-ghost" :title="$t('media.copy_url')">
              <Icon icon="mdi:content-copy" class="h-5 w-5" />
            </button>
            <button @click="downloadMedia(media)" class="btn-ghost" :title="$t('common.download')">
              <Icon icon="mdi:download" class="h-5 w-5" />
            </button>
            <button @click="editMedia(media)" class="btn-ghost" :title="$t('common.edit')">
              <Icon icon="mdi:pencil" class="h-5 w-5" />
            </button>
            <button @click="confirmDelete(media)" class="btn-ghost text-red-400 hover:text-red-300" :title="$t('common.delete')">
              <Icon icon="mdi:delete" class="h-5 w-5" />
            </button>
          </div>
        </div>

        <!-- Info -->
        <div class="card-content">
          <h3 class="card-title truncate" :title="media.filename">{{ media.filename }}</h3>
          <div class="mt-2 space-y-1 text-xs text-gray-400">
            <p>{{ formatSize(media.file_size) }}</p>
            <p v-if="media.width && media.height">{{ media.width }} × {{ media.height }}</p>
            <p>{{ $t('media.uploaded_by') }}: {{ media.uploader?.username || 'N/A' }}</p>
            <p>{{ formatDate(media.created_at) }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="mediaList.length === 0" class="empty-state">
      <Icon icon="mdi:image-multiple" class="empty-state-icon" />
      <h3 class="empty-state-title">{{ $t('media.emptyTitle') }}</h3>
      <p class="empty-state-description">{{ $t('media.emptyText') }}</p>
      <button @click="openUploadModal" class="btn btn-primary">
        <Icon icon="mdi:upload" class="h-4 w-4 mr-2" />
        {{ $t('media.upload') }}
      </button>
    </div>

    <div v-else class="empty-state">
      <Icon icon="mdi:magnify" class="empty-state-icon" />
      <h3 class="empty-state-title">{{ $t('media.noResultsTitle') }}</h3>
      <p class="empty-state-description">{{ $t('media.noResultsText') }}</p>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="pagination">
      <button
        @click="changePage(currentPage - 1)"
        :disabled="currentPage === 1"
        class="btn btn-secondary"
      >
        <Icon icon="mdi:chevron-left" class="h-5 w-5" />
      </button>
      <span class="text-gray-400">
        {{ $t('common.page') }} {{ currentPage }} {{ $t('common.of') }} {{ totalPages }}
      </span>
      <button
        @click="changePage(currentPage + 1)"
        :disabled="currentPage === totalPages"
        class="btn btn-secondary"
      >
        <Icon icon="mdi:chevron-right" class="h-5 w-5" />
      </button>
    </div>

    <!-- Upload Modal -->
    <div v-if="showUploadModal" class="modal-overlay" @click="closeUploadModal">
      <div class="modal-container" @click.stop>
        <div class="modal-panel sm:max-w-lg sm:w-full">
          <div class="modal-header">
            <h3 class="modal-title">{{ $t('media.upload_title') }}</h3>
            <button @click="closeUploadModal" class="btn-ghost">
              <Icon icon="mdi:close" class="h-5 w-5" />
            </button>
          </div>

          <div class="modal-content">
            <div
              @dragover.prevent="isDragging = true"
              @dragleave.prevent="isDragging = false"
              @drop.prevent="handleDrop"
              :class="['border-2 border-dashed rounded-lg p-8 text-center transition-colors', isDragging ? 'border-blue-500 bg-blue-500/10' : 'border-gray-600']"
            >
              <Icon icon="mdi:cloud-upload" class="h-16 w-16 mx-auto text-gray-400 mb-4" />
              <p class="text-gray-300 mb-2">{{ $t('media.drag_drop') }}</p>
              <p class="text-sm text-gray-500 mb-4">{{ $t('media.or') }}</p>
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

            <div v-if="selectedFile" class="mt-4 p-4 bg-gray-800 rounded-lg">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <Icon :icon="getFileIcon(selectedFile.type)" class="h-8 w-8 text-blue-500" />
                  <div>
                    <p class="text-sm font-medium">{{ selectedFile.name }}</p>
                    <p class="text-xs text-gray-400">{{ formatSize(selectedFile.size) }}</p>
                  </div>
                </div>
                <button @click="selectedFile = null" class="btn-ghost text-red-400">
                  <Icon icon="mdi:close" class="h-5 w-5" />
                </button>
              </div>
            </div>

            <div v-if="uploadProgress > 0 && uploadProgress < 100" class="mt-4">
              <div class="w-full bg-gray-700 rounded-full h-2">
                <div
                  class="bg-blue-500 h-2 rounded-full transition-all duration-300"
                  :style="{ width: uploadProgress + '%' }"
                ></div>
              </div>
              <p class="text-sm text-gray-400 mt-2 text-center">{{ uploadProgress }}%</p>
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
            <button @click="closeEditModal" class="btn-ghost">
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
            <button @click="closeViewModal" class="btn-ghost">
              <Icon icon="mdi:close" class="h-5 w-5" />
            </button>
          </div>

          <div class="modal-content">
            <img
              v-if="isImage(viewingMedia?.mime_type)"
              :src="getMediaUrl(viewingMedia?.url)"
              :alt="viewingMedia?.filename"
              class="w-full h-auto rounded-lg"
            />
            <div v-else class="flex items-center justify-center h-64">
              <Icon :icon="getFileIcon(viewingMedia?.mime_type)" class="h-32 w-32 text-gray-500" />
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
                <a :href="getMediaUrl(viewingMedia?.url)" target="_blank" class="text-blue-400 hover:text-blue-300">
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

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="closeDeleteModal">
      <div class="modal-container" @click.stop>
        <div class="modal-panel sm:max-w-lg sm:w-full">
          <div class="modal-header">
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon icon="mdi:alert" class="h-6 w-6 text-red-500" />
              </div>
              <div>
                <h3 class="modal-title">{{ $t('media.confirmDeleteTitle') }}</h3>
                <p class="modal-subtitle">
                  {{ $t('media.confirmDeleteText', { name: mediaToDelete?.filename }) }}
                </p>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeDeleteModal" class="btn btn-secondary w-full sm:w-auto">
              {{ $t('common.cancel') }}
            </button>
            <button
              @click="deleteMedia"
              :disabled="deleting"
              class="btn btn-danger w-full sm:w-auto"
            >
              <Icon v-if="deleting" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
              {{ deleting ? $t('common.deleting') : $t('common.delete') }}
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
import { getAdminMediaList, uploadAdminMedia, updateAdminMedia, deleteAdminMedia } from '@/services/api'

const { t } = useI18n()

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

// Delete
const showDeleteModal = ref(false)
const mediaToDelete = ref(null)
const deleting = ref(false)

// Computed
const filteredMedia = computed(() => {
  let filtered = [...mediaList.value]

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(media =>
      media.filename.toLowerCase().includes(query)
    )
  }

  return filtered
})

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
    const response = await getAdminMediaList(params)
    mediaList.value = response.data.data || []
    totalPages.value = response.data.total_pages || 1
    calculateStats()
  } catch (error) {
    console.error('Failed to fetch media:', error)
  } finally {
    loading.value = false
  }
}

const calculateStats = () => {
  stats.value.totalFiles = mediaList.value.length
  stats.value.totalImages = mediaList.value.filter(m => m.mime_type.startsWith('image/')).length
  stats.value.totalDocuments = mediaList.value.filter(m => !m.mime_type.startsWith('image/') && !m.mime_type.startsWith('video/')).length
  stats.value.totalSize = mediaList.value.reduce((sum, m) => sum + m.file_size, 0)
}

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
      calculateStats()
    }
    closeUploadModal()
  } catch (error) {
    console.error('Failed to upload file:', error)
    alert(t('media.upload_error'))
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
  } catch (error) {
    console.error('Failed to update media:', error)
    alert(t('media.update_error'))
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
  alert(t('media.url_copied'))
}

const downloadMedia = (media) => {
  const link = document.createElement('a')
  link.href = getMediaUrl(media.url)
  link.download = media.filename
  link.click()
}

const confirmDelete = (media) => {
  mediaToDelete.value = media
  showDeleteModal.value = true
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  mediaToDelete.value = null
}

const deleteMedia = async () => {
  if (!mediaToDelete.value) return

  try {
    deleting.value = true
    await deleteAdminMedia(mediaToDelete.value.id)
    mediaList.value = mediaList.value.filter(m => m.id !== mediaToDelete.value.id)
    calculateStats()
    closeDeleteModal()
  } catch (error) {
    console.error('Failed to delete media:', error)
    alert(t('media.delete_error'))
  } finally {
    deleting.value = false
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
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
}

const handleImageError = (event) => {
  event.target.src = '/placeholder.png'
}

// Lifecycle
onMounted(() => {
  fetchMedia()
})
</script>

<style scoped>
.grid-cols-auto {
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
}

.pagination {
  @apply flex items-center justify-center gap-4 mt-6;
}
</style>
