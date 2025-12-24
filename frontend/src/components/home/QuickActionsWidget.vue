<template>
  <div class="card p-6">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
      <Icon icon="mdi:lightning-bolt" class="h-5 w-5 text-yellow-500" />
      {{ $t('home.quickActions.title') }}
    </h3>

    <!-- Admin Actions -->
    <div v-if="role === 'admin'" class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <ActionButton
        icon="mdi:account-multiple"
        :label="$t('home.quickActions.manageUsers')"
        to="/admin/users"
        color="blue"
      />
      <ActionButton
        icon="mdi:account-group"
        :label="$t('home.quickActions.manageGroups')"
        to="/admin/groups"
        color="purple"
      />
      <ActionButton
        icon="mdi:application"
        :label="$t('home.quickActions.manageApps')"
        to="/admin/applications"
        color="green"
      />
      <ActionButton
        icon="mdi:bullhorn"
        :label="$t('home.quickActions.createAnnouncement')"
        to="/admin/announcements"
        color="orange"
      />
    </div>

    <!-- Group Admin Actions -->
    <div v-else-if="role === 'group_admin'" class="grid grid-cols-2 md:grid-cols-3 gap-3">
      <ActionButton
        icon="mdi:account-group"
        :label="$t('home.quickActions.myGroups')"
        to="/group-admin"
        color="indigo"
      />
      <ActionButton
        icon="mdi:application"
        :label="$t('home.quickActions.myApps')"
        to="/group-admin/applications"
        color="teal"
      />
      <ActionButton
        icon="mdi:newspaper"
        :label="$t('home.quickActions.createNews')"
        to="/admin/news"
        color="orange"
      />
    </div>

    <!-- Editor Actions -->
    <div v-else-if="role === 'editor'" class="grid grid-cols-2 gap-3">
      <ActionButton
        icon="mdi:newspaper"
        :label="$t('home.quickActions.createNews')"
        to="/admin/news"
        color="orange"
      />
      <ActionButton
        icon="mdi:calendar-plus"
        :label="$t('home.quickActions.createEvent')"
        to="/admin/events"
        color="blue"
      />
    </div>
  </div>
</template>

<script setup>
import { Icon } from '@iconify/vue'
import ActionButton from './ActionButton.vue'

defineProps({
  role: {
    type: String,
    required: true
  }
})
</script>
