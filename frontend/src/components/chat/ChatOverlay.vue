<template>
  <div class="chat-overlay">
    <!-- Minimized Bubble -->
    <div 
      v-if="!chatStore.isOpen" 
      class="chat-bubble shadow-lg hover:shadow-xl transition-all cursor-pointer bg-blue-600 text-white rounded-full w-14 h-14 flex items-center justify-center relative z-50"
      @click="chatStore.toggleChat()"
    >
      <Icon icon="mdi:chat" class="text-2xl" />
      <span 
        v-if="totalUnread > 0" 
        class="absolute -top-1 -right-1 bg-red-500 text-white text-xs font-bold rounded-full w-5 h-5 flex items-center justify-center border-2 border-white"
      >
        {{ totalUnread > 99 ? '99+' : totalUnread }}
      </span>
    </div>

    <!-- Expanded Window -->
    <div v-else class="chat-window shadow-2xl rounded-t-lg bg-white flex flex-col z-50">
      <!-- Header -->
      <div class="chat-header bg-blue-600 text-white p-3 rounded-t-lg flex justify-between items-center cursor-pointer" @click="toggleMinimize">
        <div class="flex items-center gap-2">
           <Icon icon="mdi:chat" class="text-xl" />
           <span class="font-bold">Messagerie</span>
           <span v-if="!chatStore.isConnected" class="text-xs bg-red-500 px-1 rounded">Offline</span>
        </div>
        <button @click.stop="chatStore.toggleChat()" class="hover:bg-blue-700 rounded p-1">
          <Icon icon="mdi:close" />
        </button>
      </div>

      <div class="flex flex-1 overflow-hidden h-full">
        <!-- Sidebar (Contacts) -->
        <div v-if="!isMobile || !chatStore.activeConversation" class="w-full md:w-1/3 border-r flex flex-col bg-gray-50 h-full">
           <!-- Search (Future) -->
           <!-- <div class="p-2 border-b"><input ... ></div> -->
           
           <div class="overflow-y-auto flex-1 p-2 space-y-2">
              <div v-if="chatStore.contacts.groups.length > 0" class="mb-2">
                <h3 class="text-xs font-bold text-gray-500 uppercase px-2 mb-1">Groupes</h3>
                <div 
                  v-for="group in chatStore.contacts.groups" 
                  :key="group.id"
                  @click="chatStore.openConversation('group', group)"
                  class="flex items-center gap-2 p-2 rounded hover:bg-gray-200 cursor-pointer"
                  :class="{'bg-blue-100': chatStore.activeConversation?.id === group.id && chatStore.activeConversation?.type === 'group'}"
                >
                   <div class="w-8 h-8 rounded bg-gray-300 flex items-center justify-center text-gray-600">
                     <Icon icon="mdi:account-group" />
                   </div>
                   <div class="flex-1 min-w-0">
                     <div class="font-medium truncate">{{ group.name }}</div>
                   </div>
                   <span v-if="chatStore.unreadCounts['group_'+group.id]" class="bg-blue-500 text-white text-xs px-1.5 py-0.5 rounded-full">
                     {{ chatStore.unreadCounts['group_'+group.id] }}
                   </span>
                </div>
              </div>

              <div>
                <h3 class="text-xs font-bold text-gray-500 uppercase px-2 mb-1">Utilisateurs</h3>
                <div 
                  v-for="user in chatStore.contacts.users" 
                  :key="user.id"
                  @click="chatStore.openConversation('user', user)"
                  class="flex items-center gap-2 p-2 rounded hover:bg-gray-200 cursor-pointer"
                  :class="{'bg-blue-100': chatStore.activeConversation?.id === user.id && chatStore.activeConversation?.type === 'user'}"
                >
                   <div class="relative">
                      <img 
                        :src="user.avatar_url || `https://ui-avatars.com/api/?name=${user.first_name}+${user.last_name}&background=random`" 
                        class="w-8 h-8 rounded-full object-cover"
                      >
                      <!-- Online Status (Mocked for now or realized via WS) -->
                      <!-- <span class="absolute bottom-0 right-0 w-2.5 h-2.5 bg-green-500 border-2 border-white rounded-full"></span> -->
                   </div>
                   <div class="flex-1 min-w-0">
                     <div class="font-medium truncate">{{ user.first_name }} {{ user.last_name }}</div>
                     <div class="text-xs text-gray-500 truncate">{{ user.job_title }}</div>
                   </div>
                   <span v-if="chatStore.unreadCounts['user_'+user.id]" class="bg-blue-500 text-white text-xs px-1.5 py-0.5 rounded-full">
                     {{ chatStore.unreadCounts['user_'+user.id] }}
                   </span>
                </div>
              </div>
           </div>
        </div>

        <!-- Conversation View -->
        <div v-if="chatStore.activeConversation" class="flex-1 flex flex-col h-full bg-white" :class="{'hidden md:flex': !isMobile && !chatStore.activeConversation}">
           <!-- Active Header -->
           <div class="p-3 border-b flex justify-between items-center bg-white">
              <div class="flex items-center gap-2">
                 <button v-if="isMobile" @click="chatStore.closeConversation()" class="md:hidden mr-1 text-gray-600"><span class="mdi mdi-arrow-left"></span></button>
                 <span class="font-bold truncate">{{ chatStore.activeConversation.name }}</span>
              </div>
              
              <div class="relative">
                 <button @click.stop="toggleMenu" class="text-white hover:bg-blue-700 p-1 rounded transition-colors flex items-center justify-center">
                    <Icon icon="mdi:dots-vertical" class="text-xl" />
                 </button>
                 <div v-if="showMenu" class="absolute right-0 top-full mt-1 bg-white border rounded shadow-lg w-48 z-20 py-1 text-gray-800">
                    <button 
                      @click="confirmClearHistory"
                      class="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
                    >
                      <Icon icon="mdi:delete-sweep" /> Vider l'historique
                    </button>
                 </div>
              </div>
           </div>

           <!-- Messages Area -->
           <div class="flex-1 overflow-y-auto p-4 space-y-3 flex flex-col" ref="messagesContainer" @click="showMenu = false">
              <!-- Reversed flex-col to stick to bottom -->
               <!-- Note: Data is usually chronological (old -> new). To use flex-col-reverse, we need new -> old order in DOM. 
                    Let's stick to standard flex-col and scroll to bottom for simplicity first. -->
               <div v-for="msg in currentMessages" :key="msg.id" 
                    class="flex flex-col max-w-[80%]"
                    :class="msg.sender_id == myId ? 'self-end items-end' : 'self-start items-start'"
               >
                  <span v-if="chatStore.activeConversation?.type === 'group' && msg.sender_id != myId" class="text-[10px] font-bold text-gray-500 mb-0.5 px-1 truncate max-w-full">
                     {{ msg.sender?.first_name }} {{ msg.sender?.last_name || msg.sender?.username }}
                  </span>
                  <div 
                    class="px-3 py-2 rounded-lg break-words text-sm group flex items-start gap-2"
                    :class="msg.sender_id == myId ? 'bg-blue-600 text-white rounded-tr-none' : 'bg-gray-100 text-gray-800 rounded-tl-none'"
                  >
                    <span class="flex-1">{{ msg.content }}</span>
                    
                    <!-- Delete Button (Inline) -->
                    <button 
                      v-if="msg.sender_id == myId"
                      @click.stop="deleteMessage(msg.id)"
                      class="delete-btn p-1 rounded transition-colors flex items-center justify-center"
                      :class="msg.sender_id == myId ? 'text-white/60 hover:text-white hover:bg-red-500' : 'text-gray-400 hover:text-red-500 hover:bg-red-50'"
                      title="Supprimer"
                    >
                      <Icon icon="mdi:delete" class="text-sm" />
                    </button>
                  </div>
                  <span class="text-[10px] text-gray-400 mt-1">
                    {{ formatTime(msg.created_at) }}
                  </span>
               </div>
               
               <div v-if="currentMessages.length === 0" class="text-center text-gray-400 text-sm mt-4">
                  Début de la conversation
               </div>
           </div>

           <!-- Input Area -->
           <div class="p-3 border-t flex gap-2">
              <input 
                v-model="newMessage" 
                @keyup.enter="sendMessage"
                type="text" 
                placeholder="Écrivez un message..." 
                class="flex-1 border rounded-full px-4 py-2 text-sm focus:outline-none focus:border-blue-500 bg-gray-50"
              >
              <button 
                @click="sendMessage"
                class="bg-blue-600 text-white rounded-full w-9 h-9 flex items-center justify-center hover:bg-blue-700 disabled:opacity-50"
                :disabled="!newMessage.trim()"
              >
                <Icon icon="mdi:send" class="text-sm" />
              </button>
           </div>
        </div>
        
        <!-- Empty State (Desktop) -->
        <div v-else-if="!isMobile" class="flex-1 flex flex-col items-center justify-center text-gray-400 p-4">
           <span class="mdi mdi-message-outline text-4xl mb-2"></span>
           <p>Sélectionnez une conversation</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue';
import { useChatStore } from '@/stores/chat';
import { useAuthStore } from '@/stores/auth';
import { storeToRefs } from 'pinia';
import { Icon } from '@iconify/vue';

const chatStore = useChatStore();
const authStore = useAuthStore();
const { activeConversation, messages } = storeToRefs(chatStore);

const newMessage = ref('');
const messagesContainer = ref(null);
const isMobile = ref(window.innerWidth < 768);
const showMenu = ref(false);

const myId = computed(() => authStore.user?.id);

const toggleMenu = () => {
  showMenu.value = !showMenu.value;
};

const totalUnread = computed(() => {
  return Object.values(chatStore.unreadCounts).reduce((a, b) => a + b, 0);
});

const currentMessages = computed(() => {
  if (!activeConversation.value) return [];
  const key = activeConversation.value.type === 'user' 
     ? `user_${activeConversation.value.id}` 
     : `group_${activeConversation.value.id}`;
  return messages.value[key] || [];
});

const scrollToBottom = () => {
  nextTick(() => {
    // If using flex-col standard
    if (messagesContainer.value) {
        // However, in the template I used flex-col-reverse logic in comments but standard div order.
        // Actually, let's fix the template to strictly follow one logic.
        // Standard chat: div order = chronological. Scroll Top = Max.
        if (messagesContainer.value) {
            messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
        }
    }
  });
};

const sendMessage = () => {
  if (!newMessage.value.trim()) return;
  chatStore.sendMessage(newMessage.value);
  newMessage.value = '';
  scrollToBottom();
};

const formatTime = (dateStr) => {
  const date = new Date(dateStr);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};

// Listeners
onMounted(() => {
    window.addEventListener('resize', () => {
        isMobile.value = window.innerWidth < 768;
    });
    
    // Auto-connect if logged in
    if(authStore.token) {
        chatStore.connect();
    }
});

watch(currentMessages, () => {
   scrollToBottom();
}, { deep: true });

watch(activeConversation, (newVal) => {
    if(newVal) scrollToBottom();
});

const toggleMinimize = () => {
   // Currently clicking header closes chat, logic matches "toggleChat" which toggles isOpen.
   // Minimizing behavior (keeping state but hiding window content) logic is handled by isOpen=false.
   // So toggleChat is fine for now as a simple open/close.
   chatStore.toggleChat();
};

const deleteMessage = (msgId) => {
  if (confirm('Supprimer ce message ?')) {
    chatStore.deleteMessage(msgId);
  }
};

const confirmClearHistory = () => {
   if (confirm('Voulez-vous vraiment effacer tout l\'historique de cette conversation ?')) {
      chatStore.clearHistory();
   }
};

</script>

<style scoped>
.chat-overlay {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 9999;
}

.chat-window {
  width: 350px;
  height: 500px;
  max-width: calc(100vw - 40px);
  max-height: calc(100vh - 100px);
  display: flex;
  flex-direction: column;
}

@media (min-width: 768px) {
    .chat-window {
        width: 600px; /* Wider on desktop for sidebar + chat */
    }
}
</style>
