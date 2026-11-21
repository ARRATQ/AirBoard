// Contenu de l'éditeur riche (Tiptap) : JSON { type: 'doc', content: [...] },
// sa forme sérialisée, ou du HTML pour les anciens articles.

// Nœuds qui ne font que structurer : vides, ils ne constituent pas un contenu.
const STRUCTURAL_NODES = new Set([
  'doc', 'paragraph', 'heading', 'text', 'hardBreak',
  'bulletList', 'orderedList', 'listItem', 'blockquote',
])

const hasContent = (node) => {
  if (!node || typeof node !== 'object') return false
  if (node.type === 'text') return Boolean(node.text?.trim())
  if (node.type && !STRUCTURAL_NODES.has(node.type)) return true // image, vidéo, tableau…
  return Array.isArray(node.content) && node.content.some(hasContent)
}

const htmlIsEmpty = (html) => {
  if (/<(img|video|iframe|table|hr)\b/i.test(html)) return false
  return !html.replace(/<[^>]*>/g, '').replace(/&nbsp;/g, ' ').trim()
}

/**
 * Le contenu ne contient ni texte ni média.
 * @param {object|string|null|undefined} content
 * @returns {boolean}
 */
export function isRichTextEmpty(content) {
  if (!content) return true
  if (typeof content === 'string') {
    try {
      return !hasContent(JSON.parse(content))
    } catch {
      return htmlIsEmpty(content)
    }
  }
  return !hasContent(content)
}
