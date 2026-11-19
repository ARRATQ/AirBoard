<template>
  <header class="page-head">
    <div class="page-head__text">
      <p v-if="eyebrow" class="page-head__eyebrow">{{ eyebrow }}</p>
      <h1 class="page-head__title">
        <slot name="title">{{ title }}</slot>
      </h1>
      <p v-if="description || $slots.description" class="page-head__desc">
        <slot name="description">{{ description }}</slot>
      </p>
    </div>
    <div v-if="$slots.actions" class="page-head__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup>
// En-tête de page A+ : surtitre mono, titre 22 px/600, description, actions alignées en fin.
defineProps({
  eyebrow: { type: String, default: '' },
  title: { type: String, default: '' },
  description: { type: String, default: '' },
})
</script>

<style scoped>
.page-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 14px;
  margin-block-end: 20px;
  padding-block-end: 14px;
  border-block-end: 1px solid var(--line);
}

.page-head__text { min-width: 0; }

.page-head__eyebrow {
  margin: 0 0 4px;
  font-family: var(--f-num);
  font-size: 11px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
}

.page-head__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  line-height: 1.25;
  letter-spacing: -0.01em;
  color: var(--text);
}

.page-head__desc {
  margin: 4px 0 0;
  color: var(--muted);
  max-width: 75ch;
}

.page-head__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
</style>
