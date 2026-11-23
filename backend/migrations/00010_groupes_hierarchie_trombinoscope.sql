-- Groupes hiérarchiques et groupes liés au trombinoscope.
--   groups.parent_id : groupe parent. Un membre d'un sous-groupe voit ce qui cible les groupes
--     au-dessus ; un rôle limité à un groupe couvre aussi ses sous-groupes.
--   org_units.group_id : groupe d'utilisateurs lié à une entité du trombinoscope (au plus une
--     entité par groupe). Les personnes affectées à un poste de l'entité, liées à un compte, en
--     sont membres.
--   user_groups.source : 'manual' (ajout à la main, SSO, groupe par défaut) ou 'directory'
--     (ajout de la synchronisation du trombinoscope, qui ne retire que ceux-là).

-- +goose Up
ALTER TABLE public.groups
    ADD COLUMN IF NOT EXISTS parent_id bigint REFERENCES public.groups (id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_groups_parent_id ON public.groups (parent_id);

ALTER TABLE public.org_units
    ADD COLUMN IF NOT EXISTS group_id bigint REFERENCES public.groups (id) ON DELETE SET NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_org_units_group_id
    ON public.org_units (group_id) WHERE group_id IS NOT NULL AND deleted_at IS NULL;

ALTER TABLE public.user_groups
    ADD COLUMN IF NOT EXISTS source text NOT NULL DEFAULT 'manual';
CREATE INDEX IF NOT EXISTS idx_user_groups_source ON public.user_groups (source) WHERE source <> 'manual';

-- Appartenance effective : groupes du compte et tous les groupes au-dessus (groupes supprimés
-- exclus). UNION (et non UNION ALL) arrête la récursion même si une boucle existait.
CREATE OR REPLACE VIEW public.user_effective_groups AS
WITH RECURSIVE up (user_id, group_id) AS (
    SELECT ug.user_id, ug.group_id
    FROM public.user_groups ug
    JOIN public.groups g ON g.id = ug.group_id AND g.deleted_at IS NULL
    UNION
    SELECT up.user_id, p.id
    FROM up
    JOIN public.groups g ON g.id = up.group_id
    JOIN public.groups p ON p.id = g.parent_id AND p.deleted_at IS NULL
)
SELECT user_id, group_id FROM up;

-- +goose Down
DROP VIEW IF EXISTS public.user_effective_groups;
DROP INDEX IF EXISTS public.idx_user_groups_source;
ALTER TABLE public.user_groups DROP COLUMN IF EXISTS source;
DROP INDEX IF EXISTS public.idx_org_units_group_id;
ALTER TABLE public.org_units DROP COLUMN IF EXISTS group_id;
DROP INDEX IF EXISTS public.idx_groups_parent_id;
ALTER TABLE public.groups DROP COLUMN IF EXISTS parent_id;
