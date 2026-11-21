-- Index de performance absents des modèles, auparavant vérifiés à chaque démarrage
-- par database.EnsurePerformanceIndexes : clés étrangères et colonnes filtrées par
-- les pages les plus fréquentées (accueil, ressources, actualités, rapports,
-- commentaires, visibilité par groupe).

-- +goose Up

CREATE INDEX IF NOT EXISTS idx_application_clicks_app_clicked ON public.application_clicks (application_id, clicked_at);
CREATE INDEX IF NOT EXISTS idx_application_clicks_user ON public.application_clicks (user_id);
CREATE INDEX IF NOT EXISTS idx_applications_app_group ON public.applications (app_group_id);
CREATE INDEX IF NOT EXISTS idx_comments_entity ON public.comments (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_comments_user ON public.comments (user_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent ON public.comments (parent_id);
CREATE INDEX IF NOT EXISTS idx_news_author ON public.news (author_id);
CREATE INDEX IF NOT EXISTS idx_news_category ON public.news (category_id);
CREATE INDEX IF NOT EXISTS idx_events_category ON public.events (category_id);
CREATE INDEX IF NOT EXISTS idx_polls_author ON public.polls (author_id);

-- Tables de jointure : la clé primaire commence par l'autre colonne, la recherche
-- « par groupe » / « par utilisateur » n'était pas indexée.
CREATE INDEX IF NOT EXISTS idx_user_groups_user ON public.user_groups (user_id);
CREATE INDEX IF NOT EXISTS idx_group_admins_group ON public.group_admins (group_id);
CREATE INDEX IF NOT EXISTS idx_group_app_groups_group ON public.group_app_groups (group_id);
CREATE INDEX IF NOT EXISTS idx_news_target_groups_group ON public.news_target_groups (group_id);
CREATE INDEX IF NOT EXISTS idx_event_target_groups_group ON public.event_target_groups (group_id);
CREATE INDEX IF NOT EXISTS idx_poll_target_groups_group ON public.poll_target_groups (group_id);

-- Une lecture et une réaction par utilisateur et par article. Des doublons existants
-- empêchent l'unicité : aucune donnée n'est supprimée, un index simple suffixé
-- _nonunique sert alors les lectures (comportement repris de l'ancien démarrage).
-- +goose StatementBegin
DO $$
DECLARE
    spec text[];
BEGIN
    FOREACH spec SLICE 1 IN ARRAY ARRAY[
        ['idx_news_reads_news_user', 'news_reads'],
        ['idx_news_reactions_news_user', 'news_reactions']
    ] LOOP
        IF EXISTS (SELECT 1 FROM pg_class WHERE relname = spec[1] AND relnamespace = 'public'::regnamespace) THEN
            CONTINUE;
        END IF;
        BEGIN
            EXECUTE format('CREATE UNIQUE INDEX %I ON public.%I (news_id, user_id)', spec[1], spec[2]);
        EXCEPTION WHEN unique_violation THEN
            RAISE WARNING 'index unique % impossible (doublons) : index simple à la place', spec[1];
            EXECUTE format('CREATE INDEX IF NOT EXISTS %I ON public.%I (news_id, user_id)',
                           spec[1] || '_nonunique', spec[2]);
        END;
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS public.idx_application_clicks_app_clicked;
DROP INDEX IF EXISTS public.idx_application_clicks_user;
DROP INDEX IF EXISTS public.idx_applications_app_group;
DROP INDEX IF EXISTS public.idx_comments_entity;
DROP INDEX IF EXISTS public.idx_comments_user;
DROP INDEX IF EXISTS public.idx_comments_parent;
DROP INDEX IF EXISTS public.idx_news_author;
DROP INDEX IF EXISTS public.idx_news_category;
DROP INDEX IF EXISTS public.idx_events_category;
DROP INDEX IF EXISTS public.idx_polls_author;
DROP INDEX IF EXISTS public.idx_user_groups_user;
DROP INDEX IF EXISTS public.idx_group_admins_group;
DROP INDEX IF EXISTS public.idx_group_app_groups_group;
DROP INDEX IF EXISTS public.idx_news_target_groups_group;
DROP INDEX IF EXISTS public.idx_event_target_groups_group;
DROP INDEX IF EXISTS public.idx_poll_target_groups_group;
DROP INDEX IF EXISTS public.idx_news_reads_news_user;
DROP INDEX IF EXISTS public.idx_news_reads_news_user_nonunique;
DROP INDEX IF EXISTS public.idx_news_reactions_news_user;
DROP INDEX IF EXISTS public.idx_news_reactions_news_user_nonunique;
