<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ t('gamificationAdmin.title') }}</template>
      <template #description>{{ t('gamificationAdmin.subtitle') }}</template>
    </PageHeader>

    <!-- Règles de points -->
    <section class="gm-section">
      <h2 class="gm-section__title">{{ t('gamificationAdmin.rules') }}</h2>

      <DataTable
        :columns="ruleColumns"
        :rows="ruleRows"
        :search-keys="['reason']"
        :page-size="25"
      >
        <template #cell-reason="{ row }">
          <span class="gm-mono">{{ row.reason }}</span>
        </template>

        <template #cell-points="{ row }">
          <label>
            <span class="sr-only">{{ t('gamificationAdmin.points') }} : {{ row.reason }}</span>
            <input v-model.number="row.raw.points" type="number" min="0" class="form-input gm-points" />
          </label>
        </template>

        <template #cell-enabled="{ row }">
          <label class="gm-check">
            <input v-model="row.raw.enabled" type="checkbox" class="form-checkbox" />
            <span class="sr-only">{{ t('common.active') }} : {{ row.reason }}</span>
          </label>
        </template>

        <template #cell-actions="{ row }">
          <BaseButton variant="secondary" size="sm" icon="mdi:content-save-outline" @click="saveRule(row.raw)">
            {{ t('common.save') }}
          </BaseButton>
        </template>
      </DataTable>
    </section>

    <!-- Badges -->
    <section class="gm-section">
      <div class="gm-section__head">
        <h2 class="gm-section__title">{{ t('gamificationAdmin.badges') }}</h2>
        <BaseButton variant="primary" icon="mdi:plus" @click="openCreate">{{ t('common.create') }}</BaseButton>
      </div>

      <EmptyState v-if="badges.length === 0" icon="mdi:medal-outline" :title="t('gamificationAdmin.badges')">
        <BaseButton variant="primary" size="sm" icon="mdi:plus" @click="openCreate">{{ t('common.create') }}</BaseButton>
      </EmptyState>

      <DataTable
        v-else
        :columns="badgeColumns"
        :rows="badgeRows"
        :search-keys="['name', 'code', 'description', 'trigger']"
        :page-size="25"
      >
        <template #cell-name="{ row }">
          <span class="gm-badge">
            <span class="gm-badge__icon" :style="{ backgroundColor: row.raw.color || undefined }">
              <Icon :icon="row.raw.icon || 'mdi:medal'" aria-hidden="true" />
            </span>
            <span class="gm-badge__text">
              <span class="gm-badge__name">{{ row.name }}</span>
              <span class="gm-mono">{{ row.code }}</span>
              <span v-if="row.description" class="gm-badge__desc">{{ row.description }}</span>
            </span>
          </span>
        </template>

        <template #cell-trigger="{ row }">
          <span class="gm-mono">{{ row.trigger }} · {{ row.metric }} · {{ row.threshold }}</span>
        </template>

        <template #cell-status="{ row }">
          <StatusBadge :tone="row.raw.is_active ? 'ok' : 'neutral'">{{ row.status }}</StatusBadge>
        </template>

        <template #cell-actions="{ row }">
          <RowActions :label="row.name" @edit="openEdit(row.raw)" @delete="removeBadge(row.raw)" />
        </template>
      </DataTable>
    </section>

        <Teleport to="body">
      <div v-if="showModal" class="gm-overlay" @click.self="closeModal">
        <div class="gm-modal" role="dialog" aria-modal="true">
          <h3 class="gm-modal__title">{{ editingId ? t('common.edit') : t('common.create') }}</h3>
          <form class="gm-form" @submit.prevent="saveBadge">
            <div>
              <label class="form-label">Code technique</label>
              <input v-model="form.code" class="form-input" placeholder="ex: suggestion_master" required />
              <p class="gm-hint">Identifiant unique interne (non visible par l'utilisateur).</p>
            </div>

            <div>
              <label class="form-label">Nom du badge</label>
              <input v-model="form.name" class="form-input" :placeholder="t('common.name')" required />
              <p class="gm-hint">Texte affiché aux utilisateurs.</p>
            </div>

            <div>
              <label class="form-label">Icône</label>
              <input v-model="form.icon" class="form-input" placeholder="mdi:medal" />
              <p class="gm-hint">Format Iconify, ex: <code>mdi:medal</code>.</p>
            </div>

            <div>
              <label class="form-label">Couleur</label>
              <div class="gm-colorrow">
                <input v-model="form.color" class="form-input" placeholder="#FBBF24" />
                <input
                  v-model="form.color"
                  type="color"
                  class="gm-color"
                  aria-label="Choisir une couleur"
                />
              </div>
              <div class="gm-presets">
                <button
                  v-for="preset in colorPresets"
                  :key="preset"
                  type="button"
                  class="gm-preset"
                  :style="{ backgroundColor: preset }"
                  :title="preset"
                  @click="form.color = preset"
                ></button>
              </div>
              <p class="gm-hint">Couleur hexadécimale du badge.</p>
            </div>

            <div class="gm-form__wide">
              <label class="form-label">Aperçu du badge</label>
              <div class="gm-preview">
                <div class="gm-preview__badge" :style="badgePreviewStyle">
                  <Icon :icon="safePreviewIcon" class="gm-preview__icon" />
                </div>
                <div>
                  <p class="gm-preview__name">{{ form.name || 'Nom du badge' }}</p>
                  <p class="gm-hint">{{ form.icon || 'mdi:medal' }} · {{ form.color || '#FBBF24' }}</p>
                </div>
              </div>
            </div>

            <div>
              <label class="form-label">Bonus XP du badge</label>
              <input v-model.number="form.xp_reward" type="number" class="form-input" :placeholder="t('gamificationAdmin.badgeReward')" />
              <p class="gm-hint">XP accordé une seule fois au déblocage du badge.</p>
            </div>

            <div>
              <label class="form-label">Catégorie</label>
              <select v-model="form.category" class="form-select">
                <option value="user">user</option>
                <option value="contributor">contributor</option>
              </select>
              <p class="gm-hint"><code>contributor</code> est réservé aux profils contributeurs.</p>
            </div>

            <div>
              <label class="form-label">Événement déclencheur</label>
              <select v-model="form.trigger_reason" class="form-select">
                <option value="">Sélectionner un trigger</option>
                <option v-for="reason in reasonOptions" :key="reason" :value="reason">{{ reason }}</option>
              </select>
              <p class="gm-hint">{{ selectedReasonHelp }}</p>
            </div>

            <div>
              <label class="form-label">Métrique de progression</label>
              <select v-model="form.metric" class="form-select">
                <option value="">Sélectionner une métrique</option>
                <option v-for="metric in metricOptions" :key="metric" :value="metric">{{ metric }}</option>
              </select>
              <p class="gm-hint">{{ selectedMetricHelp }}</p>
            </div>

            <div>
              <label class="form-label">Seuil de déblocage</label>
              <input v-model.number="form.threshold" type="number" min="0" class="form-input" :placeholder="t('gamificationAdmin.threshold')" />
              <p class="gm-hint">Nombre d'actions nécessaires pour obtenir le badge.</p>
            </div>

            <label class="gm-check"><input v-model="form.is_active" type="checkbox" class="form-checkbox" />Badge actif</label>
            <div class="gm-form__wide">
              <label class="form-label">Description</label>
              <textarea v-model="form.description" rows="3" class="form-textarea" :placeholder="t('common.description')"></textarea>
            </div>
            <div class="gm-form__wide gm-form__actions">
              <BaseButton variant="secondary" @click="closeModal">{{ t('common.cancel') }}</BaseButton>
              <BaseButton variant="primary" type="submit">{{ t('common.save') }}</BaseButton>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConfirm } from '@/composables/useConfirm'
import { Icon } from '@iconify/vue'
import { gamificationService } from '@/services/api'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import RowActions from '@/components/ui/RowActions.vue'

const { t } = useI18n()
const { confirm: askConfirm } = useConfirm()

const rules = ref([])
const badges = ref([])
const showModal = ref(false)
const editingId = ref(null)
const reasonOptions = ref([])
const colorPresets = ['#FBBF24', '#3B82F6', '#10B981', '#EF4444', '#8B5CF6', '#14B8A6', '#F97316', '#6366F1']
const metricOptions = [
  'reason_count',
  'app_click_distinct',
  'news_read_count',
  'news_publish_count',
  'event_publish_count',
  'poll_vote_count',
  'poll_create_count',
  'comment_count',
  'suggestion_create_count',
  'suggestion_vote_count',
  'suggestion_comment_count'
]

const reasonHelpMap = {
  daily_login: 'Se déclenche à chaque connexion journalière récompensée.',
  app_click: 'Se déclenche à chaque ouverture de ressource.',
  news_read: 'Se déclenche à chaque lecture d\'article.',
  news_publish: 'Se déclenche à chaque publication d\'article.',
  event_publish: 'Se déclenche à chaque création d\'événement.',
  poll_vote: 'Se déclenche à chaque vote sur un sondage.',
  poll_create: 'Se déclenche à chaque création de sondage.',
  comment_create: 'Se déclenche à chaque commentaire créé.',
  suggestion_create: 'Se déclenche à chaque suggestion envoyée.',
  suggestion_vote: 'Se déclenche à chaque vote de suggestion.',
  suggestion_comment: 'Se déclenche à chaque commentaire sur suggestion.'
}

const metricHelpMap = {
  reason_count: 'Compter le nombre de fois que le trigger est reçu.',
  app_click_distinct: 'Compter le nombre de ressources différentes ouvertes.',
  news_read_count: 'Compter le total d\'articles lus.',
  news_publish_count: 'Compter le total d\'articles publiés.',
  event_publish_count: 'Compter le total d\'événements publiés.',
  poll_vote_count: 'Compter le total de votes sur sondages.',
  poll_create_count: 'Compter le total de sondages créés.',
  comment_count: 'Compter le total de commentaires.',
  suggestion_create_count: 'Compter le total de suggestions créées.',
  suggestion_vote_count: 'Compter le total de votes sur suggestions.',
  suggestion_comment_count: 'Compter le total de commentaires sur suggestions.'
}

const reasonToMetricDefault = {
  daily_login: 'reason_count',
  app_click: 'app_click_distinct',
  news_read: 'news_read_count',
  news_publish: 'news_publish_count',
  event_publish: 'event_publish_count',
  poll_vote: 'poll_vote_count',
  poll_create: 'poll_create_count',
  comment_create: 'comment_count',
  suggestion_create: 'suggestion_create_count',
  suggestion_vote: 'suggestion_vote_count',
  suggestion_comment: 'suggestion_comment_count'
}

const selectedReasonHelp = computed(() => {
  return reasonHelpMap[form.value.trigger_reason] || 'Choisissez l\'événement qui fait progresser ce badge.'
})

const selectedMetricHelp = computed(() => {
  return metricHelpMap[form.value.metric] || 'Choisissez la façon de mesurer la progression vers le badge.'
})

const safePreviewIcon = computed(() => form.value.icon?.trim() || 'mdi:medal')

const badgePreviewStyle = computed(() => {
  const color = form.value.color?.trim() || '#FBBF24'
  return { backgroundColor: color }
})
const form = ref({
  code: '',
  name: '',
  description: '',
  icon: 'mdi:medal',
  color: '#FBBF24',
  xp_reward: 0,
  category: 'user',
  trigger_reason: '',
  metric: '',
  threshold: 0,
  is_active: true,
  is_secret: false
})

const ruleColumns = computed(() => [
  { key: 'reason', label: t('ui.table.columns.reason'), sortable: true },
  { key: 'points', label: t('gamificationAdmin.points'), width: '140px' },
  { key: 'enabled', label: t('common.active'), width: '90px' },
  { key: 'actions', label: '', align: 'end', width: '120px' },
])

// Les champs points / actif s'éditent directement sur l'objet de la règle (raw).
const ruleRows = computed(() => rules.value.map(rule => ({ id: rule.id, reason: rule.reason, raw: rule })))

const badgeColumns = computed(() => [
  { key: 'name', label: t('ui.table.columns.name'), sortable: true },
  { key: 'trigger', label: t('ui.table.columns.type'), hideSm: true },
  { key: 'status', label: t('ui.table.columns.status'), sortable: true },
  { key: 'actions', label: '', align: 'end', width: '80px' },
])

const badgeRows = computed(() =>
  badges.value.map(badge => ({
    id: badge.id,
    name: badge.name,
    code: badge.code,
    description: badge.description || '',
    trigger: badge.trigger_reason || '-',
    metric: badge.metric || '-',
    threshold: badge.threshold || 0,
    status: badge.is_active ? t('common.active') : t('common.inactive'),
    raw: badge,
  })),
)

const loadData = async () => {
  const [rulesData, badgesData] = await Promise.all([
    gamificationService.getRules(),
    gamificationService.getAdminAchievements()
  ])
  rules.value = rulesData.rules || []
  badges.value = badgesData.achievements || []
  reasonOptions.value = (rules.value || []).map(r => r.reason)
}

const saveRule = async (rule) => {
  try {
    await gamificationService.upsertRule({ reason: rule.reason, points: Number(rule.points) || 0, enabled: !!rule.enabled })
    window.$toast?.success(t('common.save'))
  } catch (e) {
    window.$toast?.error(e.response?.data?.error || 'Error')
  }
}

const openCreate = () => {
  editingId.value = null
  form.value = { code: '', name: '', description: '', icon: 'mdi:medal', color: '#FBBF24', xp_reward: 0, category: 'user', trigger_reason: '', metric: '', threshold: 0, is_active: true, is_secret: false }
  showModal.value = true
}

const openEdit = (badge) => {
  editingId.value = badge.id
  form.value = { ...badge }
  showModal.value = true
}

watch(() => form.value.trigger_reason, (reason) => {
  if (!reason) return
  const suggestedMetric = reasonToMetricDefault[reason]
  if (!form.value.metric || form.value.metric === 'reason_count') {
    form.value.metric = suggestedMetric
  }
})

const closeModal = () => {
  showModal.value = false
}

const saveBadge = async () => {
  try {
    const payload = { ...form.value, xp_reward: Number(form.value.xp_reward) || 0, threshold: Number(form.value.threshold) || 0 }
    if (editingId.value) {
      await gamificationService.updateAchievement(editingId.value, payload)
    } else {
      await gamificationService.createAchievement(payload)
    }
    closeModal()
    await loadData()
    window.$toast?.success(t('common.save'))
  } catch (e) {
    window.$toast?.error(e.response?.data?.error || 'Error')
  }
}

const removeBadge = async (badge) => {
  const confirmed = await askConfirm({
    message: t('gamificationAdmin.confirmDeleteBadge', { name: badge.name }),
    confirmLabel: t('common.delete'),
    tone: 'danger',
  })
  if (!confirmed) return
  try {
    await gamificationService.deleteAchievement(badge.id)
    await loadData()
    window.$toast?.success(t('common.delete'))
  } catch (e) {
    window.$toast?.error(e.response?.data?.error || 'Error')
  }
}

onMounted(loadData)
</script>

<style scoped>
.gm-section { margin-block-end: 28px; }
.gm-section__head { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-block-end: 10px; }
.gm-section__title { margin: 0 0 10px; font-size: 15px; font-weight: 600; color: var(--text); }
.gm-section__head .gm-section__title { margin: 0; }

.gm-mono { font-family: var(--f-num); font-size: 11.5px; letter-spacing: 0.04em; color: var(--muted); }
.gm-points { max-width: 110px; }
.gm-check { display: inline-flex; align-items: center; gap: 8px; font-size: 13px; color: var(--text); }

.gm-badge { display: flex; align-items: flex-start; gap: 10px; }

.gm-badge__icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 30px;
  height: 30px;
  border-radius: 4px;
  background: var(--acc);
  color: var(--acc-ink);
  font-size: 18px;
}

.gm-badge__text { display: flex; flex-direction: column; min-width: 0; }
.gm-badge__name { font-weight: 600; color: var(--text); }
.gm-badge__desc { margin-block-start: 2px; font-size: 12px; color: var(--muted); }

.gm-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: rgb(0 0 0 / 0.5);
}

.gm-modal {
  width: 100%;
  max-width: 672px;
  max-height: 92vh;
  padding: 18px;
  overflow-y: auto;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
  color: var(--text);
}

.gm-modal__title { margin: 0 0 14px; font-size: 16px; font-weight: 600; }
.gm-form { display: grid; grid-template-columns: 1fr; gap: 14px; }
.gm-form__actions { display: flex; justify-content: flex-end; gap: 8px; }
.gm-hint { margin: 4px 0 0; font-size: 12px; color: var(--muted); }

.gm-colorrow, .gm-presets { display: flex; align-items: center; gap: 8px; }
.gm-presets { margin-block-start: 8px; }

.gm-color {
  width: 48px;
  height: 40px;
  padding: 3px;
  border: 1px solid var(--line);
  border-radius: 3px;
  background: var(--surface);
  cursor: pointer;
}

.gm-preset { width: 22px; height: 22px; padding: 0; border: 1px solid var(--line); border-radius: 3px; cursor: pointer; }
.gm-preset:focus-visible { outline: 2px solid var(--acc); outline-offset: 2px; }

.gm-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--sunken);
}

.gm-preview__badge { display: grid; place-items: center; width: 44px; height: 44px; border-radius: 4px; color: #fff; }
.gm-preview__icon { width: 24px; height: 24px; }
.gm-preview__name { margin: 0; font-weight: 600; color: var(--text); }

@media (min-width: 768px) {
  .gm-form { grid-template-columns: 1fr 1fr; }
  .gm-form__wide { grid-column: 1 / -1; }
}
</style>
