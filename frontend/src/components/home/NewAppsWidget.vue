<template>
  <div class="new-apps-widget">
    <!-- Widget Header -->
    <div class="widget-header">
      <div class="header-left">
        <Icon icon="mdi:new-box" class="header-icon" />
        <h3 class="widget-title">{{ $t('home.newApps.title') }}</h3>
      </div>
      <router-link to="/dashboard" class="view-all-link">
        <span>{{ $t('home.viewAll') }}</span>
        <Icon icon="mdi:arrow-right" class="link-icon" />
      </router-link>
    </div>

    <!-- Empty State -->
    <div v-if="props.apps.length === 0" class="empty-state">
      <Icon icon="mdi:application" class="empty-icon" />
      <p class="empty-text">{{ $t('home.newApps.empty') }}</p>
    </div>

    <!-- Apps List - Same style as GridView -->
    <div v-else class="apps-list">
      <div
        v-for="app in props.apps"
        :key="app.id"
        class="app-list-item"
        @click="openApp(app)"
      >
        <!-- App Icon -->
        <div
          class="h-9 w-9 rounded-lg flex items-center justify-center flex-shrink-0 shadow-sm"
          :style="{ backgroundColor: app.color || '#6366f1' }"
        >
          <Icon :icon="app.icon || 'mdi:application'" class="h-4 w-4 text-white" />
        </div>

        <!-- App Info -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-1">
            <h3 class="font-semibold text-xs text-gray-900 dark:text-white truncate">
              {{ app.name }}
            </h3>
            <span class="new-badge">
              {{ $t('home.newApps.newBadge') }}
            </span>
            <Icon
              v-if="app.open_in_new_tab"
              icon="mdi:open-in-new"
              class="h-3 w-3 text-gray-400 dark:text-gray-500 flex-shrink-0"
            />
          </div>
          <p v-if="app.app_group" class="text-xs text-gray-600 dark:text-gray-400 truncate mt-0.5">
            {{ app.app_group.name }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Icon } from '@iconify/vue'

const props = defineProps({
  apps: {
    type: Array,
    default: () => []
  }
})

const openApp = (app) => {
  if (app.open_in_new_tab) {
    window.open(app.url, '_blank', 'noopener,noreferrer')
  } else {
    window.location.href = app.url
  }
}
</script>

<style scoped>
.new-apps-widget {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Widget Header */
.widget-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(16, 185, 129, 0.1);
}

.dark .widget-header {
  border-bottom-color: rgba(16, 185, 129, 0.2);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.header-icon {
  font-size: 1.125rem;
  color: #10b981;
}

.widget-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
  line-height: 1.2;
}

.dark .widget-title {
  color: white;
}

.view-all-link {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: #6366f1;
  text-decoration: none;
  transition: all 0.2s ease;
}

.view-all-link:hover {
  color: #4f46e5;
  gap: 0.5rem;
}

.link-icon {
  font-size: 1rem;
  transition: transform 0.2s ease;
}

.view-all-link:hover .link-icon {
  transform: translateX(2px);
}

/* Empty State */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1rem 0.5rem;
  text-align: center;
}

.empty-icon {
  font-size: 2rem;
  color: #9ca3af;
  margin-bottom: 0.375rem;
}

.empty-text {
  font-size: 0.875rem;
  color: #6b7280;
}

.dark .empty-text {
  color: #9ca3af;
}

/* Apps List */
.apps-list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

/* List Item - Same style as GridView */
.app-list-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.5rem;
  background-color: rgb(255 255 255);
  border: 1px solid rgb(229 231 235);
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.app-list-item:hover {
  background-color: rgb(249 250 251);
  border-color: rgb(209 213 219);
  box-shadow: 0 2px 4px 0 rgb(0 0 0 / 0.05);
  transform: translateX(2px);
}

.dark .app-list-item {
  background-color: rgb(31 41 55);
  border-color: rgb(55 65 81);
}

.dark .app-list-item:hover {
  background-color: rgb(55 65 81);
  border-color: rgb(75 85 99);
}

/* New Badge */
.new-badge {
  padding: 0.125rem 0.375rem;
  font-size: 0.625rem;
  font-weight: 600;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border-radius: 9999px;
  flex-shrink: 0;
  line-height: 1;
}

/* Responsive */
@media (max-width: 640px) {
  .widget-header {
    margin-bottom: 1rem;
    padding-bottom: 0.75rem;
  }

  .header-icon {
    font-size: 1.25rem;
  }

  .widget-title {
    font-size: 1rem;
  }
}
</style>
