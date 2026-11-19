# Plan de refonte Airboard

> Rédigé le 2026-11-19 à partir de trois audits (design/UX, performance front, backend).
> À reprendre **un chantier par session**, dans l'ordre. Chaque chantier est livrable seul.

## Objectif

Une application **rapide**, **agréable à regarder et à parcourir**, et **facile à comprendre**
pour un utilisateur comme pour un administrateur.

## Décisions déjà prises

- **Tous les liens visibles par un utilisateur normal sont importants et utilisés** : aucun module
  utilisateur n'est retiré ni masqué. La simplification porte sur le **côté admin** et sur
  l'organisation des pages.
- Les utilisateurs se connectent via **OAuth Microsoft** (`handlers/oauth.go`). Le SSO par en-têtes
  Authentik est désactivé en prod (`SSO_ENABLED=false`).
- Prod sur Coolify avec `JWT_SECRET` défini.
- Catégories d'actualités, modification/suppression de tags et annuaire (personnes, affectations) :
  **admins globaux seuls** (2026-11-19). Les admins de groupe gardent la **création** de tags.

## Déjà fait

| Commit | Contenu |
|---|---|
| `8cc4c99` | Admin plus jamais réinitialisé au démarrage (`ADMIN_PASSWORD`, `ADMIN_RESET_PASSWORD`) ; secrets JWT publics refusés ; SSO Authentik protégé par `SSO_PROXY_SECRET` ; tri ORDER BY en liste blanche (`utils.SafeOrderClause`) |
| `ef33e9f` | OAuth : `state` obligatoire et lié au navigateur (anti login-CSRF) |
| (chantier 1) | Refresh token refusé comme access token ; rôle/compte actif relus en base (cache 1 min) ; limite 20/min/IP sur login/register + verrouillage IP+compte et compte (compteur réparé) ; démo derrière `SEED_DEMO_DATA` ; avatar vérifié par contenu ; corps limités (5 Mo / 12 Mo, nginx 12 Mo) ; origine WebSocket ; mot de passe 8 car. ; recherche avec apostrophes ; `TRUSTED_PROXIES` |

| `330875b` | Chantier 2, front : transition de page sans `out-in` (fondu 120 ms, navigation ~700 ms → ~200 ms) ; profil, palette et réglages admin en parallèle au démarrage ; `getAppSettings` partagé (3 → 1 appel), version une fois par session, mises à jour vérifiées au plus toutes les 4 h, intervalle nettoyé |
| `627be0c` | Chantier 2, front : ChatOverlay, ChatbotWidget (+ `@n8n/chat`) et HelpDrawer (+ `marked`) chargés au premier clic ; `getActiveChatbots` appelé une fois. Le WebSocket du chat reste ouvert au démarrage (au moment d'inactivité) car il alimente le badge des messages non lus et le son |
| `c50e23f` | Chantier 2, front : compteur de notifications toutes les 60 s, en pause quand l'onglet est caché, relu au retour |
| `2ccbd96` | Chantier 2, build : chunk `ui` vide supprimé, `console.log`/`debug` retirés en prod ; proxy de dev configurable (`API_PROXY_TARGET`) |
| `a5f47d3` | Chantier 2, back : pool DB (25/10, 30 min) ; 16 index (clics, commentaires, clés étrangères, côté inverse des tables de jointure dont `user_groups(user_id)`) ; unicité `news_reads` et `news_reactions` (index simple + avertissement si doublons en prod) — package `backend/database` |
| `9a29669` | Chantier 2, back : vue et XP comptées une seule fois en concurrence (`ON CONFLICT DO NOTHING`, point 7) |
| `4a73cfb` | Chantier 2, back : N+1 supprimés sur `/home` (108 → 55 requêtes SQL), `/dashboard`, rapports par groupe et par utilisateur (nombre de requêtes constant). Comptage par `go test -tags integration -run TestQueryCounts ./handlers/` |
| `7b4e0f5` | Chantier 2 : gzip des réponses API dans nginx aussi derrière un proxy (`gzip_proxied any`) |

Mesures du chantier 2 (local) : JS+CSS au premier chargement 240 → 205 kB compressés ; appels API au
chargement de `/home` 9 → 8 (utilisateur) et 11 → 9 (admin), puis 7 / 8 quand la vérification de mise à
jour a moins de 4 h.

À vérifier après déploiement : une connexion Microsoft réelle, le chat (WebSocket : `FRONTEND_URL` doit être
l'URL publique https), un envoi d'avatar de plus de 1 Mo, et le log de désactivation du compte démo `user`.

---

## Chantier 1 — Sécurité restante ✅ fait

1. **Refresh token accepté comme access token** : `verifyToken` (`middleware/auth.go:151-196`) ne vérifie
   pas le champ `type`. → Rejeter `type=refresh`.
2. **Rôle lu dans le JWT, pas en base** : un utilisateur rétrogradé ou désactivé garde ses droits jusqu'à
   expiration (pas de contrôle `is_active` dans `RequireAuth`). → Charger rôle + `is_active` depuis la
   base, avec un cache d'environ 1 minute.
3. **Aucune limite de tentatives de connexion** : `middleware/rate_limit.go` n'est jamais monté, et il
   paniquerait (`userID.(string)` alors que c'est un `uint`, ligne 26). Le verrouillage est en mémoire,
   par `IP:username` (`handlers/auth.go:66`). → Corriger et monter `AuthRateLimit()` sur `/auth/*`,
   ajouter un verrouillage par nom d'utilisateur, ne faire confiance qu'au proxy nginx (`main.go:246-251`).
4. **Compte démo `user/user123` créé en prod** (`main.go`, seed). → Le conditionner à `SEED_DEMO_DATA=true`.
5. **Permissions admin de groupe trop larges** : catégories/tags d'actualités (`news_categories.go:108,222`)
   et personnes/affectations de l'annuaire (`directory_people.go:120,232`) sont modifiables globalement
   par tout admin de groupe.
6. Plus petits points :
   - l'upload d'avatar fait confiance au `Content-Type` envoyé par le client (`auth.go:838`) ;
   - pas de limite de taille de corps de requête ;
   - le WebSocket accepte toutes les origines (`chat/client.go:38`) ;
   - longueur minimale du mot de passe à 6 caractères au changement contre 8 à l'inscription
     (`models.go:251`) ;
   - la liste noire de la recherche bloque les apostrophes (« l'équipe » échoue, `search.go:96`).

## Chantier 2 — Performance (gains rapides) — en partie fait

Fait : front 1, 4, 5, 6 et le chunk `ui` / `console.log` du point 8 ; back 1, 3, 4, 7 et 8 (cache du
middleware, chantier 1). Gzip (back 2) traité dans nginx, par où passe toute l'API en prod.
Reste : front 2 (icônes locales), 3 (langues à la demande), 7 (polices, `loading="lazy"`), le reste du 8 ;
back 5 (`NOT EXISTS` : seules les requêtes de `home.go` touchées ici sont converties), 6 (pagination),
et les rapports par rôle et de détail (~60 requêtes, nombre constant, non N+1).

**Frontend** (premier chargement : environ 235 kB de JS compressé avant la moindre route)
1. Retirer la transition de page `out-in` de 0,3 s (`App.vue:31-41`), qui ajoute environ 600 ms à chaque
   navigation. Garder au plus un fondu de 120 ms.
2. Embarquer les icônes en local. Environ 436 icônes `mdi:` sont aujourd'hui téléchargées une à une
   depuis api.iconify.design. → `@iconify-json/mdi` + `addCollection` (sous-ensemble), ou unplugin-icons.
3. Charger les langues à la demande (`i18n/index.js:3-6` les charge toutes, environ 300 kB bruts).
4. Charger les widgets flottants au premier clic : ChatOverlay, ChatbotWidget, HelpDrawer (qui embarque
   `marked`) via `defineAsyncComponent`. Ouvrir le WebSocket du chat seulement quand c'est utile. Ne
   récupérer `getActiveChatbots` qu'une fois (appel en double : `FloatingSpeedDial.vue:46` et
   `ChatbotWidget.vue:206`).
5. Supprimer les appels en double ou en série au démarrage :
   - App.vue enchaîne profil → réglages → réglages de l'app en série ; les passer en `Promise.all` ;
   - `getAppSettings` est appelé 3 fois (App, `Sidebar.vue:679`, `Dashboard.vue:582`) ;
   - version et check-updates : une fois par session ;
   - l'intervalle de 4 h n'est jamais nettoyé (`version.js:138`).
6. Notifications : poll toutes les 30 s même onglet caché (`NotificationBell.vue:281`). → Mettre en pause
   sur `document.hidden`, passer à 60 s ou plus, ou pousser le compteur via le WebSocket.
7. Polices :
   - Tajawal est chargée pour tout le monde (à charger seulement pour `ar`) ;
   - réduire Inter aux graisses réellement utilisées ;
   - ajouter `loading="lazy"` aux 17 `<img>`.
8. Divers :
   - retirer les `console.log` en prod avec `esbuild.drop` (`api.js:61-78`, garde du routeur) ;
   - supprimer le chunk `ui` vide ;
   - TiptapRenderer : choisir les langages lowlight à la main (environ 30 kB compressés gagnés) ;
   - mettre la météo en cache (open-meteo/nominatim à chaque montage de Home, `Home.vue:676-721`) ;
   - garder une seule librairie de dates (dayjs n'est utilisé que dans `GamificationHub.vue:261`).

**Backend**
1. Pool de connexions DB (`main.go:31`) : `SetMaxOpenConns(25)`, idle 10, durée de vie 30 min.
2. Compression gzip des réponses.
3. Corriger les requêtes N+1 :
   - `dashboard.go:57,109` (une requête par AppGroup) ;
   - `home.go:502` (2 COUNT par actualité), `home.go:550-600`, `home.go:692` ;
   - `reports.go:509-600` (environ 9 requêtes par utilisateur) et `reports.go:332-470`.
4. Index manquants :
   - `application_clicks(application_id, clicked_at)` et `(user_id)` ;
   - `comments(entity_type, entity_id)` ;
   - unicité de `news_reads(news_id, user_id)` et `news_reactions(news_id, user_id)` ;
   - `*_target_groups(group_id)` ;
   - les clés étrangères `author_id` et `category_id`.
5. Remplacer `(SELECT COUNT(*) …) = 0` par `NOT EXISTS` (30 occurrences).
6. Paginer `GetUsers` (`admin.go:615`), `GetComments` (`comments.go:133`), `GetPendingComments`, et les
   listes de group_admin.
7. Compteur de vues et XP en concurrence (`news.go:753-773`). → Index unique + `ON CONFLICT DO NOTHING`,
   et attribuer l'XP seulement si `RowsAffected == 1`.
8. `RequireAuth` lance une requête `group_admins` à chaque appel (`middleware/auth.go:79`). → Mettre en cache.

## Chantier 3 — Fondations du design

Problème : deux systèmes de couleurs cohabitent. La palette runtime (`--accent`, `main.css:6-23`,
`config/palettes.js`) ne s'applique qu'à une partie de l'interface. Le reste utilise le bleu Tailwind
`primary-*` (343 occurrences), `blue-*` (493), plus du violet et de l'indigo codés en dur, et 545 couleurs
hexadécimales dans 64 fichiers. Des surcharges globales (`main.css:878-893`) servent de rustine.

1. **Un seul jeu de tokens** : faire pointer `tailwind.config.js` (`primary`, surfaces, bordures, texte)
   sur les variables CSS de la palette, puis supprimer les surcharges globales.
2. **Une boîte à outils `components/ui/`** (aujourd'hui il n'y a que IconInput, LoadingOverlay et
   NotificationContainer) : `BaseButton`, `BaseModal` (piège du focus, Échap, `role="dialog"`),
   `ConfirmDialog`, `PageHeader`, `DataTable` (recherche + pagination), `EmptyState`, `FormField`.
3. **Remplacer les 47 `alert()`/`confirm()` natifs** par des toasts et le `ConfirmDialog`. Toasts
   traduits (`stores/app.js:274-299`), avec `aria-live`, et placés pour ne plus chevaucher le bouton
   flottant.
4. Harmoniser les arrondis, les titres de page (`.page-title`), et les ombres.
5. Corriger `admin/SettingsManagement.vue`, stylé uniquement pour le mode sombre (cassé en mode clair).
6. Supprimer le code mort : `components/home/*` (environ 3 800 lignes dont `HeroSection` 991), les
   dépendances `@headlessui/vue` et `@heroicons/vue` jamais importées, et les scripts/SQL de debug à la
   racine.

> Avant de coder : choisir une direction visuelle précise et la valider avec des maquettes.

### Direction A+ (validée le 2026-11-19)

Maquette de référence : [`plans/maquettes/airboard-a-plus.html`](maquettes/airboard-a-plus.html)
(les tokens CSS sont sur `.tok`).

Sobre, inspirée des panneaux de départs d'aéroport, avec retenue : le panneau d'annonces et les tableaux
de liste ont l'allure d'un tableau d'affichage sombre ; les chiffres clés défilent comme des palettes.
Aucun gadget qui invente des données.

**Règles**
- La couleur signale les exceptions. L'état normal est en encre ; seuls l'alerte, le retard,
  l'annulation, le « à confirmer » et le férié sont colorés.
- Un vocabulaire de statuts commun à tous les modules (`StatusBadge`) : petites capitales mono, bordure
  1 px `currentColor`, rayon 3 px. Clignotement réservé à l'urgence, coupé sous `prefers-reduced-motion`.
- Partir de l'existant : aucun champ ni aucune fonction absents du modèle. Les « applications » sont des
  liens vers des ressources (applis en ligne, documents et sites Office 365, liens web), pas des services
  surveillés.
- Aucun module visible par un utilisateur ne disparaît. Clair et sombre ; fr, en, es, ar (RTL : propriétés
  CSS logiques).

**Typographie** : IBM Plex Sans (+ IBM Plex Sans Arabic pour `ar`) ; IBM Plex Mono pour chiffres, dates,
heures et étiquettes en capitales. Remplace Inter et Tajawal (`display=swap`, graisses 400/500/600).
Corps 13,5 px ; titre de page 22 px/600 ; surtitre mono 11 px en capitales, interlettrage 0,1 em.

**Forme** : rayons 3 px (badges, petits boutons), 4 px (boutons, champs, tuiles), 6 px (cartes,
panneaux) ; `rounded-full` seulement pour avatars et pastilles. Pas d'ombre sur les cartes (bordure 1 px),
une seule ombre pour les éléments flottants. Lignes de tableau de 40 px. Lien de menu actif : bordure de
début de 3 px couleur `fill`.

**Tokens** (variables CSS générées par `stores/app.js` pour chaque palette, palette personnalisée
comprise) : `--acc` (couleur de la palette), `--acc-ink` (variante foncée, contraste ≥ 4,5:1 avec le
blanc), `--fill` (= `acc-ink` en clair, `acc` en sombre).

| Token | Clair | Sombre |
|---|---|---|
| bg | color-mix(oklab, acc 3 %, #eceef0) | #0e1012 |
| surface | #fff | #15181b |
| sunken | color-mix(oklab, acc 2 %, #f4f5f6) | #111316 |
| line | color-mix(oklab, acc 6 %, #d6dade) | #272c31 |
| text / muted | #121518 / #596068 | #e6e9ec / #959da5 |
| on-fill | #fff | #141210 |
| ok / warn / bad / info | #15803d / #b45309 / #b91c1c / #1d4ed8 | #4ade80 / #fbbf24 / #f87171 / #93c5fd |
| board / board-2 | #15181b / #1d2125 | #050607 / #0d0f11 |
| board-text / board-dim | #f4f1e8 / #959ba2 | #f4f1e8 / #80868d |
| board-hi / board-ok / board-bad | #ffc857 / #5ee08f / #ff7b72 | idem |

**Lots d'exécution** (un commit chacun) : 0 figer la direction ; 1 tokens, polices et Tailwind ;
2 composants `components/ui/` + Vitest ; 3 `alert()`/`confirm()` → toasts et `ConfirmDialog` ;
4 harmonisation (`PageHeader`, `StatusBadge`) ; 5 `SettingsManagement.vue` en mode clair ; 6 code mort.

**Reporté hors chantier 3** : tour de contrôle admin (chantier 4) ; type de ressource déduit de l'URL,
« Nouveau » pendant 14 jours, note admin datée, calendrier restylé et vue Liste en tableau d'affichage
(chantier 5).

## Chantier 4 — Navigation et organisation

1. **Menu admin** : environ 37 liens pour un admin, parce que tout le bloc « Administration de groupe »
   s'affiche aussi pour lui (`Sidebar.vue:102`), en doublon des entrées admin globales. Le groupe
   « Contenu » mélange 12 éléments. → Regrouper par domaine (Contenu, Communauté, Applications,
   Utilisateurs, Système) et ranger les taxonomies (catégories, types) dans leur page parente
   (onglets).
2. **Créer une page d'accueil admin** : aujourd'hui `/admin` redirige vers les réglages
   (`router/index.js:250`).
3. **Fusionner les doublons admin / admin de groupe** : 7 pages et 2 modales ont divergé (NewsEditor
   473 lignes de différence, EventEditor 458, EventsManagement 385…). → Un seul composant, avec un
   périmètre selon le rôle.
4. **Accueil vs « Ressources » (Dashboard)** : favoris, annonces et statistiques sont en double, et le
   vocabulaire est flou (Applications, Ressources, AppGroups ; « Mes AppGroups »). Choisir un seul terme.
   Unifier la page d'arrivée après connexion (`/home` pour la connexion, `Dashboard` dans la garde du
   routeur).
5. Bugs :
   - lien actif surligné à tort, à cause de `startsWith` (`Sidebar.vue:713-721`) ;
   - icônes sans libellé quand le menu est replié ;
   - double hamburger sur Dashboard (`Dashboard.vue:36-41`).

## Chantier 5 — Refonte page par page

Commencer par `Home.vue` (1 975 lignes, dont environ 1 250 de CSS ; très dense : bandeau, recherche,
météo, XP, 7 KPI, actualités, colonne latérale). Le découper en composants et revoir la hiérarchie
visuelle. Ensuite :
- les pages utilisateur sans mise en page responsive : `EventsCenter`, `PollDetailPage`,
  `DirectoryPage`, `GamificationHub` (qui n'a pas non plus d'état de chargement) ;
- les pages admin, en les migrant sur `PageHeader` + `DataTable` + `ConfirmDialog`.

## Chantier 6 — Langues, arabe et accessibilité

1. Il manque environ 412 clés en `es` et en `ar` (events 166, profile 49, help 48, reports 40…), les
   29 clés `oauth.*` en `fr`, et 25 clés utilisées dans le code n'existent dans aucune langue.
2. Écrans écrits en dur :
   - en anglais : `Login`, `Register`, `NewsManagement`, `NotFound`, `Unauthorized` ;
   - en français : `UserModal`, `GroupModal`, `ApplicationModal`, `AppGroupModal`, `ChatOverlay`,
     `RichTextEditor`, et les KPI de `Home.vue:528-554` ;
   - `meta.title` des routes (`router/index.js:474-640`) ;
   - pluriels construits avec `.slice(0,-1)` (`Dashboard.vue:240`).
3. **Arabe (droite à gauche)** : environ 430 classes de direction physique (`ml-`, `mr-`, `text-left`,
   `left-`…), aucune logique. → Migrer vers `ms-`, `me-`, `ps-`, `start-`, `text-start` ; inverser le
   hamburger, les toasts et le bouton flottant.
4. **Accessibilité** :
   - 63 `div`/`span` cliquables à remplacer par des `button` ou des liens (KPI de `Home.vue:213`…) ;
   - piège du focus et Échap dans toutes les modales (via `BaseModal`) ;
   - le raccourci `⌘K` s'affiche aussi sous Windows et Linux.

## Chantier 7 — Structure du backend

1. Découper `main.go` (1 522 lignes) : migrations, index SQL, routes (`routes.go`) et données initiales
   (`seed.go`).
2. Migrations versionnées (goose ou atlas) au lieu d'AutoMigrate + DDL manuel.
3. Factoriser la logique de visibilité « groupes de l'utilisateur + groupes gérés », copiée 11 fois,
   et les 24 recherches `user_groups`.
4. Un seul format de réponse d'erreur (aujourd'hui 342 `gin.H{"error"}` contre 316 `models.ErrorResponse`).
5. Ajouter `recover()` dans les 13 goroutines lancées sans suivi. Lancer le serveur via `http.Server`
   avec timeouts et arrêt propre.
6. Tests table-driven sur les règles d'autorisation (`CanManage*`) et sur les handlers critiques.
7. Bruit au démarrage : le seed recrée des catégories de suggestion déjà existantes (erreur de clé
   unique `idx_suggestion_categories_slug`, `main.go:1443`).

---

## Comment reprendre

Ouvrir une session neuve et écrire par exemple :
> « Lis `plans/refonte-airboard.md` et attaque le chantier 2 (performance). »

Pour les chantiers 3 à 5, commencer par une discussion de conception avec maquettes avant de coder.
