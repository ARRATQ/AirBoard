import { Node, mergeAttributes } from '@tiptap/core'
import { VueNodeViewRenderer } from '@tiptap/vue-3'
import CalloutView from './CalloutView.vue'

export const CALLOUT_TYPES = [
  { value: 'info',    label: 'Info',      icon: 'mdi:information',   color: '#3B82F6' },
  { value: 'warning', label: 'Attention', icon: 'mdi:alert',          color: '#F59E0B' },
  { value: 'success', label: 'Succès',    icon: 'mdi:check-circle',   color: '#22C55E' },
  { value: 'danger',  label: 'Danger',    icon: 'mdi:alert-circle',   color: '#EF4444' },
]

export const CalloutExtension = Node.create({
  name: 'callout',
  group: 'block',
  content: 'block+',
  defining: true,

  addAttributes() {
    return {
      type: {
        default: 'info',
        parseHTML: element => element.getAttribute('data-callout-type') || 'info',
        renderHTML: ({ type }) => ({ 'data-callout-type': type }),
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-callout-type]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return ['div', mergeAttributes({ class: 'callout-block' }, HTMLAttributes), 0]
  },

  addNodeView() {
    return VueNodeViewRenderer(CalloutView)
  },
})
