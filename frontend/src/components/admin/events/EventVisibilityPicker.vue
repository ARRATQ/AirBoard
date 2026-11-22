<template>
  <!-- Visibilité d'un événement : public, ou réservé à des groupes choisis parmi `groups` -->
  <div class="card bg-gradient-to-br from-primary-50 to-primary-100 dark:from-primary-900/20 dark:to-primary-800/20 border-2 border-primary-200 dark:border-primary-800">
    <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
      <Icon icon="lucide:users" class="h-5 w-5 text-primary-500" />
      {{ t('events.editor.visibility') }}
    </h3>

    <div class="space-y-3">
      <p class="text-xs text-gray-600 dark:text-gray-400">
        {{ t('groupAdmin.eventEditor.targetGroupsHelp') }}
      </p>

      <!-- Événement public -->
      <div class="bg-white dark:bg-gray-800 rounded-lg p-3 border border-primary-200 dark:border-primary-700">
        <label class="flex items-center cursor-pointer">
          <input
            :checked="isPublic"
            type="checkbox"
            class="form-checkbox h-5 w-5 text-primary-600 rounded focus:ring-2 focus:ring-primary-500"
            @change="emit('update:isPublic', $event.target.checked)"
          />
          <span class="ms-3 text-sm font-medium text-gray-900 dark:text-white">
            <Icon icon="lucide:earth" class="inline h-4 w-4 me-1" />
            {{ t('groupAdmin.eventEditor.publicEvent') }}
          </span>
        </label>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 ms-8">
          {{ t('groupAdmin.eventEditor.publicEventHelp') }}
        </p>
      </div>

      <!-- Groupes ciblés (événement non public) -->
      <div v-if="!isPublic" class="space-y-2">
        <ComboSelect
          :model-value="groupIds"
          kind="group"
          :options="groupOptions"
          :label="t('events.editor.visibility')"
          :placeholder="t('polls.editor.targetGroupsPlaceholder')"
          @update:model-value="emit('update:groupIds', $event)"
        />
        <p v-if="groupIds.length === 0" class="text-sm text-orange-600 dark:text-orange-400" role="status">
          {{ t('groupAdmin.eventEditor.noGroupSelected') }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import ComboSelect from '@/components/ui/ComboSelect.vue'
import { groupOption } from '@/utils/comboFilter'

const props = defineProps({
  // v-model:groupIds — identifiants des groupes ciblés
  groupIds: {
    type: Array,
    default: () => []
  },
  // v-model:isPublic
  isPublic: {
    type: Boolean,
    default: true
  },
  // Groupes proposés (tous les groupes pour l'admin, les groupes administrés pour l'admin de groupe)
  groups: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:groupIds', 'update:isPublic'])

const { t } = useI18n()

const groupOptions = computed(() => props.groups.map((group) => groupOption(group)))
</script>
