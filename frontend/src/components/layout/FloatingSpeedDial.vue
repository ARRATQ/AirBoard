<template>
  <!-- Panneaux chat (positionnés fixed par leurs propres composants).
       Chargés au premier clic : leur code n'est pas téléchargé avant. -->
  <ChatbotWidget
    v-if="hasChatbots && botLoaded"
    ref="chatbotRef"
    :hide-toggle="true"
    :bots="bots"
    :start-open="true"
  />
  <ChatOverlay v-if="chatLoaded" :hide-toggle="true" />

  <!-- Speed Dial -->
  <div class="speed-dial-container">
    <!-- Items (glissent vers le haut au clic sur le FAB) -->
    <div class="speed-dial-items" :class="{ 'speed-dial-items--open': dialOpen }">
      <!-- Assistants IA -->
      <div v-if="hasChatbots" class="speed-dial-item" @click="openBot">
        <span class="speed-dial-label">Assistants IA</span>
        <button class="speed-dial-btn speed-dial-btn--indigo" :title="$t('chatbots.selectBot')">
          <Icon icon="mdi:robot-outline" class="h-5 w-5 text-white" />
        </button>
      </div>

      <!-- Messagerie -->
      <div class="speed-dial-item" @click="openChat">
        <span class="speed-dial-label">Messagerie</span>
        <button class="speed-dial-btn speed-dial-btn--blue" title="Messagerie équipe">
          <Icon icon="mdi:chat" class="h-5 w-5 text-white" />
          <span v-if="totalUnread > 0" class="speed-dial-badge">
            {{ totalUnread > 99 ? '99+' : totalUnread }}
          </span>
        </button>
      </div>
    </div>

    <!-- FAB principal -->
    <div
      class="speed-dial-fab-wrapper"
      :class="{ 'speed-dial-fab-wrapper--floating': !dialOpen }"
    >
      <!-- Halo externe pulsant -->
      <div
        v-if="!dialOpen"
        class="speed-dial-halo"
        :class="faceMode === 'ai' ? 'speed-dial-halo--ai' : 'speed-dial-halo--user'"
      />

      <button
        class="speed-dial-fab"
        :class="[
          { 'speed-dial-fab--open': dialOpen },
          !dialOpen ? (faceMode === 'ai' ? 'speed-dial-fab--ai' : 'speed-dial-fab--user') : ''
        ]"
        @click="toggleDial"
        :title="dialOpen ? $t('common.close') : 'Messages'"
      >
        <!-- Icône fermer -->
        <transition name="sd-face">
          <div v-if="dialOpen" key="close" class="sd-face-slot">
            <Icon icon="mdi:close" class="h-6 w-6 text-white" />
          </div>
        </transition>

        <!-- Visage utilisateur -->
        <transition name="sd-face">
          <div v-if="!dialOpen && faceMode === 'user'" key="user" class="sd-face-slot">
            <Icon icon="mdi:account-circle" class="text-white" style="font-size: 26px; margin-top: 2px;" />
            <span style="font-size: 7px; color: rgba(255,255,255,0.7); letter-spacing: 0.05em;">CHAT</span>
          </div>
        </transition>

        <!-- Visage IA -->
        <transition name="sd-face">
          <div v-if="!dialOpen && faceMode === 'ai'" key="ai" class="sd-face-slot">
            <Icon icon="mdi:robot-happy-outline" class="text-white" style="font-size: 26px; margin-top: 2px;" />
            <span style="font-size: 7px; color: rgba(255,255,255,0.7); letter-spacing: 0.05em;">AI</span>
          </div>
        </transition>

        <span v-if="!dialOpen && totalUnread > 0" class="speed-dial-fab-badge">
          {{ totalUnread > 99 ? '99+' : totalUnread }}
        </span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, defineAsyncComponent } from 'vue'
import { Icon } from '@iconify/vue'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import { chatbotsService } from '@/services/api'
import { playChatNotificationSound } from '@/utils/chatNotificationSound'

const ChatOverlay = defineAsyncComponent(() => import('@/components/chat/ChatOverlay.vue'))
const ChatbotWidget = defineAsyncComponent(() => import('@/components/ChatbotWidget.vue'))

const chatStore = useChatStore()
const authStore = useAuthStore()
const chatbotRef = ref(null)
const dialOpen = ref(false)
const bots = ref([])
const hasChatbots = computed(() => bots.value.length > 0)
const botLoaded = ref(false)
const chatLoaded = ref(chatStore.isOpen)
const faceMode = ref('user')
let faceInterval = null
let idleHandle = null

const totalUnread = computed(() =>
  Object.values(chatStore.unreadCounts).reduce((a, b) => a + b, 0)
)

// Le panneau de messagerie est chargé dès que le chat s'ouvre (clic ou ouverture
// d'une conversation depuis une autre page), puis reste monté.
watch(() => chatStore.isOpen, (open) => {
  if (open) chatLoaded.value = true
})

// Son à l'arrivée d'un message, même quand le panneau n'a jamais été ouvert.
watch(totalUnread, (newTotal, oldTotal) => {
  if (newTotal > oldTotal && newTotal > 0) playChatNotificationSound()
})

const toggleDial = () => {
  if (!hasChatbots.value) {
    openChat()
    return
  }
  dialOpen.value = !dialOpen.value
}

const openChat = () => {
  dialOpen.value = false
  chatStore.toggleChat()
}

const openBot = () => {
  dialOpen.value = false
  // Premier clic : le widget se charge puis s'ouvre lui-même (start-open)
  if (!botLoaded.value) {
    botLoaded.value = true
    return
  }
  if (chatbotRef.value?.isOpen) {
    chatbotRef.value?.closeChat()
  } else {
    chatbotRef.value?.openChat()
  }
}

// Connexion WebSocket (badge des messages non lus) une fois la page affichée.
const connectChatWhenIdle = () => {
  const connect = () => {
    idleHandle = null
    if (authStore.token) chatStore.connect()
  }
  if ('requestIdleCallback' in window) {
    idleHandle = window.requestIdleCallback(connect, { timeout: 3000 })
  } else {
    idleHandle = window.setTimeout(connect, 1500)
  }
}

onMounted(async () => {
  connectChatWhenIdle()

  faceInterval = setInterval(() => {
    if (!dialOpen.value) faceMode.value = faceMode.value === 'user' ? 'ai' : 'user'
  }, 3000)

  try {
    const list = await chatbotsService.getActiveChatbots()
    bots.value = Array.isArray(list) ? list : []
  } catch {
    bots.value = []
  }
})

onUnmounted(() => {
  if (faceInterval) clearInterval(faceInterval)
  if (idleHandle !== null) {
    if ('cancelIdleCallback' in window) window.cancelIdleCallback(idleHandle)
    else window.clearTimeout(idleHandle)
  }
})
</script>

<style scoped>
/* ── Conteneur principal ────────────────────────────────── */
.speed-dial-container {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 10001;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 12px;
}

/* ── Items wrapper ──────────────────────────────────────── */
.speed-dial-items {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-end;
  pointer-events: none; /* évite de bloquer les clics sur les panneaux en dessous */
}

.speed-dial-items--open {
  pointer-events: auto;
}

/* ── Chaque item ────────────────────────────────────────── */
.speed-dial-item {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  opacity: 0;
  transform: translateY(16px) scale(0.85);
  pointer-events: none;
  transition: opacity 0.22s ease, transform 0.28s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* Animations décalées — l'item le plus proche du FAB apparaît en premier */
.speed-dial-items--open .speed-dial-item:last-child {
  transition-delay: 0ms;
}
.speed-dial-items--open .speed-dial-item:first-child {
  transition-delay: 65ms;
}

.speed-dial-items--open .speed-dial-item {
  opacity: 1;
  transform: none;
  pointer-events: auto;
}

/* ── Labels ─────────────────────────────────────────────── */
.speed-dial-label {
  background: var(--bg-card);
  color: var(--text-primary);
  padding: 5px 12px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  box-shadow: 0 2px 10px rgba(var(--shadow-rgb), 0.2);
  user-select: none;
}

/* ── Boutons secondaires ────────────────────────────────── */
.speed-dial-btn {
  position: relative;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 3px 12px rgba(0, 0, 0, 0.25);
  transition: transform 0.15s, box-shadow 0.15s;
  flex-shrink: 0;
}

.speed-dial-btn:hover {
  transform: scale(1.1);
  box-shadow: 0 5px 16px rgba(0, 0, 0, 0.3);
}

.speed-dial-btn--blue {
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
}

.speed-dial-btn--indigo {
  background: linear-gradient(135deg, #7c3aed, #4f46e5);
}

/* ── Badge non-lus sur bouton secondaire ────────────────── */
.speed-dial-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  background: #ef4444;
  color: white;
  font-size: 10px;
  font-weight: 700;
  border-radius: 99px;
  min-width: 18px;
  height: 18px;
  padding: 0 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid white;
}

/* ── FAB wrapper flottant ───────────────────────────────── */
@keyframes sd-float {
  0%, 100% { transform: translateY(0px); }
  50%       { transform: translateY(-9px); }
}

@keyframes sd-halo-expand {
  0%   { transform: scale(1);   opacity: 0.5; }
  100% { transform: scale(2);   opacity: 0; }
}

.speed-dial-fab-wrapper {
  position: relative;
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.speed-dial-fab-wrapper--floating {
  animation: sd-float 3.2s ease-in-out infinite;
}

.speed-dial-fab-wrapper--floating:hover {
  animation-play-state: paused;
}

/* Halo externe */
.speed-dial-halo {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  animation: sd-halo-expand 3.2s ease-out infinite;
  pointer-events: none;
}
.speed-dial-halo--user { background: rgba(99, 102, 241, 0.45); }
.speed-dial-halo--ai   { background: rgba(168, 85, 247, 0.45); }

/* ── FAB principal ──────────────────────────────────────── */
.speed-dial-fab {
  position: relative;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  transition: box-shadow 0.2s, transform 0.2s, background 0.7s ease;
}

.speed-dial-fab--user {
  background: linear-gradient(135deg, #2563eb, #4f46e5);
  box-shadow: 0 4px 16px rgba(37, 99, 235, 0.5), 0 0 0 3px rgba(99, 102, 241, 0.2);
}

.speed-dial-fab--ai {
  background: linear-gradient(135deg, #7c3aed, #a855f7);
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.5), 0 0 0 3px rgba(168, 85, 247, 0.2);
}

.speed-dial-fab--open {
  background: linear-gradient(135deg, #374151, #1f2937);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
}

.speed-dial-fab:hover {
  transform: scale(1.07);
}

/* Slot de visage */
.sd-face-slot {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

/* ── Transitions de visage ──────────────────────────────── */
.sd-face-enter-active {
  transition: opacity 0.35s ease, transform 0.35s cubic-bezier(0.34, 1.3, 0.64, 1);
}
.sd-face-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}
.sd-face-enter-from { opacity: 0; transform: scale(0.55) rotate(-20deg); }
.sd-face-leave-to   { opacity: 0; transform: scale(0.55) rotate(20deg); }

/* ── Badge non-lus sur FAB ──────────────────────────────── */
.speed-dial-fab-badge {
  position: absolute;
  top: -2px;
  right: -2px;
  background: #ef4444;
  color: white;
  font-size: 10px;
  font-weight: 700;
  border-radius: 99px;
  min-width: 20px;
  height: 20px;
  padding: 0 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid white;
  animation: badge-bounce 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes badge-bounce {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.15); }
}

/* ── Mobile (≤ 600px) ───────────────────────────────────── */
@media (max-width: 600px) {
  .speed-dial-container {
    bottom: 16px;
    right: 16px;
  }

  .speed-dial-label {
    font-size: 12px;
    padding: 4px 10px;
  }
}
</style>
