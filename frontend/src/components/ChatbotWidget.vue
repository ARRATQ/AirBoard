<template>
  <div v-if="activeChatbots.length > 0" class="chatbot-widget-root">

    <!-- Panneau de chat -->
    <transition name="chatbot-panel-anim">
      <div
        v-if="chatOpen && selectedBot"
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
            <!-- Changer de bot -->
            <button
              v-if="activeChatbots.length > 1"
              @click="backToSelector"
              class="chatbot-panel-btn"
              :title="$t('chatbots.switchBot')"
            >
              <Icon icon="mdi:swap-horizontal" class="h-4 w-4" />
            </button>
            <!-- Plein écran / Réduire -->
            <button
              @click="isFullscreen = !isFullscreen"
              class="chatbot-panel-btn"
              :title="isFullscreen ? $t('chatbots.exitFullscreen') : $t('chatbots.fullscreen')"
            >
              <Icon :icon="isFullscreen ? 'mdi:fullscreen-exit' : 'mdi:fullscreen'" class="h-4 w-4" />
            </button>
            <!-- Fermer -->
            <button @click="chatOpen = false" class="chatbot-panel-btn" :title="$t('common.close')">
              <Icon icon="mdi:close" class="h-4 w-4" />
            </button>
          </div>
        </div>

        <!-- Conteneur @n8n/chat — :key force réinitialisation si bot change -->
        <div
          :key="selectedBot.id"
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

    <!-- Sélecteur multi-bots -->
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

    <!-- Bouton toggle -->
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
// true dès que le premier message utilisateur est envoyé → masque l'en-tête n8n
const headerHidden = ref(false)
let messageObserver = null

// Couleur principale (du bot sélectionné ou premier bot ou défaut)
const primaryColor = computed(() => {
  if (selectedBot.value?.color) return selectedBot.value.color
  if (activeChatbots.value[0]?.color) return activeChatbots.value[0].color
  return '#4f46e5'
})

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

const toggleChat = () => {
  chatOpen.value = !chatOpen.value
  // Quitter le plein écran quand on ferme
  if (!chatOpen.value) isFullscreen.value = false
}

const selectBot = (bot) => {
  selectedBot.value = bot
  headerHidden.value = false
}

const backToSelector = () => {
  selectedBot.value = null
  isFullscreen.value = false
  headerHidden.value = false
  stopObserver()
}

const stopObserver = () => {
  if (messageObserver) {
    messageObserver.disconnect()
    messageObserver = null
  }
}

/**
 * Observe le conteneur n8n pour détecter le premier message utilisateur
 * et masquer l'en-tête n8n avec une transition douce.
 */
const startMessageObserver = (containerId) => {
  stopObserver()
  const container = document.getElementById(containerId)
  if (!container) return

  messageObserver = new MutationObserver(() => {
    // Les messages utilisateur ont la classe "chat-message--from-me" dans @n8n/chat
    const userMsg = container.querySelector(
      '.chat-message--from-me, [data-author="user"], .chat-message-bubble--user'
    )
    if (userMsg) {
      headerHidden.value = true
      stopObserver()
    }
  })

  messageObserver.observe(container, { childList: true, subtree: true })
}

const mountChat = async (bot) => {
  if (!bot || !chatOpen.value) return
  await nextTick()

  const containerId = `n8n-chat-${bot.id}`
  const container = document.getElementById(containerId)
  if (!container) return

  // Éviter la double initialisation (ex: fermer/rouvrir sans changer de bot)
  if (container.children.length > 0) {
    // Relancer l'observateur si la session précédente n'avait pas encore de message
    if (!headerHidden.value && !bot.hide_header) startMessageObserver(containerId)
    return
  }

  try {
    const { createChat } = await import('@n8n/chat')
    await import('@n8n/chat/style.css')

    let initialMessages = []
    try {
      initialMessages = JSON.parse(bot.initial_messages || '[]')
    } catch {
      initialMessages = []
    }

    // i18n personnalisé (titre + sous-titre de la fenêtre @n8n/chat)
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

    // Lancer l'observateur seulement si le header n8n est visible
    if (!bot.hide_header) {
      startMessageObserver(containerId)
    }
  } catch (error) {
    console.error("ChatbotWidget: impossible d'initialiser @n8n/chat", error)
  }
}

watch([selectedBot, chatOpen], ([bot, open]) => {
  if (bot && open) {
    headerHidden.value = false
    mountChat(bot)
  } else {
    stopObserver()
  }
})

onMounted(() => {
  loadActiveChatbots()
})
</script>

<style scoped>
/* Positionné à GAUCHE du ChatOverlay (right: 20px + 56px + 14px = 90px) */
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

/* ── Panneau de chat (mode normal) ────────────────────── */
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
  transition: width 0.25s ease, height 0.25s ease, bottom 0.25s ease, right 0.25s ease, border-radius 0.25s ease;
}

/* ── Mode plein écran ─────────────────────────────────── */
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

/* En-tête du panneau */
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
  --chat--window--height: 100%;
}

:global(.dark) .chatbot-panel {
  background: #111827;
}

/* ── Réduction de l'en-tête interne @n8n/chat ─────────── */
/*
  @n8n/chat rend son propre header (titre + sous-titre) au-dessus des messages.
  On le rend compact par défaut (on a déjà notre propre en-tête coloré).
*/
:global(#n8n-chat-root .chat-header),
:global([id^="n8n-chat-"] .chat-header) {
  padding: 0.5rem 0.75rem !important;
  min-height: unset !important;
}

:global([id^="n8n-chat-"] .chat-header p),
:global([id^="n8n-chat-"] .chat-header h1),
:global([id^="n8n-chat-"] .chat-header h2) {
  font-size: 0.75rem !important;
  line-height: 1.2 !important;
  margin: 0 !important;
}

/* Transition douce pour masquer l'en-tête n8n au premier message */
:global([id^="n8n-chat-"] .chat-header) {
  transition: max-height 0.35s ease, opacity 0.35s ease, padding 0.35s ease;
  max-height: 80px;
  opacity: 1;
  overflow: hidden;
}

/* État masqué (classe ajoutée dynamiquement) */
.n8n-header-hidden :global(.chat-header) {
  max-height: 0 !important;
  opacity: 0 !important;
  padding-top: 0 !important;
  padding-bottom: 0 !important;
}

/* Option admin : masquer le header n8n dès le départ */
.n8n-hide-header :global(.chat-header) {
  display: none !important;
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

/* ── Transitions ───────────────────────────────────────── */
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
</style>
