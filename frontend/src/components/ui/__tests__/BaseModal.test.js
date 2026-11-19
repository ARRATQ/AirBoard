import { afterEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import BaseModal from '../BaseModal.vue'
import { $, flush, makeI18n } from './helpers'

const mountModal = (props = {}, slots = {}) =>
  mount(BaseModal, {
    props: { open: true, title: 'Modifier le groupe', ...props },
    slots: { default: '<input id="name" />', ...slots },
    global: { plugins: [makeI18n()] },
    attachTo: document.body,
  })

let wrapper
afterEach(() => {
  wrapper?.unmount()
  document.body.innerHTML = ''
})

describe('BaseModal', () => {
  it('renders an accessible dialog labelled by its title', async () => {
    wrapper = mountModal({ description: 'Nom et membres' })
    await flush()

    const dialog = $('[role="dialog"]')
    expect(dialog).not.toBeNull()
    expect(dialog.getAttribute('aria-modal')).toBe('true')
    const title = document.getElementById(dialog.getAttribute('aria-labelledby'))
    expect(title.textContent.trim()).toBe('Modifier le groupe')
    const desc = document.getElementById(dialog.getAttribute('aria-describedby'))
    expect(desc.textContent.trim()).toBe('Nom et membres')
  })

  it('renders nothing when closed', async () => {
    wrapper = mountModal({ open: false })
    await flush()
    expect($('[role="dialog"]')).toBeNull()
  })

  it('emits close on Escape and on the close button', async () => {
    wrapper = mountModal()
    await flush()

    $('[role="dialog"]').dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await flush()
    expect(wrapper.emitted('close')).toHaveLength(1)

    $('.base-modal__close').click()
    expect(wrapper.emitted('close')).toHaveLength(2)
    expect($('.base-modal__close').getAttribute('aria-label')).toBe('Fermer')
  })

  it('does not emit close when not closable', async () => {
    wrapper = mountModal({ closable: false })
    await flush()
    expect($('.base-modal__close')).toBeNull()

    $('[role="dialog"]').dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await flush()
    expect(wrapper.emitted('close')).toBeUndefined()
  })

  it('moves focus inside the dialog', async () => {
    wrapper = mountModal()
    await flush()
    await flush()
    expect($('[role="dialog"]').contains(document.activeElement)).toBe(true)
  })

  it('renders the footer slot', async () => {
    wrapper = mountModal({}, { footer: '<button type="button">Enregistrer</button>' })
    await flush()
    expect($('.base-modal__footer').textContent).toContain('Enregistrer')
  })
})
