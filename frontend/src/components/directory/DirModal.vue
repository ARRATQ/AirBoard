<template>
  <teleport to="body">
    <div class="modal-overlay" @click.self="$emit('close')" @keydown.esc="$emit('close')">
      <div class="modal-container">
        <form
          class="modal-panel modal-form w-full"
          :class="widthClass"
          role="dialog"
          aria-modal="true"
          :aria-label="title"
          novalidate
          @submit.prevent="$emit('submit')"
        >
          <div class="modal-header flex items-start justify-between gap-4">
            <div>
              <h3 class="modal-title">{{ title }}</h3>
              <p v-if="subtitle" class="modal-subtitle">{{ subtitle }}</p>
            </div>
            <button type="button" class="btn-ghost rounded-full p-1" :aria-label="$t('common.close')" @click="$emit('close')">
              <Icon icon="mdi:close" class="h-5 w-5" />
            </button>
          </div>

          <div class="modal-content">
            <p v-if="error" class="mb-4 rounded-lg border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/30 dark:text-red-300">
              {{ error }}
            </p>
            <slot />
          </div>

          <div class="modal-footer-bar">
            <span class="modal-footer-hint">{{ $t('common.requiredFieldsHint') }}</span>
            <div class="modal-footer-actions">
              <button type="button" class="btn btn-secondary" @click="$emit('close')">{{ $t('common.cancel') }}</button>
              <button type="submit" class="btn btn-primary" :disabled="saving">
                <Icon v-if="saving" icon="mdi:loading" class="animate-spin h-4 w-4 mr-1" />
                {{ saving ? $t('common.saving') : submitLabel || $t('common.save') }}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </teleport>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps({
  title: { type: String, required: true },
  subtitle: { type: String, default: '' },
  saving: { type: Boolean, default: false },
  error: { type: String, default: '' },
  submitLabel: { type: String, default: '' },
  size: { type: String, default: 'md' }, // sm | md | lg
})
defineEmits(['close', 'submit'])

const widthClass = computed(() => ({ sm: 'max-w-lg', md: 'max-w-3xl', lg: 'max-w-5xl' })[props.size] || 'max-w-3xl')
</script>
