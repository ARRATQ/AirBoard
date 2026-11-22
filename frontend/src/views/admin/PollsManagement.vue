<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ $t(`${pageKeys}.title`) }}</template>
      <template #description>{{ $t(`${pageKeys}.subtitle`) }}</template>
      <template #actions>
        <BaseButton variant="primary" icon="mdi:plus" @click="openCreateModal">
          {{ $t('polls.adminPollsManagement.newPoll') }}
        </BaseButton>
      </template>
    </PageHeader>

    <!-- Chiffres clés -->
    <dl class="pl-stats">
      <div class="pl-stats__item">
        <dt>{{ $t(`${pageKeys}.stats.total`) }}</dt>
        <dd>{{ stats.total }}</dd>
      </div>
      <div class="pl-stats__item">
        <dt>{{ $t('polls.adminPollsManagement.stats.active') }}</dt>
        <dd>{{ stats.active }}</dd>
      </div>
      <div class="pl-stats__item">
        <dt>{{ $t('polls.adminPollsManagement.stats.closed') }}</dt>
        <dd>{{ stats.closed }}</dd>
      </div>
      <!-- Votes : uniquement en périmètre global (comme l'ancienne page admin) -->
      <div v-if="isGlobal" class="pl-stats__item">
        <dt>{{ $t('polls.adminPollsManagement.stats.votes') }}</dt>
        <dd>{{ stats.votes }}</dd>
      </div>
    </dl>

    <!-- Aucun sondage et aucun filtre : invitation à en créer un -->
    <EmptyState
      v-if="!isLoading && polls.length === 0 && !hasActiveFilters"
      icon="mdi:poll-box-outline"
      :title="$t('polls.adminPollsManagement.empty.title')"
      :description="$t(`${pageKeys}.empty.description`)"
    >
      <BaseButton variant="primary" size="sm" icon="mdi:plus" @click="openCreateModal">
        {{ $t('polls.adminPollsManagement.empty.createButton') }}
      </BaseButton>
    </EmptyState>

    <!-- Recherche, statut et type sont gérés côté serveur (jusqu'à 100 sondages par chargement) -->
    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :loading="isLoading"
      :searchable="false"
      :footer="false"
      :page-size="100"
      :empty-title="$t('polls.adminPollsManagement.empty.title')"
    >
      <template #toolbar>
        <label class="pl-search">
          <Icon icon="mdi:magnify" class="pl-search__icon" aria-hidden="true" />
          <input
            v-model="filters.search"
            type="search"
            class="pl-search__input"
            :placeholder="$t('polls.adminPollsManagement.filters.searchPlaceholder')"
            :aria-label="$t('polls.adminPollsManagement.filters.search')"
            @input="debounceSearch"
          />
        </label>
        <label>
          <span class="sr-only">{{ $t('polls.adminPollsManagement.filters.status') }}</span>
          <select v-model="filters.status" class="ui-select" @change="loadPolls">
            <option value="">{{ $t('polls.adminPollsManagement.filters.status_all') }}</option>
            <option value="active">{{ $t('polls.adminPollsManagement.filters.status_active') }}</option>
            <option value="closed">{{ $t('polls.adminPollsManagement.filters.status_closed') }}</option>
          </select>
        </label>
        <label>
          <span class="sr-only">{{ $t('polls.adminPollsManagement.filters.type') }}</span>
          <select v-model="filters.type" class="ui-select" @change="loadPolls">
            <option value="">{{ $t('polls.adminPollsManagement.filters.type_all') }}</option>
            <option value="single">{{ $t('polls.adminPollsManagement.filters.type_single') }}</option>
            <option value="multiple">{{ $t('polls.adminPollsManagement.filters.type_multiple') }}</option>
          </select>
        </label>
      </template>

      <template #cell-title="{ row }">
        <span class="pl-title">
          <span class="pl-title__line">
            <span class="pl-title__text">{{ row.title }}</span>
            <StatusBadge v-if="row.raw.is_anonymous" tone="info">{{ $t('polls.adminPollsManagement.status.anonymous') }}</StatusBadge>
          </span>
          <span v-if="row.raw.description" class="pl-title__sub">{{ row.raw.description }}</span>
        </span>
      </template>

      <template #cell-typeLabel="{ row }">
        <span class="pl-mono">{{ row.typeLabel }}</span>
      </template>

      <template #cell-status="{ row }">
        <StatusBadge :tone="row.raw.is_active ? 'ok' : 'neutral'">{{ row.status }}</StatusBadge>
      </template>

      <template #cell-date="{ row }">
        <span class="pl-mono">{{ row.dateLabel }}</span>
        <span v-if="row.untilLabel" class="pl-mono pl-mono--block">{{ $t('polls.adminPollsManagement.details.until') }} {{ row.untilLabel }}</span>
      </template>

      <template #cell-optionCount="{ row }">
        <span class="pl-mono">{{ row.optionCount }}</span>
      </template>

      <template #cell-groupNames="{ row }">
        <span v-if="row.groupNames" class="pl-groups">{{ row.groupNames }}</span>
      </template>

      <template #cell-actions="{ row }">
        <RowActions :label="row.title" @edit="openEditModal(row.raw)" @delete="confirmDelete(row.raw)">
          <button
            type="button"
            class="row-actions__btn"
            :title="$t('polls.adminPollsManagement.actions.viewResults')"
            :aria-label="`${$t('polls.adminPollsManagement.actions.viewResults')} : ${row.title}`"
            @click="viewResults(row.raw)"
          >
            <Icon icon="mdi:chart-bar" aria-hidden="true" />
          </button>
          <button
            v-if="row.raw.is_active && canClose(row.raw)"
            type="button"
            class="row-actions__btn"
            :title="$t('polls.adminPollsManagement.actions.close')"
            :aria-label="`${$t('polls.adminPollsManagement.actions.close')} : ${row.title}`"
            @click="closePoll(row.raw)"
          >
            <Icon icon="mdi:lock-outline" aria-hidden="true" />
          </button>
        </RowActions>
      </template>
    </DataTable>

    <!-- Poll Editor Modal -->
    <PollEditor
      :show="showModal"
      :poll="selectedPoll"
      :groups="groups"
      @close="closeModal"
      @success="handleSuccess"
    />

    <!-- Résultats consultables depuis l'admin : l'API les donne toujours à l'admin et à l'auteur -->
    <BaseModal
      :open="Boolean(resultsPoll)"
      :title="resultsPoll ? `${$t('polls.adminPollsManagement.actions.viewResults')} : ${resultsPoll.title}` : ''"
      size="lg"
      @close="resultsPoll = null"
    >
      <PollResults v-if="resultsPoll" :poll="resultsPoll" @view-details="openPollPage(resultsPoll)" />
    </BaseModal>
  </div>
</template>

<script setup>
import StatusBadge from '@/components/ui/StatusBadge.vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import RowActions from '@/components/ui/RowActions.vue'
import { dateLocaleFor } from '@/utils/homeFormat'
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { pollsService, adminService, groupAdminService } from '@/services/api'
import PollEditor from '@/components/polls/PollEditor.vue'
import PollResults from '@/components/polls/PollResults.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import { useI18n } from 'vue-i18n'
import { useConfirm } from '@/composables/useConfirm'
import { useAuthStore } from '@/stores/auth'
import { useManagementScope } from '@/composables/useManagementScope'
import { GROUP_ADMIN } from '@/config/access'

const router = useRouter()
const { t, locale } = useI18n()
const { confirm: askConfirm } = useConfirm()
const authStore = useAuthStore()
// Admin global → tous les sondages (API admin) ; admin de groupe → ceux de ses groupes (API group-admin)
const { isGlobal } = useManagementScope(GROUP_ADMIN)

// Textes d'en-tête propres à chaque périmètre (« Gestion des sondages » / « Mes sondages »)
const pageKeys = computed(() => (isGlobal.value ? 'polls.adminPollsManagement' : 'groupAdmin.pollsManagement'))

const polls = ref([])
const groups = ref([])
const isLoading = ref(false)
const showModal = ref(false)
const selectedPoll = ref(null)

const stats = ref({
  total: 0,
  active: 0,
  closed: 0,
  votes: 0
})

const filters = ref({
  status: '',
  type: '',
  search: ''
})

const hasActiveFilters = computed(() => Boolean(filters.value.status || filters.value.type || filters.value.search))

const columns = computed(() => [
  { key: 'title', label: t('ui.table.columns.title'), sortable: true },
  { key: 'typeLabel', label: t('ui.table.columns.type'), sortable: true, hideSm: true },
  { key: 'status', label: t('ui.table.columns.status'), sortable: true },
  { key: 'author', label: t('ui.table.columns.author'), sortable: true, hideSm: true },
  { key: 'date', label: t('ui.table.columns.date'), sortable: true, hideSm: true },
  { key: 'optionCount', label: t('polls.adminPollsManagement.details.options'), sortable: true, align: 'end', hideSm: true },
  { key: 'groupNames', label: t('ui.table.columns.groups'), hideSm: true },
  { key: 'actions', label: '', align: 'end', width: '150px' },
])

const rows = computed(() =>
  polls.value.map(poll => ({
    id: poll.id,
    title: poll.title,
    typeLabel: poll.poll_type === 'single' ? t('polls.adminPollsManagement.status.singleChoice') : t('polls.adminPollsManagement.status.multipleChoices'),
    status: poll.is_active ? t('polls.adminPollsManagement.status.active') : t('polls.adminPollsManagement.status.closed'),
    author: poll.author ? `${poll.author.first_name} ${poll.author.last_name}`.trim() : '',
    date: poll.created_at || '',
    dateLabel: formatDate(poll.created_at),
    untilLabel: poll.end_date ? formatDate(poll.end_date) : '',
    optionCount: poll.options?.length || 0,
    // Admin de groupe : noms de ses groupes ; admin global : nombre de groupes
    groupNames: poll.target_groups?.length
      ? (isGlobal.value ? t('polls.adminPollsManagement.details.groups', { count: poll.target_groups.length }) : poll.target_groups.map(g => g.name).join(', '))
      : '',
    raw: poll,
  })),
)

let searchTimeout = null

// Charger les sondages
const loadPolls = async () => {
  try {
    isLoading.value = true
    const params = { ...filters.value, page: 1, page_size: 100 }
    const response = isGlobal.value
      ? await pollsService.getPolls(params)
      : await pollsService.getPollsAsGroupAdmin(params)

    const list = response.polls || []
    polls.value = list

    // Calculer les stats
    stats.value = {
      ...stats.value,
      total: response.total ?? list.length,
      active: list.filter(p => p.is_active).length,
      closed: list.filter(p => !p.is_active).length
    }
    // TODO: Charger les stats de votes depuis l'API analytics
  } catch (error) {
    console.error(t('polls.adminPollsManagement.messages.loadError'), error)
  } finally {
    isLoading.value = false
  }
}

// Charger les groupes ciblables : tous (admin global) ou ceux administrés (admin de groupe)
const loadGroups = async () => {
  try {
    if (isGlobal.value) {
      const response = await adminService.getGroups()
      groups.value = response.data || response
    } else {
      const data = await groupAdminService.getManagedGroups()
      groups.value = data || []
    }
  } catch (error) {
    console.error('Error loading groups:', error)
  }
}

// Le backend n'autorise la fermeture qu'à l'admin global ou à l'auteur du sondage
const canClose = (poll) => isGlobal.value || poll.author_id === authStore.user?.id

// Debounce search
const debounceSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    loadPolls()
  }, 500)
}

// Ouvrir le modal de création
const openCreateModal = () => {
  selectedPoll.value = null
  showModal.value = true
}

// Ouvrir le modal d'édition
const openEditModal = (poll) => {
  selectedPoll.value = poll
  showModal.value = true
}

// Fermer le modal
const closeModal = () => {
  showModal.value = false
  selectedPoll.value = null
}

// Gérer le succès
const handleSuccess = () => {
  loadPolls()
}

// Voir les résultats : dans une fenêtre, sans quitter l'admin ni passer par la page de vote
const resultsPoll = ref(null)
const viewResults = (poll) => {
  resultsPoll.value = poll
}
const openPollPage = (poll) => {
  resultsPoll.value = null
  router.push(`/polls/${poll.id}`)
}

// Fermer un sondage
const closePoll = async (poll) => {
  if (!(await askConfirm({ message: t('polls.adminPollsManagement.confirmation.closeMessage'), tone: 'danger', confirmLabel: t('polls.actions.close') }))) {
    return
  }

  try {
    if (isGlobal.value) {
      await pollsService.closePoll(poll.id)
    } else {
      await pollsService.closePollAsGroupAdmin(poll.id)
    }
    if (window.$toast) {
      window.$toast.success(t('polls.adminPollsManagement.messages.closeSuccess'))
    }
    await loadPolls()
  } catch (error) {
    console.error('Erreur lors de la fermeture:', error)
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('polls.adminPollsManagement.messages.closeError'))
    }
  }
}

// Confirmer la suppression
const confirmDelete = async (poll) => {
  if (!(await askConfirm({ message: t('polls.adminPollsManagement.confirmation.deleteMessage', { title: poll.title }), tone: 'danger' }))) {
    return
  }

  try {
    if (isGlobal.value) {
      await pollsService.deletePoll(poll.id)
    } else {
      await pollsService.deletePollAsGroupAdmin(poll.id)
    }
    if (window.$toast) {
      window.$toast.success(t('polls.adminPollsManagement.messages.deleteSuccess'))
    }
    await loadPolls()
  } catch (error) {
    console.error('Erreur lors de la suppression:', error)
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('polls.adminPollsManagement.messages.deleteError'))
    }
  }
}

// Formater la date
const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleDateString(dateLocaleFor(locale.value), { day: '2-digit', month: 'short', year: 'numeric' })
}

onMounted(() => {
  loadPolls()
  loadGroups()
})
</script>

<style scoped>
.pl-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 28px;
  margin: 0 0 16px;
  padding: 0;
}
.pl-stats__item { display: flex; align-items: baseline; gap: 8px; margin: 0; }
.pl-stats__item dt { font-size: 12.5px; color: var(--muted); }
.pl-stats__item dd { margin: 0; font-family: var(--f-num); font-size: 18px; font-weight: 600; color: var(--text); }

.pl-search { position: relative; display: flex; align-items: center; flex: 1 1 220px; max-width: 340px; }
.pl-search__icon { position: absolute; inset-inline-start: 9px; width: 16px; height: 16px; color: var(--muted); pointer-events: none; }
.pl-search__input {
  width: 100%;
  height: 32px;
  padding-inline: 30px 10px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
  font: inherit;
  font-size: 13px;
  color: var(--text);
}
.pl-search__input:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

.pl-title { display: flex; flex-direction: column; min-width: 0; }
.pl-title__line { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; }
.pl-title__text { font-weight: 600; color: var(--text); }
.pl-title__sub {
  max-width: 52ch;
  overflow: hidden;
  font-size: 12px;
  color: var(--muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}
.pl-mono { font-family: var(--f-num); font-size: 11.5px; letter-spacing: 0.04em; color: var(--muted); white-space: nowrap; }
.pl-mono--block { display: block; }
.pl-groups { font-size: 12px; color: var(--muted); }
</style>
