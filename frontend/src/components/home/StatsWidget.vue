<template>
  <div class="stats-widget">
    <!-- Widget Header -->
    <div class="widget-header">
      <Icon icon="mdi:chart-box-outline" class="header-icon" />
      <h3 class="widget-title">{{ $t('home.stats.title') }}</h3>
    </div>

    <!-- Admin Stats - Compact List -->
    <div v-if="role === 'admin'" class="stats-list">
      <div class="stat-row clickable" @click="navigateTo('/admin/users')">
        <Icon icon="mdi:account-multiple" class="stat-icon stat-icon-blue" />
        <span class="stat-value">{{ stats.total_users || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.users') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
      <div class="stat-row clickable" @click="navigateTo('/admin/groups')">
        <Icon icon="mdi:account-group" class="stat-icon stat-icon-purple" />
        <span class="stat-value">{{ stats.total_groups || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.groups') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
      <div class="stat-row clickable" @click="navigateTo('/admin/app-groups')">
        <Icon icon="mdi:folder-multiple" class="stat-icon stat-icon-pink" />
        <span class="stat-value">{{ stats.total_app_groups || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.appGroups') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
      <div class="stat-row clickable" @click="navigateTo('/admin/applications')">
        <Icon icon="mdi:application" class="stat-icon stat-icon-green" />
        <span class="stat-value">{{ stats.total_apps || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.apps') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
      <div class="stat-row clickable" @click="navigateTo('/admin/news')">
        <Icon icon="mdi:newspaper-variant-outline" class="stat-icon stat-icon-orange" />
        <span class="stat-value">{{ stats.total_news || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.news') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
      <div class="stat-row clickable" @click="navigateTo('/admin/events')">
        <Icon icon="mdi:calendar-star" class="stat-icon stat-icon-red" />
        <span class="stat-value">{{ stats.total_events || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.events') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
      <div class="stat-row clickable" @click="navigateTo('/admin/polls')">
        <Icon icon="mdi:poll" class="stat-icon stat-icon-yellow" />
        <span class="stat-value">{{ stats.total_polls || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.polls') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
    </div>

    <!-- Group Admin Stats - Compact List -->
    <div v-else-if="role === 'group_admin'" class="stats-list">
      <div class="stat-row clickable" @click="navigateTo('/group-admin/dashboard')">
        <Icon icon="mdi:account-group" class="stat-icon stat-icon-indigo" />
        <span class="stat-value">{{ stats.managed_groups_count || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.managedGroups') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
      <div class="stat-row">
        <Icon icon="mdi:account-multiple" class="stat-icon stat-icon-cyan" />
        <span class="stat-value">{{ stats.total_members_count || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.totalMembers') }}</span>
      </div>
      <div class="stat-row clickable" @click="navigateTo('/group-admin/app-groups')">
        <Icon icon="mdi:folder-multiple" class="stat-icon stat-icon-pink" />
        <span class="stat-value">{{ stats.managed_app_groups_count || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.managedAppGroups') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
      <div class="stat-row clickable" @click="navigateTo('/group-admin/applications')">
        <Icon icon="mdi:application" class="stat-icon stat-icon-teal" />
        <span class="stat-value">{{ stats.managed_apps_count || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.managedApps') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
      <div class="stat-row clickable" @click="navigateTo('/group-admin/news')">
        <Icon icon="mdi:newspaper-variant-outline" class="stat-icon stat-icon-orange" />
        <span class="stat-value">{{ stats.managed_news_count || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.managedNews') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
      <div class="stat-row clickable" @click="navigateTo('/group-admin/polls')">
        <Icon icon="mdi:poll" class="stat-icon stat-icon-yellow" />
        <span class="stat-value">{{ stats.managed_polls_count || 0 }}</span>
        <span class="stat-label">{{ $t('home.stats.managedPolls') }}</span>
        <Icon icon="mdi:chevron-right" class="stat-arrow" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { Icon } from '@iconify/vue'
import { useRouter } from 'vue-router'

defineProps({
  stats: {
    type: Object,
    required: true
  },
  role: {
    type: String,
    required: true
  }
})

const router = useRouter()

const navigateTo = (path) => {
  router.push(path)
}
</script>

<style scoped>
.stats-widget {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Widget Header */
.widget-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(102, 126, 234, 0.1);
}

.dark .widget-header {
  border-bottom-color: rgba(102, 126, 234, 0.2);
}

.header-icon {
  font-size: 1.125rem;
  color: #667eea;
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

/* Stats List - Compact Layout */
.stats-list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  flex: 1;
}

.stat-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.5rem;
  background: rgba(102, 126, 234, 0.05);
  border-radius: 8px;
  transition: all 0.3s ease;
  position: relative;
}

.dark .stat-row {
  background: rgba(102, 126, 234, 0.1);
}

.stat-row.clickable {
  cursor: pointer;
}

.stat-row:hover {
  background: rgba(102, 126, 234, 0.1);
  transform: translateX(4px);
}

.dark .stat-row:hover {
  background: rgba(102, 126, 234, 0.15);
}

.stat-row.clickable:hover {
  background: rgba(102, 126, 234, 0.12);
  transform: translateX(6px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
}

.dark .stat-row.clickable:hover {
  background: rgba(102, 126, 234, 0.2);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.25);
}

.stat-row.clickable:active {
  transform: translateX(2px);
}

.stat-arrow {
  font-size: 1rem;
  color: #9ca3af;
  margin-left: auto;
  transition: all 0.3s ease;
}

.stat-row.clickable:hover .stat-arrow {
  color: #667eea;
  transform: translateX(4px);
}

.dark .stat-arrow {
  color: #6b7280;
}

.dark .stat-row.clickable:hover .stat-arrow {
  color: #818cf8;
}

.stat-icon {
  font-size: 1.25rem;
  flex-shrink: 0;
}

.stat-value {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1f2937;
  min-width: 2.5rem;
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.dark .stat-value {
  color: white;
}

.stat-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: #6b7280;
  flex: 1;
}

.dark .stat-label {
  color: #9ca3af;
}

/* Icon Colors */
.stat-icon-blue {
  color: #3b82f6;
}

.stat-icon-purple {
  color: #8b5cf6;
}

.stat-icon-green {
  color: #10b981;
}

.stat-icon-orange {
  color: #f59e0b;
}

.stat-icon-indigo {
  color: #6366f1;
}

.stat-icon-cyan {
  color: #06b6d4;
}

.stat-icon-teal {
  color: #14b8a6;
}

.stat-icon-pink {
  color: #ec4899;
}

.stat-icon-red {
  color: #ef4444;
}

.stat-icon-yellow {
  color: #eab308;
}

/* Responsive */
@media (max-width: 640px) {
  .widget-header {
    margin-bottom: 0.375rem;
    padding-bottom: 0.375rem;
  }

  .header-icon {
    font-size: 1rem;
  }

  .widget-title {
    font-size: 0.875rem;
  }

  .stat-row {
    gap: 0.375rem;
    padding: 0.3rem 0.4rem;
    border-radius: 6px;
  }

  .stat-icon {
    font-size: 1.125rem;
  }

  .stat-value {
    font-size: 1rem;
    min-width: 2rem;
  }

  .stat-label {
    font-size: 0.6875rem;
  }

  .stat-arrow {
    font-size: 0.875rem;
  }
}
</style>
