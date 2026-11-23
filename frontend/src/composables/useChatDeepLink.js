import { watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

// Paramètre posé par la redirection /chat (lien « Ouvrir la conversation » des emails).
export const CHAT_QUERY = 'chat'
export const CHAT_QUERY_OPEN = 'open'

/**
 * Ouvre la messagerie quand l'URL porte ?chat=open, puis retire le paramètre
 * pour qu'un rechargement ou un retour arrière ne la rouvre pas.
 * @param {() => boolean} isOpen  état actuel du panneau
 * @param {() => void} open       ouvre le panneau
 */
export function useChatDeepLink(isOpen, open) {
  const route = useRoute()
  const router = useRouter()

  watch(
    () => route.query[CHAT_QUERY],
    (value) => {
      if (value !== CHAT_QUERY_OPEN) return
      if (!isOpen()) open()
      const query = Object.fromEntries(Object.entries(route.query).filter(([key]) => key !== CHAT_QUERY))
      router.replace({ query, hash: route.hash })
    },
    { immediate: true }
  )
}
