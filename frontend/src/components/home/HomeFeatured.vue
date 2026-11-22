<template>
  <section class="featured" aria-labelledby="home-featured-title">
    <div class="home-card__head featured__head">
      <h2 id="home-featured-title" class="home-card__title">{{ $t('home.featured.title') }}</h2>
      <router-link to="/news" class="home-card__more">
        {{ $t('home.viewAll') }} <span class="home-card__arrow" aria-hidden="true">→</span>
      </router-link>
    </div>

    <ul class="featured__grid">
      <li v-for="article in articles" :key="article.id" class="featured__item">
        <article class="fcard" :style="{ '--tc': typeOf(article).color }">
          <img
            v-if="article.cover_image"
            :src="article.cover_image"
            alt=""
            class="fcard__cover"
            loading="lazy"
            width="640"
            height="240"
          />
          <div class="fcard__body">
            <p class="fcard__meta">
              <span v-if="typeOf(article).name" class="tag tag--custom fcard__type">{{ typeOf(article).name }}</span>
              <span v-if="article.category" class="fcard__cat">
                <Icon :icon="article.category.icon || 'lucide:folder'" :style="{ color: article.category.color }" aria-hidden="true" />
                <span dir="auto">{{ article.category.name }}</span>
              </span>
            </p>
            <h3 class="fcard__title">
              <router-link :to="`/news/${article.slug}`" class="fcard__link">{{ article.title }}</router-link>
            </h3>
            <p v-if="article.summary" class="fcard__summary">{{ article.summary }}</p>
            <p class="fcard__foot">
              <span v-if="authorOf(article)" class="fcard__author" dir="auto">{{ authorOf(article) }}</span>
              <time :datetime="article.published_at || article.created_at">
                {{ relativeDay(article.published_at || article.created_at, dateLocale) }}
              </time>
              <span v-if="minutesOf(article)">{{ minutesOf(article) }}</span>
            </p>
          </div>
        </article>
      </li>
    </ul>
  </section>
</template>

<script setup>
// « À la une » : les deux derniers articles publiés, en cartes. La carte entière ouvre
// l'article (lien du titre étendu). Couleur et nom du type viennent des types connus de
// l'accueil (onglets du fil), sinon d'une couleur neutre.
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { dateLocaleFor, relativeDay } from '@/utils/homeFormat'
import { readingMinutes } from '@/utils/readingTime'
import { newsTypeOf } from '@/utils/homeHighlights'

const props = defineProps({
  articles: { type: Array, required: true },
  types: { type: Array, default: () => [] },
})

const FALLBACK_COLOR = '#64748b'
const { t, locale } = useI18n()
const dateLocale = computed(() => dateLocaleFor(locale.value))

const typeOf = (article) => {
  const type = newsTypeOf(article, props.types)
  return { name: type?.name || '', color: type?.color || FALLBACK_COLOR }
}
const authorOf = (article) => article.author?.full_name || article.author?.username || ''
const minutesOf = (article) => {
  const n = article.reading_time || readingMinutes(article.content)
  return n ? t('news.minutes', { n }) : ''
}
</script>

<style scoped>
.featured__head { margin-block-end: 10px; }

.featured__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 16px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.featured__item { min-width: 0; }

.fcard {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  border: 1px solid var(--line);
  border-top: 3px solid var(--tc);
  border-radius: var(--r-l);
  background: var(--surface);
  transition: border-color 150ms ease, box-shadow 150ms ease;
}

.fcard:hover {
  border-color: color-mix(in oklab, var(--tc) 45%, var(--line));
  box-shadow: 0 4px 14px -6px rgb(var(--shadow-rgb) / 0.25);
}

.fcard:focus-within { outline: 2px solid var(--acc); outline-offset: 2px; }

.fcard__cover {
  display: block;
  width: 100%;
  height: auto;
  aspect-ratio: 8 / 3;
  object-fit: cover;
  background: var(--sunken);
}

.fcard__body {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
  padding: 14px 16px 12px;
}

.fcard__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px 10px;
  margin: 0;
  font-size: 12px;
  color: var(--muted);
}

.fcard__type { --tag-c: var(--tc); }

.fcard__cat { display: inline-flex; align-items: center; gap: 4px; min-width: 0; }
.fcard__cat :deep(svg) { flex: none; width: 14px; height: 14px; }

.fcard__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  line-height: 1.3;
  color: var(--text);
}

.fcard__link {
  display: -webkit-box;
  overflow: hidden;
  color: inherit;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

/* Toute la carte est cliquable */
.fcard__link::after { position: absolute; inset: 0; content: ''; }
.fcard__link:focus-visible { outline: none; }

.fcard__summary {
  display: -webkit-box;
  margin: 0;
  overflow: hidden;
  font-size: 14px;
  line-height: 1.5;
  color: var(--muted);
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

.fcard__foot {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 12px;
  margin: auto 0 0;
  padding-top: 10px;
  border-top: 1px solid var(--line);
  font-size: 12px;
  color: var(--muted);
}

.fcard__author { min-width: 0; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; max-width: 18em; }

@media (prefers-reduced-motion: reduce) {
  .fcard { transition: none; }
}
</style>
