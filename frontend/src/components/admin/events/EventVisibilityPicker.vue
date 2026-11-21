<template>
  <!-- Visibilité d'un événement : public, ou réservé à des groupes choisis parmi `groups` -->
  <div class="card bg-gradient-to-br from-primary-50 to-primary-100 dark:from-primary-900/20 dark:to-primary-800/20 border-2 border-primary-200 dark:border-primary-800">
    <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
      <Icon icon="mdi:account-group" class="h-5 w-5 text-primary-500" />
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
            <Icon icon="mdi:earth" class="inline h-4 w-4 me-1" />
            {{ t('groupAdmin.eventEditor.publicEvent') }}
          </span>
        </label>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 ms-8">
          {{ t('groupAdmin.eventEditor.publicEventHelp') }}
        </p>
      </div>

      <!-- Groupes ciblés (événement non public) -->
      <div v-if="!isPublic" class="space-y-2">
        <div
          v-if="groupIds.length > 0"
          class="flex flex-wrap gap-2 p-3 bg-white dark:bg-gray-800 rounded-lg border border-primary-200 dark:border-primary-700"
        >
          <span
            v-for="groupId in groupIds"
            :key="groupId"
            class="inline-flex items-center gap-1 px-3 py-1.5 bg-primary-100 dark:bg-primary-900/30 text-primary-800 dark:text-primary-300 rounded-full text-sm font-medium"
          >
            <Icon icon="mdi:account-group" class="h-3 w-3" />
            {{ groupName(groupId) }}
            <button
              type="button"
              class="ms-1 hover:text-red-600 dark:hover:text-red-400 transition-colors"
              :aria-label="t('common.delete')"
              @click="removeGroup(groupId)"
            >
              <Icon icon="mdi:close-circle" class="h-4 w-4" />
            </button>
          </span>
        </div>
        <div
          v-else
          class="text-sm text-orange-500 dark:text-orange-400 italic p-3 bg-white dark:bg-gray-800 rounded-lg border border-orange-200 dark:border-orange-700"
        >
          {{ t('groupAdmin.eventEditor.noGroupSelected') }}
        </div>

        <select
          class="w-full px-3 py-2 border-2 border-primary-200 dark:border-primary-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm"
          @change="addGroup"
        >
          <option value="">{{ t('groupAdmin.eventEditor.addGroup') }}</option>
          <option v-for="group in selectableGroups" :key="group.id" :value="group.id">
            {{ group.name }}
          </option>
        </select>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'

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

const selectableGroups = computed(() => props.groups.filter((group) => !props.groupIds.includes(group.id)))

const groupName = (groupId) => props.groups.find((group) => group.id === groupId)?.name || ''

const addGroup = (event) => {
  const groupId = parseInt(event.target.value, 10)
  if (groupId && !props.groupIds.includes(groupId)) {
    emit('update:groupIds', [...props.groupIds, groupId])
  }
  // Remet le menu sur « + Ajouter un groupe »
  event.target.value = ''
}

const removeGroup = (groupId) => {
  emit('update:groupIds', props.groupIds.filter((id) => id !== groupId))
}
</script>
