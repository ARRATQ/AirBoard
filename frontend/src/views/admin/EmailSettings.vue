<template>
  <div class="content-area">
    <!-- Header -->
    <div class="page-header">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="page-title">{{ $t('email.title') }}</h1>
          <p class="page-subtitle">{{ $t('email.subtitle') }}</p>
        </div>
      </div>
    </div>

    <div class="max-w-4xl space-y-6">
      <!-- SMTP Configuration Section -->
      <div class="card">
        <div class="card-header flex items-center justify-between">
          <div class="flex items-center gap-3">
            <Icon icon="mdi:email-cog" class="h-6 w-6 text-blue-400" />
            <h3 class="card-title">{{ $t('email.smtpSettings') }}</h3>
          </div>
          <div class="flex items-center gap-2">
            <span :class="['text-sm font-medium', smtpConfig.is_enabled ? 'text-green-400 dark:text-green-300' : 'text-gray-400 dark:text-gray-500']">
              {{ smtpConfig.is_enabled ? $t('email.enabled') : $t('email.disabled') }}
            </span>
            <label class="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" v-model="smtpConfig.is_enabled" class="sr-only peer">
              <div class="w-11 h-6 bg-gray-600 dark:bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-500"></div>
            </label>
          </div>
        </div>

        <div class="p-6 space-y-6">
          <!-- Server Configuration -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-group">
              <label class="form-label form-label-required">{{ $t('email.host') }}</label>
              <input
                v-model="smtpConfig.host"
                type="text"
                class="form-input"
                :placeholder="$t('email.hostPlaceholder')"
              />
            </div>
            <div class="form-group">
              <label class="form-label form-label-required">{{ $t('email.port') }}</label>
              <input
                v-model.number="smtpConfig.port"
                type="number"
                class="form-input"
                placeholder="587"
              />
            </div>
          </div>

          <!-- Credentials -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-group">
              <label class="form-label">{{ $t('email.username') }}</label>
              <input
                v-model="smtpConfig.username"
                type="text"
                class="form-input"
                :placeholder="$t('email.usernamePlaceholder')"
              />
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('email.password') }}</label>
              <div class="relative">
                <input
                  v-model="smtpPassword"
                  :type="showPassword ? 'text' : 'password'"
                  class="form-input pr-10"
                  :placeholder="smtpConfig.id ? $t('email.passwordKeep') : $t('email.passwordPlaceholder')"
                />
                <button
                  type="button"
                  @click="showPassword = !showPassword"
                  class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 dark:text-gray-500 hover:text-white dark:hover:text-gray-300"
                >
                  <Icon :icon="showPassword ? 'mdi:eye-off' : 'mdi:eye'" class="h-5 w-5" />
                </button>
              </div>
            </div>
          </div>

          <!-- From Configuration -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-group">
              <label class="form-label form-label-required">{{ $t('email.fromEmail') }}</label>
              <input
                v-model="smtpConfig.from_email"
                type="email"
                class="form-input"
                :placeholder="$t('email.fromEmailPlaceholder')"
              />
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('email.fromName') }}</label>
              <input
                v-model="smtpConfig.from_name"
                type="text"
                class="form-input"
                placeholder="Airboard"
              />
            </div>
          </div>

          <!-- TLS Options -->
          <div class="flex flex-wrap gap-6">
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="smtpConfig.use_tls" class="form-checkbox" />
              <span class="text-sm text-gray-300 dark:text-gray-400">{{ $t('email.useTLS') }}</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="smtpConfig.use_starttls" class="form-checkbox" />
              <span class="text-sm text-gray-300 dark:text-gray-400">{{ $t('email.useSTARTTLS') }}</span>
            </label>
          </div>

          <!-- Last Test Info -->
          <div v-if="smtpConfig.last_tested_at" class="p-3 rounded-lg" :class="smtpConfig.last_test_success ? 'bg-green-100 dark:bg-green-800/30 border border-green-200 dark:border-green-700' : 'bg-red-100 dark:bg-red-800/30 border border-red-200 dark:border-red-700'">
            <div class="flex items-center gap-2">
              <Icon :icon="smtpConfig.last_test_success ? 'mdi:check-circle' : 'mdi:alert-circle'"
                    :class="smtpConfig.last_test_success ? 'text-green-700 dark:text-green-300' : 'text-red-700 dark:text-red-300'" class="h-5 w-5" />
              <span class="text-sm" :class="smtpConfig.last_test_success ? 'text-green-800 dark:text-green-200' : 'text-red-800 dark:text-red-200'">
                {{ smtpConfig.last_test_success ? $t('email.lastTestSuccess') : $t('email.lastTestFailed') }}
                - {{ formatDate(smtpConfig.last_tested_at) }}
              </span>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex flex-wrap gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
            <button @click="saveSMTPConfig" :disabled="savingSmtp" class="btn btn-primary">
              <Icon v-if="savingSmtp" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
              <Icon v-else icon="mdi:content-save" class="h-4 w-4 mr-2" />
              {{ $t('common.save') }}
            </button>
            <button @click="showTestModal = true" :disabled="!smtpConfig.id" class="btn btn-secondary">
              <Icon icon="mdi:email-fast" class="h-4 w-4 mr-2" />
              {{ $t('email.testEmail') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Email Templates Section -->
      <div class="card">
        <div class="card-header">
          <div class="flex items-center gap-3">
            <Icon icon="mdi:file-document-edit" class="h-6 w-6 text-purple-400" />
            <h3 class="card-title">{{ $t('email.templates') }}</h3>
          </div>
        </div>

        <div class="p-6">
          <!-- Template Tabs -->
          <div class="flex flex-wrap gap-2 mb-6">
            <button
              v-for="template in templates"
              :key="template.type"
              @click="selectTemplate(template)"
              :class="[
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                selectedTemplate?.type === template.type
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-700 dark:bg-gray-600 text-gray-300 dark:text-gray-200 hover:bg-gray-600 dark:hover:bg-gray-500'
              ]"
            >
              <Icon :icon="getTemplateIcon(template.type)" class="h-4 w-4 inline mr-2" />
              {{ getTemplateName(template.type) }}
            </button>
          </div>

          <!-- Selected Template Editor -->
          <div v-if="selectedTemplate" class="space-y-6">
            <!-- Template Enable Toggle -->
            <div class="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
              <div class="flex items-center gap-3">
                <Icon :icon="getTemplateIcon(selectedTemplate.type)" class="h-6 w-6 text-blue-400" />
                <div>
                  <h4 class="font-medium text-white">{{ selectedTemplate.name }}</h4>
                  <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('email.enableTemplate') }}</p>
                </div>
              </div>
              <label class="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" v-model="selectedTemplate.is_enabled" class="sr-only peer">
                <div class="w-11 h-6 bg-gray-600 dark:bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-500"></div>
              </label>
            </div>

            <!-- Subject -->
            <div class="form-group">
              <label class="form-label form-label-required">{{ $t('email.subject') }}</label>
              <input
                v-model="selectedTemplate.subject"
                type="text"
                class="form-input"
                :placeholder="$t('email.subjectPlaceholder')"
              />
            </div>

            <!-- Variables Available -->
            <div class="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
              <div class="flex items-center gap-2 mb-3">
                <Icon icon="mdi:code-braces" class="h-5 w-5 text-yellow-400" />
                <span class="font-medium text-white">{{ $t('email.variables') }}</span>
              </div>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="variable in getVariablesForType(selectedTemplate.type)"
                  :key="variable.name"
                  @click="copyVariable(variable.name)"
                  class="px-2 py-1 bg-gray-600 dark:bg-gray-500 text-yellow-300 dark:text-yellow-200 rounded text-xs font-mono cursor-pointer hover:bg-gray-500 dark:hover:bg-gray-400 transition-colors"
                  :title="variable.description"
                >
                  {{ variable.name }}
                </span>
              </div>
            </div>

            <!-- HTML Body -->
            <div class="form-group">
              <div class="flex items-center justify-between mb-2">
                <label class="form-label form-label-required">{{ $t('email.htmlBody') }}</label>
                <button @click="previewTemplate" class="text-sm text-blue-400 dark:text-blue-300 hover:text-blue-300 dark:hover:text-blue-200 flex items-center gap-1">
                  <Icon icon="mdi:eye" class="h-4 w-4" />
                  {{ $t('email.preview') }}
                </button>
              </div>
              <textarea
                v-model="selectedTemplate.html_body"
                rows="15"
                class="form-textarea font-mono text-sm"
                :placeholder="$t('email.htmlBodyPlaceholder')"
              ></textarea>
            </div>

            <!-- Template Actions -->
            <div class="flex flex-wrap gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
              <button @click="saveTemplate" :disabled="savingTemplate" class="btn btn-primary">
                <Icon v-if="savingTemplate" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
                <Icon v-else icon="mdi:content-save" class="h-4 w-4 mr-2" />
                {{ $t('common.save') }}
              </button>
              <button @click="resetTemplate" class="btn btn-secondary">
                <Icon icon="mdi:restore" class="h-4 w-4 mr-2" />
                {{ $t('email.resetTemplate') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Email Logs Section -->
      <div class="card">
        <div class="card-header flex items-center justify-between">
          <div class="flex items-center gap-3">
            <Icon icon="mdi:history" class="h-6 w-6 text-orange-400" />
            <h3 class="card-title">{{ $t('email.logs') }}</h3>
          </div>
          <button @click="loadLogs" class="btn btn-ghost btn-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">
            <Icon icon="mdi:refresh" class="h-4 w-4" />
          </button>
        </div>

        <div class="p-6">
          <div v-if="logs.length === 0" class="text-center py-8 text-gray-400 dark:text-gray-500">
            <Icon icon="mdi:email-off" class="h-12 w-12 mx-auto mb-3 opacity-50" />
            <p>{{ $t('email.noLogs') }}</p>
          </div>

          <div v-else class="overflow-x-auto">
            <table class="min-w-full">
              <thead>
                <tr class="border-b border-gray-200 dark:border-gray-700">
                  <th class="text-left py-3 px-4 text-sm font-medium text-gray-400 dark:text-gray-500">{{ $t('email.logType') }}</th>
                  <th class="text-left py-3 px-4 text-sm font-medium text-gray-400 dark:text-gray-500">{{ $t('email.logTitle') }}</th>
                  <th class="text-center py-3 px-4 text-sm font-medium text-gray-400 dark:text-gray-500">{{ $t('email.logRecipients') }}</th>
                  <th class="text-center py-3 px-4 text-sm font-medium text-gray-400 dark:text-gray-500">{{ $t('email.logStatus') }}</th>
                  <th class="text-left py-3 px-4 text-sm font-medium text-gray-400 dark:text-gray-500">{{ $t('email.logDate') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="log in logs" :key="log.id" class="border-b border-gray-100 dark:border-gray-700/50 hover:bg-gray-50 dark:hover:bg-gray-700/30">
                  <td class="py-3 px-4">
                    <span class="flex items-center gap-2">
                      <Icon :icon="getTemplateIcon(log.template_type)" class="h-4 w-4 text-gray-400 dark:text-gray-500" />
                      <span class="text-sm text-gray-300 dark:text-gray-400">{{ getTemplateName(log.template_type) }}</span>
                    </span>
                  </td>
                  <td class="py-3 px-4 text-sm text-gray-900 dark:text-white">{{ log.content_title }}</td>
                  <td class="py-3 px-4 text-center">
                    <span class="text-sm">
                      <span class="text-green-400">{{ log.success_count }}</span>
                      <span class="text-gray-500 dark:text-gray-400">/</span>
                      <span class="text-gray-300 dark:text-gray-400">{{ log.recipient_count }}</span>
                    </span>
                  </td>
                  <td class="py-3 px-4 text-center">
                    <span :class="getStatusClass(log.status)" class="px-2 py-1 rounded text-xs font-medium">
                      {{ getStatusText(log.status) }}
                    </span>
                  </td>
                  <td class="py-3 px-4 text-sm text-gray-500 dark:text-gray-400">{{ formatDate(log.created_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- Test Email Modal -->
    <div v-if="showTestModal" class="modal-overlay" @click.self="showTestModal = false">
      <div class="modal-container max-w-md">
        <div class="modal-panel">
          <div class="modal-header">
            <h3 class="text-lg font-semibold text-white">{{ $t('email.testEmail') }}</h3>
            <button @click="showTestModal = false" class="text-gray-400 dark:text-gray-500 hover:text-white dark:hover:text-gray-300">
              <Icon icon="mdi:close" class="h-5 w-5" />
            </button>
          </div>
          <div class="p-6">
            <div class="form-group">
              <label class="form-label form-label-required">{{ $t('email.testEmailAddress') }}</label>
              <input
                v-model="testEmail"
                type="email"
                class="form-input"
                :placeholder="$t('email.testEmailPlaceholder')"
              />
            </div>
          </div>
          <div class="modal-footer">
            <button @click="showTestModal = false" class="btn btn-secondary">{{ $t('common.cancel') }}</button>
            <button @click="sendTestEmail" :disabled="sendingTest" class="btn btn-primary">
              <Icon v-if="sendingTest" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
              <Icon v-else icon="mdi:send" class="h-4 w-4 mr-2" />
              {{ $t('email.send') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Preview Modal -->
    <div v-if="showPreviewModal" class="modal-overlay" @click.self="showPreviewModal = false">
      <div class="modal-container max-w-4xl">
        <div class="modal-panel">
          <div class="modal-header">
            <div>
              <h3 class="text-lg font-semibold text-white">{{ $t('email.preview') }}</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ previewSubject }}</p>
            </div>
            <button @click="showPreviewModal = false" class="text-gray-400 dark:text-gray-500 hover:text-white dark:hover:text-gray-300">
              <Icon icon="mdi:close" class="h-5 w-5" />
            </button>
          </div>
          <div class="p-0 bg-white dark:bg-gray-800">
            <iframe
              :srcdoc="previewBody"
              class="w-full h-[500px] border-0"
              sandbox="allow-same-origin"
            ></iframe>
          </div>
          <div class="modal-footer">
            <button @click="showPreviewModal = false" class="btn btn-secondary">{{ $t('common.close') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { emailService } from '@/services/api'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

// State
const smtpConfig = ref({
  host: '',
  port: 587,
  username: '',
  from_email: '',
  from_name: 'Airboard',
  use_tls: false,
  use_starttls: true,
  is_enabled: false
})
const smtpPassword = ref('')
const showPassword = ref(false)
const savingSmtp = ref(false)

const templates = ref([])
const selectedTemplate = ref(null)
const savingTemplate = ref(false)
const templateVariables = ref({})

const logs = ref([])

const showTestModal = ref(false)
const testEmail = ref('')
const sendingTest = ref(false)

const showPreviewModal = ref(false)
const previewSubject = ref('')
const previewBody = ref('')

// Methods
const loadSMTPConfig = async () => {
  try {
    const config = await emailService.getSMTPConfig()
    smtpConfig.value = config
  } catch (error) {
    console.error('Error loading SMTP config:', error)
  }
}

const saveSMTPConfig = async () => {
  savingSmtp.value = true
  try {
    const data = {
      ...smtpConfig.value,
      password: smtpPassword.value
    }
    const result = await emailService.updateSMTPConfig(data)
    smtpConfig.value = result
    smtpPassword.value = ''
    appStore.showSuccess(t('email.smtpSaved'))
  } catch (error) {
    appStore.showError(error.response?.data?.error || t('email.smtpSaveError'))
  } finally {
    savingSmtp.value = false
  }
}

const loadTemplates = async () => {
  try {
    templates.value = await emailService.getTemplates()
    if (templates.value.length > 0 && !selectedTemplate.value) {
      selectedTemplate.value = { ...templates.value[0] }
    }
  } catch (error) {
    console.error('Error loading templates:', error)
  }
}

const loadTemplateVariables = async () => {
  try {
    templateVariables.value = await emailService.getTemplateVariables()
  } catch (error) {
    console.error('Error loading template variables:', error)
  }
}

const selectTemplate = (template) => {
  selectedTemplate.value = { ...template }
}

const saveTemplate = async () => {
  if (!selectedTemplate.value) return
  savingTemplate.value = true
  try {
    const result = await emailService.updateTemplate(selectedTemplate.value.type, {
      subject: selectedTemplate.value.subject,
      html_body: selectedTemplate.value.html_body,
      plain_text_body: selectedTemplate.value.plain_text_body,
      is_enabled: selectedTemplate.value.is_enabled
    })
    // Update in templates list
    const index = templates.value.findIndex(t => t.type === result.type)
    if (index !== -1) {
      templates.value[index] = result
    }
    selectedTemplate.value = { ...result }
    appStore.showSuccess(t('email.templateSaved'))
  } catch (error) {
    appStore.showError(error.response?.data?.error || t('email.templateSaveError'))
  } finally {
    savingTemplate.value = false
  }
}

const resetTemplate = async () => {
  if (!selectedTemplate.value) return
  if (!confirm(t('email.resetConfirm'))) return

  try {
    const result = await emailService.resetTemplate(selectedTemplate.value.type)
    const index = templates.value.findIndex(t => t.type === result.type)
    if (index !== -1) {
      templates.value[index] = result
    }
    selectedTemplate.value = { ...result }
    appStore.showSuccess(t('email.templateReset'))
  } catch (error) {
    appStore.showError(error.response?.data?.error || t('email.templateResetError'))
  }
}

const previewTemplate = async () => {
  if (!selectedTemplate.value) return
  try {
    const result = await emailService.previewTemplate(selectedTemplate.value.type)
    previewSubject.value = result.subject
    previewBody.value = result.body
    showPreviewModal.value = true
  } catch (error) {
    appStore.showError(error.response?.data?.error || t('email.previewError'))
  }
}

const loadLogs = async () => {
  try {
    logs.value = await emailService.getLogs()
  } catch (error) {
    console.error('Error loading logs:', error)
  }
}

const sendTestEmail = async () => {
  if (!testEmail.value) return
  sendingTest.value = true
  try {
    await emailService.testSMTPConfig(testEmail.value)
    appStore.showSuccess(t('email.testSuccess'))
    showTestModal.value = false
    testEmail.value = ''
    await loadSMTPConfig()
  } catch (error) {
    appStore.showError(error.response?.data?.error || t('email.testFailed'))
  } finally {
    sendingTest.value = false
  }
}

const copyVariable = (variable) => {
  navigator.clipboard.writeText(variable)
  appStore.showSuccess(t('email.variableCopied'))
}

const getTemplateIcon = (type) => {
  const icons = {
    news: 'mdi:newspaper',
    application: 'mdi:application',
    event: 'mdi:calendar',
    announcement: 'mdi:bullhorn'
  }
  return icons[type] || 'mdi:email'
}

const getTemplateName = (type) => {
  const translations = {
    news: t('email.templateNews'),
    application: t('email.templateApplication'),
    event: t('email.templateEvent'),
    announcement: t('email.templateAnnouncement')
  }
  return translations[type] || type
}

const getVariablesForType = (type) => {
  return templateVariables.value[type] || []
}

const getStatusClass = (status) => {
  const classes = {
    completed: 'bg-green-100 dark:bg-green-800/50 text-green-800 dark:text-green-200 border border-green-200 dark:border-green-700',
    sending: 'bg-yellow-100 dark:bg-yellow-800/50 text-yellow-800 dark:text-yellow-200 border border-yellow-200 dark:border-yellow-700',
    failed: 'bg-red-100 dark:bg-red-800/50 text-red-800 dark:text-red-200 border border-red-200 dark:border-red-700',
    pending: 'bg-gray-100 dark:bg-gray-600 text-gray-700 dark:text-gray-200 border border-gray-200 dark:border-gray-700'
  }
  return classes[status] || classes.pending
}

const getStatusText = (status) => {
  const translations = {
    pending: t('email.statusPending'),
    sending: t('email.statusSending'),
    completed: t('email.statusCompleted'),
    failed: t('email.statusFailed')
  }
  return translations[status] || status
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString()
}

// Lifecycle
onMounted(async () => {
  await Promise.all([
    loadSMTPConfig(),
    loadTemplates(),
    loadTemplateVariables(),
    loadLogs()
  ])
})
</script>

<style scoped>
.form-checkbox {
  @apply h-4 w-4 rounded border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-600 text-blue-500 focus:ring-blue-500 focus:ring-offset-white dark:focus:ring-offset-gray-900;
}
</style>
