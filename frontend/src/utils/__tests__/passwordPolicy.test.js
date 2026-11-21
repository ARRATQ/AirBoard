import { describe, it, expect } from 'vitest'
import { passwordProblem, PASSWORD_MIN_LENGTH } from '../passwordPolicy'

describe('passwordProblem (miroir de utils/auth_security.go)', () => {
  it.each([
    ['', 'tooShort'],
    ['Ab1!xyz', 'tooShort'],
    ['abcdef1!', 'upper'],
    ['ABCDEF1!', 'lower'],
    ['Abcdefg!', 'digit'],
    ['Abcdefg1', 'special'],
    ['Xadmin1!z', 'common'],
    ['MyPassword1!', 'common'],
    ['Zx9!Kq2#', null],
    ['Ete-Lumiere7', null]
  ])('%j → %s', (pw, expected) => {
    expect(passwordProblem(pw)).toBe(expected)
  })

  it('demande au moins 8 caractères comme le backend', () => {
    expect(PASSWORD_MIN_LENGTH).toBe(8)
  })

  it('refuse au-delà de 128 caractères', () => {
    expect(passwordProblem('Aa1!' + 'x'.repeat(125))).toBe('tooLong')
  })
})
