import { afterEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import PersonDetailModal from '../PersonDetailModal.vue'
import { makeI18n, $ } from '@/components/ui/__tests__/helpers'

const person = { id: 3, first_name: 'Inès', last_name: 'Haddad', is_active: true, custom_fields: {} }

const mountModal = () =>
  mount(PersonDetailModal, {
    props: { person, unitPath: () => [] },
    global: { plugins: [makeI18n()] },
    attachTo: document.body,
  })

describe('PersonDetailModal', () => {
  let wrapper
  afterEach(() => wrapper?.unmount())

  it('moves focus into the dialog so Escape closes it', async () => {
    wrapper = mountModal()
    expect(document.activeElement).toBe($('.person-detail__close'))
    document.activeElement.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('returns focus to the element that opened it', () => {
    const opener = document.createElement('button')
    document.body.appendChild(opener)
    opener.focus()
    wrapper = mountModal()
    wrapper.unmount()
    wrapper = null
    expect(document.activeElement).toBe(opener)
    opener.remove()
  })
})
