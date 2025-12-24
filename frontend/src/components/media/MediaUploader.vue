<template>
  <div class="media-uploader">
    <!-- Upload zone -->
    <div
      class="upload-zone"
      :class="{ 'drag-over': isDragOver, 'uploading': isUploading }"
      @drop.prevent="handleDrop"
      @dragover.prevent="isDragOver = true"
      @dragleave.prevent="isDragOver = false"
      @click="triggerFileInput"
    >
      <input
        ref="fileInput"
        type="file"
        :accept="acceptedTypes"
        multiple
        class="hidden"
        @change="handleFileSelect"
      />

      <div v-if="!isUploading" class="upload-content">
        <Icon icon="mdi:cloud-upload" class="upload-icon" />
        <p class="upload-text">
          {{ $t('media.dropFiles') }}
        </p>
        <p class="upload-hint">
          {{ $t('media.orClickToSelect') }}
        </p>
        <p class="upload-types">
          {{ acceptedTypesText }}
        </p>
      </div>

      <div v-else class="uploading-content">
        <Icon icon="mdi:loading" class="animate-spin upload-icon" />
        <p class="upload-text">
          {{ $t('media.uploading') }} {{ uploadProgress }}%
        </p>
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: uploadProgress + '%' }"></div>
        </div>
      </div>
    </div>

    <!-- Error message -->
    <div v-if="errorMessage" class="error-message">
      <Icon icon="mdi:alert-circle" class="h-5 w-5" />
      <span>{{ errorMessage }}</span>
    </div>

    <!-- Upload queue -->
    <div v-if="uploadQueue.length > 0" class="upload-queue">
      <h3 class="queue-title">{{ $t('media.uploadQueue') }}</h3>
      <div
        v-for="item in uploadQueue"
        :key="item.id"
        class="queue-item"
        :class="{ 'success': item.status === 'success', 'error': item.status === 'error' }"
      >
        <Icon
          :icon="getFileIcon(item.file.type)"
          class="file-icon"
        />
        <div class="file-info">
          <p class="file-name">{{ item.file.name }}</p>
          <p class="file-size">{{ formatFileSize(item.file.size) }}</p>
        </div>
        <div class="file-status">
          <Icon
            v-if="item.status === 'uploading'"
            icon="mdi:loading"
            class="animate-spin text-blue-500"
          />
          <Icon
            v-else-if="item.status === 'success'"
            icon="mdi:check-circle"
            class="text-green-500"
          />
          <Icon
            v-else-if="item.status === 'error'"
            icon="mdi:alert-circle"
            class="text-red-500"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Icon } from '@iconify/vue'
import { mediaService } from '@/services/api'

const props = defineProps({
  accept: {
    type: String,
    default: 'image/*,application/pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.zip,.rar'
  },
  maxSizeMB: {
    type: Number,
    default: 10
  },
  asGroupAdmin: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['upload-success', 'upload-error'])

const fileInput = ref(null)
const isDragOver = ref(false)
const isUploading = ref(false)
const uploadProgress = ref(0)
const errorMessage = ref('')
const uploadQueue = ref([])

const acceptedTypes = computed(() => props.accept)

const acceptedTypesText = computed(() => {
  const types = []
  if (props.accept.includes('image')) types.push('Images')
  if (props.accept.includes('pdf')) types.push('PDF')
  if (props.accept.includes('doc')) types.push('Word')
  if (props.accept.includes('xls')) types.push('Excel')
  if (props.accept.includes('ppt')) types.push('PowerPoint')
  if (props.accept.includes('zip')) types.push('Archives')
  return types.join(', ')
})

const triggerFileInput = () => {
  if (!isUploading.value) {
    fileInput.value?.click()
  }
}

const handleFileSelect = (event) => {
  const files = Array.from(event.target.files || [])
  handleFiles(files)
  // Reset input
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const handleDrop = (event) => {
  isDragOver.value = false
  const files = Array.from(event.dataTransfer?.files || [])
  handleFiles(files)
}

const handleFiles = async (files) => {
  if (files.length === 0) return

  errorMessage.value = ''

  // Validate files
  const validFiles = []
  for (const file of files) {
    if (file.size > props.maxSizeMB * 1024 * 1024) {
      errorMessage.value = `${file.name} is too large (max ${props.maxSizeMB}MB)`
      continue
    }
    validFiles.push(file)
  }

  if (validFiles.length === 0) return

  // Add to queue
  const queueItems = validFiles.map((file, index) => ({
    id: Date.now() + index,
    file,
    status: 'pending'
  }))
  uploadQueue.value.push(...queueItems)

  // Upload files
  for (const item of queueItems) {
    await uploadFile(item)
  }
}

const uploadFile = async (queueItem) => {
  queueItem.status = 'uploading'
  isUploading.value = true

  try {
    const uploadMethod = props.asGroupAdmin
      ? mediaService.uploadMediaAsGroupAdmin
      : mediaService.uploadMedia

    const result = await uploadMethod(queueItem.file, (progressEvent) => {
      uploadProgress.value = Math.round((progressEvent.loaded * 100) / progressEvent.total)
    })

    queueItem.status = 'success'
    emit('upload-success', result.media)

    // Remove from queue after 2 seconds
    setTimeout(() => {
      uploadQueue.value = uploadQueue.value.filter(item => item.id !== queueItem.id)
    }, 2000)
  } catch (error) {
    queueItem.status = 'error'
    errorMessage.value = error.response?.data?.message || 'Upload failed'
    emit('upload-error', error)

    // Remove from queue after 5 seconds
    setTimeout(() => {
      uploadQueue.value = uploadQueue.value.filter(item => item.id !== queueItem.id)
    }, 5000)
  } finally {
    isUploading.value = false
    uploadProgress.value = 0
  }
}

const getFileIcon = (mimeType) => {
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
</script>

<script>
export default {
  name: 'MediaUploader'
}
</script>

<style scoped>
.media-uploader {
  @apply w-full;
}

.upload-zone {
  @apply border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg p-8 text-center cursor-pointer transition-all;
  @apply hover:border-blue-500 dark:hover:border-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20;
}

.upload-zone.drag-over {
  @apply border-blue-500 bg-blue-50 dark:bg-blue-900/30;
}

.upload-zone.uploading {
  @apply cursor-not-allowed opacity-75;
}

.upload-content, .uploading-content {
  @apply flex flex-col items-center gap-2;
}

.upload-icon {
  @apply w-12 h-12 text-gray-400 dark:text-gray-500;
}

.upload-text {
  @apply text-lg font-medium text-gray-700 dark:text-gray-300;
}

.upload-hint {
  @apply text-sm text-gray-500 dark:text-gray-400;
}

.upload-types {
  @apply text-xs text-gray-400 dark:text-gray-500 mt-2;
}

.progress-bar {
  @apply w-full max-w-xs h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden;
}

.progress-fill {
  @apply h-full bg-blue-500 transition-all duration-300;
}

.error-message {
  @apply mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg flex items-center gap-2 text-red-700 dark:text-red-400;
}

.upload-queue {
  @apply mt-6;
}

.queue-title {
  @apply text-sm font-medium text-gray-700 dark:text-gray-300 mb-3;
}

.queue-item {
  @apply flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg mb-2;
}

.queue-item.success {
  @apply bg-green-50 dark:bg-green-900/20;
}

.queue-item.error {
  @apply bg-red-50 dark:bg-red-900/20;
}

.file-icon {
  @apply w-8 h-8 text-gray-500 dark:text-gray-400;
}

.file-info {
  @apply flex-1;
}

.file-name {
  @apply text-sm font-medium text-gray-700 dark:text-gray-300 truncate;
}

.file-size {
  @apply text-xs text-gray-500 dark:text-gray-400;
}

.file-status {
  @apply w-6 h-6;
}
</style>
