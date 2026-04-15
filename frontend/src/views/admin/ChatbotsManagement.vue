<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6">
    <div class="w-full">
      <!-- Header -->
      <div class="flex justify-between items-center mb-8">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
            {{ $t('chatbots.title') }}
          </h1>
          <p class="mt-1 text-gray-600 dark:text-gray-400">
            {{ $t('chatbots.subtitle') }}
          </p>
        </div>
        <button
          @click="openCreateModal"
          class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors flex items-center gap-2"
        >
          <Icon icon="mdi:plus" class="h-5 w-5" />
          {{ $t('chatbots.new') }}
        </button>
      </div>

      <!-- Loading State -->
      <div v-if="isLoading" class="flex justify-center items-center py-12">
        <Icon icon="mdi:loading" class="h-8 w-8 animate-spin text-blue-600" />
      </div>

      <!-- Table -->
      <div v-else-if="chatbots.length > 0" class="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                {{ $t('chatbots.nameField') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                {{ $t('chatbots.webhookUrl') }}
              </th>
              <th class="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                {{ $t('common.status') }}
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                {{ $t('common.actions') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
            <tr
              v-for="chatbot in chatbots"
              :key="chatbot.id"
              class="hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
            >
              <!-- Nom + icône + description -->
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="h-9 w-9 rounded-lg bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center flex-shrink-0">
                    <Icon :icon="chatbot.icon || 'mdi:robot-outline'" class="h-5 w-5 text-indigo-600 dark:text-indigo-400" />
                  </div>
                  <div>
                    <p class="font-medium text-gray-900 dark:text-white">{{ chatbot.name }}</p>
                    <p v-if="chatbot.description" class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate max-w-xs">
                      {{ chatbot.description }}
                    </p>
                  </div>
                </div>
              </td>

              <!-- Webhook URL tronqué -->
              <td class="px-6 py-4">
                <p class="text-sm text-gray-600 dark:text-gray-300 font-mono truncate max-w-xs" :title="chatbot.webhook_url">
                  {{ chatbot.webhook_url }}
                </p>
              </td>

              <!-- Toggle actif/inactif -->
              <td class="px-6 py-4 text-center">
                <button
                  @click="toggleActive(chatbot)"
                  :disabled="togglingId === chatbot.id"
                  :title="chatbot.is_active ? $t('common.deactivate') : $t('common.activate')"
                  class="relative inline-flex items-center h-6 w-11 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50"
                  :class="chatbot.is_active ? 'bg-indigo-600' : 'bg-gray-300 dark:bg-gray-600'"
                >
                  <span
                    class="inline-block h-4 w-4 rounded-full bg-white shadow transform transition-transform"
                    :class="chatbot.is_active ? 'translate-x-6' : 'translate-x-1'"
                  />
                </button>
              </td>

              <!-- Actions -->
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end gap-2">
                  <button
                    @click="openEditModal(chatbot)"
                    class="p-2 text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
                    :title="$t('common.edit')"
                  >
                    <Icon icon="mdi:pencil" class="h-4 w-4" />
                  </button>
                  <button
                    @click="confirmDelete(chatbot)"
                    class="p-2 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                    :title="$t('common.delete')"
                  >
                    <Icon icon="mdi:delete" class="h-4 w-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-12">
        <Icon icon="mdi:robot-outline" class="h-16 w-16 mx-auto text-gray-400 mb-4" />
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
          {{ $t('chatbots.emptyTitle') }}
        </h3>
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          {{ $t('chatbots.emptyText') }}
        </p>
        <button
          @click="openCreateModal"
          class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
        >
          {{ $t('chatbots.create') }}
        </button>
      </div>

      <!-- Modal création/édition -->
      <ChatbotModal
        :show="showModal"
        :is-edit="isEditMode"
        :chatbot="editingChatbot"
        @close="showModal = false"
        @saved="onSaved"
      />

      <!-- Modal confirmation suppression -->
      <div
        v-if="showDeleteConfirm"
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
        @click.self="showDeleteConfirm = false"
      >
        <div class="bg-white dark:bg-gray-800 rounded-lg max-w-md w-full p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="h-12 w-12 bg-red-100 dark:bg-red-900 rounded-full flex items-center justify-center">
              <Icon icon="mdi:alert" class="h-6 w-6 text-red-600 dark:text-red-400" />
            </div>
            <h3 class="text-xl font-bold text-gray-900 dark:text-white">
              {{ $t('chatbots.confirmDeleteTitle') }}
            </h3>
          </div>

          <p class="text-gray-700 dark:text-gray-300 mb-6">
            {{ $t('chatbots.confirmDeleteText', { name: chatbotToDelete?.name }) }}
          </p>

          <div class="flex gap-3">
            <button
              @click="showDeleteConfirm = false"
              class="flex-1 px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white rounded-lg font-medium hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
            >
              {{ $t('common.cancel') }}
            </button>
            <button
              @click="deleteChatbot"
              :disabled="isDeleting"
              class="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-red-400 text-white rounded-lg font-medium transition-colors"
            >
              {{ isDeleting ? $t('common.deleting') : $t('common.delete') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { chatbotsAdminService } from '@/services/api'
import ChatbotModal from '@/components/admin/ChatbotModal.vue'

const chatbots = ref([])
const isLoading = ref(false)
const isDeleting = ref(false)
const togglingId = ref(null)
const showModal = ref(false)
const isEditMode = ref(false)
const editingChatbot = ref(null)
const showDeleteConfirm = ref(false)
const chatbotToDelete = ref(null)

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

const confirmDelete = (chatbot) => {
  chatbotToDelete.value = chatbot
  showDeleteConfirm.value = true
}

const deleteChatbot = async () => {
  isDeleting.value = true
  try {
    await chatbotsAdminService.deleteChatbot(chatbotToDelete.value.id)
    showDeleteConfirm.value = false
    await loadChatbots()
  } catch (error) {
    console.error('Error deleting chatbot:', error)
  } finally {
    isDeleting.value = false
  }
}

onMounted(() => {
  loadChatbots()
})
</script>
