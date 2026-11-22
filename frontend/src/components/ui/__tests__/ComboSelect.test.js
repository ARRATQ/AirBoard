import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ComboSelect from '../ComboSelect.vue'
import { makeI18n } from './helpers'
import { matchesQuery, personOption } from '@/utils/comboFilter'

const options = [
  { id: 1, label: 'Équipe Réseau', sublabel: '3 membres', color: '#2563eb' },
  { id: 2, label: 'Développeurs', sublabel: '8 membres' },
  { id: 3, label: 'Direction', sublabel: '2 membres' },
]

const mountCombo = (props = {}) =>
  mount(ComboSelect, {
    props: {
      modelValue: [],
      options,
      label: 'Groupes',
      // v-model simulé : la valeur émise redevient la prop
      'onUpdate:modelValue': (value) => wrapper.setProps({ modelValue: value }),
      ...props,
    },
    global: { plugins: [makeI18n()] },
    attachTo: document.body,
  })

let wrapper
const mountWith = (props) => {
  wrapper = mountCombo(props)
  return wrapper
}

const optionLabels = (w) => w.findAll('[role="option"] .combo__option-label').map((o) => o.text())
const chipLabels = (w) => w.findAll('.combo__chip-label').map((c) => c.text())

afterEach(() => {
  wrapper?.unmount()
  wrapper = null
  vi.useRealTimers()
})

describe('matchesQuery', () => {
  it('ignore la casse et les accents, sur le nom et le détail', () => {
    expect(matchesQuery(options[0], 'equipe')).toBe(true)
    expect(matchesQuery(options[1], 'DEVELOP')).toBe(true)
    expect(matchesQuery(options[1], '8 membres')).toBe(true)
    expect(matchesQuery(options[2], 'reseau')).toBe(false)
  })

  it('exige chaque mot saisi', () => {
    expect(matchesQuery(options[0], 'équipe res')).toBe(true)
    expect(matchesQuery(options[0], 'équipe dir')).toBe(false)
  })

  it('nomme une personne par son nom, sinon son identifiant', () => {
    expect(personOption({ id: 4, first_name: 'Ana', last_name: 'Bel', email: 'a@example.test' }))
      .toEqual({ id: 4, label: 'Ana Bel', sublabel: 'a@example.test' })
    expect(personOption({ id: 5, username: 'zed' }).label).toBe('zed')
  })
})

describe('ComboSelect', () => {
  it('affiche les sélectionnés en étiquettes et filtre les propositions', async () => {
    const w = mountWith({ modelValue: [2] })
    expect(chipLabels(w)).toEqual(['Développeurs'])
    const input = w.find('input')
    await input.trigger('focus')
    expect(input.attributes('aria-expanded')).toBe('true')
    expect(optionLabels(w)).toEqual(['Équipe Réseau', 'Développeurs', 'Direction'])
    await input.setValue('dir')
    expect(optionLabels(w)).toEqual(['Direction'])
  })

  it('sélectionne au clavier et retire avec Retour arrière', async () => {
    const w = mountWith()
    const input = w.find('input')
    await input.trigger('focus')
    await input.trigger('keydown', { key: 'ArrowDown' })
    await input.trigger('keydown', { key: 'Enter' })
    expect(w.props('modelValue')).toEqual([2])
    expect(chipLabels(w)).toEqual(['Développeurs'])
    await input.trigger('keydown', { key: 'Backspace' })
    expect(w.props('modelValue')).toEqual([])
  })

  it('un clic sur une option déjà choisie la retire', async () => {
    const w = mountWith({ modelValue: [1] })
    await w.find('input').trigger('focus')
    await w.findAll('[role="option"]')[0].trigger('click')
    expect(w.props('modelValue')).toEqual([])
  })

  it('ne retire pas un sélectionné verrouillé', async () => {
    const w = mountWith({ modelValue: [9, 1], lockedIds: [9], labels: [{ id: 9, label: 'Autre service' }] })
    expect(chipLabels(w)).toEqual(['Autre service', 'Équipe Réseau'])
    expect(w.findAll('.combo__chip-remove')).toHaveLength(1)
    const input = w.find('input')
    await input.trigger('keydown', { key: 'Backspace' })
    await input.trigger('keydown', { key: 'Backspace' })
    expect(w.props('modelValue')).toEqual([9])
  })

  it("Échap ferme la liste sans laisser l'événement fermer la fenêtre", async () => {
    const w = mountWith()
    const input = w.find('input')
    await input.trigger('focus')
    const event = new KeyboardEvent('keydown', { key: 'Escape', cancelable: true, bubbles: true })
    input.element.dispatchEvent(event)
    await w.vm.$nextTick()
    expect(event.defaultPrevented).toBe(true)
    expect(input.attributes('aria-expanded')).toBe('false')
  })

  it('interroge le serveur pendant la saisie et garde le nom du résultat choisi', async () => {
    vi.useFakeTimers()
    const search = vi.fn().mockResolvedValue([{ id: 42, label: 'Zoé Martin', sublabel: 'zoe@example.test' }])
    const w = mountWith({ options: [], search, kind: 'person' })
    const input = w.find('input')
    await input.trigger('focus')
    await input.setValue('zoé')
    await vi.advanceTimersByTimeAsync(300)
    expect(search).toHaveBeenCalledWith('zoé')
    expect(optionLabels(w)).toEqual(['Zoé Martin'])
    await w.findAll('[role="option"]')[0].trigger('click')
    expect(w.props('modelValue')).toEqual([42])
    expect(chipLabels(w)).toEqual(['Zoé Martin'])
    expect(w.find('.combo__avatar').text()).toBe('ZM')
  })
})
