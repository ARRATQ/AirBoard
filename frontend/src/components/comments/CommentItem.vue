<template>
  <div class="comment-item bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700">
    <!-- Header -->
    <div class="flex items-start justify-between mb-2">
      <div class="flex items-center gap-3">
        <div class="flex-shrink-0">
          <div class="h-10 w-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-600
                      flex items-center justify-center text-white font-semibold">
            {{ getUserInitials(comment.user) }}
          </div>
        </div>
        <div>
          <p class="font-medium text-gray-900 dark:text-white">
            {{ getUserName(comment.user) }}
          </p>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ formatDate(comment.created_at) }}
            <span v-if="comment.created_at !== comment.updated_at" class="ml-1">
              ({{ $t('comments.edited') }})
            </span>
          </p>
        </div>
      </div>

      <!-- Actions -->
      <div v-if="canManage" class="flex items-center gap-2">
        <button
          v-if="canEdit"
          class="text-gray-500 hover:text-blue-600 transition-colors"
          :title="$t('comments.edit')"
          @click="$emit('edit', comment)"
        >
          <Icon icon="mdi:pencil" class="h-5 w-5" />
        </button>
        <button
          v-if="canDelete"
          class="text-gray-500 hover:text-red-600 transition-colors"
          :title="$t('comments.delete')"
          @click="$emit('delete', comment.id)"
        >
          <Icon icon="mdi:delete" class="h-5 w-5" />
        </button>
      </div>
    </div>

    <!-- Content -->
    <div class="ml-13">
      <p class="text-gray-700 dark:text-gray-300 whitespace-pre-wrap">
        {{ comment.content }}
      </p>

      <!-- Moderation badges -->
      <div v-if="showModerationBadges" class="flex items-center gap-2 mt-2">
        <span
          v-if="!comment.is_approved"
          class="px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300"
        >
          <Icon icon="mdi:clock-outline" class="inline h-3 w-3 mr-1" />
          {{ $t('comments.pending') }}
        </span>
        <span
          v-if="comment.is_flagged"
          class="px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300"
        >
          <Icon icon="mdi:flag" class="inline h-3 w-3 mr-1" />
          {{ $t('comments.flagged') }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  comment: {
    type: Object,
    required: true
  }
})

defineEmits(['edit', 'delete'])

const authStore = useAuthStore()
const { t } = useI18n()

const canEdit = computed(() => {
  return authStore.user && (
    authStore.user.id === props.comment.user_id ||
    authStore.user.role === 'admin' ||
    authStore.user.role === 'editor'
  )
})

const canDelete = computed(() => {
  return authStore.user && (
    authStore.user.id === props.comment.user_id ||
    authStore.user.role === 'admin' ||
    authStore.user.role === 'editor' ||
    authStore.user.role === 'group_admin'
  )
})

const canManage = computed(() => canEdit.value || canDelete.value)

const showModerationBadges = computed(() => {
  return authStore.user && (
    authStore.user.role === 'admin' ||
    authStore.user.role === 'editor'
  )
})

const getUserName = (user) => {
  if (!user) return t('comments.anonymous')
  return user.first_name && user.last_name
    ? `${user.first_name} ${user.last_name}`
    : user.username || user.email
}

const getUserInitials = (user) => {
  if (!user) return '?'
  if (user.first_name && user.last_name) {
    return `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
  }
  if (user.username) {
    return user.username.substring(0, 2).toUpperCase()
  }
  return user.email[0].toUpperCase()
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const now = new Date()
  const diffInSeconds = Math.floor((now - date) / 1000)

  if (diffInSeconds < 60) {
    return t('comments.time.just_now')
  }
  if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60)
    return t('comments.time.minutes_ago', { count: minutes })
  }
  if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600)
    return t('comments.time.hours_ago', { count: hours })
  }
  if (diffInSeconds < 604800) {
    const days = Math.floor(diffInSeconds / 86400)
    return t('comments.time.days_ago', { count: days })
  }

  return date.toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'short',
    year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
  })
}
</script>
