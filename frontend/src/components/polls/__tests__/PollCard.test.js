import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { makeI18n, flush } from '@/components/ui/__tests__/helpers'

vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))
vi.mock('@/services/api', () => ({
  pollsService: { getResults: vi.fn().mockResolvedValue({ user_voted: false }), vote: vi.fn(), removeVote: vi.fn() },
}))

import PollCard from '../PollCard.vue'
import { useAuthStore } from '@/stores/auth'

const basePoll = {
  id: 1,
  title: 'Cantine',
  poll_type: 'single',
  is_active: true,
  show_results: 'closed',
  options: [{ id: 1, text: 'Oui' }, { id: 2, text: 'Non' }],
}

const mountCard = async (poll) => {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = { id: 9, role: 'user' }
  const wrapper = mount(PollCard, {
    props: { poll, defaultExpanded: true },
    global: { plugins: [pinia, makeI18n()] },
  })
  await flush()
  return wrapper
}

describe('PollCard', () => {
  it('sondage dont la date de fin est passée : affiché comme fermé', async () => {
    const wrapper = await mountCard({ ...basePoll, end_date: '2020-01-01T00:00:00Z' })
    expect(wrapper.find('.tag--bad').exists()).toBe(true)
    expect(wrapper.find('.waiting-message').text()).toBe('Ce sondage est fermé.')
    expect(wrapper.find('.poll-options').exists()).toBe(false)
  })

  it('sondage en cours : on peut voter', async () => {
    const wrapper = await mountCard({ ...basePoll, end_date: '2099-01-01T00:00:00Z' })
    expect(wrapper.find('.tag--bad').exists()).toBe(false)
    expect(wrapper.find('.poll-options').exists()).toBe(true)
  })
})
