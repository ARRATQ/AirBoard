<template>
  <Teleport to="body">
    <Transition name="help-fade">
      <div v-if="modelValue" class="help-backdrop" aria-hidden="true" @click="close" />
    </Transition>

    <Transition name="help-pop">
      <div v-if="modelValue" class="help-layer">
        <div v-focus-trap role="dialog" aria-modal="true" aria-labelledby="help-title" class="help">
          <header class="help__head">
            <span class="help__mark" aria-hidden="true"><Icon icon="lucide:circle-help" /></span>
            <div class="help__heading">
              <h2 id="help-title" class="help__title">{{ $t('help.title') }}</h2>
              <p class="help__subtitle">{{ $t('help.subtitle') }}</p>
            </div>
            <button type="button" class="help__close" :aria-label="$t('common.close')" @click="close">
              <Icon icon="lucide:x" aria-hidden="true" />
            </button>
          </header>

          <div class="help__body">
            <!-- Mobile : liste déroulante des guides -->
            <div class="help__picker">
              <label for="help-topic" class="sr-only">{{ $t('help.chooseTopic') }}</label>
              <select id="help-topic" class="form-select" :value="currentId" @change="select($event.target.value)">
                <optgroup v-for="group in groups" :key="group.audience" :label="$t(`help.audiences.${group.audience}`)">
                  <option v-for="topic in group.topics" :key="topic.id" :value="topic.id">{{ $t(`help.topics.${topic.id}`) }}</option>
                </optgroup>
              </select>
            </div>

            <nav class="help__nav" :aria-label="$t('help.chooseTopic')">
              <section v-for="group in groups" :key="group.audience" class="help__group">
                <h3 class="help__group-title">{{ $t(`help.audiences.${group.audience}`) }}</h3>
                <ul>
                  <li v-for="topic in group.topics" :key="topic.id">
                    <button
                      type="button"
                      class="help__topic"
                      :class="{ 'help__topic--on': topic.id === currentId }"
                      :aria-current="topic.id === currentId ? 'page' : undefined"
                      @click="select(topic.id)"
                    >
                      <Icon :icon="topic.icon" aria-hidden="true" />
                      <span>{{ $t(`help.topics.${topic.id}`) }}</span>
                    </button>
                  </li>
                </ul>
              </section>
            </nav>

            <article ref="article" class="help__content" aria-live="polite" @click="followLink">
              <template v-if="current">
                <p class="help__eyebrow">
                  <span>{{ $t(`help.audiences.${current.audience}`) }}</span>
                  <span v-if="scope" class="help__scope" :class="`help__scope--${scope}`">
                    <Icon :icon="scope === 'all' ? 'lucide:globe' : 'lucide:users-round'" aria-hidden="true" />
                    {{ $t(`help.scope.${scope}`) }}
                  </span>
                </p>
                <div v-if="loading" class="help__state">
                  <Icon icon="lucide:loader-circle" class="help__spin" aria-hidden="true" />
                  <span class="sr-only">{{ $t('common.loading') }}</span>
                </div>
                <p v-else-if="loadError" class="help__state">{{ $t('help.loadError') }}</p>
                <!-- eslint-disable-next-line vue/no-v-html -- Markdown de l'application, nettoyé par DOMPurify -->
                <div v-else class="help-prose" v-html="html" />
              </template>
            </article>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
// Aide intégrée : guides choisis selon les permissions de la personne (help/catalog.js),
// ouverte sur le guide de la page affichée. Les passages propres à une portée (« tout le
// site » / « vos groupes ») ne s'affichent qu'à qui a cette portée.
import { computed, nextTick, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAuthStore } from '@/stores/auth'
import { useEscapeClose } from '@/composables/useEscapeClose'
import { getRoles } from '@/config/access'
import { HELP_AUDIENCES, visibleTopics, topicScope, topicForRoute, applyScope } from '@/help/catalog'
import { loadTopic } from '@/help/content'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])

const { locale } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const roles = computed(() => getRoles(authStore.user))
const topics = computed(() => visibleTopics(roles.value))
const groups = computed(() =>
  HELP_AUDIENCES
    .map((audience) => ({ audience, topics: topics.value.filter((t) => t.audience === audience) }))
    .filter((g) => g.topics.length > 0)
)

const currentId = ref('')
const current = computed(() => topics.value.find((t) => t.id === currentId.value) || null)
const scope = computed(() => (current.value ? topicScope(current.value, roles.value) : ''))

const markdown = ref('')
const loading = ref(false)
const loadError = ref(false)
const article = ref(null)
let request = 0

const html = computed(() => DOMPurify.sanitize(marked.parse(applyScope(markdown.value, scope.value), { gfm: true })))

const load = async () => {
  if (!current.value) return
  const id = current.value.id
  const mine = ++request
  loading.value = true
  loadError.value = false
  try {
    const text = await loadTopic(locale.value, id)
    if (mine !== request) return
    markdown.value = text
  } catch {
    if (mine === request) loadError.value = true
  } finally {
    if (mine === request) loading.value = false
  }
  await nextTick()
  if (article.value) article.value.scrollTop = 0
}

const select = (id) => {
  currentId.value = id
}

const close = () => emit('update:modelValue', false)

// Liens internes des guides (« [Ressources](/resources) ») : navigation dans l'application
const followLink = (event) => {
  const link = event.target.closest?.('a')
  const href = link?.getAttribute('href')
  if (!href || !href.startsWith('/') || href.startsWith('//')) return
  event.preventDefault()
  close()
  router.push(href)
}

watch(() => props.modelValue, (open) => {
  if (!open) return
  const forPage = topicForRoute(route.path, topics.value)
  currentId.value = (forPage || topics.value[0])?.id || ''
}, { immediate: true })

// immediate : la fenêtre n'est créée qu'à la première ouverture, guide déjà choisi
watch([currentId, locale], load, { immediate: true })

useEscapeClose(() => props.modelValue, close)
</script>

<style scoped>
.help-backdrop { position: fixed; inset: 0; z-index: 40; background: rgb(0 0 0 / 0.5); }

.help-layer {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: grid;
  place-items: center;
  padding: 16px;
  pointer-events: none;
}

.help {
  pointer-events: auto;
  display: flex;
  flex-direction: column;
  width: min(100%, 960px);
  height: min(100%, 720px);
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: var(--r-l);
  background: var(--bg-card);
  box-shadow: var(--shadow-float);
}

.help__head {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
}

.help__mark {
  display: grid;
  flex: none;
  place-items: center;
  width: 36px;
  height: 36px;
  border-radius: var(--r-m);
  background: var(--acc);
  color: var(--on-fill, #fff);
  font-size: 20px;
}

.help__heading { flex: 1; min-width: 0; }
.help__title { margin: 0; font-size: 17px; font-weight: 700; color: var(--text); }
.help__subtitle { margin: 0; font-size: 13px; color: var(--muted); }

.help__close {
  display: grid;
  place-items: center;
  width: 36px;
  height: 36px;
  border-radius: var(--r-s);
  color: var(--muted);
  font-size: 20px;
}

.help__close:hover { background: var(--hover); color: var(--text); }
.help__close:focus-visible, .help__topic:focus-visible { outline: 2px solid var(--acc); outline-offset: 2px; }

.help__body { display: flex; flex: 1; min-height: 0; }
.help__picker { display: none; }

.help__nav {
  flex: none;
  width: 250px;
  overflow-y: auto;
  padding: 14px 10px;
  border-inline-end: 1px solid var(--border);
  background: var(--bg-surface);
}

.help__group + .help__group { margin-top: 14px; }

.help__group-title {
  margin: 0 0 4px;
  padding: 0 10px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--muted);
}

.help__topic {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 7px 10px;
  border-radius: var(--r-s);
  font-size: 14px;
  color: var(--text);
  text-align: start;
}

.help__topic svg { flex: none; width: 16px; height: 16px; color: var(--muted); }
.help__topic:hover { background: var(--hover); }
.help__topic--on { background: var(--bg-card); box-shadow: inset 3px 0 0 var(--acc); font-weight: 600; }
[dir='rtl'] .help__topic--on { box-shadow: inset -3px 0 0 var(--acc); }
.help__topic--on svg { color: var(--acc-ink, var(--acc)); }

.help__content { flex: 1; min-width: 0; overflow-y: auto; padding: 22px 28px 32px; }

.help__eyebrow {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin: 0 0 6px;
  font-size: 11.5px;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--muted);
}

.help__scope {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border: 1px solid var(--border);
  border-radius: 999px;
  letter-spacing: 0;
  text-transform: none;
  color: var(--text);
  background: var(--bg-surface);
}

.help__scope svg { width: 13px; height: 13px; }
.help__scope--groups svg { color: var(--acc-ink, var(--acc)); }

.help__state { display: grid; place-items: center; min-height: 10rem; color: var(--muted); }
.help__spin { width: 24px; height: 24px; animation: help-spin 0.9s linear infinite; }
@keyframes help-spin { to { transform: rotate(360deg); } }

/* Texte des guides */
.help-prose { max-width: 68ch; font-size: 15px; line-height: 1.65; color: var(--text); }
.help-prose :deep(h1) { margin: 0 0 12px; font-size: 24px; line-height: 1.25; font-weight: 700; }
.help-prose :deep(h2) { margin: 26px 0 8px; font-size: 17px; font-weight: 700; }
.help-prose :deep(h3) { margin: 18px 0 6px; font-size: 15px; font-weight: 700; }
.help-prose :deep(p) { margin: 0 0 10px; }
.help-prose :deep(ul), .help-prose :deep(ol) { margin: 0 0 12px; padding-inline-start: 22px; }
.help-prose :deep(ul) { list-style: disc; }
.help-prose :deep(ol) { list-style: decimal; }
.help-prose :deep(li) { margin: 3px 0; }
.help-prose :deep(li::marker) { color: var(--muted); }
.help-prose :deep(strong) { font-weight: 650; }
.help-prose :deep(a) { color: var(--acc-ink, var(--acc)); text-decoration: underline; text-underline-offset: 2px; }
.help-prose :deep(code) { padding: 1px 5px; border-radius: 4px; background: var(--bg-surface); font-size: 0.9em; }

/* Encadré « Bon à savoir » : citations Markdown */
.help-prose :deep(blockquote) {
  margin: 14px 0;
  padding: 10px 14px;
  border-inline-start: 3px solid var(--acc);
  border-radius: var(--r-s);
  background: var(--bg-surface);
}

.help-prose :deep(blockquote p:last-child) { margin-bottom: 0; }

.help-prose :deep(table) {
  width: 100%;
  margin: 12px 0 16px;
  border-collapse: collapse;
  font-size: 0.92em;
}

.help-prose :deep(th),
.help-prose :deep(td) {
  padding: 8px 10px;
  border-bottom: 1px solid var(--border);
  text-align: start;
  vertical-align: top;
}

.help-prose :deep(th) {
  color: var(--muted);
  font-size: 0.8em;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

/* Transitions */
.help-fade-enter-active, .help-fade-leave-active { transition: opacity 0.2s ease; }
.help-fade-enter-from, .help-fade-leave-to { opacity: 0; }
.help-pop-enter-active, .help-pop-leave-active { transition: opacity 0.2s ease, transform 0.2s ease; }
.help-pop-enter-from, .help-pop-leave-to { opacity: 0; transform: scale(0.97); }

@media (max-width: 720px) {
  .help-layer { padding: 0; }
  .help { width: 100%; height: 100%; border: 0; border-radius: 0; }
  .help__body { flex-direction: column; }
  .help__nav { display: none; }
  .help__picker { display: block; padding: 10px 16px; border-bottom: 1px solid var(--border); background: var(--bg-surface); }
  .help__content { padding: 18px 16px 28px; }
}

@media (prefers-reduced-motion: reduce) {
  .help-fade-enter-active, .help-fade-leave-active, .help-pop-enter-active, .help-pop-leave-active { transition: none; }
  .help__spin { animation: none; }
}
</style>
