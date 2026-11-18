<template>
  <div class="dir-admin">
    <header class="flex flex-wrap items-end justify-between gap-4 mb-6">
      <div>
        <p class="dir-admin__eyebrow">{{ $t('directory.title') }}</p>
        <h1 class="text-3xl font-bold" style="color: var(--text-primary)">{{ $t('directory.manageTitle') }}</h1>
        <p class="mt-1" style="color: var(--text-muted)">
          {{ canAdmin ? $t('directory.manageHelpAdmin') : $t('directory.manageHelpGroupAdmin') }}
        </p>
      </div>
      <router-link to="/directory" class="btn btn-secondary">
        <Icon icon="mdi:eye-outline" class="h-4 w-4 mr-1" /> {{ $t('directory.viewPublic') }}
      </router-link>
    </header>

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

    <div v-if="loading" class="py-16 flex justify-center" style="color: var(--text-muted)">
      <Icon icon="mdi:loading" class="animate-spin h-8 w-8" />
    </div>
    <div v-else-if="error" class="py-16 text-center space-y-3" style="color: var(--text-muted)">
      <p>{{ error }}</p>
      <button type="button" class="btn btn-secondary" @click="load">{{ $t('common.retry') }}</button>
    </div>

    <!-- ====== STRUCTURE ====== -->
    <section v-else-if="activeTab === 'structure'">
      <div class="flex justify-between items-center gap-3 mb-4 flex-wrap">
        <p class="text-sm" style="color: var(--text-muted)">{{ $t('directory.structureHelp') }}</p>
        <button v-if="canAdmin" type="button" class="btn btn-primary" @click="openUnit(null, null)">
          <Icon icon="mdi:plus" class="h-4 w-4 mr-1" /> {{ $t('directory.newUnit') }}
        </button>
      </div>

      <p v-if="!flatUnits.length" class="dir-admin__empty">{{ $t('directory.noUnits') }}</p>

      <ul class="space-y-3">
        <li
          v-for="unit in flatUnits"
          :key="unit.id"
          class="dir-unit"
          :style="{ '--unit-color': unit.color, marginInlineStart: `${Math.min(unit.depth, 6) * 1.5}rem` }"
        >
          <div class="dir-unit__head">
            <span class="dir-unit__dot" />
            <span class="font-semibold" style="color: var(--text-primary)">{{ unit.name }}</span>
            <span v-if="unit.code" class="text-xs font-mono" style="color: var(--text-muted)">{{ unit.code }}</span>
            <div v-if="canAdmin" class="dir-unit__actions">
              <button type="button" class="dir-icon-btn" :title="$t('directory.newPosition')" @click="openPosition(null, unit.id)">
                <Icon icon="mdi:briefcase-plus-outline" />
              </button>
              <button type="button" class="dir-icon-btn" :title="$t('directory.newSubUnit')" @click="openUnit(null, unit.id)">
                <Icon icon="mdi:file-tree-outline" />
              </button>
              <button type="button" class="dir-icon-btn" :title="$t('common.edit')" @click="openUnit(unit)">
                <Icon icon="mdi:pencil-outline" />
              </button>
              <button type="button" class="dir-icon-btn dir-icon-btn--danger" :title="$t('common.delete')" @click="removeUnit(unit)">
                <Icon icon="mdi:trash-can-outline" />
              </button>
            </div>
          </div>

          <p v-if="!(unit.positions || []).length" class="px-4 pb-3 text-xs italic" style="color: var(--text-muted)">
            {{ $t('directory.noPositions') }}
          </p>
          <div v-for="position in sortedPositions(unit)" :key="position.id" class="dir-pos">
            <div class="dir-pos__title">
              <Icon v-if="position.is_head" icon="mdi:star-four-points" class="h-3.5 w-3.5 text-amber-500" :title="$t('directory.head')" />
              <span>{{ position.title }}</span>
              <template v-if="canAdmin">
                <button type="button" class="dir-icon-btn dir-icon-btn--sm" :title="$t('common.edit')" @click="openPosition(position)">
                  <Icon icon="mdi:pencil-outline" />
                </button>
                <button type="button" class="dir-icon-btn dir-icon-btn--sm dir-icon-btn--danger" :title="$t('common.delete')" @click="removePosition(position)">
                  <Icon icon="mdi:trash-can-outline" />
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
                  <Icon icon="mdi:close" class="h-3.5 w-3.5" />
                </button>
              </span>
              <span v-if="!(position.assignments || []).length" class="text-xs italic" style="color: var(--text-muted)">{{ $t('directory.vacant') }}</span>
              <button type="button" class="dir-chip dir-chip--add" @click="openAssignment(position)">
                <Icon icon="mdi:account-plus-outline" class="h-4 w-4" /> {{ $t('directory.assign') }}
              </button>
            </div>
          </div>
        </li>
      </ul>
    </section>

    <!-- ====== PERSONNES ====== -->
    <section v-else-if="activeTab === 'people'">
      <div class="flex justify-between items-center gap-3 mb-4 flex-wrap">
        <label class="dir-admin__search">
          <Icon icon="mdi:magnify" class="h-4 w-4" />
          <input v-model="peopleSearch" type="search" :placeholder="$t('directory.searchPlaceholder')" />
        </label>
        <button type="button" class="btn btn-primary" @click="openPerson(null)">
          <Icon icon="mdi:account-plus" class="h-4 w-4 mr-1" /> {{ $t('directory.newPerson') }}
        </button>
      </div>

      <p v-if="!people.length" class="dir-admin__empty">{{ $t('directory.noPeople') }}</p>
      <div v-else class="card p-0 overflow-x-auto">
        <table class="table w-full">
          <thead>
            <tr>
              <th>{{ $t('directory.person') }}</th>
              <th>{{ $t('directory.positions') }}</th>
              <th>{{ $t('directory.contact') }}</th>
              <th>{{ $t('directory.status') }}</th>
              <th class="text-end">{{ $t('directory.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in filteredPeople" :key="p.id">
              <td>
                <div class="flex items-center gap-3">
                  <PersonAvatar :person="p" size="sm" />
                  <div>
                    <p class="font-medium">{{ fullName(p) }}</p>
                    <p v-if="p.user_id" class="text-xs flex items-center gap-1" style="color: var(--text-muted)">
                      <Icon icon="mdi:link-variant" class="h-3 w-3" /> {{ $t('directory.linked') }}
                    </p>
                  </div>
                </div>
              </td>
              <td class="text-sm">
                <p v-for="r in rolesByPerson.get(p.id) || []" :key="r.assignment.id">
                  {{ r.position.title }} <span style="color: var(--text-muted)">· {{ r.unit.name }}</span>
                </p>
                <span v-if="!(rolesByPerson.get(p.id) || []).length" class="italic" style="color: var(--text-muted)">—</span>
              </td>
              <td class="text-sm">
                <p>{{ p.email }}</p>
                <p style="color: var(--text-muted)">{{ p.phone }}</p>
              </td>
              <td>
                <span class="badge" :class="p.is_active ? 'badge-success' : 'badge-secondary'">
                  {{ p.is_active ? $t('common.active') : $t('common.inactive') }}
                </span>
              </td>
              <td class="text-end whitespace-nowrap">
                <button type="button" class="dir-icon-btn" :title="$t('common.edit')" @click="openPerson(p)">
                  <Icon icon="mdi:pencil-outline" />
                </button>
                <button type="button" class="dir-icon-btn dir-icon-btn--danger" :title="$t('common.delete')" @click="removePerson(p)">
                  <Icon icon="mdi:trash-can-outline" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- ====== CHAMPS DYNAMIQUES (admin) ====== -->
    <section v-else-if="activeTab === 'fields' && canAdmin">
      <div class="flex justify-between items-center gap-3 mb-4 flex-wrap">
        <p class="text-sm max-w-2xl" style="color: var(--text-muted)">{{ $t('directory.fieldsHelp') }}</p>
        <button type="button" class="btn btn-primary" @click="openField(null)">
          <Icon icon="mdi:plus" class="h-4 w-4 mr-1" /> {{ $t('directory.newField') }}
        </button>
      </div>
      <p v-if="!fields.length" class="dir-admin__empty">{{ $t('directory.noCustomFields') }}</p>
      <ul class="grid gap-3 md:grid-cols-2">
        <li v-for="f in fields" :key="f.id" class="dir-field">
          <span class="dir-field__icon"><Icon :icon="FIELD_TYPE_ICONS[f.type]" class="h-5 w-5" /></span>
          <div class="flex-1 min-w-0">
            <p class="font-semibold truncate" style="color: var(--text-primary)">{{ f.label }}</p>
            <p class="text-xs font-mono" style="color: var(--text-muted)">{{ f.key }} · {{ $t(`directory.fieldTypes.${f.type}`) }}</p>
            <div class="flex flex-wrap gap-1 mt-2">
              <span v-if="f.required" class="badge badge-secondary">{{ $t('directory.fieldRequired') }}</span>
              <span v-if="f.visibility === 'managers'" class="badge bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-200">
                <Icon icon="mdi:lock-outline" class="h-3 w-3 mr-1" />{{ $t('directory.visibility.managers') }}
              </span>
              <span v-if="f.show_on_card" class="badge badge-secondary">{{ $t('directory.onCard') }}</span>
            </div>
          </div>
          <div class="flex flex-col gap-1">
            <button type="button" class="dir-icon-btn" :title="$t('common.edit')" @click="openField(f)"><Icon icon="mdi:pencil-outline" /></button>
            <button type="button" class="dir-icon-btn dir-icon-btn--danger" :title="$t('common.delete')" @click="removeField(f)"><Icon icon="mdi:trash-can-outline" /></button>
          </div>
        </li>
      </ul>
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
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import PersonAvatar from '@/components/directory/PersonAvatar.vue'
import UnitFormModal from '@/components/directory/UnitFormModal.vue'
import PositionFormModal from '@/components/directory/PositionFormModal.vue'
import AssignmentFormModal from '@/components/directory/AssignmentFormModal.vue'
import PersonFormModal from '@/components/directory/PersonFormModal.vue'
import FieldFormModal from '@/components/directory/FieldFormModal.vue'
import { directoryService } from '@/services/api'
import { useDirectory, fullName, notify, apiError, FIELD_TYPE_ICONS } from '@/components/directory/useDirectory'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const {
  loading, error, units, people, fields, canAdmin, load,
  peopleById, flatUnits, rolesByPerson, descendantIds,
} = useDirectory()

const tabs = [
  { id: 'structure', icon: 'mdi:sitemap-outline', label: 'directory.tabStructure', count: () => units.value.length },
  { id: 'people', icon: 'mdi:account-multiple-outline', label: 'directory.tabPeople', count: () => people.value.length },
  { id: 'fields', icon: 'mdi:form-select', label: 'directory.tabFields', count: () => fields.value.length, adminOnly: true },
]
const visibleTabs = computed(() => tabs.filter((tab) => !tab.adminOnly || canAdmin.value))

const activeTab = ref(['structure', 'people', 'fields'].includes(route.query.tab) ? route.query.tab : 'structure')
const setTab = (id) => {
  activeTab.value = id
  router.replace({ query: { ...route.query, tab: id, person: undefined } })
}

const peopleSearch = ref('')
const normalize = (s) => (s || '').normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase()
const filteredPeople = computed(() => {
  const q = normalize(peopleSearch.value)
  if (!q) return people.value
  return people.value.filter((p) => normalize(`${fullName(p)} ${p.email} ${p.phone}`).includes(q))
})

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

const onSaved = async (messageKey) => {
  closeModal()
  notify('success', t(messageKey))
  await load()
}

// ---- Suppressions ----
const confirmAndRun = async (message, action, successKey) => {
  if (!window.confirm(message)) return
  try {
    await action()
    notify('success', t(successKey))
    await load()
  } catch (e) {
    notify('error', apiError(e, t('directory.errors.deleteFailed')))
  }
}

const removeUnit = (unit) =>
  confirmAndRun(t('directory.confirm.unit', { name: unit.name }), () => directoryService.deleteUnit(unit.id), 'directory.deleted.unit')
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
  // Lien profond depuis la fiche publique : ?person=ID ouvre l'édition
  const personId = Number(route.query.person)
  if (personId && peopleById.value.has(personId)) {
    activeTab.value = 'people'
    openPerson(peopleById.value.get(personId))
  }
})
</script>

<style scoped>
.dir-admin { padding: clamp(1rem, 2.5vw, 2rem); max-width: 80rem; margin: 0 auto; }
.dir-admin__eyebrow { font-size: 0.72rem; font-weight: 700; letter-spacing: 0.14em; text-transform: uppercase; color: var(--accent); }
.dir-admin__tabs { display: flex; gap: 0.25rem; margin-bottom: 1.5rem; border-bottom: 1px solid var(--border); overflow-x: auto; }
.dir-admin__tab {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.7rem 1rem;
  margin-bottom: -1px;
  border-bottom: 2px solid transparent;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-muted);
  white-space: nowrap;
}
.dir-admin__tab:hover { color: var(--text-primary); }
.dir-admin__tab--active { color: var(--accent); border-bottom-color: var(--accent); }
.dir-admin__tab-count { font-size: 0.7rem; padding: 0.05rem 0.45rem; border-radius: 9999px; background: var(--bg-surface); }
.dir-admin__empty { padding: 3rem 1rem; text-align: center; color: var(--text-muted); border: 1px dashed var(--border); border-radius: 0.85rem; }
.dir-admin__search {
  display: flex; align-items: center; gap: 0.5rem; padding: 0 0.75rem; height: 2.5rem; min-width: min(100%, 20rem);
  border: 1px solid var(--border); border-radius: 0.65rem; background: var(--bg-card); color: var(--text-muted);
}
.dir-admin__search:focus-within { border-color: var(--accent); }
.dir-admin__search input { flex: 1; background: transparent; border: 0; outline: none; color: var(--text-primary); }

.dir-unit { border: 1px solid var(--border); border-inline-start: 4px solid var(--unit-color); border-radius: 0.85rem; background: var(--bg-card); }
.dir-unit__head { display: flex; align-items: center; gap: 0.6rem; padding: 0.75rem 1rem; }
.dir-unit__dot { width: 0.6rem; height: 0.6rem; border-radius: 9999px; background: var(--unit-color); }
.dir-unit__actions { margin-inline-start: auto; display: flex; gap: 0.15rem; }
.dir-pos { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem 1rem; padding: 0.55rem 1rem; border-top: 1px solid var(--border); }
.dir-pos__title { display: flex; align-items: center; gap: 0.3rem; min-width: 12rem; font-size: 0.85rem; font-weight: 600; color: var(--text-primary); }
.dir-pos__holders { display: flex; flex-wrap: wrap; align-items: center; gap: 0.4rem; }
.dir-chip {
  display: inline-flex; align-items: center; gap: 0.35rem; padding: 0.15rem 0.35rem 0.15rem 0.15rem;
  border-radius: 9999px; border: 1px solid var(--border); background: var(--bg-surface); font-size: 0.8rem; color: var(--text-primary);
}
.dir-chip__name:hover { text-decoration: underline; }
.dir-chip__remove { display: inline-flex; padding: 0.15rem; border-radius: 9999px; color: var(--text-muted); }
.dir-chip__remove:hover { color: var(--danger, #dc2626); background: var(--hover); }
.dir-chip--add { padding: 0.25rem 0.65rem; border-style: dashed; color: var(--accent); background: transparent; }
.dir-chip--add:hover { border-color: var(--accent); background: color-mix(in srgb, var(--accent) 8%, transparent); }
.dir-icon-btn { display: inline-flex; padding: 0.4rem; border-radius: 0.45rem; color: var(--text-muted); font-size: 1.05rem; transition: background-color 150ms ease, color 150ms ease; }
.dir-icon-btn:hover { background: var(--hover); color: var(--text-primary); }
.dir-icon-btn:focus-visible { outline: 2px solid var(--accent); }
.dir-icon-btn--sm { padding: 0.2rem; font-size: 0.9rem; }
.dir-icon-btn--danger:hover { color: var(--danger, #dc2626); }
.dir-field { display: flex; gap: 0.85rem; padding: 1rem; border: 1px solid var(--border); border-radius: 0.85rem; background: var(--bg-card); }
.dir-field__icon {
  display: inline-flex; align-items: center; justify-content: center; width: 2.5rem; height: 2.5rem; flex-shrink: 0;
  border-radius: 0.65rem; color: var(--accent); background: color-mix(in srgb, var(--accent) 12%, transparent);
}
</style>
