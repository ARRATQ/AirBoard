-- Migration pour corriger les contraintes d'unicité sur les slugs
-- Permet la réutilisation des slugs après soft delete

-- 1. News slug
DROP INDEX IF EXISTS idx_news_slug;
CREATE UNIQUE INDEX IF NOT EXISTS idx_news_slug ON news(slug) WHERE deleted_at IS NULL;

-- 2. NewsCategory slug
DROP INDEX IF EXISTS idx_category_slug;
CREATE UNIQUE INDEX IF NOT EXISTS idx_category_slug ON news_categories(slug) WHERE deleted_at IS NULL;

-- 3. Tag slug
DROP INDEX IF EXISTS idx_tag_slug;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_slug ON tags(slug) WHERE deleted_at IS NULL;

-- 4. Tag name
DROP INDEX IF EXISTS idx_tag_name;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_name ON tags(name) WHERE deleted_at IS NULL;

-- 5. Event slug
DROP INDEX IF EXISTS idx_event_slug;
CREATE UNIQUE INDEX IF NOT EXISTS idx_event_slug ON events(slug) WHERE deleted_at IS NULL;

-- 6. EventCategory name
DROP INDEX IF EXISTS idx_event_category_name;
CREATE UNIQUE INDEX IF NOT EXISTS idx_event_category_name ON event_categories(name) WHERE deleted_at IS NULL;

-- 7. EventCategory slug
DROP INDEX IF EXISTS idx_event_category_slug;
CREATE UNIQUE INDEX IF NOT EXISTS idx_event_category_slug ON event_categories(slug) WHERE deleted_at IS NULL;

-- Vérification des index créés
SELECT
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE indexname IN (
    'idx_news_slug',
    'idx_category_slug',
    'idx_tag_slug',
    'idx_tag_name',
    'idx_event_slug',
    'idx_event_category_name',
    'idx_event_category_slug'
)
ORDER BY tablename, indexname;
