<template>
  <!-- Ciblage par groupe : obligatoire pour l'admin de groupe (au moins un de ses groupes) -->
  <div class="card bg-gradient-to-br from-primary-50 to-primary-100 dark:from-primary-900/20 dark:to-primary-800/20 border-2 border-primary-200 dark:border-primary-800">
    <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
      <Icon icon="mdi:account-group" class="h-5 w-5 text-primary-500" />
      {{ $t('groupAdmin.newsEditor.targetGroups') }} {{ $t('groupAdmin.newsEditor.targetGroupsRequired') }}
    </h3>

    <div class="space-y-3">
      <p class="text-xs text-gray-600 dark:text-gray-400">
        {{ $t('groupAdmin.newsEditor.targetGroupsHelp') }}
      </p>

      <div
        v-if="modelValue.length > 0"
        class="flex flex-wrap gap-2 p-3 bg-white dark:bg-gray-800 rounded-lg border border-primary-200 dark:border-primary-700"
      >
        <span
          v-for="groupId in modelValue"
          :key="groupId"
          class="inline-flex items-center gap-1 px-3 py-1.5 bg-primary-100 dark:bg-primary-900/30 text-primary-800 dark:text-primary-300 rounded-full text-sm font-medium"
        >
          <Icon icon="mdi:account-group" class="h-3 w-3" />
          {{ groupName(groupId) }}
          <Icon
            v-if="isLocked(groupId)"
            icon="mdi:lock-outline"
            class="h-3.5 w-3.5 opacity-70"
            :title="$t('adminShared.news.lockedGroup')"
            :aria-label="$t('adminShared.news.lockedGroup')"
          />
          <button
            v-else
            type="button"
            class="ml-1 hover:text-red-600 dark:hover:text-red-400 transition-colors"
            :aria-label="$t('adminShared.news.removeItem', { name: groupName(groupId) })"
            @click="removeGroup(groupId)"
          >
            <Icon icon="mdi:close-circle" class="h-4 w-4" />
          </button>
        </span>
      </div>
      <div
        v-else
        class="text-sm text-red-500 dark:text-red-400 italic p-3 bg-white dark:bg-gray-800 rounded-lg border border-red-200 dark:border-red-700"
      >
        {{ $t('groupAdmin.newsEditor.noGroupSelected') }}
      </div>

      <select
        class="w-full px-3 py-2 border-2 border-primary-200 dark:border-primary-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm"
        :aria-label="$t('groupAdmin.newsEditor.addGroup')"
        @change="addGroup"
      >
        <option value="">{{ $t('groupAdmin.newsEditor.addGroup') }}</option>
        <option v-for="group in availableGroups" :key="group.id" :value="group.id">
          {{ group.name }}
        </option>
      </select>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps({
  // Identifiants des groupes ciblés
  modelValue: { type: Array, required: true },
  // Groupes que l'utilisateur administre
  groups: { type: Array, required: true },
  // Groupes déjà ciblés qu'il n'administre pas : affichés, mais non retirables
  lockedGroups: { type: Array, default: () => [] }
})

const emit = defineEmits(['update:modelValue'])

const availableGroups = computed(() =>
  props.groups.filter((group) => !props.modelValue.includes(group.id))
)

const findGroup = (groupId) =>
  props.groups.find((g) => g.id === groupId) || props.lockedGroups.find((g) => g.id === groupId)

const groupName = (groupId) => findGroup(groupId)?.name || `#${groupId}`
const isLocked = (groupId) => !props.groups.some((g) => g.id === groupId)

const addGroup = (event) => {
  const groupId = parseInt(event.target.value, 10)
  if (groupId && !props.modelValue.includes(groupId)) {
    emit('update:modelValue', [...props.modelValue, groupId])
  }
  event.target.value = ''
}

const removeGroup = (groupId) => {
  emit('update:modelValue', props.modelValue.filter((id) => id !== groupId))
}
</script>
