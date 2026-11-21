import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { generateHTML } from '@tiptap/html'
import StarterKit from '@tiptap/starter-kit'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import { createLowlight, common } from 'lowlight'
import TiptapRenderer from '../TiptapRenderer.vue'

const doc = {
  type: 'doc',
  content: [
    { type: 'paragraph', content: [{ type: 'text', text: 'Bonjour' }] },
    { type: 'codeBlock', attrs: { language: 'javascript' }, content: [{ type: 'text', text: 'const a = 1 < 2' }] },
    { type: 'codeBlock', content: [{ type: 'text', text: 'SELECT 1;' }] },
    { type: 'image', attrs: { src: '/uploads/a.png', alt: 'a' } },
  ],
}

describe('TiptapRenderer', () => {
  it('rend les blocs de code comme l’ancienne configuration lowlight (generateHTML ne colore pas)', () => {
    const codeOnly = { ...doc, content: doc.content.slice(0, 3) }
    const withLowlight = generateHTML(codeOnly, [
      StarterKit.configure({ codeBlock: false }),
      CodeBlockLowlight.configure({ lowlight: createLowlight(common) }),
    ])
    const plain = generateHTML(codeOnly, [StarterKit])
    expect(plain).toBe(withLowlight)
    expect(plain).toMatch(/<pre[^>]*><code class="language-javascript">const a = 1 &lt; 2<\/code><\/pre>/)
  })

  it('affiche le contenu et charge les images en différé', () => {
    const wrapper = mount(TiptapRenderer, { props: { content: doc } })
    expect(wrapper.html()).toContain('language-javascript')
    const img = wrapper.find('img')
    expect(img.attributes('src')).toBe('/uploads/a.png')
    expect(img.attributes('loading')).toBe('lazy')
  })
  it('ouvre tous les liens dans un nouvel onglet, sans accès à la page d’origine', () => {
    // Un lien enregistré sans target (target: null) sortait tel quel de generateHTML.
    const withLink = {
      type: 'doc',
      content: [{
        type: 'paragraph',
        content: [{ type: 'text', text: 'site', marks: [{ type: 'link', attrs: { href: 'https://example.com', target: null, rel: null } }] }],
      }],
    }
    const link = mount(TiptapRenderer, { props: { content: withLink } }).find('a')
    expect(link.attributes('href')).toBe('https://example.com')
    expect(link.attributes('target')).toBe('_blank')
    expect(link.attributes('rel').split(' ')).toEqual(expect.arrayContaining(['noopener', 'noreferrer']))
  })
})
