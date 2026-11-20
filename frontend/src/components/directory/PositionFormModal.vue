<template>
  <DirModal
    :title="position ? $t('directory.editPosition') : $t('directory.newPosition')"
    :saving="saving"
    :error="error"
    size="sm"
    @close="$emit('close')"
    @submit="submit"
  >
    <div class="form-group">
      <label class="form-label form-label-required" for="pos-title">{{ $t('directory.positionTitle') }}</label>
      <input id="pos-title" v-model.trim="form.title" class="form-input" maxlength="200" required autofocus :placeholder="$t('directory.positionTitlePlaceholder')" />
    </div>
    <div class="grid grid-cols-3 gap-4">
      <div class="form-group col-span-2">
        <label class="form-label form-label-required" for="pos-unit">{{ $t('directory.unit') }}</label>
        <select id="pos-unit" v-model="form.unit_id" class="form-select" required>
          <option v-for="u in flatUnits" :key="u.id" :value="u.id">{{ '\u00a0\u00a0'.repeat(u.depth) }}{{ u.name }}</option>
        </select>
      </div>
      <div class="form-group">
        <label class="form-label" for="pos-order">{{ $t('directory.order') }}</label>
        <input id="pos-order" v-model.number="form.order" type="number" class="form-input" />
      </div>
    </div>
    <label class="flex items-start gap-3 cursor-pointer">
      <input v-model="form.is_head" type="checkbox" class="mt-1" />
      <span>
        <span class="block text-sm font-medium" style="color: var(--text-primary)">{{ $t('directory.isHead') }}</span>
        <span class="block text-xs" style="color: var(--text-muted)">{{ $t('directory.isHeadHelp') }}</span>
      </span>
    </label>
  </DirModal>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DirModal from './DirModal.vue'
import { directoryService } from '@/services/api'
import { apiError } from './useDirectory'

const props = defineProps({
  position: { type: Object, default: null },
  unitId: { type: Number, default: null },
  flatUnits: { type: Array, required: true },
})
const emit = defineEmits(['close', 'saved'])
const { t } = useI18n()

const form = reactive({
  title: props.position?.title || '',
  unit_id: props.position?.unit_id ?? props.unitId ?? props.flatUnits[0]?.id ?? null,
  is_head: props.position?.is_head ?? false,
  order: props.position?.order ?? 0,
})
const saving = ref(false)
const error = ref('')

const submit = async () => {
  if (!form.title || !form.unit_id) {
    error.value = t('directory.errors.positionRequired')
    return
  }
  saving.value = true
  error.value = ''
  try {
    const payload = { ...form, order: Number(form.order) || 0 }
    const saved = props.position
      ? await directoryService.updatePosition(props.position.id, payload)
      : await directoryService.createPosition(payload)
    emit('saved', saved)
  } catch (e) {
    error.value = apiError(e, t('directory.errors.saveFailed'))
  } finally {
    saving.value = false
  }
}
</script>
