<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6">
    <div class="max-w-5xl mx-auto space-y-6">
      <div class="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('suggestions.manageCategories') }}</h1>
          <p class="text-gray-600 dark:text-gray-400 mt-1">{{ t('suggestions.manageCategoriesHelp') }}</p>
        </div>
        <button class="btn-primary" @click="openCreate">{{ t('common.create') }}</button>
      </div>

      <div class="card p-5 border border-gray-200 dark:border-gray-700">
        <div v-if="loading" class="loading-state py-8">
          <Icon icon="mdi:loading" class="animate-spin text-2xl text-amber-500" />
        </div>

        <div v-else-if="categories.length === 0" class="empty-state py-8">
          <Icon icon="mdi:format-list-bulleted-square" class="text-4xl text-gray-400" />
          <p>{{ t('suggestions.noCategory') }}</p>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="category in categories"
            :key="category.id"
            class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 flex items-center justify-between gap-3"
          >
            <div>
              <p class="font-semibold text-gray-900 dark:text-white">{{ category.name }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400">slug: {{ category.slug }} · order: {{ category.order }}</p>
            </div>
            <div class="flex items-center gap-2">
              <label class="text-sm text-gray-600 dark:text-gray-300 flex items-center gap-2 mr-2">
                <input type="checkbox" :checked="category.is_active" @change="toggleActive(category, $event.target.checked)" />
                {{ t('common.active') }}
              </label>
              <button class="btn-secondary btn-sm" @click="openEdit(category)">{{ t('common.edit') }}</button>
              <button class="btn-secondary btn-sm text-red-600" @click="removeCategory(category)">{{ t('common.delete') }}</button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="showForm" class="card p-5 border border-gray-200 dark:border-gray-700">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          {{ editingId ? t('common.edit') : t('common.create') }}
        </h2>

        <form class="grid grid-cols-1 md:grid-cols-2 gap-4" @submit.prevent="saveCategory">
          <div>
            <label class="form-label form-label-required">{{ t('suggestions.form.category') }}</label>
            <input v-model="form.name" class="form-input" required />
          </div>

          <div>
            <label class="form-label form-label-required">Slug</label>
            <input v-model="form.slug" class="form-input" required />
          </div>

          <div>
            <label class="form-label">Order</label>
            <input v-model.number="form.order" type="number" class="form-input" />
          </div>

          <div class="flex items-center gap-2 mt-7">
            <input v-model="form.is_active" type="checkbox" id="is_active" />
            <label for="is_active">{{ t('common.active') }}</label>
          </div>

          <div class="md:col-span-2 flex justify-end gap-2">
            <button type="button" class="btn-secondary" @click="closeForm">{{ t('common.cancel') }}</button>
            <button type="submit" class="btn-primary">{{ t('common.save') }}</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { suggestionsService } from '@/services/api'

const { t } = useI18n()

const loading = ref(false)
const categories = ref([])
const showForm = ref(false)
const editingId = ref(null)
const form = ref({
  name: '',
  slug: '',
  order: 0,
  is_active: true
})

const loadCategories = async () => {
  try {
    loading.value = true
    const response = await suggestionsService.getAdminCategories({ include_inactive: 1 })
    categories.value = response.categories || []
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.categoryLoadError'))
    }
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  editingId.value = null
  form.value = { name: '', slug: '', order: 0, is_active: true }
  showForm.value = true
}

const openEdit = (category) => {
  editingId.value = category.id
  form.value = {
    name: category.name,
    slug: category.slug,
    order: category.order,
    is_active: category.is_active
  }
  showForm.value = true
}

const closeForm = () => {
  showForm.value = false
}

const saveCategory = async () => {
  const payload = {
    name: form.value.name.trim(),
    slug: form.value.slug.trim(),
    order: Number(form.value.order) || 0,
    is_active: !!form.value.is_active
  }

  try {
    if (editingId.value) {
      await suggestionsService.updateCategory(editingId.value, payload)
    } else {
      await suggestionsService.createCategory(payload)
    }

    if (window.$toast) {
      window.$toast.success(t('common.save'))
    }

    showForm.value = false
    await loadCategories()
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.categorySaveError'))
    }
  }
}

const removeCategory = async (category) => {
  if (!confirm(`${t('common.delete')} "${category.name}" ?`)) return

  try {
    await suggestionsService.deleteCategory(category.id)
    if (window.$toast) {
      window.$toast.success(t('common.delete'))
    }
    await loadCategories()
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.categoryDeleteError'))
    }
  }
}

const toggleActive = async (category, isActive) => {
  try {
    await suggestionsService.updateCategory(category.id, {
      name: category.name,
      slug: category.slug,
      order: category.order,
      is_active: isActive
    })
    await loadCategories()
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.categorySaveError'))
    }
  }
}

onMounted(loadCategories)
</script>
