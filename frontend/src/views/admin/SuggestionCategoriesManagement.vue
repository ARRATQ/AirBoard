<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ t('suggestions.manageCategories') }}</template>
      <template #description>{{ t('suggestions.manageCategoriesHelp') }}</template>
      <template #actions>
        <BaseButton variant="primary" icon="mdi:plus" @click="openCreate">{{ t('common.create') }}</BaseButton>
      </template>
    </PageHeader>

    <EmptyState
      v-if="!loading && categories.length === 0"
      icon="mdi:format-list-bulleted-square"
      :title="t('suggestions.noCategory')"
    >
      <BaseButton variant="primary" size="sm" icon="mdi:plus" @click="openCreate">{{ t('common.create') }}</BaseButton>
    </EmptyState>

    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :loading="loading"
      :search-keys="['name', 'slug']"
      :page-size="25"
    >
      <template #cell-name="{ row }">
        <span class="sc-name">{{ row.name }}</span>
      </template>

      <template #cell-slug="{ row }">
        <span class="sc-mono" dir="ltr">{{ row.slug }}</span>
      </template>

      <template #cell-order="{ row }">
        <span class="sc-mono">{{ row.order }}</span>
      </template>

      <template #cell-status="{ row }">
        <label class="sc-active">
          <input type="checkbox" :checked="row.raw.is_active" @change="toggleActive(row.raw, $event.target.checked)" />
          <span>{{ t('common.active') }}</span>
        </label>
      </template>

      <template #cell-actions="{ row }">
        <RowActions :label="row.name" @edit="openEdit(row.raw)" @delete="removeCategory(row.raw)" />
      </template>
    </DataTable>

    <Teleport to="body">
      <div v-if="showModal" class="sc-overlay" @click.self="closeForm">
        <div class="sc-modal" role="dialog" aria-modal="true" :aria-label="editingId ? t('common.edit') : t('common.create')">
          <div class="sc-modal__head">
            <h2 class="sc-modal__title">{{ editingId ? t('common.edit') : t('common.create') }}</h2>
            <button type="button" class="row-actions__btn" :aria-label="t('common.cancel')" @click="closeForm">
              <Icon icon="mdi:close" aria-hidden="true" />
            </button>
          </div>

          <form class="sc-form" @submit.prevent="saveCategory">
            <div>
              <label class="form-label form-label-required">{{ t('suggestions.form.category') }}</label>
              <input v-model="form.name" class="form-input" required />
            </div>

            <div>
              <label class="form-label form-label-required">{{ $t('ui.table.columns.slug') }}</label>
              <input v-model="form.slug" class="form-input" required />
            </div>

            <div>
              <label class="form-label">{{ t('ui.table.columns.order') }}</label>
              <input v-model.number="form.order" type="number" class="form-input" />
            </div>

            <div class="sc-form__check">
              <input id="is_active" v-model="form.is_active" type="checkbox" />
              <label for="is_active">{{ t('common.active') }}</label>
            </div>

            <div class="sc-form__actions">
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
import { useI18n } from 'vue-i18n'
import { useConfirm } from '@/composables/useConfirm'
import { Icon } from '@iconify/vue'
import { suggestionsService } from '@/services/api'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import RowActions from '@/components/ui/RowActions.vue'

const { t } = useI18n()
const { confirm: askConfirm } = useConfirm()

const loading = ref(false)
const categories = ref([])
const showModal = ref(false)
const editingId = ref(null)
const form = ref({
  name: '',
  slug: '',
  order: 0,
  is_active: true
})

const columns = computed(() => [
  { key: 'name', label: t('ui.table.columns.name'), sortable: true },
  { key: 'slug', label: t('ui.table.columns.slug'), sortable: true, hideSm: true },
  { key: 'order', label: t('ui.table.columns.order'), sortable: true, align: 'end', width: '80px', hideSm: true },
  { key: 'status', label: t('ui.table.columns.status') },
  { key: 'actions', label: '', align: 'end', width: '80px' },
])

const rows = computed(() =>
  [...categories.value]
    .sort((a, b) => (a.order || 0) - (b.order || 0))
    .map(category => ({
      id: category.id,
      name: category.name,
      slug: category.slug,
      order: category.order ?? 0,
      raw: category,
    })),
)

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
  showModal.value = true
}

const openEdit = (category) => {
  editingId.value = category.id
  form.value = {
    name: category.name,
    slug: category.slug,
    order: category.order,
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

    showModal.value = false
    await loadCategories()
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.categorySaveError'))
    }
  }
}

const removeCategory = async (category) => {
  const confirmed = await askConfirm({
    title: t('common.delete'),
    message: `${t('common.delete')} "${category.name}" ?`,
    confirmLabel: t('common.delete'),
    tone: 'danger',
  })
  if (!confirmed) return

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

<style scoped>
.sc-name { font-weight: 600; color: var(--text); }

.sc-mono {
  font-family: var(--f-num);
  font-size: 11.5px;
  letter-spacing: 0.04em;
  color: var(--muted);
  white-space: nowrap;
}

.sc-active { display: inline-flex; align-items: center; gap: 6px; font-size: 13px; color: var(--text); cursor: pointer; }

.sc-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: rgb(0 0 0 / 0.5);
}

.sc-modal {
  width: 100%;
  max-width: 640px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
  color: var(--text);
}

.sc-modal__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-block-end: 1px solid var(--line);
}

.sc-modal__title { margin: 0; font-size: 16px; font-weight: 600; }

.sc-form { display: grid; grid-template-columns: 1fr; gap: 14px; padding: 16px; }

.sc-form__check { display: flex; align-items: center; gap: 8px; }
.sc-form__actions { display: flex; justify-content: flex-end; gap: 8px; }

@media (min-width: 768px) {
  .sc-form { grid-template-columns: 1fr 1fr; }
  .sc-form__actions { grid-column: 1 / -1; }
}
</style>
