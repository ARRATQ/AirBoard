// Échap ferme les modales maison (BaseModal et ConfirmDialog le font déjà via Headless UI).
// Une pile commune garantit que seule la modale ouverte en dernier se ferme, et un seul
// écouteur sur document sert toutes les modales, quel que soit l'élément qui a le focus.
import { onBeforeUnmount, toValue, watch } from 'vue'

const stack = []

// Une boîte de dialogue Headless UI (BaseModal, ConfirmDialog) ouverte par-dessus
// gère elle-même Échap : on la laisse faire.
const headlessDialogOpen = () =>
  Boolean(document.querySelector('[id^="headlessui-dialog-"][aria-modal="true"]'))

const onKeydown = (event) => {
  if (event.key !== 'Escape' || event.defaultPrevented || stack.length === 0) return
  if (headlessDialogOpen()) return
  event.preventDefault()
  stack[stack.length - 1].close()
}

const add = (entry) => {
  if (stack.includes(entry)) return
  stack.push(entry)
  if (stack.length === 1) document.addEventListener('keydown', onKeydown)
}

const remove = (entry) => {
  const index = stack.indexOf(entry)
  if (index === -1) return
  stack.splice(index, 1)
  if (stack.length === 0) document.removeEventListener('keydown', onKeydown)
}

/**
 * Ferme la modale avec Échap tant qu'elle est ouverte.
 * @param {import('vue').MaybeRefOrGetter<boolean>} isOpen état d'ouverture (ref, computed ou getter)
 * @param {() => void} close action de fermeture (la même que le bouton Annuler)
 */
export function useEscapeClose(isOpen, close) {
  const entry = { close }
  watch(
    () => Boolean(toValue(isOpen)),
    (open) => (open ? add(entry) : remove(entry)),
    { immediate: true },
  )
  onBeforeUnmount(() => remove(entry))
}
