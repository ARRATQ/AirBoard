import Image from '@tiptap/extension-image'
import { VueNodeViewRenderer } from '@tiptap/vue-3'
import ImageResizeView from './ImageResizeView.vue'

/**
 * Extends the base Image extension with:
 * - `width` attribute (stored as percentage string, e.g. "50%")
 * - Interactive NodeView with size presets + drag-to-resize handle
 *
 * The `renderHTML` method ensures the width is serialised as an inline style
 * so `generateHTML()` in TiptapRenderer correctly outputs `style="width: 50%"`.
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
    }
  },

  addNodeView() {
    return VueNodeViewRenderer(ImageResizeView)
  },
})
