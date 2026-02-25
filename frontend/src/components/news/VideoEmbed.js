import { Node, mergeAttributes } from '@tiptap/core'
import { VueNodeViewRenderer } from '@tiptap/vue-3'
import VideoEmbedView from './VideoEmbedView.vue'

/**
 * Convert a public YouTube or Vimeo URL to its embed URL.
 * Returns null if the URL is not recognized.
 */
export const getEmbedUrl = (url) => {
  if (!url) return null

  // YouTube: various URL forms
  const ytMatch = url.match(
    /(?:youtube\.com\/(?:watch\?v=|shorts\/|embed\/)|youtu\.be\/)([A-Za-z0-9_-]{11})/
  )
  if (ytMatch) return `https://www.youtube-nocookie.com/embed/${ytMatch[1]}`

  // Vimeo
  const vimeoMatch = url.match(/vimeo\.com\/(?:video\/)?(\d+)/)
  if (vimeoMatch) return `https://player.vimeo.com/video/${vimeoMatch[1]}`

  return null
}

export const detectService = (url) => {
  if (!url) return 'unknown'
  if (url.includes('youtu')) return 'youtube'
  if (url.includes('vimeo')) return 'vimeo'
  return 'unknown'
}

export const VideoEmbed = Node.create({
  name: 'videoEmbed',
  group: 'block',
  atom: true,
  draggable: true,

  addAttributes() {
    return {
      src: { default: null },
      service: { default: 'youtube' },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-video-embed]' }]
  },

  // For generateHTML / clipboard: render a responsive iframe wrapper
  renderHTML({ HTMLAttributes }) {
    return [
      'div',
      mergeAttributes({ class: 'video-embed-wrapper', 'data-video-embed': '' }),
      [
        'iframe',
        {
          src: HTMLAttributes.src,
          frameborder: '0',
          allowfullscreen: '',
          allow:
            'accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share',
          loading: 'lazy',
          class: 'video-embed-iframe',
        },
      ],
    ]
  },

  addNodeView() {
    return VueNodeViewRenderer(VideoEmbedView)
  },
})
