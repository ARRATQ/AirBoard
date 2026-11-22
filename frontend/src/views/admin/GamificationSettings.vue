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
        :search-keys="['label', 'reason']"
        :page-size="25"
      >
        <template #cell-reason="{ row }">
          <span class="gm-rule">
            <span>{{ row.label }}</span>
            <span class="gm-mono">{{ row.reason }}</span>
          </span>
        </template>

        <template #cell-points="{ row }">
          <label>
            <span class="sr-only">{{ t('gamificationAdmin.points') }} : {{ row.label }}</span>
            <input v-model.number="row.raw.points" type="number" min="0" class="form-input gm-points" />
          </label>
        </template>

        <template #cell-enabled="{ row }">
          <label class="gm-check">
            <input v-model="row.raw.enabled" type="checkbox" class="form-checkbox" />
            <span class="sr-only">{{ t('common.active') }} : {{ row.label }}</span>
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
            <span class="gm-badge__icon" :style="{ '--c': row.raw.color || undefined }">
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
          <span>{{ row.trigger }} · {{ row.metric }} · {{ row.threshold }}</span>
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
        <div v-focus-trap class="gm-modal" role="dialog" aria-modal="true" aria-labelledby="gm-modal-title">
          <h3 id="gm-modal-title" class="gm-modal__title">{{ editingId ? t('common.edit') : t('common.create') }}</h3>
          <form class="gm-form" @submit.prevent="saveBadge">
            <div>
              <label for="gm-code" class="form-label">{{ t('gamificationAdmin.form.code') }}</label>
              <input id="gm-code" v-model="form.code" class="form-input" placeholder="suggestion_master" dir="ltr" required />
              <p class="gm-hint">{{ t('gamificationAdmin.form.codeHint') }}</p>
            </div>

            <div>
              <label for="gm-name" class="form-label">{{ t('gamificationAdmin.form.name') }}</label>
              <input id="gm-name" v-model="form.name" class="form-input" required />
              <p class="gm-hint">{{ t('gamificationAdmin.form.nameHint') }}</p>
            </div>

            <div>
              <IconInput id="gm-icon" v-model="form.icon" :label="t('gamificationAdmin.form.icon')" placeholder="lucide:medal" />
              <p class="gm-hint">{{ t('gamificationAdmin.form.iconHint', { example: 'mdi:medal' }) }}</p>
            </div>

            <div>
              <label for="gm-color" class="form-label">{{ t('gamificationAdmin.form.color') }}</label>
              <div class="gm-colorrow">
                <input id="gm-color" v-model="form.color" class="form-input" placeholder="#FBBF24" dir="ltr" />
                <input
                  v-model="form.color"
                  type="color"
                  class="gm-color"
                  :aria-label="t('gamificationAdmin.form.pickColor')"
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
                  :aria-label="preset"
                  @click="form.color = preset"
                ></button>
              </div>
              <p class="gm-hint">{{ t('gamificationAdmin.form.colorHint') }}</p>
            </div>

            <div class="gm-form__wide">
              <p class="form-label">{{ t('gamificationAdmin.form.preview') }}</p>
              <div class="gm-preview">
                <div class="gm-preview__badge" :style="badgePreviewStyle">
                  <Icon :icon="safePreviewIcon" class="gm-preview__icon" aria-hidden="true" />
                </div>
                <div>
                  <p class="gm-preview__name">{{ form.name || t('gamificationAdmin.form.name') }}</p>
                  <p class="gm-hint" dir="ltr">{{ form.icon || 'mdi:medal' }} · {{ form.color || '#FBBF24' }}</p>
                </div>
              </div>
            </div>

            <div>
              <label for="gm-xp" class="form-label">{{ t('gamificationAdmin.form.xpReward') }}</label>
              <input id="gm-xp" v-model.number="form.xp_reward" type="number" min="0" class="form-input" />
              <p class="gm-hint">{{ t('gamificationAdmin.form.xpRewardHint') }}</p>
            </div>

            <div>
              <label for="gm-category" class="form-label">{{ t('gamificationAdmin.form.category') }}</label>
              <select id="gm-category" v-model="form.category" class="form-select">
                <option value="user">{{ t('gamificationAdmin.categories.user') }}</option>
                <option value="contributor">{{ t('gamificationAdmin.categories.contributor') }}</option>
              </select>
              <p class="gm-hint">{{ t('gamificationAdmin.form.categoryHint', { code: t('gamificationAdmin.categories.contributor') }) }}</p>
            </div>

            <div>
              <label for="gm-trigger" class="form-label">{{ t('gamificationAdmin.form.trigger') }}</label>
              <select id="gm-trigger" v-model="form.trigger_reason" class="form-select">
                <option value="">{{ t('gamificationAdmin.form.triggerPlaceholder') }}</option>
                <option v-for="reason in reasonOptions" :key="reason" :value="reason">{{ labelFor('reasonLabels', reason) }}</option>
              </select>
              <p class="gm-hint">{{ selectedReasonHelp }}</p>
            </div>

            <div>
              <label for="gm-metric" class="form-label">{{ t('gamificationAdmin.form.metric') }}</label>
              <select id="gm-metric" v-model="form.metric" class="form-select">
                <option value="">{{ t('gamificationAdmin.form.metricPlaceholder') }}</option>
                <option v-for="metric in metricOptions" :key="metric" :value="metric">{{ labelFor('metricLabels', metric) }}</option>
              </select>
              <p class="gm-hint">{{ selectedMetricHelp }}</p>
            </div>

            <div>
              <label for="gm-threshold" class="form-label">{{ t('gamificationAdmin.form.threshold') }}</label>
              <input id="gm-threshold" v-model.number="form.threshold" type="number" min="0" class="form-input" />
              <p class="gm-hint">{{ t('gamificationAdmin.form.thresholdHint') }}</p>
            </div>

            <label class="gm-check"><input v-model="form.is_active" type="checkbox" class="form-checkbox" />{{ t('gamificationAdmin.form.active') }}</label>
            <div class="gm-form__wide">
              <label for="gm-description" class="form-label">{{ t('common.description') }}</label>
              <textarea id="gm-description" v-model="form.description" rows="3" class="form-textarea"></textarea>
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
import IconInput from '@/components/ui/IconInput.vue'
import { computed, onMounted, ref, watch } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
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

const { t, te } = useI18n()
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

// Aide contextuelle : une phrase par déclencheur ou métrique connus, sinon la consigne générale.
const helpFor = (group, key, fallback) => {
  const path = `gamificationAdmin.${group}.${key}`
  return key && te(path) ? t(path) : t(fallback)
}

// Libellé lisible d'un déclencheur ou d'une métrique ; le code brut si aucune traduction.
const labelFor = (group, code) => {
  const path = `gamificationAdmin.${group}.${code}`
  return code && te(path) ? t(path) : code
}

const selectedReasonHelp = computed(() => helpFor('reasons', form.value.trigger_reason, 'gamificationAdmin.form.triggerHelp'))
const selectedMetricHelp = computed(() => helpFor('metrics', form.value.metric, 'gamificationAdmin.form.metricHelp'))

const safePreviewIcon = computed(() => form.value.icon?.trim() || 'mdi:medal')

const badgePreviewStyle = computed(() => {
  const color = form.value.color?.trim() || '#FBBF24'
  return { '--c': color }
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
const ruleRows = computed(() => rules.value.map(rule => ({ id: rule.id, reason: rule.reason, label: labelFor('reasonLabels', rule.reason), raw: rule })))

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
    trigger: badge.trigger_reason ? labelFor('reasonLabels', badge.trigger_reason) : '-',
    metric: badge.metric ? labelFor('metricLabels', badge.metric) : '-',
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
    window.$toast?.success(t('common.feedback.saved'))
  } catch (e) {
    window.$toast?.error(t('common.feedback.saveError'))
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
    window.$toast?.success(t('common.feedback.saved'))
  } catch (e) {
    window.$toast?.error(t('common.feedback.saveError'))
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
    window.$toast?.success(t('common.feedback.deleted'))
  } catch (e) {
    window.$toast?.error(t('common.feedback.deleteError'))
  }
}

onMounted(loadData)

// Échap ferme la modale
useEscapeClose(showModal, () => closeModal())
</script>

<style scoped>
.gm-section { margin-block-end: 28px; }
.gm-section__head { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-block-end: 10px; }
.gm-section__title { margin: 0 0 10px; font-size: 15px; font-weight: 600; color: var(--text); }
.gm-section__head .gm-section__title { margin: 0; }

.gm-mono { font-family: var(--f-num); font-size: 11.5px; letter-spacing: 0.04em; color: var(--muted); }
.gm-rule { display: flex; flex-direction: column; gap: 2px; }
.gm-points { max-width: 110px; }
.gm-check { display: inline-flex; align-items: center; gap: 8px; font-size: 13px; color: var(--text); }

.gm-badge { display: flex; align-items: flex-start; gap: 10px; }

.gm-badge__icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 30px;
  height: 30px;
  border: 1px solid currentColor;
  border-radius: var(--r-s);
  color: color-mix(in oklab, var(--c, var(--acc-ink)) 78%, var(--text));
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
  border-radius: var(--r-s);
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
  border-radius: var(--r-s);
  background: var(--surface);
  cursor: pointer;
}

.gm-preset { width: 22px; height: 22px; padding: 0; border: 1px solid var(--line); border-radius: var(--r-s); cursor: pointer; }
.gm-preset:focus-visible { outline: 2px solid var(--acc); outline-offset: 2px; }

.gm-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--sunken);
}

.gm-preview__badge {
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border: 1px solid currentColor;
  border-radius: var(--r-s);
  color: color-mix(in oklab, var(--c) 78%, var(--text));
}
.gm-preview__icon { width: 24px; height: 24px; }
.gm-preview__name { margin: 0; font-weight: 600; color: var(--text); }

@media (min-width: 768px) {
  .gm-form { grid-template-columns: 1fr 1fr; }
  .gm-form__wide { grid-column: 1 / -1; }
}
</style>
