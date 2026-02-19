<template>
  <div class="w-full p-6 space-y-8">
    <div>
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">{{ t('gamificationAdmin.title') }}</h1>
      <p class="text-gray-600 dark:text-gray-400">{{ t('gamificationAdmin.subtitle') }}</p>
    </div>

    <section class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('gamificationAdmin.rules') }}</h2>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full min-w-[720px]">
          <thead>
            <tr class="text-left border-b border-gray-200 dark:border-gray-700 text-sm text-gray-600 dark:text-gray-300">
              <th class="py-2 pr-4">Reason</th>
              <th class="py-2 pr-4">{{ t('gamificationAdmin.points') }}</th>
              <th class="py-2 pr-4">{{ t('common.active') }}</th>
              <th class="py-2">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rule in rules" :key="rule.id" class="border-b border-gray-100 dark:border-gray-700/60">
              <td class="py-3 pr-4 font-mono text-sm text-gray-800 dark:text-gray-200">{{ rule.reason }}</td>
              <td class="py-3 pr-4 w-40">
                <input v-model.number="rule.points" type="number" min="0" class="form-input" />
              </td>
              <td class="py-3 pr-4">
                <input v-model="rule.enabled" type="checkbox" class="form-checkbox" />
              </td>
              <td class="py-3">
                <button class="gm-btn gm-btn-secondary gm-btn-sm" @click="saveRule(rule)">{{ t('common.save') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('gamificationAdmin.badges') }}</h2>
        <button class="gm-btn gm-btn-primary" @click="openCreate">{{ t('common.create') }}</button>
      </div>

      <div class="space-y-3">
        <article
          v-for="badge in badges"
          :key="badge.id"
          class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 flex items-start justify-between gap-4"
        >
          <div class="min-w-0">
            <p class="font-semibold text-gray-900 dark:text-white">{{ badge.name }} <span class="text-xs text-gray-500">({{ badge.code }})</span></p>
            <p class="text-sm text-gray-600 dark:text-gray-300">{{ badge.description }}</p>
            <p class="text-xs text-gray-500 mt-1">{{ badge.trigger_reason || '-' }} · {{ badge.metric || '-' }} · {{ badge.threshold || 0 }}</p>
          </div>
          <div class="flex items-center gap-2">
            <button class="gm-btn gm-btn-secondary gm-btn-sm" @click="openEdit(badge)">{{ t('common.edit') }}</button>
            <button class="gm-btn gm-btn-danger gm-btn-sm" @click="removeBadge(badge)">{{ t('common.delete') }}</button>
          </div>
        </article>
      </div>
    </section>

    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="closeModal">
        <div class="w-full max-w-2xl bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{{ editingId ? t('common.edit') : t('common.create') }}</h3>
          <form class="grid grid-cols-1 md:grid-cols-2 gap-4" @submit.prevent="saveBadge">
            <div>
              <label class="form-label">Code technique</label>
              <input v-model="form.code" class="form-input" placeholder="ex: suggestion_master" required />
              <p class="text-xs text-gray-500 mt-1">Identifiant unique interne (non visible par l'utilisateur).</p>
            </div>

            <div>
              <label class="form-label">Nom du badge</label>
              <input v-model="form.name" class="form-input" :placeholder="t('common.name')" required />
              <p class="text-xs text-gray-500 mt-1">Texte affiché aux utilisateurs.</p>
            </div>

            <div>
              <label class="form-label">Icône</label>
              <input v-model="form.icon" class="form-input" placeholder="mdi:medal" />
              <p class="text-xs text-gray-500 mt-1">Format Iconify, ex: <code>mdi:medal</code>.</p>
            </div>

            <div>
              <label class="form-label">Couleur</label>
              <div class="flex items-center gap-2">
                <input v-model="form.color" class="form-input" placeholder="#FBBF24" />
                <input
                  v-model="form.color"
                  type="color"
                  class="h-10 w-12 p-1 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 cursor-pointer"
                  aria-label="Choisir une couleur"
                />
              </div>
              <div class="flex items-center gap-2 mt-2">
                <button
                  v-for="preset in colorPresets"
                  :key="preset"
                  type="button"
                  class="h-6 w-6 rounded-full border border-white shadow-sm ring-1 ring-gray-300 dark:ring-gray-600"
                  :style="{ backgroundColor: preset }"
                  :title="preset"
                  @click="form.color = preset"
                ></button>
              </div>
              <p class="text-xs text-gray-500 mt-1">Couleur hexadécimale du badge.</p>
            </div>

            <div class="md:col-span-2">
              <label class="form-label">Aperçu du badge</label>
              <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/40 p-4 flex items-center gap-3">
                <div class="h-11 w-11 rounded-full flex items-center justify-center text-white shadow-sm" :style="badgePreviewStyle">
                  <Icon :icon="safePreviewIcon" class="h-6 w-6" />
                </div>
                <div>
                  <p class="font-semibold text-gray-900 dark:text-white">{{ form.name || 'Nom du badge' }}</p>
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ form.icon || 'mdi:medal' }} · {{ form.color || '#FBBF24' }}</p>
                </div>
              </div>
            </div>

            <div>
              <label class="form-label">Bonus XP du badge</label>
              <input v-model.number="form.xp_reward" type="number" class="form-input" :placeholder="t('gamificationAdmin.badgeReward')" />
              <p class="text-xs text-gray-500 mt-1">XP accordé une seule fois au déblocage du badge.</p>
            </div>

            <div>
              <label class="form-label">Catégorie</label>
              <select v-model="form.category" class="form-select">
                <option value="user">user</option>
                <option value="contributor">contributor</option>
              </select>
              <p class="text-xs text-gray-500 mt-1"><code>contributor</code> est réservé aux profils contributeurs.</p>
            </div>

            <div>
              <label class="form-label">Événement déclencheur</label>
              <select v-model="form.trigger_reason" class="form-select">
                <option value="">Sélectionner un trigger</option>
                <option v-for="reason in reasonOptions" :key="reason" :value="reason">{{ reason }}</option>
              </select>
              <p class="text-xs text-gray-500 mt-1">{{ selectedReasonHelp }}</p>
            </div>

            <div>
              <label class="form-label">Métrique de progression</label>
              <select v-model="form.metric" class="form-select">
                <option value="">Sélectionner une métrique</option>
                <option v-for="metric in metricOptions" :key="metric" :value="metric">{{ metric }}</option>
              </select>
              <p class="text-xs text-gray-500 mt-1">{{ selectedMetricHelp }}</p>
            </div>

            <div>
              <label class="form-label">Seuil de déblocage</label>
              <input v-model.number="form.threshold" type="number" min="0" class="form-input" :placeholder="t('gamificationAdmin.threshold')" />
              <p class="text-xs text-gray-500 mt-1">Nombre d'actions nécessaires pour obtenir le badge.</p>
            </div>

            <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300 mt-7"><input v-model="form.is_active" type="checkbox" class="form-checkbox" />Badge actif</label>
            <div class="md:col-span-2">
              <label class="form-label">Description</label>
              <textarea v-model="form.description" rows="3" class="form-textarea" :placeholder="t('common.description')"></textarea>
            </div>
            <div class="md:col-span-2 flex justify-end gap-2">
              <button type="button" class="gm-btn gm-btn-secondary" @click="closeModal">{{ t('common.cancel') }}</button>
              <button type="submit" class="gm-btn gm-btn-primary">{{ t('common.save') }}</button>
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
import { Icon } from '@iconify/vue'
import { gamificationService } from '@/services/api'

const { t } = useI18n()

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
  app_click: 'Se déclenche à chaque ouverture d\'application.',
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
  app_click_distinct: 'Compter le nombre d\'applications différentes ouvertes.',
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
  if (!confirm(`Supprimer ${badge.name} ?`)) return
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
.gm-btn {
  @apply inline-flex items-center justify-center gap-2 rounded-lg font-semibold transition-all duration-200;
  @apply px-4 py-2 text-sm;
}

.gm-btn-sm {
  @apply px-3 py-1.5 text-xs;
}

.gm-btn-primary {
  @apply text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-sm;
  @apply hover:from-amber-600 hover:to-orange-600 hover:shadow-md;
  @apply focus:outline-none focus:ring-2 focus:ring-amber-400 focus:ring-offset-2;
  @apply dark:focus:ring-offset-gray-800;
}

.gm-btn-secondary {
  @apply border border-gray-300 bg-white text-gray-700;
  @apply hover:bg-gray-50 hover:border-gray-400;
  @apply dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700;
  @apply focus:outline-none focus:ring-2 focus:ring-gray-300 focus:ring-offset-2;
  @apply dark:focus:ring-offset-gray-800;
}

.gm-btn-danger {
  @apply border border-red-300 bg-red-50 text-red-700;
  @apply hover:bg-red-100 hover:border-red-400;
  @apply dark:border-red-700 dark:bg-red-900/20 dark:text-red-300 dark:hover:bg-red-900/35;
  @apply focus:outline-none focus:ring-2 focus:ring-red-300 focus:ring-offset-2;
  @apply dark:focus:ring-offset-gray-800;
}
</style>
