<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ $t('moderation.title') }}</template>
      <template #description>{{ $t('moderation.description') }}</template>
    </PageHeader>

    <div class="cm-tabs" role="tablist">
      <button
        type="button"
        role="tab"
        class="cm-tab"
        :class="{ 'cm-tab--on': activeTab === 'pending' }"
        :aria-selected="activeTab === 'pending'"
        @click="activeTab = 'pending'"
      >
        {{ $t('moderation.pending') }}
        <span v-if="pendingCount > 0" class="cm-tab__count">{{ pendingCount }}</span>
      </button>
      <button
        type="button"
        role="tab"
        class="cm-tab"
        :class="{ 'cm-tab--on': activeTab === 'settings' }"
        :aria-selected="activeTab === 'settings'"
        @click="activeTab = 'settings'"
      >
        {{ $t('moderation.settings') }}
      </button>
    </div>

    <!-- Commentaires en attente -->
    <template v-if="activeTab === 'pending'">
      <EmptyState
        v-if="!loading && pendingComments.length === 0"
        icon="mdi:check-circle-outline"
        :title="$t('moderation.no_pending')"
      />

      <DataTable
        v-else
        :columns="columns"
        :rows="rows"
        :loading="loading"
        :search-keys="['author', 'content', 'target']"
        :page-size="25"
      >
        <template #cell-author="{ row }">
          <span class="cm-author">{{ row.author }}</span>
        </template>

        <template #cell-target="{ row }">
          <span class="cm-mono">{{ row.target }}</span>
        </template>

        <template #cell-content="{ row }">
          <p class="cm-content">{{ row.content }}</p>
        </template>

        <template #cell-createdAt="{ row }">
          <span class="cm-mono">{{ row.createdAt }}</span>
        </template>

        <template #cell-status="{ row }">
          <StatusBadge v-if="row.raw.is_flagged" tone="bad">{{ $t('moderation.flagged') }}</StatusBadge>
          <StatusBadge v-else tone="warn">{{ $t('moderation.pending') }}</StatusBadge>
        </template>

        <template #cell-actions="{ row }">
          <RowActions :label="row.author" :editable="false" @delete="deleteComment(row.raw.id)">
            <button
              type="button"
              class="row-actions__btn cm-approve"
              :aria-label="`${$t('moderation.approve')} : ${row.author}`"
              :title="$t('moderation.approve')"
              @click="moderateComment(row.raw.id, true, false)"
            >
              <Icon icon="mdi:check" aria-hidden="true" />
            </button>
            <button
              type="button"
              class="row-actions__btn cm-flag"
              :aria-label="`${$t('moderation.flag')} : ${row.author}`"
              :title="$t('moderation.flag')"
              @click="moderateComment(row.raw.id, false, true)"
            >
              <Icon icon="mdi:flag-outline" aria-hidden="true" />
            </button>
          </RowActions>
        </template>
      </DataTable>
    </template>

    <!-- Paramètres -->
    <section v-if="activeTab === 'settings'" class="cm-settings">
      <h2 class="cm-settings__title">{{ $t('moderation.settings_title') }}</h2>

      <div v-for="toggle in toggles" :key="toggle.key" class="cm-row">
        <div class="cm-row__text">
          <label :for="`cm-${toggle.key}`" class="cm-row__label">{{ $t(`moderation.${toggle.label}`) }}</label>
          <p class="cm-row__desc">{{ $t(`moderation.${toggle.label}_desc`) }}</p>
        </div>
        <label class="cm-switch">
          <input
            :id="`cm-${toggle.key}`"
            v-model="settings[toggle.key]"
            type="checkbox"
            class="cm-switch__input"
            :disabled="toggle.dependsOnGlobal && !settings.comments_enabled"
          >
          <span class="cm-switch__track" aria-hidden="true"></span>
        </label>
      </div>

      <div class="cm-field">
        <label for="cm-max-length" class="cm-row__label">{{ $t('moderation.max_length') }}</label>
        <input
          id="cm-max-length"
          v-model.number="settings.max_comment_length"
          type="number"
          min="100"
          max="5000"
          class="form-input"
        >
        <p class="cm-row__desc">{{ $t('moderation.max_length_desc') }}</p>
      </div>

      <div class="cm-save">
        <BaseButton variant="primary" @click="saveSettings">{{ $t('moderation.save_settings') }}</BaseButton>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { commentsService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import { useConfirm } from '@/composables/useConfirm'
import { dateLocaleFor } from '@/utils/homeFormat'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import RowActions from '@/components/ui/RowActions.vue'
import BaseButton from '@/components/ui/BaseButton.vue'

const { t, locale } = useI18n()
const { confirm: askConfirm } = useConfirm()
const appStore = useAppStore()

const activeTab = ref('pending')
const loading = ref(false)
const pendingComments = ref([])
const settings = ref({
  comments_enabled: true,
  news_comments_enabled: true,
  app_comments_enabled: false,
  event_comments_enabled: true,
  require_moderation: false,
  max_comment_length: 1000
})

// Réglages à bascule ; les types de commentaires dépendent de l'activation globale.
const toggles = [
  { key: 'comments_enabled', label: 'enable_comments', dependsOnGlobal: false },
  { key: 'news_comments_enabled', label: 'news_comments', dependsOnGlobal: true },
  { key: 'app_comments_enabled', label: 'app_comments', dependsOnGlobal: true },
  { key: 'event_comments_enabled', label: 'event_comments', dependsOnGlobal: true },
  { key: 'require_moderation', label: 'require_moderation', dependsOnGlobal: false },
]

const columns = computed(() => [
  { key: 'author', label: t('ui.table.columns.author'), sortable: true },
  { key: 'target', label: t('ui.table.columns.type'), sortable: true, hideSm: true },
  { key: 'content', label: t('ui.table.columns.content') },
  { key: 'createdAt', label: t('ui.table.columns.date'), sortable: true, hideSm: true },
  { key: 'status', label: t('ui.table.columns.status'), hideSm: true },
  { key: 'actions', label: '', align: 'end', width: '130px' },
])

const rows = computed(() =>
  pendingComments.value.map(comment => ({
    id: comment.id,
    author: getUserName(comment.user),
    target: `${comment.entity_type} #${comment.entity_id}`,
    content: comment.content,
    createdAt: formatDate(comment.created_at),
    raw: comment,
  })),
)

const pendingCount = computed(() => pendingComments.value.length)

const getUserName = (user) => {
  if (!user) return t('comments.anonymous')
  return user.first_name && user.last_name
    ? `${user.first_name} ${user.last_name}`
    : user.username || user.email
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleDateString(dateLocaleFor(locale.value), {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const loadPendingComments = async () => {
  loading.value = true
  try {
    const response = await commentsService.getPendingComments()
    pendingComments.value = response.comments || []
  } catch (error) {
    console.error('Erreur lors du chargement des commentaires en attente:', error)
  } finally {
    loading.value = false
  }
}

const loadSettings = async () => {
  try {
    const response = await commentsService.getSettings()
    settings.value = response
  } catch (error) {
    console.error('Erreur lors du chargement des paramètres:', error)
  }
}

const moderateComment = async (commentId, isApproved, isFlagged) => {
  try {
    await commentsService.moderateComment({
      comment_id: commentId,
      is_approved: isApproved,
      is_flagged: isFlagged
    })
    await loadPendingComments()
  } catch (error) {
    console.error('Erreur lors de la modération du commentaire:', error)
    appStore.showError(t('moderation.moderateError'))
  }
}

const deleteComment = async (commentId) => {
  const confirmed = await askConfirm({
    message: t('comments.confirmDelete'),
    confirmLabel: t('common.delete'),
    tone: 'danger',
  })
  if (!confirmed) return

  try {
    await commentsService.deleteComment(commentId)
    await loadPendingComments()
  } catch (error) {
    console.error('Erreur lors de la suppression du commentaire:', error)
    appStore.showError(t('comments.deleteError'))
  }
}

const saveSettings = async () => {
  try {
    await commentsService.updateSettings(settings.value)
    appStore.showSuccess(t('moderation.settingsSaved'))
  } catch (error) {
    console.error('Erreur lors de la mise à jour des paramètres:', error)
    appStore.showError(t('moderation.settingsSaveError'))
  }
}

onMounted(() => {
  loadPendingComments()
  loadSettings()
})
</script>

<style scoped>
.cm-tabs { display: flex; gap: 20px; margin-block-end: 16px; border-block-end: 1px solid var(--line); }

.cm-tab {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin-block-end: -1px;
  padding: 10px 2px;
  border: 0;
  border-block-end: 2px solid transparent;
  background: none;
  color: var(--muted);
  font-size: 13.5px;
  font-weight: 600;
  cursor: pointer;
}

.cm-tab:hover { color: var(--text); }
.cm-tab--on { border-block-end-color: var(--acc); color: var(--text); }
.cm-tab:focus-visible { outline: 2px solid var(--acc); outline-offset: 2px; }

.cm-tab__count {
  padding: 1px 6px;
  border: 1px solid var(--warn);
  border-radius: 3px;
  color: var(--warn);
  font-family: var(--f-num);
  font-size: 10.5px;
}

.cm-author { font-weight: 600; color: var(--text); }

.cm-content {
  max-width: 60ch;
  margin: 0;
  color: var(--text);
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.cm-mono {
  font-family: var(--f-num);
  font-size: 11.5px;
  letter-spacing: 0.04em;
  color: var(--muted);
}

.cm-approve:hover { color: var(--ok); }
.cm-flag:hover { color: var(--warn); }

.cm-settings {
  max-width: 720px;
  padding: 16px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
}

.cm-settings__title { margin: 0 0 8px; font-size: 15px; font-weight: 600; color: var(--text); }

.cm-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-block: 12px;
  border-block-end: 1px solid var(--line);
}

.cm-row__label { display: block; font-size: 13.5px; font-weight: 600; color: var(--text); }
.cm-row__desc { margin: 2px 0 0; font-size: 12.5px; color: var(--muted); }

.cm-field { display: grid; gap: 6px; padding-block: 14px; }
.cm-save { padding-block-start: 4px; }

.cm-switch { position: relative; flex: none; display: inline-flex; cursor: pointer; }
.cm-switch__input { position: absolute; opacity: 0; inset: 0; margin: 0; cursor: pointer; }

.cm-switch__track {
  position: relative;
  width: 36px;
  height: 20px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--sunken);
}

.cm-switch__track::after {
  content: '';
  position: absolute;
  inset-block-start: 2px;
  inset-inline-start: 2px;
  width: 14px;
  height: 14px;
  border-radius: 2px;
  background: var(--muted);
  transition: transform 0.15s;
}

.cm-switch__input:checked + .cm-switch__track { border-color: var(--acc); background: var(--acc); }
.cm-switch__input:checked + .cm-switch__track::after { background: var(--surface); transform: translateX(16px); }
[dir='rtl'] .cm-switch__input:checked + .cm-switch__track::after { transform: translateX(-16px); }
.cm-switch__input:disabled + .cm-switch__track { opacity: 0.5; }
.cm-switch__input:disabled { cursor: not-allowed; }
.cm-switch__input:focus-visible + .cm-switch__track { outline: 2px solid var(--acc); outline-offset: 2px; }

@media (prefers-reduced-motion: reduce) {
  .cm-switch__track::after { transition: none; }
}
</style>
