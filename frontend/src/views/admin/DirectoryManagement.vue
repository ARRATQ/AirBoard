<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ $t('directory.manageTitle') }}</template>
      <template #description>
        {{ canAdmin ? $t('directory.manageHelpAdmin') : $t('directory.manageHelpGroupAdmin') }}
      </template>
      <template #actions>
        <BaseButton :tag="RouterLink" to="/directory" variant="secondary" icon="lucide:eye">
          {{ $t('directory.viewPublic') }}
        </BaseButton>
      </template>
    </PageHeader>

    <nav class="dir-admin__tabs" role="tablist">
      <button
        v-for="tab in visibleTabs"
        :key="tab.id"
        type="button"
        role="tab"
        :aria-selected="activeTab === tab.id"
        class="dir-admin__tab"
        :class="{ 'dir-admin__tab--active': activeTab === tab.id }"
        @click="setTab(tab.id)"
      >
        <Icon :icon="tab.icon" class="h-4 w-4" />
        {{ $t(tab.label) }}
        <span class="dir-admin__tab-count">{{ tab.count() }}</span>
      </button>
    </nav>

    <div v-if="loading" class="dir-admin__state">
      <Icon icon="lucide:loader-circle" class="dir-admin__spin" aria-hidden="true" />
      <span class="sr-only">{{ $t('common.loading') }}</span>
    </div>
    <div v-else-if="error" class="dir-admin__state dir-admin__state--col">
      <p>{{ error }}</p>
      <BaseButton variant="secondary" size="sm" @click="load">{{ $t('common.retry') }}</BaseButton>
    </div>

    <!-- ====== STRUCTURE ====== -->
    <section v-else-if="activeTab === 'structure'">
      <div class="dir-admin__bar">
        <p class="dir-admin__help">{{ $t('directory.structureHelp') }}</p>
        <div v-if="canAdmin" class="dir-admin__bar-actions">
          <BaseButton
            variant="secondary"
            icon="lucide:refresh-cw"
            :loading="syncing"
            :title="$t('directory.groups.syncHint')"
            @click="syncGroups"
          >
            {{ $t('directory.groups.sync') }}
          </BaseButton>
          <BaseButton variant="primary" icon="lucide:plus" @click="openUnit(null, null)">
            {{ $t('directory.newUnit') }}
          </BaseButton>
        </div>
      </div>

      <EmptyState v-if="!flatUnits.length" icon="lucide:network" :title="$t('directory.noUnits')" />

      <ul class="dir-tree">
        <li
          v-for="unit in flatUnits"
          :key="unit.id"
          class="dir-unit"
          :style="{ '--unit-color': unit.color, marginInlineStart: `${Math.min(unit.depth, 6) * 1.5}rem` }"
        >
          <div class="dir-unit__head">
            <span class="dir-unit__dot" />
            <span class="dir-unit__name">{{ unit.name }}</span>
            <span v-if="unit.code" class="dir-mono">{{ unit.code }}</span>
            <span v-if="canAdmin && groupOf(unit)" class="dir-unit__group">
              <Icon icon="lucide:users" aria-hidden="true" />
              <span class="sr-only">{{ $t('directory.groups.linkedTo') }}</span>{{ groupOf(unit).name }}
            </span>
            <div v-if="canAdmin" class="dir-unit__actions">
              <button
                type="button"
                class="dir-icon-btn"
                :class="{ 'dir-icon-btn--on': groupOf(unit) }"
                :title="$t('directory.groups.button')"
                :aria-label="$t('directory.groups.buttonFor', { name: unit.name })"
                @click="openUnitGroup(unit)"
              >
                <Icon icon="lucide:users-round" />
              </button>
              <button type="button" class="dir-icon-btn" :title="$t('directory.newPosition')" :aria-label="$t('directory.newPosition')" @click="openPosition(null, unit.id)">
                <Icon icon="lucide:briefcase-plus" />
              </button>
              <button type="button" class="dir-icon-btn" :title="$t('directory.newSubUnit')" :aria-label="$t('directory.newSubUnit')" @click="openUnit(null, unit.id)">
                <Icon icon="lucide:folder-tree" />
              </button>
              <button type="button" class="dir-icon-btn" :title="$t('common.edit')" :aria-label="$t('common.edit')" @click="openUnit(unit)">
                <Icon icon="lucide:pencil" />
              </button>
              <button type="button" class="dir-icon-btn dir-icon-btn--danger" :title="$t('common.delete')" :aria-label="$t('common.delete')" @click="removeUnit(unit)">
                <Icon icon="lucide:trash-2" />
              </button>
            </div>
          </div>

          <p v-if="!(unit.positions || []).length" class="dir-muted dir-unit__none">
            {{ $t('directory.noPositions') }}
          </p>
          <div v-for="position in sortedPositions(unit)" :key="position.id" class="dir-pos">
            <div class="dir-pos__title">
              <Icon v-if="position.is_head" icon="lucide:sparkle" class="dir-head" :title="$t('directory.head')" />
              <span>{{ position.title }}</span>
              <template v-if="canAdmin">
                <button type="button" class="dir-icon-btn dir-icon-btn--sm" :title="$t('common.edit')" :aria-label="$t('common.edit')" @click="openPosition(position)">
                  <Icon icon="lucide:pencil" />
                </button>
                <button type="button" class="dir-icon-btn dir-icon-btn--sm dir-icon-btn--danger" :title="$t('common.delete')" :aria-label="$t('common.delete')" @click="removePosition(position)">
                  <Icon icon="lucide:trash-2" />
                </button>
              </template>
            </div>
            <div class="dir-pos__holders">
              <span v-for="a in position.assignments || []" :key="a.id" class="dir-chip">
                <PersonAvatar :person="peopleById.get(a.person_id)" size="xs" />
                <button type="button" class="dir-chip__name" :title="$t('directory.editAssignment')" @click="openAssignment(position, a)">
                  {{ fullName(peopleById.get(a.person_id)) }}
                </button>
                <span v-if="a.is_interim" class="dir-interim">{{ $t('directory.interim') }}</span>
                <button type="button" class="dir-chip__remove" :aria-label="$t('directory.unassign')" @click="removeAssignment(a)">
                  <Icon icon="lucide:x" class="h-3.5 w-3.5" />
                </button>
              </span>
              <span v-if="!(position.assignments || []).length" class="dir-muted">{{ $t('directory.vacant') }}</span>
              <button type="button" class="dir-chip dir-chip--add" @click="openAssignment(position)">
                <Icon icon="lucide:user-plus" class="h-4 w-4" /> {{ $t('directory.assign') }}
              </button>
            </div>
          </div>
        </li>
      </ul>
    </section>

    <!-- ====== PERSONNES ====== -->
    <section v-else-if="activeTab === 'people'">
      <EmptyState v-if="!people.length" icon="lucide:users" :title="$t('directory.noPeople')">
        <BaseButton variant="primary" size="sm" icon="lucide:user-plus" @click="openPerson(null)">{{ $t('directory.newPerson') }}</BaseButton>
      </EmptyState>

      <DataTable
        v-else
        :columns="peopleColumns"
        :rows="peopleRows"
        :search-keys="['name', 'email', 'phone', 'rolesText']"
        :search-placeholder="$t('directory.searchPlaceholder')"
        :page-size="25"
      >
        <template #toolbar>
          <BaseButton variant="primary" icon="lucide:user-plus" @click="openPerson(null)">{{ $t('directory.newPerson') }}</BaseButton>
        </template>

        <template #cell-name="{ row }">
          <div class="dir-person">
            <PersonAvatar :person="row.raw" size="sm" />
            <div>
              <p class="dir-person__name">{{ row.name }}</p>
              <p v-if="row.raw.user_id" class="dir-muted dir-person__link">
                <Icon icon="lucide:link" class="h-3 w-3" aria-hidden="true" /> {{ $t('directory.linked') }}
              </p>
            </div>
          </div>
        </template>

        <template #cell-rolesText="{ row }">
          <p v-for="r in rolesByPerson.get(row.id) || []" :key="r.assignment.id" class="dir-role">
            {{ r.position.title }} <span class="dir-muted">· {{ r.unit.name }}</span>
          </p>
          <span v-if="!(rolesByPerson.get(row.id) || []).length" class="dir-muted">—</span>
        </template>

        <template #cell-email="{ row }">
          <p class="dir-role">{{ row.email }}</p>
          <p class="dir-mono" dir="ltr">{{ row.phone }}</p>
        </template>

        <template #cell-status="{ row }">
          <StatusBadge :tone="row.raw.is_active ? 'ok' : 'neutral'">{{ row.status }}</StatusBadge>
        </template>

        <template #cell-actions="{ row }">
          <RowActions :label="row.name" @edit="openPerson(row.raw)" @delete="removePerson(row.raw)" />
        </template>
      </DataTable>
    </section>

    <!-- ====== CHAMPS DYNAMIQUES (admin) ====== -->
    <section v-else-if="activeTab === 'fields' && canAdmin">
      <div class="dir-admin__bar">
        <p class="dir-admin__help">{{ $t('directory.fieldsHelp') }}</p>
        <BaseButton variant="primary" icon="lucide:plus" @click="openField(null)">{{ $t('directory.newField') }}</BaseButton>
      </div>

      <EmptyState v-if="!fields.length" icon="lucide:square-chevron-down" :title="$t('directory.noCustomFields')" />

      <DataTable
        v-else
        :columns="fieldColumns"
        :rows="fieldRows"
        :search-keys="['label', 'key', 'typeLabel']"
        :page-size="25"
      >
        <template #cell-label="{ row }">
          <span class="dir-field">
            <Icon :icon="FIELD_TYPE_ICONS[row.raw.type]" class="dir-field__icon" aria-hidden="true" />
            <span class="dir-unit__name">{{ row.label }}</span>
          </span>
        </template>

        <template #cell-key="{ row }">
          <span class="dir-mono">{{ row.key }}</span>
        </template>

        <template #cell-flags="{ row }">
          <span class="dir-flags">
            <StatusBadge v-if="row.raw.required" tone="neutral">{{ $t('directory.fieldRequired') }}</StatusBadge>
            <StatusBadge v-if="row.raw.visibility === 'managers'" tone="warn">{{ $t('directory.visibility.managers') }}</StatusBadge>
            <StatusBadge v-if="row.raw.show_on_card" tone="info">{{ $t('directory.onCard') }}</StatusBadge>
          </span>
        </template>

        <template #cell-actions="{ row }">
          <RowActions :label="row.label" @edit="openField(row.raw)" @delete="removeField(row.raw)" />
        </template>
      </DataTable>
    </section>

    <!-- ====== MODALS ====== -->
    <UnitFormModal
      v-if="modal.type === 'unit'"
      :unit="modal.item"
      :parent-id="modal.parentId"
      :flat-units="flatUnits"
      :descendant-ids="descendantIds"
      @close="closeModal"
      @saved="onSaved('directory.saved.unit')"
    />
    <PositionFormModal
      v-if="modal.type === 'position'"
      :position="modal.item"
      :unit-id="modal.parentId"
      :flat-units="flatUnits"
      @close="closeModal"
      @saved="onSaved('directory.saved.position')"
    />
    <AssignmentFormModal
      v-if="modal.type === 'assignment'"
      :position="modal.position"
      :assignment="modal.item"
      :people="people"
      @close="closeModal"
      @saved="onSaved('directory.saved.assignment')"
    />
    <PersonFormModal
      v-if="modal.type === 'person'"
      :person="modal.item"
      :fields="fields"
      @close="closeModal"
      @saved="onSaved('directory.saved.person')"
    />
    <UnitGroupModal
      v-if="modal.type === 'unitGroup'"
      :unit="modal.item"
      :groups="groups"
      :branch-size="descendantIds(modal.item.id).size"
      @close="closeModal"
      @saved="onGroupSaved"
    />
    <FieldFormModal
      v-if="modal.type === 'field'"
      :field="modal.item"
      :next-order="fields.length"
      @close="closeModal"
      @saved="onSaved('directory.saved.field')"
    />
  </div>
</template>

<script setup>
import StatusBadge from '@/components/ui/StatusBadge.vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import RowActions from '@/components/ui/RowActions.vue'
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useConfirm } from '@/composables/useConfirm'
import { Icon } from '@iconify/vue'
import PersonAvatar from '@/components/directory/PersonAvatar.vue'
import UnitFormModal from '@/components/directory/UnitFormModal.vue'
import PositionFormModal from '@/components/directory/PositionFormModal.vue'
import AssignmentFormModal from '@/components/directory/AssignmentFormModal.vue'
import PersonFormModal from '@/components/directory/PersonFormModal.vue'
import FieldFormModal from '@/components/directory/FieldFormModal.vue'
import UnitGroupModal from '@/components/directory/UnitGroupModal.vue'
import { adminService, directoryService } from '@/services/api'
import { useDirectory, fullName, notify, apiError, FIELD_TYPE_ICONS } from '@/components/directory/useDirectory'

const { t } = useI18n()
const { confirm: askConfirm } = useConfirm()
const route = useRoute()
const router = useRouter()

const {
  loading, error, units, people, fields, canAdmin, load,
  peopleById, flatUnits, rolesByPerson, descendantIds,
} = useDirectory()

const tabs = [
  { id: 'structure', icon: 'lucide:network', label: 'directory.tabStructure', count: () => units.value.length },
  { id: 'people', icon: 'lucide:users', label: 'directory.tabPeople', count: () => people.value.length },
  { id: 'fields', icon: 'lucide:square-chevron-down', label: 'directory.tabFields', count: () => fields.value.length, adminOnly: true },
]
const visibleTabs = computed(() => tabs.filter((tab) => !tab.adminOnly || canAdmin.value))

const activeTab = ref(['structure', 'people', 'fields'].includes(route.query.tab) ? route.query.tab : 'structure')
const setTab = (id) => {
  activeTab.value = id
  router.replace({ query: { ...route.query, tab: id, person: undefined } })
}

const peopleColumns = computed(() => [
  { key: 'name', label: t('directory.person'), sortable: true },
  { key: 'rolesText', label: t('directory.positions'), hideSm: true },
  { key: 'email', label: t('directory.contact'), hideSm: true },
  { key: 'status', label: t('directory.status'), sortable: true },
  { key: 'actions', label: '', align: 'end', width: '80px' },
])

const peopleRows = computed(() =>
  people.value.map((p) => ({
    id: p.id,
    name: fullName(p),
    email: p.email || '',
    phone: p.phone || '',
    rolesText: (rolesByPerson.value.get(p.id) || []).map((r) => `${r.position.title} ${r.unit.name}`).join(' '),
    status: p.is_active ? t('common.active') : t('common.inactive'),
    raw: p,
  })),
)

const fieldColumns = computed(() => [
  { key: 'label', label: t('ui.table.columns.name'), sortable: true },
  { key: 'key', label: t('ui.table.columns.key'), sortable: true, hideSm: true },
  { key: 'typeLabel', label: t('ui.table.columns.type'), sortable: true, hideSm: true },
  { key: 'flags', label: t('ui.table.columns.visibility'), hideSm: true },
  { key: 'actions', label: '', align: 'end', width: '80px' },
])

const fieldRows = computed(() =>
  fields.value.map((f) => ({
    id: f.id,
    label: f.label,
    key: f.key,
    typeLabel: t(`directory.fieldTypes.${f.type}`),
    raw: f,
  })),
)

const sortedPositions = (unit) =>
  [...(unit.positions || [])].sort((a, b) => Number(b.is_head) - Number(a.is_head) || a.order - b.order)

// ---- Modals ----
const modal = reactive({ type: null, item: null, parentId: null, position: null })
const openModal = (type, item = null, extra = {}) => Object.assign(modal, { type, item, parentId: null, position: null, ...extra })
const closeModal = () => Object.assign(modal, { type: null, item: null, parentId: null, position: null })

const openUnit = (unit, parentId = null) => openModal('unit', unit, { parentId })
const openPosition = (position, unitId = null) => openModal('position', position, { parentId: unitId })
const openAssignment = (position, assignment = null) => openModal('assignment', assignment, { position })
const openPerson = (person) => openModal('person', person)
const openField = (field) => openModal('field', field)
const openUnitGroup = (unit) => openModal('unitGroup', unit)

// ---- Groupes d'utilisateurs liés aux entités (admin) ----
const groups = ref([])
const groupsById = computed(() => new Map(groups.value.map((g) => [g.id, g])))
const groupOf = (unit) => (unit.group_id ? groupsById.value.get(unit.group_id) : null)
const loadGroups = async () => {
  if (!canAdmin.value) return
  try {
    const data = await adminService.getGroups()
    groups.value = Array.isArray(data) ? data : []
  } catch (e) {
    notify('error', apiError(e, t('directory.groups.loadFailed')))
  }
}

const onGroupSaved = async (message) => {
  closeModal()
  notify('success', message)
  await Promise.all([load(), loadGroups()])
}

const syncing = ref(false)
const syncGroups = async () => {
  syncing.value = true
  try {
    const res = await directoryService.syncGroups()
    notify('success', t('directory.groups.synced', { linked: res.linked_people || 0, added: res.added || 0, removed: res.removed || 0 }))
    await Promise.all([load(), loadGroups()])
  } catch (e) {
    notify('error', apiError(e, t('directory.groups.syncFailed')))
  } finally {
    syncing.value = false
  }
}

const onSaved = async (messageKey) => {
  closeModal()
  notify('success', t(messageKey))
  await load()
}

// ---- Suppressions ----
const confirmAndRun = async (message, action, successKey) => {
  const confirmed = await askConfirm({ message, confirmLabel: t('common.delete'), tone: 'danger' })
  if (!confirmed) return
  try {
    await action()
    notify('success', t(successKey))
    await load()
  } catch (e) {
    notify('error', apiError(e, t('directory.errors.deleteFailed')))
  }
}

// Une unité qui a des sous-unités n'est pas supprimable (l'API répond 409) : le dire avant de demander confirmation
const removeUnit = (unit) => {
  if (flatUnits.value.some((u) => u.parent_id === unit.id)) {
    notify('error', t('directory.errors.hasSubUnits'))
    return
  }
  return confirmAndRun(t('directory.confirm.unit', { name: unit.name }), () => directoryService.deleteUnit(unit.id), 'directory.deleted.unit')
}
const removePosition = (position) =>
  confirmAndRun(t('directory.confirm.position', { name: position.title }), () => directoryService.deletePosition(position.id), 'directory.deleted.position')
const removeAssignment = (a) =>
  confirmAndRun(t('directory.confirm.assignment', { name: fullName(peopleById.value.get(a.person_id)) }), () => directoryService.deleteAssignment(a.id), 'directory.deleted.assignment')
const removePerson = (p) =>
  confirmAndRun(t('directory.confirm.person', { name: fullName(p) }), () => directoryService.deletePerson(p.id), 'directory.deleted.person')
const removeField = (f) =>
  confirmAndRun(t('directory.confirm.field', { name: f.label }), () => directoryService.deleteField(f.id), 'directory.deleted.field')

onMounted(async () => {
  await load()
  loadGroups()
  // Lien profond depuis la fiche publique : ?person=ID ouvre l'édition
  const personId = Number(route.query.person)
  if (personId && peopleById.value.has(personId)) {
    activeTab.value = 'people'
    openPerson(peopleById.value.get(personId))
  }
})
</script>

<style scoped>
.dir-admin__tabs { display: flex; gap: 4px; margin-block-end: 16px; border-block-end: 1px solid var(--line); overflow-x: auto; }

.dir-admin__tab {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  margin-block-end: -1px;
  padding: 10px 14px;
  border: 0;
  border-block-end: 2px solid transparent;
  background: none;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--muted);
  white-space: nowrap;
  cursor: pointer;
}

.dir-admin__tab:hover { color: var(--text); }
.dir-admin__tab--active { color: var(--text); border-block-end-color: var(--acc); }
.dir-admin__tab:focus-visible { outline: 2px solid var(--acc); outline-offset: -2px; }

.dir-admin__tab-count {
  padding: 0 6px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  font-family: var(--f-num);
  font-size: 10.5px;
  color: var(--muted);
}

.dir-admin__bar { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 12px; margin-block-end: 14px; }
.dir-admin__help { max-width: 640px; margin: 0; font-size: 13px; color: var(--muted); }

.dir-admin__state { display: flex; justify-content: center; padding-block: 56px; color: var(--muted); }
.dir-admin__state--col { flex-direction: column; align-items: center; gap: 12px; }
.dir-admin__spin { width: 28px; height: 28px; animation: dir-rot 1s linear infinite; }
@keyframes dir-rot { to { transform: rotate(360deg); } }

.dir-muted { font-size: 12px; color: var(--muted); }
.dir-mono { font-family: var(--f-num); font-size: 11.5px; letter-spacing: 0.04em; color: var(--muted); }

.dir-person { display: flex; align-items: center; gap: 10px; }
.dir-person__name { margin: 0; font-weight: 600; color: var(--text); }
.dir-person__link { display: flex; align-items: center; gap: 4px; margin: 0; }
.dir-role { margin: 0; font-size: 13px; color: var(--text); }

.dir-field { display: inline-flex; align-items: center; gap: 8px; }
.dir-field__icon { flex: none; width: 18px; height: 18px; color: var(--muted); }
.dir-flags { display: inline-flex; flex-wrap: wrap; gap: 4px; }

.dir-tree { display: grid; gap: 10px; margin: 0; padding: 0; list-style: none; }

.dir-unit {
  border: 1px solid var(--line);
  border-inline-start: 3px solid var(--unit-color);
  border-radius: var(--r-s);
  background: var(--surface);
}

.dir-unit__head { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 10px; padding: 10px 14px; }
.dir-unit__dot { width: 8px; height: 8px; border-radius: 50%; background: var(--unit-color); }
.dir-unit__name { font-weight: 600; color: var(--text); }
.dir-unit__none { margin: 0; padding: 0 14px 10px; font-style: italic; }
.dir-unit__actions { margin-inline-start: auto; display: flex; gap: 2px; }
.dir-head { width: 14px; height: 14px; color: var(--warn); }

.dir-pos { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 16px; padding: 8px 14px; border-block-start: 1px solid var(--line); }
.dir-pos__title { display: flex; align-items: center; gap: 4px; min-width: 12rem; font-size: 13px; font-weight: 600; color: var(--text); }
.dir-pos__holders { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; }

.dir-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 6px 2px 2px;
  padding-inline: 2px 6px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--sunken);
  font-size: 12.5px;
  color: var(--text);
}

.dir-chip__name { background: none; border: 0; color: inherit; cursor: pointer; }
.dir-chip__name:hover { text-decoration: underline; }
.dir-chip__remove { display: inline-flex; padding: 2px; border: 0; border-radius: var(--r-s); background: none; color: var(--muted); cursor: pointer; }
.dir-chip__remove:hover { color: var(--bad); background: var(--surface); }
.dir-chip--add { padding: 3px 9px; border-style: dashed; background: transparent; color: var(--acc-ink); cursor: pointer; }
.dir-chip--add:hover { border-color: var(--acc); }
.dir-chip:focus-visible, .dir-chip__remove:focus-visible, .dir-chip__name:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

.dir-icon-btn {
  display: inline-grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  color: var(--muted);
  font-size: 17px;
  cursor: pointer;
}

.dir-icon-btn:hover { background: var(--sunken); color: var(--text); }
.dir-icon-btn:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }
.dir-icon-btn--sm { width: 24px; height: 24px; font-size: 14px; }
.dir-icon-btn--danger:hover { color: var(--bad); }
.dir-icon-btn--on { color: var(--acc); }

.dir-admin__bar-actions { display: flex; flex-wrap: wrap; gap: 8px; }

.dir-unit__group {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 22ch;
  padding: 1px 8px;
  border: 1px solid color-mix(in srgb, var(--acc) 35%, var(--line));
  border-radius: 999px;
  overflow: hidden;
  font-size: 11.5px;
  font-weight: 600;
  color: var(--text);
  text-overflow: ellipsis;
  white-space: nowrap;
  background: color-mix(in srgb, var(--acc) 8%, transparent);
}

.dir-unit__group svg { flex: none; width: 12px; height: 12px; color: var(--acc); }

@media (prefers-reduced-motion: reduce) {
  .dir-admin__spin { animation: none; }
}
</style>
