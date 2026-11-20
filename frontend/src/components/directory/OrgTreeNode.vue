<template>
  <li class="ot-node">
    <article
      class="ot-card"
      :class="{ 'ot-card--root': depth === 0 }"
      :style="{ '--unit-color': unit.color || 'var(--accent)' }"
      :dir="textDir"
      :data-unit-id="unit.id"
    >
      <header class="ot-card__head">
        <button type="button" class="ot-card__name" :title="$t('directory.focusUnit')" @click="$emit('focus', unit.id)">
          {{ unit.name }}
        </button>
        <span v-if="unit.code" class="ot-card__code">{{ unit.code }}</span>
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
/* ---- Connecteurs en équerre (arbre de haut en bas) ---- */
.ot-node {
  --line: color-mix(in srgb, var(--text-muted) 45%, transparent);
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.75rem 0.65rem 0;
  list-style: none;
}
.ot-node::before,
.ot-node::after {
  content: '';
  position: absolute;
  top: 0;
  width: 50%;
  height: 1.75rem;
  border-top: 2px solid var(--line);
}
.ot-node::before { right: 50%; }
.ot-node::after { left: 50%; border-left: 2px solid var(--line); }
.ot-node:only-child::before,
.ot-node:only-child::after { display: none; }
.ot-node:only-child { padding-top: 0; }
.ot-node:first-child::before,
.ot-node:last-child::after { border: 0 none; }
.ot-node:last-child::before { border-right: 2px solid var(--line); border-top-right-radius: 0.75rem; }
.ot-node:first-child::after { border-top-left-radius: 0.75rem; }
.ot-children {
  position: relative;
  display: flex;
  justify-content: center;
  padding-top: 1.75rem;
}
.ot-children::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  height: 1.75rem;
  border-left: 2px solid var(--line);
}
/* Racine sans trait au-dessus */
.ot-root > .ot-node::before,
.ot-root > .ot-node::after { display: none; }
.ot-root > .ot-node { padding-top: 0; }

/* ---- Carte d'unité ---- */
.ot-card {
  position: relative;
  width: 15.5rem;
  border: 1px solid var(--border);
  border-radius: 1rem;
  background: var(--bg-card);
  overflow: hidden;
  box-shadow:
    0 1px 0 rgba(var(--shadow-rgb, 15, 23, 42), 0.04),
    0 14px 30px -18px rgba(var(--shadow-rgb, 15, 23, 42), 0.45);
  transition: transform 200ms cubic-bezier(0.16, 1, 0.3, 1), box-shadow 200ms ease, border-color 200ms ease;
}
.ot-card::before {
  content: '';
  position: absolute;
  inset: 0 0 auto 0;
  height: 4px;
  background: var(--unit-color);
}
.ot-card:hover {
  transform: translateY(-2px);
  border-color: color-mix(in srgb, var(--unit-color) 55%, var(--border));
}
.ot-card--root { width: 18rem; }
.ot-card__head {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.85rem 0.9rem 0.6rem;
  background: linear-gradient(180deg, color-mix(in srgb, var(--unit-color) 10%, transparent), transparent);
}
.ot-card__name {
  flex: 1;
  min-width: 0;
  text-align: start;
  font-weight: 700;
  font-size: 0.92rem;
  line-height: 1.25;
  color: var(--text-primary);
}
.ot-card--root .ot-card__name { font-size: 1.05rem; }
.ot-card__name:hover { color: var(--unit-color); text-decoration: underline; text-underline-offset: 3px; }
.ot-card__code {
  flex-shrink: 0;
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  padding: 0.12rem 0.4rem;
  border-radius: 0.35rem;
  color: var(--unit-color);
  background: color-mix(in srgb, var(--unit-color) 14%, transparent);
}
.ot-card__head-person {
  display: flex;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.25rem 0.75rem 0.75rem;
}
.ot-lead {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.2rem;
  min-width: 0;
  padding: 0.35rem;
  border-radius: 0.75rem;
  text-align: center;
  color: var(--text-primary);
}
button.ot-lead:hover { background: var(--hover); }
button.ot-lead:focus-visible { outline: 2px solid var(--accent); }
.ot-lead__name { margin-top: 0.25rem; font-weight: 600; font-size: 0.85rem; line-height: 1.2; }
.ot-lead__title {
  display: inline-flex;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.72rem;
  color: var(--unit-color);
  font-weight: 500;
}
.ot-lead--vacant .ot-lead__name { color: var(--text-muted); font-style: italic; }
.ot-card__team {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.55rem 0.8rem;
  border-top: 1px solid var(--border);
  background: var(--bg-surface);
}
.ot-stack { display: flex; align-items: center; padding-inline-start: 0.35rem; }
.ot-stack__item { margin-inline-start: -0.45rem; border-radius: 9999px; transition: transform 150ms ease; }
.ot-stack__item:hover { transform: translateY(-2px); z-index: 1; }
.ot-stack__item:focus-visible { outline: 2px solid var(--accent); }
.ot-stack__more {
  margin-inline-start: 0.3rem;
  font-size: 0.7rem;
  font-weight: 700;
  color: var(--text-muted);
}
.ot-card__count {
  margin-inline-start: auto;
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  white-space: nowrap;
}

/* ---- Bouton déplier/replier ---- */
.ot-toggle {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  margin-top: -0.7rem;
  padding: 0.2rem 0.45rem;
  border: 1px solid var(--border);
  border-radius: 9999px;
  background: var(--bg-card);
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-muted);
  transition: color 150ms ease, border-color 150ms ease;
}
.ot-toggle:hover { color: var(--accent); border-color: var(--accent); }
.ot-toggle:focus-visible { outline: 2px solid var(--accent); }
.ot-toggle--closed { color: var(--accent); border-color: color-mix(in srgb, var(--accent) 45%, var(--border)); }
.ot-node > .ot-toggle + .ot-children { margin-top: -0.7rem; }

@media (prefers-reduced-motion: reduce) {
  .ot-card, .ot-stack__item { transition: none; }
}
</style>
