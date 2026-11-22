<template>
  <!-- Ciblage par groupe : obligatoire pour l'admin de groupe (au moins un de ses groupes) -->
  <div class="card bg-gradient-to-br from-primary-50 to-primary-100 dark:from-primary-900/20 dark:to-primary-800/20 border-2 border-primary-200 dark:border-primary-800">
    <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
      <Icon icon="lucide:users" class="h-5 w-5 text-primary-500" />
      {{ $t('groupAdmin.newsEditor.targetGroups') }} {{ $t('groupAdmin.newsEditor.targetGroupsRequired') }}
    </h3>

    <div class="space-y-3">
      <p class="text-xs text-gray-600 dark:text-gray-400">
        {{ $t('groupAdmin.newsEditor.targetGroupsHelp') }}
      </p>

      <ComboSelect
        :model-value="modelValue"
        kind="group"
        :options="groupOptions"
        :labels="lockedOptions"
        :locked-ids="lockedIds"
        :label="$t('groupAdmin.newsEditor.targetGroups')"
        :placeholder="$t('polls.editor.targetGroupsPlaceholder')"
        @update:model-value="emit('update:modelValue', $event)"
      />
      <p v-if="modelValue.length === 0" class="text-sm text-red-500 dark:text-red-400" role="status">
        {{ $t('groupAdmin.newsEditor.noGroupSelected') }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import ComboSelect from '@/components/ui/ComboSelect.vue'
import { groupOption } from '@/utils/comboFilter'

const props = defineProps({
  // Identifiants des groupes ciblés
  modelValue: { type: Array, required: true },
  // Groupes que l'utilisateur administre
  groups: { type: Array, required: true },
  // Groupes déjà ciblés qu'il n'administre pas : affichés, mais non retirables
  lockedGroups: { type: Array, default: () => [] }
})

const emit = defineEmits(['update:modelValue'])

const groupOptions = computed(() => props.groups.map((group) => groupOption(group)))
// Groupes déjà ciblés hors de ceux qu'on gère : affichés, mais non retirables
const lockedOptions = computed(() => props.lockedGroups.map((group) => groupOption(group)))
const lockedIds = computed(() => {
  const managed = new Set(props.groups.map((g) => g.id))
  return props.modelValue.filter((id) => !managed.has(id))
})
</script>
