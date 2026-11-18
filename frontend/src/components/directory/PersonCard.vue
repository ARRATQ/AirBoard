<template>
  <button type="button" class="person-card" :class="{ 'person-card--inactive': !person.is_active }" @click="$emit('open', person)">
    <span class="person-card__band" :style="{ backgroundColor: primaryRole?.unit.color || 'var(--accent)' }" />
    <PersonAvatar :person="person" size="lg" class="person-card__avatar" />
    <span class="person-card__name">{{ name }}</span>
    <span v-if="primaryRole" class="person-card__role">
      {{ primaryRole.position.title }}
      <span v-if="primaryRole.assignment.is_interim" class="dir-interim">{{ $t('directory.interim') }}</span>
    </span>
    <span v-if="primaryRole" class="person-card__unit">{{ primaryRole.unit.name }}</span>
    <span v-else class="person-card__unit person-card__unit--none">{{ $t('directory.noPosition') }}</span>
    <span v-if="roles.length > 1" class="person-card__more">+{{ roles.length - 1 }} {{ $t('directory.otherPositions') }}</span>

    <span v-if="cardFields.length" class="person-card__fields">
      <span v-for="f in cardFields" :key="f.key" class="person-card__field">
        <Icon :icon="FIELD_TYPE_ICONS[f.type]" class="h-3.5 w-3.5 shrink-0" />
        <span class="truncate">{{ formatFieldValue(f, person.custom_fields[f.key], locale) }}</span>
      </span>
    </span>

    <span class="person-card__contacts">
      <a v-if="person.email" :href="`mailto:${person.email}`" :title="person.email" @click.stop>
        <Icon icon="mdi:email-outline" class="h-4 w-4" />
      </a>
      <a v-if="person.phone" :href="`tel:${person.phone.replace(/\s+/g, '')}`" :title="person.phone" @click.stop>
        <Icon icon="mdi:phone-outline" class="h-4 w-4" />
      </a>
    </span>
  </button>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import PersonAvatar from './PersonAvatar.vue'
import { fullName, formatFieldValue, FIELD_TYPE_ICONS } from './useDirectory'

const props = defineProps({
  person: { type: Object, required: true },
  roles: { type: Array, default: () => [] },
  fields: { type: Array, default: () => [] },
})
defineEmits(['open'])

const { locale } = useI18n()
const name = computed(() => fullName(props.person))
const primaryRole = computed(() => props.roles[0] || null)
const cardFields = computed(() =>
  props.fields.filter((f) => f.show_on_card && props.person.custom_fields?.[f.key])
)
</script>

<style scoped>
.person-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.15rem;
  padding: 1.5rem 1rem 1rem;
  border: 1px solid var(--border);
  border-radius: 1rem;
  background: var(--bg-card);
  color: var(--text-primary);
  text-align: center;
  overflow: hidden;
  transition: transform 180ms cubic-bezier(0.16, 1, 0.3, 1), box-shadow 180ms ease, border-color 180ms ease;
}
.person-card:hover {
  transform: translateY(-3px);
  border-color: var(--accent);
  box-shadow: 0 12px 28px -12px rgba(var(--shadow-rgb, 15, 23, 42), 0.35);
}
.person-card:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}
.person-card--inactive { opacity: 0.55; }
.person-card__band {
  position: absolute;
  inset: 0 0 auto 0;
  height: 3.25rem;
  opacity: 0.14;
}
.person-card__avatar { position: relative; margin-bottom: 0.6rem; }
.person-card__name { font-weight: 700; font-size: 1rem; line-height: 1.25; }
.person-card__role { font-size: 0.85rem; font-weight: 500; color: var(--accent); display: inline-flex; gap: 0.35rem; align-items: center; flex-wrap: wrap; justify-content: center; }
.person-card__unit { font-size: 0.78rem; color: var(--text-muted); }
.person-card__unit--none { font-style: italic; }
.person-card__more { font-size: 0.72rem; color: var(--text-muted); }
.person-card__fields {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  width: 100%;
  margin-top: 0.6rem;
  padding-top: 0.6rem;
  border-top: 1px dashed var(--border);
}
.person-card__field { display: flex; align-items: center; gap: 0.35rem; justify-content: center; font-size: 0.75rem; color: var(--text-muted); min-width: 0; }
.person-card__contacts { display: flex; gap: 0.5rem; margin-top: 0.75rem; }
.person-card__contacts a {
  display: inline-flex;
  padding: 0.4rem;
  border-radius: 9999px;
  border: 1px solid var(--border);
  color: var(--text-muted);
  transition: color 150ms ease, border-color 150ms ease, background-color 150ms ease;
}
.person-card__contacts a:hover { color: var(--accent); border-color: var(--accent); background: var(--hover); }
</style>
