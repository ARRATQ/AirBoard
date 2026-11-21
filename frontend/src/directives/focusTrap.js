// v-focus-trap : garde le focus clavier dans une modale maison tant qu'elle est affichée,
// et le rend à l'élément qui l'avait ouverte à la fermeture.
//
//   <div v-focus-trap role="dialog" aria-modal="true" aria-labelledby="…">…</div>
//   <div v-focus-trap="isOpen" …>   (pour une modale affichée avec v-show)
//
// Les modales Headless UI (BaseModal, ConfirmDialog) ont leur propre piège : quand l'une
// d'elles est ouverte par-dessus, on la laisse faire. Une pile garantit que seule la
// modale ouverte en dernier retient le focus.
import { nextTick } from 'vue'

const FOCUSABLE = [
  'a[href]',
  'area[href]',
  'button:not([disabled])',
  'input:not([disabled]):not([type="hidden"])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  'iframe',
  'summary',
  '[contenteditable="true"]',
  '[tabindex]:not([tabindex="-1"])',
].join(',')

const stack = []

const headlessDialogOpen = () =>
  Boolean(document.querySelector('[id^="headlessui-dialog-"][aria-modal="true"]'))

const isVisible = (node) =>
  !node.closest('[hidden], [inert]') && (typeof node.checkVisibility !== 'function' || node.checkVisibility())

/** Éléments focalisables et visibles d'un conteneur, dans l'ordre du document. */
export const focusableIn = (container) =>
  [...container.querySelectorAll(FOCUSABLE)].filter(isVisible)

const focusFirst = (container) => {
  const target = container.querySelector('[autofocus]') || focusableIn(container)[0]
  if (target) {
    target.focus()
    return
  }
  if (!container.hasAttribute('tabindex')) container.setAttribute('tabindex', '-1')
  container.focus()
}

const onKeydown = (event) => {
  if (event.key !== 'Tab' || stack.length === 0 || headlessDialogOpen()) return
  const container = stack[stack.length - 1]
  const items = focusableIn(container)
  const active = document.activeElement
  if (items.length === 0) {
    event.preventDefault()
    container.focus()
    return
  }
  const first = items[0]
  const last = items[items.length - 1]
  if (!container.contains(active)) {
    event.preventDefault()
    ;(event.shiftKey ? last : first).focus()
  } else if (event.shiftKey && (active === first || active === container)) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && active === last) {
    event.preventDefault()
    first.focus()
  }
}

const activate = (el) => {
  if (el.__focusTrap) return
  el.__focusTrap = { previous: document.activeElement }
  stack.push(el)
  if (stack.length === 1) document.addEventListener('keydown', onKeydown, true)
  // Attendre le rendu des enfants (v-if internes, transitions) avant de placer le focus.
  nextTick(() => {
    if (el.__focusTrap && !el.contains(document.activeElement)) focusFirst(el)
  })
}

const deactivate = (el) => {
  const state = el.__focusTrap
  if (!state) return
  delete el.__focusTrap
  const index = stack.indexOf(el)
  if (index !== -1) stack.splice(index, 1)
  if (stack.length === 0) document.removeEventListener('keydown', onKeydown, true)
  const { previous } = state
  if (previous && previous.isConnected && typeof previous.focus === 'function') previous.focus()
}

export const focusTrap = {
  mounted(el, binding) {
    if (binding.value !== false) activate(el)
  },
  updated(el, binding) {
    if (binding.value === binding.oldValue) return
    if (binding.value === false) deactivate(el)
    else activate(el)
  },
  beforeUnmount(el) {
    deactivate(el)
  },
}

export default focusTrap
