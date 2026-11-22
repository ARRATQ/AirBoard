-- Sondages et modération passent aux permissions (polls.manage, comments.moderate), et
-- l'ancien rôle de compte « editor » disparaît : un compte est « admin » ou « user », ses
-- autres droits viennent des rôles créés par l'administrateur.
-- Les éditeurs créaient des sondages et modéraient les commentaires sur tout le site : le
-- rôle « Éditeur » reçoit ces deux permissions sur tout le site. Chaque compte « editor »
-- a (ou reçoit ici) ce rôle avant de redevenir « user » : il ne perd rien.

-- +goose Up
INSERT INTO public.access_role_permissions (role_id, permission, scope)
SELECT r.id, p.permission, 'all'
FROM public.access_roles r
CROSS JOIN (VALUES ('polls.manage'), ('comments.moderate')) AS p (permission)
WHERE lower(r.name) = lower('Éditeur')
ON CONFLICT DO NOTHING;

UPDATE public.access_roles
SET description = 'Gère les actualités, les événements, les sondages et la modération, ainsi que les ressources de ses groupes.',
    updated_at = now()
WHERE lower(name) = lower('Éditeur')
  AND description = 'Gère les actualités et les événements, ainsi que les ressources de ses groupes.';

INSERT INTO public.user_access_roles (user_id, access_role_id)
SELECT u.id, r.id
FROM public.users u
JOIN public.access_roles r ON lower(r.name) = lower('Éditeur')
WHERE u.role = 'editor'
ON CONFLICT DO NOTHING;

UPDATE public.users u
SET role = 'user'
WHERE u.role = 'editor'
  AND EXISTS (
      SELECT 1
      FROM public.user_access_roles ur
      JOIN public.access_roles r ON r.id = ur.access_role_id
      WHERE ur.user_id = u.id AND lower(r.name) = lower('Éditeur')
  );

-- +goose Down
-- Approximation : les membres du rôle « Éditeur » redeviennent des comptes « editor ».
UPDATE public.users u
SET role = 'editor'
WHERE u.role = 'user'
  AND EXISTS (
      SELECT 1
      FROM public.user_access_roles ur
      JOIN public.access_roles r ON r.id = ur.access_role_id
      WHERE ur.user_id = u.id AND lower(r.name) = lower('Éditeur')
  );

DELETE FROM public.access_role_permissions
WHERE permission IN ('polls.manage', 'comments.moderate');

UPDATE public.access_roles
SET description = 'Gère les actualités et les événements, ainsi que les ressources de ses groupes.'
WHERE lower(name) = lower('Éditeur')
  AND description = 'Gère les actualités, les événements, les sondages et la modération, ainsi que les ressources de ses groupes.';
