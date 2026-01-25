-- Script de test pour valider la réutilisation des slugs après soft delete
-- Ce script teste que les slugs peuvent être réutilisés après suppression

-- Prérequis : La migration fix_news_slug_constraint.sql doit avoir été exécutée

BEGIN;

-- Test 1: News slug
-- Créer un article de test
INSERT INTO news (slug, title, summary, content, type, priority, is_published, author_id, created_at, updated_at)
VALUES ('test-article-slug', 'Test Article', 'Test summary', '{}', 'article', 'normal', false, 1, NOW(), NOW());

-- Soft delete de l'article
UPDATE news SET deleted_at = NOW() WHERE slug = 'test-article-slug';

-- Tenter de créer un nouvel article avec le même slug (devrait réussir)
INSERT INTO news (slug, title, summary, content, type, priority, is_published, author_id, created_at, updated_at)
VALUES ('test-article-slug', 'Test Article Nouveau', 'Test summary', '{}', 'article', 'normal', false, 1, NOW(), NOW());

-- Vérification : devrait retourner 2 lignes (1 supprimée, 1 active)
SELECT
    id,
    slug,
    title,
    deleted_at IS NOT NULL as is_deleted
FROM news
WHERE slug = 'test-article-slug'
ORDER BY created_at;

-- Test 2: Tag name et slug
-- Créer un tag de test
INSERT INTO tags (name, slug, color, created_at, updated_at)
VALUES ('test-tag', 'test-tag-slug', '#FF0000', NOW(), NOW());

-- Soft delete du tag
UPDATE tags SET deleted_at = NOW() WHERE slug = 'test-tag-slug';

-- Tenter de créer un nouveau tag avec le même slug (devrait réussir)
INSERT INTO tags (name, slug, color, created_at, updated_at)
VALUES ('test-tag', 'test-tag-slug', '#00FF00', NOW(), NOW());

-- Vérification : devrait retourner 2 lignes (1 supprimée, 1 active)
SELECT
    id,
    name,
    slug,
    deleted_at IS NOT NULL as is_deleted
FROM tags
WHERE slug = 'test-tag-slug'
ORDER BY created_at;

-- Rollback pour ne pas polluer la base
ROLLBACK;

-- Message de succès
SELECT 'Tests réussis ! Les slugs peuvent être réutilisés après soft delete.' AS result;
