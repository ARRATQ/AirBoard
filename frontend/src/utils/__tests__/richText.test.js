import { describe, it, expect } from 'vitest'
import { isRichTextEmpty } from '../richText'

const doc = (...content) => ({ type: 'doc', content })
const p = (...content) => ({ type: 'paragraph', ...(content.length ? { content } : {}) })
const text = (value) => ({ type: 'text', text: value })

describe('isRichTextEmpty', () => {
  it('treats missing or blank content as empty', () => {
    expect(isRichTextEmpty(null)).toBe(true)
    expect(isRichTextEmpty('')).toBe(true)
    expect(isRichTextEmpty(doc())).toBe(true)
    expect(isRichTextEmpty(doc(p()))).toBe(true)
    expect(isRichTextEmpty(doc(p(text('   ')), p({ type: 'hardBreak' })))).toBe(true)
    expect(isRichTextEmpty(doc({ type: 'bulletList', content: [{ type: 'listItem', content: [p()] }] }))).toBe(true)
  })

  it('finds text anywhere in the document', () => {
    expect(isRichTextEmpty(doc(p(text('Bonjour'))))).toBe(false)
    expect(isRichTextEmpty(doc({ type: 'heading', attrs: { level: 2 }, content: [text('Titre')] }))).toBe(false)
  })

  it('counts media without text as content', () => {
    expect(isRichTextEmpty(doc({ type: 'image', attrs: { src: '/a.png' } }))).toBe(false)
  })

  it('accepts serialized JSON and legacy HTML strings', () => {
    expect(isRichTextEmpty(JSON.stringify(doc(p())))).toBe(true)
    expect(isRichTextEmpty(JSON.stringify(doc(p(text('x')))))).toBe(false)
    expect(isRichTextEmpty('<p>  </p>')).toBe(true)
    expect(isRichTextEmpty('<p>Texte</p>')).toBe(false)
    expect(isRichTextEmpty('<p><img src="/a.png"></p>')).toBe(false)
  })
})
