<template>
  <div class="fixed inset-0 z-50 overflow-y-auto" @click.self="$emit('close')">
    <div class="flex items-center justify-center min-h-screen px-4">
      <div class="fixed inset-0 bg-black opacity-50" @click="$emit('close')"></div>

      <div class="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full p-6 z-10">
        <!-- Header -->
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-xl font-semibold text-gray-900 dark:text-white">
            {{ $t('comments.edit_title') }}
          </h3>
          <button
            class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            @click="$emit('close')"
          >
            <Icon icon="mdi:close" class="h-6 w-6" />
          </button>
        </div>

        <!-- Form -->
        <div>
          <textarea
            v-model="editedContent"
            :maxlength="maxLength"
            class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg
                   focus:ring-2 focus:ring-blue-500 focus:border-transparent
                   bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                   resize-none"
            rows="5"
          />
          <div class="flex items-center justify-between mt-2">
            <span class="text-sm text-gray-500 dark:text-gray-400">
              {{ editedContent.length }} / {{ maxLength }}
            </span>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex items-center justify-end gap-3 mt-6">
          <button
            class="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700
                   rounded-lg transition-colors"
            @click="$emit('close')"
          >
            {{ $t('common.cancel') }}
          </button>
          <button
            :disabled="!editedContent.trim()"
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700
                   disabled:opacity-50 disabled:cursor-not-allowed
                   transition-colors"
            @click="save"
          >
            {{ $t('common.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps({
  comment: {
    type: Object,
    required: true
  },
  maxLength: {
    type: Number,
    default: 1000
  }
})

const emit = defineEmits(['close', 'save'])

const editedContent = ref('')

const save = () => {
  if (!editedContent.value.trim()) return
  emit('save', {
    ...props.comment,
    content: editedContent.value.trim()
  })
}

onMounted(() => {
  editedContent.value = props.comment.content
})
</script>
