<template>
  <div class="card">
    <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
      <Icon icon="mdi:tag-multiple" class="h-5 w-5 text-primary-500" />
      {{ $t('groupAdmin.newsEditor.tagsSection') }}
    </h3>

    <div class="space-y-3">
      <!-- Tags sélectionnés -->
      <div
        v-if="modelValue.length > 0"
        class="flex flex-wrap gap-2 p-3 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700"
      >
        <span
          v-for="tagId in modelValue"
          :key="tagId"
          class="tag tag--acc"
        >
          <Icon icon="mdi:tag" class="h-3 w-3" />
          {{ tagName(tagId) }}
          <button
            type="button"
            class="ms-1 hover:text-red-600 dark:hover:text-red-400 transition-colors"
            :aria-label="$t('adminShared.news.removeItem', { name: tagName(tagId) })"
            @click="removeTag(tagId)"
          >
            <Icon icon="mdi:close-circle" class="h-3.5 w-3.5" />
          </button>
        </span>
      </div>
      <div
        v-else
        class="text-sm text-gray-500 dark:text-gray-400 italic p-3 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700"
      >
        {{ $t('groupAdmin.newsEditor.noTagSelected') }}
      </div>

      <!-- Ajout d'un tag existant -->
      <select
        class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm"
        :aria-label="$t('groupAdmin.newsEditor.addTag')"
        @change="addTag"
      >
        <option value="">{{ $t('groupAdmin.newsEditor.addTag') }}</option>
        <option v-for="tag in availableTags" :key="tag.id" :value="tag.id">
          {{ tag.name }}
        </option>
      </select>

      <!-- Création d'un tag (admin, éditeur et admin de groupe y ont droit) -->
      <div class="flex gap-2">
        <input
          v-model="newTagName"
          type="text"
          :placeholder="$t('adminShared.news.newTagPlaceholder')"
          :aria-label="$t('adminShared.news.newTagPlaceholder')"
          class="flex-1 min-w-0 px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm"
          @keydown.enter.prevent="submitTag"
        />
        <input
          v-model="newTagColor"
          type="color"
          :title="$t('adminShared.news.tagColor')"
          :aria-label="$t('adminShared.news.tagColor')"
          class="h-8 w-8 flex-shrink-0 border border-gray-300 dark:border-gray-600 rounded-lg cursor-pointer"
        />
        <button
          type="button"
          :disabled="!newTagName.trim() || isCreating"
          class="px-3 py-1.5 bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white rounded-lg transition-colors flex items-center gap-1 text-sm font-medium disabled:cursor-not-allowed"
          @click="submitTag"
        >
          <Icon v-if="isCreating" icon="mdi:loading" class="h-4 w-4 animate-spin" />
          <Icon v-else icon="mdi:plus" class="h-4 w-4" />
          <span>{{ $t('common.create') }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Icon } from '@iconify/vue'

const DEFAULT_TAG_COLOR = '#3B82F6'

const props = defineProps({
  // Identifiants des tags sélectionnés
  modelValue: { type: Array, required: true },
  // Tous les tags existants
  tags: { type: Array, required: true },
  // Crée un tag ({ name, color }) ; renvoie true si le champ peut être vidé
  createTag: { type: Function, required: true }
})

const emit = defineEmits(['update:modelValue'])

const newTagName = ref('')
const newTagColor = ref(DEFAULT_TAG_COLOR)
const isCreating = ref(false)

const availableTags = computed(() =>
  props.tags.filter((tag) => !props.modelValue.includes(tag.id))
)

const tagName = (tagId) => props.tags.find((tag) => tag.id === tagId)?.name || ''

const addTag = (event) => {
  const tagId = parseInt(event.target.value, 10)
  if (tagId && !props.modelValue.includes(tagId)) {
    emit('update:modelValue', [...props.modelValue, tagId])
  }
  event.target.value = ''
}

const removeTag = (tagId) => {
  emit('update:modelValue', props.modelValue.filter((id) => id !== tagId))
}

const submitTag = async () => {
  const name = newTagName.value.trim()
  if (!name || isCreating.value) return

  isCreating.value = true
  try {
    const done = await props.createTag({ name, color: newTagColor.value })
    if (done) {
      newTagName.value = ''
      newTagColor.value = DEFAULT_TAG_COLOR
    }
  } finally {
    isCreating.value = false
  }
}
</script>
