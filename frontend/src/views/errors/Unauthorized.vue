<template>
  <ErrorPage
    code="403"
    icon="lucide:lock"
    tone="bad"
    :title="t('errors.unauthorized.title')"
    :message="t('errors.unauthorized.message')"
  />
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import ErrorPage from './ErrorPage.vue'

const { t } = useI18n()
</script>
