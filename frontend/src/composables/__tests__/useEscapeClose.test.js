import { describe, it, expect, vi, afterEach } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { useEscapeClose } from '../useEscapeClose'

const pressEscape = (key = 'Escape') => {
  const event = new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true })
  document.dispatchEvent(event)
  return event
}

const Modal = defineComponent({
  props: { open: { type: [Object, Function], required: true }, onClose: { type: Function, required: true } },
  setup(props) {
    useEscapeClose(props.open, props.onClose)
    return () => h('div')
  },
})

const mounted = []
const mountModal = (open, onClose) => {
  const wrapper = mount(Modal, { props: { open, onClose }, attachTo: document.body })
  mounted.push(wrapper)
  return wrapper
}

afterEach(() => {
  mounted.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ''
})

describe('useEscapeClose', () => {
  it('closes an open modal on Escape only', async () => {
    const open = ref(true)
    const onClose = vi.fn()
    mountModal(open, onClose)
    await nextTick()

    pressEscape('Enter')
    expect(onClose).not.toHaveBeenCalled()
    const event = pressEscape()
    expect(onClose).toHaveBeenCalledTimes(1)
    expect(event.defaultPrevented).toBe(true)
  })

  it('ignores Escape while the modal is closed', async () => {
    const open = ref(false)
    const onClose = vi.fn()
    mountModal(open, onClose)
    await nextTick()

    pressEscape()
    expect(onClose).not.toHaveBeenCalled()
  })

  it('closes only the most recently opened modal', async () => {
    const first = vi.fn()
    const second = vi.fn()
    const secondOpen = ref(false)
    mountModal(ref(true), first)
    mountModal(secondOpen, second)
    secondOpen.value = true
    await nextTick()

    pressEscape()
    expect(second).toHaveBeenCalledTimes(1)
    expect(first).not.toHaveBeenCalled()

    secondOpen.value = false
    await nextTick()
    pressEscape()
    expect(first).toHaveBeenCalledTimes(1)
  })

  it('stops listening once the component is unmounted', async () => {
    const onClose = vi.fn()
    const wrapper = mountModal(ref(true), onClose)
    await nextTick()
    wrapper.unmount()
    mounted.splice(mounted.indexOf(wrapper), 1)

    pressEscape()
    expect(onClose).not.toHaveBeenCalled()
  })

  it('leaves Escape to an open Headless UI dialog stacked above', async () => {
    const onClose = vi.fn()
    mountModal(ref(true), onClose)
    await nextTick()
    const dialog = document.createElement('div')
    dialog.id = 'headlessui-dialog-1'
    dialog.setAttribute('aria-modal', 'true')
    document.body.appendChild(dialog)

    pressEscape()
    expect(onClose).not.toHaveBeenCalled()
  })

  it('accepts a getter as the open state', async () => {
    const event = ref({ id: 1 })
    const onClose = vi.fn()
    mountModal(() => Boolean(event.value), onClose)
    await nextTick()

    pressEscape()
    expect(onClose).toHaveBeenCalledTimes(1)
  })
})
