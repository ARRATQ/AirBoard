import { describe, it, expect, vi, afterEach } from 'vitest'
import { defineComponent, h } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { useChatDeepLink } from '../useChatDeepLink'

const Page = defineComponent({ render: () => h('div') })

const makeRouter = () =>
  createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/home', component: Page },
      { path: '/chat', redirect: { path: '/home', query: { chat: 'open' } } },
    ],
  })

const mounted = []
const mountWith = async (router, url, state = { open: false }) => {
  const open = vi.fn(() => { state.open = true })
  const Host = defineComponent({
    setup() {
      useChatDeepLink(() => state.open, open)
      return () => h('div')
    },
  })
  router.push(url)
  await router.isReady()
  mounted.push(mount(Host, { global: { plugins: [router] } }))
  await flushPromises()
  return open
}

afterEach(() => mounted.splice(0).forEach((w) => w.unmount()))

describe('useChatDeepLink', () => {
  it('opens the chat from the /chat link and cleans the URL', async () => {
    const router = makeRouter()
    const open = await mountWith(router, '/chat')
    expect(open).toHaveBeenCalledTimes(1)
    expect(router.currentRoute.value.fullPath).toBe('/home')
  })

  it('keeps other query parameters', async () => {
    const router = makeRouter()
    await mountWith(router, '/home?chat=open&tab=2')
    expect(router.currentRoute.value.query).toEqual({ tab: '2' })
  })

  it('does not toggle an already open chat closed', async () => {
    const router = makeRouter()
    const open = await mountWith(router, '/chat', { open: true })
    expect(open).not.toHaveBeenCalled()
    expect(router.currentRoute.value.query.chat).toBeUndefined()
  })

  it('reacts to a later navigation and ignores other values', async () => {
    const router = makeRouter()
    const open = await mountWith(router, '/home?chat=closed')
    expect(open).not.toHaveBeenCalled()
    await router.push('/chat')
    await flushPromises()
    expect(open).toHaveBeenCalledTimes(1)
  })
})
