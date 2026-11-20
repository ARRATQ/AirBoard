<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ $t('chatbots.title') }}</template>
      <template #description>{{ $t('chatbots.subtitle') }}</template>
      <template #actions>
        <BaseButton variant="primary" icon="mdi:plus" @click="openCreateModal">
          {{ $t('chatbots.new') }}
        </BaseButton>
      </template>
    </PageHeader>

    <EmptyState
      v-if="!isLoading && chatbots.length === 0"
      icon="mdi:robot-outline"
      :title="$t('chatbots.emptyTitle')"
      :description="$t('chatbots.emptyText')"
    >
      <BaseButton variant="primary" size="sm" icon="mdi:plus" @click="openCreateModal">
        {{ $t('chatbots.create') }}
      </BaseButton>
    </EmptyState>

    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :loading="isLoading"
      :search-keys="['name', 'description', 'webhookUrl']"
      :page-size="25"
    >
      <template #cell-name="{ row }">
        <span class="cb-name">
          <Icon :icon="row.raw.icon || 'mdi:robot-outline'" class="cb-name__icon" aria-hidden="true" />
          <span class="cb-name__text">
            <span class="cb-name__title">{{ row.name }}</span>
            <span v-if="row.description" class="cb-name__desc">{{ row.description }}</span>
          </span>
        </span>
      </template>

      <template #cell-webhookUrl="{ row }">
        <span class="cb-url" dir="ltr" :title="row.webhookUrl">{{ row.webhookUrl }}</span>
      </template>

      <template #cell-status="{ row }">
        <span class="cb-status">
          <StatusBadge :tone="row.raw.is_active ? 'ok' : 'neutral'">{{ row.status }}</StatusBadge>
          <button
            type="button"
            role="switch"
            class="cb-switch"
            :class="{ 'cb-switch--on': row.raw.is_active }"
            :aria-checked="row.raw.is_active ? 'true' : 'false'"
            :aria-label="`${row.raw.is_active ? $t('common.deactivate') : $t('common.activate')} : ${row.name}`"
            :title="row.raw.is_active ? $t('common.deactivate') : $t('common.activate')"
            :disabled="togglingId === row.id"
            @click="toggleActive(row.raw)"
          >
            <span class="cb-switch__thumb" />
          </button>
        </span>
      </template>

      <template #cell-actions="{ row }">
        <RowActions :label="row.name" @edit="openEditModal(row.raw)" @delete="confirmDelete(row.raw)" />
      </template>
    </DataTable>

    <ChatbotModal
      :show="showModal"
      :is-edit="isEditMode"
      :chatbot="editingChatbot"
      @close="showModal = false"
      @saved="onSaved"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import RowActions from '@/components/ui/RowActions.vue'
import { useConfirm } from '@/composables/useConfirm'
import { chatbotsAdminService } from '@/services/api'
import ChatbotModal from '@/components/admin/ChatbotModal.vue'

const { t } = useI18n()
const { confirm: askConfirm } = useConfirm()

const chatbots = ref([])
const isLoading = ref(false)
const togglingId = ref(null)
const showModal = ref(false)
const isEditMode = ref(false)
const editingChatbot = ref(null)

const columns = computed(() => [
  { key: 'name', label: t('chatbots.nameField'), sortable: true },
  { key: 'webhookUrl', label: t('chatbots.webhookUrl'), sortable: true, hideSm: true },
  { key: 'status', label: t('common.status'), sortable: true },
  { key: 'actions', label: '', align: 'end', width: '80px' },
])

const rows = computed(() =>
  chatbots.value.map(chatbot => ({
    id: chatbot.id,
    name: chatbot.name,
    description: chatbot.description || '',
    webhookUrl: chatbot.webhook_url || '',
    status: chatbot.is_active ? t('common.active') : t('common.inactive'),
    raw: chatbot,
  })),
)

const loadChatbots = async () => {
  isLoading.value = true
  try {
    const data = await chatbotsAdminService.getChatbots({ limit: 100 })
    chatbots.value = data.data || []
  } catch (error) {
    console.error('Error loading chatbots:', error)
  } finally {
    isLoading.value = false
  }
}

const openCreateModal = () => {
  isEditMode.value = false
  editingChatbot.value = null
  showModal.value = true
}

const openEditModal = (chatbot) => {
  isEditMode.value = true
  editingChatbot.value = { ...chatbot }
  showModal.value = true
}

const onSaved = () => {
  showModal.value = false
  loadChatbots()
}

const toggleActive = async (chatbot) => {
  togglingId.value = chatbot.id
  try {
    await chatbotsAdminService.updateChatbot(chatbot.id, {
      name: chatbot.name,
      description: chatbot.description,
      webhook_url: chatbot.webhook_url,
      icon: chatbot.icon,
      is_active: !chatbot.is_active,
      initial_messages: chatbot.initial_messages
    })
    chatbot.is_active = !chatbot.is_active
  } catch (error) {
    console.error('Error toggling chatbot:', error)
  } finally {
    togglingId.value = null
  }
}

const confirmDelete = async (chatbot) => {
  const confirmed = await askConfirm({
    title: t('chatbots.confirmDeleteTitle'),
    message: t('chatbots.confirmDeleteText', { name: chatbot.name }),
    confirmLabel: t('common.delete'),
    tone: 'danger',
  })
  if (!confirmed) return

  try {
    await chatbotsAdminService.deleteChatbot(chatbot.id)
    await loadChatbots()
  } catch (error) {
    console.error('Error deleting chatbot:', error)
  }
}

onMounted(() => {
  loadChatbots()
})
</script>

<style scoped>
.cb-name { display: flex; align-items: flex-start; gap: 10px; min-width: 0; }
.cb-name__icon { flex: none; width: 20px; height: 20px; margin-block-start: 1px; color: var(--acc-ink); }
.cb-name__text { display: flex; flex-direction: column; min-width: 0; }
.cb-name__title { font-weight: 600; color: var(--text); }

.cb-name__desc {
  max-width: 42ch;
  overflow: hidden;
  font-size: 12px;
  color: var(--muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cb-url {
  display: block;
  max-width: 36ch;
  overflow: hidden;
  font-family: var(--f-num);
  font-size: 11.5px;
  color: var(--muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cb-status { display: inline-flex; align-items: center; gap: 10px; }

.cb-switch {
  position: relative;
  flex: none;
  width: 34px;
  height: 20px;
  padding: 0;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--sunken);
  cursor: pointer;
  transition: background-color 0.15s;
}

.cb-switch--on { border-color: var(--acc); background: var(--acc); }
.cb-switch:disabled { opacity: 0.5; cursor: default; }
.cb-switch:focus-visible { outline: 2px solid var(--acc); outline-offset: 2px; }

.cb-switch__thumb {
  position: absolute;
  inset-block: 2px;
  inset-inline-start: 2px;
  width: 14px;
  border-radius: 2px;
  background: var(--surface);
  box-shadow: 0 0 0 1px var(--line);
  transition: inset-inline-start 0.15s;
}

.cb-switch--on .cb-switch__thumb { inset-inline-start: 16px; }

@media (prefers-reduced-motion: reduce) {
  .cb-switch, .cb-switch__thumb { transition: none; }
}

@media (max-width: 640px) {
  .cb-name__desc { max-width: 22ch; }
}
</style>
