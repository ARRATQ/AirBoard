<template>
  <DirModal
    :title="field ? $t('directory.editField') : $t('directory.newField')"
    :saving="saving"
    :error="error"
    size="md"
    @close="$emit('close')"
    @submit="submit"
  >
    <div class="grid gap-x-6 md:grid-cols-2">
      <div>
        <div class="form-group">
          <label class="form-label form-label-required" for="fld-label">{{ $t('directory.fieldLabel') }}</label>
          <input id="fld-label" v-model.trim="form.label" class="form-input" maxlength="100" required autofocus @input="syncKey" />
        </div>
        <div class="form-group">
          <label class="form-label" for="fld-key">{{ $t('directory.fieldKey') }}</label>
          <input
            id="fld-key"
            v-model.trim="form.key"
            class="form-input font-mono text-sm"
            maxlength="60"
            :disabled="!!field"
            pattern="[a-z][a-z0-9_]*"
            @input="keyTouched = true"
          />
          <p class="text-xs mt-1" style="color: var(--text-muted)">{{ field ? $t('directory.fieldKeyLocked') : $t('directory.fieldKeyHelp') }}</p>
        </div>
        <div class="form-group">
          <span class="form-label form-label-required">{{ $t('directory.fieldType') }}</span>
          <div class="grid grid-cols-2 gap-2">
            <label
              v-for="type in FIELD_TYPES"
              :key="type"
              class="field-type"
              :class="{ 'field-type--active': form.type === type }"
            >
              <input v-model="form.type" type="radio" :value="type" class="sr-only" />
              <Icon :icon="FIELD_TYPE_ICONS[type]" class="h-4 w-4" />
              {{ $t(`directory.fieldTypes.${type}`) }}
            </label>
          </div>
        </div>
      </div>

      <div>
        <div v-if="form.type === 'select'" class="form-group">
          <label class="form-label form-label-required" for="fld-options">{{ $t('directory.fieldOptions') }}</label>
          <textarea id="fld-options" v-model="optionsText" class="form-textarea" rows="5" :placeholder="$t('directory.fieldOptionsPlaceholder')" />
          <p class="text-xs mt-1" style="color: var(--text-muted)">{{ $t('directory.fieldOptionsHelp') }}</p>
        </div>
        <div class="form-group">
          <span class="form-label">{{ $t('directory.fieldVisibility') }}</span>
          <div class="space-y-2">
            <label v-for="v in ['all', 'managers']" :key="v" class="flex items-start gap-3 cursor-pointer">
              <input v-model="form.visibility" type="radio" :value="v" class="mt-1" />
              <span>
                <span class="block text-sm font-medium" style="color: var(--text-primary)">{{ $t(`directory.visibility.${v}`) }}</span>
                <span class="block text-xs" style="color: var(--text-muted)">{{ $t(`directory.visibility.${v}Help`) }}</span>
              </span>
            </label>
          </div>
        </div>
        <div class="form-group space-y-2">
          <label class="flex items-center gap-3 cursor-pointer">
            <input v-model="form.required" type="checkbox" />
            <span class="text-sm" style="color: var(--text-primary)">{{ $t('directory.fieldRequired') }}</span>
          </label>
          <label class="flex items-center gap-3 cursor-pointer">
            <input v-model="form.show_on_card" type="checkbox" />
            <span class="text-sm" style="color: var(--text-primary)">{{ $t('directory.fieldShowOnCard') }}</span>
          </label>
        </div>
        <div class="form-group w-32">
          <label class="form-label" for="fld-order">{{ $t('directory.order') }}</label>
          <input id="fld-order" v-model.number="form.order" type="number" class="form-input" />
        </div>
      </div>
    </div>
  </DirModal>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import DirModal from './DirModal.vue'
import { directoryService } from '@/services/api'
import { apiError, FIELD_TYPES, FIELD_TYPE_ICONS } from './useDirectory'

const props = defineProps({
  field: { type: Object, default: null },
  nextOrder: { type: Number, default: 0 },
})
const emit = defineEmits(['close', 'saved'])
const { t } = useI18n()

const form = reactive({
  label: props.field?.label || '',
  key: props.field?.key || '',
  type: props.field?.type || 'text',
  required: props.field?.required ?? false,
  visibility: props.field?.visibility || 'all',
  show_on_card: props.field?.show_on_card ?? false,
  order: props.field?.order ?? props.nextOrder,
})
const optionsText = ref((props.field?.options || []).join('\n'))
const keyTouched = ref(!!props.field)
const saving = ref(false)
const error = ref('')

const slugify = (s) =>
  s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase()
    .replace(/[^a-z0-9]+/g, '_').replace(/^_+|_+$/g, '').replace(/^(\d)/, 'f_$1').slice(0, 60)

const syncKey = () => {
  if (!keyTouched.value) form.key = slugify(form.label)
}

const submit = async () => {
  if (!form.label) {
    error.value = t('directory.errors.labelRequired')
    return
  }
  const options = optionsText.value.split('\n').map((o) => o.trim()).filter(Boolean)
  if (form.type === 'select' && !options.length) {
    error.value = t('directory.errors.optionsRequired')
    return
  }
  saving.value = true
  error.value = ''
  try {
    const payload = { ...form, options, order: Number(form.order) || 0 }
    const saved = props.field
      ? await directoryService.updateField(props.field.id, payload)
      : await directoryService.createField(payload)
    emit('saved', saved)
  } catch (e) {
    error.value = apiError(e, t('directory.errors.saveFailed'))
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.field-type {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.65rem;
  border: 1px solid var(--border);
  border-radius: 0.55rem;
  font-size: 0.85rem;
  color: var(--text-primary);
  cursor: pointer;
  transition: border-color 150ms ease, background-color 150ms ease;
}
.field-type:hover { background: var(--hover); }
.field-type:focus-within { outline: 2px solid var(--accent); outline-offset: 1px; }
.field-type--active {
  border-color: var(--accent);
  background: color-mix(in srgb, var(--accent) 10%, transparent);
  color: var(--accent);
  font-weight: 600;
}
</style>
