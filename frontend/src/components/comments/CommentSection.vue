<template>
  <div class="comment-section">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ $t('comments.title') }} ({{ comments.length }})
      </h3>
    </div>

    <!-- Formulaire d'ajout de commentaire -->
    <div v-if="canComment" class="mb-6">
      <textarea
        v-model="newComment"
        :placeholder="$t('comments.placeholder')"
        :maxlength="maxCommentLength"
        class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg
               focus:ring-2 focus:ring-blue-500 focus:border-transparent
               bg-white dark:bg-gray-700 text-gray-900 dark:text-white
               resize-none"
        rows="3"
        @keydown.ctrl.enter="submitComment"
      />
      <div class="flex items-center justify-between mt-2">
        <span class="text-sm text-gray-500 dark:text-gray-400">
          {{ newComment.length }} / {{ maxCommentLength }}
        </span>
        <button
          :disabled="!newComment.trim() || isSubmitting"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700
                 disabled:opacity-50 disabled:cursor-not-allowed
                 transition-colors"
          @click="submitComment"
        >
          <Icon v-if="isSubmitting" icon="mdi:loading" class="animate-spin h-5 w-5" />
          <span v-else>{{ $t('comments.submit') }}</span>
        </button>
      </div>
    </div>

    <!-- Message si commentaires désactivés -->
    <div v-else-if="!commentsEnabled" class="text-center py-8 text-gray-500 dark:text-gray-400">
      <Icon icon="mdi:comment-off-outline" class="h-12 w-12 mx-auto mb-2" />
      <p>{{ $t('comments.disabled') }}</p>
    </div>

    <!-- Liste des commentaires -->
    <div v-if="comments.length > 0" class="space-y-4">
      <CommentItem
        v-for="comment in sortedComments"
        :key="comment.id"
        :comment="comment"
        @edit="editComment"
        @delete="deleteCommentData"
      />
    </div>

    <!-- Message si aucun commentaire -->
    <div v-else-if="commentsEnabled" class="text-center py-8 text-gray-500 dark:text-gray-400">
      <Icon icon="mdi:comment-outline" class="h-12 w-12 mx-auto mb-2" />
      <p>{{ $t('comments.empty') }}</p>
    </div>

    <!-- Modal d'édition -->
    <CommentEditModal
      v-if="editingComment"
      :comment="editingComment"
      :max-length="maxCommentLength"
      @close="editingComment = null"
      @save="updateCommentData"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { getComments, createComment, updateComment, deleteComment, getCommentSettings } from '@/services/api'
import CommentItem from './CommentItem.vue'
import CommentEditModal from './CommentEditModal.vue'

const props = defineProps({
  entityType: {
    type: String,
    required: true,
    validator: (value) => ['news', 'application', 'event', 'suggestion'].includes(value)
  },
  entityId: {
    type: Number,
    required: true
  }
})

const authStore = useAuthStore()
const comments = ref([])
const newComment = ref('')
const isSubmitting = ref(false)
const editingComment = ref(null)
const commentsEnabled = ref(true)
const maxCommentLength = ref(1000)

const canComment = computed(() => {
  return authStore.isAuthenticated && commentsEnabled.value
})

const sortedComments = computed(() => {
  return [...comments.value].sort((a, b) =>
    new Date(b.created_at) - new Date(a.created_at)
  )
})

const loadComments = async () => {
  try {
    const response = await getComments(props.entityType, props.entityId)
    comments.value = response.comments || []
  } catch (error) {
    console.error('Erreur lors du chargement des commentaires:', error)
  }
}

const loadSettings = async () => {
  try {
    const settings = await getCommentSettings()
    commentsEnabled.value = settings.comments_enabled
    maxCommentLength.value = settings.max_comment_length || 1000

    // Vérifier si commentaires activés pour ce type d'entité
    if (props.entityType === 'news' && !settings.news_comments_enabled) {
      commentsEnabled.value = false
    } else if (props.entityType === 'application' && !settings.app_comments_enabled) {
      commentsEnabled.value = false
    } else if (props.entityType === 'event' && !settings.event_comments_enabled) {
      commentsEnabled.value = false
    }
  } catch (error) {
    console.error('Erreur lors du chargement des paramètres:', error)
  }
}

const submitComment = async () => {
  if (!newComment.value.trim() || isSubmitting.value) return

  isSubmitting.value = true
  try {
    await createComment({
      content: newComment.value.trim(),
      entity_type: props.entityType,
      entity_id: props.entityId
    })
    newComment.value = ''
    await loadComments()
  } catch (error) {
    console.error('Erreur lors de la création du commentaire:', error)
    alert('Erreur lors de la création du commentaire')
  } finally {
    isSubmitting.value = false
  }
}

const editComment = (comment) => {
  editingComment.value = { ...comment }
}

const updateCommentData = async (updatedComment) => {
  try {
    await updateComment(updatedComment.id, {
      content: updatedComment.content,
      entity_type: props.entityType,
      entity_id: props.entityId
    })
    editingComment.value = null
    await loadComments()
  } catch (error) {
    console.error('Erreur lors de la mise à jour du commentaire:', error)
    alert('Erreur lors de la mise à jour du commentaire')
  }
}

const deleteCommentData = async (commentId) => {
  if (!confirm('Êtes-vous sûr de vouloir supprimer ce commentaire ?')) return

  try {
    await deleteComment(commentId)
    await loadComments()
  } catch (error) {
    console.error('Erreur lors de la suppression du commentaire:', error)
    alert('Erreur lors de la suppression du commentaire')
  }
}

onMounted(() => {
  loadSettings()
  loadComments()
})
</script>

<style scoped>
.comment-section {
  @apply mt-8;
}
</style>
