<template>
  <div class="media-gallery">
    <!-- Filters -->
    <div class="filters">
      <button
        v-for="filter in filters"
        :key="filter.value"
        @click="selectedFilter = filter.value"
        :class="{ 'active': selectedFilter === filter.value }"
        class="filter-button"
      >
        <Icon :icon="filter.icon" class="h-5 w-5" />
        <span>{{ filter.label }}</span>
      </button>
    </div>

    <!-- Media grid -->
    <div v-if="loading" class="loading">
      <Icon icon="mdi:loading" class="animate-spin h-8 w-8" />
      <p>{{ $t('media.loading') }}</p>
    </div>

    <div v-else-if="mediaList.length === 0" class="empty-state">
      <Icon icon="mdi:image-off" class="h-16 w-16 text-gray-300" />
      <p>{{ $t('media.noMedia') }}</p>
    </div>

    <div v-else class="media-grid">
      <div
        v-for="media in mediaList"
        :key="media.id"
        @click="selectMedia(media)"
        class="media-item"
        :class="{ 'selected': selectedMediaId === media.id }"
      >
        <!-- Image preview -->
        <div v-if="media.mime_type.startsWith('image/')" class="media-preview">
          <img :src="media.url" :alt="media.filename" />
        </div>

        <!-- File preview -->
        <div v-else class="media-preview file-preview">
          <Icon :icon="getFileIcon(media)" class="h-12 w-12" />
        </div>

        <!-- Media info -->
        <div class="media-info">
          <p class="media-filename">{{ media.filename }}</p>
          <p class="media-size">{{ formatFileSize(media.file_size) }}</p>
        </div>

        <!-- Selected indicator -->
        <div v-if="selectedMediaId === media.id" class="selected-indicator">
          <Icon icon="mdi:check-circle" class="h-6 w-6" />
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="pagination">
      <button
        @click="currentPage--"
        :disabled="currentPage === 1"
        class="pagination-button"
      >
        <Icon icon="mdi:chevron-left" class="h-5 w-5" />
      </button>
      <span class="pagination-info">
        Page {{ currentPage }} / {{ totalPages }}
      </span>
      <button
        @click="currentPage++"
        :disabled="currentPage === totalPages"
        class="pagination-button"
      >
        <Icon icon="mdi:chevron-right" class="h-5 w-5" />
      </button>
    </div>

    <!-- Actions -->
    <div class="actions">
      <button @click="$emit('cancel')" class="button-secondary">
        {{ $t('common.cancel') }}
      </button>
      <button
        @click="insertSelected"
        :disabled="!selectedMediaId"
        class="button-primary"
      >
        {{ $t('media.insert') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { mediaService } from '@/services/api'

const emit = defineEmits(['select', 'cancel'])

const filters = [
  { label: 'All', value: '', icon: 'mdi:file-multiple' },
  { label: 'Images', value: 'image', icon: 'mdi:image' },
  { label: 'Documents', value: 'document', icon: 'mdi:file-document' },
  { label: 'PDF', value: 'pdf', icon: 'mdi:file-pdf-box' },
]

const selectedFilter = ref('')
const loading = ref(false)
const mediaList = ref([])
const currentPage = ref(1)
const totalPages = ref(1)
const selectedMediaId = ref(null)

const fetchMedia = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: 20,
    }
    if (selectedFilter.value) {
      params.type = selectedFilter.value
    }

    const response = await mediaService.getMediaList(params)
    mediaList.value = response.data || []
    totalPages.value = response.total_pages || 1
  } catch (error) {
    console.error('Failed to fetch media:', error)
  } finally {
    loading.value = false
  }
}

const selectMedia = (media) => {
  selectedMediaId.value = media.id
}

const insertSelected = () => {
  const media = mediaList.value.find(m => m.id === selectedMediaId.value)
  if (media) {
    emit('select', media)
  }
}

const getFileIcon = (media) => {
  const mimeType = media.mime_type
  if (mimeType.startsWith('image/')) return 'mdi:file-image'
  if (mimeType === 'application/pdf') return 'mdi:file-pdf-box'
  if (mimeType.includes('word') || mimeType.includes('document')) return 'mdi:file-word'
  if (mimeType.includes('sheet') || mimeType.includes('excel')) return 'mdi:file-excel'
  if (mimeType.includes('presentation') || mimeType.includes('powerpoint')) return 'mdi:file-powerpoint'
  if (mimeType.includes('zip') || mimeType.includes('rar')) return 'mdi:folder-zip'
  return 'mdi:file'
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

// Watch for filter and page changes
watch([selectedFilter, currentPage], () => {
  fetchMedia()
})

onMounted(() => {
  fetchMedia()
})
</script>

<script>
export default {
  name: 'MediaGallery'
}
</script>

<style scoped>
.media-gallery {
  @apply flex flex-col gap-4;
}

.filters {
  @apply flex gap-2 flex-wrap;
}

.filter-button {
  @apply flex items-center gap-2 px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300;
  @apply hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors;
}

.filter-button.active {
  @apply bg-blue-500 text-white border-blue-500;
}

.loading {
  @apply flex flex-col items-center justify-center gap-2 py-12 text-gray-500 dark:text-gray-400;
}

.empty-state {
  @apply flex flex-col items-center justify-center gap-2 py-12 text-gray-400 dark:text-gray-500;
}

.media-grid {
  @apply grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4;
}

.media-item {
  @apply relative cursor-pointer rounded-lg border-2 border-gray-200 dark:border-gray-700 overflow-hidden transition-all;
  @apply hover:border-blue-500 hover:shadow-lg;
}

.media-item.selected {
  @apply border-blue-500 shadow-lg;
}

.media-preview {
  @apply aspect-square bg-gray-100 dark:bg-gray-800 flex items-center justify-center;
}

.media-preview img {
  @apply w-full h-full object-cover;
}

.file-preview {
  @apply text-gray-400 dark:text-gray-500;
}

.media-info {
  @apply p-2 bg-white dark:bg-gray-900;
}

.media-filename {
  @apply text-xs font-medium text-gray-700 dark:text-gray-300 truncate;
}

.media-size {
  @apply text-xs text-gray-500 dark:text-gray-400;
}

.selected-indicator {
  @apply absolute top-2 right-2 bg-blue-500 text-white rounded-full p-1;
}

.pagination {
  @apply flex items-center justify-center gap-4;
}

.pagination-button {
  @apply p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300;
  @apply hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors;
}

.pagination-info {
  @apply text-sm text-gray-600 dark:text-gray-400;
}

.actions {
  @apply flex gap-2 justify-end pt-4 border-t border-gray-200 dark:border-gray-700;
}

.button-secondary {
  @apply px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300;
  @apply hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors;
}

.button-primary {
  @apply px-4 py-2 rounded-lg bg-blue-500 text-white;
  @apply hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors;
}
</style>
