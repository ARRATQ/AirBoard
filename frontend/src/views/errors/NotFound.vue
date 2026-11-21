<template>
  <ErrorPage code="404" icon="mdi:compass-off" :title="t('errors.notFound.title')" :message="t('errors.notFound.message')" />
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import ErrorPage from './ErrorPage.vue'

const { t } = useI18n()
</script>
