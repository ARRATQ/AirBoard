<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6">
    <div class="max-w-5xl mx-auto space-y-6">
      <div class="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('admin.newsTypes') }}</h1>
          <p class="text-gray-600 dark:text-gray-400 mt-1">{{ t('admin.newsTypesHelp') }}</p>
        </div>
        <button class="cat-btn cat-btn-primary" @click="openCreate">{{ t('common.create') }}</button>
      </div>

      <div class="card p-5 border border-gray-200 dark:border-gray-700">
        <div v-if="loading" class="loading-state py-8">
          <Icon icon="mdi:loading" class="animate-spin text-2xl text-amber-500" />
        </div>

        <div v-else-if="types.length === 0" class="empty-state py-8">
          <Icon icon="mdi:shape-outline" class="text-4xl text-gray-400" />
          <p>{{ t('admin.noType') }}</p>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="type in types"
            :key="type.id"
            class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 flex items-center justify-between gap-3"
          >
            <div class="flex items-center gap-3 flex-1">
              <Icon :icon="type.icon || 'mdi:shape'" class="text-2xl" :style="{ color: type.color }" />
              <div class="flex-1">
                <p class="font-semibold text-gray-900 dark:text-white">{{ type.name }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  slug: <code class="bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded">{{ type.slug }}</code>
                  · {{ t('common.order') }}: {{ type.order }}
                </p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <label class="text-sm text-gray-600 dark:text-gray-300 flex items-center gap-2 mr-2">
                <input type="checkbox" :checked="type.is_active" @change="toggleActive(type, $event.target.checked)" />
                {{ t('common.active') }}
              </label>
              <button class="cat-btn cat-btn-secondary cat-btn-sm" @click="openEdit(type)">{{ t('common.edit') }}</button>
              <button class="cat-btn cat-btn-danger cat-btn-sm" @click="removeType(type)">{{ t('common.delete') }}</button>
            </div>
          </div>
        </div>
      </div>

      <Teleport to="body">
        <div
          v-if="showModal"
          class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4"
          @click.self="closeForm"
        >
          <div class="w-full max-w-2xl bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700">
            <div class="flex items-center justify-between px-5 py-4 border-b border-gray-200 dark:border-gray-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ editingId ? t('common.edit') : t('common.create') }}
              </h2>
              <button class="cat-btn cat-btn-secondary cat-btn-sm" type="button" @click="closeForm">
                <Icon icon="mdi:close" class="h-4 w-4" />
              </button>
            </div>

            <form class="p-5 space-y-4" @submit.prevent="saveType">
              <!-- Name -->
              <div>
                <label class="form-label form-label-required">{{ t('common.name') }}</label>
                <input v-model="form.name" class="form-input" required />
              </div>

              <!-- Icon and Color Row -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- Icon -->
                <div>
                  <label class="form-label">{{ t('common.icon') }}</label>
                  <div class="flex gap-2">
                    <input v-model="form.icon" class="form-input flex-1" placeholder="mdi:shape" />
                    <div class="flex items-center justify-center px-4 bg-gray-100 dark:bg-gray-700 rounded-lg border border-gray-300 dark:border-gray-600">
                      <Icon :icon="form.icon || 'mdi:shape'" class="text-2xl" :style="{ color: form.color }" />
                    </div>
                  </div>
                </div>

                <!-- Color -->
                <div>
                  <label class="form-label">{{ t('common.color') }}</label>
                  <div class="flex gap-2">
                    <input v-model="form.color" type="color" class="form-input w-16 h-10 p-1" />
                    <input v-model="form.color" class="form-input flex-1" placeholder="#3B82F6" />
                  </div>
                </div>
              </div>

              <!-- Order -->
              <div>
                <label class="form-label">{{ t('common.order') }}</label>
                <input v-model.number="form.order" type="number" class="form-input" />
              </div>

              <!-- Is Active -->
              <div class="flex items-center gap-2 mt-7">
                <input v-model="form.is_active" type="checkbox" id="is_active" />
                <label for="is_active">{{ t('common.active') }}</label>
              </div>

              <!-- Buttons -->
              <div class="flex justify-end gap-2 pt-4">
                <button type="button" class="cat-btn cat-btn-secondary" @click="closeForm">{{ t('common.cancel') }}</button>
                <button type="submit" class="cat-btn cat-btn-primary">{{ t('common.save') }}</button>
              </div>
            </form>
          </div>
        </div>
      </Teleport>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { newsService } from '@/services/api'
import { useNewsTypesStore } from '@/stores/newsTypes'

const { t } = useI18n()
const typesStore = useNewsTypesStore()

const loading = ref(false)
const types = ref([])
const showModal = ref(false)
const editingId = ref(null)
const form = ref({
  name: '',
  icon: 'mdi:shape',
  color: '#3B82F6',
  order: 0,
  is_active: true
})

const loadTypes = async () => {
  try {
    loading.value = true
    const response = await newsService.getTypes()
    types.value = Array.isArray(response) ? response : response.types || []
    // Refresh the store cache
    await typesStore.refreshTypes()
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('admin.typeLoadError'))
    }
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  editingId.value = null
  form.value = {
    name: '',
    icon: 'mdi:shape',
    color: '#3B82F6',
    order: 0,
    is_active: true
  }
  showModal.value = true
}

const openEdit = (type) => {
  editingId.value = type.id
  form.value = {
    name: type.name,
    icon: type.icon || 'mdi:shape',
    color: type.color || '#3B82F6',
    order: type.order || 0,
    is_active: type.is_active
  }
  showModal.value = true
}

const closeForm = () => {
  showModal.value = false
}

const saveType = async () => {
  const payload = {
    name: form.value.name.trim(),
    icon: form.value.icon.trim() || 'mdi:shape',
    color: form.value.color,
    order: Number(form.value.order) || 0,
    is_active: !!form.value.is_active
  }

  try {
    if (editingId.value) {
      await newsService.updateType(editingId.value, payload)
    } else {
      await newsService.createType(payload)
    }

    if (window.$toast) {
      window.$toast.success(t('common.save'))
    }

    showModal.value = false
    await loadTypes()
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('admin.typeSaveError'))
    }
  }
}

const removeType = async (type) => {
  if (!confirm(`${t('common.delete')} "${type.name}" ?`)) return

  try {
    await newsService.deleteType(type.id)
    if (window.$toast) {
      window.$toast.success(t('common.delete'))
    }
    await loadTypes()
  } catch (error) {
    if (window.$toast) {
      const message = error.response?.data?.error || t('admin.typeDeleteError')
      window.$toast.error(message)
    }
  }
}

const toggleActive = async (type, isActive) => {
  try {
    await newsService.updateType(type.id, {
      name: type.name,
      icon: type.icon || 'mdi:shape',
      color: type.color || '#3B82F6',
      order: type.order || 0,
      is_active: isActive
    })
    await loadTypes()
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('admin.typeSaveError'))
    }
  }
}

onMounted(loadTypes)
</script>

<style scoped>
.cat-btn {
  @apply inline-flex items-center justify-center gap-2 rounded-lg font-semibold transition-all duration-200;
  @apply px-4 py-2 text-sm;
}

.cat-btn-sm {
  @apply px-3 py-1.5 text-xs;
}

.cat-btn-primary {
  @apply text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-sm;
  @apply hover:from-amber-600 hover:to-orange-600 hover:shadow-md;
  @apply focus:outline-none focus:ring-2 focus:ring-amber-400 focus:ring-offset-2;
  @apply dark:focus:ring-offset-gray-900;
}

.cat-btn-secondary {
  @apply border border-gray-300 bg-white text-gray-700;
  @apply hover:bg-gray-50 hover:border-gray-400;
  @apply dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700;
  @apply focus:outline-none focus:ring-2 focus:ring-gray-300 focus:ring-offset-2;
  @apply dark:focus:ring-offset-gray-900;
}

.cat-btn-danger {
  @apply border border-red-300 bg-red-50 text-red-700;
  @apply hover:bg-red-100 hover:border-red-400;
  @apply dark:border-red-700 dark:bg-red-900/20 dark:text-red-300 dark:hover:bg-red-900/35;
  @apply focus:outline-none focus:ring-2 focus:ring-red-300 focus:ring-offset-2;
  @apply dark:focus:ring-offset-gray-900;
}

code {
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
}
</style>
