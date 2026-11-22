-- Fin des admins de groupe (table group_admins) : leurs droits passent aux permissions.
-- Chaque admin de groupe reçoit le rôle « Éditeur de groupe » (ressources, actualités,
-- événements et sondages, limités à ses groupes) et devient membre des groupes qu'il
-- administrait, pour garder la main dessus. Puis la table disparaît.
-- Le rôle n'est créé que s'il y a des admins de groupe à convertir.

-- +goose Up
INSERT INTO public.access_roles (name, description, created_at, updated_at)
SELECT 'Éditeur de groupe',
       'Gère les ressources, les actualités, les événements et les sondages de ses groupes.',
       now(), now()
WHERE EXISTS (SELECT 1 FROM public.group_admins)
  AND NOT EXISTS (SELECT 1 FROM public.access_roles WHERE lower(name) = lower('Éditeur de groupe'));

INSERT INTO public.access_role_permissions (role_id, permission, scope)
SELECT r.id, p.permission, 'groups'
FROM public.access_roles r
CROSS JOIN (VALUES ('resources.manage'), ('news.manage'), ('events.manage'), ('polls.manage')) AS p (permission)
WHERE lower(r.name) = lower('Éditeur de groupe')
  AND EXISTS (SELECT 1 FROM public.group_admins)
ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_roles (user_id, access_role_id)
SELECT DISTINCT ga.user_id, r.id
FROM public.group_admins ga
JOIN public.access_roles r ON lower(r.name) = lower('Éditeur de groupe')
ON CONFLICT DO NOTHING;

INSERT INTO public.user_groups (user_id, group_id)
SELECT DISTINCT ga.user_id, ga.group_id
FROM public.group_admins ga
WHERE NOT EXISTS (
    SELECT 1 FROM public.user_groups ug WHERE ug.user_id = ga.user_id AND ug.group_id = ga.group_id
)
ON CONFLICT DO NOTHING;

DROP TABLE IF EXISTS public.group_admins;

-- +goose Down
-- La table revient vide : les droits restent portés par le rôle « Éditeur de groupe ».
CREATE TABLE IF NOT EXISTS public.group_admins (
    user_id bigint NOT NULL,
    group_id bigint NOT NULL,
    CONSTRAINT group_admins_pkey PRIMARY KEY (user_id, group_id),
    CONSTRAINT fk_group_admins_group FOREIGN KEY (group_id) REFERENCES public.groups(id),
    CONSTRAINT fk_group_admins_user FOREIGN KEY (user_id) REFERENCES public.users(id)
);
CREATE INDEX IF NOT EXISTS idx_group_admins_group ON public.group_admins (group_id);
