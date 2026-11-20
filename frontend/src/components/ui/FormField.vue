<template>
  <div class="field">
    <label :for="fieldId" class="field__label">
      {{ label }}<span v-if="required" class="field__req" aria-hidden="true">*</span>
    </label>
    <!-- Le slot reçoit id et describedby à poser sur le champ : <input :id="id" :aria-describedby="describedby"> -->
    <slot :id="fieldId" :describedby="describedby" :invalid="Boolean(error)" />
    <p v-if="error" :id="`${fieldId}-error`" class="field__error" role="alert">{{ error }}</p>
    <p v-else-if="hint" :id="`${fieldId}-hint`" class="field__hint">{{ hint }}</p>
  </div>
</template>

<script>
// Compteur partagé par toutes les instances : identifiants uniques dans la page.
let fieldCounter = 0
</script>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  label: { type: String, required: true },
  id: { type: String, default: '' },
  hint: { type: String, default: '' },
  error: { type: String, default: '' },
  required: { type: Boolean, default: false },
})

fieldCounter += 1
const autoId = `field-${fieldCounter}`
const fieldId = computed(() => props.id || autoId)

const describedby = computed(() => {
  if (props.error) return `${fieldId.value}-error`
  if (props.hint) return `${fieldId.value}-hint`
  return undefined
})
</script>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: 5px;
  margin-block-end: 14px;
}

.field__label {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--text);
}

.field__req {
  margin-inline-start: 3px;
  color: var(--bad);
}

.field__hint,
.field__error {
  margin: 0;
  font-size: 12px;
}

.field__hint { color: var(--muted); }
.field__error { color: var(--bad); }
</style>
