<template>
  <div class="content-area poll-page">
    <div v-if="loading" class="poll-page__state" role="status">
      <Icon icon="mdi:loading" class="poll-page__spin" aria-hidden="true" />
      <span>{{ t('polls.loading') }}</span>
    </div>

    <EmptyState v-else-if="error" icon="mdi:alert-circle-outline" :title="t('polls.detail.notFound')" :description="error">
      <BaseButton tag="router-link" to="/polls" icon="mdi:arrow-left">{{ t('polls.detail.backToPolls') }}</BaseButton>
    </EmptyState>

    <div v-else-if="poll" class="poll-page__body">
      <nav class="crumbs" :aria-label="t('polls.detail.breadcrumbText')">
        <router-link to="/polls" class="crumbs__link">{{ t('polls.detail.breadcrumbText') }}</router-link>
        <Icon icon="mdi:chevron-right" class="crumbs__sep" aria-hidden="true" />
        <span class="crumbs__current" aria-current="page">{{ poll.title }}</span>
      </nav>

      <PollCard
        :poll="poll"
        :compact="false"
        :can-manage="canManage"
        default-expanded
        @vote-success="handleVoteSuccess"
        @edit="openEditor"
        @delete="confirmDelete"
        @close="confirmClose"
      />

      <section v-if="poll.target_groups && poll.target_groups.length > 0" class="poll-info">
        <h2 class="poll-info__title">{{ t('polls.detail.targetGroupsTitle') }}</h2>
        <ul class="poll-info__groups">
          <li v-for="group in poll.target_groups" :key="group.id" class="poll-info__group">
            <span class="poll-info__dot" :style="{ background: group.color || 'var(--muted)' }" aria-hidden="true"></span>
            {{ group.name }}
          </li>
        </ul>
      </section>

      <section v-if="poll.news" class="poll-info">
        <h2 class="poll-info__title">{{ t('polls.detail.linkedArticleTitle') }}</h2>
        <router-link :to="`/news/${poll.news.slug}`" class="poll-info__link">
          <Icon icon="mdi:newspaper-variant-outline" class="poll-info__icon" aria-hidden="true" />
          <span class="poll-info__text">
            <span class="poll-info__name">{{ poll.news.title }}</span>
            <span v-if="poll.news.summary" class="poll-info__summary">{{ poll.news.summary }}</span>
          </span>
          <Icon icon="mdi:chevron-right" class="poll-info__arrow" aria-hidden="true" />
        </router-link>
      </section>

      <section v-if="poll.announcement" class="poll-info">
        <h2 class="poll-info__title">{{ t('polls.detail.linkedAnnouncementTitle') }}</h2>
        <div class="poll-info__link poll-info__link--static">
          <Icon icon="mdi:bullhorn-outline" class="poll-info__icon" aria-hidden="true" />
          <span class="poll-info__text">
            <span class="poll-info__name">{{ poll.announcement.title }}</span>
            <span v-if="poll.announcement.content" class="poll-info__summary">{{ poll.announcement.content }}</span>
          </span>
        </div>
      </section>
    </div>

    <!-- Modal d'édition -->
    <PollEditor
      :show="showEditor"
      :poll="poll"
      :groups="groups"
      @close="showEditor = false"
      @success="handleEditSuccess"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConfirm } from '@/composables/useConfirm'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { pollsService, adminService } from '@/services/api'
import PollCard from '@/components/polls/PollCard.vue'
import PollEditor from '@/components/polls/PollEditor.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()
const { confirm: askConfirm } = useConfirm()

const poll = ref(null)
const groups = ref([])
const loading = ref(false)
const error = ref(null)
const showEditor = ref(false)

// Vérifier si l'utilisateur peut gérer ce sondage
const canManage = computed(() => {
  if (!poll.value) return false
  const user = authStore.user

  if (user.role === 'admin') return true
  if (user.role === 'editor') return true

  // Vérifier si l'utilisateur est admin d'au moins un groupe
  if (user.admin_of_groups && user.admin_of_groups.length > 0) {
    // Peut gérer si auteur ou si le sondage cible un de ses groupes
    if (poll.value.author_id === user.id) return true

    const managedGroupIds = user.managed_group_ids || []
    const targetGroupIds = poll.value.target_groups?.map(g => g.id) || []

    return targetGroupIds.some(id => managedGroupIds.includes(id))
  }

  return false
})

// Charger le sondage
const loadPoll = async () => {
  try {
    loading.value = true
    error.value = null
    poll.value = await pollsService.getPollById(route.params.id)
  } catch (err) {
    console.error('Error loading poll:', err)
    error.value = err.response?.data?.error || t('polls.detail.loadError')
  } finally {
    loading.value = false
  }
}

// Charger les groupes (pour l'éditeur)
const loadGroups = async () => {
  try {
    const response = await adminService.getGroups()
    groups.value = response.data || response
  } catch (err) {
    console.error('Error loading groups:', err)
  }
}

// Gérer le succès du vote
const handleVoteSuccess = () => {
  loadPoll()
}

// Ouvrir l'éditeur
const openEditor = () => {
  if (groups.value.length === 0) {
    loadGroups()
  }
  showEditor.value = true
}

// Gérer le succès de l'édition
const handleEditSuccess = () => {
  loadPoll()
}

// Confirmer la suppression
const confirmDelete = async () => {
  if (!(await askConfirm({ message: t('polls.actions.confirmDelete'), tone: 'danger' }))) {
    return
  }

  try {
    const user = authStore.user

    if (user.role === 'admin') {
      await pollsService.deletePoll(poll.value.id)
    } else if (user.role === 'editor') {
      await pollsService.deletePollAsEditor(poll.value.id)
    } else if (user.admin_of_groups && user.admin_of_groups.length > 0) {
      await pollsService.deletePollAsGroupAdmin(poll.value.id)
    }

    if (window.$toast) {
      window.$toast.success(t('polls.actions.deleteSuccess'))
    }

    router.push('/polls')
  } catch (err) {
    console.error('Erreur lors de la suppression:', err)
    if (window.$toast) {
      window.$toast.error(err.response?.data?.error || t('polls.actions.deleteError'))
    }
  }
}

// Confirmer la fermeture
const confirmClose = async () => {
  if (!(await askConfirm({ message: t('polls.actions.confirmClose'), tone: 'danger', confirmLabel: t('polls.actions.close') }))) {
    return
  }

  try {
    const user = authStore.user

    if (user.role === 'admin') {
      await pollsService.closePoll(poll.value.id)
    } else if (user.admin_of_groups && user.admin_of_groups.length > 0) {
      await pollsService.closePollAsGroupAdmin(poll.value.id)
    }

    if (window.$toast) {
      window.$toast.success(t('polls.actions.closeSuccess'))
    }

    await loadPoll()
  } catch (err) {
    console.error('Erreur lors de la fermeture:', err)
    if (window.$toast) {
      window.$toast.error(err.response?.data?.error || t('polls.actions.closeError'))
    }
  }
}

onMounted(() => {
  loadPoll()
})
</script>

<style scoped>
/* Détail d'un sondage (direction A+) : le sondage ouvert d'emblée, puis ses liens. */
.poll-page__body {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 860px;
  margin-inline: auto;
}

.poll-page__state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding-block: 64px;
  color: var(--muted);
  font-size: 13px;
}

.poll-page__spin { width: 30px; height: 30px; animation: poll-spin 1s linear infinite; }
@keyframes poll-spin { to { transform: rotate(360deg); } }

.crumbs {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  font-family: var(--f-num);
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.crumbs__link { color: var(--muted); }
.crumbs__link:hover { color: var(--text); text-decoration: underline; }
.crumbs__link:focus-visible { outline: 2px solid var(--acc); outline-offset: 2px; border-radius: 2px; }
.crumbs__sep { width: 14px; height: 14px; color: var(--muted); }
.crumbs__current { min-width: 0; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; color: var(--text); }

.poll-info {
  padding: 14px 16px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
}

.poll-info__title {
  margin: 0 0 10px;
  font-family: var(--f-num);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
}

.poll-info__groups { display: flex; flex-wrap: wrap; gap: 6px; margin: 0; padding: 0; list-style: none; }

.poll-info__group {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border: 1px solid var(--line);
  border-radius: 3px;
  font-size: 12.5px;
  color: var(--text);
}

.poll-info__dot { width: 8px; height: 8px; border-radius: 50%; }

.poll-info__link {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--line);
  border-radius: 3px;
  color: var(--text);
}

a.poll-info__link:hover { border-color: var(--acc); }
a.poll-info__link:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

.poll-info__icon { flex: none; width: 22px; height: 22px; color: var(--muted); }
.poll-info__text { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.poll-info__name { font-weight: 600; }

.poll-info__summary {
  font-size: 12.5px;
  color: var(--muted);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.poll-info__arrow { flex: none; width: 20px; height: 20px; color: var(--muted); }

[dir='rtl'] .crumbs__sep,
[dir='rtl'] .poll-info__arrow { transform: scaleX(-1); }

@media (prefers-reduced-motion: reduce) {
  .poll-page__spin { animation: none; }
}
</style>
