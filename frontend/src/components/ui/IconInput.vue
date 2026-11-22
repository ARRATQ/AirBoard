<template>
  <div class="form-group">
    <label v-if="label" :for="id" class="form-label" :class="{ 'form-label-required': required }">
      {{ label }}
    </label>

    <div class="ii__row">
      <span class="ii__preview" aria-hidden="true">
        <Icon :icon="isIconName(modelValue) ? modelValue : 'lucide:circle-help'" />
      </span>
      <input
        :id="id"
        :value="modelValue"
        type="text"
        dir="ltr"
        spellcheck="false"
        autocomplete="off"
        :placeholder="placeholder || t('ui.iconInput.placeholder')"
        :required="required"
        class="form-input ii__input"
        @input="$emit('update:modelValue', $event.target.value.trim())"
      />
      <button
        type="button"
        class="btn btn-secondary ii__toggle"
        :aria-expanded="open"
        :aria-controls="`${id}-panel`"
        @click="open = !open"
      >
        <Icon :icon="open ? 'lucide:x' : 'lucide:search'" aria-hidden="true" />
        {{ open ? t('ui.iconInput.close') : t('ui.iconInput.browse') }}
      </button>
    </div>

    <div v-if="error" class="form-error">{{ error }}</div>
    <p v-else-if="modelValue && !isIconName(modelValue)" class="form-error">{{ t('ui.iconInput.invalid') }}</p>

    <section v-if="open" :id="`${id}-panel`" class="ii__panel" :aria-label="t('ui.iconInput.browse')">
      <div class="ii__tools">
        <input
          ref="searchField"
          v-model="query"
          type="search"
          class="form-input ii__search"
          :aria-label="t('ui.iconInput.searchLabel')"
          :placeholder="t('ui.iconInput.searchPlaceholder')"
        />
        <div class="ii__styles" role="group" :aria-label="t('ui.iconInput.styleLabel')">
          <button
            v-for="key in STYLE_KEYS"
            :key="key"
            type="button"
            class="ii__style"
            :aria-pressed="style === key"
            @click="style = key"
          >
            {{ t(`ui.iconInput.styles.${key}`) }}
          </button>
        </div>
      </div>

      <p v-if="state === 'loading'" class="ii__note" role="status">{{ t('ui.iconInput.loading') }}</p>
      <p v-else-if="state === 'error'" class="ii__note ii__note--bad" role="alert">{{ t('ui.iconInput.searchError') }}</p>
      <p v-else-if="state === 'empty'" class="ii__note" role="status">{{ t('ui.iconInput.noResults') }}</p>
      <p v-else-if="!query.trim()" class="ii__note">{{ t('ui.iconInput.popular') }}</p>

      <ul v-if="shown.length" class="ii__grid">
        <li v-for="name in shown" :key="name">
          <button
            type="button"
            class="ii__choice"
            :class="{ 'is-selected': name === modelValue }"
            :aria-pressed="name === modelValue"
            :aria-label="name"
            :title="name"
            @click="choose(name)"
          >
            <Icon :icon="name" aria-hidden="true" />
          </button>
        </li>
      </ul>

      <p class="ii__source">{{ t('ui.iconInput.source') }}</p>
    </section>
  </div>
</template>

<script setup>
// Champ d'icône de l'administration : saisie libre « collection:nom », ou choix dans le
// catalogue Iconify (recherche filtrable par style : trait fin, plein, tous).
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { ICON_STYLES, isIconName, searchIcons } from '@/utils/iconifySearch'

const { t } = useI18n()

const props = defineProps({
  modelValue: { type: String, default: '' },
  label: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  required: { type: Boolean, default: false },
  id: { type: String, default: () => `icon-input-${Math.random().toString(36).slice(2, 11)}` },
  error: { type: String, default: '' },
  category: { type: String, default: 'general' },
})

const emit = defineEmits(['update:modelValue'])

const STYLE_KEYS = Object.keys(ICON_STYLES)
const SEARCH_DELAY_MS = 300

// Point de départ avant toute recherche, par usage du champ
const SUGGESTIONS = {
  general: ['lucide:house', 'lucide:settings', 'lucide:user', 'lucide:folder', 'lucide:file-text', 'lucide:chart-line', 'lucide:house', 'lucide:settings', 'lucide:user', 'lucide:folder'],
  application: ['lucide:app-window', 'lucide:globe', 'lucide:database', 'lucide:server', 'lucide:cloud', 'lucide:mail', 'lucide:message-circle', 'lucide:calendar', 'lucide:wrench', 'lucide:app-window', 'lucide:globe', 'lucide:git-branch'],
  group: ['lucide:folders', 'lucide:users', 'lucide:shield-check', 'lucide:code', 'lucide:chart-pie', 'lucide:briefcase', 'lucide:folders', 'lucide:users'],
  social: ['mdi:github', 'mdi:gitlab', 'mdi:slack', 'mdi:discord', 'mdi:microsoft-teams', 'mdi:google'],
}

const open = ref(false)
const query = ref('')
const style = ref('all')
const results = ref([])
const state = ref('idle') // idle | loading | done | empty | error
const searchField = ref(null)

const shown = computed(() =>
  query.value.trim() ? results.value : (SUGGESTIONS[props.category] || SUGGESTIONS.general),
)

let timer = null
let controller = null

async function runSearch() {
  controller?.abort()
  if (!query.value.trim()) {
    results.value = []
    state.value = 'idle'
    return
  }
  controller = new AbortController()
  state.value = 'loading'
  try {
    const names = await searchIcons(query.value, style.value, { signal: controller.signal })
    results.value = names
    state.value = names.length ? 'done' : 'empty'
  } catch (err) {
    if (err?.name === 'AbortError') return
    results.value = []
    state.value = 'error'
  }
}

watch([query, style], () => {
  clearTimeout(timer)
  timer = setTimeout(runSearch, SEARCH_DELAY_MS)
})

watch(open, async (isOpen) => {
  if (!isOpen) return
  await nextTick()
  searchField.value?.focus()
})

onBeforeUnmount(() => {
  clearTimeout(timer)
  controller?.abort()
})

function choose(name) {
  emit('update:modelValue', name)
}
</script>

<style scoped>
.ii__row { display: flex; gap: 8px; align-items: stretch; }

.ii__preview {
  display: grid;
  flex: none;
  place-items: center;
  width: var(--ctl-h);
  height: var(--ctl-h);
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
  color: var(--text);
}

.ii__preview :deep(svg) { width: 18px; height: 18px; }

.ii__input { flex: 1; min-width: 0; font-family: var(--f-num); }
.ii__toggle { flex: none; }
.ii__toggle :deep(svg) { width: 16px; height: 16px; }

.ii__panel {
  margin-top: 8px;
  padding: 12px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--sunken);
}

.ii__tools { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; }
.ii__search { flex: 1 1 14rem; min-width: 0; }

.ii__styles {
  display: inline-flex;
  flex: none;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
}

.ii__style {
  min-height: calc(var(--ctl-h) - 2px);
  padding: 0 10px;
  color: var(--muted);
  font-size: var(--fs-small);
}

.ii__style + .ii__style { border-inline-start: 1px solid var(--line); }
.ii__style:hover { color: var(--text); }
.ii__style[aria-pressed='true'] { background: var(--sunken); color: var(--text); font-weight: 600; }
.ii__style:focus-visible { outline: 2px solid var(--acc); outline-offset: -2px; }

.ii__note { margin: 10px 0 0; font-size: var(--fs-small); color: var(--muted); }
.ii__note--bad { color: var(--bad); }

.ii__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(40px, 1fr));
  gap: 6px;
  max-height: 232px;
  margin: 10px 0 0;
  padding: 0;
  overflow-y: auto;
  list-style: none;
}

.ii__choice {
  display: grid;
  place-items: center;
  width: 100%;
  aspect-ratio: 1;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
  color: var(--text);
  transition: border-color 120ms ease;
}

.ii__choice :deep(svg) { width: 20px; height: 20px; }
.ii__choice:hover { border-color: var(--muted); }
.ii__choice.is-selected { border-color: var(--acc-ink); box-shadow: inset 0 0 0 1px var(--acc-ink); color: var(--acc-ink); }
.ii__choice:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

.ii__source { margin: 10px 0 0; font-size: var(--fs-eyebrow); color: var(--muted); }

@media (prefers-reduced-motion: reduce) {
  .ii__choice { transition: none; }
}
</style>
