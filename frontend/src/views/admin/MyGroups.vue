<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ $t('myGroups.title') }}</template>
      <template #description>{{ scopeAll ? $t('myGroups.subtitleAll') : $t('myGroups.subtitleGroups') }}</template>
      <template #actions>
        <BaseButton v-if="canCreate" variant="primary" icon="lucide:plus" @click="openCreate(null)">
          {{ scopeAll ? $t('myGroups.newGroup') : $t('myGroups.newSubGroup') }}
        </BaseButton>
      </template>
    </PageHeader>

    <div v-if="loading" class="mg-state">
      <Icon icon="lucide:loader-circle" class="mg-spin" aria-hidden="true" />
      <span class="sr-only">{{ $t('common.loading') }}</span>
    </div>
    <div v-else-if="loadError" class="mg-state mg-state--col">
      <p>{{ loadError }}</p>
      <BaseButton variant="secondary" size="sm" @click="load">{{ $t('common.retry') }}</BaseButton>
    </div>

    <EmptyState
      v-else-if="!groups.length"
      icon="lucide:users-round"
      :title="$t('myGroups.emptyTitle')"
      :description="$t('myGroups.emptyText')"
    />

    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :search-keys="['name', 'membersText']"
      :search-placeholder="$t('myGroups.searchPlaceholder')"
      :page-size="50"
    >
      <template #cell-name="{ row }">
        <span class="mg-name" :style="{ '--mg-depth': row.depth }">
          <Icon v-if="row.depth > 0" icon="lucide:corner-down-right" class="mg-name__branch rtl-flip" aria-hidden="true" />
          <span class="mg-name__swatch" :style="{ backgroundColor: row.raw.color || undefined }">
            <Icon icon="lucide:users" aria-hidden="true" />
          </span>
          <span class="mg-name__text">
            <span class="mg-name__title">{{ row.name }}</span>
            <span v-if="row.parentPath" class="mg-name__path">
              <span class="sr-only">{{ $t('groups.parentLabel') }}</span>{{ row.parentPath }}
            </span>
            <span class="mg-tags">
              <span v-if="!row.raw.editable && row.raw.org_unit_name" class="mg-tag">
                <Icon icon="lucide:network" aria-hidden="true" />{{ $t('groups.linkedUnit', { name: row.raw.org_unit_name }) }}
              </span>
              <span v-else-if="!row.raw.editable" class="mg-tag">
                <Icon icon="lucide:lock" aria-hidden="true" />{{ $t('myGroups.adminOnly') }}
              </span>
            </span>
          </span>
        </span>
      </template>

      <template #cell-members="{ row }">
        <span class="mg-members" :title="row.membersText">
          <span class="mg-mono">{{ $t('groups.membersCount', row.members) }}</span>
          <span v-if="row.preview" class="mg-members__names">{{ row.preview }}</span>
        </span>
      </template>

      <template #cell-actions="{ row }">
        <RowActions
          :label="row.name"
          :editable="row.raw.editable"
          :deletable="row.raw.editable"
          @edit="openEdit(row.raw)"
          @delete="remove(row.raw)"
        >
          <button
            v-if="!row.raw.is_default"
            type="button"
            class="row-actions__btn"
            :aria-label="$t('myGroups.addSubGroupTo', { name: row.name })"
            :title="$t('myGroups.addSubGroup')"
            @click="openCreate(row.raw.id)"
          >
            <Icon icon="lucide:folder-plus" aria-hidden="true" />
          </button>
        </RowActions>
      </template>
    </DataTable>

    <SubGroupModal
      v-if="modal.open"
      :group="modal.group"
      :parent-id="modal.parentId"
      :groups="groups"
      :scope-all="scopeAll"
      @close="closeModal"
      @saved="onSaved"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import RowActions from '@/components/ui/RowActions.vue'
import SubGroupModal from '@/components/admin/SubGroupModal.vue'
import { useConfirm } from '@/composables/useConfirm'
import { useAppStore } from '@/stores/app'
import { userGroupsService } from '@/services/api'
import { orderGroupTree } from '@/utils/groupTree'

const { t } = useI18n()
const appStore = useAppStore()
const { confirm: askConfirm } = useConfirm()

const groups = ref([])
const scopeAll = ref(false)
const loading = ref(true)
const loadError = ref('')

const memberName = (m) => [m.first_name, m.last_name].filter(Boolean).join(' ') || m.username
const MEMBER_PREVIEW = 3

const columns = computed(() => [
  { key: 'name', label: t('ui.table.columns.name') },
  { key: 'members', label: t('ui.table.columns.members'), hideSm: true },
  { key: 'actions', label: '', align: 'end', width: '130px' },
])

const rows = computed(() =>
  orderGroupTree(groups.value).map(({ group, depth, path }) => {
    const names = (group.members || []).map(memberName)
    return {
      id: group.id,
      name: group.name,
      depth,
      parentPath: path.join(' › '),
      members: names.length,
      membersText: names.join(', '),
      preview: names.slice(0, MEMBER_PREVIEW).join(', ') + (names.length > MEMBER_PREVIEW ? '…' : ''),
      raw: group,
    }
  }),
)

// Création possible : sous un groupe visible (hors groupe par défaut), ou à la racine
const canCreate = computed(() => scopeAll.value || groups.value.some((g) => !g.is_default))

const load = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const data = await userGroupsService.list()
    groups.value = Array.isArray(data?.groups) ? data.groups : []
    scopeAll.value = !!data?.can_create_root
  } catch (e) {
    loadError.value = e?.response?.data?.error || t('myGroups.errors.load')
  } finally {
    loading.value = false
  }
}

const modal = reactive({ open: false, group: null, parentId: null })
const openCreate = (parentId) => Object.assign(modal, { open: true, group: null, parentId })
const openEdit = (group) => Object.assign(modal, { open: true, group, parentId: null })
const closeModal = () => Object.assign(modal, { open: false, group: null, parentId: null })

const onSaved = async (message) => {
  closeModal()
  appStore.showSuccess(message)
  await load()
}

const remove = async (group) => {
  const confirmed = await askConfirm({
    title: t('myGroups.confirmDeleteTitle'),
    message: t('myGroups.confirmDeleteText', { name: group.name }),
    confirmLabel: t('common.delete'),
    tone: 'danger',
  })
  if (!confirmed) return
  try {
    await userGroupsService.remove(group.id)
    appStore.showSuccess(t('myGroups.deleted'))
    await load()
  } catch (e) {
    appStore.showError(e?.response?.data?.error || t('myGroups.errors.delete'))
  }
}

onMounted(load)
</script>

<style scoped>
.mg-state { display: grid; place-items: center; min-height: 12rem; color: var(--muted); }
.mg-state--col { gap: 10px; }
.mg-spin { width: 24px; height: 24px; animation: mg-spin 0.9s linear infinite; }
@keyframes mg-spin { to { transform: rotate(360deg); } }

.mg-name {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  min-width: 0;
  padding-inline-start: calc(min(var(--mg-depth, 0), 6) * 18px);
}

.mg-name__branch { flex: none; width: 14px; height: 14px; margin-top: 7px; margin-inline-end: -4px; color: var(--muted); }

.mg-name__swatch {
  flex: none;
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: var(--r-s);
  background: var(--acc);
  color: var(--acc-on, #fff);
  font-size: 16px;
}

.mg-name__text { display: flex; flex-direction: column; min-width: 0; }
.mg-name__title { font-weight: 600; color: var(--text); }
.mg-name__path { font-size: 11.5px; color: var(--muted); }

.mg-tags { display: flex; flex-wrap: wrap; gap: 4px; }
.mg-tags:empty { display: none; }

.mg-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  width: fit-content;
  margin-top: 3px;
  padding: 1px 7px;
  border: 1px solid var(--border);
  border-radius: 999px;
  font-size: 11.5px;
  color: var(--muted);
  background: var(--bg-surface);
}

.mg-tag svg { width: 12px; height: 12px; }

.mg-members { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.mg-members__names { max-width: 36ch; overflow: hidden; font-size: 12px; color: var(--muted); text-overflow: ellipsis; white-space: nowrap; }

.mg-mono {
  font-family: var(--f-num);
  font-size: 11.5px;
  letter-spacing: 0.04em;
  color: var(--muted);
  white-space: nowrap;
}

@media (max-width: 640px) {
  .mg-name { padding-inline-start: calc(min(var(--mg-depth, 0), 3) * 12px); }
}

@media (prefers-reduced-motion: reduce) {
  .mg-spin { animation: none; }
}
</style>
