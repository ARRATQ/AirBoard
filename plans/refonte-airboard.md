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
  **admins globaux seuls** (2026-11-19). Créer un tag reste ouvert à qui gère les actualités ou les
  événements.

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
| `d71444d` | Chantier 2, front : icônes Iconify embarquées (plugin Vite `vite-plugins/iconsPlugin.js`) : chaque module reçoit les icônes qu'il cite ; celles stockées en base (applications, catégories) viennent de 640 petits paquets locaux chargés à la demande (mdi, lucide, heroicons) ; api.iconify.design retiré de la CSP |
| `286a734` | Chantier 2, front : langues chargées à la demande (anglais de secours en arrière-plan) ; IBM Plex Sans Arabic chargée seulement en `ar` |
| `982979b` | Chantier 2, front : `loading="lazy"` sur les images sous la ligne de flottaison et celles des articles ; lowlight retiré de TiptapRenderer, où il ne colorait rien (`generateHTML` ne lance pas les plugins ProseMirror) : 36 → 1,2 kB compressés |
| `076bc49` | Chantier 2, front : météo de l'accueil en cache 30 min (`composables/useWeather.js`), ville gardée par zone d'environ 1 km |
| `192507f` | Chantier 2, front : date-fns seule, dayjs retiré |
| `956194f` | Chantier 2, back : `NOT EXISTS` au lieu de `(SELECT COUNT(*) …) = 0` (26 occurrences restantes : events, home, news, polls, search), vérifié par `TestVisibilityQueries` |
| `7077ab0` | Chantier 2 : `GET /admin/users` paginé côté SQL (page, `page_size` ≤ 500, recherche, rôle, statut, tri en liste blanche) ; page Utilisateurs et sélecteur d'admins de groupe adaptés ; `PaginationBar` partagé |
| `2ccc83f` | Chantier 2 : commentaires et file de modération paginés (20 par page, ≤ 100), clé `comments` conservée ; section commentaires et page de modération avec `PaginationBar` |
| `2457a4c` | Chantier 2 : listes app-groups et applications de l'admin de groupe paginées (100 par défaut, ≤ 1000) ; le front demande 1000 lignes en attendant une vraie pagination |
| `41d7491` | Correctif : en arabe, la barre de pagination retournait toute la page (`:global([dir='rtl']) .x` compilé en `[dir='rtl']`), défaut déjà présent dans DataTable sur main |
| (suivi) | Correctifs : les sélecteurs de groupes d'applications de l'admin voient tous les groupes (et non 10) ; les liens des articles s'ouvrent toujours dans un nouvel onglet (`DOMPurify.addHook`) |
| (suivi) | Après fusion du chantier 4 : graisse 700 chargée ; `/group-admin/managed-groups` renvoie `user_count` au lieu des utilisateurs (adresses e-mail comprises) ; icône `mdi:poll-check` remplacée ; écrans du chantier 4 adaptés aux listes paginées |
| `728934c` | Chantier 4 : lien actif exact (`utils/navigation.js`), libellés et `aria-label` du menu replié, un seul hamburger |
| `ff5e747` | Chantier 4 : une page par module pour l'admin et l'admin de groupe (actualités, événements, sondages, ressources, catégories) ; rôles et périmètre dans `config/access.js` et `useManagementScope` ; anciennes URL `/group-admin/…` redirigées |
| `c0ba933` | Chantier 4 : menu admin décrit dans `config/navigation.js`, rangé par domaine ; taxonomies en onglets (`SectionTabs`) ; annuaire sous `/admin/directory` |
| `ade5e18` | Chantier 4 : page « Vue d'ensemble » à `/admin` (chiffres clés, à traiter, raccourcis), aussi pour l'admin de groupe ; `/group-admin` redirige |
| `4eebfc2` | Chantier 4 : « Ressources » et « Catégories de ressources » partout ; `/dashboard` → `/resources` ; page d'arrivée unique `/home` ; page Ressources traduite, sans statistiques admin |

Mesures du chantier 2 (local) : JS+CSS au premier chargement 240 → 205 kB compressés ; appels API au
chargement de `/home` 9 → 8 (utilisateur) et 11 → 9 (admin), puis 7 / 8 quand la vérification de mise à
jour a moins de 4 h.

Mesures de la seconde partie (build de prod servi en local, admin, `/home` en français, gzip niveau 9) :

| | main (`c4918a7`) | branche |
|---|---|---|
| JS+CSS du point d'entrée (avant toute route) | 222,2 kB | 161,0 kB |
| JS+CSS bloquants pour afficher l'accueil (entrée + route + langue) | 232,7 kB | 193,7 kB |
| Chargés ensuite en arrière-plan | — | anglais de secours 16,6 kB ; table et 7 paquets d'icônes 38,9 kB |
| Requêtes au chargement de `/home` | 32 | 29 (30 en `ar`) |
| dont api.iconify.design | 7 | 0 |
| dont polices (Google Fonts) | 10 | 4 (5 en `ar`) |
| dont API | 8 | 8 |

Le total chargé par l'accueil est un peu plus élevé (249 kB contre 233 kB plus les réponses JSON
d'Iconify), mais le chemin bloquant est plus court de 39 kB et aucun domaine tiers n'est plus sollicité
pour les icônes. Les paquets d'icônes ont un nom haché : le navigateur les garde en cache.

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

## Chantier 2 — Performance (gains rapides) — fait

Fait : front 1 à 8 ; back 1 à 8. Gzip (back 2) traité dans nginx, par où passe toute l'API en prod.
Polices (front 7) : la police arabe n'est plus chargée que pour `ar` ; la graisse 700 d'IBM Plex est
chargée (le code utilisait le gras, dessiné jusque-là avec la 600 ; IBM Plex s'arrête à 700, donc 800
et `font-black` s'affichent en 700).

Reste, ou à suivre :
- écrans de ressources et de catégories (admin et admin de groupe) : ils demandent une page de 1000
  lignes ; leur donner une vraie pagination si les volumes l'exigent ;
- rapports par rôle et de détail (~60 requêtes, nombre constant, non N+1) ;
- ~~une icône d'une autre collection que mdi, lucide ou heroicons saisie en base ne s'affiche plus~~
  → fait : elle s'affiche avec un pictogramme neutre servi localement (`utils/iconPlaceholder.js`).
  Dans la base de dev : `et:newspaper` (type d'article) et `fluent-emoji-flat:…` (chatbot).

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

## Chantier 3 — Fondations du design ✅ fait (direction A+)

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
| chrome (menu latéral seul) | #f3f4f6 | #0b0c0e |
| bg (page de contenu) | #fff | #141619 |
| surface (cartes, barre du haut) | #fff | #141619 |
| sunken (en-têtes de tableau, champs) | #f5f6f7 | #101215 |
| line | #e2e4e8 | #25292e |
| text / muted | #121518 / #596068 | #e6e9ec / #959da5 |
| on-fill | #fff | #141210 |
| ok / warn / bad / info | #15803d / #b45309 / #b91c1c / #1d4ed8 | #4ade80 / #fbbf24 / #f87171 / #93c5fd |
| board / board-2 (annonces, chiffres clés seulement) | #1b1e22 / #24282d | #1d2125 / #252a2f |
| board-text / board-dim | #f4f1e8 / #959ba2 | #f4f1e8 / #80868d |
| board-hi / board-ok / board-bad | #ffc857 / #5ee08f / #ff7b72 | idem |

Fonds « menu gris, contenu blanc » (choix du 2026-11-20) : aucune teinte d'accent dans les fonds,
séparation par bordures. Tableaux et résultat des jours ouvrés en clair (surface, line, en-tête sunken,
chiffre principal en acc-ink) ; le noir board ne reste que sur le panneau d'annonces et le bandeau
des chiffres clés.

**Lots d'exécution** (un commit chacun) : 0 figer la direction ; 1 tokens, polices et Tailwind ;
2 composants `components/ui/` + Vitest ; 3 `alert()`/`confirm()` → toasts et `ConfirmDialog` ;
4 harmonisation (`PageHeader`, `StatusBadge`) ; 5 `SettingsManagement.vue` en mode clair ; 6 code mort.

**Reporté hors chantier 3** : tour de contrôle admin (chantier 4) ; type de ressource déduit de l'URL,
« Nouveau » pendant 14 jours, note admin datée, calendrier restylé et vue Liste en tableau d'affichage
(chantier 5).

## Chantier 4 — Navigation et organisation ✅ fait (branche `claude/loving-nash-195cdb`)

Décisions : menu admin Vue d'ensemble / Contenu / Communauté / Ressources / Utilisateurs / Système ;
l'admin de groupe voit une liste courte (Vue d'ensemble, Actualités, Événements, Sondages, Ressources),
l'éditeur Actualités et Événements ; terme retenu « Ressources » + « Catégories de ressources ».

Corrigé ensuite côté backend (même PR) : `CreateAppGroup` garde le groupe propriétaire choisi ;
`false` et `0` (actif, privé, nouvel onglet, ordre) sont enregistrés à la création et à la modification ;
filtres publié/brouillon, récurrent et journée entière de la liste des événements ; recherche d'un
caractère acceptée (longueur comptée en caractères) ; liste de termes interdits retirée (l'apostrophe,
« update » ou « create » sont acceptés), seuls les jokers de LIKE sont échappés.

Constat de départ :

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

## Chantier 5 — Refonte page par page ✅ fait (branche `claude/chantier-5-pages-0bceae`)

Fait : accueil découpé en composants (`components/home/`) ; portail Ressources A+ (type déduit de
l'URL, « Nouveau » 14 jours, note de l'admin datée) ; agenda Mois/Semaine/Jour et liste en tableau,
jours ouvrés en direct ; sondage, trombinoscope (organigramme et fiches) et My XP (chargement, erreur,
Réessayer) ; les pages d'admin sur `PageHeader` + `DataTable` + `RowActions` + `useConfirm`. Échap
sur toutes les modales, toasts traduits, titres d'onglet. Côté serveur : `false` respecté à la
création, modération appliquée, noms et slugs réutilisables après suppression (index uniques partiels
posés en SQL : GORM 1.25 ajoutait une contrainte complète), filtres et statistiques des médias côté
serveur. `DataTable` a un mode `remote` pour les listes paginées par le serveur (utilisateurs).

Constat de départ : commencer par `Home.vue` (1 975 lignes, dont environ 1 250 de CSS ; très dense : bandeau, recherche,
météo, XP, 7 KPI, actualités, colonne latérale). Le découper en composants et revoir la hiérarchie
visuelle. Ensuite :
- les pages utilisateur sans mise en page responsive : `EventsCenter`, `PollDetailPage`,
  `DirectoryPage`, `GamificationHub` (qui n'a pas non plus d'état de chargement) ;
- les pages admin, en les migrant sur `PageHeader` + `DataTable` + `ConfirmDialog`.

## Chantier 6 — Langues, arabe et accessibilité ✅ fait (2026-11-21)

**Fait** (branche `claude/chantier-6-i18n-rtl-a11y`) :
- **Langues** : fr/en/es/ar ont les mêmes 2 574 clés (avant : 2 225 en fr/en, 2 228 en es/ar). Environ
  480 textes en dur traduits : Login, Register, OAuthCallback, 404/403 (nouveau gabarit `ErrorPage`),
  ChatOverlay, éditeur riche, paramètres, e-mail/OAuth, analyses, rapports, médias, cartes, bouton
  flottant… Les « (s) » sont remplacés par des pluriels vue-i18n. Test `locales/__tests__/locales.test.js` :
  parité des clés, pas de valeur vide, pas de « (s) », toute clé littérale `t('…')` de `src/` existe en fr.
- **Arabe** : 318 classes Tailwind physiques passées en logiques (0 restante hors centrage `left-1/2`),
  propriétés CSS logiques dans les panneaux de chat, l'éditeur et les encadrés ; surcharges `[dir=rtl]`
  devenues inutiles supprimées ; `space-x`/`divide-x` inversés par une règle globale ; chevrons et flèches
  en `rtl-flip` (BaseButton le fait seul). Bouton flottant, chat et toasts côté fin de ligne.
- **Accessibilité** : directive globale `v-focus-trap` (piège, retour du focus, pile, cède à Headless UI)
  posée avec `role="dialog"`, `aria-modal`, `aria-labelledby` sur une quarantaine de modales maison ;
  26 éléments cliquables non interactifs devenus boutons ou liens (6 fonds de modale restent des `div`
  `aria-hidden`) ; contour `:focus-visible` commun en `var(--acc)` ; `.btn-reset`.
- Vérifié en navigateur (ar 375/1440, fr 1440, clair/sombre) : aucun débordement horizontal, aucune
  erreur de page, Tab reste dans la modale, Échap la ferme et rend le focus.

**Reste / à surveiller** :
- ~~Pluriels arabes limités à 3 formes~~ → fait : règle CLDR à 6 formes (`i18n/arabicPlural.js`), 36
  messages réécrits avec le duel et l'accord des nombres, test de cohérence.
- ~~Aide « 6 caractères » de l'inscription~~ → fait : `utils/passwordPolicy.js` reproduit la règle du
  backend (8 caractères, majuscule, minuscule, chiffre, caractère spécial, mots courants refusés) ;
  le changement de mot de passe demande aussi 8 caractères.
- ~~`readingTime.js` renvoie « N min »~~ → fait (`readingMinutes` + `news.minutes`). Les poignées de
  redimensionnement d'image restent physiques (la position est enregistrée dans l'article).
- ~~`RichTextEditor.insertFile` concatène du HTML~~ → fait : contenu structuré.
- Liens « Conditions d'utilisation » et « Confidentialité » vers `#` → remplacés par du texte simple ;
  à relier le jour où ces pages existent.

État initial (2026-11-19) :

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

## Chantier 7 — Structure du backend ✅ fait (branche `claude/chantier-7-backend`)

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

### Fait — partie A (branche `claude/chantier-7a-structure`)

- **Point 1** : `main.go` découpé (déplacement pur) en `db.go`, `routes.go`, `seed.go`,
  `seed_defaults.go`, `seed_demo.go`, `server.go` ; `main()` tient sur un écran.
- **Point 2** : migrations goose embarquées (`backend/migrations/`, mode d'emploi dans
  `backend/MIGRATIONS.md`), AutoMigrate et DDL manuel retirés du démarrage. `00001_baseline`
  est extrait du schéma que l'ancien démarrage construisait et rendu idempotent pour reprendre
  une base de production existante ; `00002` index uniques partiels, `00003` index de
  performance (ex-`EnsurePerformanceIndexes`). Vérifié : ancien code sur base vide = nouveau
  code sur base vide (schémas identiques), base construite par `main` puis migrée = identique,
  copie de la base de dev migrée = seule `goose_db_version` ajoutée. Test d'intégration qui
  échoue si un champ de modèle n'a pas de migration.
- **Point 5 (moitié serveur)** : `http.Server` avec délais (en-têtes 10 s, lecture et écriture
  2 min, inactivité 2 min ; WebSocket du chat non coupé, test dédié) et arrêt propre sur
  SIGINT/SIGTERM (8 s, puis fermeture de la base).
- **Point 7** : seed idempotent (`ON CONFLICT DO NOTHING`), plus de « record not found » au
  démarrage ; test d'intégration en deux passages.

### Fait — partie B (branche `claude/chantier-7b-handlers`)

- **Point 3 fait.** `handlers/visibility.go` regroupe la visibilité (`newsTargets`, `eventTargets`,
  `pollTargets` : `visibleTo`, `publishedVisibleTo`, `ownOrPublishedVisibleTo`, `manageableBy`) et
  les recherches de groupes (`userGroupIDs`, `adminGroupIDs`, `attachManagedGroups`,
  `canSeeTargeted`…). La page d'accueil calcule les groupes du lecteur une seule fois (55 → 44
  requêtes). Deux failles corrigées au passage : le `OR` de visibilité n'était pas parenthésé
  (brouillons, sondages inactifs et contenus supprimés visibles des groupes ciblés) et
  `GetEventBySlug` ne vérifiait pas le ciblage.
- **Point 4 fait.** Corps d'erreur unique `{"error": "...", "code": "..."}` via le paquet
  `apierror` (705 appels, statuts inchangés). Les erreurs de binding donnent un message lisible et
  les erreurs SQL brutes ne sortent plus. Côté front, `services/apiError.js` lit `error` puis
  `message`.
- **Point 5, partie recover faite.** `utils.SafeGo` couvre les 17 goroutines sans suivi et les 13 de
  l'accueil. Le `CustomRecovery` de Gin répond aussi au format commun.
- **Point 6 fait.** Tests table-driven des gardes de rôle, `CanManage*`, `canManageDirectory`, SQL
  de visibilité (sans base), et tests d'intégration (tag `integration`) sur le ciblage, le périmètre
  admin de groupe et le verrouillage de connexion.

## Chantier 8 — Rôles et permissions ✅ fait (PR #17 à #22)

- Un rôle `admin` intouchable, plus des rôles créés par l'admin à partir d'un catalogue fixe
  (`models/access.go`) : ressources, actualités, événements, sondages, modération. Chaque
  permission vaut sur tout le site ou sur des groupes ; plusieurs rôles s'additionnent.
- Fin du rôle de compte `editor` (#19) et des admins de groupe (#20, migration `00008` vers un
  rôle « Éditeur de groupe »).
- Groupes choisis à l'attribution d'un rôle (#21, table `user_access_role_groups`) ; sans choix,
  les groupes de la personne hors groupe par défaut. Sélecteurs à saisie semi-automatique
  (`ui/ComboSelect`) au lieu des listes à cocher.
- Nettoyage (#22) : textes, commentaires et README alignés sur ce modèle ; une actualité, un
  événement ou un sondage écrit avec une gestion limitée à ses groupes cible au moins un de ses
  groupes (serveur et formulaires). Aide intégrée masquée (`HELP_ENABLED` dans `Sidebar.vue`)
  tant que ses pages ne sont pas refaites.
- Les chemins d'API `/editor` (permission sur tout le site) et `/group-admin` (limitée à ses
  groupes) gardent leurs noms historiques.

## Après le chantier 8 — Groupes (PR #23 et #24, 2026-11-23)

- **Sous-groupes** (#23, migration `00010`) : `groups.parent_id`, vue `user_effective_groups`
  (un membre d'un sous-groupe voit aussi ce qui cible les groupes au-dessus). Une permission sur un
  groupe couvre ses sous-groupes.
- **Groupes liés au trombinoscope** (#23) : une entité peut avoir son groupe ; ses membres y sont
  ajoutés et retirés automatiquement (`user_groups.source = 'directory'`), y compris quand une
  personne change d'entité.
- **Permission `groups.manage`** (#24, page Admin → Mes groupes) : sur ses groupes, créer,
  renommer et supprimer des sous-groupes avec seulement des membres de ces groupes ; sur tout le
  site, gérer tous les groupes. Le groupe par défaut et les groupes liés au trombinoscope restent
  à l'admin.

## Pistes ouvertes

- **Aide intégrée** : masquée (`HELP_ENABLED = false` dans `Sidebar.vue`). Son contenu date
  d'avant les rôles et la refonte des pages ; à réécrire (fr/en/es/ar) avant de la réactiver.

---

## Comment reprendre

Ouvrir une session neuve et écrire par exemple :
> « Lis `plans/refonte-airboard.md` et attaque le chantier 2 (performance). »

Pour les chantiers 3 à 5, commencer par une discussion de conception avec maquettes avant de coder.
