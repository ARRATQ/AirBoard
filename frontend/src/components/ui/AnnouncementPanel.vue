<template>
  <section
    class="ann"
    :class="`ann--${current.type || 'info'}`"
    :aria-label="$t('home.announcements.region')"
    @mouseenter="pause"
    @mouseleave="resume"
    @focusin="pause"
    @focusout="resume"
  >
    <span class="ann__type">
      <Icon :icon="iconFor(current.type)" class="ann__type-icon" aria-hidden="true" />
      <span class="ann__type-label">{{ $t(`announcements.type_${current.type || 'info'}`) }}</span>
    </span>

    <button
      type="button"
      class="ann__text"
      :title="$t('home.announcements.open')"
      aria-live="polite"
      @click="openModal"
    >
      <Transition name="ann-fade" mode="out-in">
        <span :key="index" class="ann__line">
          <strong v-if="current.title" class="ann__title">{{ current.title }}</strong>
          <span v-if="current.content" class="ann__content">{{ current.content }}</span>
        </span>
      </Transition>
    </button>

    <span v-if="announcements.length > 1" class="ann__nav">
      <button type="button" class="ann__btn" :aria-label="$t('home.announcements.prev')" @click="prev">
        <Icon icon="mdi:chevron-left" class="ann__chev" aria-hidden="true" />
      </button>
      <span class="ann__count">{{ index + 1 }}/{{ announcements.length }}</span>
      <button type="button" class="ann__btn" :aria-label="$t('home.announcements.next')" @click="next">
        <Icon icon="mdi:chevron-right" class="ann__chev" aria-hidden="true" />
      </button>
      <button
        v-if="!motionReduced"
        type="button"
        class="ann__btn"
        :aria-label="$t(userPaused ? 'home.announcements.play' : 'home.announcements.pause')"
        :aria-pressed="userPaused"
        @click="togglePause"
      >
        <Icon :icon="userPaused ? 'mdi:play' : 'mdi:pause'" class="ann__chev ann__chev--fixed" aria-hidden="true" />
      </button>
    </span>

    <BaseModal :open="modalOpen" :title="current.title || $t('home.announcements.region')" @close="modalOpen = false">
      <p class="ann-modal__type">{{ $t(`announcements.type_${current.type || 'info'}`) }}</p>
      <p class="ann-modal__content">{{ current.content }}</p>
    </BaseModal>
  </section>
</template>

<script setup>
// Panneau d'annonces de la barre du haut : le seul élément « tableau d'affichage » sombre avec
// le bandeau des chiffres clés. Défilement automatique toutes les 5 s, en pause au survol, au
// focus, pendant la lecture d'une annonce, et coupé sous prefers-reduced-motion.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Icon } from '@iconify/vue'
import BaseModal from '@/components/ui/BaseModal.vue'

const props = defineProps({
  announcements: { type: Array, required: true },
})

const ROTATE_MS = 5000
const ICONS = {
  info: 'mdi:information-outline',
  warning: 'mdi:alert-outline',
  success: 'mdi:check-circle-outline',
  error: 'mdi:alert-circle-outline',
}

const index = ref(0)
const modalOpen = ref(false)
let timer = null

const current = computed(() => props.announcements[index.value] || {})
const iconFor = (type) => ICONS[type] || ICONS.info

const count = () => props.announcements.length
const next = () => { index.value = (index.value + 1) % count() }
const prev = () => { index.value = (index.value - 1 + count()) % count() }

const motionReduced = Boolean(window.matchMedia?.('(prefers-reduced-motion: reduce)').matches)
const userPaused = ref(false)

function stop() {
  clearInterval(timer)
  timer = null
}

function start() {
  stop()
  if (count() > 1 && !modalOpen.value && !motionReduced && !userPaused.value) timer = setInterval(next, ROTATE_MS)
}

function togglePause() {
  userPaused.value = !userPaused.value
  if (userPaused.value) stop()
}

const pause = stop
const resume = start

function openModal() {
  stop()
  modalOpen.value = true
}

watch(modalOpen, (open) => { if (!open) start() })
watch(() => props.announcements, () => { index.value = 0; start() })

onMounted(start)
onBeforeUnmount(stop)
</script>

<style scoped>
.ann {
  --ann-tone: var(--board-dim);
  flex: 1 1 320px;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  height: 34px;
  padding-inline: 12px 4px;
  border-radius: var(--r-s);
  background: var(--board);
  color: var(--board-text);
  font-family: var(--f-num);
  font-size: 12.5px;
}

/* La couleur ne signale que l'exception : alerte, erreur, bravo. */
.ann--warning { --ann-tone: var(--board-hi); }
.ann--error { --ann-tone: var(--board-bad); }
.ann--success { --ann-tone: var(--board-ok); }
.ann--info { --ann-tone: #9ecbff; }

.ann__type {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--ann-tone);
}

.ann__type-icon { width: 15px; height: 15px; }

.ann__text {
  flex: 1;
  min-width: 0;
  height: 100%;
  padding: 0;
  border: 0;
  background: none;
  color: inherit;
  font: inherit;
  text-align: start;
  cursor: pointer;
}

.ann__text:focus-visible,
.ann__btn:focus-visible {
  outline: 2px solid var(--board-hi);
  outline-offset: -2px;
  border-radius: var(--r-s);
}

.ann__line {
  display: block;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.ann__title { font-weight: 600; margin-inline-end: 8px; }
.ann__content { color: var(--board-dim); }
.ann__text:hover .ann__content { color: var(--board-text); }

.ann__nav {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 2px;
  color: var(--board-dim);
  font-size: 11px;
}

.ann__count { min-width: 3ch; text-align: center; direction: ltr; }

.ann__btn {
  display: grid;
  place-items: center;
  width: 26px;
  height: 26px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  color: var(--board-text);
  cursor: pointer;
}

.ann__btn:hover { background: rgb(255 255 255 / 0.1); }
.ann__chev { width: 16px; height: 16px; }
[dir='rtl'] .ann__chev:not(.ann__chev--fixed) { transform: scaleX(-1); }

.ann-fade-enter-active,
.ann-fade-leave-active { transition: opacity 180ms ease, transform 180ms ease; }
.ann-fade-enter-from { opacity: 0; transform: translateY(4px); }
.ann-fade-leave-to { opacity: 0; transform: translateY(-4px); }

@media (prefers-reduced-motion: reduce) {
  .ann-fade-enter-active,
  .ann-fade-leave-active { transition: none; }
}

.ann-modal__type {
  margin: 0 0 8px;
  font-family: var(--f-num);
  font-size: 11px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
}

.ann-modal__content {
  margin: 0;
  white-space: pre-line;
  line-height: 1.55;
  color: var(--text);
}
</style>
