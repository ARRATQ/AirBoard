<template>
  <DirModal
    :title="assignment ? $t('directory.editAssignment') : $t('directory.assignPerson')"
    :subtitle="position.title"
    :saving="saving"
    :error="error"
    size="sm"
    @close="$emit('close')"
    @submit="submit"
  >
    <div class="form-group">
      <label class="form-label form-label-required" for="asg-person">{{ $t('directory.person') }}</label>
      <template v-if="assignment">
        <p class="form-input" style="background: var(--bg-surface)">{{ currentPersonName }}</p>
      </template>
      <template v-else>
        <input
          v-model="query"
          type="search"
          class="form-input mb-2"
          :placeholder="$t('directory.searchPerson')"
          :aria-label="$t('directory.searchPerson')"
        />
        <select id="asg-person" v-model="form.person_id" class="form-select" size="6" required>
          <option v-for="p in candidates" :key="p.id" :value="p.id">
            {{ p.last_name }} {{ p.first_name }}{{ p.email ? ` — ${p.email}` : '' }}
          </option>
        </select>
        <p v-if="!candidates.length" class="text-xs mt-1" style="color: var(--text-muted)">{{ $t('directory.noPersonFound') }}</p>
      </template>
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div class="form-group">
        <label class="form-label" for="asg-start">{{ $t('directory.startDate') }}</label>
        <input id="asg-start" v-model="form.start_date" type="date" class="form-input" />
      </div>
      <div class="form-group">
        <label class="form-label" for="asg-order">{{ $t('directory.order') }}</label>
        <input id="asg-order" v-model.number="form.order" type="number" class="form-input" />
      </div>
    </div>
    <label class="flex items-start gap-3 cursor-pointer">
      <input v-model="form.is_interim" type="checkbox" class="mt-1" />
      <span>
        <span class="block text-sm font-medium" style="color: var(--text-primary)">{{ $t('directory.interim') }}</span>
        <span class="block text-xs" style="color: var(--text-muted)">{{ $t('directory.interimHelp') }}</span>
      </span>
    </label>
  </DirModal>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DirModal from './DirModal.vue'
import { directoryService } from '@/services/api'
import { apiError, fullName } from './useDirectory'

const props = defineProps({
  position: { type: Object, required: true },
  assignment: { type: Object, default: null },
  people: { type: Array, required: true },
})
const emit = defineEmits(['close', 'saved'])
const { t } = useI18n()

const form = reactive({
  person_id: props.assignment?.person_id ?? null,
  is_interim: props.assignment?.is_interim ?? false,
  start_date: props.assignment?.start_date ? props.assignment.start_date.slice(0, 10) : '',
  order: props.assignment?.order ?? 0,
})
const query = ref('')
const saving = ref(false)
const error = ref('')

const currentPersonName = computed(() => fullName(props.people.find((p) => p.id === form.person_id)))

const normalize = (s) => (s || '').normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase()

// Exclut les personnes déjà titulaires de ce poste
const candidates = computed(() => {
  const taken = new Set((props.position.assignments || []).map((a) => a.person_id))
  const q = normalize(query.value)
  return props.people
    .filter((p) => !taken.has(p.id))
    .filter((p) => !q || normalize(`${fullName(p)} ${p.email}`).includes(q))
    .sort((a, b) => a.last_name.localeCompare(b.last_name))
})

const submit = async () => {
  if (!form.person_id) {
    error.value = t('directory.errors.personRequired')
    return
  }
  saving.value = true
  error.value = ''
  try {
    const payload = { position_id: props.position.id, ...form, order: Number(form.order) || 0 }
    const saved = props.assignment
      ? await directoryService.updateAssignment(props.assignment.id, payload)
      : await directoryService.createAssignment(payload)
    emit('saved', saved)
  } catch (e) {
    error.value = apiError(e, t('directory.errors.saveFailed'))
  } finally {
    saving.value = false
  }
}
</script>
