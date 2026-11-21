import { describe, expect, it } from 'vitest'
import { apiErrorMessage } from '../apiError'

const axiosError = (data) => ({ response: { status: 400, data } })

describe('apiErrorMessage', () => {
  it('lit le message du format { error, code }', () => {
    expect(apiErrorMessage(axiosError({ error: 'Tag déjà existant', code: 'conflict' }), 'défaut')).toBe('Tag déjà existant')
  })

  it('lit encore `message` des anciennes réponses', () => {
    expect(apiErrorMessage(axiosError({ message: 'Ancien format' }), 'défaut')).toBe('Ancien format')
  })

  it('préfère `error` à `message`', () => {
    expect(apiErrorMessage(axiosError({ error: 'Nouveau', message: 'Ancien' }), 'défaut')).toBe('Nouveau')
  })

  it.each([
    ['sans réponse (erreur réseau)', new Error('Network Error')],
    ['corps vide', axiosError(undefined)],
    ['message vide', axiosError({ error: '  ' })],
    ['message non textuel', axiosError({ error: { detail: 1 } })],
    ['erreur absente', undefined],
  ])('retombe sur le message par défaut : %s', (_, error) => {
    expect(apiErrorMessage(error, 'défaut')).toBe('défaut')
  })
})
