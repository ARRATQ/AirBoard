import Image from '@tiptap/extension-image'
import { VueNodeViewRenderer } from '@tiptap/vue-3'
import ImageResizeView from './ImageResizeView.vue'

/**
 * Extends the base Image extension with:
 * - `width` attribute (stored as percentage string, e.g. "50%")
 * - `align` attribute (left | center | right)
 * - Interactive NodeView with size presets, alignment buttons + drag-to-resize handle
 *
 * The `renderHTML` methods ensure attributes are serialised so `generateHTML()`
 * in TiptapRenderer correctly outputs `style="width: 50%"` and `data-align="center"`.
 */
export const ResizableImage = Image.extend({
  addAttributes() {
    return {
      ...this.parent?.(),
      width: {
        default: null,
        parseHTML: (element) =>
          element.style.width || element.getAttribute('width') || null,
        renderHTML: ({ width }) => {
          if (!width) return {}
          return { style: `width: ${width}` }
        },
      },
      align: {
        default: 'center',
        parseHTML: (element) => element.getAttribute('data-align') || 'center',
        renderHTML: ({ align }) => {
          if (!align) return {}
          return { 'data-align': align }
        },
      },
    }
  },

  addNodeView() {
    return VueNodeViewRenderer(ImageResizeView)
  },
})
