<template>
  <!-- Minimized Bubble animée -->
  <div
    v-if="!chatStore.isOpen && !props.hideToggle"
    class="fixed z-[9999] cursor-pointer co-fab"
    :style="{ bottom: '20px', right: '20px' }"
    @click="chatStore.toggleChat()"
  >
    <!-- Halo externe pulsant -->
    <div class="co-fab-halo" :class="faceMode === 'ai' ? 'co-fab-halo--ai' : 'co-fab-halo--user'" />

    <div
      class="w-14 h-14 rounded-full flex items-center justify-center relative co-fab-ring"
      :class="[faceMode === 'ai' ? 'co-fab-ring--ai' : 'co-fab-ring--user', totalUnread > 0 ? 'co-fab-unread-pulse' : '']"
    >
      <!-- Visage utilisateur -->
      <transition name="co-face">
        <div v-if="faceMode === 'user'" key="user" class="absolute inset-0 flex flex-col items-center justify-center gap-0">
          <Icon icon="mdi:account-circle" class="text-white" style="font-size: 28px; margin-top: 2px;" />
          <span class="text-white/70 leading-none" style="font-size: 8px; letter-spacing: 0.04em;">CHAT</span>
        </div>
      </transition>

      <!-- Visage IA -->
      <transition name="co-face">
        <div v-if="faceMode === 'ai'" key="ai" class="absolute inset-0 flex flex-col items-center justify-center">
          <Icon icon="mdi:robot-happy-outline" class="text-white" style="font-size: 28px; margin-top: 2px;" />
          <span class="text-white/70 leading-none" style="font-size: 8px; letter-spacing: 0.04em;">AI</span>
        </div>
      </transition>

      <!-- Badge non-lus -->
      <span
        v-if="totalUnread > 0"
        class="absolute -top-1 -right-1 bg-red-500 text-white text-[10px] font-bold rounded-full min-w-[19px] h-[19px] px-1 flex items-center justify-center border-2 border-white z-10"
      >
        {{ totalUnread > 99 ? '99+' : totalUnread }}
      </span>
    </div>
  </div>

  <!-- Chat Window -->
  <transition name="co-window-anim">
    <div
      v-if="chatStore.isOpen"
      class="co-window flex flex-col overflow-hidden bg-white dark:bg-gray-900"
      :class="isFullscreen ? 'co-window--fullscreen' : 'co-window--normal'"
    >
      <!-- ── Header ─────────────────────────────────────── -->
      <div
        class="flex items-center justify-between px-3.5 py-2.5 flex-shrink-0"
        style="background: linear-gradient(135deg, #2563eb, #4f46e5);"
      >
        <div class="flex items-center gap-2.5 min-w-0">
          <div class="w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0" style="background:rgba(255,255,255,0.2)">
            <Icon icon="mdi:chat" class="h-4 w-4 text-white" />
          </div>
          <div class="min-w-0">
            <p class="font-semibold text-white text-sm leading-tight">Messagerie</p>
            <p class="text-white/60 text-xs leading-tight flex items-center gap-1">
              <span class="inline-block w-1.5 h-1.5 rounded-full" :class="chatStore.isConnected ? 'bg-green-400' : 'bg-red-400'" />
              {{ chatStore.isConnected ? 'Connecté' : 'Hors ligne' }}
            </p>
          </div>
        </div>
        <div class="flex items-center gap-1 flex-shrink-0">
          <button
            class="w-7 h-7 rounded-full flex items-center justify-center text-white transition-colors"
            style="background:rgba(255,255,255,0.15)"
            @mouseover="$event.currentTarget.style.background='rgba(255,255,255,0.28)'"
            @mouseleave="$event.currentTarget.style.background='rgba(255,255,255,0.15)'"
            @click="isFullscreen = !isFullscreen"
            :title="isFullscreen ? 'Réduire' : 'Plein écran'"
          >
            <Icon :icon="isFullscreen ? 'mdi:fullscreen-exit' : 'mdi:fullscreen'" class="h-4 w-4" />
          </button>
          <button
            class="w-7 h-7 rounded-full flex items-center justify-center text-white transition-colors"
            style="background:rgba(255,255,255,0.15)"
            @mouseover="$event.currentTarget.style.background='rgba(255,255,255,0.28)'"
            @mouseleave="$event.currentTarget.style.background='rgba(255,255,255,0.15)'"
            @click="chatStore.toggleChat()"
            title="Fermer"
          >
            <Icon icon="mdi:close" class="h-4 w-4" />
          </button>
        </div>
      </div>

      <!-- ── Body ──────────────────────────────────────── -->
      <div class="flex flex-1 overflow-hidden min-h-0">

        <!-- Sidebar contacts -->
        <div
          v-if="!isMobile || !chatStore.activeConversation"
          class="flex flex-col border-r border-gray-100 dark:border-gray-700/60 flex-shrink-0 min-h-0"
          :class="isFullscreen ? 'w-72' : 'w-full md:w-[220px]'"
          style="background:#fafafa"
          :style="{ background: isDark ? '#1f2937' : '#fafafa' }"
        >
          <!-- Search -->
          <div class="px-2.5 py-2 border-b border-gray-100 dark:border-gray-700/60 flex-shrink-0">
            <div class="relative">
              <Icon icon="mdi:magnify" class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400 text-sm" />
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Rechercher..."
                class="w-full pl-8 pr-7 py-1.5 rounded-lg text-sm border-none outline-none bg-gray-100 dark:bg-gray-700/60 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-400/30"
              >
              <button v-if="searchQuery" @click="searchQuery = ''" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600">
                <Icon icon="mdi:close-circle" class="text-xs" />
              </button>
            </div>
          </div>

          <!-- Contact list -->
          <div class="overflow-y-auto flex-1 py-1.5 px-1.5">

            <!-- Non-lus -->
            <template v-if="hasUnreadUsers">
              <p class="text-[10px] font-bold text-blue-500 uppercase tracking-wide px-1.5 pt-1 pb-0.5">Non lus</p>
              <div
                v-for="user in unreadUsers" :key="'u_'+user.id"
                class="flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer transition-colors mb-0.5"
                :class="chatStore.activeConversation?.id === user.id && chatStore.activeConversation?.type === 'user'
                  ? 'bg-blue-50 dark:bg-blue-900/20'
                  : 'hover:bg-gray-200/60 dark:hover:bg-gray-700/50'"
                @click="chatStore.openConversation('user', user)"
              >
                <div class="relative flex-shrink-0">
                  <img :src="user.avatar_url || `https://ui-avatars.com/api/?name=${user.first_name}+${user.last_name}&background=3b82f6&color=fff`" class="w-8 h-8 rounded-full object-cover" />
                  <span v-if="user.is_online" class="absolute bottom-0 right-0 w-2 h-2 bg-green-400 border-2 border-white dark:border-gray-800 rounded-full" />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="text-xs font-semibold text-gray-800 dark:text-gray-100 truncate">{{ user.first_name }} {{ user.last_name }}</div>
                  <div class="text-[10px] text-gray-400 truncate">{{ user.job_title || 'Membre' }}</div>
                </div>
                <span class="bg-blue-500 text-white text-[10px] font-bold rounded-full min-w-[16px] h-4 px-1 flex items-center justify-center flex-shrink-0">{{ chatStore.unreadCounts['user_'+user.id] }}</span>
              </div>
            </template>

            <!-- Groupes -->
            <template v-if="filteredGroups.length > 0">
              <p class="text-[10px] font-bold text-gray-400 uppercase tracking-wide px-1.5 pt-2 pb-0.5">Groupes</p>
              <div
                v-for="group in filteredGroups" :key="group.id"
                class="flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer transition-colors mb-0.5"
                :class="chatStore.activeConversation?.id === group.id && chatStore.activeConversation?.type === 'group'
                  ? 'bg-blue-50 dark:bg-blue-900/20'
                  : 'hover:bg-gray-200/60 dark:hover:bg-gray-700/50'"
                @click="chatStore.openConversation('group', group)"
              >
                <div class="w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0" style="background:linear-gradient(135deg,#7c3aed,#4f46e5)">
                  <Icon icon="mdi:account-group" class="h-4 w-4 text-white" />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="text-xs font-semibold text-gray-800 dark:text-gray-100 truncate">{{ group.name }}</div>
                  <div class="text-[10px] text-gray-400">Groupe</div>
                </div>
                <span v-if="chatStore.unreadCounts['group_'+group.id]" class="bg-blue-500 text-white text-[10px] font-bold rounded-full min-w-[16px] h-4 px-1 flex items-center justify-center">{{ chatStore.unreadCounts['group_'+group.id] }}</span>
              </div>
            </template>

            <!-- Users par groupe -->
            <template v-for="(groupData, groupName) in groupedUsers" :key="groupName">
              <p class="text-[10px] font-bold text-gray-400 uppercase tracking-wide px-1.5 pt-2 pb-0.5">{{ groupName }}</p>
              <div
                v-for="user in groupData.users" :key="user.id"
                class="flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer transition-colors mb-0.5"
                :class="chatStore.activeConversation?.id === user.id && chatStore.activeConversation?.type === 'user'
                  ? 'bg-blue-50 dark:bg-blue-900/20'
                  : 'hover:bg-gray-200/60 dark:hover:bg-gray-700/50'"
                @click="chatStore.openConversation('user', user)"
              >
                <div class="relative flex-shrink-0">
                  <img :src="user.avatar_url || `https://ui-avatars.com/api/?name=${user.first_name}+${user.last_name}&background=3b82f6&color=fff`" class="w-8 h-8 rounded-full object-cover" />
                  <span v-if="user.is_online" class="absolute bottom-0 right-0 w-2 h-2 bg-green-400 border-2 border-white dark:border-gray-800 rounded-full" />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="text-xs font-semibold text-gray-800 dark:text-gray-100 truncate">{{ user.first_name }} {{ user.last_name }}</div>
                  <div class="text-[10px] text-gray-400 truncate">{{ user.job_title || 'Membre' }}</div>
                </div>
                <span v-if="chatStore.unreadCounts['user_'+user.id]" class="bg-blue-500 text-white text-[10px] font-bold rounded-full min-w-[16px] h-4 px-1 flex items-center justify-center">{{ chatStore.unreadCounts['user_'+user.id] }}</span>
              </div>
            </template>

            <!-- Autres -->
            <template v-if="otherUsers.length > 0">
              <p class="text-[10px] font-bold text-gray-400 uppercase tracking-wide px-1.5 pt-2 pb-0.5">Contacts</p>
              <div
                v-for="user in otherUsers" :key="'o_'+user.id"
                class="flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer transition-colors mb-0.5"
                :class="chatStore.activeConversation?.id === user.id && chatStore.activeConversation?.type === 'user'
                  ? 'bg-blue-50 dark:bg-blue-900/20'
                  : 'hover:bg-gray-200/60 dark:hover:bg-gray-700/50'"
                @click="chatStore.openConversation('user', user)"
              >
                <div class="relative flex-shrink-0">
                  <img :src="user.avatar_url || `https://ui-avatars.com/api/?name=${user.first_name}+${user.last_name}&background=3b82f6&color=fff`" class="w-8 h-8 rounded-full object-cover" />
                  <span v-if="user.is_online" class="absolute bottom-0 right-0 w-2 h-2 bg-green-400 border-2 border-white dark:border-gray-800 rounded-full" />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="text-xs font-semibold text-gray-800 dark:text-gray-100 truncate">{{ user.first_name }} {{ user.last_name }}</div>
                  <div class="text-[10px] text-gray-400 truncate">{{ user.job_title || 'Membre' }}</div>
                </div>
                <span v-if="chatStore.unreadCounts['user_'+user.id]" class="bg-blue-500 text-white text-[10px] font-bold rounded-full min-w-[16px] h-4 px-1 flex items-center justify-center">{{ chatStore.unreadCounts['user_'+user.id] }}</span>
              </div>
            </template>

            <!-- Vide -->
            <div
              v-if="!hasUnreadUsers && filteredGroups.length === 0 && Object.keys(groupedUsers).length === 0 && otherUsers.length === 0"
              class="flex flex-col items-center justify-center py-10 text-gray-400"
            >
              <Icon icon="mdi:account-search-outline" class="text-3xl opacity-30 mb-1" />
              <p class="text-xs">Aucun contact trouvé</p>
            </div>
          </div>
        </div>

        <!-- ── Vue conversation ──────────────────────────── -->
        <div v-if="chatStore.activeConversation" class="flex-1 flex flex-col min-h-0 bg-white dark:bg-gray-900">

          <!-- Conversation header -->
          <div class="flex items-center justify-between px-3 py-2 border-b border-gray-100 dark:border-gray-700/60 flex-shrink-0 bg-white dark:bg-gray-900">
            <div class="flex items-center gap-2 min-w-0">
              <button v-if="isMobile" @click="chatStore.closeConversation()" class="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 mr-0.5 flex-shrink-0 transition-colors">
                <Icon icon="mdi:arrow-left" class="text-xl" />
              </button>
              <div class="relative flex-shrink-0">
                <img
                  v-if="chatStore.activeConversation.type === 'user'"
                  :src="chatStore.activeConversation.avatar || `https://ui-avatars.com/api/?name=${chatStore.activeConversation.name}&background=3b82f6&color=fff`"
                  class="w-9 h-9 rounded-full object-cover"
                />
                <div v-else class="w-9 h-9 rounded-full flex items-center justify-center" style="background:linear-gradient(135deg,#7c3aed,#4f46e5)">
                  <Icon icon="mdi:account-group" class="h-5 w-5 text-white" />
                </div>
                <span v-if="chatStore.activeConversation.type === 'user' && chatStore.activeConversation.is_online" class="absolute bottom-0 right-0 w-2.5 h-2.5 bg-green-400 border-2 border-white dark:border-gray-900 rounded-full" />
              </div>
              <div class="min-w-0">
                <p class="text-sm font-semibold text-gray-900 dark:text-white truncate leading-tight">{{ chatStore.activeConversation.name }}</p>
                <p class="text-xs leading-tight" :class="chatStore.activeConversation.is_online ? 'text-green-500' : 'text-gray-400'">
                  {{ chatStore.activeConversation.type === 'group' ? 'Groupe' : (chatStore.activeConversation.is_online ? 'En ligne' : 'Hors ligne') }}
                </p>
              </div>
            </div>

            <!-- Menu ⋮ -->
            <div class="relative flex-shrink-0">
              <button @click.stop="toggleMenu" class="w-8 h-8 rounded-lg flex items-center justify-center text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-700 dark:hover:text-gray-300 transition-colors">
                <Icon icon="mdi:dots-vertical" class="text-lg" />
              </button>
              <transition name="co-menu-drop">
                <div v-if="showMenu" class="absolute right-0 top-full mt-1.5 bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700 rounded-xl shadow-xl w-52 py-1.5 z-30">
                  <button @click="toggleNotifications" class="w-full flex items-center gap-2 px-3 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors">
                    <Icon :icon="notificationsEnabled ? 'mdi:bell-ring-outline' : 'mdi:bell-off-outline'" :class="notificationsEnabled ? 'text-green-500' : 'text-gray-400'" class="flex-shrink-0" />
                    <span class="flex-1 text-xs text-left">Sons de notification</span>
                    <span class="text-[9px] font-bold uppercase px-1.5 py-0.5 rounded-md" :class="notificationsEnabled ? 'bg-green-50 text-green-600 dark:bg-green-900/20' : 'bg-gray-100 text-gray-500 dark:bg-gray-700'">{{ notificationsEnabled ? 'On' : 'Off' }}</span>
                  </button>
                  <div class="border-t border-gray-100 dark:border-gray-700 my-1" />
                  <button @click="confirmClearHistory" class="w-full flex items-center gap-2 px-3 py-2 text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors">
                    <Icon icon="mdi:delete-sweep-outline" class="flex-shrink-0" />
                    <span class="text-xs">Vider l'historique</span>
                  </button>
                </div>
              </transition>
            </div>
          </div>

          <!-- Zone messages -->
          <div
            ref="messagesContainer"
            class="flex-1 overflow-y-auto flex flex-col gap-2.5 px-3.5 py-3 min-h-0"
            style="background:#f8faff"
            @click="showMenu = false"
          >
            <!-- Empty state -->
            <div v-if="currentMessages.length === 0" class="flex flex-col items-center justify-center py-10 text-gray-400">
              <div class="w-12 h-12 rounded-xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-2">
                <Icon icon="mdi:chat-outline" class="text-2xl opacity-40" />
              </div>
              <p class="text-xs font-medium">Début de la conversation</p>
              <p class="text-[11px] mt-0.5 opacity-60">Envoyez un premier message !</p>
            </div>

            <!-- Messages -->
            <div
              v-for="msg in currentMessages" :key="msg.id"
              class="flex flex-col"
              :class="msg.sender_id == myId ? 'items-end' : 'items-start'"
            >
              <!-- Nom expéditeur (groupes, messages reçus) -->
              <span
                v-if="chatStore.activeConversation?.type === 'group' && msg.sender_id != myId"
                class="text-[10px] font-semibold text-gray-400 mb-1 ml-8"
              >
                {{ msg.sender?.first_name }} {{ msg.sender?.last_name || msg.sender?.username }}
              </span>

              <div class="flex items-end gap-1.5 flex-nowrap" :class="msg.sender_id == myId ? 'flex-row-reverse' : 'flex-row'">
                <!-- Avatar (messages reçus) -->
                <img
                  v-if="msg.sender_id != myId"
                  :src="msg.sender?.avatar_url || `https://ui-avatars.com/api/?name=${msg.sender?.first_name}+${msg.sender?.last_name}&background=6366f1&color=fff&size=64`"
                  class="w-6 h-6 rounded-full object-cover flex-shrink-0 mb-0.5"
                />

                <!-- Bulle -->
                <div class="co-msg-wrap group">
                  <div
                    class="co-msg-bubble"
                    :class="msg.sender_id == myId ? 'co-msg-bubble--me' : 'co-msg-bubble--other'"
                  >
                    <span class="co-msg-text" v-html="formatMessage(msg.content)" />
                    <button
                      v-if="msg.sender_id == myId"
                      @click.stop="deleteMessage(msg.id)"
                      class="co-del-btn"
                      title="Supprimer"
                    >
                      <Icon icon="mdi:delete-outline" class="h-3 w-3" />
                    </button>
                  </div>
                  <span class="text-[10px] text-gray-400 mt-0.5 block" :class="msg.sender_id == myId ? 'text-right' : 'text-left'">
                    {{ formatTime(msg.created_at) }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Zone saisie -->
          <div class="flex-shrink-0 border-t border-gray-100 dark:border-gray-700/60 bg-white dark:bg-gray-900">
            <!-- Emoji picker -->
            <transition name="co-emoji-pop">
              <div v-if="showEmojiPicker" class="border-b border-gray-100 dark:border-gray-700/60 bg-white dark:bg-gray-900 p-2">
                <div class="grid grid-cols-8 gap-0.5 max-h-32 overflow-y-auto">
                  <button
                    v-for="emoji in commonEmojis" :key="emoji"
                    @click="addEmoji(emoji)"
                    class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-lg flex items-center justify-center transition-transform hover:scale-125"
                  >{{ emoji }}</button>
                </div>
              </div>
            </transition>

            <div class="flex items-center gap-2 px-3 py-2.5">
              <button
                @click="showEmojiPicker = !showEmojiPicker"
                class="flex-shrink-0 p-1.5 rounded-lg transition-colors"
                :class="showEmojiPicker ? 'text-blue-500 bg-blue-50 dark:bg-blue-900/20' : 'text-gray-400 hover:text-blue-500 hover:bg-gray-100 dark:hover:bg-gray-800'"
                title="Emoji"
              >
                <Icon icon="mdi:emoticon-outline" class="h-5 w-5" />
              </button>

              <input
                v-model="newMessage"
                @keyup.enter="sendMessage"
                @focus="showEmojiPicker = false"
                type="text"
                placeholder="Écrire un message..."
                class="flex-1 min-w-0 border border-gray-200 dark:border-gray-700 rounded-full px-4 py-2 text-sm bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-400 outline-none focus:border-blue-400 focus:ring-2 focus:ring-blue-400/20 focus:bg-white dark:focus:bg-gray-800 transition-all"
              >

              <button
                @click="sendMessage"
                :disabled="!newMessage.trim()"
                class="flex-shrink-0 w-9 h-9 rounded-full flex items-center justify-center text-white transition-all disabled:opacity-40 disabled:cursor-not-allowed hover:scale-105"
                style="background:linear-gradient(135deg,#2563eb,#3b82f6); box-shadow:0 2px 8px rgba(37,99,235,0.35)"
                title="Envoyer"
              >
                <Icon icon="mdi:send" class="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>

        <!-- Empty state desktop -->
        <div
          v-else-if="!isMobile"
          class="flex-1 flex flex-col items-center justify-center text-gray-400 dark:text-gray-600"
          style="background:#f8faff"
        >
          <div class="w-14 h-14 rounded-2xl flex items-center justify-center mb-3" style="background:linear-gradient(135deg,#dbeafe,#e0e7ff)">
            <Icon icon="mdi:chat-outline" class="text-3xl text-blue-400" />
          </div>
          <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Sélectionnez une conversation</p>
          <p class="text-xs mt-1 text-gray-400">Choisissez un contact à gauche</p>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue';
import { useChatStore } from '@/stores/chat';
import { useAuthStore } from '@/stores/auth';
import { storeToRefs } from 'pinia';
import { Icon } from '@iconify/vue';
import DOMPurify from 'dompurify';

const props = defineProps({
  hideToggle: { type: Boolean, default: false }
});

const chatStore = useChatStore();
const authStore = useAuthStore();
const { activeConversation, messages } = storeToRefs(chatStore);

const newMessage = ref('');
const searchQuery = ref('');
const showEmojiPicker = ref(false);
const messagesContainer = ref(null);
const isMobile = ref(window.innerWidth < 768);
const showMenu = ref(false);
const isFullscreen = ref(false);
const isDark = ref(document.documentElement.classList.contains('dark'));

const commonEmojis = [
  '😀','😂','🤣','😊','😍','😘','😉','😎',
  '🤔','🤨','😐','🙄','😏','😮','😴','😌',
  '😛','😜','🤤','😒','😓','😔','😕','🙃',
  '😲','☹️','😤','😢','😭','🤯','😬','😰',
  '😱','🥵','🥶','😳','🤪','😵','😡','😠',
  '😇','🥳','🥴','🥺','🤫','🧐','🤓','😈',
  '💀','👻','👽','🤖','💩','😺','😸','😻',
  '👍','👎','👌','✌️','🤞','🤝','👏','🙌',
  '🔥','✨','⭐','❤️','💯','🎉','🎁','🚀'
];

const addEmoji = (emoji) => { newMessage.value += emoji; };
const myId = computed(() => authStore.user?.id);

const filteredGroups = computed(() => {
  if (!searchQuery.value) return chatStore.contacts.groups;
  const q = searchQuery.value.toLowerCase();
  return chatStore.contacts.groups.filter(g => g.name.toLowerCase().includes(q));
});

const filteredUsers = computed(() => {
  let users = chatStore.contacts.users;
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase();
    users = users.filter(u =>
      `${u.first_name} ${u.last_name}`.toLowerCase().includes(q) ||
      u.username.toLowerCase().includes(q)
    );
  }
  return users;
});

const unreadUsers = computed(() => filteredUsers.value.filter(u => chatStore.unreadCounts[`user_${u.id}`] > 0));
const hasUnreadUsers = computed(() => unreadUsers.value.length > 0);

const groupedUsers = computed(() => {
  const groups = {};
  filteredUsers.value.forEach(user => {
    if (user.groups && user.groups.length > 0) {
      user.groups.forEach(g => {
        if (!groups[g.name]) groups[g.name] = { color: g.color, users: [] };
        groups[g.name].users.push(user);
      });
    }
  });
  return groups;
});

const otherUsers = computed(() => filteredUsers.value.filter(u => !u.groups || u.groups.length === 0));

const notificationsEnabled = ref(localStorage.getItem('chat_notifications') !== 'false');
const toggleNotifications = () => {
  notificationsEnabled.value = !notificationsEnabled.value;
  localStorage.setItem('chat_notifications', String(notificationsEnabled.value));
  showMenu.value = false;
};
const toggleMenu = () => { showMenu.value = !showMenu.value; };

const playNotificationSound = async () => {
  if (!notificationsEnabled.value) return;
  try {
    const audioCtx = new (window.AudioContext || window.webkitAudioContext)();
    if (audioCtx.state === 'suspended') await audioCtx.resume();
    const oscillator = audioCtx.createOscillator();
    const gainNode = audioCtx.createGain();
    oscillator.connect(gainNode);
    gainNode.connect(audioCtx.destination);
    oscillator.type = 'sine';
    oscillator.frequency.setValueAtTime(880, audioCtx.currentTime);
    gainNode.gain.setValueAtTime(0, audioCtx.currentTime);
    gainNode.gain.linearRampToValueAtTime(0.15, audioCtx.currentTime + 0.01);
    gainNode.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + 0.4);
    oscillator.start();
    oscillator.stop(audioCtx.currentTime + 0.4);
    setTimeout(() => audioCtx.close(), 500);
  } catch (e) { console.warn('Audio notification error:', e); }
};

const formatMessage = (content) => {
  if (!content) return '';
  const urlRegex = /(https?:\/\/[^\s]+)/g;
  const formatted = content.replace(urlRegex, (url) =>
    `<a href="${url}" target="_blank" rel="noopener noreferrer" class="underline opacity-80 hover:opacity-100 break-all">${url}</a>`
  );
  return DOMPurify.sanitize(formatted, { ALLOWED_TAGS: ['a'], ALLOWED_ATTR: ['href', 'target', 'rel', 'class'] });
};

const totalUnread = computed(() => Object.values(chatStore.unreadCounts).reduce((a, b) => a + b, 0));

const currentMessages = computed(() => {
  if (!activeConversation.value) return [];
  const key = activeConversation.value.type === 'user'
    ? `user_${activeConversation.value.id}`
    : `group_${activeConversation.value.id}`;
  return messages.value[key] || [];
});

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
  });
};

const sendMessage = () => {
  if (!newMessage.value.trim()) return;
  chatStore.sendMessage(newMessage.value);
  newMessage.value = '';
  scrollToBottom();
};

const formatTime = (dateStr) => new Date(dateStr).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

const deleteMessage = (msgId) => {
  if (confirm('Supprimer ce message ?')) chatStore.deleteMessage(msgId);
};

const confirmClearHistory = () => {
  showMenu.value = false;
  if (confirm("Voulez-vous vraiment effacer tout l'historique de cette conversation ?")) chatStore.clearHistory();
};

const previousTotalUnread = ref(0);
const faceMode = ref('user');
let faceInterval = null;

onMounted(() => {
  window.addEventListener('resize', () => { isMobile.value = window.innerWidth < 768; });
  const observer = new MutationObserver(() => { isDark.value = document.documentElement.classList.contains('dark'); });
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });
  if (authStore.token) chatStore.connect();
  previousTotalUnread.value = totalUnread.value;

  faceInterval = setInterval(() => {
    faceMode.value = faceMode.value === 'user' ? 'ai' : 'user';
  }, 3000);
});

onUnmounted(() => {
  if (faceInterval) clearInterval(faceInterval);
});

watch(currentMessages, () => scrollToBottom(), { deep: true });
watch(activeConversation, (val) => { if (val) scrollToBottom(); });
watch(totalUnread, (newTotal, oldTotal) => {
  if (newTotal > oldTotal && newTotal > 0) playNotificationSound();
  previousTotalUnread.value = newTotal;
});
</script>

<style scoped>
/* ── Fenêtre — classes préfixées co- pour éviter conflit @n8n/chat ── */
.co-window {
  box-shadow: 0 20px 60px rgba(0,0,0,0.18), 0 0 0 1px rgba(0,0,0,0.06);
}

.co-window--normal {
  position: fixed;
  bottom: 80px;
  right: 20px;
  width: 360px;
  height: 520px;
  max-width: calc(100vw - 32px);
  max-height: calc(100vh - 110px);
  border-radius: 16px;
  overflow: hidden;
  z-index: 9999;
  transition: width 0.25s ease, height 0.25s ease, border-radius 0.25s ease;
}

@media (min-width: 768px) {
  .co-window--normal {
    width: 620px;
    height: 560px;
  }
}

.co-window--fullscreen {
  position: fixed !important;
  inset: 0 !important;
  width: 100vw !important;
  height: 100vh !important;
  border-radius: 0 !important;
  z-index: 10002;
}

/* ── Bulles — noms uniques pour éviter conflit @n8n/chat ── */
.co-msg-wrap {
  display: flex;
  flex-direction: column;
  max-width: 240px;
  min-width: 0;
}

@media (min-width: 480px) {
  .co-msg-wrap {
    max-width: 280px;
  }
}

@media (min-width: 768px) {
  .co-msg-wrap {
    max-width: 340px;
  }
}

.co-window--fullscreen .co-msg-wrap {
  max-width: min(600px, 65vw);
}

.co-msg-bubble {
  padding: 8px 12px;
  border-radius: 16px;
  font-size: 13.5px;
  line-height: 1.5;
  word-break: break-word;
  display: flex;
  align-items: flex-start;
  gap: 6px;
}

.co-msg-bubble--me {
  background: linear-gradient(135deg, #2563eb, #3b82f6);
  color: white;
  border-bottom-right-radius: 4px;
}

.co-msg-bubble--other {
  background: white;
  color: #1f2937;
  border-bottom-left-radius: 4px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
  border: 1px solid #f3f4f6;
}

:global(.dark) .co-msg-bubble--other {
  background: #1e293b;
  color: #e2e8f0;
  border-color: rgba(255,255,255,0.06);
}

.co-msg-text { flex: 1; }

.co-del-btn {
  opacity: 0;
  background: rgba(255,255,255,0.15);
  border: none;
  border-radius: 4px;
  padding: 2px;
  cursor: pointer;
  color: inherit;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 0.15s, background 0.15s;
  flex-shrink: 0;
}
.group:hover .co-del-btn { opacity: 1; }
.co-del-btn:hover { background: rgba(239,68,68,0.6); }

/* ── FAB flottant animé ─────────────────────────────────── */
@keyframes co-float {
  0%, 100% { transform: translateY(0px); }
  50%       { transform: translateY(-9px); }
}

@keyframes co-halo-expand {
  0%   { transform: scale(1);   opacity: 0.55; }
  100% { transform: scale(1.9); opacity: 0; }
}

.co-fab {
  animation: co-float 3.2s ease-in-out infinite;
}
.co-fab:hover {
  animation-play-state: paused;
}

/* Anneau coloré */
.co-fab-ring {
  transition: background 0.7s ease, box-shadow 0.7s ease;
}
.co-fab-ring--user {
  background: linear-gradient(135deg, #2563eb, #4f46e5);
  box-shadow: 0 6px 22px rgba(37,99,235,0.55), 0 0 0 3px rgba(99,102,241,0.25);
}
.co-fab-ring--ai {
  background: linear-gradient(135deg, #7c3aed, #a855f7);
  box-shadow: 0 6px 22px rgba(124,58,237,0.55), 0 0 0 3px rgba(168,85,247,0.25);
}

/* Halo externe */
.co-fab-halo {
  position: absolute;
  inset: 0;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  animation: co-halo-expand 3.2s ease-out infinite;
  pointer-events: none;
}
.co-fab-halo--user { background: rgba(99,102,241,0.45); }
.co-fab-halo--ai   { background: rgba(168,85,247,0.45); }

/* Pulse non-lus */
@keyframes co-unread-pulse {
  0%, 100% { box-shadow: 0 6px 22px rgba(37,99,235,0.55), 0 0 0 3px rgba(99,102,241,0.25); }
  50%       { box-shadow: 0 8px 30px rgba(37,99,235,0.75), 0 0 0 5px rgba(99,102,241,0.35); }
}
.co-fab-unread-pulse { animation: co-unread-pulse 1.8s ease-in-out infinite; }

/* ── Transition de visage ───────────────────────────────── */
.co-face-enter-active {
  transition: opacity 0.35s ease, transform 0.35s cubic-bezier(0.34,1.3,0.64,1);
}
.co-face-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
  position: absolute;
}
.co-face-enter-from { opacity: 0; transform: scale(0.6) rotate(-20deg); }
.co-face-leave-to   { opacity: 0; transform: scale(0.6) rotate(20deg); }

/* ── Transitions ────────────────────────────────────────── */
.co-window-anim-enter-active,
.co-window-anim-leave-active {
  transition: opacity 0.22s ease, transform 0.26s cubic-bezier(0.34, 1.2, 0.64, 1);
}
.co-window-anim-enter-from,
.co-window-anim-leave-to {
  opacity: 0;
  transform: translateY(14px) scale(0.97);
}

.co-menu-drop-enter-active, .co-menu-drop-leave-active { transition: opacity 0.15s, transform 0.18s; }
.co-menu-drop-enter-from, .co-menu-drop-leave-to { opacity: 0; transform: translateY(-6px) scale(0.96); }

.co-emoji-pop-enter-active, .co-emoji-pop-leave-active { transition: opacity 0.15s, transform 0.18s; }
.co-emoji-pop-enter-from, .co-emoji-pop-leave-to { opacity: 0; transform: translateY(6px); }

/* ── Mobile ─────────────────────────────────────────────── */
@media (max-width: 600px) {
  .co-window--normal {
    left: 8px !important;
    right: 8px !important;
    width: auto !important;
    max-width: none !important;
    bottom: 80px !important;
    height: 72vh !important;
    max-height: 540px;
  }
  .co-window--fullscreen {
    inset: 0 !important;
  }
}
</style>
