<template>
  <div v-if="activeChatbots.length > 0" class="chatbot-widget-root">

    <!--
      v-if="selectedBot"  → détruit/recrée le DOM quand on change de bot (nouvelle session n8n)
      v-show="chatOpen"   → masque/affiche sans détruire le DOM (session n8n préservée entre ouvertures)
    -->
    <template v-if="selectedBot">
      <transition name="chatbot-panel-anim">
        <div
          v-show="chatOpen"
          class="chatbot-panel"
          :class="{ 'chatbot-panel--fullscreen': isFullscreen }"
        >
          <!-- En-tête coloré -->
          <div
            class="chatbot-panel-header"
            :style="{ background: `linear-gradient(135deg, ${selectedBot.color || '#7c3aed'}, ${selectedBot.color || '#4f46e5'})` }"
          >
            <div class="flex items-center gap-2 min-w-0">
              <Icon :icon="selectedBot.icon || 'mdi:robot-outline'" class="h-5 w-5 text-white flex-shrink-0" />
              <div class="min-w-0">
                <p class="font-semibold text-white text-sm leading-tight truncate">
                  {{ selectedBot.welcome_title || selectedBot.name }}
                </p>
                <p v-if="selectedBot.welcome_subtitle" class="text-white/70 text-xs leading-tight truncate">
                  {{ selectedBot.welcome_subtitle }}
                </p>
              </div>
            </div>
            <div class="flex items-center gap-1 flex-shrink-0">
              <button
                v-if="activeChatbots.length > 1"
                @click="backToSelector"
                class="chatbot-panel-btn"
                :title="$t('chatbots.switchBot')"
              >
                <Icon icon="mdi:swap-horizontal" class="h-4 w-4" />
              </button>
              <button
                @click="isFullscreen = !isFullscreen"
                class="chatbot-panel-btn"
                :title="isFullscreen ? $t('chatbots.exitFullscreen') : $t('chatbots.fullscreen')"
              >
                <Icon :icon="isFullscreen ? 'mdi:fullscreen-exit' : 'mdi:fullscreen'" class="h-4 w-4" />
              </button>
              <button @click="chatOpen = false" class="chatbot-panel-btn" :title="$t('common.close')">
                <Icon icon="mdi:close" class="h-4 w-4" />
              </button>
            </div>
          </div>

          <!-- Conteneur @n8n/chat (jamais détruit tant que selectedBot ne change pas) -->
          <div
            :id="`n8n-chat-${selectedBot.id}`"
            class="chatbot-chat-container"
            :class="{
              'n8n-hide-header': selectedBot.hide_header,
              'n8n-header-hidden': headerHidden
            }"
            :style="{ '--chat--color-primary': selectedBot.color || '#4f46e5', '--chat--color-secondary': selectedBot.color || '#7c3aed' }"
          />
        </div>
      </transition>
    </template>

    <!-- ── Zone avatar — EXTÉRIEURE au panel, à sa gauche ── -->
    <transition name="avatar-zone-anim">
      <div
        v-if="showAvatar && selectedBot?.show_avatar_intro && !isFullscreen"
        class="chatbot-avatar-zone"
      >
        <!-- Bulle de bienvenue (temporaire) au-dessus du visage -->
        <transition name="bubble-anim">
          <div
            v-if="showBubble"
            class="chatbot-bubble"
            @click="dismissBubble"
          >
            <button class="chatbot-bubble-close" @click.stop="dismissBubble" :title="$t('common.close')">
              <Icon icon="mdi:close" class="h-3 w-3" />
            </button>
            <p class="font-semibold text-sm text-gray-900 dark:text-white leading-tight">
              {{ selectedBot.welcome_title || selectedBot.name }}
            </p>
            <p
              v-if="selectedBot.welcome_subtitle"
              class="text-xs text-gray-500 dark:text-gray-400 mt-1 leading-snug"
            >
              {{ selectedBot.welcome_subtitle }}
            </p>
          </div>
        </transition>

        <!-- Visage de l'avatar -->
        <div
          class="chatbot-avatar-face"
          :style="{
            background: `radial-gradient(circle at 38% 38%, ${lighten(selectedBot.color || '#4f46e5')}, ${selectedBot.color || '#4f46e5'})`
          }"
        >
          <div class="chatbot-avatar-shine" />
          <Icon :icon="selectedBot.icon || 'mdi:robot-outline'" class="chatbot-avatar-icon" />
        </div>

        <!-- Nom du bot -->
        <span class="chatbot-avatar-label">{{ selectedBot.name }}</span>
      </div>
    </transition>

    <!-- ── Sélecteur multi-bots ─────────────────────────── -->
    <transition name="chatbot-menu-anim">
      <div v-if="chatOpen && !selectedBot && activeChatbots.length > 1" class="chatbot-selector-menu">
        <p class="text-xs font-semibold text-gray-500 dark:text-gray-400 px-3 pt-2 pb-1 uppercase tracking-wide">
          {{ $t('chatbots.chooseBotTitle') }}
        </p>
        <button
          v-for="bot in activeChatbots"
          :key="bot.id"
          @click="selectBot(bot)"
          class="chatbot-selector-item"
        >
          <div
            class="h-8 w-8 rounded-lg flex items-center justify-center flex-shrink-0"
            :style="{ background: (bot.color || '#4f46e5') + '22', border: '1px solid ' + (bot.color || '#4f46e5') }"
          >
            <Icon :icon="bot.icon || 'mdi:robot-outline'" class="h-4 w-4" :style="{ color: bot.color || '#4f46e5' }" />
          </div>
          <div class="text-left">
            <p class="text-sm font-medium text-gray-900 dark:text-white">{{ bot.name }}</p>
            <p v-if="bot.description" class="text-xs text-gray-500 dark:text-gray-400 truncate max-w-40">
              {{ bot.description }}
            </p>
          </div>
        </button>
      </div>
    </transition>

    <!-- ── Bouton toggle ────────────────────────────────── -->
    <button
      class="chatbot-toggle-btn"
      :class="{ 'chatbot-toggle-btn--open': chatOpen }"
      :style="{ background: chatOpen ? 'linear-gradient(135deg, #374151, #1f2937)' : `linear-gradient(135deg, ${primaryColor}, ${primaryColor})` }"
      @click="toggleChat"
      :title="chatOpen ? $t('common.close') : (selectedBot ? selectedBot.name : $t('chatbots.selectBot'))"
    >
      <Icon
        :icon="chatOpen ? 'mdi:close' : (selectedBot ? selectedBot.icon || 'mdi:robot-outline' : 'mdi:robot-outline')"
        class="h-6 w-6 text-white transition-transform duration-200"
      />
    </button>
  </div>
</template>

<script setup>
import { ref, watch, computed, onMounted, nextTick } from 'vue'
import { Icon } from '@iconify/vue'
import { chatbotsService } from '@/services/api'

const activeChatbots = ref([])
const selectedBot = ref(null)
const chatOpen = ref(false)
const isFullscreen = ref(false)
const headerHidden = ref(false)

const showAvatar = ref(false)
const showBubble = ref(false)
let bubbleDismissTimer = null
let messageObserver = null

const primaryColor = computed(() => {
  if (selectedBot.value?.color) return selectedBot.value.color
  if (activeChatbots.value[0]?.color) return activeChatbots.value[0].color
  return '#4f46e5'
})

const lighten = (hex) => {
  try {
    const n = parseInt(hex.replace('#', ''), 16)
    const r = Math.min(255, ((n >> 16) & 0xff) + 60)
    const g = Math.min(255, ((n >> 8) & 0xff) + 60)
    const b = Math.min(255, (n & 0xff) + 60)
    return `#${((1 << 24) | (r << 16) | (g << 8) | b).toString(16).slice(1)}`
  } catch {
    return hex
  }
}

const loadActiveChatbots = async () => {
  try {
    const bots = await chatbotsService.getActiveChatbots()
    activeChatbots.value = Array.isArray(bots) ? bots : []
    if (activeChatbots.value.length === 1) {
      selectedBot.value = activeChatbots.value[0]
    }
  } catch (error) {
    console.error('ChatbotWidget: impossible de charger les chatbots', error)
  }
}

const resetAvatarState = () => {
  showAvatar.value = false
  showBubble.value = false
  clearTimeout(bubbleDismissTimer)
  bubbleDismissTimer = null
}

const dismissBubble = () => {
  showBubble.value = false
  clearTimeout(bubbleDismissTimer)
  bubbleDismissTimer = null
}

const toggleChat = () => {
  chatOpen.value = !chatOpen.value
  if (!chatOpen.value) {
    isFullscreen.value = false
    resetAvatarState()
  }
}

const selectBot = (bot) => {
  selectedBot.value = bot
  headerHidden.value = false
  resetAvatarState()
}

const backToSelector = () => {
  selectedBot.value = null
  isFullscreen.value = false
  headerHidden.value = false
  resetAvatarState()
  stopObserver()
}

const stopObserver = () => {
  if (messageObserver) {
    messageObserver.disconnect()
    messageObserver = null
  }
}

const startMessageObserver = (containerId) => {
  stopObserver()
  const container = document.getElementById(containerId)
  if (!container) return

  messageObserver = new MutationObserver(() => {
    const userMsg = container.querySelector(
      '.chat-message--from-me, [data-author="user"], .chat-message-bubble--user'
    )
    if (userMsg) {
      headerHidden.value = true
      dismissBubble()
      stopObserver()
    }
  })

  messageObserver.observe(container, { childList: true, subtree: true })
}

const launchAvatar = (bot) => {
  const hasBubbleContent = bot.welcome_title || bot.welcome_subtitle
  setTimeout(() => {
    showAvatar.value = true
    if (hasBubbleContent) {
      showBubble.value = true
      bubbleDismissTimer = setTimeout(dismissBubble, 6000)
    }
  }, 350)
}

const mountChat = async (bot) => {
  if (!bot || !chatOpen.value) return
  await nextTick()

  const containerId = `n8n-chat-${bot.id}`
  const container = document.getElementById(containerId)
  if (!container) return

  // Conteneur déjà initialisé (v-show : DOM préservé entre ouvertures)
  // → on relance juste l'observateur et l'avatar si besoin
  if (container.children.length > 0) {
    if (!headerHidden.value && !bot.hide_header) startMessageObserver(containerId)
    if (bot.show_avatar_intro && !showAvatar.value) launchAvatar(bot)
    return
  }

  // Première ouverture pour ce bot → initialisation complète
  try {
    const { createChat } = await import('@n8n/chat')
    await import('@n8n/chat/style.css')

    let initialMessages = []
    try {
      initialMessages = JSON.parse(bot.initial_messages || '[]')
    } catch {
      initialMessages = []
    }

    const i18nConfig = {}
    if (bot.welcome_title || bot.welcome_subtitle) {
      i18nConfig.en = {
        ...(bot.welcome_title ? { title: bot.welcome_title } : {}),
        ...(bot.welcome_subtitle ? { subtitle: bot.welcome_subtitle } : {})
      }
    }

    createChat({
      webhookUrl: bot.webhook_url,
      mode: 'fullscreen',
      target: `#${containerId}`,
      initialMessages,
      showWelcomeScreen: false,
      loadPreviousSession: true,
      enableStreaming: false,
      ...(Object.keys(i18nConfig).length > 0 ? { i18n: i18nConfig } : {})
    })

    if (!bot.hide_header) startMessageObserver(containerId)
    if (bot.show_avatar_intro) launchAvatar(bot)
  } catch (error) {
    console.error("ChatbotWidget: impossible d'initialiser @n8n/chat", error)
  }
}

// Quand le bot change (sélection d'un nouveau bot) → le v-if recrée le DOM,
// on monte le chat dès que le panel est ouvert.
// Quand chatOpen passe à true pour le même bot → mountChat trouve le DOM
// déjà initialisé (container.children.length > 0) et ne recrée pas l'instance.
watch([selectedBot, chatOpen], ([bot, open], [oldBot]) => {
  if (bot && open) {
    // Réinitialiser headerHidden uniquement si le bot a changé
    if (bot !== oldBot) headerHidden.value = false
    mountChat(bot)
  } else {
    stopObserver()
    if (!open) resetAvatarState()
  }
})

onMounted(() => {
  loadActiveChatbots()
})
</script>

<style scoped>
/* Positionné à GAUCHE du ChatOverlay */
.chatbot-widget-root {
  position: fixed;
  bottom: 20px;
  right: 90px;
  z-index: 10000;
}

/* ── Bouton toggle ─────────────────────────────────────── */
.chatbot-toggle-btn {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.3);
  border: none;
  transition: box-shadow 0.2s, transform 0.2s;
  position: relative;
  z-index: 1;
}

.chatbot-toggle-btn:hover {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.4);
  transform: scale(1.05);
}

/* ── Panneau de chat ───────────────────────────────────── */
.chatbot-panel {
  position: absolute;
  bottom: 70px;
  right: 0;
  width: 380px;
  height: 520px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.25);
  display: flex;
  flex-direction: column;
  background: white;
  transition: width 0.25s ease, height 0.25s ease, bottom 0.25s ease,
              right 0.25s ease, border-radius 0.25s ease;
}

:global(.dark) .chatbot-panel {
  background: #111827;
}

.chatbot-panel--fullscreen {
  position: fixed;
  bottom: 0;
  right: 0;
  left: 0;
  top: 0;
  width: 100vw;
  height: 100vh;
  border-radius: 0;
  z-index: 10001;
}

.chatbot-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.625rem 0.75rem;
  flex-shrink: 0;
}

.chatbot-panel-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  color: white;
  background: rgba(255, 255, 255, 0.15);
  border: none;
  cursor: pointer;
  transition: background 0.15s;
}

.chatbot-panel-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* Zone de chat */
.chatbot-chat-container {
  flex: 1;
  overflow: hidden;
  min-height: 0;
  --chat--window--height: 100%;
}

/* ── Correctifs CSS @n8n/chat ──────────────────────────── */

/* La fenêtre n8n doit occuper exactement le conteneur */
:global([id^="n8n-chat-"] .chat-window) {
  height: 100% !important;
  border-radius: 0 !important;
  box-shadow: none !important;
  border: none !important;
}

/* Supprimer l'outline bleue au focus du champ de saisie */
:global([id^="n8n-chat-"] textarea),
:global([id^="n8n-chat-"] input[type="text"]) {
  outline: none !important;
  box-shadow: none !important;
}

:global([id^="n8n-chat-"] textarea:focus),
:global([id^="n8n-chat-"] input[type="text"]:focus) {
  outline: none !important;
  box-shadow: none !important;
  border-color: transparent !important;
}

/* Supprimer le cadre/ring autour de la zone de saisie */
:global([id^="n8n-chat-"] .chat-input),
:global([id^="n8n-chat-"] .chat-inputs),
:global([id^="n8n-chat-"] .chat-inputs-wrapper),
:global([id^="n8n-chat-"] .chat-input__wrapper),
:global([id^="n8n-chat-"] .chat-input-wrapper) {
  outline: none !important;
  box-shadow: none !important;
  border-left: none !important;
  border-right: none !important;
  border-bottom: none !important;
}

/* En-tête interne n8n — compact */
:global([id^="n8n-chat-"] .chat-header) {
  padding: 0.5rem 0.75rem !important;
  min-height: unset !important;
  transition: max-height 0.35s ease, opacity 0.35s ease, padding 0.35s ease;
  max-height: 80px;
  overflow: hidden;
}

:global([id^="n8n-chat-"] .chat-header p),
:global([id^="n8n-chat-"] .chat-header h1),
:global([id^="n8n-chat-"] .chat-header h2) {
  font-size: 0.75rem !important;
  line-height: 1.2 !important;
  margin: 0 !important;
}

.n8n-header-hidden :global(.chat-header) {
  max-height: 0 !important;
  opacity: 0 !important;
  padding-top: 0 !important;
  padding-bottom: 0 !important;
}

.n8n-hide-header :global(.chat-header) {
  display: none !important;
}

/* ══════════════════════════════════════════════════════════
   Zone avatar — EXTÉRIEURE au panneau, à sa gauche
   right: 395px = 380px (panel) + 15px (gap)
   ══════════════════════════════════════════════════════════ */
.chatbot-avatar-zone {
  position: absolute;
  bottom: 70px;
  right: 395px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  z-index: 9999;
}

.chatbot-avatar-face {
  position: relative;
  width: 88px;
  height: 88px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow:
    0 0 0 3px white,
    0 8px 32px rgba(0, 0, 0, 0.35);
  animation: avatar-float 3s ease-in-out infinite;
  flex-shrink: 0;
  overflow: hidden;
}

.chatbot-avatar-shine {
  position: absolute;
  top: 10px;
  left: 14px;
  width: 26px;
  height: 16px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.35);
  transform: rotate(-30deg);
  pointer-events: none;
}

.chatbot-avatar-icon {
  width: 48px;
  height: 48px;
  color: white;
  position: relative;
  z-index: 1;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.25));
}

.chatbot-avatar-label {
  font-size: 0.7rem;
  font-weight: 600;
  color: white;
  background: rgba(0, 0, 0, 0.55);
  padding: 0.2rem 0.65rem;
  border-radius: 99px;
  backdrop-filter: blur(6px);
  white-space: nowrap;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}

@keyframes avatar-float {
  0%, 100% { transform: translateY(0px); }
  50%       { transform: translateY(-8px); }
}

/* ── Bulle de bienvenue ───────────────────────────────── */
.chatbot-bubble {
  position: relative;
  background: white;
  border-radius: 14px 14px 14px 4px;
  padding: 0.75rem 2rem 0.75rem 1rem;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.16);
  max-width: 210px;
  cursor: pointer;
}

:global(.dark) .chatbot-bubble {
  background: #1f2937;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.5);
}

.chatbot-bubble-close {
  position: absolute;
  top: 6px;
  right: 6px;
  color: #9ca3af;
  background: none;
  border: none;
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: color 0.15s;
}

.chatbot-bubble-close:hover {
  color: #4b5563;
}

/* ── Transitions zone avatar ───────────────────────────── */
.avatar-zone-anim-enter-active {
  animation: avatar-zone-in 0.55s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}
.avatar-zone-anim-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}
.avatar-zone-anim-leave-to {
  opacity: 0;
  transform: translateY(24px) scale(0.7);
}

@keyframes avatar-zone-in {
  from { opacity: 0; transform: translateY(36px) scale(0.55); }
  to   { opacity: 1; transform: translateY(0) scale(1); }
}

/* ── Transitions bulle ────────────────────────────────── */
.bubble-anim-enter-active {
  animation: bubble-in 0.38s cubic-bezier(0.34, 1.56, 0.64, 1) 0.3s both;
}
.bubble-anim-leave-active {
  transition: opacity 0.22s ease, transform 0.22s ease;
}
.bubble-anim-leave-to {
  opacity: 0;
  transform: scale(0.88) translateY(6px);
}

@keyframes bubble-in {
  from { opacity: 0; transform: scale(0.7) translateY(10px); transform-origin: bottom center; }
  to   { opacity: 1; transform: scale(1) translateY(0); }
}

/* ── Sélecteur multi-bots ──────────────────────────────── */
.chatbot-selector-menu {
  position: absolute;
  bottom: 70px;
  right: 0;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 0.75rem;
  padding: 0.5rem;
  min-width: 16rem;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
}

:global(.dark) .chatbot-selector-menu {
  background: #1f2937;
  border-color: #374151;
}

.chatbot-selector-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  width: 100%;
  padding: 0.5rem 0.625rem;
  border-radius: 0.5rem;
  transition: background 0.15s;
  cursor: pointer;
  border: none;
  background: transparent;
}

.chatbot-selector-item:hover {
  background: #f3f4f6;
}

:global(.dark) .chatbot-selector-item:hover {
  background: #374151;
}

/* ── Transitions panel / menu ──────────────────────────── */
.chatbot-panel-anim-enter-active,
.chatbot-panel-anim-leave-active,
.chatbot-menu-anim-enter-active,
.chatbot-menu-anim-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}

.chatbot-panel-anim-enter-from,
.chatbot-panel-anim-leave-to,
.chatbot-menu-anim-enter-from,
.chatbot-menu-anim-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.97);
}

/* ══════════════════════════════════════════════════════════
   Responsive mobile (≤ 600px)
   Le panel passe en position: fixed pleine largeur.
   Le widget-root se rapproche du bord droit.
   ══════════════════════════════════════════════════════════ */
@media (max-width: 600px) {
  .chatbot-widget-root {
    right: 12px;
  }

  .chatbot-panel {
    /* Surcharge : fixed pleine largeur avec marges latérales */
    position: fixed !important;
    left: 8px !important;
    right: 8px !important;
    width: auto !important;
    bottom: 80px !important;
    height: 72vh !important;
    max-height: 520px;
    border-radius: 12px !important;
  }

  /* Fullscreen reste intact */
  .chatbot-panel--fullscreen {
    left: 0 !important;
    right: 0 !important;
    bottom: 0 !important;
    height: 100vh !important;
    max-height: none !important;
    border-radius: 0 !important;
  }

  .chatbot-selector-menu {
    position: fixed !important;
    left: 8px !important;
    right: 8px !important;
    bottom: 80px !important;
    max-height: 60vh;
    overflow-y: auto;
  }

  /* Avatar caché sur mobile (pas assez de place à gauche du panel) */
  .chatbot-avatar-zone {
    display: none;
  }
}
</style>
