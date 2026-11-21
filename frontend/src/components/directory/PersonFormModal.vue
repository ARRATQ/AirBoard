<template>
  <DirModal
    :title="person ? $t('directory.editPerson') : $t('directory.newPerson')"
    :saving="saving"
    :error="error"
    size="lg"
    @close="$emit('close')"
    @submit="submit"
  >
    <div class="grid gap-x-8 md:grid-cols-2">
      <!-- Identité -->
      <section>
        <h4 class="person-form__section">{{ $t('directory.identity') }}</h4>
        <div class="flex items-center gap-4 mb-4">
          <PersonAvatar :person="previewPerson" size="lg" />
          <div class="flex flex-wrap gap-2">
            <button type="button" class="btn btn-secondary text-xs" :disabled="uploading" @click="fileInput?.click()">
              <Icon :icon="uploading ? 'mdi:loading' : 'mdi:upload'" class="h-4 w-4 mr-1" :class="{ 'animate-spin': uploading }" />
              {{ $t('directory.uploadPhoto') }}
            </button>
            <button type="button" class="btn btn-secondary text-xs" @click="showGallery = true">
              <Icon icon="mdi:image-multiple-outline" class="h-4 w-4 mr-1" /> {{ $t('directory.pickFromLibrary') }}
            </button>
            <button v-if="form.photo_url" type="button" class="btn btn-ghost text-xs" @click="form.photo_url = ''">
              <Icon icon="mdi:close" class="h-4 w-4 mr-1" /> {{ $t('directory.removePhoto') }}
            </button>
            <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onFile" />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="form-group">
            <label class="form-label form-label-required" for="p-first">{{ $t('directory.firstName') }}</label>
            <input id="p-first" v-model.trim="form.first_name" class="form-input" maxlength="100" required />
          </div>
          <div class="form-group">
            <label class="form-label form-label-required" for="p-last">{{ $t('directory.lastName') }}</label>
            <input id="p-last" v-model.trim="form.last_name" class="form-input" maxlength="100" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="p-email">{{ $t('directory.email') }}</label>
            <input id="p-email" v-model.trim="form.email" type="email" class="form-input" maxlength="200" />
          </div>
          <div class="form-group">
            <label class="form-label" for="p-phone">{{ $t('directory.phone') }}</label>
            <input id="p-phone" v-model.trim="form.phone" type="tel" class="form-input" maxlength="50" />
          </div>
        </div>

        <!-- Lien optionnel vers un compte Airboard -->
        <div class="form-group">
          <span class="form-label">{{ $t('directory.linkedAccount') }}</span>
          <div v-if="form.user_id" class="flex items-center justify-between gap-2 rounded-lg border px-3 py-2" style="border-color: var(--border)">
            <span class="flex items-center gap-2 text-sm" style="color: var(--text-primary)">
              <Icon icon="mdi:account-check-outline" class="h-4 w-4 text-green-500" />
              {{ linkedUser ? `${linkedUser.first_name} ${linkedUser.last_name} (${linkedUser.username})` : $t('directory.linkedAccountId', { id: form.user_id }) }}
            </span>
            <button type="button" class="btn btn-ghost text-xs" @click="unlink">{{ $t('directory.unlink') }}</button>
          </div>
          <div v-else class="relative">
            <input
              v-model="userQuery"
              type="search"
              class="form-input"
              :placeholder="$t('directory.searchAccount')"
              @input="searchUsers"
            />
            <ul v-if="userResults.length" class="person-form__results">
              <li v-for="u in userResults" :key="u.id">
                <button type="button" @click="linkUser(u)">
                  <span class="font-medium">{{ u.first_name }} {{ u.last_name }}</span>
                  <span class="text-xs" style="color: var(--text-muted)">{{ u.username }} · {{ u.email }}</span>
                </button>
              </li>
            </ul>
          </div>
          <p class="text-xs mt-1" style="color: var(--text-muted)">{{ $t('directory.linkedAccountHelp') }}</p>
        </div>

        <label class="flex items-center gap-3 cursor-pointer">
          <input v-model="form.is_active" type="checkbox" />
          <span class="text-sm" style="color: var(--text-primary)">{{ $t('directory.activeHelp') }}</span>
        </label>
      </section>

      <!-- Champs dynamiques -->
      <section class="mt-6 md:mt-0">
        <h4 class="person-form__section">{{ $t('directory.additionalInfo') }}</h4>
        <p v-if="!fields.length" class="text-sm italic" style="color: var(--text-muted)">{{ $t('directory.noCustomFields') }}</p>
        <div v-for="f in fields" :key="f.id" class="form-group">
          <label class="form-label flex items-center gap-1" :class="{ 'form-label-required': f.required }" :for="`cf-${f.key}`">
            {{ f.label }}
            <Icon v-if="f.visibility === 'managers'" icon="mdi:lock-outline" class="h-3.5 w-3.5 text-amber-500" :title="$t('directory.fieldManagersOnly')" />
          </label>
          <textarea
            v-if="f.type === 'textarea'"
            :id="`cf-${f.key}`"
            v-model="form.custom_fields[f.key]"
            class="form-textarea"
            rows="3"
            :required="f.required"
          />
          <select
            v-else-if="f.type === 'select'"
            :id="`cf-${f.key}`"
            v-model="form.custom_fields[f.key]"
            class="form-select"
            :required="f.required"
          >
            <option value="">—</option>
            <option v-for="o in f.options" :key="o" :value="o">{{ o }}</option>
          </select>
          <input
            v-else
            :id="`cf-${f.key}`"
            v-model.trim="form.custom_fields[f.key]"
            :type="INPUT_TYPES[f.type] || 'text'"
            class="form-input"
            :required="f.required"
            :placeholder="f.type === 'url' ? 'https://' : ''"
          />
        </div>
      </section>
    </div>

    <!-- Médiathèque -->
    <teleport to="body">
      <div v-if="showGallery" class="modal-overlay" style="z-index: 60" @click.self="showGallery = false">
        <div class="modal-container">
          <div class="modal-panel w-full max-w-4xl p-4">
            <MediaGallery @select="pickMedia" @cancel="showGallery = false" />
          </div>
        </div>
      </div>
    </teleport>
  </DirModal>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import DirModal from './DirModal.vue'
import PersonAvatar from './PersonAvatar.vue'
import MediaGallery from '@/components/media/MediaGallery.vue'
import { directoryService, mediaService } from '@/services/api'
import { apiError } from './useDirectory'

const INPUT_TYPES = { email: 'email', phone: 'tel', url: 'url', date: 'date', text: 'text' }
const MAX_PHOTO_MB = 5

const props = defineProps({
  person: { type: Object, default: null },
  fields: { type: Array, required: true },
})
const emit = defineEmits(['close', 'saved'])
const { t } = useI18n()

const form = reactive({
  first_name: props.person?.first_name || '',
  last_name: props.person?.last_name || '',
  email: props.person?.email || '',
  phone: props.person?.phone || '',
  photo_url: props.person?.photo_url || '',
  user_id: props.person?.user_id ?? null,
  is_active: props.person?.is_active ?? true,
  custom_fields: Object.fromEntries(props.fields.map((f) => [f.key, props.person?.custom_fields?.[f.key] || ''])),
})
const saving = ref(false)
const error = ref('')
const uploading = ref(false)
const showGallery = ref(false)
const fileInput = ref(null)
const userQuery = ref('')
const userResults = ref([])
const linkedUser = ref(null)

const previewPerson = computed(() => ({ ...form, id: props.person?.id }))

const onFile = async (event) => {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file) return
  if (!file.type.startsWith('image/')) {
    error.value = t('directory.errors.photoType')
    return
  }
  if (file.size > MAX_PHOTO_MB * 1024 * 1024) {
    error.value = t('directory.errors.photoSize', { size: MAX_PHOTO_MB })
    return
  }
  uploading.value = true
  error.value = ''
  try {
    const res = await mediaService.uploadMediaAsGroupAdmin(file)
    form.photo_url = res.media?.url || ''
  } catch (e) {
    error.value = apiError(e, t('directory.errors.uploadFailed'))
  } finally {
    uploading.value = false
  }
}

const pickMedia = (media) => {
  if (media?.mime_type?.startsWith('image/')) {
    form.photo_url = media.url
    showGallery.value = false
  } else {
    error.value = t('directory.errors.photoType')
  }
}

let searchTimer = null
const searchUsers = () => {
  clearTimeout(searchTimer)
  const q = userQuery.value.trim()
  if (q.length < 2) {
    userResults.value = []
    return
  }
  searchTimer = setTimeout(async () => {
    try {
      userResults.value = await directoryService.searchUsers(q)
    } catch (e) {
      userResults.value = []
      error.value = apiError(e, t('directory.errors.searchFailed'))
    }
  }, 250)
}

// Lier un compte : pré-remplit uniquement les champs encore vides
const linkUser = (u) => {
  linkedUser.value = u
  form.user_id = u.id
  if (!form.first_name) form.first_name = u.first_name || ''
  if (!form.last_name) form.last_name = u.last_name || ''
  if (!form.email) form.email = u.email || ''
  if (!form.phone) form.phone = u.phone || ''
  if (!form.photo_url && u.avatar_url) form.photo_url = u.avatar_url
  userQuery.value = ''
  userResults.value = []
}

const unlink = () => {
  form.user_id = null
  linkedUser.value = null
}

const submit = async () => {
  if (!form.first_name || !form.last_name) {
    error.value = t('directory.errors.nameRequired')
    return
  }
  const missing = props.fields.find((f) => f.required && !String(form.custom_fields[f.key] || '').trim())
  if (missing) {
    error.value = t('directory.errors.fieldRequired', { field: missing.label })
    return
  }
  saving.value = true
  error.value = ''
  try {
    const payload = { ...form, custom_fields: { ...form.custom_fields } }
    const saved = props.person
      ? await directoryService.updatePerson(props.person.id, payload)
      : await directoryService.createPerson(payload)
    emit('saved', saved)
  } catch (e) {
    error.value = apiError(e, t('directory.errors.saveFailed'))
  } finally {
    saving.value = false
  }
}

// Échap ferme la modale
useEscapeClose(showGallery, () => { showGallery.value = false })
</script>

<style scoped>
.person-form__section {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 0.85rem;
}
.person-form__results {
  position: absolute;
  z-index: 20;
  inset-inline: 0;
  top: calc(100% + 0.25rem);
  max-height: 14rem;
  overflow-y: auto;
  border: 1px solid var(--border);
  border-radius: 0.6rem;
  background: var(--bg-card);
  box-shadow: 0 12px 24px -12px rgba(var(--shadow-rgb, 15, 23, 42), 0.35);
}
.person-form__results button {
  display: flex;
  flex-direction: column;
  width: 100%;
  padding: 0.5rem 0.75rem;
  text-align: start;
  color: var(--text-primary);
}
.person-form__results button:hover { background: var(--hover); }
</style>
