<template>
  <div ref="root" class="combo" :class="{ 'combo--open': open, 'combo--up': openUp, 'combo--disabled': disabled }">
    <div class="combo__field" @click="focusInput">
      <ul v-if="chips.length" class="combo__chips" :aria-label="selectedLabel">
        <li v-for="chip in chips" :key="chip.id" class="combo__chip" :class="{ 'combo__chip--locked': chip.locked }">
          <span v-if="kind === 'person'" class="combo__avatar" aria-hidden="true">{{ initials(chip.label) }}</span>
          <span v-else class="combo__dot" :style="{ backgroundColor: chip.color || undefined }" aria-hidden="true"></span>
          <span class="combo__chip-label">{{ chip.label }}</span>
          <Icon v-if="chip.locked" icon="lucide:lock" class="combo__chip-lock" :aria-label="t('combo.locked')" />
          <button
            v-else-if="!disabled"
            type="button"
            class="combo__chip-remove"
            :aria-label="t('combo.remove', { name: chip.label })"
            @click.stop="remove(chip.id)"
          >
            <Icon icon="lucide:x" aria-hidden="true" />
          </button>
        </li>
      </ul>
      <input
        :id="inputId"
        ref="input"
        v-model="query"
        type="text"
        class="combo__input"
        role="combobox"
        autocomplete="off"
        :aria-label="label"
        :aria-expanded="open"
        :aria-controls="listId"
        :aria-activedescendant="open && active >= 0 ? optionId(active) : undefined"
        aria-autocomplete="list"
        :placeholder="chips.length ? '' : placeholder"
        :disabled="disabled"
        @focus="openList"
        @keydown="onKeydown"
      />
      <Icon v-if="loading" icon="lucide:loader-circle" class="combo__spinner animate-spin" aria-hidden="true" />
      <Icon v-else icon="lucide:search" class="combo__search" aria-hidden="true" />
    </div>

    <ul
      v-show="open"
      :id="listId"
      class="combo__list"
      role="listbox"
      aria-multiselectable="true"
      :aria-label="label"
    >
      <li
        v-for="(option, index) in results"
        :id="optionId(index)"
        :key="option.id"
        role="option"
        class="combo__option"
        :class="{ 'combo__option--active': index === active }"
        :aria-selected="isSelected(option.id)"
        @mousedown.prevent
        @mouseenter="active = index"
        @click="toggle(option)"
      >
        <span v-if="kind === 'person'" class="combo__avatar" aria-hidden="true">{{ initials(option.label) }}</span>
        <span v-else class="combo__dot" :style="{ backgroundColor: option.color || undefined }" aria-hidden="true"></span>
        <span class="combo__option-text">
          <span class="combo__option-label">{{ option.label }}</span>
          <span v-if="option.sublabel" class="combo__option-sub">{{ option.sublabel }}</span>
        </span>
        <Icon v-if="isSelected(option.id)" icon="lucide:check" class="combo__check" aria-hidden="true" />
      </li>
      <li v-if="!loading && results.length === 0" class="combo__empty" role="presentation">
        {{ query.trim() ? t('combo.noMatch') : emptyText || t('combo.noOption') }}
      </li>
    </ul>
    <p class="sr-only" aria-live="polite">{{ open ? t('combo.results', results.length) : '' }}</p>
  </div>
</template>

<script setup>
// Sélection multiple avec recherche : remplace les longues listes à cocher (groupes, personnes).
// Les options connues sont filtrées sur place ; `search` (facultatif) interroge le serveur.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { matchesQuery } from '@/utils/comboFilter'

const props = defineProps({
  // Identifiants sélectionnés
  modelValue: { type: Array, required: true },
  // Options connues : { id, label, sublabel?, color? } (servent aussi d'étiquettes aux sélectionnés)
  options: { type: Array, default: () => [] },
  // Étiquettes de sélectionnés absents des options (non proposés, ex. groupes verrouillés)
  labels: { type: Array, default: () => [] },
  // Recherche distante facultative : (texte) => Promise<options>
  search: { type: Function, default: null },
  // Sélectionnés que l'utilisateur ne peut pas retirer
  lockedIds: { type: Array, default: () => [] },
  // 'group' (pastille de couleur) ou 'person' (initiales)
  kind: { type: String, default: 'group' },
  label: { type: String, required: true },
  placeholder: { type: String, default: '' },
  emptyText: { type: String, default: '' },
  inputId: { type: String, default: undefined },
  disabled: { type: Boolean, default: false },
  maxResults: { type: Number, default: 50 },
})

const emit = defineEmits(['update:modelValue'])
const { t } = useI18n()

const SEARCH_DEBOUNCE_MS = 250
const uid = `combo-${Math.random().toString(36).slice(2, 9)}`
const listId = `${uid}-list`
const optionId = (index) => `${uid}-opt-${index}`

const root = ref(null)
const input = ref(null)
const query = ref('')
const open = ref(false)
// La liste s'ouvre vers le haut quand la place manque dessous (bas d'une fenêtre qui défile)
const openUp = ref(false)
const LIST_ROOM = 256
const active = ref(-1)
const loading = ref(false)
const remote = ref([])
// Étiquettes vues (options et résultats distants), pour afficher les sélectionnés
const known = ref(new Map())

const remember = (list) => {
  const next = new Map(known.value)
  for (const option of list) next.set(option.id, option)
  known.value = next
}
watch(() => props.options, (list) => remember(list), { immediate: true })
watch(() => props.labels, (list) => remember(list), { immediate: true })

const isSelected = (id) => props.modelValue.includes(id)
const isLocked = (id) => props.lockedIds.includes(id)

const chips = computed(() =>
  props.modelValue.map((id) => {
    const option = known.value.get(id)
    return { id, label: option?.label ?? `#${id}`, color: option?.color, locked: isLocked(id) }
  }),
)

const results = computed(() => {
  const source = props.search && query.value.trim() ? remote.value : props.options
  return source.filter((option) => matchesQuery(option, query.value)).slice(0, props.maxResults)
})

const selectedLabel = computed(() => t('combo.selected', { label: props.label }))

const initials = (label) =>
  String(label || '?')
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase()

let searchTimer = null
let searchRequest = 0
watch(query, (text) => {
  active.value = results.value.length ? 0 : -1
  if (!props.search) return
  clearTimeout(searchTimer)
  if (!text.trim()) {
    remote.value = []
    loading.value = false
    return
  }
  loading.value = true
  searchTimer = setTimeout(async () => {
    const request = ++searchRequest
    try {
      const list = await props.search(text.trim())
      if (request !== searchRequest) return
      remote.value = Array.isArray(list) ? list : []
      remember(remote.value)
      active.value = remote.value.length ? 0 : -1
    } catch (error) {
      if (request !== searchRequest) return
      console.error('Recherche impossible :', error)
      remote.value = []
    } finally {
      if (request === searchRequest) loading.value = false
    }
  }, SEARCH_DEBOUNCE_MS)
})

const focusInput = () => {
  if (!props.disabled) input.value?.focus()
}
// Bas visible : celui de la fenêtre, ou du premier parent qui fait défiler son contenu
const visibleBottom = (el) => {
  let bottom = window.innerHeight
  for (let node = el.parentElement; node; node = node.parentElement) {
    const { overflowY } = getComputedStyle(node)
    if (overflowY === 'auto' || overflowY === 'scroll') {
      bottom = Math.min(bottom, node.getBoundingClientRect().bottom)
      break
    }
  }
  return bottom
}
// Hauteur que prendra la liste : une ligne par proposition, plafonnée par max-height
const OPTION_H = 40
const listHeight = () => Math.min(LIST_ROOM, Math.max(1, results.value.length) * OPTION_H + 16)
const placeList = () => {
  if (!root.value) return
  const rect = root.value.getBoundingClientRect()
  const below = visibleBottom(root.value) - rect.bottom
  openUp.value = below < listHeight() && rect.top > below
}
const openList = () => {
  if (props.disabled) return
  if (!open.value) placeList()
  open.value = true
  if (active.value < 0 && results.value.length) active.value = 0
}
const close = () => {
  open.value = false
  active.value = -1
}

const toggle = (option) => {
  if (isSelected(option.id)) {
    remove(option.id)
  } else {
    remember([option])
    emit('update:modelValue', [...props.modelValue, option.id])
  }
  query.value = ''
  focusInput()
}
const remove = (id) => {
  if (isLocked(id)) return
  emit('update:modelValue', props.modelValue.filter((value) => value !== id))
}

const move = (step) => {
  if (!open.value) openList()
  const count = results.value.length
  if (!count) return
  active.value = (active.value + step + count) % count
}

const onKeydown = (event) => {
  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      move(1)
      break
    case 'ArrowUp':
      event.preventDefault()
      move(-1)
      break
    case 'Enter':
      if (open.value && results.value[active.value]) {
        event.preventDefault()
        toggle(results.value[active.value])
      }
      break
    case 'Escape':
      if (open.value) {
        // Ferme la liste sans fermer la fenêtre qui la contient (useEscapeClose)
        event.preventDefault()
        close()
      }
      break
    case 'Backspace':
      if (!query.value) {
        const removable = props.modelValue.filter((id) => !isLocked(id))
        if (removable.length) remove(removable[removable.length - 1])
      }
      break
    case 'Tab':
      close()
      break
  }
}

const onDocumentPointer = (event) => {
  if (root.value && !root.value.contains(event.target)) close()
}
onMounted(() => document.addEventListener('pointerdown', onDocumentPointer))
onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocumentPointer)
  clearTimeout(searchTimer)
})
</script>

<style scoped>
.combo {
  position: relative;
}

.combo__field {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  min-height: var(--ctl-h);
  padding-block: 3px;
  padding-inline: 4px 30px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background-color: var(--surface);
  cursor: text;
  transition: border-color 120ms ease, box-shadow 120ms ease;
}

.combo:focus-within .combo__field {
  border-color: var(--acc);
  box-shadow: 0 0 0 2px color-mix(in oklab, var(--acc) 28%, transparent);
}

.combo--disabled .combo__field {
  background-color: var(--sunken);
  cursor: not-allowed;
}

.combo__chips {
  display: contents;
}

.combo__chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  min-height: 24px;
  padding-inline: 6px 2px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background-color: var(--sunken);
  color: var(--text);
  font-size: var(--fs-small);
  line-height: 1.2;
}

.combo__chip--locked {
  padding-inline-end: 6px;
  color: var(--muted);
}

.combo__chip-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.combo__chip-lock {
  width: 12px;
  height: 12px;
  flex: none;
}

.combo__chip-remove {
  display: grid;
  place-items: center;
  width: 20px;
  height: 20px;
  flex: none;
  border-radius: var(--r-s);
  color: var(--muted);
}

.combo__chip-remove:hover {
  background-color: color-mix(in oklab, var(--bad) 12%, transparent);
  color: var(--bad);
}

.combo__chip-remove:focus-visible {
  outline: 2px solid var(--fill);
  outline-offset: 1px;
}

.combo__chip-remove :deep(svg) {
  width: 13px;
  height: 13px;
}

.combo__input {
  flex: 1 1 120px;
  min-width: 0;
  height: 24px;
  padding: 0 6px;
  border: 0;
  background: transparent;
  color: var(--text);
  font-size: var(--fs-body);
}

.combo__input:focus {
  outline: none;
  box-shadow: none;
}

.combo__input::placeholder {
  color: var(--muted);
}

.combo__search,
.combo__spinner {
  position: absolute;
  inset-inline-end: 9px;
  top: calc(var(--ctl-h) / 2 - 7px);
  width: 14px;
  height: 14px;
  color: var(--muted);
  pointer-events: none;
}

.combo__dot {
  width: 9px;
  height: 9px;
  flex: none;
  border-radius: 50%;
  background-color: var(--muted);
}

.combo__avatar {
  display: grid;
  place-items: center;
  width: 20px;
  height: 20px;
  flex: none;
  border-radius: 50%;
  background-color: var(--board);
  color: var(--board-text);
  font-family: var(--f-num);
  font-size: 9.5px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.combo__list {
  position: absolute;
  z-index: 30;
  inset-inline: 0;
  top: calc(100% + 4px);
  max-height: 248px;
  overflow-y: auto;
  padding: 4px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background-color: var(--surface);
  box-shadow: var(--shadow-float);
  animation: combo-in 120ms ease-out;
}

.combo--up .combo__list {
  top: auto;
  bottom: calc(100% + 4px);
}

.combo__option {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 36px;
  padding: 4px 8px;
  border-radius: var(--r-s);
  color: var(--text);
  font-size: var(--fs-body);
  cursor: pointer;
}

.combo__option--active {
  background-color: var(--hover);
}

.combo__option[aria-selected='true'] .combo__option-label {
  font-weight: 600;
}

.combo__option-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.combo__option-label,
.combo__option-sub {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.combo__option-sub {
  color: var(--muted);
  font-size: var(--fs-small);
}

.combo__check {
  width: 15px;
  height: 15px;
  flex: none;
  color: var(--fill);
}

.combo__empty {
  padding: 10px 8px;
  color: var(--muted);
  font-size: var(--fs-small);
}

@keyframes combo-in {
  from {
    opacity: 0;
    transform: translateY(-3px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .combo__field {
    transition: none;
  }

  .combo__list {
    animation: none;
  }
}
</style>
