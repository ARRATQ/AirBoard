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
          :aria-pressed="editor.isActive('bold')"
          :title="t('news.editor.bold')"
          :aria-label="t('news.editor.bold')"
        >
          <Icon icon="lucide:bold" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleItalic().run()"
          :class="{ 'is-active': editor.isActive('italic') }"
          class="toolbar-button"
          :aria-pressed="editor.isActive('italic')"
          :title="t('news.editor.italic')"
          :aria-label="t('news.editor.italic')"
        >
          <Icon icon="lucide:italic" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleUnderline().run()"
          :class="{ 'is-active': editor.isActive('underline') }"
          class="toolbar-button"
          :aria-pressed="editor.isActive('underline')"
          :title="t('news.editor.underline')"
          :aria-label="t('news.editor.underline')"
        >
          <Icon icon="lucide:underline" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleStrike().run()"
          :class="{ 'is-active': editor.isActive('strike') }"
          class="toolbar-button"
          :aria-pressed="editor.isActive('strike')"
          :title="t('news.editor.strike')"
          :aria-label="t('news.editor.strike')"
        >
          <Icon icon="lucide:strikethrough" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleCode().run()"
          :class="{ 'is-active': editor.isActive('code') }"
          class="toolbar-button"
          :aria-pressed="editor.isActive('code')"
          :title="t('news.editor.inlineCode')"
          :aria-label="t('news.editor.inlineCode')"
        >
          <Icon icon="lucide:code" class="h-5 w-5" aria-hidden="true" />
        </button>
      </div>

      <!-- Color & Highlight -->
      <div class="toolbar-group">
        <!-- Text color -->
        <div class="color-picker-wrapper">
          <button
            type="button"
            class="toolbar-button color-button"
            :title="t('news.editor.textColor')"
            :aria-label="t('news.editor.textColor')"
            aria-haspopup="true"
            :aria-expanded="showColorPicker"
            @click="toggleColorPicker"
          >
            <Icon icon="lucide:baseline" class="h-5 w-5" aria-hidden="true" />
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
                :title="t(color.nameKey)"
                :aria-label="t(color.nameKey)"
                @click="setTextColor(color.value)"
              ></button>
            </div>
            <button type="button" class="color-remove-btn" @click="removeTextColor">
              <Icon icon="lucide:remove-formatting" class="h-4 w-4" aria-hidden="true" /> {{ t('common.clear') }}
            </button>
          </div>
        </div>

        <!-- Highlight color -->
        <div class="color-picker-wrapper">
          <button
            type="button"
            class="toolbar-button color-button"
            :title="t('news.editor.highlight')"
            :aria-label="t('news.editor.highlight')"
            aria-haspopup="true"
            :aria-expanded="showHighlightPicker"
            @click="toggleHighlightPicker"
          >
            <Icon icon="lucide:highlighter" class="h-5 w-5" aria-hidden="true" />
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
                :title="t(color.nameKey)"
                :aria-label="t(color.nameKey)"
                @click="setHighlight(color.value)"
              ></button>
            </div>
            <button type="button" class="color-remove-btn" @click="removeHighlight">
              <Icon icon="lucide:remove-formatting" class="h-4 w-4" aria-hidden="true" /> {{ t('common.clear') }}
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
          :aria-pressed="editor.isActive('heading', { level: 1 })"
          :title="t('news.editor.heading1')"
          :aria-label="t('news.editor.heading1')"
        >
          <Icon icon="lucide:heading-1" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
          :class="{ 'is-active': editor.isActive('heading', { level: 2 }) }"
          class="toolbar-button"
          :aria-pressed="editor.isActive('heading', { level: 2 })"
          :title="t('news.editor.heading2')"
          :aria-label="t('news.editor.heading2')"
        >
          <Icon icon="lucide:heading-2" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleHeading({ level: 3 }).run()"
          :class="{ 'is-active': editor.isActive('heading', { level: 3 }) }"
          class="toolbar-button"
          :aria-pressed="editor.isActive('heading', { level: 3 })"
          :title="t('news.editor.heading3')"
          :aria-label="t('news.editor.heading3')"
        >
          <Icon icon="lucide:heading-3" class="h-5 w-5" aria-hidden="true" />
        </button>
      </div>

      <!-- Text Alignment -->
      <div class="toolbar-group">
        <button
          type="button"
          @click="editor.chain().focus().setTextAlign('left').run()"
          :class="{ 'is-active': editor.isActive({ textAlign: 'left' }) }"
          class="toolbar-button"
          :aria-pressed="editor.isActive({ textAlign: 'left' })"
          :title="t('news.editor.alignLeft')"
          :aria-label="t('news.editor.alignLeft')"
        >
          <Icon icon="lucide:align-left" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().setTextAlign('center').run()"
          :class="{ 'is-active': editor.isActive({ textAlign: 'center' }) }"
          class="toolbar-button"
          :aria-pressed="editor.isActive({ textAlign: 'center' })"
          :title="t('news.editor.alignCenter')"
          :aria-label="t('news.editor.alignCenter')"
        >
          <Icon icon="lucide:align-center" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().setTextAlign('right').run()"
          :class="{ 'is-active': editor.isActive({ textAlign: 'right' }) }"
          class="toolbar-button"
          :aria-pressed="editor.isActive({ textAlign: 'right' })"
          :title="t('news.editor.alignRight')"
          :aria-label="t('news.editor.alignRight')"
        >
          <Icon icon="lucide:align-right" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().setTextAlign('justify').run()"
          :class="{ 'is-active': editor.isActive({ textAlign: 'justify' }) }"
          class="toolbar-button"
          :aria-pressed="editor.isActive({ textAlign: 'justify' })"
          :title="t('news.editor.alignJustify')"
          :aria-label="t('news.editor.alignJustify')"
        >
          <Icon icon="lucide:align-justify" class="h-5 w-5" aria-hidden="true" />
        </button>
      </div>

      <!-- Lists -->
      <div class="toolbar-group">
        <button
          type="button"
          @click="editor.chain().focus().toggleBulletList().run()"
          :class="{ 'is-active': editor.isActive('bulletList') }"
          class="toolbar-button"
          :aria-pressed="editor.isActive('bulletList')"
          :title="t('news.editor.bulletList')"
          :aria-label="t('news.editor.bulletList')"
        >
          <Icon icon="lucide:list" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleOrderedList().run()"
          :class="{ 'is-active': editor.isActive('orderedList') }"
          class="toolbar-button"
          :aria-pressed="editor.isActive('orderedList')"
          :title="t('news.editor.orderedList')"
          :aria-label="t('news.editor.orderedList')"
        >
          <Icon icon="lucide:list-ordered" class="h-5 w-5" aria-hidden="true" />
        </button>
      </div>

      <!-- Blocks -->
      <div class="toolbar-group">
        <button
          type="button"
          @click="editor.chain().focus().toggleCodeBlock().run()"
          :class="{ 'is-active': editor.isActive('codeBlock') }"
          class="toolbar-button"
          :aria-pressed="editor.isActive('codeBlock')"
          :title="t('news.editor.codeBlock')"
          :aria-label="t('news.editor.codeBlock')"
        >
          <Icon icon="lucide:braces" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().toggleBlockquote().run()"
          :class="{ 'is-active': editor.isActive('blockquote') }"
          class="toolbar-button"
          :aria-pressed="editor.isActive('blockquote')"
          :title="t('news.editor.blockquote')"
          :aria-label="t('news.editor.blockquote')"
        >
          <Icon icon="lucide:quote" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().setHorizontalRule().run()"
          class="toolbar-button"
          :title="t('news.editor.horizontalRule')"
          :aria-label="t('news.editor.horizontalRule')"
        >
          <Icon icon="lucide:minus" class="h-5 w-5" aria-hidden="true" />
        </button>

        <!-- Callout -->
        <div class="color-picker-wrapper">
          <button
            type="button"
            :title="t('news.editor.callout.label')"
            :aria-label="t('news.editor.callout.label')"
            aria-haspopup="true"
            :aria-expanded="showCalloutMenu"
            @click="toggleCalloutMenu"
            :class="{ 'is-active': editor.isActive('callout') }"
            class="toolbar-button"
          >
            <Icon icon="lucide:message-square-warning" class="h-5 w-5" aria-hidden="true" />
          </button>
          <div v-if="showCalloutMenu" class="callout-type-dropdown" @mouseleave="closeCalloutMenu">
            <button
              v-for="type in calloutTypes"
              :key="type.value"
              type="button"
              class="callout-type-item"
              @click="insertCallout(type.value)"
            >
              <Icon :icon="type.icon" class="h-4 w-4" aria-hidden="true" :style="{ color: type.color }" />
              <span>{{ t(type.labelKey) }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Table -->
      <div class="toolbar-group">
        <div class="color-picker-wrapper">
          <button
            type="button"
            :title="t('news.editor.table')"
            :aria-label="t('news.editor.table')"
            aria-haspopup="true"
            :aria-expanded="showTableMenu"
            @click="toggleTableMenu"
            :class="{ 'is-active': editor.isActive('table') }"
            class="toolbar-button"
          >
            <Icon icon="lucide:table" class="h-5 w-5" aria-hidden="true" />
          </button>
          <div v-if="showTableMenu" class="table-dropdown" @mouseleave="closeTableMenu">
            <button type="button" class="table-menu-item" @click="insertTable">
              <Icon icon="lucide:table" class="h-4 w-4" aria-hidden="true" /> {{ t('news.editor.insertTable') }}
            </button>
            <template v-if="editor.isActive('table')">
              <hr class="table-menu-divider" />
              <button type="button" class="table-menu-item" @click="editor.chain().focus().addColumnBefore().run()">
                <Icon icon="lucide:between-vertical-start" class="h-4 w-4" aria-hidden="true" /> {{ t('news.editor.addColumnBefore') }}
              </button>
              <button type="button" class="table-menu-item" @click="editor.chain().focus().addColumnAfter().run()">
                <Icon icon="lucide:between-vertical-end" class="h-4 w-4" aria-hidden="true" /> {{ t('news.editor.addColumnAfter') }}
              </button>
              <button type="button" class="table-menu-item" @click="editor.chain().focus().deleteColumn().run()">
                <Icon icon="lucide:between-vertical-start" class="h-4 w-4" aria-hidden="true" /> {{ t('news.editor.deleteColumn') }}
              </button>
              <hr class="table-menu-divider" />
              <button type="button" class="table-menu-item" @click="editor.chain().focus().addRowBefore().run()">
                <Icon icon="lucide:between-horizontal-start" class="h-4 w-4" aria-hidden="true" /> {{ t('news.editor.addRowBefore') }}
              </button>
              <button type="button" class="table-menu-item" @click="editor.chain().focus().addRowAfter().run()">
                <Icon icon="lucide:between-horizontal-end" class="h-4 w-4" aria-hidden="true" /> {{ t('news.editor.addRowAfter') }}
              </button>
              <button type="button" class="table-menu-item" @click="editor.chain().focus().deleteRow().run()">
                <Icon icon="lucide:between-horizontal-start" class="h-4 w-4" aria-hidden="true" /> {{ t('news.editor.deleteRow') }}
              </button>
              <hr class="table-menu-divider" />
              <button type="button" class="table-menu-item table-menu-danger" @click="editor.chain().focus().deleteTable().run()">
                <Icon icon="lucide:table-2" class="h-4 w-4" aria-hidden="true" /> {{ t('news.editor.deleteTable') }}
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
          :aria-pressed="editor.isActive('link')"
          :title="t('news.editor.addLink')"
          :aria-label="t('news.editor.addLink')"
        >
          <Icon icon="lucide:link" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          v-if="editor.isActive('link')"
          type="button"
          @click="editor.chain().focus().unsetLink().run()"
          class="toolbar-button"
          :title="t('news.editor.removeLink')"
          :aria-label="t('news.editor.removeLink')"
        >
          <Icon icon="lucide:unlink" class="h-5 w-5" aria-hidden="true" />
        </button>
      </div>

      <!-- Media -->
      <div class="toolbar-group">
        <button
          type="button"
          @click="showMediaModal = true"
          class="toolbar-button"
          :title="t('news.editor.insertImage')"
          :aria-label="t('news.editor.insertImage')"
        >
          <Icon icon="lucide:image-plus" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="showFileModal = true"
          class="toolbar-button"
          :title="t('news.editor.attachFile')"
          :aria-label="t('news.editor.attachFile')"
        >
          <Icon icon="lucide:file-plus" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="insertVideo"
          class="toolbar-button"
          :title="t('news.editor.embedVideo')"
          :aria-label="t('news.editor.embedVideo')"
        >
          <Icon icon="lucide:video" class="h-5 w-5" aria-hidden="true" />
        </button>
      </div>

      <!-- Undo/Redo -->
      <div class="toolbar-group ms-auto">
        <button
          type="button"
          @click="editor.chain().focus().undo().run()"
          :disabled="!editor.can().undo()"
          class="toolbar-button"
          :title="t('news.editor.undo')"
          :aria-label="t('news.editor.undo')"
        >
          <Icon icon="lucide:undo-2" class="h-5 w-5" aria-hidden="true" />
        </button>
        <button
          type="button"
          @click="editor.chain().focus().redo().run()"
          :disabled="!editor.can().redo()"
          class="toolbar-button"
          :title="t('news.editor.redo')"
          :aria-label="t('news.editor.redo')"
        >
          <Icon icon="lucide:redo-2" class="h-5 w-5" aria-hidden="true" />
        </button>
      </div>
    </div>

    <!-- Editor content -->
    <EditorContent :editor="editor" class="editor-content" />

    <!-- Media Modal -->
    <teleport to="body">
      <div v-if="showMediaModal" class="modal-overlay" @click.self="showMediaModal = false">
        <div
          v-focus-trap
          class="modal-content"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="`${uid}-media-title`"
        >
          <div class="modal-header">
            <h3 :id="`${uid}-media-title`" class="modal-title">{{ t('news.editor.insertImage') }}</h3>
            <button type="button" :aria-label="t('common.close')" @click="showMediaModal = false" class="modal-close">
              <Icon icon="lucide:x" class="h-6 w-6" aria-hidden="true" />
            </button>
          </div>
          <div class="modal-body">
            <div class="tabs">
              <button
                type="button"
                :aria-pressed="mediaTab === 'upload'"
                @click="mediaTab = 'upload'"
                :class="{ 'active': mediaTab === 'upload' }"
                class="tab-button"
              >
                <Icon icon="lucide:upload" class="h-5 w-5" aria-hidden="true" />
                {{ t('media.upload') }}
              </button>
              <button
                type="button"
                :aria-pressed="mediaTab === 'gallery'"
                @click="mediaTab = 'gallery'"
                :class="{ 'active': mediaTab === 'gallery' }"
                class="tab-button"
              >
                <Icon icon="lucide:images" class="h-5 w-5" aria-hidden="true" />
                {{ t('directory.pickFromLibrary') }}
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
        <div
          v-focus-trap
          class="modal-content"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="`${uid}-file-title`"
        >
          <div class="modal-header">
            <h3 :id="`${uid}-file-title`" class="modal-title">{{ t('news.editor.attachFile') }}</h3>
            <button type="button" :aria-label="t('common.close')" @click="showFileModal = false" class="modal-close">
              <Icon icon="lucide:x" class="h-6 w-6" aria-hidden="true" />
            </button>
          </div>
          <div class="modal-body">
            <div class="tabs">
              <button
                type="button"
                :aria-pressed="fileTab === 'upload'"
                @click="fileTab = 'upload'"
                :class="{ 'active': fileTab === 'upload' }"
                class="tab-button"
              >
                <Icon icon="lucide:upload" class="h-5 w-5" aria-hidden="true" />
                {{ t('media.upload') }}
              </button>
              <button
                type="button"
                :aria-pressed="fileTab === 'gallery'"
                @click="fileTab = 'gallery'"
                :class="{ 'active': fileTab === 'gallery' }"
                class="tab-button"
              >
                <Icon icon="lucide:file" class="h-5 w-5" aria-hidden="true" />
                {{ t('directory.pickFromLibrary') }}
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
import { ref, computed, watch, onBeforeUnmount, useId } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
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
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
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

const { t } = useI18n()
const appStore = useAppStore()
// Identifiant unique : titres des modales (plusieurs éditeurs possibles sur une page)
const uid = useId()

const props = defineProps({
  modelValue: {
    type: [String, Object],
    default: ''
  },
  placeholder: {
    type: String,
    default: ''
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

// Palette de couleurs pour le texte (noms traduits au rendu)
const textColors = [
  { nameKey: 'news.editor.colors.black', value: '#000000' },
  { nameKey: 'news.editor.colors.darkGray', value: '#374151' },
  { nameKey: 'news.editor.colors.gray', value: '#6B7280' },
  { nameKey: 'news.editor.colors.red', value: '#EF4444' },
  { nameKey: 'news.editor.colors.orange', value: '#F97316' },
  { nameKey: 'news.editor.colors.yellow', value: '#EAB308' },
  { nameKey: 'news.editor.colors.green', value: '#22C55E' },
  { nameKey: 'news.editor.colors.cyan', value: '#06B6D4' },
  { nameKey: 'news.editor.colors.blue', value: '#3B82F6' },
  { nameKey: 'news.editor.colors.indigo', value: '#6366F1' },
  { nameKey: 'news.editor.colors.purple', value: '#A855F7' },
  { nameKey: 'news.editor.colors.pink', value: '#EC4899' },
]

// Palette de couleurs pour la surbrillance
const highlightColors = [
  { nameKey: 'news.editor.colors.yellow', value: '#FEF08A' },
  { nameKey: 'news.editor.colors.lightGreen', value: '#BBF7D0' },
  { nameKey: 'news.editor.colors.lightBlue', value: '#BAE6FD' },
  { nameKey: 'news.editor.colors.lightPink', value: '#FBCFE8' },
  { nameKey: 'news.editor.colors.lightOrange', value: '#FED7AA' },
  { nameKey: 'news.editor.colors.lightPurple', value: '#DDD6FE' },
  { nameKey: 'news.editor.colors.lightRed', value: '#FECACA' },
  { nameKey: 'news.editor.colors.lightCyan', value: '#A5F3FC' },
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
  const url = window.prompt(t('news.editor.videoPrompt'))
  if (!url) return
  const embedUrl = getEmbedUrl(url)
  if (!embedUrl) {
    appStore.showWarning(t('news.editor.invalidVideoUrl'))
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
    // Tiptap 3 : StarterKit inclut Link et Underline, configurés séparément ci-dessous
    StarterKit.configure({
      codeBlock: false,
      link: false,
      underline: false,
    }),
    Link.configure({
      openOnClick: false,
      HTMLAttributes: {
        class: 'text-primary-600 dark:text-primary-400 underline hover:text-primary-800 dark:hover:text-primary-300',
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
      placeholder: props.placeholder || t('news.editor.placeholder'),
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
    // Contenu structuré plutôt que du HTML concaténé : le nom du fichier
    // reste du texte et ne peut pas injecter de balises.
    editor.value.chain().focus().insertContent({
      type: 'paragraph',
      content: [{
        type: 'text',
        text: `📎 ${media.filename}`,
        marks: [{ type: 'link', attrs: { href: media.url, target: '_blank', class: 'file-attachment' } }]
      }]
    }).run()
    showFileModal.value = false
  }
}

onBeforeUnmount(() => {
  if (editor.value) {
    editor.value.destroy()
  }
})

// Échap ferme la modale
useEscapeClose(showMediaModal, () => { showMediaModal.value = false })
useEscapeClose(showFileModal, () => { showFileModal.value = false })
</script>

<style scoped>
.rich-text-editor {
  @apply rounded-lg;
  background-color: var(--bg-card);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  max-height: 80vh;
  /* overflow must NOT be set here — it would clip the toolbar dropdowns */
}

.editor-toolbar {
  @apply flex flex-wrap items-center gap-1 p-2 rounded-t-lg;
  background-color: var(--bg-surface);
  border-bottom: 1px solid var(--border);
  /* flex-shrink: 0 keeps the toolbar always visible; no sticky needed */
  flex-shrink: 0;
  z-index: 20;
}

.toolbar-group {
  @apply flex items-center gap-1 pe-2 me-2;
  border-inline-end: 1px solid var(--border);
  position: relative;
}

.toolbar-group:last-child {
  @apply border-e-0 me-0 pe-0;
}

.toolbar-button {
  @apply p-2 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-gray-700 dark:text-gray-300;
}

.toolbar-button.is-active {
  @apply bg-primary-100 dark:bg-primary-900 text-primary-600 dark:text-primary-400;
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
  inset-inline-start: 0;
  z-index: 50;
  @apply rounded-lg shadow-xl p-3;
  background-color: var(--bg-card);
  border: 1px solid var(--border);
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
  inset-inline-start: 0;
  z-index: 50;
  @apply rounded-lg shadow-xl py-1;
  background-color: var(--bg-card);
  border: 1px solid var(--border);
  min-width: 200px;
}

.table-menu-item {
  @apply flex items-center gap-2 w-full px-3 py-2 text-sm transition-colors text-start;
  color: var(--text-primary);
}
.table-menu-item:hover {
  background-color: var(--hover);
}

.table-menu-danger {
  @apply text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20;
}

.table-menu-divider {
  @apply border-gray-200 dark:border-gray-700 my-1;
}

.editor-content {
  @apply min-h-[300px];
  /* This is now the scrollable zone — toolbar stays above and is never clipped */
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

/* Tiptap editor styles */
.editor-content :deep(.ProseMirror) {
  @apply focus:outline-none;
}

.editor-content :deep(.ProseMirror p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  @apply text-gray-400 dark:text-gray-500 float-start h-0 pointer-events-none;
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
  @apply list-disc ms-6 mb-4 space-y-1;
}

.editor-content :deep(.ProseMirror ol) {
  @apply list-decimal ms-6 mb-4 space-y-1;
}

.editor-content :deep(.ProseMirror li) {
  @apply ms-2;
}

.editor-content :deep(.ProseMirror ul ul),
.editor-content :deep(.ProseMirror ol ol) {
  @apply ms-6 mb-0 mt-1;
}

/* Code block styling */
.editor-content :deep(.ProseMirror pre) {
  @apply bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto;
}

.editor-content :deep(.ProseMirror code) {
  @apply px-1 py-0.5 rounded text-sm;
  background-color: var(--bg-surface);
}

/* Link styling */
.editor-content :deep(.ProseMirror a) {
  @apply text-primary-600 dark:text-primary-400 underline hover:text-primary-800 dark:hover:text-primary-300;
}

/* Table styling */
.editor-content :deep(.ProseMirror .tableWrapper) {
  overflow-x: auto;
  margin: 1rem 0;
  /* Allow resize handles to overflow the wrapper edges */
  padding-inline-end: 4px;
}

.editor-content :deep(.ProseMirror table) {
  @apply border-collapse w-full;
  /* table-fixed is required so Tiptap's colgroup widths are respected */
  table-layout: fixed;
}

.editor-content :deep(.ProseMirror th),
.editor-content :deep(.ProseMirror td) {
  @apply border border-gray-300 dark:border-gray-600 px-3 py-2;
  position: relative; /* required for the resize handle */
  min-width: 40px;
  /* Allow text to wrap rather than forcing column expansion */
  word-break: break-word;
  overflow-wrap: break-word;
}

.editor-content :deep(.ProseMirror th) {
  @apply font-semibold;
  background-color: var(--bg-surface);
}

/* ── Column resize handle ── */
.editor-content :deep(.ProseMirror .column-resize-handle) {
  position: absolute;
  inset-inline-end: -3px;
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
  inset-inline-start: 2px;
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
  inset-inline-start: 0; inset-inline-end: 0; top: 0; bottom: 0;
  background: rgba(99, 102, 241, 0.12);
  pointer-events: none;
}

/* Image styling */
.editor-content :deep(.ProseMirror img) {
  @apply rounded-lg max-w-full h-auto my-4 cursor-pointer;
}

/* File attachment styling */
.editor-content :deep(.ProseMirror a.file-attachment) {
  @apply inline-flex items-center gap-2 px-3 py-2 rounded-lg no-underline;
  background-color: var(--bg-surface);
}
.editor-content :deep(.ProseMirror a.file-attachment:hover) {
  background-color: var(--hover);
}

/* Highlight mark */
.editor-content :deep(.ProseMirror mark) {
  border-radius: var(--r-s);
  padding: 0 2px;
}

/* Modal styles */
.modal-overlay {
  @apply fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4;
}

.modal-content {
  @apply rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col;
  background-color: var(--bg-card);
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
  @apply border-primary-500 text-primary-600 dark:text-primary-400 font-medium;
}

.tab-content {
  @apply py-4;
}

/* Callout dropdown */
.callout-type-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  inset-inline-start: 0;
  z-index: 50;
  @apply rounded-lg shadow-xl py-1;
  background-color: var(--bg-card);
  border: 1px solid var(--border);
  min-width: 140px;
}

.callout-type-item {
  @apply flex items-center gap-2 w-full px-3 py-2 text-sm transition-colors text-start;
  color: var(--text-primary);
}
.callout-type-item:hover {
  background-color: var(--hover);
}
</style>
