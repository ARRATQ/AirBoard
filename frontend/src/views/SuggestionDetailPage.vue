<template>
  <div class="w-full p-6 space-y-6">
    <div class="flex items-center justify-between gap-3 flex-wrap">
      <RouterLink to="/suggestions" class="suggestion-back-btn">
        <Icon icon="mdi:arrow-left" />
        {{ t('suggestions.backToList') }}
      </RouterLink>

      <select
        v-if="authStore.isAdmin && suggestion"
        class="form-select min-w-[220px]"
        :value="suggestion.status"
        @change="updateStatus($event.target.value)"
      >
        <option v-for="status in statuses" :key="status" :value="status">
          {{ t(`suggestions.statuses.${status}`) }}
        </option>
      </select>
    </div>

    <div v-if="loading" class="loading-state">
      <Icon icon="mdi:loading" class="animate-spin text-2xl text-amber-500" />
      <p>{{ t('suggestions.loading') }}</p>
    </div>

    <div v-else-if="!suggestion" class="empty-state">
      <Icon icon="mdi:lightbulb-off-outline" class="text-5xl text-gray-400" />
      <h3>{{ t('suggestions.notFound') }}</h3>
    </div>

    <template v-else>
      <article class="card p-6 border border-gray-200 dark:border-gray-700">
        <h1 class="text-2xl md:text-3xl font-bold text-gray-900 dark:text-white">{{ suggestion.title }}</h1>

        <div class="mt-3 flex items-center gap-2 flex-wrap text-xs">
          <span class="px-2 py-1 rounded-full bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300">
            {{ suggestion.category_ref?.name || suggestion.category || t('suggestions.noCategory') }}
          </span>
          <span class="px-2 py-1 rounded-full bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300">
            {{ t(`suggestions.statuses.${suggestion.status}`) }}
          </span>
          <span class="text-gray-500 dark:text-gray-400">
            {{ formatAuthor(suggestion) }} · {{ formatDate(suggestion.created_at) }}
          </span>
        </div>

        <p class="mt-5 text-gray-800 dark:text-gray-200 whitespace-pre-line leading-relaxed">{{ suggestion.description }}</p>

        <div v-if="suggestion.poll" class="mt-4 text-sm text-gray-500 dark:text-gray-400 flex items-center gap-2">
          <Icon icon="mdi:poll" class="h-4 w-4" />
          <span>{{ t('suggestions.linkedPoll') }}: {{ suggestion.poll.title }}</span>
        </div>

        <div class="mt-6 flex items-center gap-3 flex-wrap">
          <button
            class="vote-button"
            :class="suggestion.user_vote === 'up' ? 'vote-button-active-up' : 'vote-button-up'"
            @click="handleVote('up')"
          >
            <Icon icon="mdi:thumb-up" class="h-4 w-4" />
            <span>{{ t('suggestions.voteFor') }}</span>
            <span class="vote-count">{{ suggestion.upvotes || 0 }}</span>
          </button>

          <button
            class="vote-button"
            :class="suggestion.user_vote === 'down' ? 'vote-button-active-down' : 'vote-button-down'"
            @click="handleVote('down')"
          >
            <Icon icon="mdi:thumb-down" class="h-4 w-4" />
            <span>{{ t('suggestions.voteAgainst') }}</span>
            <span class="vote-count">{{ suggestion.downvotes || 0 }}</span>
          </button>

          <button v-if="suggestion.user_vote" class="suggestion-unvote-btn" @click="handleVote('none')">
            {{ t('suggestions.removeVote') }}
          </button>
        </div>
      </article>

      <article class="card p-6 border border-gray-200 dark:border-gray-700">
        <CommentSection
          entity-type="suggestion"
          :entity-id="suggestion.id"
        />
      </article>
    </template>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { Icon } from '@iconify/vue'
import { suggestionsService } from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import CommentSection from '@/components/comments/CommentSection.vue'

const { t, locale } = useI18n()
const route = useRoute()
const authStore = useAuthStore()

const statuses = ['new', 'reviewing', 'planned', 'in_progress', 'implemented', 'rejected']

const loading = ref(false)
const suggestion = ref(null)

const loadSuggestion = async () => {
  try {
    loading.value = true
    suggestion.value = await suggestionsService.getSuggestionById(route.params.id)
  } catch (error) {
    console.error('Failed to load suggestion:', error)
    suggestion.value = null
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.loadError'))
    }
  } finally {
    loading.value = false
  }
}

const handleVote = async (voteType) => {
  if (!suggestion.value) return
  try {
    const result = await suggestionsService.voteSuggestion(suggestion.value.id, voteType)
    suggestion.value.upvotes = result.upvotes
    suggestion.value.downvotes = result.downvotes
    suggestion.value.user_vote = result.user_vote || ''
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.voteError'))
    }
  }
}

const updateStatus = async (status) => {
  if (!authStore.isAdmin || !suggestion.value) return
  try {
    const updated = await suggestionsService.updateSuggestionStatus(suggestion.value.id, status)
    suggestion.value.status = updated.status
    if (window.$toast) {
      window.$toast.success(t('suggestions.statusUpdated'))
    }
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.statusUpdateError'))
    }
  }
}

const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleDateString(locale.value)
}

const formatAuthor = (item) => {
  if (item.is_anonymous) return t('suggestions.anonymous')
  const first = item.user?.first_name || ''
  const last = item.user?.last_name || ''
  const fullName = `${first} ${last}`.trim()
  return fullName || item.user?.username || t('suggestions.anonymous')
}

onMounted(loadSuggestion)
</script>

<style scoped>
.vote-button {
  @apply inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold border transition-all duration-200;
}

.vote-button-up {
  @apply border-green-200 text-green-700 bg-green-50 hover:bg-green-100 dark:border-green-800 dark:text-green-300 dark:bg-green-900/20 dark:hover:bg-green-900/35;
}

.vote-button-down {
  @apply border-red-200 text-red-700 bg-red-50 hover:bg-red-100 dark:border-red-800 dark:text-red-300 dark:bg-red-900/20 dark:hover:bg-red-900/35;
}

.vote-button-active-up {
  @apply border-green-500 text-green-800 bg-green-100 shadow-sm ring-2 ring-green-500/30 dark:text-green-200 dark:bg-green-900/45;
}

.vote-button-active-down {
  @apply border-red-500 text-red-800 bg-red-100 shadow-sm ring-2 ring-red-500/30 dark:text-red-200 dark:bg-red-900/45;
}

.vote-count {
  @apply inline-flex items-center justify-center min-w-[22px] h-6 px-2 rounded-full text-xs font-bold bg-white/70 dark:bg-black/20;
}

.suggestion-back-btn {
  @apply inline-flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm font-semibold;
  @apply border border-gray-300 text-gray-700 bg-white hover:bg-gray-50;
  @apply dark:border-gray-600 dark:text-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700;
  @apply transition-colors duration-200;
}

.suggestion-unvote-btn {
  @apply inline-flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm font-semibold;
  @apply border border-rose-200 text-rose-700 bg-rose-50 hover:bg-rose-100;
  @apply dark:border-rose-800 dark:text-rose-300 dark:bg-rose-900/20 dark:hover:bg-rose-900/35;
  @apply focus:outline-none focus:ring-2 focus:ring-rose-300 focus:ring-offset-2;
  @apply dark:focus:ring-offset-gray-900 transition-colors duration-200;
}
</style>
