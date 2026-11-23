<template>
  <DirModal
    :title="$t('directory.groups.title', { name: unit.name })"
    :subtitle="$t('directory.groups.subtitle')"
    :saving="saving"
    :error="error"
    :submit-label="submitLabel"
    size="md"
    @close="$emit('close')"
    @submit="submit"
  >
    <div v-if="currentGroup" class="ugm-current" role="status">
      <Icon icon="lucide:link" class="ugm-current__icon" aria-hidden="true" />
      <span>{{ $t('directory.groups.current', { name: currentGroup.name }) }}</span>
    </div>

    <fieldset class="ugm-choices">
      <legend class="sr-only">{{ $t('directory.groups.choiceLegend') }}</legend>

      <label class="ugm-choice" :class="{ 'ugm-choice--active': mode === 'create' }">
        <input v-model="mode" type="radio" value="create" name="ugm-mode" />
        <span class="ugm-choice__body">
          <span class="ugm-choice__title">{{ $t('directory.groups.createTitle') }}</span>
          <span class="ugm-choice__desc">{{ $t('directory.groups.createDesc', branchSize) }}</span>
        </span>
      </label>

      <label class="ugm-choice" :class="{ 'ugm-choice--active': mode === 'link' }">
        <input v-model="mode" type="radio" value="link" name="ugm-mode" />
        <span class="ugm-choice__body">
          <span class="ugm-choice__title">{{ $t('directory.groups.linkTitle') }}</span>
          <span class="ugm-choice__desc">{{ $t('directory.groups.linkDesc') }}</span>
        </span>
      </label>
      <div v-if="mode === 'link'" class="form-group ugm-select">
        <label for="ugm-group" class="form-label form-label-required">{{ $t('directory.groups.group') }}</label>
        <select id="ugm-group" v-model="groupId" class="form-select" required>
          <option :value="null" disabled>{{ $t('directory.groups.pick') }}</option>
          <option v-for="row in linkable" :key="row.group.id" :value="row.group.id">{{ indentedLabel(row) }}</option>
        </select>
        <p v-if="!linkable.length" class="form-help">{{ $t('directory.groups.noFreeGroup') }}</p>
      </div>

      <label v-if="currentGroup" class="ugm-choice" :class="{ 'ugm-choice--active': mode === 'unlink' }">
        <input v-model="mode" type="radio" value="unlink" name="ugm-mode" />
        <span class="ugm-choice__body">
          <span class="ugm-choice__title">{{ $t('directory.groups.unlinkTitle') }}</span>
          <span class="ugm-choice__desc">{{ $t('directory.groups.unlinkDesc') }}</span>
        </span>
      </label>
    </fieldset>

    <p class="ugm-note">
      <Icon icon="lucide:info" class="ugm-note__icon" aria-hidden="true" />
      <span>{{ $t('directory.groups.rules') }}</span>
    </p>
  </DirModal>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import DirModal from './DirModal.vue'
import { directoryService } from '@/services/api'
import { apiError } from './useDirectory'
import { orderGroupTree, indentedLabel } from '@/utils/groupTree'

const props = defineProps({
  unit: { type: Object, required: true },
  // Groupes d'utilisateurs (liste de l'admin : org_unit_id indique l'entité déjà liée)
  groups: { type: Array, required: true },
  // Nombre d'entités de la branche (l'entité et ses sous-entités)
  branchSize: { type: Number, default: 1 },
})
const emit = defineEmits(['close', 'saved'])
const { t } = useI18n()

const currentGroup = computed(() => props.groups.find((g) => g.id === props.unit.group_id) || null)
// Groupes libres (ou celui de l'entité), hors groupe par défaut
const linkable = computed(() =>
  orderGroupTree(props.groups).filter(
    ({ group }) => !group.is_default && (group.org_unit_id == null || group.org_unit_id === props.unit.id),
  ),
)

const mode = ref(props.unit.group_id ? 'link' : 'create')
const groupId = ref(props.unit.group_id ?? null)
const saving = ref(false)
const error = ref('')

const submitLabel = computed(() => {
  if (mode.value === 'create') return t('directory.groups.createSubmit')
  if (mode.value === 'unlink') return t('directory.groups.unlinkSubmit')
  return t('directory.groups.linkSubmit')
})

const syncSummary = (sync) =>
  sync && (sync.added || sync.removed) ? ` ${t('directory.groups.syncSummary', { added: sync.added || 0, removed: sync.removed || 0 })}` : ''

const submit = async () => {
  error.value = ''
  if (mode.value === 'link' && !groupId.value) {
    error.value = t('directory.groups.pickRequired')
    return
  }
  saving.value = true
  try {
    if (mode.value === 'create') {
      const res = await directoryService.createUnitGroups(props.unit.id)
      emit('saved', t('directory.groups.created', { created: res.created || 0, linked: res.linked || 0 }) + syncSummary(res.sync))
    } else {
      const target = mode.value === 'unlink' ? null : groupId.value
      const res = await directoryService.linkUnitGroup(props.unit.id, target)
      emit('saved', t(target ? 'directory.groups.linked' : 'directory.groups.unlinked') + syncSummary(res.sync))
    }
  } catch (e) {
    error.value = apiError(e, t('directory.errors.saveFailed'))
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.ugm-current {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-block-end: 14px;
  padding: 9px 12px;
  border-radius: var(--r-s, 8px);
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
  background: color-mix(in srgb, var(--acc) 10%, transparent);
}

.ugm-current__icon { width: 16px; height: 16px; color: var(--acc); flex: none; }

.ugm-choices { display: flex; flex-direction: column; gap: 8px; margin: 0; padding: 0; border: 0; }

.ugm-choice {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 11px 12px;
  border: 1px solid var(--border);
  border-radius: var(--r-s, 8px);
  background: var(--bg-surface);
  cursor: pointer;
  transition: border-color 0.15s ease, background-color 0.15s ease;
}

.ugm-choice:hover { border-color: var(--acc); }
.ugm-choice:focus-within { outline: 2px solid var(--acc); outline-offset: 2px; }

.ugm-choice--active {
  border-color: var(--acc);
  background: color-mix(in srgb, var(--acc) 7%, var(--bg-surface));
}

.ugm-choice input { margin-top: 3px; accent-color: var(--acc); }
.ugm-choice__body { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.ugm-choice__title { font-size: 14px; font-weight: 600; color: var(--text); }
.ugm-choice__desc { font-size: 12.5px; color: var(--muted); }

.ugm-select { margin: 2px 0 4px; padding-inline-start: 30px; }

.ugm-note {
  display: flex;
  gap: 8px;
  margin-top: 14px;
  font-size: 12.5px;
  line-height: 1.5;
  color: var(--muted);
}

.ugm-note__icon { width: 15px; height: 15px; margin-top: 2px; flex: none; }

@media (prefers-reduced-motion: reduce) {
  .ugm-choice { transition: none; }
}
</style>
