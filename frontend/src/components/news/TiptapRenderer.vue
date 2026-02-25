<template>
  <div class="tiptap-content" v-html="sanitizedContent"></div>
</template>

<script setup>
import { computed } from 'vue'
import { generateHTML } from '@tiptap/html'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import { ResizableImage } from './ResizableImage.js'
import { CalloutExtension } from './CalloutExtension.js'
import { VideoEmbed } from './VideoEmbed.js'
import Underline from '@tiptap/extension-underline'
import { Table } from '@tiptap/extension-table'
import { TableRow } from '@tiptap/extension-table-row'
import { TableCell } from '@tiptap/extension-table-cell'
import { TableHeader } from '@tiptap/extension-table-header'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import TextAlign from '@tiptap/extension-text-align'
import { TextStyle } from '@tiptap/extension-text-style'
import Color from '@tiptap/extension-color'
import Highlight from '@tiptap/extension-highlight'
import { createLowlight, common } from 'lowlight'
import DOMPurify from 'dompurify'

// Create lowlight instance with common languages
const lowlight = createLowlight(common)

const props = defineProps({
  content: {
    type: [Object, String],
    required: true
  }
})

const renderedContent = computed(() => {
  if (!props.content) return ''

  let contentObj = props.content

  // If content is a string, try to parse it as JSON
  if (typeof props.content === 'string') {
    try {
      contentObj = JSON.parse(props.content)
    } catch (e) {
      // If parsing fails, treat it as plain text and escape it
      const escapedText = props.content.replace(/</g, '&lt;').replace(/>/g, '&gt;')
      return `<p>${escapedText}</p>`
    }
  }

  // Generate HTML from Tiptap JSON
  return generateHTML(contentObj, [
    StarterKit.configure({
      codeBlock: false
    }),
    Link.configure({
      openOnClick: false
    }),
    ResizableImage.configure({
      inline: false,
    }),
    CalloutExtension,
    VideoEmbed,
    Underline,
    Table,
    TableRow,
    TableCell,
    TableHeader,
    CodeBlockLowlight.configure({
      lowlight
    }),
    TextAlign.configure({
      types: ['heading', 'paragraph']
    }),
    TextStyle,
    Color,
    Highlight.configure({
      multicolor: true,
    }),
  ])
})

// Sanitize HTML content to prevent XSS attacks
const sanitizedContent = computed(() => {
  return DOMPurify.sanitize(renderedContent.value, {
    ALLOWED_TAGS: [
      'p', 'br', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
      'ul', 'ol', 'li', 'strong', 'em', 'u', 's', 'del',
      'a', 'code', 'pre', 'blockquote', 'hr',
      'table', 'thead', 'tbody', 'tr', 'th', 'td',
      'img', 'span', 'div', 'iframe'
    ],
    ALLOWED_ATTR: [
      'href', 'class', 'target', 'rel', 'src', 'alt',
      'width', 'height', 'style', 'data-*',
      'frameborder', 'allowfullscreen', 'allow', 'loading', 'scrolling'
    ],
    ALLOWED_URI_REGEXP: /^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|data):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i,
    ALLOW_DATA_ATTR: true,
    ADD_ATTR: ['target'],
    // Force target="_blank" and rel="noopener noreferrer" on all links
    HOOK_AFTER_SANITIZE_ATTRIBUTES: (node) => {
      if (node.tagName === 'A') {
        node.setAttribute('target', '_blank')
        node.setAttribute('rel', 'noopener noreferrer')
      }
    }
  })
})
</script>

<style>
.tiptap-content {
  @apply text-gray-800 dark:text-gray-200;
}

.tiptap-content h1 {
  @apply text-3xl font-bold mb-4 text-gray-900 dark:text-white;
}

.tiptap-content h2 {
  @apply text-2xl font-bold mb-3 text-gray-900 dark:text-white;
}

.tiptap-content h3 {
  @apply text-xl font-bold mb-2 text-gray-900 dark:text-white;
}

.tiptap-content p {
  @apply mb-4 leading-relaxed;
}

.tiptap-content ul {
  @apply list-disc ml-6 mb-4 space-y-1;
}

.tiptap-content ol {
  @apply list-decimal ml-6 mb-4 space-y-1;
}

.tiptap-content li {
  @apply ml-2;
}

.tiptap-content ul ul,
.tiptap-content ol ol {
  @apply ml-6 mb-0 mt-1;
}

.tiptap-content a {
  @apply text-primary-500 hover:text-primary-600 underline;
}

.tiptap-content blockquote {
  @apply border-l-4 border-gray-300 dark:border-gray-600 pl-4 italic my-4 text-gray-700 dark:text-gray-300;
}

.tiptap-content code {
  @apply bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-sm font-mono text-red-600 dark:text-red-400;
}

.tiptap-content pre {
  @apply bg-gray-900 dark:bg-gray-950 text-gray-100 p-4 rounded-lg overflow-x-auto mb-4;
}

.tiptap-content pre code {
  @apply bg-transparent text-gray-100 p-0;
}

.tiptap-content table {
  @apply w-full border-collapse mb-4;
}

.tiptap-content th {
  @apply bg-gray-100 dark:bg-gray-800 border border-gray-300 dark:border-gray-600 px-4 py-2 text-left font-semibold;
}

.tiptap-content td {
  @apply border border-gray-300 dark:border-gray-600 px-4 py-2;
}

.tiptap-content hr {
  @apply border-gray-300 dark:border-gray-600 my-6;
}

.tiptap-content strong {
  @apply font-bold;
}

.tiptap-content em {
  @apply italic;
}

.tiptap-content img {
  @apply max-w-full h-auto rounded-lg my-4;
}

/* ── Callout blocks ──────────────────────────────── */
.tiptap-content [data-callout-type] {
  border-left: 4px solid;
  border-radius: 8px;
  padding: 12px 16px;
  margin: 1rem 0;
}
.tiptap-content [data-callout-type="info"]    { background: #EFF6FF; border-color: #3B82F6; }
.tiptap-content [data-callout-type="warning"] { background: #FFFBEB; border-color: #F59E0B; }
.tiptap-content [data-callout-type="success"] { background: #F0FDF4; border-color: #22C55E; }
.tiptap-content [data-callout-type="danger"]  { background: #FEF2F2; border-color: #EF4444; }

.dark .tiptap-content [data-callout-type="info"]    { background: #1E3A5F; border-color: #3B82F6; }
.dark .tiptap-content [data-callout-type="warning"] { background: #3D2C00; border-color: #F59E0B; }
.dark .tiptap-content [data-callout-type="success"] { background: #052E16; border-color: #22C55E; }
.dark .tiptap-content [data-callout-type="danger"]  { background: #3B0A0A; border-color: #EF4444; }

.tiptap-content [data-callout-type] p:last-child { margin-bottom: 0; }

/* ── Video embed ─────────────────────────────────── */
.tiptap-content .video-embed-wrapper {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 10px;
  overflow: hidden;
  margin: 1.5rem 0;
  background: #000;
}
.tiptap-content .video-embed-iframe {
  width: 100%;
  height: 100%;
  border: none;
  display: block;
}
</style>
