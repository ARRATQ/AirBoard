# Changelog

Tous les changements importants dans ce projet sont documentés dans ce fichier.

Le format est basé sur [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) et ce projet suit [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.0.0] - 2026-02-20

### 🎉 Major Features

#### **Gamification System**
- User levels and XP progression system
- Achievement badges with milestone tracking
- Global and category-based leaderboards
- XP earning from multiple actions (login, views, reactions, comments, polls)
- Dashboard widget showing personal stats and achievements
- Achievement notifications and history tracking

#### **Suggestion Box System**
- Community-driven feature request and idea submission
- Category-based organization (Feature, Bug, Improvement, Other)
- Voting system with upvote/downvote
- Comment threads on suggestions
- Admin workflow (Pending → Approved → Implemented)
- Status tracking and user notifications
- Export suggestions for analysis

#### **Global Search System**
- Full-text search across all content types (apps, articles, announcements, FAQs)
- Advanced filtering (by type, category, tags, date range)
- Instant results with keyword highlighting
- Permission-aware search (respects user access)
- Multi-language support with typo tolerance
- Search history and popular searches
- Performance optimizations with indexing and caching

### ✨ Enhancements

#### **News Types Management** (Security & Quality)
- New `/admin/news-types` management page
- SQL injection prevention (ID parameter validation)
- Input validation (max length, format checks)
- Automatic slug generation with collision detection
- Database index on `news.type` for performance
- Foreign key support (`NewsTypeID`) for referential integrity
- Better error messages (409 Conflict on duplicate names)
- Slug uniqueness handling with numeric suffixes

#### **News Display Improvements**
- News articles now grouped by type on home page
- Each type has its own branded widget with color and icon
- "View All" link per type for filtering to News Center
- Better UX with visual hierarchy of content types
- Fix: i18n now uses type names from database (no more broken keys for custom types)
- Fix: Removed double navigation on article click

#### **Frontend Store Management**
- News types store now properly refreshed after mutations
- Newly created types appear immediately without page reload
- Improved state synchronization

#### **Database & Model**
- Added `NewsTypeID` foreign key to `News` model
- Backward compatible with existing `Type` string field
- `NewsType` model with soft deletes and indexed slugs
- Proper constraint handling on cascade operations
- `NewsTypeRequest` with strict validation (max 100 chars, min/max ranges)

### 🐛 Bug Fixes

- **SQL Injection Risk**: All ID parameters now validated before database queries
- **Double Navigation**: NewsCard component no longer navigates + emits event
- **Stale Cache**: News types store properly invalidated on mutations
- **Count Query Error**: DeleteNewsType now checks Count() return error
- **Unique Constraint**: Better error handling for duplicate type names
- **Missing Validation**: Input validation on NewsType fields (name, icon, color, order)

### 🔒 Security Improvements

- ID parameter validation in all news_types handlers
- Input sanitization on NewsType creation/update
- CSS injection protection (color field validation)
- SQL injection prevention with parameterized queries
- Proper constraint violation handling

### 📚 Documentation

- Added comprehensive feature descriptions in README.md
- Version bump to 2.0.0 in package.json and README
- This CHANGELOG documenting all changes

### 🔄 Dependencies

No breaking changes to dependencies.

---

## [1.6.1] - 2025-12-15

### ✨ Features
- In-app notifications system
- Notification types and priority levels
- Notification center with filtering
- Dashboard notification widget

### 🐛 Bug Fixes
- Fixed notification timing issues
- Improved notification rendering performance

---

## [1.6.0] - 2025-12-01

### ✨ Features
- Announcement system with dynamic banners
- Support for multiple announcement types
- Scheduled announcements with validity periods
- Admin announcement management

---

## [1.5.0] - 2025-11-15

### ✨ Features
- Polls and surveys system
- Interactive voting interface
- Result visualization and analytics
- Poll integration with news articles

---

## [1.4.0] - 2025-10-01

### ✨ Features
- Group Admin role for delegated management
- Private application groups
- Group Admin endpoints and permissions
- Group management dashboard

### 🐛 Bug Fixes
- Fixed permission checks in group operations

---

## [1.3.0] - 2025-09-01

### ✨ Features
- Advanced news analytics
- Article view tracking per user
- Reaction statistics
- Analytics dashboard

---

## [1.2.0] - 2025-08-15

### ✨ Features
- News Hub with Tiptap rich editor
- Article categories and tags
- Reaction system (👍, ❤️, 🎉)
- Display modes (grid, list, table)

---

## [1.1.0] - 2025-07-01

### ✨ Features
- User management system
- Group management
- Role-based access control
- User favorites

### 🐛 Bug Fixes
- Fixed middleware authentication chain

---

## [1.0.0] - 2025-06-01

### ✨ Initial Release Features
- Dual authentication (classic + SSO)
- Application portal with icons and colors
- User dashboard with favorites
- Group-based access control
- Multi-language support (fr, en, es, ar)
- SSO integration with Authentik
- Docker-ready deployment
- Health checks and monitoring

---

## Format Guide

- `🎉` = Major Feature
- `✨` = Enhancement/Feature
- `🐛` = Bug Fix
- `🔒` = Security
- `📚` = Documentation
- `🔄` = Dependencies
- `⚡` = Performance

For questions or contributions, see [CONTRIBUTING.md](CONTRIBUTING.md) or open an issue on GitHub.
