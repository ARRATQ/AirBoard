-- Groupes choisis à l'attribution d'un rôle. Pour un rôle attribué à un compte, les
-- permissions de portée « ses groupes » couvrent les groupes choisis ici ; sans aucun groupe
-- choisi, elles couvrent les groupes dont le compte est membre (hors groupe par défaut),
-- comme avant cette migration : les attributions existantes ne changent pas.

-- +goose Up
CREATE TABLE IF NOT EXISTS public.user_access_role_groups (
    user_id bigint NOT NULL,
    access_role_id bigint NOT NULL,
    group_id bigint NOT NULL REFERENCES public.groups (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, access_role_id, group_id),
    FOREIGN KEY (user_id, access_role_id)
        REFERENCES public.user_access_roles (user_id, access_role_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_access_role_groups_group ON public.user_access_role_groups (group_id);

-- +goose Down
DROP TABLE IF EXISTS public.user_access_role_groups;
