import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import PersonCard from '../PersonCard.vue'
import { makeI18n } from '@/components/ui/__tests__/helpers'

const person = {
  id: 7,
  first_name: 'Inès',
  last_name: 'Haddad',
  email: 'ines.haddad@example.org',
  phone: '01 23 45 67 89',
  is_active: true,
  custom_fields: {},
}

const roles = [
  {
    position: { title: 'Responsable RH' },
    unit: { name: 'Ressources humaines', color: '#0F766E' },
    assignment: { is_interim: false },
  },
]

const mountCard = (props = {}) =>
  mount(PersonCard, {
    props: { person, roles, ...props },
    global: { plugins: [makeI18n()] },
  })

describe('PersonCard', () => {
  it('does not nest contact links inside a button', () => {
    const wrapper = mountCard()
    expect(wrapper.findAll('button a')).toHaveLength(0)
    expect(wrapper.findAll('a')).toHaveLength(2)
  })

  it('opens the person when the name button is clicked', async () => {
    const wrapper = mountCard()
    const open = wrapper.find('button')
    expect(open.text()).toBe('Inès Haddad')
    await open.trigger('click')
    expect(wrapper.emitted('open')).toEqual([[person]])
  })

  it('does not open the person when a contact link is clicked', async () => {
    const wrapper = mountCard()
    await wrapper.find('a[href^="mailto:"]').trigger('click')
    expect(wrapper.emitted('open')).toBeUndefined()
  })

  it('builds mailto and tel links without spaces in the phone number', () => {
    const wrapper = mountCard()
    const hrefs = wrapper.findAll('a').map((a) => a.attributes('href'))
    expect(hrefs).toEqual(['mailto:ines.haddad@example.org', 'tel:0123456789'])
  })

  it('omits contact links when email and phone are missing', () => {
    const wrapper = mountCard({ person: { ...person, email: '', phone: '' } })
    expect(wrapper.findAll('a')).toHaveLength(0)
  })
})
