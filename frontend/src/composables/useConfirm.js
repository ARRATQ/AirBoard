// Confirmation asynchrone qui remplace window.confirm() :
//   const { confirm } = useConfirm()
//   if (!(await confirm({ message: t('…'), tone: 'danger' }))) return
// Un seul <ConfirmDialog /> est monté dans App.vue et lit cet état.

import { reactive, readonly } from 'vue'

const EMPTY = {
  open: false,
  title: '',
  message: '',
  confirmLabel: '',
  cancelLabel: '',
  tone: 'default', // default | danger
}

const state = reactive({ ...EMPTY })
let pending = null

/**
 * Ouvre la boîte de confirmation.
 * @param {string | { title?: string, message: string, confirmLabel?: string, cancelLabel?: string, tone?: 'default'|'danger' }} options
 * @returns {Promise<boolean>} true si confirmé, false sinon (Annuler, Échap, clic extérieur).
 */
export function confirm(options) {
  const opts = typeof options === 'string' ? { message: options } : { ...options }
  // Une nouvelle demande annule la précédente restée ouverte.
  if (pending) pending(false)
  Object.assign(state, EMPTY, opts, { open: true })
  return new Promise((resolve) => {
    pending = resolve
  })
}

export function settleConfirm(value) {
  const resolve = pending
  pending = null
  state.open = false
  if (resolve) resolve(Boolean(value))
}

export const confirmState = readonly(state)

export function useConfirm() {
  return { confirm }
}
