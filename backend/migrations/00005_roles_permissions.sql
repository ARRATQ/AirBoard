-- Rôles personnalisés et permissions : l'administrateur crée des rôles, leur confie des
-- permissions du catalogue (définies dans le code, models/access.go), chacune pour tout le
-- site ou pour les seuls groupes dont la personne est membre, puis les attribue.
-- Premier module géré ainsi : les ressources. Les éditeurs existants reçoivent le rôle
-- « Éditeur », qui gère les ressources de leurs groupes.

-- +goose Up
CREATE TABLE IF NOT EXISTS public.access_roles (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    description text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_access_roles_name ON public.access_roles (lower(name));

CREATE TABLE IF NOT EXISTS public.access_role_permissions (
    role_id bigint NOT NULL REFERENCES public.access_roles (id) ON DELETE CASCADE,
    permission text NOT NULL,
    scope text NOT NULL DEFAULT 'groups' CHECK (scope IN ('all', 'groups')),
    PRIMARY KEY (role_id, permission)
);

CREATE TABLE IF NOT EXISTS public.user_access_roles (
    user_id bigint NOT NULL REFERENCES public.users (id) ON DELETE CASCADE,
    access_role_id bigint NOT NULL REFERENCES public.access_roles (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, access_role_id)
);
CREATE INDEX IF NOT EXISTS idx_user_access_roles_role ON public.user_access_roles (access_role_id);

INSERT INTO public.access_roles (name, description, created_at, updated_at)
SELECT 'Éditeur', 'Gère les ressources et leurs catégories dans ses groupes.', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.access_roles WHERE lower(name) = lower('Éditeur'));

INSERT INTO public.access_role_permissions (role_id, permission, scope)
SELECT id, 'resources.manage', 'groups' FROM public.access_roles WHERE lower(name) = lower('Éditeur')
ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_roles (user_id, access_role_id)
SELECT u.id, r.id
FROM public.users u
JOIN public.access_roles r ON lower(r.name) = lower('Éditeur')
WHERE u.role = 'editor' AND u.deleted_at IS NULL
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS public.user_access_roles;
DROP TABLE IF EXISTS public.access_role_permissions;
DROP TABLE IF EXISTS public.access_roles;
