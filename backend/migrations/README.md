# Migrations SQL

Ce dossier contient les migrations SQL manuelles pour corriger certains problèmes que GORM AutoMigrate ne peut pas résoudre automatiquement.

## Migration : fix_news_slug_constraint.sql

**Problème résolu :**
Lorsqu'un article (news) est supprimé (soft delete), son slug reste en base de données avec un `deleted_at` non NULL. Les anciennes contraintes d'unicité empêchaient la création d'un nouvel article avec le même titre/slug.

**Solution :**
Utilisation d'index uniques partiels (partial unique indexes) qui ignorent les enregistrements supprimés (`WHERE deleted_at IS NULL`).

**Tables concernées :**
- `news` (colonne `slug`)
- `news_categories` (colonne `slug`)
- `tags` (colonnes `slug` et `name`)
- `events` (colonne `slug`)
- `event_categories` (colonnes `name` et `slug`)

**Exécution :**
Cette migration s'exécute automatiquement au démarrage du backend (voir `backend/main.go`).

Pour l'exécuter manuellement :
```bash
# Depuis le conteneur postgres
docker exec -i airboard-postgres-1 psql -U airboard -d airboard < backend/migrations/fix_news_slug_constraint.sql

# Ou depuis psql
psql -U airboard -d airboard -f backend/migrations/fix_news_slug_constraint.sql
```

**Test de validation :**
1. Créer un article avec le titre "Test Article" (slug généré : `test-article`)
2. Supprimer cet article (soft delete)
3. Créer un nouvel article avec le même titre "Test Article"
4. ✅ Le nouvel article devrait être créé avec succès avec le slug `test-article`

**Note technique :**
PostgreSQL supporte les index uniques partiels depuis la version 7.2. Ils permettent de créer une contrainte d'unicité qui ne s'applique que sur un sous-ensemble des lignes (ici, uniquement les lignes non supprimées).
