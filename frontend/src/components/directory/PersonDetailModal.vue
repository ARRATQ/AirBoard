<template>
  <teleport to="body">
    <div class="modal-overlay" @click.self="$emit('close')" @keydown.esc="$emit('close')">
      <div class="modal-container">
        <div class="modal-panel modal-form w-full max-w-3xl" role="dialog" aria-modal="true" :aria-label="name">
          <div class="person-detail__hero" :style="{ '--unit-color': roles[0]?.unit.color || 'var(--accent)' }">
            <button type="button" class="person-detail__close" :aria-label="$t('common.close')" @click="$emit('close')">
              <Icon icon="mdi:close" class="h-5 w-5" />
            </button>
            <PersonAvatar :person="person" size="xl" />
            <div class="min-w-0">
              <h2 class="person-detail__name">{{ name }}</h2>
              <p v-if="roles[0]" class="person-detail__title">
                {{ roles[0].position.title }}
                <span v-if="roles[0].assignment.is_interim" class="dir-interim">{{ $t('directory.interim') }}</span>
              </p>
              <span v-if="!person.is_active" class="badge badge-secondary mt-2">{{ $t('common.inactive') }}</span>
            </div>
          </div>

          <div class="modal-content">
            <div class="grid gap-6 md:grid-cols-2">
              <section>
                <h3 class="person-detail__section">{{ $t('directory.contact') }}</h3>
                <dl class="person-detail__list">
                  <div v-if="person.email">
                    <dt><Icon icon="mdi:email-outline" /> {{ $t('directory.email') }}</dt>
                    <dd><a :href="`mailto:${person.email}`">{{ person.email }}</a></dd>
                  </div>
                  <div v-if="person.phone">
                    <dt><Icon icon="mdi:phone-outline" /> {{ $t('directory.phone') }}</dt>
                    <dd><a :href="`tel:${person.phone.replace(/\s+/g, '')}`">{{ person.phone }}</a></dd>
                  </div>
                  <div v-for="f in filledFields" :key="f.key">
                    <dt>
                      <Icon :icon="FIELD_TYPE_ICONS[f.type]" /> {{ f.label }}
                      <Icon v-if="f.visibility === 'managers'" icon="mdi:lock-outline" class="text-amber-500" :title="$t('directory.fieldManagersOnly')" />
                    </dt>
                    <dd :class="{ 'whitespace-pre-line': f.type === 'textarea' }">
                      <a v-if="fieldHref(f, value(f))" :href="fieldHref(f, value(f))" target="_blank" rel="noopener noreferrer">
                        {{ formatFieldValue(f, value(f), locale) }}
                      </a>
                      <template v-else>{{ formatFieldValue(f, value(f), locale) }}</template>
                    </dd>
                  </div>
                  <p v-if="!person.email && !person.phone && !filledFields.length" class="text-sm text-gray-400 italic">
                    {{ $t('directory.noDetails') }}
                  </p>
                </dl>
              </section>

              <section>
                <h3 class="person-detail__section">{{ $t('directory.positions') }}</h3>
                <p v-if="!roles.length" class="text-sm text-gray-400 italic">{{ $t('directory.noPosition') }}</p>
                <ul v-else class="space-y-3">
                  <li v-for="r in roles" :key="r.assignment.id" class="person-detail__role" :style="{ borderColor: r.unit.color }">
                    <p class="font-semibold flex items-center gap-2 flex-wrap" style="color: var(--text-primary)">
                      <Icon v-if="r.position.is_head" icon="mdi:star-four-points" class="text-amber-500" :title="$t('directory.head')" />
                      {{ r.position.title }}
                      <span v-if="r.assignment.is_interim" class="dir-interim">{{ $t('directory.interim') }}</span>
                    </p>
                    <p class="text-xs mt-1" style="color: var(--text-muted)">
                      <template v-for="(u, i) in unitPath(r.unit.id)" :key="u.id">
                        <span v-if="i > 0"> › </span>{{ u.name }}
                      </template>
                    </p>
                    <p v-if="r.assignment.start_date" class="text-xs mt-1" style="color: var(--text-muted)">
                      {{ $t('directory.since') }} {{ formatDate(r.assignment.start_date) }}
                    </p>
                  </li>
                </ul>
              </section>
            </div>
          </div>

          <div class="modal-footer-bar">
            <span class="modal-footer-hint" />
            <div class="modal-footer-actions">
              <button v-if="canManage" type="button" class="btn btn-secondary" @click="$emit('edit', person)">
                <Icon icon="mdi:pencil" class="h-4 w-4 mr-1" /> {{ $t('common.edit') }}
              </button>
              <button type="button" class="btn btn-primary" @click="$emit('close')">{{ $t('common.close') }}</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import PersonAvatar from './PersonAvatar.vue'
import { fullName, formatFieldValue, fieldHref, FIELD_TYPE_ICONS } from './useDirectory'

const props = defineProps({
  person: { type: Object, required: true },
  roles: { type: Array, default: () => [] },
  fields: { type: Array, default: () => [] },
  unitPath: { type: Function, required: true },
  canManage: { type: Boolean, default: false },
})
defineEmits(['close', 'edit'])

const { locale } = useI18n()
const name = computed(() => fullName(props.person))
const value = (f) => props.person.custom_fields?.[f.key]
const filledFields = computed(() => props.fields.filter((f) => value(f)))
const formatDate = (iso) => new Date(iso).toLocaleDateString(locale.value)
</script>

<style scoped>
.person-detail__hero {
  position: relative;
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 2rem 1.5rem 1.5rem;
  background:
    radial-gradient(120% 140% at 0% 0%, color-mix(in srgb, var(--unit-color) 22%, transparent), transparent 60%),
    var(--bg-surface);
  border-bottom: 1px solid var(--border);
}
.person-detail__close {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  padding: 0.4rem;
  border-radius: 9999px;
  color: var(--text-muted);
}
.person-detail__close:hover { background: var(--hover); color: var(--text-primary); }
[dir='rtl'] .person-detail__close { right: auto; left: 0.75rem; }
.person-detail__name { font-size: 1.6rem; font-weight: 800; line-height: 1.15; color: var(--text-primary); }
.person-detail__title { margin-top: 0.25rem; font-weight: 500; color: var(--accent); display: flex; gap: 0.5rem; align-items: center; flex-wrap: wrap; }
.person-detail__section {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 0.75rem;
}
.person-detail__list { display: flex; flex-direction: column; gap: 0.75rem; }
.person-detail__list dt { display: flex; align-items: center; gap: 0.35rem; font-size: 0.75rem; color: var(--text-muted); }
.person-detail__list dd { font-size: 0.95rem; color: var(--text-primary); word-break: break-word; }
.person-detail__list a { color: var(--accent); }
.person-detail__list a:hover { text-decoration: underline; }
.person-detail__role { padding-inline-start: 0.75rem; border-inline-start: 3px solid; }
@media (max-width: 640px) {
  .person-detail__hero { flex-direction: column; text-align: center; }
}
</style>
