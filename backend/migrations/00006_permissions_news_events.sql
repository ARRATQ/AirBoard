-- Actualités et événements passent aux permissions (news.manage, events.manage).
-- Les éditeurs les géraient sur tout le site : le rôle « Éditeur » reçoit ces deux
-- permissions avec la portée « tout le site » ; l'admin peut ensuite la restreindre à
-- « ses groupes » depuis la page Rôles. Les comptes passés éditeurs depuis la migration
-- 00005 reçoivent aussi le rôle.

-- +goose Up
INSERT INTO public.access_role_permissions (role_id, permission, scope)
SELECT r.id, p.permission, 'all'
FROM public.access_roles r
CROSS JOIN (VALUES ('news.manage'), ('events.manage')) AS p (permission)
WHERE lower(r.name) = lower('Éditeur')
ON CONFLICT DO NOTHING;

UPDATE public.access_roles
SET description = 'Gère les actualités et les événements, ainsi que les ressources de ses groupes.',
    updated_at = now()
WHERE lower(name) = lower('Éditeur')
  AND description = 'Gère les ressources et leurs catégories dans ses groupes.';

INSERT INTO public.user_access_roles (user_id, access_role_id)
SELECT u.id, r.id
FROM public.users u
JOIN public.access_roles r ON lower(r.name) = lower('Éditeur')
WHERE u.role = 'editor' AND u.deleted_at IS NULL
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM public.access_role_permissions
WHERE permission IN ('news.manage', 'events.manage');

UPDATE public.access_roles
SET description = 'Gère les ressources et leurs catégories dans ses groupes.'
WHERE lower(name) = lower('Éditeur')
  AND description = 'Gère les actualités et les événements, ainsi que les ressources de ses groupes.';
