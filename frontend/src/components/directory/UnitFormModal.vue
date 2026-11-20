<template>
  <DirModal
    :title="unit ? $t('directory.editUnit') : $t('directory.newUnit')"
    :saving="saving"
    :error="error"
    size="md"
    @close="$emit('close')"
    @submit="submit"
  >
    <div class="grid gap-x-6 md:grid-cols-2">
      <div>
        <div class="form-group">
          <label class="form-label form-label-required" for="unit-name">{{ $t('directory.unitName') }}</label>
          <input id="unit-name" v-model.trim="form.name" class="form-input" maxlength="200" required autofocus />
        </div>
        <div class="form-group">
          <label class="form-label" for="unit-code">{{ $t('directory.unitCode') }}</label>
          <input id="unit-code" v-model.trim="form.code" class="form-input" maxlength="50" :placeholder="$t('directory.unitCodePlaceholder')" />
        </div>
        <div class="form-group">
          <label class="form-label" for="unit-parent">{{ $t('directory.parentUnit') }}</label>
          <select id="unit-parent" v-model="form.parent_id" class="form-select">
            <option :value="null">{{ $t('directory.noParent') }}</option>
            <option v-for="u in parentOptions" :key="u.id" :value="u.id">{{ '\u00a0\u00a0'.repeat(u.depth) }}{{ u.name }}</option>
          </select>
        </div>
      </div>
      <div>
        <div class="grid grid-cols-2 gap-4">
          <div class="form-group">
            <label class="form-label" for="unit-color">{{ $t('directory.color') }}</label>
            <div class="flex items-center gap-2">
              <input id="unit-color" v-model="form.color" type="color" class="h-10 w-12 cursor-pointer rounded border" style="border-color: var(--border)" />
              <input v-model="form.color" class="form-input" maxlength="20" aria-label="hex" />
            </div>
          </div>
          <div class="form-group">
            <label class="form-label" for="unit-order">{{ $t('directory.order') }}</label>
            <input id="unit-order" v-model.number="form.order" type="number" class="form-input" />
          </div>
        </div>
        <div class="form-group">
          <label class="form-label" for="unit-desc">{{ $t('directory.description') }}</label>
          <textarea id="unit-desc" v-model="form.description" class="form-textarea" rows="4" maxlength="2000" />
        </div>
      </div>
    </div>
  </DirModal>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DirModal from './DirModal.vue'
import { directoryService } from '@/services/api'
import { apiError } from './useDirectory'

const props = defineProps({
  unit: { type: Object, default: null },
  parentId: { type: Number, default: null },
  flatUnits: { type: Array, required: true },
  descendantIds: { type: Function, required: true },
})
const emit = defineEmits(['close', 'saved'])
const { t } = useI18n()

const form = reactive({
  name: props.unit?.name || '',
  code: props.unit?.code || '',
  description: props.unit?.description || '',
  color: props.unit?.color || '#3B82F6',
  parent_id: props.unit ? props.unit.parent_id ?? null : props.parentId,
  order: props.unit?.order ?? 0,
})
const saving = ref(false)
const error = ref('')

// Une unité ne peut pas être rattachée à elle-même ni à ses descendants
const parentOptions = computed(() => {
  if (!props.unit) return props.flatUnits
  const excluded = props.descendantIds(props.unit.id)
  return props.flatUnits.filter((u) => !excluded.has(u.id))
})

const submit = async () => {
  if (!form.name) {
    error.value = t('directory.errors.nameRequired')
    return
  }
  saving.value = true
  error.value = ''
  try {
    const payload = { ...form, order: Number(form.order) || 0 }
    const saved = props.unit
      ? await directoryService.updateUnit(props.unit.id, payload)
      : await directoryService.createUnit(payload)
    emit('saved', saved)
  } catch (e) {
    error.value = apiError(e, t('directory.errors.saveFailed'))
  } finally {
    saving.value = false
  }
}
</script>
