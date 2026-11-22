<template>
  <!-- Le nom porte l'action d'ouverture ; son ::after couvre toute la carte.
       Les liens de contact restent hors du bouton (HTML valide) et passent au-dessus. -->
  <article
    class="person-card"
    :class="{ 'person-card--inactive': !person.is_active }"
    :style="{ '--unit-color': primaryRole?.unit.color || 'var(--line)' }"
  >
    <PersonAvatar :person="person" size="lg" class="person-card__avatar" />
    <button type="button" class="person-card__name" @click="$emit('open', person)">{{ name }}</button>
    <span v-if="primaryRole" class="person-card__role">
      {{ primaryRole.position.title }}
      <span v-if="primaryRole.assignment.is_interim" class="dir-interim">{{ $t('directory.interim') }}</span>
    </span>
    <span v-if="primaryRole" class="person-card__unit">{{ primaryRole.unit.name }}</span>
    <span v-else class="person-card__unit person-card__unit--none">{{ $t('directory.noPosition') }}</span>
    <span v-if="roles.length > 1" class="person-card__more">{{ $t('directory.otherPositions', { count: roles.length - 1 }) }}</span>

    <span v-if="cardFields.length" class="person-card__fields">
      <span v-for="f in cardFields" :key="f.key" class="person-card__field">
        <Icon :icon="FIELD_TYPE_ICONS[f.type]" class="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
        <span class="truncate">{{ formatFieldValue(f, person.custom_fields[f.key], locale) }}</span>
      </span>
    </span>

    <span v-if="person.email || person.phone" class="person-card__contacts">
      <a v-if="person.email" :href="`mailto:${person.email}`" :title="person.email" :aria-label="person.email">
        <Icon icon="lucide:mail" class="h-4 w-4" aria-hidden="true" />
      </a>
      <a v-if="person.phone" :href="`tel:${person.phone.replace(/\s+/g, '')}`" :title="person.phone" :aria-label="person.phone">
        <Icon icon="lucide:phone" class="h-4 w-4" aria-hidden="true" />
      </a>
    </span>
  </article>
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
  gap: 2px;
  padding: 18px 14px 14px;
  border: 1px solid var(--line);
  border-block-start: 3px solid var(--unit-color);
  border-radius: var(--r-s);
  background: var(--surface);
  color: var(--text);
  text-align: center;
  cursor: pointer;
  transition: border-color 150ms ease;
}
.person-card:hover { border-inline-color: var(--muted); border-block-end-color: var(--muted); }
.person-card__name {
  padding: 0;
  border: 0;
  background: none;
  color: inherit;
  font: inherit;
  font-weight: 600;
  font-size: 15px;
  line-height: 1.25;
  cursor: pointer;
}
/* Zone cliquable étendue à toute la carte */
.person-card__name::after {
  content: '';
  position: absolute;
  inset: -3px -1px -1px;
  border-radius: var(--r-s);
}
.person-card__name:focus-visible { outline: none; }
.person-card__name:focus-visible::after { outline: 2px solid var(--acc); outline-offset: 2px; }
.person-card--inactive { opacity: 0.55; }
.person-card__avatar { margin-bottom: 8px; }
.person-card__role {
  display: inline-flex;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  color: var(--text);
}
.person-card__unit { font-size: 12px; color: var(--muted); }
.person-card__unit--none { font-style: italic; }
.person-card__more { font-family: var(--f-num); font-size: 11px; color: var(--muted); }
.person-card__fields {
  display: flex;
  flex-direction: column;
  gap: 3px;
  width: 100%;
  margin-top: 9px;
  padding-top: 9px;
  border-top: 1px solid var(--line);
}
.person-card__field {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  min-width: 0;
  font-size: 12px;
  color: var(--muted);
}
.person-card__contacts { position: relative; z-index: 1; display: flex; gap: 6px; margin-top: 11px; }
.person-card__contacts a {
  display: inline-grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  color: var(--muted);
  transition: color 150ms ease, border-color 150ms ease;
}
.person-card__contacts a:hover { color: var(--text); border-color: var(--muted); }
.person-card__contacts a:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

@media (prefers-reduced-motion: reduce) {
  .person-card, .person-card__contacts a { transition: none; }
}
</style>
