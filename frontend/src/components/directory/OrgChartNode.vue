<template>
  <li class="org-node" :style="{ '--unit-color': unit.color || 'var(--accent)' }">
    <div class="org-node__card">
      <button
        type="button"
        class="org-node__header"
        :aria-expanded="hasContent ? String(open) : undefined"
        @click="hasContent && (open = !open)"
      >
        <Icon
          v-if="hasContent"
          icon="mdi:chevron-right"
          class="org-node__chevron h-5 w-5"
          :class="{ 'org-node__chevron--open': open }"
        />
        <span v-else class="w-5" />
        <span class="org-node__name">{{ unit.name }}</span>
        <span v-if="unit.code" class="org-node__code">{{ unit.code }}</span>
        <span class="org-node__count" :title="$t('directory.peopleCount', { count: peopleCount })">
          <Icon icon="mdi:account-multiple-outline" class="h-3.5 w-3.5" /> {{ peopleCount }}
        </span>
      </button>

      <div v-show="open" v-if="positions.length" class="org-node__positions">
        <div v-for="position in positions" :key="position.id" class="org-pos" :class="{ 'org-pos--head': position.is_head }">
          <p class="org-pos__title">
            <Icon v-if="position.is_head" icon="mdi:star-four-points" class="h-3.5 w-3.5 text-amber-500" />
            {{ position.title }}
          </p>
          <div class="org-pos__holders">
            <template v-if="holders(position).length">
              <button
                v-for="h in holders(position)"
                :key="h.assignment.id"
                type="button"
                class="org-holder"
                @click="$emit('open-person', h.person)"
              >
                <PersonAvatar :person="h.person" size="sm" />
                <span class="org-holder__name">{{ fullName(h.person) }}</span>
                <span v-if="h.assignment.is_interim" class="dir-interim">{{ $t('directory.interim') }}</span>
              </button>
            </template>
            <span v-else class="org-holder org-holder--vacant">
              <PersonAvatar :person="null" size="sm" />
              <span class="org-holder__name">{{ $t('directory.vacant') }}</span>
            </span>
          </div>
        </div>
      </div>
    </div>

    <ul v-if="children.length" v-show="open" class="org-node__children">
      <OrgChartNode
        v-for="child in children"
        :key="child.id"
        :unit="child"
        :children-of="childrenOf"
        :people-by-id="peopleById"
        :count-people="countPeople"
        :depth="depth + 1"
        :force-open="forceOpen"
        @open-person="$emit('open-person', $event)"
      />
    </ul>
  </li>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { Icon } from '@iconify/vue'
import PersonAvatar from './PersonAvatar.vue'
import { fullName } from './useDirectory'

defineOptions({ name: 'OrgChartNode' })

const props = defineProps({
  unit: { type: Object, required: true },
  childrenOf: { type: Map, required: true },
  peopleById: { type: Map, required: true },
  countPeople: { type: Function, required: true },
  depth: { type: Number, default: 0 },
  forceOpen: { type: Boolean, default: null },
})
defineEmits(['open-person'])

const open = ref(props.depth < 2)
watch(() => props.forceOpen, (v) => { if (v !== null) open.value = v })

const children = computed(() => props.childrenOf.get(props.unit.id) || [])
const positions = computed(() =>
  [...(props.unit.positions || [])].sort((a, b) => Number(b.is_head) - Number(a.is_head) || a.order - b.order)
)
const hasContent = computed(() => children.value.length > 0 || positions.value.length > 0)
const peopleCount = computed(() => props.countPeople(props.unit.id))

const holders = (position) =>
  (position.assignments || [])
    .map((assignment) => ({ assignment, person: props.peopleById.get(assignment.person_id) }))
    .filter((h) => h.person)
</script>

<style scoped>
.org-node { position: relative; list-style: none; }
.org-node__card {
  border: 1px solid var(--border);
  border-inline-start: 4px solid var(--unit-color);
  border-radius: 0.85rem;
  background: var(--bg-card);
  overflow: hidden;
  box-shadow: 0 1px 0 rgba(var(--shadow-rgb, 15, 23, 42), 0.04);
}
.org-node__header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.75rem 1rem;
  text-align: start;
  color: var(--text-primary);
  background: linear-gradient(90deg, color-mix(in srgb, var(--unit-color) 9%, transparent), transparent 70%);
}
.org-node__header:hover { background-color: var(--hover); }
.org-node__header:focus-visible { outline: 2px solid var(--accent); outline-offset: -2px; }
.org-node__chevron { color: var(--text-muted); transition: transform 200ms cubic-bezier(0.16, 1, 0.3, 1); }
.org-node__chevron--open { transform: rotate(90deg); }
[dir='rtl'] .org-node__chevron:not(.org-node__chevron--open) { transform: rotate(180deg); }
.org-node__name { font-weight: 700; font-size: 1rem; }
.org-node__code {
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  padding: 0.1rem 0.45rem;
  border-radius: 0.35rem;
  color: var(--unit-color);
  background: color-mix(in srgb, var(--unit-color) 14%, transparent);
}
.org-node__count {
  margin-inline-start: auto;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.75rem;
  color: var(--text-muted);
}
.org-node__positions {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(15rem, 1fr));
  gap: 0.5rem;
  padding: 0.25rem 1rem 1rem;
}
.org-pos {
  padding: 0.6rem 0.75rem;
  border-radius: 0.6rem;
  background: var(--bg-surface);
  border: 1px solid transparent;
}
.org-pos--head {
  grid-column: 1 / -1;
  border-color: color-mix(in srgb, var(--unit-color) 35%, transparent);
  background: color-mix(in srgb, var(--unit-color) 6%, var(--bg-surface));
}
.org-pos__title {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 0.4rem;
}
.org-pos__holders { display: flex; flex-wrap: wrap; gap: 0.4rem; }
.org-holder {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.2rem 0.65rem 0.2rem 0.2rem;
  border-radius: 9999px;
  color: var(--text-primary);
  transition: background-color 150ms ease;
}
[dir='rtl'] .org-holder { padding: 0.2rem 0.2rem 0.2rem 0.65rem; }
button.org-holder:hover { background: var(--hover); }
button.org-holder:focus-visible { outline: 2px solid var(--accent); }
.org-holder__name { font-size: 0.875rem; font-weight: 500; }
.org-holder--vacant { color: var(--text-muted); font-style: italic; }

.org-node__children {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 0.75rem;
  padding-inline-start: 2rem;
}
/* Connecteurs de l'arbre */
.org-node__children::before {
  content: '';
  position: absolute;
  inset-inline-start: 0.9rem;
  top: -0.75rem;
  bottom: 1.4rem;
  border-inline-start: 2px solid var(--border);
}
.org-node__children > .org-node::before {
  content: '';
  position: absolute;
  inset-inline-start: -1.1rem;
  top: 1.4rem;
  width: 1.1rem;
  border-top: 2px solid var(--border);
}
@media (max-width: 640px) {
  .org-node__children { padding-inline-start: 1rem; }
  .org-node__children::before { inset-inline-start: 0.35rem; }
  .org-node__children > .org-node::before { inset-inline-start: -0.65rem; width: 0.65rem; }
}
</style>
