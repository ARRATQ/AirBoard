# Migrations du schéma

Le schéma Postgres est décrit par des fichiers SQL versionnés dans `backend/migrations/`,
embarqués dans le binaire (`//go:embed`) et appliqués au démarrage par
[goose](https://github.com/pressly/goose) (`database.Migrate`, appelé par `main`).
GORM ne crée plus rien : `AutoMigrate` n'est plus appelé.

- goose note les versions appliquées dans la table `goose_db_version`.
- Chaque fichier s'exécute dans une transaction : une migration qui échoue est
  annulée en entier et le serveur refuse de démarrer (`Erreur lors des migrations`).
- Un verrou consultatif Postgres empêche deux instances de migrer en même temps.

| Fichier | Contenu |
|---|---|
| `00001_baseline.sql` | Le schéma qu'AutoMigrate construisait à partir des modèles. |
| `00002_soft_delete_unique_indexes.sql` | Index uniques partiels (`WHERE deleted_at IS NULL`) sur noms et slugs, suppression des anciens index complets, unicité des votes et avis. |
| `00003_performance_indexes.sql` | Index de performance (clés étrangères, tables de jointure, lectures/réactions). |
| `00004_icones_lucide.sql` | Icônes en base passées de mdi à Lucide (table de correspondance, logos de marque inchangés), nouvelles valeurs par défaut. Anciennes valeurs gardées dans `icon_lucide_backup` pour `Down`. |
| `00005_roles_permissions.sql` | Rôles personnalisés : `access_roles`, `access_role_permissions` (permission du catalogue `models/access.go`, portée `all` ou `groups`), `user_access_roles`. Rôle « Éditeur » (`resources.manage` sur ses groupes) attribué aux comptes `role = 'editor'`. |
| `00006_permissions_news_events.sql` | Actualités et événements par permission : le rôle « Éditeur » reçoit `news.manage` et `events.manage` sur tout le site (ce que les éditeurs avaient déjà) ; rôle réattribué aux comptes `role = 'editor'`. |
| `00007_permissions_polls_comments.sql` | Sondages et modération par permission : le rôle « Éditeur » reçoit `polls.manage` et `comments.moderate` sur tout le site. Fin du rôle de compte `editor` : ses comptes (dotés du rôle « Éditeur ») redeviennent `user`. |
| `00008_retrait_admins_groupe.sql` | Fin des admins de groupe : chaque compte de `group_admins` reçoit le rôle « Éditeur de groupe » (ressources, actualités, événements et sondages, portée « ses groupes ») et devient membre des groupes qu'il administrait ; la table `group_admins` est supprimée (le retour arrière la recrée vide). |

## Ajouter une migration

1. Modifier le modèle dans `backend/models/` si besoin.
2. Créer `backend/migrations/NNNNN_description.sql`, numéro suivant le dernier, sans trou :

   ```sql
   -- Pourquoi ce changement.

   -- +goose Up
   ALTER TABLE public.news ADD COLUMN IF NOT EXISTS reading_time bigint DEFAULT 0;

   -- +goose Down
   ALTER TABLE public.news DROP COLUMN IF EXISTS reading_time;
   ```

   - Écrire le SQL que GORM aurait produit : `bigint` pour `int`/`uint`, `text` pour
     `string`, `timestamp with time zone` pour `time.Time`, `boolean`, et la valeur
     du tag `default:`.
   - Toujours `IF NOT EXISTS` / `IF EXISTS` : le test `migrations_test.go` le vérifie.
   - Fonctions ou blocs `DO $$ … $$` : les entourer de `-- +goose StatementBegin` et
     `-- +goose StatementEnd`.
   - **Pas de `uniqueIndex` GORM sur une seule colonne** d'un modèle à suppression
     douce : GORM 1.25 ajoute une contrainte UNIQUE complète qui bloque la recréation
     d'un élément supprimé. Poser un index unique partiel en SQL, comme dans 00002.
   - Ne jamais modifier une migration déjà déployée : en écrire une nouvelle.
3. Vérifier sur une base jetable (jamais la base de dev partagée) :

   ```bash
   createdb -T template0 airboard_migration_test
   AIRBOARD_MIGRATION_TEST_DSN="host=localhost user=airboard password=… dbname=airboard_migration_test sslmode=disable" \
     go test -tags integration -run 'TestMigrate|TestSeed' -v . ./database/
   dropdb airboard_migration_test
   ```

   `TestMigrateIsIdempotentAndCoversModels` migre deux fois puis vérifie que chaque
   modèle GORM a sa table et toutes ses colonnes : un champ ajouté sans migration le
   fait échouer. Un nouveau modèle doit aussi être ajouté à sa liste `persistedModels`.

## Premier démarrage sur une base existante (production)

La base de production a été construite par l'ancien démarrage (AutoMigrate + DDL
manuel) et n'a pas de table `goose_db_version`. Au premier démarrage de cette version,
goose applique les trois migrations. Elles sont idempotentes :

- tables, séquences, colonnes et index ne sont créés que s'ils manquent
  (`CREATE … IF NOT EXISTS`, `ADD COLUMN IF NOT EXISTS`) : une base déjà à jour n'est
  pas modifiée, une base d'une version plus ancienne reçoit les colonnes et tables
  manquantes, comme AutoMigrate le faisait ;
- une contrainte n'est ajoutée que si aucune contrainte de même nom ou de même
  définition n'existe (un index unique complet sur les mêmes colonnes compte) ;
- les anciens index uniques complets sur noms et slugs sont retirés, remplacés par
  les index partiels ;
- les données ne sont jamais modifiées. Si des doublons empêchent un index unique,
  l'index n'est pas créé (index simple `_nonunique` pour les lectures et réactions) ;
  si des lignes orphelines empêchent une clé étrangère, elle est posée `NOT VALID`.
  Ces cas sont signalés au démarrage par des lignes `Avertissement:` ;
- les tables d'anciennes versions qui ne correspondent plus à aucun modèle sont
  laissées en place.

Ce que la reprise ne corrige pas : une colonne dont le type a changé dans le modèle
(AutoMigrate pouvait parfois la convertir), une colonne `NOT NULL` sans valeur par
défaut ajoutée à une table déjà remplie (échec, comme avec AutoMigrate).

Avant la première mise en production de cette version, faire une sauvegarde
(`pg_dump`) de la base.
