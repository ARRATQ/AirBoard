<template>
  <li class="ot-node">
    <article
      class="ot-card"
      :class="{ 'ot-card--root': depth === 0 }"
      :style="{ '--unit-color': unit.color || 'var(--acc)' }"
      :dir="textDir"
      :data-unit-id="unit.id"
    >
      <header class="ot-card__head">
        <button type="button" class="ot-card__name" :title="$t('directory.focusUnit')" @click="$emit('focus', unit.id)">
          {{ unit.name }}
        </button>
        <span v-if="unit.code" class="tag ot-card__code">{{ unit.code }}</span>
      </header>

      <!-- Responsable mis en avant -->
      <div v-if="headPosition" class="ot-card__head-person">
        <template v-if="headHolders.length">
          <button
            v-for="h in headHolders.slice(0, 2)"
            :key="h.assignment.id"
            type="button"
            class="ot-lead"
            @click="$emit('open-person', h.person)"
          >
            <PersonAvatar :person="h.person" :size="headHolders.length > 1 ? 'md' : 'lg'" />
            <span class="ot-lead__name">{{ fullName(h.person) }}</span>
            <span class="ot-lead__title">
              {{ headPosition.title }}
              <span v-if="h.assignment.is_interim" class="dir-interim">{{ $t('directory.interim') }}</span>
            </span>
          </button>
        </template>
        <div v-else class="ot-lead ot-lead--vacant">
          <PersonAvatar :person="null" size="lg" />
          <span class="ot-lead__name">{{ $t('directory.vacant') }}</span>
          <span class="ot-lead__title">{{ headPosition.title }}</span>
        </div>
      </div>

      <!-- Autres membres -->
      <footer class="ot-card__team">
        <div v-if="members.length" class="ot-stack">
          <button
            v-for="m in members.slice(0, MAX_AVATARS)"
            :key="m.assignment.id"
            type="button"
            class="ot-stack__item"
            :title="`${fullName(m.person)} — ${m.position.title}`"
            @click="$emit('open-person', m.person)"
          >
            <PersonAvatar :person="m.person" size="sm" />
          </button>
          <span v-if="members.length > MAX_AVATARS" class="ot-stack__more">+{{ members.length - MAX_AVATARS }}</span>
        </div>
        <span class="ot-card__count">
          <Icon icon="mdi:account-multiple-outline" class="h-3.5 w-3.5" />
          {{ peopleCount }}
          <template v-if="vacantCount">
            · <Icon icon="mdi:account-question-outline" class="h-3.5 w-3.5" /> {{ vacantCount }}
          </template>
        </span>
      </footer>
    </article>

    <button
      v-if="children.length"
      type="button"
      class="ot-toggle"
      :class="{ 'ot-toggle--closed': !open }"
      :aria-expanded="String(open)"
      :title="open ? $t('directory.collapse') : $t('directory.expand')"
      @click="open = !open"
    >
      <Icon :icon="open ? 'mdi:minus' : 'mdi:plus'" class="h-3 w-3" />
      <span v-if="!open">{{ children.length }}</span>
    </button>

    <ul v-if="children.length && open" class="ot-children">
      <OrgTreeNode
        v-for="child in children"
        :key="child.id"
        :unit="child"
        :children-of="childrenOf"
        :people-by-id="peopleById"
        :count-people="countPeople"
        :depth="depth + 1"
        :force-open="forceOpen"
        :text-dir="textDir"
        @open-person="$emit('open-person', $event)"
        @focus="$emit('focus', $event)"
      />
    </ul>
  </li>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { Icon } from '@iconify/vue'
import PersonAvatar from './PersonAvatar.vue'
import { fullName } from './useDirectory'

defineOptions({ name: 'OrgTreeNode' })

const MAX_AVATARS = 5

const props = defineProps({
  unit: { type: Object, required: true },
  childrenOf: { type: Map, required: true },
  peopleById: { type: Map, required: true },
  countPeople: { type: Function, required: true },
  depth: { type: Number, default: 0 },
  forceOpen: { type: Boolean, default: null },
  textDir: { type: String, default: 'ltr' },
})
defineEmits(['open-person', 'focus'])

const open = ref(props.depth < 2)
watch(() => props.forceOpen, (v) => { if (v !== null) open.value = v })

const children = computed(() => props.childrenOf.get(props.unit.id) || [])
const positions = computed(() =>
  [...(props.unit.positions || [])].sort((a, b) => Number(b.is_head) - Number(a.is_head) || a.order - b.order)
)
const headPosition = computed(() => positions.value.find((p) => p.is_head) || null)

const holdersOf = (position) =>
  (position.assignments || [])
    .map((assignment) => ({ assignment, position, person: props.peopleById.get(assignment.person_id) }))
    .filter((h) => h.person)

const headHolders = computed(() => (headPosition.value ? holdersOf(headPosition.value) : []))
const members = computed(() =>
  positions.value.filter((p) => p !== headPosition.value).flatMap(holdersOf)
)
const vacantCount = computed(() => positions.value.filter((p) => !holdersOf(p).length).length)
const peopleCount = computed(() => props.countPeople(props.unit.id))
</script>

<style scoped>
/* ---- Connecteurs en équerre (arbre de haut en bas), en propriétés logiques pour l'arabe ---- */
.ot-node {
  --ot-wire: var(--line);
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 28px 10px 0;
  list-style: none;
}
.ot-node::before,
.ot-node::after {
  content: '';
  position: absolute;
  top: 0;
  width: 50%;
  height: 28px;
  border-top: 1px solid var(--ot-wire);
}
.ot-node::before { inset-inline-end: 50%; }
.ot-node::after { inset-inline-start: 50%; border-inline-start: 1px solid var(--ot-wire); }
.ot-node:only-child::before,
.ot-node:only-child::after { display: none; }
.ot-node:only-child { padding-top: 0; }
.ot-node:first-child::before,
.ot-node:last-child::after { border: 0 none; }
.ot-node:last-child::before { border-inline-end: 1px solid var(--ot-wire); border-start-end-radius: var(--r-s); }
.ot-node:first-child::after { border-start-start-radius: var(--r-s); }
.ot-children {
  position: relative;
  display: flex;
  justify-content: center;
  padding-top: 28px;
}
.ot-children::before {
  content: '';
  position: absolute;
  top: 0;
  inset-inline-start: 50%;
  height: 28px;
  border-inline-start: 1px solid var(--ot-wire);
}
/* Racine sans trait au-dessus */
.ot-root > .ot-node::before,
.ot-root > .ot-node::after { display: none; }
.ot-root > .ot-node { padding-top: 0; }

/* ---- Carte d'unité : filet de la couleur de l'unité, pas d'ombre ---- */
.ot-card {
  position: relative;
  width: 15.5rem;
  border: 1px solid var(--line);
  border-block-start: 3px solid var(--unit-color);
  border-radius: var(--r-s);
  background: var(--surface);
  overflow: hidden;
  transition: border-color 150ms ease;
}
.ot-card:hover { border-inline-color: var(--muted); border-block-end-color: var(--muted); }
.ot-card--root { width: 18rem; }
.ot-card__head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px 6px;
}
.ot-card__name {
  flex: 1;
  min-width: 0;
  padding: 0;
  border: 0;
  background: none;
  text-align: start;
  font: inherit;
  font-weight: 600;
  font-size: 14px;
  line-height: 1.25;
  color: var(--text);
  cursor: pointer;
}
.ot-card--root .ot-card__name { font-size: 15.5px; }
.ot-card__name:hover { text-decoration: underline; text-underline-offset: 3px; }
.ot-card__name:focus-visible { outline: 2px solid var(--acc); outline-offset: 2px; border-radius: 2px; }
.ot-card__head-person {
  display: flex;
  justify-content: center;
  gap: 8px;
  padding: 2px 10px 10px;
}
.ot-lead {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  min-width: 0;
  padding: 6px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  font: inherit;
  text-align: center;
  color: var(--text);
}
button.ot-lead { cursor: pointer; }
button.ot-lead:hover { background: var(--sunken); }
button.ot-lead:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }
.ot-lead__name { margin-top: 3px; font-weight: 600; font-size: 13px; line-height: 1.2; }
.ot-lead__title {
  display: inline-flex;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: var(--muted);
}
.ot-lead--vacant .ot-lead__name { color: var(--muted); font-style: italic; font-weight: 500; }
.ot-card__team {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 7px 12px;
  border-top: 1px solid var(--line);
  background: var(--sunken);
}
.ot-stack { display: flex; align-items: center; padding-inline-start: 6px; }
.ot-stack__item { margin-inline-start: -7px; border-radius: 9999px; box-shadow: 0 0 0 2px var(--sunken); }
.ot-stack__item:hover { z-index: 1; }
.ot-stack__item:focus-visible { z-index: 1; outline: 2px solid var(--acc); outline-offset: 1px; }
.ot-stack__more {
  margin-inline-start: 5px;
  font-family: var(--f-num);
  font-size: 11px;
  color: var(--muted);
}
.ot-card__count {
  margin-inline-start: auto;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-family: var(--f-num);
  font-size: 11.5px;
  color: var(--muted);
  white-space: nowrap;
}

/* ---- Bouton déplier/replier ---- */
.ot-toggle {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  margin-top: -11px;
  padding: 2px 7px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
  font-family: var(--f-num);
  font-size: 11px;
  color: var(--muted);
  cursor: pointer;
  transition: color 150ms ease, border-color 150ms ease;
}
.ot-toggle:hover { color: var(--text); border-color: var(--muted); }
.ot-toggle:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }
.ot-toggle--closed { color: var(--text); }
.ot-node > .ot-toggle + .ot-children { margin-top: -11px; }

@media (prefers-reduced-motion: reduce) {
  .ot-card, .ot-toggle { transition: none; }
}
</style>
