// Garde-fous des traductions : les quatre langues ont les mêmes clés, et toute clé littérale
// utilisée dans le code existe dans la langue de référence (fr).
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative, resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const SRC = resolve(__dirname, '../..')
const LOCALES = ['fr', 'en', 'es', 'ar']
const REFERENCE = 'fr'

const flatten = (obj, prefix = '', out = {}) => {
  for (const [key, value] of Object.entries(obj)) {
    const path = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object') flatten(value, path, out)
    else out[path] = value
  }
  return out
}

const messages = Object.fromEntries(
  LOCALES.map((locale) => [
    locale,
    flatten(JSON.parse(readFileSync(join(SRC, 'locales', `${locale}.json`), 'utf8'))),
  ]),
)

const sourceFiles = (dir) =>
  readdirSync(dir).flatMap((name) => {
    const path = join(dir, name)
    if (statSync(path).isDirectory()) return name === '__tests__' ? [] : sourceFiles(path)
    return /\.(vue|js)$/.test(name) && !name.endsWith('.test.js') ? [path] : []
  })

// t('a.b'), $t("a.b"), i18n.global.t('a.b'), te('a.b') — clés littérales uniquement.
const KEY_CALL = /(?<![\w$])(?:\$t|\$te|t|te|tc|\$tc)\(\s*(['"`])([A-Za-z][\w-]*(?:\.[\w-]+)+)\1/g

const usedKeys = () => {
  const found = new Map()
  for (const file of sourceFiles(SRC)) {
    const text = readFileSync(file, 'utf8')
    for (const match of text.matchAll(KEY_CALL)) {
      if (!found.has(match[2])) found.set(match[2], relative(SRC, file))
    }
  }
  return found
}

describe('fichiers de langue', () => {
  const referenceKeys = Object.keys(messages[REFERENCE])

  it.each(LOCALES.filter((l) => l !== REFERENCE))('%s a toutes les clés de fr', (locale) => {
    const missing = referenceKeys.filter((key) => !(key in messages[locale]))
    expect(missing).toEqual([])
  })

  it.each(LOCALES.filter((l) => l !== REFERENCE))('%s n’a pas de clé absente de fr', (locale) => {
    const extra = Object.keys(messages[locale]).filter((key) => !(key in messages[REFERENCE]))
    expect(extra).toEqual([])
  })

  it.each(LOCALES)('%s n’a aucune traduction vide', (locale) => {
    const empty = Object.entries(messages[locale])
      .filter(([, value]) => typeof value === 'string' && value.trim() === '')
      .map(([key]) => key)
    expect(empty).toEqual([])
  })

  it.each(LOCALES)('%s ne fabrique pas de pluriel avec « (s) » (utiliser les formes « | »)', (locale) => {
    const fake = Object.entries(messages[locale])
      .filter(([, value]) => typeof value === 'string' && /\w\((?:s|e|es|x)\)/.test(value))
      .map(([key]) => key)
    expect(fake).toEqual([])
  })

  it('les pluriels arabes ont les six formes CLDR (zéro | un | deux | 3-10 | 11-99 | 100+)', () => {
    const partial = Object.entries(messages.ar)
      .filter(([, value]) => typeof value === 'string' && value.includes('|'))
      .filter(([, value]) => value.split('|').length !== 6 || value.split('|').some((form) => form.trim() === ''))
      .map(([key]) => key)
    expect(partial).toEqual([])
  })

  it('toute clé littérale utilisée dans src/ existe en fr', () => {
    const missing = [...usedKeys()]
      .filter(([key]) => !(key in messages[REFERENCE]))
      .map(([key, file]) => `${key} (${file})`)
    expect(missing).toEqual([])
  })
})
