<template>
  <DirModal
    :title="group ? $t('myGroups.modal.editTitle', { name: group.name }) : scopeAll ? $t('myGroups.modal.createTitle') : $t('myGroups.newSubGroup')"
    :subtitle="scopeAll ? $t('myGroups.modal.subtitleAll') : $t('myGroups.modal.subtitleGroups')"
    :saving="saving"
    :error="error"
    :submit-label="group ? $t('common.save') : $t('myGroups.modal.create')"
    size="md"
    @close="$emit('close')"
    @submit="submit"
  >
    <div class="grid grid-cols-1 sm:grid-cols-[1fr_9rem] gap-4">
      <div class="form-group mb-0">
        <label for="sg-name" class="form-label form-label-required">{{ $t('myGroups.modal.name') }}</label>
        <input id="sg-name" v-model="form.name" class="form-input" maxlength="100" required autofocus />
      </div>
      <div class="form-group mb-0">
        <label for="sg-color" class="form-label">{{ $t('myGroups.modal.color') }}</label>
        <input id="sg-color" v-model="form.color" type="color" class="sg-color" />
      </div>
    </div>

    <div class="form-group mt-4 mb-0">
      <label for="sg-parent" class="form-label" :class="{ 'form-label-required': !scopeAll }">{{ $t('myGroups.modal.parent') }}</label>
      <select id="sg-parent" v-model="form.parent_id" class="form-select" :required="!scopeAll" @change="loadCandidates">
        <option v-if="scopeAll" :value="null">{{ $t('groupModal.parentNone') }}</option>
        <option v-else :value="null" disabled>{{ $t('myGroups.modal.pickParent') }}</option>
        <option v-for="row in parents" :key="row.group.id" :value="row.group.id">{{ indentedLabel(row) }}</option>
      </select>
      <p class="form-help">{{ $t('groupModal.parentHint') }}</p>
    </div>

    <div class="form-group mt-4 mb-0">
      <label for="sg-description" class="form-label">{{ $t('common.description') }}</label>
      <textarea id="sg-description" v-model="form.description" rows="2" class="form-textarea resize-none" maxlength="500"></textarea>
    </div>

    <div class="form-group mt-4 mb-0">
      <label for="sg-members-input" class="form-label">{{ $t('myGroups.modal.members') }}</label>
      <ComboSelect
        v-model="form.user_ids"
        input-id="sg-members-input"
        kind="person"
        :options="candidates"
        :labels="memberLabels"
        :search="searchCandidates"
        :label="$t('myGroups.modal.members')"
        :placeholder="$t('myGroups.modal.membersPlaceholder')"
        :empty-text="!scopeAll && !form.parent_id ? $t('myGroups.modal.pickParentFirst') : ''"
      />
      <p class="form-help">{{ scopeAll ? $t('myGroups.modal.membersHintAll') : $t('myGroups.modal.membersHintGroups') }}</p>
    </div>
  </DirModal>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DirModal from '@/components/directory/DirModal.vue'
import ComboSelect from '@/components/ui/ComboSelect.vue'
import { userGroupsService } from '@/services/api'
import { personOption } from '@/utils/comboFilter'
import { parentCandidates, indentedLabel } from '@/utils/groupTree'

const props = defineProps({
  // Groupe modifié (null : création)
  group: { type: Object, default: null },
  // Parent proposé à la création
  parentId: { type: Number, default: null },
  // Groupes visibles avec la permission (liste de la page)
  groups: { type: Array, required: true },
  // Portée « tout le site » : groupe principal possible et n'importe quel compte
  scopeAll: { type: Boolean, default: false },
})
const emit = defineEmits(['close', 'saved'])
const { t } = useI18n()

const form = reactive({
  name: props.group?.name || '',
  color: props.group?.color || '#6B7280',
  description: props.group?.description || '',
  parent_id: props.group ? props.group.parent_id ?? null : props.parentId,
  user_ids: (props.group?.members || []).map((m) => m.id),
})
const saving = ref(false)
const error = ref('')
const candidates = ref([])

// Parents possibles : groupes visibles, sans le groupe lui-même ni ses sous-groupes
const parents = computed(() => parentCandidates(props.groups, props.group?.id ?? null))
const memberLabels = computed(() => (props.group?.members || []).map(personOption))

const loadCandidates = async () => {
  if (!props.scopeAll && !form.parent_id) {
    candidates.value = []
    return
  }
  try {
    const list = await userGroupsService.candidates(form.parent_id, '')
    candidates.value = (Array.isArray(list) ? list : []).map(personOption)
  } catch {
    candidates.value = []
  }
}

const searchCandidates = async (text) => {
  if (!props.scopeAll && !form.parent_id) return []
  const list = await userGroupsService.candidates(form.parent_id, text)
  return (Array.isArray(list) ? list : []).map(personOption)
}

onMounted(loadCandidates)

const submit = async () => {
  error.value = ''
  if (!form.name.trim()) {
    error.value = t('myGroups.errors.nameRequired')
    return
  }
  if (!props.scopeAll && !form.parent_id) {
    error.value = t('myGroups.errors.parentRequired')
    return
  }
  saving.value = true
  const payload = {
    name: form.name.trim(),
    color: form.color,
    description: form.description,
    parent_id: form.parent_id,
    user_ids: [...form.user_ids],
  }
  try {
    if (props.group) await userGroupsService.update(props.group.id, payload)
    else await userGroupsService.create(payload)
    emit('saved', props.group ? t('myGroups.updated') : t('myGroups.created'))
  } catch (e) {
    error.value = e?.response?.data?.error || t('myGroups.errors.save')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.sg-color {
  width: 100%;
  height: 38px;
  padding: 3px;
  border: 1px solid var(--border);
  border-radius: var(--r-s, 8px);
  background: var(--bg-card);
  cursor: pointer;
}
</style>
