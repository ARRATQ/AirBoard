<template>
  <div class="form-group">
    <label v-if="label" :for="id" class="form-label" :class="{ 'form-label-required': required }">
      {{ label }}
    </label>
    <div class="icon-input-container">
      <input
        :id="id"
        :value="modelValue"
        @input="$emit('update:modelValue', $event.target.value)"
        type="text"
        :placeholder="placeholder || t('ui.iconInput.placeholder')"
        :required="required"
        class="form-input icon-input"
      />
      <div class="icon-preview">
        <Icon 
          v-if="isValidIcon" 
          :icon="modelValue" 
          class="h-5 w-5" 
          @error="onIconError"
        />
        <Icon 
          v-else 
          icon="mdi:help-circle-outline" 
          class="h-5 w-5 text-gray-400" 
        />
      </div>
    </div>
    <div v-if="showHelp" class="form-help">
      {{ helpText }}
    </div>
    <div v-if="error" class="form-error">{{ error }}</div>
    
    <!-- Suggestions populaires -->
    <div v-if="showSuggestions && suggestions.length > 0" class="mt-2">
      <p class="text-xs text-gray-500 mb-2">{{ t('ui.iconInput.popular') }}</p>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="suggestion in suggestions"
          :key="suggestion.icon"
          @click="selectSuggestion(suggestion.icon)"
          type="button"
          class="inline-flex items-center gap-2 px-3 py-1 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-md text-xs transition-colors"
        >
          <Icon :icon="suggestion.icon" class="h-3 w-3" aria-hidden="true" />
          <span>{{ suggestion.labelKey ? t(suggestion.labelKey) : suggestion.label }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  label: {
    type: String,
    default: ''
  },
  placeholder: {
    type: String,
    default: ''
  },
  required: {
    type: Boolean,
    default: false
  },
  id: {
    type: String,
    default: () => `icon-input-${Math.random().toString(36).substr(2, 9)}`
  },
  error: {
    type: String,
    default: ''
  },
  showHelp: {
    type: Boolean,
    default: true
  },
  showSuggestions: {
    type: Boolean,
    default: true
  },
  category: {
    type: String,
    default: 'general'
  }
})

const emit = defineEmits(['update:modelValue'])

const isValidIcon = ref(true)

const helpText = computed(() => {
  return t('ui.iconInput.help')
})

const suggestions = computed(() => {
  // labelKey : traduit au rendu ; label : noms de marque laissés tels quels
  const suggestionsByCategory = {
    general: [
      { icon: 'mdi:home', labelKey: 'common.home' },
      { icon: 'mdi:cog', labelKey: 'common.settings' },
      { icon: 'mdi:account', labelKey: 'ui.iconInput.suggestions.user' },
      { icon: 'mdi:folder', labelKey: 'ui.iconInput.suggestions.folder' },
      { icon: 'mdi:file-document', labelKey: 'ui.iconInput.suggestions.document' },
      { icon: 'mdi:chart-line', labelKey: 'ui.iconInput.suggestions.chart' }
    ],
    application: [
      { icon: 'mdi:application', labelKey: 'ui.iconInput.suggestions.application' },
      { icon: 'mdi:web', labelKey: 'ui.iconInput.suggestions.web' },
      { icon: 'mdi:git', label: 'Git' },
      { icon: 'mdi:database', labelKey: 'ui.iconInput.suggestions.database' },
      { icon: 'mdi:server', labelKey: 'ui.iconInput.suggestions.server' },
      { icon: 'mdi:cloud', labelKey: 'ui.iconInput.suggestions.cloud' },
      { icon: 'mdi:email', labelKey: 'ui.iconInput.suggestions.email' },
      { icon: 'mdi:chat', labelKey: 'ui.iconInput.suggestions.chat' },
      { icon: 'mdi:calendar', labelKey: 'ui.iconInput.suggestions.calendar' },
      { icon: 'mdi:tools', labelKey: 'ui.iconInput.suggestions.tools' }
    ],
    group: [
      { icon: 'mdi:folder-multiple', labelKey: 'ui.iconInput.suggestions.groups' },
      { icon: 'mdi:account-group', labelKey: 'ui.iconInput.suggestions.team' },
      { icon: 'mdi:shield-check', labelKey: 'ui.iconInput.suggestions.security' },
      { icon: 'mdi:code-tags', labelKey: 'ui.iconInput.suggestions.development' },
      { icon: 'mdi:chart-pie', labelKey: 'ui.iconInput.suggestions.analytics' },
      { icon: 'mdi:briefcase', labelKey: 'ui.iconInput.suggestions.business' }
    ],
    social: [
      { icon: 'mdi:github', label: 'GitHub' },
      { icon: 'mdi:gitlab', label: 'GitLab' },
      { icon: 'mdi:slack', label: 'Slack' },
      { icon: 'mdi:discord', label: 'Discord' },
      { icon: 'mdi:microsoft-teams', label: 'Teams' },
      { icon: 'mdi:google', label: 'Google' }
    ]
  }
  
  return suggestionsByCategory[props.category] || suggestionsByCategory.general
})

// Vérifier si l'icône est valide
watch(() => props.modelValue, (newValue) => {
  if (!newValue) {
    isValidIcon.value = false
    return
  }
  
  // Vérification basique du format
  const iconPattern = /^[a-zA-Z0-9\-_]+:[a-zA-Z0-9\-_]+$/
  isValidIcon.value = iconPattern.test(newValue)
}, { immediate: true })

const onIconError = () => {
  isValidIcon.value = false
}

const selectSuggestion = (icon) => {
  emit('update:modelValue', icon)
}
</script>