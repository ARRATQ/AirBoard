<template>
  <div class="news-type-widget">
    <!-- Widget Header -->
    <div class="widget-header">
      <div class="header-left">
        <Icon
          :icon="newsType.icon || 'mdi:newspaper'"
          class="header-icon"
          :style="{ color: newsType.color || '#3B82F6' }"
        />
        <h3 class="widget-title">{{ newsType.name }}</h3>
      </div>
      <router-link :to="`/news?type=${newsType.slug}`" class="view-all-link">
        <span>{{ $t('home.viewAll') }}</span>
        <Icon icon="mdi:arrow-right" class="link-icon" />
      </router-link>
    </div>

    <!-- Empty State -->
    <div v-if="news.length === 0" class="empty-state">
      <Icon icon="mdi:newspaper-variant-outline" class="empty-icon" />
      <p class="empty-text">{{ $t('home.noNews') }}</p>
    </div>

    <!-- News List -->
    <div v-else class="news-list">
      <NewsListItem
        v-for="article in news"
        :key="article.id"
        :article="article"
        @click="goToNews(article)"
      />
    </div>
  </div>
</template>

<script setup>
import { Icon } from '@iconify/vue'
import { useRouter } from 'vue-router'
import NewsListItem from './NewsListItem.vue'

const router = useRouter()

const props = defineProps({
  newsType: {
    type: Object,
    required: true
  },
  news: {
    type: Array,
    required: true,
    default: () => []
  }
})

const goToNews = (article) => {
  router.push(`/news/${article.slug}`)
}
</script>

<style scoped>
.news-type-widget {
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
  border-bottom: 1px solid rgba(249, 115, 22, 0.1);
}

.dark .widget-header {
  border-bottom-color: rgba(249, 115, 22, 0.2);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.header-icon {
  font-size: 1.125rem;
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

/* News List */
.news-list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
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
