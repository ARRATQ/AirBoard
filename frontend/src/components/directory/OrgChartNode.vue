<template>
  <li class="org-node" :style="{ '--unit-color': unit.color || 'var(--acc)' }">
    <div class="org-node__card">
      <button
        type="button"
        class="org-node__header"
        :aria-expanded="hasContent ? String(open) : undefined"
        @click="hasContent && (open = !open)"
      >
        <Icon
          v-if="hasContent"
          icon="lucide:chevron-right"
          class="org-node__chevron h-5 w-5"
          :class="{ 'org-node__chevron--open': open }"
        />
        <span v-else class="w-5" />
        <span class="org-node__name">{{ unit.name }}</span>
        <span v-if="unit.code" class="tag org-node__code">{{ unit.code }}</span>
        <span class="org-node__count" :title="$t('directory.peopleCount', { count: peopleCount })">
          <Icon icon="lucide:users" class="h-3.5 w-3.5" /> {{ peopleCount }}
        </span>
      </button>

      <div v-show="open" v-if="positions.length" class="org-node__positions">
        <div v-for="position in positions" :key="position.id" class="org-pos" :class="{ 'org-pos--head': position.is_head }">
          <p class="org-pos__title">
            <Icon v-if="position.is_head" icon="lucide:sparkle" class="org-pos__star h-3.5 w-3.5" aria-hidden="true" />
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
  border: 1px solid var(--line);
  border-inline-start: 3px solid var(--unit-color);
  border-radius: var(--r-s);
  background: var(--surface);
  overflow: hidden;
}
.org-node__header {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 10px 14px;
  border: 0;
  background: none;
  font: inherit;
  text-align: start;
  color: var(--text);
  cursor: pointer;
}
.org-node__header:hover { background: var(--sunken); }
.org-node__header:focus-visible { outline: 2px solid var(--acc); outline-offset: -2px; }
.org-node__chevron { color: var(--muted); transition: transform 150ms ease; }
.org-node__chevron--open { transform: rotate(90deg); }
[dir='rtl'] .org-node__chevron:not(.org-node__chevron--open) { transform: rotate(180deg); }
.org-node__name { font-weight: 600; font-size: 15px; }
.org-node__count {
  margin-inline-start: auto;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-family: var(--f-num);
  font-size: 11.5px;
  color: var(--muted);
}
.org-node__positions {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(15rem, 1fr));
  gap: 8px;
  padding: 2px 14px 14px;
}
.org-pos {
  padding: 9px 11px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--sunken);
}
.org-pos--head {
  grid-column: 1 / -1;
  border-inline-start: 2px solid var(--unit-color);
  background: var(--surface);
}
.org-pos__title {
  display: flex;
  align-items: center;
  gap: 5px;
  margin: 0 0 6px;
  font-family: var(--f-num);
  font-size: 10.5px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--muted);
}
.org-pos__star { color: var(--warn); }
.org-pos__holders { display: flex; flex-wrap: wrap; gap: 6px; }
.org-holder {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding-block: 3px;
  padding-inline: 3px 10px;
  border: 0;
  border-radius: 9999px;
  background: none;
  font: inherit;
  color: var(--text);
  transition: background-color 150ms ease;
}
button.org-holder { cursor: pointer; }
button.org-holder:hover { background: var(--line); }
button.org-holder:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }
.org-holder__name { font-size: 13.5px; font-weight: 500; }
.org-holder--vacant { color: var(--muted); font-style: italic; }

.org-node__children {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 12px;
  padding-inline-start: 32px;
}
/* Connecteurs de l'arbre */
.org-node__children::before {
  content: '';
  position: absolute;
  inset-inline-start: 14px;
  top: -12px;
  bottom: 22px;
  border-inline-start: 1px solid var(--line);
}
.org-node__children > .org-node::before {
  content: '';
  position: absolute;
  inset-inline-start: -18px;
  top: 22px;
  width: 18px;
  border-top: 1px solid var(--line);
}
@media (max-width: 640px) {
  .org-node__children { padding-inline-start: 16px; }
  .org-node__children::before { inset-inline-start: 6px; }
  .org-node__children > .org-node::before { inset-inline-start: -10px; width: 10px; }
}

@media (prefers-reduced-motion: reduce) {
  .org-node__chevron, .org-holder { transition: none; }
}
</style>
