<template>
  <div
    v-if="show"
    class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
    @click.self="$emit('close')"
  >
    <div class="bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto">
      <!-- En-tête -->
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ isEdit ? $t('chatbots.edit') : $t('chatbots.create') }}
        </h2>
        <button type="button" @click="$emit('close')"
          class="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors">
          <Icon icon="mdi:close" class="h-5 w-5 text-gray-500" />
        </button>
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-4">

        <!-- Nom -->
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            {{ $t('chatbots.nameField') }} *
          </label>
          <input v-model="form.name" type="text" required
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" />
        </div>

        <!-- Description -->
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            {{ $t('chatbots.descriptionField') }}
          </label>
          <textarea v-model="form.description" rows="2"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"></textarea>
        </div>

        <!-- Webhook URL -->
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            {{ $t('chatbots.webhookUrl') }} *
          </label>
          <input v-model="form.webhook_url" type="url" required
            placeholder="https://ai.example.com/webhook/.../chat"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 font-mono text-sm" />
        </div>

        <!-- Icône + Couleur (sur la même ligne) -->
        <div class="grid grid-cols-2 gap-4">
          <!-- Icône -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ $t('chatbots.iconField') }}
            </label>
            <div class="flex items-center gap-2">
              <div class="h-9 w-9 rounded-lg flex items-center justify-center flex-shrink-0"
                :style="{ background: form.color + '22', border: '1px solid ' + form.color }">
                <Icon :icon="form.icon || 'mdi:robot-outline'" class="h-5 w-5" :style="{ color: form.color }" />
              </div>
              <input v-model="form.icon" type="text" placeholder="mdi:robot-outline"
                class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500 font-mono text-xs" />
            </div>
          </div>

          <!-- Couleur -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ $t('chatbots.colorField') }}
            </label>
            <div class="flex items-center gap-2">
              <input v-model="form.color" type="color"
                class="h-9 w-12 rounded cursor-pointer border border-gray-300 dark:border-gray-600 p-0.5 bg-white dark:bg-gray-700" />
              <input v-model="form.color" type="text" placeholder="#4f46e5" maxlength="7"
                class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500 font-mono text-sm" />
            </div>
          </div>
        </div>

        <!-- Séparateur -->
        <div class="border-t border-gray-200 dark:border-gray-700 pt-4">
          <p class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">
            {{ $t('chatbots.welcomeSection') }}
          </p>

          <!-- Prévisualisation de l'en-tête -->
          <div class="rounded-lg overflow-hidden mb-3 shadow-sm">
            <div class="px-4 py-3 flex items-center gap-3"
              :style="{ background: 'linear-gradient(135deg, ' + form.color + 'dd, ' + form.color + ')' }">
              <div class="h-8 w-8 rounded-lg bg-white/20 flex items-center justify-center flex-shrink-0">
                <Icon :icon="form.icon || 'mdi:robot-outline'" class="h-5 w-5 text-white" />
              </div>
              <div>
                <p class="font-semibold text-white text-sm leading-tight">
                  {{ form.welcome_title || $t('chatbots.welcomeTitlePlaceholder') }}
                </p>
                <p class="text-white/75 text-xs leading-tight mt-0.5">
                  {{ form.welcome_subtitle || $t('chatbots.welcomeSubtitlePlaceholder') }}
                </p>
              </div>
            </div>
          </div>

          <!-- Titre d'accueil -->
          <div class="mb-3">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ $t('chatbots.welcomeTitle') }}
            </label>
            <input v-model="form.welcome_title" type="text"
              :placeholder="$t('chatbots.welcomeTitlePlaceholder')"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" />
          </div>

          <!-- Sous-titre d'accueil -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ $t('chatbots.welcomeSubtitle') }}
            </label>
            <input v-model="form.welcome_subtitle" type="text"
              :placeholder="$t('chatbots.welcomeSubtitlePlaceholder')"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" />
          </div>
        </div>

        <!-- Messages initiaux -->
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            {{ $t('chatbots.initialMessages') }}
          </label>
          <textarea v-model="initialMessagesText" rows="3"
            :placeholder="$t('chatbots.initialMessagesPlaceholder')"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"></textarea>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ $t('chatbots.initialMessagesHelp') }}
          </p>
        </div>

        <!-- Masquer l'en-tête interne n8n -->
        <div class="flex items-start gap-2">
          <input v-model="form.hide_header" type="checkbox" id="chatbot_hide_header"
            class="w-4 h-4 mt-0.5 text-indigo-600 rounded focus:ring-2 focus:ring-indigo-500 flex-shrink-0" />
          <div>
            <label for="chatbot_hide_header" class="text-sm font-medium text-gray-700 dark:text-gray-300 cursor-pointer">
              {{ $t('chatbots.hideHeader') }}
            </label>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ $t('chatbots.hideHeaderHelp') }}</p>
          </div>
        </div>

        <!-- Actif -->
        <div class="flex items-center gap-2">
          <input v-model="form.is_active" type="checkbox" id="chatbot_is_active"
            class="w-4 h-4 text-indigo-600 rounded focus:ring-2 focus:ring-indigo-500" />
          <label for="chatbot_is_active" class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ $t('chatbots.isActive') }}
          </label>
        </div>

        <!-- Actions -->
        <div class="flex gap-3 pt-4">
          <button type="button" @click="$emit('close')"
            class="flex-1 px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white rounded-lg font-medium hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors">
            {{ $t('common.cancel') }}
          </button>
          <button type="submit" :disabled="isSaving"
            class="flex-1 px-4 py-2 text-white rounded-lg font-medium transition-colors disabled:opacity-50"
            :style="{ background: form.color }">
            {{ isSaving ? $t('common.saving') : $t('common.save') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { chatbotsAdminService } from '@/services/api'

const props = defineProps({
  show: Boolean,
  isEdit: Boolean,
  chatbot: { type: Object, default: null }
})

const emit = defineEmits(['close', 'saved'])

const isSaving = ref(false)
const initialMessagesText = ref('')

const defaultForm = () => ({
  name: '',
  description: '',
  webhook_url: '',
  icon: 'mdi:robot-outline',
  color: '#4f46e5',
  welcome_title: '',
  welcome_subtitle: '',
  hide_header: false,
  is_active: true
})

const form = ref(defaultForm())

watch(
  () => props.chatbot,
  (val) => {
    if (val) {
      form.value = {
        name: val.name || '',
        description: val.description || '',
        webhook_url: val.webhook_url || '',
        icon: val.icon || 'mdi:robot-outline',
        color: val.color || '#4f46e5',
        welcome_title: val.welcome_title || '',
        welcome_subtitle: val.welcome_subtitle || '',
        hide_header: val.hide_header === true,
        is_active: val.is_active !== false
      }
      try {
        const msgs = JSON.parse(val.initial_messages || '[]')
        initialMessagesText.value = Array.isArray(msgs) ? msgs.join('\n') : ''
      } catch {
        initialMessagesText.value = ''
      }
    } else {
      form.value = defaultForm()
      initialMessagesText.value = ''
    }
  },
  { immediate: true }
)

const handleSubmit = async () => {
  isSaving.value = true
  try {
    const msgs = initialMessagesText.value
      .split('\n')
      .map((s) => s.trim())
      .filter(Boolean)

    const payload = {
      ...form.value,
      initial_messages: JSON.stringify(msgs)
    }

    if (props.isEdit && props.chatbot) {
      await chatbotsAdminService.updateChatbot(props.chatbot.id, payload)
    } else {
      await chatbotsAdminService.createChatbot(payload)
    }

    emit('saved')
  } catch (error) {
    console.error('Error saving chatbot:', error)
  } finally {
    isSaving.value = false
  }
}
</script>
