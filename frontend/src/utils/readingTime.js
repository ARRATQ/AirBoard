/**
 * Recursively extract plain text from a Tiptap JSON content tree.
 */
const extractText = (node) => {
  if (!node) return ''
  if (node.type === 'text') return node.text || ''
  if (Array.isArray(node.content)) return node.content.map(extractText).join(' ')
  return ''
}

/**
 * Estimate reading time from a Tiptap JSON content object.
 * @param {Object|string} content - Tiptap JSON (object or serialized string)
 * @returns {string|null} e.g. "3 min" or null if no content
 */
export const readingTime = (content) => {
  if (!content) return null
  let obj = content
  if (typeof content === 'string') {
    try { obj = JSON.parse(content) } catch { return null }
  }
  const words = extractText(obj).trim().split(/\s+/).filter(w => w.length > 0).length
  if (words === 0) return null
  const minutes = Math.ceil(words / 200)
  return minutes <= 1 ? '1 min' : `${minutes} min`
}
