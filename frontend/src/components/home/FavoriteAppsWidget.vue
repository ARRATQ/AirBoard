<template>
  <div class="favorites-widget">
    <!-- Widget Header -->
    <div class="widget-header">
      <div class="header-left">
        <div class="header-icon-wrapper">
          <Icon icon="mdi:star" class="header-icon" />
        </div>
        <h3 class="widget-title">{{ $t('home.favorites.title') }}</h3>
      </div>
      <router-link to="/dashboard" class="view-all-link">
        <span>{{ $t('home.viewAll') }}</span>
        <Icon icon="mdi:arrow-right" class="link-icon" />
      </router-link>
    </div>

    <!-- Empty State -->
    <div v-if="apps.length === 0" class="empty-state">
      <div class="empty-icon-wrapper">
        <Icon icon="mdi:star-outline" class="empty-icon" />
      </div>
      <p class="empty-text">{{ $t('home.favorites.empty') }}</p>
      <router-link to="/dashboard" class="empty-cta">
        {{ $t('home.favorites.browseApps') }}
      </router-link>
    </div>

    <!-- Apps Grid - Same style as GridView -->
    <div v-else class="apps-grid">
      <div
        v-for="app in props.apps.slice(0, 6)"
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
            <Icon
              v-if="app.open_in_new_tab"
              icon="mdi:open-in-new"
              class="h-3 w-3 text-gray-400 dark:text-gray-500 flex-shrink-0"
            />
          </div>
          <p v-if="app.description" class="text-xs text-gray-600 dark:text-gray-400 truncate mt-0.5">
            {{ app.description }}
          </p>
        </div>

        <!-- Favorite Star Indicator -->
        <div class="p-1 rounded flex-shrink-0">
          <Icon icon="mdi:star" class="h-4 w-4 text-yellow-500" />
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
.favorites-widget {
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
  border-bottom: 1px solid rgba(251, 191, 36, 0.1);
}

.dark .widget-header {
  border-bottom-color: rgba(251, 191, 36, 0.2);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.header-icon {
  font-size: 1.125rem;
  color: #fbbf24;
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

.empty-icon-wrapper {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%);
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

.dark .empty-icon-wrapper {
  background: linear-gradient(135deg, #374151 0%, #1f2937 100%);
}

.empty-icon {
  font-size: 2rem;
  color: #9ca3af;
}

.empty-text {
  font-size: 0.875rem;
  color: #6b7280;
  margin-bottom: 1rem;
}

.dark .empty-text {
  color: #9ca3af;
}

.empty-cta {
  display: inline-flex;
  align-items: center;
  padding: 0.625rem 1.25rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 0.875rem;
  font-weight: 600;
  border-radius: 10px;
  text-decoration: none;
  transition: all 0.3s ease;
}

.empty-cta:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(102, 126, 234, 0.3);
}

/* Apps Grid */
.apps-grid {
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

/* Responsive */
@media (max-width: 640px) {
  .widget-header {
    margin-bottom: 1rem;
  }

  .header-icon {
    font-size: 1.25rem;
  }

  .widget-title {
    font-size: 1rem;
  }
}
</style>
