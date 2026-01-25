-- Vérifier tous les sondages du groupe daaf
SELECT
    p.id,
    p.title,
    p.news_id,
    p.author_id,
    p.is_active,
    COALESCE(
        (SELECT string_agg(g.name, ', ')
         FROM poll_target_groups ptg
         JOIN groups g ON g.id = ptg.group_id
         WHERE ptg.poll_id = p.id
        ),
        'GLOBAL'
    ) as target_groups
FROM polls p
WHERE p.deleted_at IS NULL
ORDER BY p.id DESC
LIMIT 20;

-- Vérifier les sondages liés à un groupe spécifique (daaf)
SELECT
    p.id,
    p.title,
    p.news_id,
    p.is_active,
    g.name as group_name
FROM polls p
LEFT JOIN poll_target_groups ptg ON ptg.poll_id = p.id
LEFT JOIN groups g ON g.id = ptg.group_id
WHERE p.deleted_at IS NULL
  AND g.name LIKE '%daaf%'
ORDER BY p.id DESC;
