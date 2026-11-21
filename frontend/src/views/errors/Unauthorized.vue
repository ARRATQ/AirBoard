<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center px-4">
    <div class="max-w-md w-full text-center">
      <div class="mb-8">
        <Icon aria-hidden="true" icon="mdi:lock" class="mx-auto h-24 w-24 text-red-400" />
      </div>
      
      <h1 class="text-6xl font-bold text-gray-900 dark:text-white mb-4">403</h1>
      
      <h2 class="text-2xl font-semibold text-gray-700 dark:text-gray-300 mb-4">
        {{ t('errors.unauthorized.title') }}
      </h2>
      
      <p class="text-gray-600 dark:text-gray-400 mb-8">
        {{ t('errors.unauthorized.message') }}
      </p>
      
      <div class="space-y-4">
        <router-link to="/home" class="btn-primary btn-md w-full">
          <Icon aria-hidden="true" icon="mdi:home" class="h-5 w-5 me-2" />
          {{ t('errors.backHome') }}
        </router-link>
        
        <button type="button" @click="$router.back()" class="btn-secondary btn-md w-full">
          <Icon aria-hidden="true" icon="mdi:arrow-left" class="rtl-flip h-5 w-5 me-2" />
          {{ t('errors.previousPage') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>