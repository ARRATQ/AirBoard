<template>
  <div class="rich-text-editor">
    <!-- Toolbar (sticky) -->
    <div v-if="editor" class="editor-toolbar">
      <!-- Text formatting -->
      <div class="toolbar-group">
        <button
          type="button"
          @click="editor.chain().focus().toggleBold().run()"
          :class="{ 'is-active': editor.isActive('bold') }"
          class="toolbar-button"
          title="Gras (Ctrl+B)"
        >
          <Icon icon="mdi:format-bold" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleItalic().run()"
          :class="{ 'is-active': editor.isActive('italic') }"
          class="toolbar-button"
          title="Italique (Ctrl+I)"
        >
          <Icon icon="mdi:format-italic" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleUnderline().run()"
          :class="{ 'is-active': editor.isActive('underline') }"
          class="toolbar-button"
          title="Souligné (Ctrl+U)"
        >
          <Icon icon="mdi:format-underline" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleStrike().run()"
          :class="{ 'is-active': editor.isActive('strike') }"
          class="toolbar-button"
          title="Barré"
        >
          <Icon icon="mdi:format-strikethrough" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleCode().run()"
          :class="{ 'is-active': editor.isActive('code') }"
          class="toolbar-button"
          title="Code inline"
        >
          <Icon icon="mdi:code-tags" class="h-5 w-5" />
        </button>
      </div>

      <!-- Color & Highlight -->
      <div class="toolbar-group">
        <!-- Text color -->
        <div class="color-picker-wrapper" title="Couleur du texte">
          <button type="button" class="toolbar-button color-button" @click="toggleColorPicker">
            <Icon icon="mdi:format-color-text" class="h-5 w-5" />
            <span class="color-swatch" :style="{ background: activeTextColor }"></span>
          </button>
          <div v-if="showColorPicker" class="color-dropdown" @mouseleave="closeColorPicker">
            <div class="color-grid">
              <button
                v-for="color in textColors"
                :key="color.value"
                type="button"
                class="color-dot"
                :style="{ background: color.value }"
                :title="color.name"
                @click="setTextColor(color.value)"
              ></button>
            </div>
            <button type="button" class="color-remove-btn" @click="removeTextColor">
              <Icon icon="mdi:format-color-reset" class="h-4 w-4" /> Effacer
            </button>
          </div>
        </div>

        <!-- Highlight color -->
        <div class="color-picker-wrapper" title="Surbrillance">
          <button type="button" class="toolbar-button color-button" @click="toggleHighlightPicker">
            <Icon icon="mdi:marker" class="h-5 w-5" />
            <span class="color-swatch" :style="{ background: activeHighlightColor }"></span>
          </button>
          <div v-if="showHighlightPicker" class="color-dropdown" @mouseleave="closeHighlightPicker">
            <div class="color-grid">
              <button
                v-for="color in highlightColors"
                :key="color.value"
                type="button"
                class="color-dot"
                :style="{ background: color.value }"
                :title="color.name"
                @click="setHighlight(color.value)"
              ></button>
            </div>
            <button type="button" class="color-remove-btn" @click="removeHighlight">
              <Icon icon="mdi:format-color-reset" class="h-4 w-4" /> Effacer
            </button>
          </div>
        </div>
      </div>

      <!-- Headings -->
      <div class="toolbar-group">
        <button
          type="button"
          @click="editor.chain().focus().toggleHeading({ level: 1 }).run()"
          :class="{ 'is-active': editor.isActive('heading', { level: 1 }) }"
          class="toolbar-button"
          title="Titre 1"
        >
          <Icon icon="mdi:format-header-1" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
          :class="{ 'is-active': editor.isActive('heading', { level: 2 }) }"
          class="toolbar-button"
          title="Titre 2"
        >
          <Icon icon="mdi:format-header-2" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleHeading({ level: 3 }).run()"
          :class="{ 'is-active': editor.isActive('heading', { level: 3 }) }"
          class="toolbar-button"
          title="Titre 3"
        >
          <Icon icon="mdi:format-header-3" class="h-5 w-5" />
        </button>
      </div>

      <!-- Text Alignment -->
      <div class="toolbar-group">
        <button
          type="button"
          @click="editor.chain().focus().setTextAlign('left').run()"
          :class="{ 'is-active': editor.isActive({ textAlign: 'left' }) }"
          class="toolbar-button"
          title="Aligner à gauche"
        >
          <Icon icon="mdi:format-align-left" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().setTextAlign('center').run()"
          :class="{ 'is-active': editor.isActive({ textAlign: 'center' }) }"
          class="toolbar-button"
          title="Centrer"
        >
          <Icon icon="mdi:format-align-center" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().setTextAlign('right').run()"
          :class="{ 'is-active': editor.isActive({ textAlign: 'right' }) }"
          class="toolbar-button"
          title="Aligner à droite"
        >
          <Icon icon="mdi:format-align-right" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().setTextAlign('justify').run()"
          :class="{ 'is-active': editor.isActive({ textAlign: 'justify' }) }"
          class="toolbar-button"
          title="Justifier"
        >
          <Icon icon="mdi:format-align-justify" class="h-5 w-5" />
        </button>
      </div>

      <!-- Lists -->
      <div class="toolbar-group">
        <button
          type="button"
          @click="editor.chain().focus().toggleBulletList().run()"
          :class="{ 'is-active': editor.isActive('bulletList') }"
          class="toolbar-button"
          title="Liste à puces"
        >
          <Icon icon="mdi:format-list-bulleted" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleOrderedList().run()"
          :class="{ 'is-active': editor.isActive('orderedList') }"
          class="toolbar-button"
          title="Liste numérotée"
        >
          <Icon icon="mdi:format-list-numbered" class="h-5 w-5" />
        </button>
      </div>

      <!-- Blocks -->
      <div class="toolbar-group">
        <button
          type="button"
          @click="editor.chain().focus().toggleCodeBlock().run()"
          :class="{ 'is-active': editor.isActive('codeBlock') }"
          class="toolbar-button"
          title="Bloc de code"
        >
          <Icon icon="mdi:code-braces" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleBlockquote().run()"
          :class="{ 'is-active': editor.isActive('blockquote') }"
          class="toolbar-button"
          title="Citation"
        >
          <Icon icon="mdi:format-quote-close" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().setHorizontalRule().run()"
          class="toolbar-button"
          title="Séparateur horizontal"
        >
          <Icon icon="mdi:minus" class="h-5 w-5" />
        </button>

        <!-- Callout -->
        <div class="color-picker-wrapper" title="Callout">
          <button
            type="button"
            @click="toggleCalloutMenu"
            :class="{ 'is-active': editor.isActive('callout') }"
            class="toolbar-button"
          >
            <Icon icon="mdi:message-alert-outline" class="h-5 w-5" />
          </button>
          <div v-if="showCalloutMenu" class="callout-type-dropdown" @mouseleave="closeCalloutMenu">
            <button
              v-for="t in calloutTypes"
              :key="t.value"
              type="button"
              class="callout-type-item"
              @click="insertCallout(t.value)"
            >
              <Icon :icon="t.icon" class="h-4 w-4" :style="{ color: t.color }" />
              <span>{{ t.label }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Table -->
      <div class="toolbar-group">
        <div class="color-picker-wrapper" title="Tableau">
          <button
            type="button"
            @click="toggleTableMenu"
            :class="{ 'is-active': editor.isActive('table') }"
            class="toolbar-button"
          >
            <Icon icon="mdi:table" class="h-5 w-5" />
          </button>
          <div v-if="showTableMenu" class="table-dropdown" @mouseleave="closeTableMenu">
            <button type="button" class="table-menu-item" @click="insertTable">
              <Icon icon="mdi:table-plus" class="h-4 w-4" /> Insérer un tableau
            </button>
            <template v-if="editor.isActive('table')">
              <hr class="table-menu-divider" />
              <button type="button" class="table-menu-item" @click="editor.chain().focus().addColumnBefore().run()">
                <Icon icon="mdi:table-column-plus-before" class="h-4 w-4" /> Colonne avant
              </button>
              <button type="button" class="table-menu-item" @click="editor.chain().focus().addColumnAfter().run()">
                <Icon icon="mdi:table-column-plus-after" class="h-4 w-4" /> Colonne après
              </button>
              <button type="button" class="table-menu-item" @click="editor.chain().focus().deleteColumn().run()">
                <Icon icon="mdi:table-column-remove" class="h-4 w-4" /> Supprimer colonne
              </button>
              <hr class="table-menu-divider" />
              <button type="button" class="table-menu-item" @click="editor.chain().focus().addRowBefore().run()">
                <Icon icon="mdi:table-row-plus-before" class="h-4 w-4" /> Ligne avant
              </button>
              <button type="button" class="table-menu-item" @click="editor.chain().focus().addRowAfter().run()">
                <Icon icon="mdi:table-row-plus-after" class="h-4 w-4" /> Ligne après
              </button>
              <button type="button" class="table-menu-item" @click="editor.chain().focus().deleteRow().run()">
                <Icon icon="mdi:table-row-remove" class="h-4 w-4" /> Supprimer ligne
              </button>
              <hr class="table-menu-divider" />
              <button type="button" class="table-menu-item table-menu-danger" @click="editor.chain().focus().deleteTable().run()">
                <Icon icon="mdi:table-remove" class="h-4 w-4" /> Supprimer tableau
              </button>
            </template>
          </div>
        </div>
      </div>

      <!-- Link -->
      <div class="toolbar-group">
        <button
          type="button"
          @click="setLink"
          :class="{ 'is-active': editor.isActive('link') }"
          class="toolbar-button"
          title="Ajouter un lien"
        >
          <Icon icon="mdi:link-variant" class="h-5 w-5" />
        </button>
        <button
          v-if="editor.isActive('link')"
          type="button"
          @click="editor.chain().focus().unsetLink().run()"
          class="toolbar-button"
          title="Supprimer le lien"
        >
          <Icon icon="mdi:link-variant-off" class="h-5 w-5" />
        </button>
      </div>

      <!-- Media -->
      <div class="toolbar-group">
        <button
          type="button"
          @click="showMediaModal = true"
          class="toolbar-button"
          title="Insérer une image"
        >
          <Icon icon="mdi:image-plus" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="showFileModal = true"
          class="toolbar-button"
          title="Joindre un fichier"
        >
          <Icon icon="mdi:file-plus" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="insertVideo"
          class="toolbar-button"
          title="Intégrer une vidéo YouTube / Vimeo"
        >
          <Icon icon="mdi:video-plus" class="h-5 w-5" />
        </button>
      </div>

      <!-- Undo/Redo -->
      <div class="toolbar-group ml-auto">
        <button
          type="button"
          @click="editor.chain().focus().undo().run()"
          :disabled="!editor.can().undo()"
          class="toolbar-button"
          title="Annuler (Ctrl+Z)"
        >
          <Icon icon="mdi:undo" class="h-5 w-5" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().redo().run()"
          :disabled="!editor.can().redo()"
          class="toolbar-button"
          title="Rétablir (Ctrl+Y)"
        >
          <Icon icon="mdi:redo" class="h-5 w-5" />
        </button>
      </div>
    </div>

    <!-- Editor content -->
    <EditorContent :editor="editor" class="editor-content" />

    <!-- Media Modal -->
    <teleport to="body">
      <div v-if="showMediaModal" class="modal-overlay" @click.self="showMediaModal = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3 class="modal-title">Insérer une image</h3>
            <button @click="showMediaModal = false" class="modal-close">
              <Icon icon="mdi:close" class="h-6 w-6" />
            </button>
          </div>
          <div class="modal-body">
            <div class="tabs">
              <button
                @click="mediaTab = 'upload'"
                :class="{ 'active': mediaTab === 'upload' }"
                class="tab-button"
              >
                <Icon icon="mdi:upload" class="h-5 w-5" />
                Téléverser
              </button>
              <button
                @click="mediaTab = 'gallery'"
                :class="{ 'active': mediaTab === 'gallery' }"
                class="tab-button"
              >
                <Icon icon="mdi:image-multiple" class="h-5 w-5" />
                Médiathèque
              </button>
            </div>

            <div v-if="mediaTab === 'upload'" class="tab-content">
              <MediaUploader
                accept="image/*"
                :max-size-m-b="5"
                :as-group-admin="asGroupAdmin"
                @upload-success="handleMediaUpload"
              />
            </div>
            <div v-else class="tab-content">
              <MediaGallery
                @select="insertImage"
                @cancel="showMediaModal = false"
              />
            </div>
          </div>
        </div>
      </div>
    </teleport>

    <!-- File Modal -->
    <teleport to="body">
      <div v-if="showFileModal" class="modal-overlay" @click.self="showFileModal = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3 class="modal-title">Joindre un fichier</h3>
            <button @click="showFileModal = false" class="modal-close">
              <Icon icon="mdi:close" class="h-6 w-6" />
            </button>
          </div>
          <div class="modal-body">
            <div class="tabs">
              <button
                @click="fileTab = 'upload'"
                :class="{ 'active': fileTab === 'upload' }"
                class="tab-button"
              >
                <Icon icon="mdi:upload" class="h-5 w-5" />
                Téléverser
              </button>
              <button
                @click="fileTab = 'gallery'"
                :class="{ 'active': fileTab === 'gallery' }"
                class="tab-button"
              >
                <Icon icon="mdi:file-multiple" class="h-5 w-5" />
                Médiathèque
              </button>
            </div>

            <div v-if="fileTab === 'upload'" class="tab-content">
              <MediaUploader
                accept="application/pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.zip,.rar"
                :max-size-m-b="10"
                :as-group-admin="asGroupAdmin"
                @upload-success="handleFileUpload"
              />
            </div>
            <div v-else class="tab-content">
              <MediaGallery
                @select="insertFile"
                @cancel="showFileModal = false"
              />
            </div>
          </div>
        </div>
      </div>
    </teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import { ResizableImage } from './ResizableImage.js'
import { CalloutExtension, CALLOUT_TYPES } from './CalloutExtension.js'
import { VideoEmbed, getEmbedUrl, detectService } from './VideoEmbed.js'
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
import Placeholder from '@tiptap/extension-placeholder'
import { createLowlight } from 'lowlight'
import { Icon } from '@iconify/vue'
import MediaUploader from '@/components/media/MediaUploader.vue'
import MediaGallery from '@/components/media/MediaGallery.vue'

// Import common languages for syntax highlighting
import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import python from 'highlight.js/lib/languages/python'
import java from 'highlight.js/lib/languages/java'
import bash from 'highlight.js/lib/languages/bash'
import json from 'highlight.js/lib/languages/json'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'

// Create lowlight instance and register languages
const lowlight = createLowlight()
lowlight.register('javascript', javascript)
lowlight.register('typescript', typescript)
lowlight.register('python', python)
lowlight.register('java', java)
lowlight.register('bash', bash)
lowlight.register('json', json)
lowlight.register('xml', xml)
lowlight.register('css', css)

const props = defineProps({
  modelValue: {
    type: [String, Object],
    default: ''
  },
  placeholder: {
    type: String,
    default: 'Commencez à écrire...'
  },
  editable: {
    type: Boolean,
    default: true
  },
  asGroupAdmin: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue'])

const showMediaModal = ref(false)
const showFileModal = ref(false)
const mediaTab = ref('upload')
const fileTab = ref('upload')
const showColorPicker = ref(false)
const showHighlightPicker = ref(false)
const showTableMenu = ref(false)
const showCalloutMenu = ref(false)
const calloutTypes = CALLOUT_TYPES

// Palette de couleurs pour le texte
const textColors = [
  { name: 'Noir', value: '#000000' },
  { name: 'Gris foncé', value: '#374151' },
  { name: 'Gris', value: '#6B7280' },
  { name: 'Rouge', value: '#EF4444' },
  { name: 'Orange', value: '#F97316' },
  { name: 'Jaune', value: '#EAB308' },
  { name: 'Vert', value: '#22C55E' },
  { name: 'Cyan', value: '#06B6D4' },
  { name: 'Bleu', value: '#3B82F6' },
  { name: 'Indigo', value: '#6366F1' },
  { name: 'Violet', value: '#A855F7' },
  { name: 'Rose', value: '#EC4899' },
]

// Palette de couleurs pour la surbrillance
const highlightColors = [
  { name: 'Jaune', value: '#FEF08A' },
  { name: 'Vert clair', value: '#BBF7D0' },
  { name: 'Bleu clair', value: '#BAE6FD' },
  { name: 'Rose clair', value: '#FBCFE8' },
  { name: 'Orange clair', value: '#FED7AA' },
  { name: 'Violet clair', value: '#DDD6FE' },
  { name: 'Rouge clair', value: '#FECACA' },
  { name: 'Cyan clair', value: '#A5F3FC' },
]

const activeTextColor = computed(() => {
  if (!editor.value) return 'transparent'
  const color = editor.value.getAttributes('textStyle').color
  return color || 'transparent'
})

const activeHighlightColor = computed(() => {
  if (!editor.value) return 'transparent'
  const color = editor.value.getAttributes('highlight').color
  return color || 'transparent'
})

const toggleColorPicker = () => {
  showColorPicker.value = !showColorPicker.value
  showHighlightPicker.value = false
  showTableMenu.value = false
}

const closeColorPicker = () => {
  showColorPicker.value = false
}

const toggleHighlightPicker = () => {
  showHighlightPicker.value = !showHighlightPicker.value
  showColorPicker.value = false
  showTableMenu.value = false
}

const closeHighlightPicker = () => {
  showHighlightPicker.value = false
}

const toggleTableMenu = () => {
  showTableMenu.value = !showTableMenu.value
  showColorPicker.value = false
  showHighlightPicker.value = false
}

const closeTableMenu = () => {
  showTableMenu.value = false
}

const toggleCalloutMenu = () => {
  showCalloutMenu.value = !showCalloutMenu.value
  showColorPicker.value = false
  showHighlightPicker.value = false
  showTableMenu.value = false
}

const closeCalloutMenu = () => {
  showCalloutMenu.value = false
}

const insertCallout = (type) => {
  editor.value.chain().focus().insertContent({
    type: 'callout',
    attrs: { type },
    content: [{ type: 'paragraph' }],
  }).run()
  showCalloutMenu.value = false
}

const insertVideo = () => {
  const url = window.prompt('URL YouTube ou Vimeo')
  if (!url) return
  const embedUrl = getEmbedUrl(url)
  if (!embedUrl) {
    window.alert('URL non reconnue. Utilisez une URL YouTube ou Vimeo.')
    return
  }
  editor.value.chain().focus().insertContent({
    type: 'videoEmbed',
    attrs: { src: embedUrl, service: detectService(url) },
  }).run()
}

const setTextColor = (color) => {
  editor.value.chain().focus().setColor(color).run()
  showColorPicker.value = false
}

const removeTextColor = () => {
  editor.value.chain().focus().unsetColor().run()
  showColorPicker.value = false
}

const setHighlight = (color) => {
  editor.value.chain().focus().setHighlight({ color }).run()
  showHighlightPicker.value = false
}

const removeHighlight = () => {
  editor.value.chain().focus().unsetHighlight().run()
  showHighlightPicker.value = false
}

const insertTable = () => {
  editor.value.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()
  showTableMenu.value = false
}

const editor = useEditor({
  content: props.modelValue,
  editable: props.editable,
  extensions: [
    StarterKit.configure({
      codeBlock: false,
    }),
    Link.configure({
      openOnClick: false,
      HTMLAttributes: {
        class: 'text-blue-600 dark:text-blue-400 underline hover:text-blue-800 dark:hover:text-blue-300',
      },
    }),
    ResizableImage.configure({
      inline: false,
    }),
    CalloutExtension,
    VideoEmbed,
    Underline,
    Table.configure({
      resizable: true,
    }),
    TableRow,
    TableCell,
    TableHeader,
    CodeBlockLowlight.configure({
      lowlight,
    }),
    TextAlign.configure({
      types: ['heading', 'paragraph'],
    }),
    TextStyle,
    Color,
    Highlight.configure({
      multicolor: true,
    }),
    Placeholder.configure({
      placeholder: props.placeholder,
    }),
  ],
  onUpdate: ({ editor }) => {
    emit('update:modelValue', editor.getJSON())
  },
  editorProps: {
    attributes: {
      class: 'prose prose-sm sm:prose lg:prose-lg xl:prose-xl dark:prose-invert max-w-none focus:outline-none min-h-[300px] p-4',
    },
  },
})

// Set link
const setLink = () => {
  const previousUrl = editor.value.getAttributes('link').href
  const url = window.prompt('URL', previousUrl)

  if (url === null) {
    return
  }

  if (url === '') {
    editor.value.chain().focus().extendMarkRange('link').unsetLink().run()
    return
  }

  editor.value.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
}

// Watch for external content changes
watch(() => props.modelValue, (value) => {
  if (editor.value) {
    const isSame = JSON.stringify(editor.value.getJSON()) === JSON.stringify(value)
    if (!isSame) {
      editor.value.commands.setContent(value, false)
    }
  }
})

// Watch for editable changes
watch(() => props.editable, (value) => {
  if (editor.value) {
    editor.value.setEditable(value)
  }
})

// Media handlers
const handleMediaUpload = (media) => {
  if (media.mime_type.startsWith('image/')) {
    insertImage(media)
  }
}

const insertImage = (media) => {
  if (editor.value) {
    editor.value.chain().focus().setImage({ src: media.url, alt: media.filename }).run()
    showMediaModal.value = false
  }
}

const handleFileUpload = (media) => {
  insertFile(media)
}

const insertFile = (media) => {
  if (editor.value) {
    const fileLink = `📎 ${media.filename}`
    editor.value.chain().focus().insertContent(`<p><a href="${media.url}" target="_blank" class="file-attachment">${fileLink}</a></p>`).run()
    showFileModal.value = false
  }
}

onBeforeUnmount(() => {
  if (editor.value) {
    editor.value.destroy()
  }
})
</script>

<style scoped>
.rich-text-editor {
  @apply border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800;
  position: relative;
  max-height: 80vh;
  overflow-y: auto;
}

.editor-toolbar {
  @apply flex flex-wrap items-center gap-1 p-2 bg-gray-50 dark:bg-gray-900 border-b border-gray-300 dark:border-gray-600 rounded-t-lg;
  position: sticky;
  top: 0;
  z-index: 20;
}

.toolbar-group {
  @apply flex items-center gap-1 border-r border-gray-300 dark:border-gray-600 pr-2 mr-2;
  position: relative;
}

.toolbar-group:last-child {
  @apply border-r-0 mr-0 pr-0;
}

.toolbar-button {
  @apply p-2 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-gray-700 dark:text-gray-300;
}

.toolbar-button.is-active {
  @apply bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-400;
}

/* Color picker */
.color-picker-wrapper {
  position: relative;
}

.color-button {
  @apply flex flex-col items-center gap-0.5 p-1.5;
}

.color-swatch {
  display: block;
  width: 16px;
  height: 3px;
  border-radius: 2px;
  border: 1px solid rgba(0,0,0,0.15);
}

.color-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  z-index: 50;
  @apply bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-xl p-3;
  min-width: 160px;
}

.color-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 6px;
  margin-bottom: 8px;
}

.color-dot {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  transition: transform 0.1s, border-color 0.1s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}

.color-dot:hover {
  transform: scale(1.2);
  border-color: rgba(0,0,0,0.3);
}

.color-remove-btn {
  @apply flex items-center gap-1 text-xs text-gray-500 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 w-full pt-1 border-t border-gray-200 dark:border-gray-700;
}

/* Table menu */
.table-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  z-index: 50;
  @apply bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-xl py-1;
  min-width: 200px;
}

.table-menu-item {
  @apply flex items-center gap-2 w-full px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-left;
}

.table-menu-danger {
  @apply text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20;
}

.table-menu-divider {
  @apply border-gray-200 dark:border-gray-700 my-1;
}

.editor-content {
  @apply min-h-[300px];
}

/* Tiptap editor styles */
.editor-content :deep(.ProseMirror) {
  @apply focus:outline-none;
}

.editor-content :deep(.ProseMirror p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  @apply text-gray-400 dark:text-gray-500 float-left h-0 pointer-events-none;
}

/* Heading styles */
.editor-content :deep(.ProseMirror h1) {
  @apply text-3xl font-bold mb-4 mt-6 text-gray-900 dark:text-white;
}

.editor-content :deep(.ProseMirror h2) {
  @apply text-2xl font-bold mb-3 mt-5 text-gray-900 dark:text-white;
}

.editor-content :deep(.ProseMirror h3) {
  @apply text-xl font-bold mb-2 mt-4 text-gray-900 dark:text-white;
}

.editor-content :deep(.ProseMirror h4) {
  @apply text-lg font-bold mb-2 mt-3 text-gray-900 dark:text-white;
}

/* Paragraph styling */
.editor-content :deep(.ProseMirror p) {
  @apply mb-4 leading-relaxed;
}

/* List styling */
.editor-content :deep(.ProseMirror ul) {
  @apply list-disc ml-6 mb-4 space-y-1;
}

.editor-content :deep(.ProseMirror ol) {
  @apply list-decimal ml-6 mb-4 space-y-1;
}

.editor-content :deep(.ProseMirror li) {
  @apply ml-2;
}

.editor-content :deep(.ProseMirror ul ul),
.editor-content :deep(.ProseMirror ol ol) {
  @apply ml-6 mb-0 mt-1;
}

/* Code block styling */
.editor-content :deep(.ProseMirror pre) {
  @apply bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto;
}

.editor-content :deep(.ProseMirror code) {
  @apply bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded text-sm;
}

/* Link styling */
.editor-content :deep(.ProseMirror a) {
  @apply text-blue-600 dark:text-blue-400 underline hover:text-blue-800 dark:hover:text-blue-300;
}

/* Table styling */
.editor-content :deep(.ProseMirror .tableWrapper) {
  overflow-x: auto;
  margin: 1rem 0;
}

.editor-content :deep(.ProseMirror table) {
  @apply border-collapse table-auto w-full;
  min-width: 100%;
}

.editor-content :deep(.ProseMirror th),
.editor-content :deep(.ProseMirror td) {
  @apply border border-gray-300 dark:border-gray-600 px-3 py-2;
  position: relative; /* required for the resize handle */
  min-width: 40px;
}

.editor-content :deep(.ProseMirror th) {
  @apply bg-gray-100 dark:bg-gray-800 font-semibold;
}

/* ── Column resize handle ── */
.editor-content :deep(.ProseMirror .column-resize-handle) {
  position: absolute;
  right: -3px;
  top: 0;
  bottom: 0;
  width: 6px;
  cursor: col-resize;
  z-index: 20;
  pointer-events: all;
}

.editor-content :deep(.ProseMirror .column-resize-handle::after) {
  content: '';
  position: absolute;
  left: 2px;
  top: 10%;
  bottom: 10%;
  width: 2px;
  border-radius: 2px;
  background: #6366F1;
  opacity: 0;
  transition: opacity 0.15s;
}

/* Appear on hover */
.editor-content :deep(.ProseMirror th:hover > .column-resize-handle::after),
.editor-content :deep(.ProseMirror td:hover > .column-resize-handle::after) {
  opacity: 0.45;
}

/* Bright while dragging */
.editor-content :deep(.ProseMirror.resize-cursor .column-resize-handle::after) {
  opacity: 1 !important;
  background: #4F46E5;
}

.editor-content :deep(.ProseMirror.resize-cursor) {
  cursor: col-resize;
}

/* ── Selected cell highlight ── */
.editor-content :deep(.ProseMirror .selectedCell::after) {
  z-index: 2;
  position: absolute;
  content: '';
  left: 0; right: 0; top: 0; bottom: 0;
  background: rgba(99, 102, 241, 0.12);
  pointer-events: none;
}

/* Image styling */
.editor-content :deep(.ProseMirror img) {
  @apply rounded-lg max-w-full h-auto my-4 cursor-pointer;
}

/* File attachment styling */
.editor-content :deep(.ProseMirror a.file-attachment) {
  @apply inline-flex items-center gap-2 px-3 py-2 bg-gray-100 dark:bg-gray-800 rounded-lg no-underline hover:bg-gray-200 dark:hover:bg-gray-700;
}

/* Highlight mark */
.editor-content :deep(.ProseMirror mark) {
  border-radius: 3px;
  padding: 0 2px;
}

/* Modal styles */
.modal-overlay {
  @apply fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4;
}

.modal-content {
  @apply bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col;
}

.modal-header {
  @apply flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700;
}

.modal-title {
  @apply text-lg font-semibold text-gray-900 dark:text-white;
}

.modal-close {
  @apply p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors text-gray-500 dark:text-gray-400;
}

.modal-body {
  @apply p-4 overflow-y-auto flex-1;
}

.tabs {
  @apply flex gap-2 mb-4 border-b border-gray-200 dark:border-gray-700;
}

.tab-button {
  @apply flex items-center gap-2 px-4 py-2 border-b-2 border-transparent text-gray-600 dark:text-gray-400;
  @apply hover:text-gray-900 dark:hover:text-white transition-colors;
}

.tab-button.active {
  @apply border-blue-500 text-blue-600 dark:text-blue-400 font-medium;
}

.tab-content {
  @apply py-4;
}

/* Callout dropdown */
.callout-type-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  z-index: 50;
  @apply bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-xl py-1;
  min-width: 140px;
}

.callout-type-item {
  @apply flex items-center gap-2 w-full px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-left;
}
</style>
