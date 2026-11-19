import { afterEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ConfirmDialog from '../ConfirmDialog.vue'
import { confirm, confirmState, settleConfirm } from '@/composables/useConfirm'
import { $, $$, flush, makeI18n } from './helpers'

const mountDialog = () =>
  mount(ConfirmDialog, { global: { plugins: [makeI18n()] }, attachTo: document.body })

let wrapper
afterEach(() => {
  settleConfirm(false)
  wrapper?.unmount()
  document.body.innerHTML = ''
})

describe('ConfirmDialog', () => {
  it('resolves true when the confirm button is clicked', async () => {
    wrapper = mountDialog()
    const answer = confirm({ message: 'Supprimer ce sondage ?', tone: 'danger' })
    await flush()

    expect($('[role="alertdialog"]')).not.toBeNull()
    expect(document.body.textContent).toContain('Supprimer ce sondage ?')
    const buttons = $$('.base-modal__footer button')
    expect(buttons.map(b => b.textContent.trim())).toEqual(['Annuler', 'Supprimer'])

    buttons[1].click()
    await expect(answer).resolves.toBe(true)
    expect(confirmState.open).toBe(false)
  })

  it('resolves false when cancelled', async () => {
    wrapper = mountDialog()
    const answer = confirm('Continuer ?')
    await flush()

    $$('.base-modal__footer button')[0].click()
    await expect(answer).resolves.toBe(false)
  })

  it('resolves false on Escape', async () => {
    wrapper = mountDialog()
    const answer = confirm('Continuer ?')
    await flush()

    $('[role="alertdialog"]').dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await expect(answer).resolves.toBe(false)
  })

  it('uses custom labels and the neutral confirm label by default', async () => {
    wrapper = mountDialog()
    confirm({ message: 'Publier ?' })
    await flush()
    expect($$('.base-modal__footer button')[1].textContent.trim()).toBe('Confirmer')

    confirm({ message: 'Publier ?', confirmLabel: 'Publier', cancelLabel: 'Plus tard' })
    await flush()
    expect($$('.base-modal__footer button').map(b => b.textContent.trim())).toEqual(['Plus tard', 'Publier'])
  })

  it('cancels a pending request when a new one opens', async () => {
    wrapper = mountDialog()
    const first = confirm('Premier ?')
    confirm('Second ?')
    await expect(first).resolves.toBe(false)
  })

  it('focuses the cancel button first', async () => {
    wrapper = mountDialog()
    confirm({ message: 'Supprimer ?', tone: 'danger' })
    await flush()
    await flush()
    expect(document.activeElement?.textContent.trim()).toBe('Annuler')
  })
})
