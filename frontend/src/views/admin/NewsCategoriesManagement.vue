<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ t('admin.newsCategories') }}</template>
      <template #description>{{ t('admin.newsCategoriesHelp') }}</template>
      <template #actions>
        <BaseButton variant="primary" icon="mdi:plus" @click="openCreate">{{ t('common.create') }}</BaseButton>
      </template>
    </PageHeader>

    <EmptyState
      v-if="!loading && categories.length === 0"
      icon="mdi:tag-multiple-outline"
      :title="t('admin.noCategory')"
    />

    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :loading="loading"
      :search-keys="['name', 'slug', 'description']"
      :page-size="25"
    >
      <template #cell-name="{ row }">
        <span class="nc-name">
          <Icon :icon="row.raw.icon || 'mdi:tag'" class="nc-name__icon" :style="{ color: row.raw.color || undefined }" aria-hidden="true" />
          <span class="nc-name__text">
            <span class="nc-name__title">{{ row.name }}</span>
            <span v-if="row.description" class="nc-name__desc">{{ row.description }}</span>
          </span>
        </span>
      </template>

      <template #cell-slug="{ row }">
        <code class="nc-mono">{{ row.slug }}</code>
      </template>

      <template #cell-order="{ row }">
        <span class="nc-mono">{{ row.order }}</span>
      </template>

      <template #cell-status="{ row }">
        <label class="nc-toggle">
          <input type="checkbox" :checked="row.raw.is_active" @change="toggleActive(row.raw, $event.target.checked)" />
          <span>{{ row.status }}</span>
        </label>
      </template>

      <template #cell-actions="{ row }">
        <RowActions :label="row.name" @edit="openEdit(row.raw)" @delete="removeCategory(row.raw)" />
      </template>
    </DataTable>

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
            <BaseButton variant="ghost" size="sm" icon="mdi:close" :aria-label="t('common.cancel')" @click="closeForm" />
          </div>

          <form class="p-5 space-y-4" @submit.prevent="saveCategory">
            <!-- Name -->
            <div>
              <label class="form-label form-label-required">{{ t('common.name') }}</label>
              <input v-model="form.name" class="form-input" required />
            </div>

            <!-- Description -->
            <div>
              <label class="form-label">{{ t('common.description') }}</label>
              <textarea v-model="form.description" class="form-input" rows="3" />
            </div>

            <!-- Icon and Color Row -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <!-- Icon -->
              <div>
                <label class="form-label">{{ t('common.icon') }}</label>
                <div class="flex gap-2">
                  <input v-model="form.icon" class="form-input flex-1" placeholder="mdi:tag" />
                  <div class="flex items-center justify-center px-4 bg-gray-100 dark:bg-gray-700 rounded-lg border border-gray-300 dark:border-gray-600">
                    <Icon :icon="form.icon || 'mdi:tag'" class="text-2xl" :style="{ color: form.color }" />
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
              <BaseButton variant="secondary" @click="closeForm">{{ t('common.cancel') }}</BaseButton>
              <BaseButton variant="primary" type="submit">{{ t('common.save') }}</BaseButton>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import RowActions from '@/components/ui/RowActions.vue'
import { useI18n } from 'vue-i18n'
import { useConfirm } from '@/composables/useConfirm'
import { Icon } from '@iconify/vue'
import { newsService } from '@/services/api'

const { t } = useI18n()
const { confirm: askConfirm } = useConfirm()

const loading = ref(false)
const categories = ref([])
const showModal = ref(false)
const editingId = ref(null)
const form = ref({
  name: '',
  description: '',
  icon: 'mdi:tag',
  color: '#3B82F6',
  order: 0,
  is_active: true
})

const loadCategories = async () => {
  try {
    loading.value = true
    const response = await newsService.getCategories()
    categories.value = Array.isArray(response) ? response : response.categories || []
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('admin.categoryLoadError'))
    }
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  editingId.value = null
  form.value = {
    name: '',
    description: '',
    icon: 'mdi:tag',
    color: '#3B82F6',
    order: 0,
    is_active: true
  }
  showModal.value = true
}

const openEdit = (category) => {
  editingId.value = category.id
  form.value = {
    name: category.name,
    description: category.description || '',
    icon: category.icon || 'mdi:tag',
    color: category.color || '#3B82F6',
    order: category.order || 0,
    is_active: category.is_active
  }
  showModal.value = true
}

const closeForm = () => {
  showModal.value = false
}

const saveCategory = async () => {
  const payload = {
    name: form.value.name.trim(),
    description: form.value.description.trim(),
    icon: form.value.icon.trim() || 'mdi:tag',
    color: form.value.color,
    order: Number(form.value.order) || 0,
    is_active: !!form.value.is_active
  }

  try {
    if (editingId.value) {
      await newsService.updateCategory(editingId.value, payload)
    } else {
      await newsService.createCategory(payload)
    }

    if (window.$toast) {
      window.$toast.success(t('common.save'))
    }

    showModal.value = false
    await loadCategories()
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('admin.categorySaveError'))
    }
  }
}

const removeCategory = async (category) => {
  if (!(await askConfirm({ message: `${t('common.delete')} "${category.name}" ?`, tone: 'danger' }))) return

  try {
    await newsService.deleteCategory(category.id)
    if (window.$toast) {
      window.$toast.success(t('common.delete'))
    }
    await loadCategories()
  } catch (error) {
    if (window.$toast) {
      const message = error.response?.data?.error || t('admin.categoryDeleteError')
      window.$toast.error(message)
    }
  }
}

const toggleActive = async (category, isActive) => {
  try {
    await newsService.updateCategory(category.id, {
      name: category.name,
      description: category.description || '',
      icon: category.icon || 'mdi:tag',
      color: category.color || '#3B82F6',
      order: category.order || 0,
      is_active: isActive
    })
    await loadCategories()
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('admin.categorySaveError'))
    }
  }
}

const columns = computed(() => [
  { key: 'name', label: t('ui.table.columns.name'), sortable: true },
  { key: 'slug', label: t('ui.table.columns.slug'), sortable: true, hideSm: true },
  { key: 'order', label: t('ui.table.columns.order'), sortable: true, align: 'end', width: '80px', hideSm: true },
  { key: 'status', label: t('ui.table.columns.status'), sortable: true },
  { key: 'actions', label: '', align: 'end', width: '80px' },
])

const rows = computed(() =>
  categories.value.map(category => ({
    id: category.id,
    name: category.name,
    slug: category.slug,
    description: category.description || '',
    order: category.order || 0,
    status: category.is_active ? t('common.active') : t('common.inactive'),
    raw: category,
  })),
)

onMounted(loadCategories)
</script>

<style scoped>
.nc-name { display: flex; align-items: flex-start; gap: 10px; min-width: 0; }
.nc-name__icon { flex: none; width: 20px; height: 20px; margin-block-start: 1px; color: var(--muted); }
.nc-name__text { display: flex; flex-direction: column; min-width: 0; }
.nc-name__title { font-weight: 600; color: var(--text); }
.nc-name__desc { font-size: 12px; color: var(--muted); }
.nc-mono { font-family: var(--f-num); font-size: 11.5px; letter-spacing: 0.04em; color: var(--muted); white-space: nowrap; }
.nc-toggle { display: inline-flex; align-items: center; gap: 6px; font-size: 12.5px; color: var(--muted); cursor: pointer; }
</style>
