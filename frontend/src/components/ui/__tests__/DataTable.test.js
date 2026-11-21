import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import DataTable from '../DataTable.vue'
import { makeI18n } from './helpers'

const columns = [
  { key: 'name', label: 'Nom', sortable: true },
  { key: 'group', label: 'Groupe' },
  { key: 'logins', label: 'Connexions', sortable: true, align: 'end' },
]

const rows = [
  { id: 1, name: 'Émile Durand', group: 'RH', logins: 12 },
  { id: 2, name: 'Alice Martin', group: 'IT', logins: 3 },
  { id: 3, name: 'Chloé Bernard', group: 'Direction', logins: 40 },
]

const mountTable = (props = {}, slots = {}) =>
  mount(DataTable, {
    props: { columns, rows, ...props },
    slots,
    global: { plugins: [makeI18n()] },
  })

const names = (wrapper) => wrapper.findAll('tbody tr').map(tr => tr.find('td').text())

describe('DataTable', () => {
  it('renders one row per item with headers', () => {
    const wrapper = mountTable()
    expect(wrapper.findAll('th').map(th => th.text())).toEqual(['Nom', 'Groupe', 'Connexions'])
    expect(names(wrapper)).toEqual(['Émile Durand', 'Alice Martin', 'Chloé Bernard'])
  })

  it('filters rows, ignoring case and accents', async () => {
    const wrapper = mountTable()
    await wrapper.find('input[type="search"]').setValue('emile')
    expect(names(wrapper)).toEqual(['Émile Durand'])

    await wrapper.find('input[type="search"]').setValue('direction')
    expect(names(wrapper)).toEqual(['Chloé Bernard'])
  })

  it('shows the empty state when nothing matches', async () => {
    const wrapper = mountTable()
    await wrapper.find('input[type="search"]').setValue('zzz')
    expect(wrapper.text()).toContain('Aucun résultat.')
  })

  it('sorts ascending then descending and exposes aria-sort', async () => {
    const wrapper = mountTable()
    const header = wrapper.findAll('th')[2]
    expect(header.attributes('aria-sort')).toBe('none')

    await header.find('button').trigger('click')
    expect(header.attributes('aria-sort')).toBe('ascending')
    expect(names(wrapper)).toEqual(['Alice Martin', 'Émile Durand', 'Chloé Bernard'])

    await header.find('button').trigger('click')
    expect(header.attributes('aria-sort')).toBe('descending')
    expect(names(wrapper)).toEqual(['Chloé Bernard', 'Émile Durand', 'Alice Martin'])
  })

  it('sorts text with locale rules', async () => {
    const wrapper = mountTable()
    await wrapper.findAll('th')[0].find('button').trigger('click')
    expect(names(wrapper)).toEqual(['Alice Martin', 'Chloé Bernard', 'Émile Durand'])
  })

  it('paginates and resets to page 1 on search', async () => {
    const many = Array.from({ length: 45 }, (_, i) => ({ id: i + 1, name: `Personne ${i + 1}`, group: 'IT', logins: i }))
    const wrapper = mountTable({ rows: many, pageSize: 20 })
    expect(wrapper.findAll('tbody tr')).toHaveLength(20)
    expect(wrapper.find('.dt__range').text()).toBe('1–20 sur 45')

    const next = wrapper.find('button[aria-label="Page suivante"]')
    await next.trigger('click')
    await next.trigger('click')
    expect(wrapper.find('.dt__range').text()).toBe('41–45 sur 45')
    expect(wrapper.findAll('tbody tr')).toHaveLength(5)
    expect(next.attributes('disabled')).toBeDefined()

    await wrapper.find('input[type="search"]').setValue('Personne 1')
    expect(wrapper.find('.dt__range').text()).toBe('1–11 sur 11')
  })

  it('renders cell slots with the row', () => {
    const wrapper = mountTable({}, { 'cell-group': '<template #cell-group="{ row }"><b>{{ row.group }}!</b></template>' })
    expect(wrapper.findAll('tbody b').map(b => b.text())).toEqual(['RH!', 'IT!', 'Direction!'])
  })

  it('shows a loading row instead of data', () => {
    const wrapper = mountTable({ loading: true })
    expect(wrapper.findAll('tbody tr')).toHaveLength(1)
    expect(wrapper.find('.dt__pager').exists()).toBe(false)
  })

  it('in remote mode, emits search and sort instead of applying them', async () => {
    const wrapper = mountTable({ remote: true, pageSize: 2 })
    expect(names(wrapper)).toEqual(['Émile Durand', 'Alice Martin', 'Chloé Bernard'])

    await wrapper.find('input[type="search"]').setValue('emile')
    expect(wrapper.emitted('update:query')?.at(-1)).toEqual(['emile'])
    expect(names(wrapper)).toHaveLength(3)

    await wrapper.findAll('th')[0].find('button').trigger('click')
    expect(wrapper.emitted('update:sortKey')?.at(-1)).toEqual(['name'])
    expect(names(wrapper)[0]).toBe('Émile Durand')

    await wrapper.findAll('th')[0].find('button').trigger('click')
    expect(wrapper.emitted('update:sortDir')?.at(-1)).toEqual(['desc'])
    expect(names(wrapper)[0]).toBe('Émile Durand')
    expect(wrapper.find('.dt__pager').exists()).toBe(false)
  })
})
