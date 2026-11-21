import { afterEach, describe, expect, it } from 'vitest'
import { defineComponent, h, nextTick, ref, withDirectives } from 'vue'
import { mount } from '@vue/test-utils'
import { focusTrap } from '../focusTrap'

const pressTab = (shiftKey = false) => {
  const event = new KeyboardEvent('keydown', { key: 'Tab', shiftKey, bubbles: true, cancelable: true })
  document.activeElement.dispatchEvent(event)
  return event
}

// Une page avec un bouton d'ouverture et une modale piégée affichée sur demande.
const Page = defineComponent({
  props: { initialOpen: { type: Boolean, default: false }, empty: { type: Boolean, default: false } },
  setup(props, { expose }) {
    const open = ref(props.initialOpen)
    expose({ open })
    return () => [
      h('button', { id: 'opener', onClick: () => (open.value = true) }, 'open'),
      open.value
        ? withDirectives(
            h('div', { id: 'dialog', role: 'dialog' }, props.empty
              ? [h('p', 'texte')]
              : [
                  h('button', { id: 'first' }, 'a'),
                  h('input', { id: 'middle' }),
                  h('button', { id: 'disabled', disabled: true }, 'x'),
                  h('button', { id: 'last' }, 'b'),
                ]),
            [[focusTrap]],
          )
        : null,
    ]
  },
})

const wrappers = []
const mountPage = (props) => {
  const wrapper = mount(Page, { props, attachTo: document.body })
  wrappers.push(wrapper)
  return wrapper
}

afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ''
})

describe('v-focus-trap', () => {
  it('place le focus sur le premier élément focalisable à l’ouverture', async () => {
    const wrapper = mountPage()
    document.getElementById('opener').focus()
    wrapper.vm.open = true
    await nextTick()
    await nextTick()
    expect(document.activeElement.id).toBe('first')
  })

  it('boucle de la fin au début avec Tab, et du début à la fin avec Maj+Tab', async () => {
    mountPage({ initialOpen: true })
    await nextTick()
    document.getElementById('last').focus()
    expect(pressTab().defaultPrevented).toBe(true)
    expect(document.activeElement.id).toBe('first')

    expect(pressTab(true).defaultPrevented).toBe(true)
    expect(document.activeElement.id).toBe('last')
  })

  it('laisse le navigateur avancer normalement au milieu de la modale', async () => {
    mountPage({ initialOpen: true })
    await nextTick()
    document.getElementById('first').focus()
    expect(pressTab().defaultPrevented).toBe(false)
  })

  it('ramène dans la modale un focus parti dehors', async () => {
    mountPage({ initialOpen: true })
    await nextTick()
    document.getElementById('opener').focus()
    expect(pressTab().defaultPrevented).toBe(true)
    expect(document.activeElement.id).toBe('first')
  })

  it('focalise la modale elle-même quand elle ne contient rien de focalisable', async () => {
    mountPage({ initialOpen: true, empty: true })
    await nextTick()
    await nextTick()
    expect(document.activeElement.id).toBe('dialog')
    expect(pressTab().defaultPrevented).toBe(true)
    expect(document.activeElement.id).toBe('dialog')
  })

  it('rend le focus à l’élément d’ouverture à la fermeture', async () => {
    const wrapper = mountPage()
    document.getElementById('opener').focus()
    wrapper.vm.open = true
    await nextTick()
    await nextTick()
    wrapper.vm.open = false
    await nextTick()
    expect(document.activeElement.id).toBe('opener')
  })
})
