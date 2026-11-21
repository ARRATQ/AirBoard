-- Index uniques que GORM ne sait pas exprimer, auparavant posés à chaque démarrage
-- par main.go :
--   * noms et slugs réutilisables après suppression douce : index uniques PARTIELS
--     (WHERE deleted_at IS NULL) à la place des anciens index uniques complets ;
--   * un vote / un avis par utilisateur (index uniques composites).
-- Les modèles n'ont volontairement pas de tag uniqueIndex sur ces colonnes : avec
-- GORM 1.25, un uniqueIndex sur une seule colonne ajoute une contrainte UNIQUE
-- complète qui bloquerait la recréation d'un élément supprimé.
--
-- Idempotent : sur une base existante, les index partiels déjà posés sont gardés.
-- Si des doublons empêchent un index unique, il n'est pas créé (avertissement
-- Postgres) au lieu de bloquer le démarrage, comme le faisait main.go.

-- +goose Up

-- Anciens index uniques complets, sous les noms par défaut de GORM. Selon la
-- version de GORM qui les a créés, ce sont des contraintes ou de simples index.
ALTER TABLE public.news_categories DROP CONSTRAINT IF EXISTS idx_news_categories_name;
DROP INDEX IF EXISTS public.idx_news_categories_name;
ALTER TABLE public.news_categories DROP CONSTRAINT IF EXISTS idx_news_categories_slug;
DROP INDEX IF EXISTS public.idx_news_categories_slug;
ALTER TABLE public.news_types DROP CONSTRAINT IF EXISTS idx_news_types_name;
DROP INDEX IF EXISTS public.idx_news_types_name;
ALTER TABLE public.news_types DROP CONSTRAINT IF EXISTS idx_news_types_slug;
DROP INDEX IF EXISTS public.idx_news_types_slug;
ALTER TABLE public.suggestion_categories DROP CONSTRAINT IF EXISTS idx_suggestion_categories_name;
DROP INDEX IF EXISTS public.idx_suggestion_categories_name;
ALTER TABLE public.suggestion_categories DROP CONSTRAINT IF EXISTS idx_suggestion_categories_slug;
DROP INDEX IF EXISTS public.idx_suggestion_categories_slug;
ALTER TABLE public.events DROP CONSTRAINT IF EXISTS idx_events_slug;
DROP INDEX IF EXISTS public.idx_events_slug;
ALTER TABLE public.event_categories DROP CONSTRAINT IF EXISTS idx_event_categories_name;
DROP INDEX IF EXISTS public.idx_event_categories_name;
ALTER TABLE public.event_categories DROP CONSTRAINT IF EXISTS idx_event_categories_slug;
DROP INDEX IF EXISTS public.idx_event_categories_slug;
ALTER TABLE public.tags DROP CONSTRAINT IF EXISTS idx_tags_name;
DROP INDEX IF EXISTS public.idx_tags_name;
ALTER TABLE public.tags DROP CONSTRAINT IF EXISTS idx_tags_slug;
DROP INDEX IF EXISTS public.idx_tags_slug;

-- +goose StatementBegin
-- Pose l'index unique partiel idx sur tbl(col). Une version complète (non partielle)
-- portant déjà ce nom, contrainte ou index, est d'abord retirée.
CREATE FUNCTION pg_temp.ensure_partial_unique(idx text, tbl text, col text) RETURNS void AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint
               WHERE conname = idx AND conrelid = format('public.%I', tbl)::regclass) THEN
        EXECUTE format('ALTER TABLE public.%I DROP CONSTRAINT %I', tbl, idx);
    END IF;
    IF EXISTS (SELECT 1 FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
               WHERE c.relname = idx AND c.relnamespace = 'public'::regnamespace
                 AND i.indpred IS NULL) THEN
        EXECUTE format('DROP INDEX public.%I', idx);
    END IF;
    EXECUTE format('CREATE UNIQUE INDEX IF NOT EXISTS %I ON public.%I (%I) WHERE deleted_at IS NULL',
                   idx, tbl, col);
EXCEPTION WHEN unique_violation THEN
    RAISE WARNING 'index unique % non créé : doublons existants dans %.%', idx, tbl, col;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION pg_temp.try_unique_index(ddl text) RETURNS void AS $$
BEGIN
    EXECUTE ddl;
EXCEPTION WHEN unique_violation THEN
    RAISE WARNING 'index unique non créé (doublons existants) : %', ddl;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

SELECT pg_temp.ensure_partial_unique('idx_news_slug', 'news', 'slug');
SELECT pg_temp.ensure_partial_unique('idx_news_category_name', 'news_categories', 'name');
SELECT pg_temp.ensure_partial_unique('idx_category_slug', 'news_categories', 'slug');
SELECT pg_temp.ensure_partial_unique('idx_news_type_name', 'news_types', 'name');
SELECT pg_temp.ensure_partial_unique('idx_news_type_slug', 'news_types', 'slug');
SELECT pg_temp.ensure_partial_unique('idx_tag_name', 'tags', 'name');
SELECT pg_temp.ensure_partial_unique('idx_tag_slug', 'tags', 'slug');
SELECT pg_temp.ensure_partial_unique('idx_event_slug', 'events', 'slug');
SELECT pg_temp.ensure_partial_unique('idx_event_category_name', 'event_categories', 'name');
SELECT pg_temp.ensure_partial_unique('idx_event_category_slug', 'event_categories', 'slug');
SELECT pg_temp.ensure_partial_unique('idx_suggestion_category_name', 'suggestion_categories', 'name');
SELECT pg_temp.ensure_partial_unique('idx_suggestion_category_slug', 'suggestion_categories', 'slug');

-- Un avis par utilisateur et par élément, un vote par option, un vote par suggestion.
SELECT pg_temp.try_unique_index('CREATE UNIQUE INDEX IF NOT EXISTS idx_feedback_user_entity ON public.feedbacks (user_id, entity_type, entity_id)');
SELECT pg_temp.try_unique_index('CREATE UNIQUE INDEX IF NOT EXISTS idx_poll_vote_user_option ON public.poll_votes (poll_id, user_id, poll_option_id)');
SELECT pg_temp.try_unique_index('CREATE UNIQUE INDEX IF NOT EXISTS idx_suggestion_vote_user_unique ON public.suggestion_votes (suggestion_id, user_id)');

DROP FUNCTION pg_temp.ensure_partial_unique(text, text, text);
DROP FUNCTION pg_temp.try_unique_index(text);

-- +goose Down
-- Pas de retour arrière : revenir aux index complets rendrait les slugs supprimés
-- impossibles à réutiliser.
SELECT 1;
