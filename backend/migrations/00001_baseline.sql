-- Schéma de référence d'Airboard : ce qu'AutoMigrate (GORM 1.25) construisait à partir
-- des modèles. Généré depuis une base vide construite par l'ancien démarrage, puis rendu
-- idempotent pour pouvoir s'appliquer aussi sur une base de production existante :
-- tables, colonnes, contraintes et index ne sont créés que s'ils manquent.
-- Les contraintes sont ajoutées via des fonctions temporaires (pg_temp) qui comparent
-- par nom ET par définition, pour ne pas dupliquer une contrainte nommée autrement par
-- une ancienne version de GORM.

-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION pg_temp.ensure_pk(tbl regclass, cname text, def text) RETURNS void AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = tbl AND contype = 'p') THEN
        RETURN;
    END IF;
    EXECUTE format('ALTER TABLE %s ADD CONSTRAINT %I %s', tbl, cname, def);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION pg_temp.ensure_unique(tbl regclass, cname text, cols text[]) RETURNS void AS $$
DECLARE
    keys int2[];
BEGIN
    SELECT array_agg(a.attnum ORDER BY x.o) INTO keys
    FROM unnest(cols) WITH ORDINALITY x(name, o)
    JOIN pg_attribute a ON a.attrelid = tbl AND a.attname = x.name;
    -- Déjà unique (contrainte ou index unique complet, quel que soit son nom) : rien à faire.
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = tbl AND conname = cname)
       OR EXISTS (SELECT 1 FROM pg_index i WHERE i.indrelid = tbl AND i.indisunique
                  AND i.indpred IS NULL AND i.indkey::int2[] = keys) THEN
        RETURN;
    END IF;
    EXECUTE format('ALTER TABLE %s ADD CONSTRAINT %I UNIQUE (%s)', tbl, cname,
                   (SELECT string_agg(quote_ident(c), ', ') FROM unnest(cols) c));
EXCEPTION WHEN unique_violation THEN
    RAISE WARNING 'contrainte % non créée : doublons existants dans %', cname, tbl;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION pg_temp.ensure_fk(tbl regclass, cname text, def text) RETURNS void AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = tbl AND contype = 'f'
               AND (conname = cname OR pg_get_constraintdef(oid) = def)) THEN
        RETURN;
    END IF;
    BEGIN
        EXECUTE format('ALTER TABLE %s ADD CONSTRAINT %I %s', tbl, cname, def);
    EXCEPTION WHEN foreign_key_violation THEN
        -- Lignes orphelines dans une base existante : la contrainte protège les
        -- nouvelles écritures sans bloquer le démarrage.
        RAISE WARNING 'clé étrangère % ajoutée NOT VALID : lignes orphelines dans %', cname, tbl;
        EXECUTE format('ALTER TABLE %s ADD CONSTRAINT %I %s NOT VALID', tbl, cname, def);
    END;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION pg_temp.try_unique_index(ddl text) RETURNS void AS $$
BEGIN
    EXECUTE ddl;
EXCEPTION WHEN unique_violation THEN
    RAISE WARNING 'index unique non créé (doublons existants) : %', ddl;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Tables et colonnes

CREATE SEQUENCE IF NOT EXISTS public.achievements_id_seq;
CREATE TABLE IF NOT EXISTS public.achievements (
    id bigint DEFAULT nextval('achievements_id_seq'::regclass) NOT NULL,
    code text,
    name text NOT NULL,
    description text,
    icon text DEFAULT 'mdi:medal'::text,
    color text DEFAULT '#FBBF24'::text,
    xp_reward bigint DEFAULT 0,
    category text DEFAULT 'user'::text,
    trigger_reason text DEFAULT ''::text,
    metric text DEFAULT ''::text,
    threshold bigint DEFAULT 0,
    is_active boolean DEFAULT true,
    is_secret boolean DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.achievements
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('achievements_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS code text,
    ADD COLUMN IF NOT EXISTS name text NOT NULL,
    ADD COLUMN IF NOT EXISTS description text,
    ADD COLUMN IF NOT EXISTS icon text DEFAULT 'mdi:medal'::text,
    ADD COLUMN IF NOT EXISTS color text DEFAULT '#FBBF24'::text,
    ADD COLUMN IF NOT EXISTS xp_reward bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS category text DEFAULT 'user'::text,
    ADD COLUMN IF NOT EXISTS trigger_reason text DEFAULT ''::text,
    ADD COLUMN IF NOT EXISTS metric text DEFAULT ''::text,
    ADD COLUMN IF NOT EXISTS threshold bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS is_secret boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.achievements_id_seq OWNED BY public.achievements.id;

CREATE SEQUENCE IF NOT EXISTS public.announcements_id_seq;
CREATE TABLE IF NOT EXISTS public.announcements (
    id bigint DEFAULT nextval('announcements_id_seq'::regclass) NOT NULL,
    title text NOT NULL,
    content text,
    "type" text DEFAULT 'info'::text,
    priority bigint DEFAULT 0,
    is_active boolean DEFAULT true,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.announcements
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('announcements_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS title text NOT NULL,
    ADD COLUMN IF NOT EXISTS content text,
    ADD COLUMN IF NOT EXISTS "type" text DEFAULT 'info'::text,
    ADD COLUMN IF NOT EXISTS priority bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS start_date timestamp with time zone,
    ADD COLUMN IF NOT EXISTS end_date timestamp with time zone,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.announcements_id_seq OWNED BY public.announcements.id;

CREATE SEQUENCE IF NOT EXISTS public.app_groups_id_seq;
CREATE TABLE IF NOT EXISTS public.app_groups (
    id bigint DEFAULT nextval('app_groups_id_seq'::regclass) NOT NULL,
    name text NOT NULL,
    description text,
    color text DEFAULT '#10B981'::text,
    icon text DEFAULT 'mdi:folder'::text,
    "order" bigint DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    owner_group_id bigint,
    is_private boolean DEFAULT false
);
ALTER TABLE public.app_groups
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('app_groups_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS name text NOT NULL,
    ADD COLUMN IF NOT EXISTS description text,
    ADD COLUMN IF NOT EXISTS color text DEFAULT '#10B981'::text,
    ADD COLUMN IF NOT EXISTS icon text DEFAULT 'mdi:folder'::text,
    ADD COLUMN IF NOT EXISTS "order" bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS owner_group_id bigint,
    ADD COLUMN IF NOT EXISTS is_private boolean DEFAULT false;
ALTER SEQUENCE public.app_groups_id_seq OWNED BY public.app_groups.id;

CREATE SEQUENCE IF NOT EXISTS public.app_settings_id_seq;
CREATE TABLE IF NOT EXISTS public.app_settings (
    id bigint DEFAULT nextval('app_settings_id_seq'::regclass) NOT NULL,
    app_name text DEFAULT 'Airboard'::text,
    app_icon text DEFAULT 'mdi:view-dashboard'::text,
    dashboard_title text DEFAULT 'Dashboard'::text,
    welcome_message text DEFAULT 'Welcome to your application portal'::text,
    home_page_message text DEFAULT 'Discover your personalized workspace'::text,
    signup_enabled boolean DEFAULT true,
    default_group_id bigint,
    hero_image_url text DEFAULT ''::text,
    hero_image_url_dark text DEFAULT ''::text,
    hero_image_position text DEFAULT 'center center'::text,
    color_palette text DEFAULT 'claude'::text,
    custom_accent_color text DEFAULT '#d97757'::text,
    custom_palettes text DEFAULT '[]'::text,
    news_per_tab bigint DEFAULT 5,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE public.app_settings
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('app_settings_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS app_name text DEFAULT 'Airboard'::text,
    ADD COLUMN IF NOT EXISTS app_icon text DEFAULT 'mdi:view-dashboard'::text,
    ADD COLUMN IF NOT EXISTS dashboard_title text DEFAULT 'Dashboard'::text,
    ADD COLUMN IF NOT EXISTS welcome_message text DEFAULT 'Welcome to your application portal'::text,
    ADD COLUMN IF NOT EXISTS home_page_message text DEFAULT 'Discover your personalized workspace'::text,
    ADD COLUMN IF NOT EXISTS signup_enabled boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS default_group_id bigint,
    ADD COLUMN IF NOT EXISTS hero_image_url text DEFAULT ''::text,
    ADD COLUMN IF NOT EXISTS hero_image_url_dark text DEFAULT ''::text,
    ADD COLUMN IF NOT EXISTS hero_image_position text DEFAULT 'center center'::text,
    ADD COLUMN IF NOT EXISTS color_palette text DEFAULT 'claude'::text,
    ADD COLUMN IF NOT EXISTS custom_accent_color text DEFAULT '#d97757'::text,
    ADD COLUMN IF NOT EXISTS custom_palettes text DEFAULT '[]'::text,
    ADD COLUMN IF NOT EXISTS news_per_tab bigint DEFAULT 5,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;
ALTER SEQUENCE public.app_settings_id_seq OWNED BY public.app_settings.id;

CREATE SEQUENCE IF NOT EXISTS public.application_clicks_id_seq;
CREATE TABLE IF NOT EXISTS public.application_clicks (
    id bigint DEFAULT nextval('application_clicks_id_seq'::regclass) NOT NULL,
    user_id bigint,
    application_id bigint,
    clicked_at timestamp with time zone
);
ALTER TABLE public.application_clicks
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('application_clicks_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS user_id bigint,
    ADD COLUMN IF NOT EXISTS application_id bigint,
    ADD COLUMN IF NOT EXISTS clicked_at timestamp with time zone;
ALTER SEQUENCE public.application_clicks_id_seq OWNED BY public.application_clicks.id;

CREATE SEQUENCE IF NOT EXISTS public.applications_id_seq;
CREATE TABLE IF NOT EXISTS public.applications (
    id bigint DEFAULT nextval('applications_id_seq'::regclass) NOT NULL,
    name text NOT NULL,
    description text,
    url text NOT NULL,
    icon text DEFAULT 'mdi:application'::text,
    color text DEFAULT '#6366F1'::text,
    "order" bigint DEFAULT 0,
    is_active boolean DEFAULT true,
    open_in_new_tab boolean DEFAULT true,
    admin_note character varying(140),
    admin_note_until timestamp with time zone,
    app_group_id bigint,
    created_by_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.applications
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('applications_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS name text NOT NULL,
    ADD COLUMN IF NOT EXISTS description text,
    ADD COLUMN IF NOT EXISTS url text NOT NULL,
    ADD COLUMN IF NOT EXISTS icon text DEFAULT 'mdi:application'::text,
    ADD COLUMN IF NOT EXISTS color text DEFAULT '#6366F1'::text,
    ADD COLUMN IF NOT EXISTS "order" bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS open_in_new_tab boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS admin_note character varying(140),
    ADD COLUMN IF NOT EXISTS admin_note_until timestamp with time zone,
    ADD COLUMN IF NOT EXISTS app_group_id bigint,
    ADD COLUMN IF NOT EXISTS created_by_id bigint,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.applications_id_seq OWNED BY public.applications.id;

CREATE SEQUENCE IF NOT EXISTS public.chat_messages_id_seq;
CREATE TABLE IF NOT EXISTS public.chat_messages (
    id bigint DEFAULT nextval('chat_messages_id_seq'::regclass) NOT NULL,
    content text NOT NULL,
    "type" text DEFAULT 'text'::text,
    sender_id bigint NOT NULL,
    recipient_id bigint,
    group_id bigint,
    is_read boolean DEFAULT false,
    created_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.chat_messages
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('chat_messages_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS content text NOT NULL,
    ADD COLUMN IF NOT EXISTS "type" text DEFAULT 'text'::text,
    ADD COLUMN IF NOT EXISTS sender_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS recipient_id bigint,
    ADD COLUMN IF NOT EXISTS group_id bigint,
    ADD COLUMN IF NOT EXISTS is_read boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.chat_messages_id_seq OWNED BY public.chat_messages.id;

CREATE SEQUENCE IF NOT EXISTS public.chatbots_id_seq;
CREATE TABLE IF NOT EXISTS public.chatbots (
    id bigint DEFAULT nextval('chatbots_id_seq'::regclass) NOT NULL,
    name text NOT NULL,
    description text,
    webhook_url text NOT NULL,
    icon text DEFAULT 'mdi:robot-outline'::text,
    color text DEFAULT '#4f46e5'::text,
    welcome_title text,
    welcome_subtitle text,
    hide_header boolean DEFAULT false,
    show_avatar_intro boolean DEFAULT false,
    is_active boolean DEFAULT true,
    initial_messages text DEFAULT '[]'::text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.chatbots
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('chatbots_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS name text NOT NULL,
    ADD COLUMN IF NOT EXISTS description text,
    ADD COLUMN IF NOT EXISTS webhook_url text NOT NULL,
    ADD COLUMN IF NOT EXISTS icon text DEFAULT 'mdi:robot-outline'::text,
    ADD COLUMN IF NOT EXISTS color text DEFAULT '#4f46e5'::text,
    ADD COLUMN IF NOT EXISTS welcome_title text,
    ADD COLUMN IF NOT EXISTS welcome_subtitle text,
    ADD COLUMN IF NOT EXISTS hide_header boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS show_avatar_intro boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS initial_messages text DEFAULT '[]'::text,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.chatbots_id_seq OWNED BY public.chatbots.id;

CREATE SEQUENCE IF NOT EXISTS public.comment_settings_id_seq;
CREATE TABLE IF NOT EXISTS public.comment_settings (
    id bigint DEFAULT nextval('comment_settings_id_seq'::regclass) NOT NULL,
    comments_enabled boolean DEFAULT true,
    news_comments_enabled boolean DEFAULT true,
    app_comments_enabled boolean DEFAULT false,
    event_comments_enabled boolean DEFAULT true,
    require_moderation boolean DEFAULT false,
    allow_anonymous boolean DEFAULT false,
    max_comment_length bigint DEFAULT 1000,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE public.comment_settings
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('comment_settings_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS comments_enabled boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS news_comments_enabled boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS app_comments_enabled boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS event_comments_enabled boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS require_moderation boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS allow_anonymous boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS max_comment_length bigint DEFAULT 1000,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;
ALTER SEQUENCE public.comment_settings_id_seq OWNED BY public.comment_settings.id;

CREATE SEQUENCE IF NOT EXISTS public.comments_id_seq;
CREATE TABLE IF NOT EXISTS public.comments (
    id bigint DEFAULT nextval('comments_id_seq'::regclass) NOT NULL,
    content text NOT NULL,
    user_id bigint,
    entity_type text NOT NULL,
    entity_id bigint NOT NULL,
    is_approved boolean DEFAULT true,
    is_flagged boolean DEFAULT false,
    moderated_by bigint,
    moderated_at timestamp with time zone,
    parent_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.comments
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('comments_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS content text NOT NULL,
    ADD COLUMN IF NOT EXISTS user_id bigint,
    ADD COLUMN IF NOT EXISTS entity_type text NOT NULL,
    ADD COLUMN IF NOT EXISTS entity_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS is_approved boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS is_flagged boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS moderated_by bigint,
    ADD COLUMN IF NOT EXISTS moderated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS parent_id bigint,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.comments_id_seq OWNED BY public.comments.id;

CREATE SEQUENCE IF NOT EXISTS public.directory_fields_id_seq;
CREATE TABLE IF NOT EXISTS public.directory_fields (
    id bigint DEFAULT nextval('directory_fields_id_seq'::regclass) NOT NULL,
    "key" text NOT NULL,
    label text NOT NULL,
    "type" text DEFAULT 'text'::text NOT NULL,
    options jsonb DEFAULT '[]'::jsonb,
    required boolean DEFAULT false,
    visibility text DEFAULT 'all'::text NOT NULL,
    show_on_card boolean DEFAULT false,
    "order" bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.directory_fields
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('directory_fields_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS "key" text NOT NULL,
    ADD COLUMN IF NOT EXISTS label text NOT NULL,
    ADD COLUMN IF NOT EXISTS "type" text DEFAULT 'text'::text NOT NULL,
    ADD COLUMN IF NOT EXISTS options jsonb DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS required boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS visibility text DEFAULT 'all'::text NOT NULL,
    ADD COLUMN IF NOT EXISTS show_on_card boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS "order" bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.directory_fields_id_seq OWNED BY public.directory_fields.id;

CREATE SEQUENCE IF NOT EXISTS public.directory_people_id_seq;
CREATE TABLE IF NOT EXISTS public.directory_people (
    id bigint DEFAULT nextval('directory_people_id_seq'::regclass) NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    email text,
    phone text,
    photo_url text,
    user_id bigint,
    is_active boolean DEFAULT true,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.directory_people
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('directory_people_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS first_name text NOT NULL,
    ADD COLUMN IF NOT EXISTS last_name text NOT NULL,
    ADD COLUMN IF NOT EXISTS email text,
    ADD COLUMN IF NOT EXISTS phone text,
    ADD COLUMN IF NOT EXISTS photo_url text,
    ADD COLUMN IF NOT EXISTS user_id bigint,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS custom_fields jsonb DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.directory_people_id_seq OWNED BY public.directory_people.id;

CREATE SEQUENCE IF NOT EXISTS public.email_notification_logs_id_seq;
CREATE TABLE IF NOT EXISTS public.email_notification_logs (
    id bigint DEFAULT nextval('email_notification_logs_id_seq'::regclass) NOT NULL,
    template_type text,
    content_id bigint,
    content_title text,
    recipient_count bigint,
    success_count bigint,
    failure_count bigint,
    status text DEFAULT 'pending'::text,
    error_message text,
    sent_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone
);
ALTER TABLE public.email_notification_logs
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('email_notification_logs_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS template_type text,
    ADD COLUMN IF NOT EXISTS content_id bigint,
    ADD COLUMN IF NOT EXISTS content_title text,
    ADD COLUMN IF NOT EXISTS recipient_count bigint,
    ADD COLUMN IF NOT EXISTS success_count bigint,
    ADD COLUMN IF NOT EXISTS failure_count bigint,
    ADD COLUMN IF NOT EXISTS status text DEFAULT 'pending'::text,
    ADD COLUMN IF NOT EXISTS error_message text,
    ADD COLUMN IF NOT EXISTS sent_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS completed_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone;
ALTER SEQUENCE public.email_notification_logs_id_seq OWNED BY public.email_notification_logs.id;

CREATE SEQUENCE IF NOT EXISTS public.email_o_auth_configs_id_seq;
CREATE TABLE IF NOT EXISTS public.email_o_auth_configs (
    id bigint DEFAULT nextval('email_o_auth_configs_id_seq'::regclass) NOT NULL,
    smtp_config_id bigint,
    provider text DEFAULT 'microsoft'::text,
    tenant_id text,
    client_id text,
    client_secret text,
    scopes text DEFAULT 'https://graph.microsoft.com/.default'::text,
    auth_url text DEFAULT 'https://login.microsoftonline.com/{tenant}/oauth2/v2.0/authorize'::text,
    token_url text DEFAULT 'https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token'::text,
    access_token text,
    refresh_token text,
    token_type text DEFAULT 'Bearer'::text,
    expires_at timestamp with time zone,
    grant_type text DEFAULT 'client_credentials'::text,
    is_enabled boolean DEFAULT false,
    last_token_refresh timestamp with time zone,
    last_refresh_error text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE public.email_o_auth_configs
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('email_o_auth_configs_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS smtp_config_id bigint,
    ADD COLUMN IF NOT EXISTS provider text DEFAULT 'microsoft'::text,
    ADD COLUMN IF NOT EXISTS tenant_id text,
    ADD COLUMN IF NOT EXISTS client_id text,
    ADD COLUMN IF NOT EXISTS client_secret text,
    ADD COLUMN IF NOT EXISTS scopes text DEFAULT 'https://graph.microsoft.com/.default'::text,
    ADD COLUMN IF NOT EXISTS auth_url text DEFAULT 'https://login.microsoftonline.com/{tenant}/oauth2/v2.0/authorize'::text,
    ADD COLUMN IF NOT EXISTS token_url text DEFAULT 'https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token'::text,
    ADD COLUMN IF NOT EXISTS access_token text,
    ADD COLUMN IF NOT EXISTS refresh_token text,
    ADD COLUMN IF NOT EXISTS token_type text DEFAULT 'Bearer'::text,
    ADD COLUMN IF NOT EXISTS expires_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS grant_type text DEFAULT 'client_credentials'::text,
    ADD COLUMN IF NOT EXISTS is_enabled boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS last_token_refresh timestamp with time zone,
    ADD COLUMN IF NOT EXISTS last_refresh_error text,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;
ALTER SEQUENCE public.email_o_auth_configs_id_seq OWNED BY public.email_o_auth_configs.id;

CREATE SEQUENCE IF NOT EXISTS public.email_templates_id_seq;
CREATE TABLE IF NOT EXISTS public.email_templates (
    id bigint DEFAULT nextval('email_templates_id_seq'::regclass) NOT NULL,
    "type" text NOT NULL,
    name text NOT NULL,
    subject text NOT NULL,
    html_body text NOT NULL,
    plain_text_body text,
    is_enabled boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE public.email_templates
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('email_templates_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS "type" text NOT NULL,
    ADD COLUMN IF NOT EXISTS name text NOT NULL,
    ADD COLUMN IF NOT EXISTS subject text NOT NULL,
    ADD COLUMN IF NOT EXISTS html_body text NOT NULL,
    ADD COLUMN IF NOT EXISTS plain_text_body text,
    ADD COLUMN IF NOT EXISTS is_enabled boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;
ALTER SEQUENCE public.email_templates_id_seq OWNED BY public.email_templates.id;

CREATE SEQUENCE IF NOT EXISTS public.event_categories_id_seq;
CREATE TABLE IF NOT EXISTS public.event_categories (
    id bigint DEFAULT nextval('event_categories_id_seq'::regclass) NOT NULL,
    name character varying(100) NOT NULL,
    slug character varying(100) NOT NULL,
    description character varying(500),
    icon character varying(100) DEFAULT 'mdi:calendar'::character varying,
    color character varying(20) DEFAULT '#3B82F6'::character varying,
    "order" bigint DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.event_categories
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('event_categories_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS name character varying(100) NOT NULL,
    ADD COLUMN IF NOT EXISTS slug character varying(100) NOT NULL,
    ADD COLUMN IF NOT EXISTS description character varying(500),
    ADD COLUMN IF NOT EXISTS icon character varying(100) DEFAULT 'mdi:calendar'::character varying,
    ADD COLUMN IF NOT EXISTS color character varying(20) DEFAULT '#3B82F6'::character varying,
    ADD COLUMN IF NOT EXISTS "order" bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.event_categories_id_seq OWNED BY public.event_categories.id;

CREATE TABLE IF NOT EXISTS public.event_tags (
    event_id bigint NOT NULL,
    tag_id bigint NOT NULL
);
ALTER TABLE public.event_tags
    ADD COLUMN IF NOT EXISTS event_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS tag_id bigint NOT NULL;

CREATE TABLE IF NOT EXISTS public.event_target_groups (
    event_id bigint NOT NULL,
    group_id bigint NOT NULL
);
ALTER TABLE public.event_target_groups
    ADD COLUMN IF NOT EXISTS event_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS group_id bigint NOT NULL;

CREATE SEQUENCE IF NOT EXISTS public.events_id_seq;
CREATE TABLE IF NOT EXISTS public.events (
    id bigint DEFAULT nextval('events_id_seq'::regclass) NOT NULL,
    slug character varying(255) NOT NULL,
    title character varying(255) NOT NULL,
    description text,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone,
    is_all_day boolean DEFAULT false,
    timezone character varying(50) DEFAULT 'UTC'::character varying,
    is_recurring boolean DEFAULT false,
    recurrence_rule text,
    recurrence_end timestamp with time zone,
    recurrence_exceptions text,
    location character varying(500),
    external_links text,
    color character varying(20) DEFAULT '#3B82F6'::character varying,
    priority character varying(20) DEFAULT 'normal'::character varying,
    status character varying(20) DEFAULT 'confirmed'::character varying,
    cover_image character varying(500),
    is_published boolean DEFAULT false,
    published_at timestamp with time zone,
    is_holiday boolean DEFAULT false,
    country_code character varying(5),
    author_id bigint NOT NULL,
    category_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.events
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('events_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS slug character varying(255) NOT NULL,
    ADD COLUMN IF NOT EXISTS title character varying(255) NOT NULL,
    ADD COLUMN IF NOT EXISTS description text,
    ADD COLUMN IF NOT EXISTS start_date timestamp with time zone NOT NULL,
    ADD COLUMN IF NOT EXISTS end_date timestamp with time zone,
    ADD COLUMN IF NOT EXISTS is_all_day boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS timezone character varying(50) DEFAULT 'UTC'::character varying,
    ADD COLUMN IF NOT EXISTS is_recurring boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS recurrence_rule text,
    ADD COLUMN IF NOT EXISTS recurrence_end timestamp with time zone,
    ADD COLUMN IF NOT EXISTS recurrence_exceptions text,
    ADD COLUMN IF NOT EXISTS location character varying(500),
    ADD COLUMN IF NOT EXISTS external_links text,
    ADD COLUMN IF NOT EXISTS color character varying(20) DEFAULT '#3B82F6'::character varying,
    ADD COLUMN IF NOT EXISTS priority character varying(20) DEFAULT 'normal'::character varying,
    ADD COLUMN IF NOT EXISTS status character varying(20) DEFAULT 'confirmed'::character varying,
    ADD COLUMN IF NOT EXISTS cover_image character varying(500),
    ADD COLUMN IF NOT EXISTS is_published boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS published_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS is_holiday boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS country_code character varying(5),
    ADD COLUMN IF NOT EXISTS author_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS category_id bigint,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.events_id_seq OWNED BY public.events.id;

CREATE SEQUENCE IF NOT EXISTS public.feedbacks_id_seq;
CREATE TABLE IF NOT EXISTS public.feedbacks (
    id bigint DEFAULT nextval('feedbacks_id_seq'::regclass) NOT NULL,
    user_id bigint,
    entity_type text NOT NULL,
    entity_id bigint NOT NULL,
    feedback_type text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE public.feedbacks
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('feedbacks_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS user_id bigint,
    ADD COLUMN IF NOT EXISTS entity_type text NOT NULL,
    ADD COLUMN IF NOT EXISTS entity_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS feedback_type text NOT NULL,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;
ALTER SEQUENCE public.feedbacks_id_seq OWNED BY public.feedbacks.id;

CREATE SEQUENCE IF NOT EXISTS public.gamification_profiles_id_seq;
CREATE TABLE IF NOT EXISTS public.gamification_profiles (
    id bigint DEFAULT nextval('gamification_profiles_id_seq'::regclass) NOT NULL,
    user_id bigint,
    xp bigint DEFAULT 0,
    level bigint DEFAULT 1,
    last_daily_login timestamp with time zone,
    login_streak bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.gamification_profiles
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('gamification_profiles_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS user_id bigint,
    ADD COLUMN IF NOT EXISTS xp bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS level bigint DEFAULT 1,
    ADD COLUMN IF NOT EXISTS last_daily_login timestamp with time zone,
    ADD COLUMN IF NOT EXISTS login_streak bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.gamification_profiles_id_seq OWNED BY public.gamification_profiles.id;

CREATE SEQUENCE IF NOT EXISTS public.gamification_rules_id_seq;
CREATE TABLE IF NOT EXISTS public.gamification_rules (
    id bigint DEFAULT nextval('gamification_rules_id_seq'::regclass) NOT NULL,
    reason text NOT NULL,
    points bigint NOT NULL,
    enabled boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.gamification_rules
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('gamification_rules_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS reason text NOT NULL,
    ADD COLUMN IF NOT EXISTS points bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS enabled boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.gamification_rules_id_seq OWNED BY public.gamification_rules.id;

CREATE TABLE IF NOT EXISTS public.group_admins (
    user_id bigint NOT NULL,
    group_id bigint NOT NULL
);
ALTER TABLE public.group_admins
    ADD COLUMN IF NOT EXISTS user_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS group_id bigint NOT NULL;

CREATE TABLE IF NOT EXISTS public.group_app_groups (
    app_group_id bigint NOT NULL,
    group_id bigint NOT NULL
);
ALTER TABLE public.group_app_groups
    ADD COLUMN IF NOT EXISTS app_group_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS group_id bigint NOT NULL;

CREATE SEQUENCE IF NOT EXISTS public.groups_id_seq;
CREATE TABLE IF NOT EXISTS public.groups (
    id bigint DEFAULT nextval('groups_id_seq'::regclass) NOT NULL,
    name text NOT NULL,
    description text,
    color text DEFAULT '#3B82F6'::text,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.groups
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('groups_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS name text NOT NULL,
    ADD COLUMN IF NOT EXISTS description text,
    ADD COLUMN IF NOT EXISTS color text DEFAULT '#3B82F6'::text,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.groups_id_seq OWNED BY public.groups.id;

CREATE SEQUENCE IF NOT EXISTS public.hero_messages_id_seq;
CREATE TABLE IF NOT EXISTS public.hero_messages (
    id bigint DEFAULT nextval('hero_messages_id_seq'::regclass) NOT NULL,
    content text NOT NULL,
    author text,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.hero_messages
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('hero_messages_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS content text NOT NULL,
    ADD COLUMN IF NOT EXISTS author text,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.hero_messages_id_seq OWNED BY public.hero_messages.id;

CREATE SEQUENCE IF NOT EXISTS public.media_id_seq;
CREATE TABLE IF NOT EXISTS public.media (
    id bigint DEFAULT nextval('media_id_seq'::regclass) NOT NULL,
    filename text NOT NULL,
    storage_path text NOT NULL,
    url text NOT NULL,
    mime_type text NOT NULL,
    file_size bigint NOT NULL,
    width bigint,
    height bigint,
    thumbnail_url text,
    storage_type text DEFAULT 'local'::text,
    uploaded_by bigint,
    usage_count bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.media
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('media_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS filename text NOT NULL,
    ADD COLUMN IF NOT EXISTS storage_path text NOT NULL,
    ADD COLUMN IF NOT EXISTS url text NOT NULL,
    ADD COLUMN IF NOT EXISTS mime_type text NOT NULL,
    ADD COLUMN IF NOT EXISTS file_size bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS width bigint,
    ADD COLUMN IF NOT EXISTS height bigint,
    ADD COLUMN IF NOT EXISTS thumbnail_url text,
    ADD COLUMN IF NOT EXISTS storage_type text DEFAULT 'local'::text,
    ADD COLUMN IF NOT EXISTS uploaded_by bigint,
    ADD COLUMN IF NOT EXISTS usage_count bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.media_id_seq OWNED BY public.media.id;

CREATE SEQUENCE IF NOT EXISTS public.news_id_seq;
CREATE TABLE IF NOT EXISTS public.news (
    id bigint DEFAULT nextval('news_id_seq'::regclass) NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    summary character varying(300),
    content text,
    cover_image text,
    "type" text DEFAULT 'article'::text,
    priority text DEFAULT 'normal'::text,
    is_pinned boolean DEFAULT false,
    is_published boolean DEFAULT false,
    published_at timestamp with time zone,
    expires_at timestamp with time zone,
    view_count bigint DEFAULT 0,
    reading_time bigint,
    author_id bigint,
    category_id bigint,
    news_type_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.news
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('news_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS slug text NOT NULL,
    ADD COLUMN IF NOT EXISTS title text NOT NULL,
    ADD COLUMN IF NOT EXISTS summary character varying(300),
    ADD COLUMN IF NOT EXISTS content text,
    ADD COLUMN IF NOT EXISTS cover_image text,
    ADD COLUMN IF NOT EXISTS "type" text DEFAULT 'article'::text,
    ADD COLUMN IF NOT EXISTS priority text DEFAULT 'normal'::text,
    ADD COLUMN IF NOT EXISTS is_pinned boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS is_published boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS published_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS expires_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS view_count bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reading_time bigint,
    ADD COLUMN IF NOT EXISTS author_id bigint,
    ADD COLUMN IF NOT EXISTS category_id bigint,
    ADD COLUMN IF NOT EXISTS news_type_id bigint,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.news_id_seq OWNED BY public.news.id;

CREATE SEQUENCE IF NOT EXISTS public.news_categories_id_seq;
CREATE TABLE IF NOT EXISTS public.news_categories (
    id bigint DEFAULT nextval('news_categories_id_seq'::regclass) NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    description text,
    icon text DEFAULT 'mdi:newspaper'::text,
    color text DEFAULT '#3B82F6'::text,
    "order" bigint DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.news_categories
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('news_categories_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS name text NOT NULL,
    ADD COLUMN IF NOT EXISTS slug text NOT NULL,
    ADD COLUMN IF NOT EXISTS description text,
    ADD COLUMN IF NOT EXISTS icon text DEFAULT 'mdi:newspaper'::text,
    ADD COLUMN IF NOT EXISTS color text DEFAULT '#3B82F6'::text,
    ADD COLUMN IF NOT EXISTS "order" bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.news_categories_id_seq OWNED BY public.news_categories.id;

CREATE SEQUENCE IF NOT EXISTS public.news_reactions_id_seq;
CREATE TABLE IF NOT EXISTS public.news_reactions (
    id bigint DEFAULT nextval('news_reactions_id_seq'::regclass) NOT NULL,
    news_id bigint,
    user_id bigint,
    reaction_type text,
    created_at timestamp with time zone
);
ALTER TABLE public.news_reactions
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('news_reactions_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS news_id bigint,
    ADD COLUMN IF NOT EXISTS user_id bigint,
    ADD COLUMN IF NOT EXISTS reaction_type text,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone;
ALTER SEQUENCE public.news_reactions_id_seq OWNED BY public.news_reactions.id;

CREATE SEQUENCE IF NOT EXISTS public.news_reads_id_seq;
CREATE TABLE IF NOT EXISTS public.news_reads (
    id bigint DEFAULT nextval('news_reads_id_seq'::regclass) NOT NULL,
    news_id bigint,
    user_id bigint,
    read_at timestamp with time zone
);
ALTER TABLE public.news_reads
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('news_reads_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS news_id bigint,
    ADD COLUMN IF NOT EXISTS user_id bigint,
    ADD COLUMN IF NOT EXISTS read_at timestamp with time zone;
ALTER SEQUENCE public.news_reads_id_seq OWNED BY public.news_reads.id;

CREATE TABLE IF NOT EXISTS public.news_tags (
    news_id bigint NOT NULL,
    tag_id bigint NOT NULL
);
ALTER TABLE public.news_tags
    ADD COLUMN IF NOT EXISTS news_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS tag_id bigint NOT NULL;

CREATE TABLE IF NOT EXISTS public.news_target_groups (
    news_id bigint NOT NULL,
    group_id bigint NOT NULL
);
ALTER TABLE public.news_target_groups
    ADD COLUMN IF NOT EXISTS news_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS group_id bigint NOT NULL;

CREATE SEQUENCE IF NOT EXISTS public.news_types_id_seq;
CREATE TABLE IF NOT EXISTS public.news_types (
    id bigint DEFAULT nextval('news_types_id_seq'::regclass) NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    icon text DEFAULT 'mdi:newspaper'::text,
    color text DEFAULT '#3B82F6'::text,
    "order" bigint DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.news_types
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('news_types_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS name text NOT NULL,
    ADD COLUMN IF NOT EXISTS slug text NOT NULL,
    ADD COLUMN IF NOT EXISTS icon text DEFAULT 'mdi:newspaper'::text,
    ADD COLUMN IF NOT EXISTS color text DEFAULT '#3B82F6'::text,
    ADD COLUMN IF NOT EXISTS "order" bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.news_types_id_seq OWNED BY public.news_types.id;

CREATE SEQUENCE IF NOT EXISTS public.notifications_id_seq;
CREATE TABLE IF NOT EXISTS public.notifications (
    id bigint DEFAULT nextval('notifications_id_seq'::regclass) NOT NULL,
    user_id bigint NOT NULL,
    "type" text NOT NULL,
    category text,
    title text NOT NULL,
    message text,
    icon text DEFAULT 'mdi:bell'::text,
    color text DEFAULT '#3B82F6'::text,
    is_read boolean DEFAULT false,
    priority bigint DEFAULT 0,
    action_url text,
    metadata jsonb,
    created_at timestamp with time zone,
    read_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.notifications
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('notifications_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS user_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS "type" text NOT NULL,
    ADD COLUMN IF NOT EXISTS category text,
    ADD COLUMN IF NOT EXISTS title text NOT NULL,
    ADD COLUMN IF NOT EXISTS message text,
    ADD COLUMN IF NOT EXISTS icon text DEFAULT 'mdi:bell'::text,
    ADD COLUMN IF NOT EXISTS color text DEFAULT '#3B82F6'::text,
    ADD COLUMN IF NOT EXISTS is_read boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS priority bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS action_url text,
    ADD COLUMN IF NOT EXISTS metadata jsonb,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS read_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;

CREATE SEQUENCE IF NOT EXISTS public.o_auth_providers_id_seq;
CREATE TABLE IF NOT EXISTS public.o_auth_providers (
    id bigint DEFAULT nextval('o_auth_providers_id_seq'::regclass) NOT NULL,
    provider_name text NOT NULL,
    display_name text NOT NULL,
    icon text DEFAULT 'mdi:login'::text,
    is_enabled boolean DEFAULT false,
    client_id text,
    client_secret text,
    redirect_uri text,
    auth_url text,
    token_url text,
    user_info_url text,
    scopes text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE public.o_auth_providers
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('o_auth_providers_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS provider_name text NOT NULL,
    ADD COLUMN IF NOT EXISTS display_name text NOT NULL,
    ADD COLUMN IF NOT EXISTS icon text DEFAULT 'mdi:login'::text,
    ADD COLUMN IF NOT EXISTS is_enabled boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS client_id text,
    ADD COLUMN IF NOT EXISTS client_secret text,
    ADD COLUMN IF NOT EXISTS redirect_uri text,
    ADD COLUMN IF NOT EXISTS auth_url text,
    ADD COLUMN IF NOT EXISTS token_url text,
    ADD COLUMN IF NOT EXISTS user_info_url text,
    ADD COLUMN IF NOT EXISTS scopes text,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;
ALTER SEQUENCE public.o_auth_providers_id_seq OWNED BY public.o_auth_providers.id;

CREATE SEQUENCE IF NOT EXISTS public.org_positions_id_seq;
CREATE TABLE IF NOT EXISTS public.org_positions (
    id bigint DEFAULT nextval('org_positions_id_seq'::regclass) NOT NULL,
    unit_id bigint NOT NULL,
    title text NOT NULL,
    is_head boolean DEFAULT false,
    "order" bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.org_positions
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('org_positions_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS unit_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS title text NOT NULL,
    ADD COLUMN IF NOT EXISTS is_head boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS "order" bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.org_positions_id_seq OWNED BY public.org_positions.id;

CREATE SEQUENCE IF NOT EXISTS public.org_units_id_seq;
CREATE TABLE IF NOT EXISTS public.org_units (
    id bigint DEFAULT nextval('org_units_id_seq'::regclass) NOT NULL,
    name text NOT NULL,
    code text,
    description text,
    color text DEFAULT '#3B82F6'::text,
    parent_id bigint,
    "order" bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.org_units
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('org_units_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS name text NOT NULL,
    ADD COLUMN IF NOT EXISTS code text,
    ADD COLUMN IF NOT EXISTS description text,
    ADD COLUMN IF NOT EXISTS color text DEFAULT '#3B82F6'::text,
    ADD COLUMN IF NOT EXISTS parent_id bigint,
    ADD COLUMN IF NOT EXISTS "order" bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.org_units_id_seq OWNED BY public.org_units.id;

CREATE SEQUENCE IF NOT EXISTS public.poll_options_id_seq;
CREATE TABLE IF NOT EXISTS public.poll_options (
    id bigint DEFAULT nextval('poll_options_id_seq'::regclass) NOT NULL,
    poll_id bigint NOT NULL,
    text text NOT NULL,
    is_other boolean DEFAULT false,
    "order" bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.poll_options
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('poll_options_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS poll_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS text text NOT NULL,
    ADD COLUMN IF NOT EXISTS is_other boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS "order" bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.poll_options_id_seq OWNED BY public.poll_options.id;

CREATE TABLE IF NOT EXISTS public.poll_target_groups (
    poll_id bigint NOT NULL,
    group_id bigint NOT NULL
);
ALTER TABLE public.poll_target_groups
    ADD COLUMN IF NOT EXISTS poll_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS group_id bigint NOT NULL;

CREATE SEQUENCE IF NOT EXISTS public.poll_votes_id_seq;
CREATE TABLE IF NOT EXISTS public.poll_votes (
    id bigint DEFAULT nextval('poll_votes_id_seq'::regclass) NOT NULL,
    poll_id bigint NOT NULL,
    user_id bigint NOT NULL,
    poll_option_id bigint NOT NULL,
    voted_at timestamp with time zone,
    created_at timestamp with time zone
);
ALTER TABLE public.poll_votes
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('poll_votes_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS poll_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS user_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS poll_option_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS voted_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone;
ALTER SEQUENCE public.poll_votes_id_seq OWNED BY public.poll_votes.id;

CREATE SEQUENCE IF NOT EXISTS public.polls_id_seq;
CREATE TABLE IF NOT EXISTS public.polls (
    id bigint DEFAULT nextval('polls_id_seq'::regclass) NOT NULL,
    title text NOT NULL,
    description text,
    poll_type text DEFAULT 'single'::text,
    is_anonymous boolean DEFAULT false,
    is_active boolean DEFAULT true,
    show_results text DEFAULT 'after'::text,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    author_id bigint,
    news_id bigint,
    announcement_id bigint
);
ALTER TABLE public.polls
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('polls_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS title text NOT NULL,
    ADD COLUMN IF NOT EXISTS description text,
    ADD COLUMN IF NOT EXISTS poll_type text DEFAULT 'single'::text,
    ADD COLUMN IF NOT EXISTS is_anonymous boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS show_results text DEFAULT 'after'::text,
    ADD COLUMN IF NOT EXISTS start_date timestamp with time zone,
    ADD COLUMN IF NOT EXISTS end_date timestamp with time zone,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS author_id bigint,
    ADD COLUMN IF NOT EXISTS news_id bigint,
    ADD COLUMN IF NOT EXISTS announcement_id bigint;
ALTER SEQUENCE public.polls_id_seq OWNED BY public.polls.id;

CREATE SEQUENCE IF NOT EXISTS public.position_assignments_id_seq;
CREATE TABLE IF NOT EXISTS public.position_assignments (
    id bigint DEFAULT nextval('position_assignments_id_seq'::regclass) NOT NULL,
    position_id bigint NOT NULL,
    person_id bigint NOT NULL,
    is_interim boolean DEFAULT false,
    start_date timestamp with time zone,
    "order" bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.position_assignments
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('position_assignments_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS position_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS person_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS is_interim boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS start_date timestamp with time zone,
    ADD COLUMN IF NOT EXISTS "order" bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.position_assignments_id_seq OWNED BY public.position_assignments.id;

CREATE SEQUENCE IF NOT EXISTS public.smtp_configs_id_seq;
CREATE TABLE IF NOT EXISTS public.smtp_configs (
    id bigint DEFAULT nextval('smtp_configs_id_seq'::regclass) NOT NULL,
    host text,
    port bigint DEFAULT 587,
    username text,
    password text,
    from_email text NOT NULL,
    from_name text DEFAULT 'Airboard'::text,
    use_tls boolean,
    use_star_ttl_s boolean,
    use_o_auth boolean DEFAULT true,
    is_enabled boolean DEFAULT false,
    last_tested_at timestamp with time zone,
    last_test_success boolean DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE public.smtp_configs
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('smtp_configs_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS host text,
    ADD COLUMN IF NOT EXISTS port bigint DEFAULT 587,
    ADD COLUMN IF NOT EXISTS username text,
    ADD COLUMN IF NOT EXISTS password text,
    ADD COLUMN IF NOT EXISTS from_email text NOT NULL,
    ADD COLUMN IF NOT EXISTS from_name text DEFAULT 'Airboard'::text,
    ADD COLUMN IF NOT EXISTS use_tls boolean,
    ADD COLUMN IF NOT EXISTS use_star_ttl_s boolean,
    ADD COLUMN IF NOT EXISTS use_o_auth boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS is_enabled boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS last_tested_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS last_test_success boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;
ALTER SEQUENCE public.smtp_configs_id_seq OWNED BY public.smtp_configs.id;

CREATE SEQUENCE IF NOT EXISTS public.suggestion_categories_id_seq;
CREATE TABLE IF NOT EXISTS public.suggestion_categories (
    id bigint DEFAULT nextval('suggestion_categories_id_seq'::regclass) NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    "order" bigint DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.suggestion_categories
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('suggestion_categories_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS name text NOT NULL,
    ADD COLUMN IF NOT EXISTS slug text NOT NULL,
    ADD COLUMN IF NOT EXISTS "order" bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.suggestion_categories_id_seq OWNED BY public.suggestion_categories.id;

CREATE SEQUENCE IF NOT EXISTS public.suggestion_votes_id_seq;
CREATE TABLE IF NOT EXISTS public.suggestion_votes (
    id bigint DEFAULT nextval('suggestion_votes_id_seq'::regclass) NOT NULL,
    suggestion_id bigint NOT NULL,
    user_id bigint NOT NULL,
    vote_type text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE public.suggestion_votes
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('suggestion_votes_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS suggestion_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS user_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS vote_type text NOT NULL,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;
ALTER SEQUENCE public.suggestion_votes_id_seq OWNED BY public.suggestion_votes.id;

CREATE SEQUENCE IF NOT EXISTS public.suggestions_id_seq;
CREATE TABLE IF NOT EXISTS public.suggestions (
    id bigint DEFAULT nextval('suggestions_id_seq'::regclass) NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    category text DEFAULT ''::text,
    category_id bigint,
    status text DEFAULT 'new'::text,
    is_archived boolean DEFAULT false,
    is_anonymous boolean DEFAULT false,
    user_id bigint,
    poll_id bigint,
    poll_option_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.suggestions
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('suggestions_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS title text NOT NULL,
    ADD COLUMN IF NOT EXISTS description text NOT NULL,
    ADD COLUMN IF NOT EXISTS category text DEFAULT ''::text,
    ADD COLUMN IF NOT EXISTS category_id bigint,
    ADD COLUMN IF NOT EXISTS status text DEFAULT 'new'::text,
    ADD COLUMN IF NOT EXISTS is_archived boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS is_anonymous boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS user_id bigint,
    ADD COLUMN IF NOT EXISTS poll_id bigint,
    ADD COLUMN IF NOT EXISTS poll_option_id bigint,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.suggestions_id_seq OWNED BY public.suggestions.id;

CREATE SEQUENCE IF NOT EXISTS public.tags_id_seq;
CREATE TABLE IF NOT EXISTS public.tags (
    id bigint DEFAULT nextval('tags_id_seq'::regclass) NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    color text DEFAULT '#6B7280'::text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.tags
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('tags_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS name text NOT NULL,
    ADD COLUMN IF NOT EXISTS slug text NOT NULL,
    ADD COLUMN IF NOT EXISTS color text DEFAULT '#6B7280'::text,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.tags_id_seq OWNED BY public.tags.id;

CREATE SEQUENCE IF NOT EXISTS public.user_achievements_id_seq;
CREATE TABLE IF NOT EXISTS public.user_achievements (
    id bigint DEFAULT nextval('user_achievements_id_seq'::regclass) NOT NULL,
    user_id bigint,
    achievement_id bigint,
    unlocked_at timestamp with time zone
);
ALTER TABLE public.user_achievements
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('user_achievements_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS user_id bigint,
    ADD COLUMN IF NOT EXISTS achievement_id bigint,
    ADD COLUMN IF NOT EXISTS unlocked_at timestamp with time zone;
ALTER SEQUENCE public.user_achievements_id_seq OWNED BY public.user_achievements.id;

CREATE TABLE IF NOT EXISTS public.user_favorites (
    user_id bigint NOT NULL,
    application_id bigint NOT NULL
);
ALTER TABLE public.user_favorites
    ADD COLUMN IF NOT EXISTS user_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS application_id bigint NOT NULL;

CREATE TABLE IF NOT EXISTS public.user_groups (
    group_id bigint NOT NULL,
    user_id bigint NOT NULL
);
ALTER TABLE public.user_groups
    ADD COLUMN IF NOT EXISTS group_id bigint NOT NULL,
    ADD COLUMN IF NOT EXISTS user_id bigint NOT NULL;

CREATE SEQUENCE IF NOT EXISTS public.users_id_seq;
CREATE TABLE IF NOT EXISTS public.users (
    id bigint DEFAULT nextval('users_id_seq'::regclass) NOT NULL,
    username text NOT NULL,
    email text NOT NULL,
    password text,
    first_name text,
    last_name text,
    role text DEFAULT 'user'::text,
    is_active boolean DEFAULT true,
    sso_provider text,
    sso_id text,
    last_login timestamp with time zone,
    avatar_url text,
    phone text,
    department text,
    job_title text,
    location text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('users_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS username text NOT NULL,
    ADD COLUMN IF NOT EXISTS email text NOT NULL,
    ADD COLUMN IF NOT EXISTS password text,
    ADD COLUMN IF NOT EXISTS first_name text,
    ADD COLUMN IF NOT EXISTS last_name text,
    ADD COLUMN IF NOT EXISTS role text DEFAULT 'user'::text,
    ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true,
    ADD COLUMN IF NOT EXISTS sso_provider text,
    ADD COLUMN IF NOT EXISTS sso_id text,
    ADD COLUMN IF NOT EXISTS last_login timestamp with time zone,
    ADD COLUMN IF NOT EXISTS avatar_url text,
    ADD COLUMN IF NOT EXISTS phone text,
    ADD COLUMN IF NOT EXISTS department text,
    ADD COLUMN IF NOT EXISTS job_title text,
    ADD COLUMN IF NOT EXISTS location text,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

CREATE SEQUENCE IF NOT EXISTS public.xp_transactions_id_seq;
CREATE TABLE IF NOT EXISTS public.xp_transactions (
    id bigint DEFAULT nextval('xp_transactions_id_seq'::regclass) NOT NULL,
    user_id bigint,
    amount bigint,
    reason text,
    metadata jsonb,
    created_at timestamp with time zone
);
ALTER TABLE public.xp_transactions
    ADD COLUMN IF NOT EXISTS id bigint DEFAULT nextval('xp_transactions_id_seq'::regclass) NOT NULL,
    ADD COLUMN IF NOT EXISTS user_id bigint,
    ADD COLUMN IF NOT EXISTS amount bigint,
    ADD COLUMN IF NOT EXISTS reason text,
    ADD COLUMN IF NOT EXISTS metadata jsonb,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone;
ALTER SEQUENCE public.xp_transactions_id_seq OWNED BY public.xp_transactions.id;

-- Clés primaires
SELECT pg_temp.ensure_pk('public.achievements', 'achievements_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.announcements', 'announcements_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.app_groups', 'app_groups_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.app_settings', 'app_settings_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.application_clicks', 'application_clicks_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.applications', 'applications_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.chat_messages', 'chat_messages_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.chatbots', 'chatbots_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.comment_settings', 'comment_settings_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.comments', 'comments_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.directory_fields', 'directory_fields_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.directory_people', 'directory_people_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.email_notification_logs', 'email_notification_logs_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.email_o_auth_configs', 'email_o_auth_configs_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.email_templates', 'email_templates_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.event_categories', 'event_categories_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.event_tags', 'event_tags_pkey', 'PRIMARY KEY (event_id, tag_id)');
SELECT pg_temp.ensure_pk('public.event_target_groups', 'event_target_groups_pkey', 'PRIMARY KEY (event_id, group_id)');
SELECT pg_temp.ensure_pk('public.events', 'events_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.feedbacks', 'feedbacks_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.gamification_profiles', 'gamification_profiles_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.gamification_rules', 'gamification_rules_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.group_admins', 'group_admins_pkey', 'PRIMARY KEY (user_id, group_id)');
SELECT pg_temp.ensure_pk('public.group_app_groups', 'group_app_groups_pkey', 'PRIMARY KEY (app_group_id, group_id)');
SELECT pg_temp.ensure_pk('public.groups', 'groups_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.hero_messages', 'hero_messages_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.media', 'media_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.news', 'news_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.news_categories', 'news_categories_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.news_reactions', 'news_reactions_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.news_reads', 'news_reads_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.news_tags', 'news_tags_pkey', 'PRIMARY KEY (news_id, tag_id)');
SELECT pg_temp.ensure_pk('public.news_target_groups', 'news_target_groups_pkey', 'PRIMARY KEY (news_id, group_id)');
SELECT pg_temp.ensure_pk('public.news_types', 'news_types_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.notifications', 'notifications_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.o_auth_providers', 'o_auth_providers_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.org_positions', 'org_positions_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.org_units', 'org_units_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.poll_options', 'poll_options_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.poll_target_groups', 'poll_target_groups_pkey', 'PRIMARY KEY (poll_id, group_id)');
SELECT pg_temp.ensure_pk('public.poll_votes', 'poll_votes_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.polls', 'polls_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.position_assignments', 'position_assignments_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.smtp_configs', 'smtp_configs_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.suggestion_categories', 'suggestion_categories_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.suggestion_votes', 'suggestion_votes_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.suggestions', 'suggestions_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.tags', 'tags_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.user_achievements', 'user_achievements_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.user_favorites', 'user_favorites_pkey', 'PRIMARY KEY (user_id, application_id)');
SELECT pg_temp.ensure_pk('public.user_groups', 'user_groups_pkey', 'PRIMARY KEY (group_id, user_id)');
SELECT pg_temp.ensure_pk('public.users', 'users_pkey', 'PRIMARY KEY (id)');
SELECT pg_temp.ensure_pk('public.xp_transactions', 'xp_transactions_pkey', 'PRIMARY KEY (id)');

-- Contraintes d'unicité (tag gorm:"unique")
SELECT pg_temp.ensure_unique('public.app_groups', 'app_groups_name_key', ARRAY['name']);
SELECT pg_temp.ensure_unique('public.groups', 'groups_name_key', ARRAY['name']);
SELECT pg_temp.ensure_unique('public.media', 'media_storage_path_key', ARRAY['storage_path']);
SELECT pg_temp.ensure_unique('public.o_auth_providers', 'o_auth_providers_provider_name_key', ARRAY['provider_name']);
SELECT pg_temp.ensure_unique('public.users', 'users_email_key', ARRAY['email']);
SELECT pg_temp.ensure_unique('public.users', 'users_username_key', ARRAY['username']);

-- Index (tags gorm:"index" / "uniqueIndex")
SELECT pg_temp.try_unique_index('CREATE UNIQUE INDEX IF NOT EXISTS idx_achievements_code ON public.achievements USING btree (code)');
CREATE INDEX IF NOT EXISTS idx_achievements_deleted_at ON public.achievements USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_announcements_deleted_at ON public.announcements USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_app_groups_deleted_at ON public.app_groups USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_applications_deleted_at ON public.applications USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON public.chat_messages USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_chat_messages_deleted_at ON public.chat_messages USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_chat_messages_group_id ON public.chat_messages USING btree (group_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_is_read ON public.chat_messages USING btree (is_read);
CREATE INDEX IF NOT EXISTS idx_chat_messages_recipient_id ON public.chat_messages USING btree (recipient_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_sender_id ON public.chat_messages USING btree (sender_id);
CREATE INDEX IF NOT EXISTS idx_chatbots_deleted_at ON public.chatbots USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_comments_deleted_at ON public.comments USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_directory_fields_deleted_at ON public.directory_fields USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_directory_people_deleted_at ON public.directory_people USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_directory_people_user_id ON public.directory_people USING btree (user_id);
CREATE INDEX IF NOT EXISTS idx_email_notification_logs_template_type ON public.email_notification_logs USING btree (template_type);
SELECT pg_temp.try_unique_index('CREATE UNIQUE INDEX IF NOT EXISTS idx_email_o_auth_configs_smtp_config_id ON public.email_o_auth_configs USING btree (smtp_config_id)');
SELECT pg_temp.try_unique_index('CREATE UNIQUE INDEX IF NOT EXISTS idx_email_templates_type ON public.email_templates USING btree (type)');
CREATE INDEX IF NOT EXISTS idx_event_categories_deleted_at ON public.event_categories USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_events_author_id ON public.events USING btree (author_id);
CREATE INDEX IF NOT EXISTS idx_events_date_range ON public.events USING btree (start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_events_deleted_at ON public.events USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_events_is_holiday ON public.events USING btree (is_holiday);
CREATE INDEX IF NOT EXISTS idx_events_is_published ON public.events USING btree (is_published);
CREATE INDEX IF NOT EXISTS idx_events_is_recurring ON public.events USING btree (is_recurring);
CREATE INDEX IF NOT EXISTS idx_gamification_profiles_deleted_at ON public.gamification_profiles USING btree (deleted_at);
SELECT pg_temp.try_unique_index('CREATE UNIQUE INDEX IF NOT EXISTS idx_gamification_profiles_user_id ON public.gamification_profiles USING btree (user_id)');
CREATE INDEX IF NOT EXISTS idx_gamification_rules_deleted_at ON public.gamification_rules USING btree (deleted_at);
SELECT pg_temp.try_unique_index('CREATE UNIQUE INDEX IF NOT EXISTS idx_gamification_rules_reason ON public.gamification_rules USING btree (reason)');
CREATE INDEX IF NOT EXISTS idx_groups_deleted_at ON public.groups USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_hero_messages_deleted_at ON public.hero_messages USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_media_deleted_at ON public.media USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_news_deleted_at ON public.news USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_news_type ON public.news USING btree (type);
CREATE INDEX IF NOT EXISTS idx_news_categories_deleted_at ON public.news_categories USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_news_types_deleted_at ON public.news_types USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON public.notifications USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at ON public.notifications USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON public.notifications USING btree (is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON public.notifications USING btree (user_id);
CREATE INDEX IF NOT EXISTS idx_org_positions_deleted_at ON public.org_positions USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_org_positions_unit_id ON public.org_positions USING btree (unit_id);
CREATE INDEX IF NOT EXISTS idx_org_units_deleted_at ON public.org_units USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_org_units_parent_id ON public.org_units USING btree (parent_id);
CREATE INDEX IF NOT EXISTS idx_poll_options_deleted_at ON public.poll_options USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_poll_options_poll_id ON public.poll_options USING btree (poll_id);
CREATE INDEX IF NOT EXISTS idx_poll_votes_poll_id ON public.poll_votes USING btree (poll_id);
CREATE INDEX IF NOT EXISTS idx_poll_votes_poll_option_id ON public.poll_votes USING btree (poll_option_id);
CREATE INDEX IF NOT EXISTS idx_poll_votes_user_id ON public.poll_votes USING btree (user_id);
CREATE INDEX IF NOT EXISTS idx_polls_deleted_at ON public.polls USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_position_assignments_deleted_at ON public.position_assignments USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_position_assignments_person_id ON public.position_assignments USING btree (person_id);
CREATE INDEX IF NOT EXISTS idx_position_assignments_position_id ON public.position_assignments USING btree (position_id);
CREATE INDEX IF NOT EXISTS idx_suggestion_categories_deleted_at ON public.suggestion_categories USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_suggestion_votes_suggestion_id ON public.suggestion_votes USING btree (suggestion_id);
CREATE INDEX IF NOT EXISTS idx_suggestion_votes_user_id ON public.suggestion_votes USING btree (user_id);
CREATE INDEX IF NOT EXISTS idx_suggestions_deleted_at ON public.suggestions USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_tags_deleted_at ON public.tags USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_user_achievements_achievement_id ON public.user_achievements USING btree (achievement_id);
CREATE INDEX IF NOT EXISTS idx_user_achievements_user_id ON public.user_achievements USING btree (user_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON public.users USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_xp_transactions_created_at ON public.xp_transactions USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_xp_transactions_user_id ON public.xp_transactions USING btree (user_id);

-- Clés étrangères
SELECT pg_temp.ensure_fk('public.app_groups', 'fk_app_groups_owner_group', 'FOREIGN KEY (owner_group_id) REFERENCES groups(id)');
SELECT pg_temp.ensure_fk('public.application_clicks', 'fk_application_clicks_application', 'FOREIGN KEY (application_id) REFERENCES applications(id)');
SELECT pg_temp.ensure_fk('public.application_clicks', 'fk_application_clicks_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.applications', 'fk_app_groups_applications', 'FOREIGN KEY (app_group_id) REFERENCES app_groups(id)');
SELECT pg_temp.ensure_fk('public.applications', 'fk_applications_created_by', 'FOREIGN KEY (created_by_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.chat_messages', 'fk_chat_messages_group', 'FOREIGN KEY (group_id) REFERENCES groups(id)');
SELECT pg_temp.ensure_fk('public.chat_messages', 'fk_chat_messages_recipient', 'FOREIGN KEY (recipient_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.chat_messages', 'fk_chat_messages_sender', 'FOREIGN KEY (sender_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.comments', 'fk_comments_moderator', 'FOREIGN KEY (moderated_by) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.comments', 'fk_comments_parent', 'FOREIGN KEY (parent_id) REFERENCES comments(id)');
SELECT pg_temp.ensure_fk('public.comments', 'fk_comments_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.email_o_auth_configs', 'fk_smtp_configs_email_o_auth_config', 'FOREIGN KEY (smtp_config_id) REFERENCES smtp_configs(id)');
SELECT pg_temp.ensure_fk('public.event_tags', 'fk_event_tags_event', 'FOREIGN KEY (event_id) REFERENCES events(id)');
SELECT pg_temp.ensure_fk('public.event_tags', 'fk_event_tags_tag', 'FOREIGN KEY (tag_id) REFERENCES tags(id)');
SELECT pg_temp.ensure_fk('public.event_target_groups', 'fk_event_target_groups_event', 'FOREIGN KEY (event_id) REFERENCES events(id)');
SELECT pg_temp.ensure_fk('public.event_target_groups', 'fk_event_target_groups_group', 'FOREIGN KEY (group_id) REFERENCES groups(id)');
SELECT pg_temp.ensure_fk('public.events', 'fk_events_author', 'FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE');
SELECT pg_temp.ensure_fk('public.events', 'fk_events_category', 'FOREIGN KEY (category_id) REFERENCES event_categories(id)');
SELECT pg_temp.ensure_fk('public.feedbacks', 'fk_feedbacks_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.gamification_profiles', 'fk_gamification_profiles_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.group_admins', 'fk_group_admins_group', 'FOREIGN KEY (group_id) REFERENCES groups(id)');
SELECT pg_temp.ensure_fk('public.group_admins', 'fk_group_admins_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.group_app_groups', 'fk_group_app_groups_app_group', 'FOREIGN KEY (app_group_id) REFERENCES app_groups(id)');
SELECT pg_temp.ensure_fk('public.group_app_groups', 'fk_group_app_groups_group', 'FOREIGN KEY (group_id) REFERENCES groups(id)');
SELECT pg_temp.ensure_fk('public.media', 'fk_media_uploader', 'FOREIGN KEY (uploaded_by) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.news', 'fk_news_author', 'FOREIGN KEY (author_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.news', 'fk_news_category', 'FOREIGN KEY (category_id) REFERENCES news_categories(id)');
SELECT pg_temp.ensure_fk('public.news', 'fk_news_news_type', 'FOREIGN KEY (news_type_id) REFERENCES news_types(id)');
SELECT pg_temp.ensure_fk('public.news_reactions', 'fk_news_reactions', 'FOREIGN KEY (news_id) REFERENCES news(id)');
SELECT pg_temp.ensure_fk('public.news_reactions', 'fk_news_reactions_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.news_reads', 'fk_news_reads_news', 'FOREIGN KEY (news_id) REFERENCES news(id)');
SELECT pg_temp.ensure_fk('public.news_reads', 'fk_news_reads_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.news_tags', 'fk_news_tags_news', 'FOREIGN KEY (news_id) REFERENCES news(id)');
SELECT pg_temp.ensure_fk('public.news_tags', 'fk_news_tags_tag', 'FOREIGN KEY (tag_id) REFERENCES tags(id)');
SELECT pg_temp.ensure_fk('public.news_target_groups', 'fk_news_target_groups_group', 'FOREIGN KEY (group_id) REFERENCES groups(id)');
SELECT pg_temp.ensure_fk('public.news_target_groups', 'fk_news_target_groups_news', 'FOREIGN KEY (news_id) REFERENCES news(id)');
SELECT pg_temp.ensure_fk('public.notifications', 'fk_notifications_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.org_positions', 'fk_org_units_positions', 'FOREIGN KEY (unit_id) REFERENCES org_units(id)');
SELECT pg_temp.ensure_fk('public.poll_options', 'fk_polls_options', 'FOREIGN KEY (poll_id) REFERENCES polls(id)');
SELECT pg_temp.ensure_fk('public.poll_target_groups', 'fk_poll_target_groups_group', 'FOREIGN KEY (group_id) REFERENCES groups(id)');
SELECT pg_temp.ensure_fk('public.poll_target_groups', 'fk_poll_target_groups_poll', 'FOREIGN KEY (poll_id) REFERENCES polls(id)');
SELECT pg_temp.ensure_fk('public.poll_votes', 'fk_poll_options_votes', 'FOREIGN KEY (poll_option_id) REFERENCES poll_options(id)');
SELECT pg_temp.ensure_fk('public.poll_votes', 'fk_poll_votes_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.poll_votes', 'fk_polls_votes', 'FOREIGN KEY (poll_id) REFERENCES polls(id)');
SELECT pg_temp.ensure_fk('public.polls', 'fk_polls_announcement', 'FOREIGN KEY (announcement_id) REFERENCES announcements(id)');
SELECT pg_temp.ensure_fk('public.polls', 'fk_polls_author', 'FOREIGN KEY (author_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.polls', 'fk_polls_news', 'FOREIGN KEY (news_id) REFERENCES news(id)');
SELECT pg_temp.ensure_fk('public.position_assignments', 'fk_org_positions_assignments', 'FOREIGN KEY (position_id) REFERENCES org_positions(id)');
SELECT pg_temp.ensure_fk('public.suggestion_votes', 'fk_suggestions_votes', 'FOREIGN KEY (suggestion_id) REFERENCES suggestions(id)');
SELECT pg_temp.ensure_fk('public.suggestions', 'fk_suggestions_category_ref', 'FOREIGN KEY (category_id) REFERENCES suggestion_categories(id)');
SELECT pg_temp.ensure_fk('public.suggestions', 'fk_suggestions_poll', 'FOREIGN KEY (poll_id) REFERENCES polls(id)');
SELECT pg_temp.ensure_fk('public.suggestions', 'fk_suggestions_poll_option', 'FOREIGN KEY (poll_option_id) REFERENCES poll_options(id)');
SELECT pg_temp.ensure_fk('public.suggestions', 'fk_suggestions_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.user_achievements', 'fk_user_achievements_achievement', 'FOREIGN KEY (achievement_id) REFERENCES achievements(id)');
SELECT pg_temp.ensure_fk('public.user_achievements', 'fk_user_achievements_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.user_favorites', 'fk_user_favorites_application', 'FOREIGN KEY (application_id) REFERENCES applications(id)');
SELECT pg_temp.ensure_fk('public.user_favorites', 'fk_user_favorites_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');
SELECT pg_temp.ensure_fk('public.user_groups', 'fk_user_groups_group', 'FOREIGN KEY (group_id) REFERENCES groups(id)');
SELECT pg_temp.ensure_fk('public.user_groups', 'fk_user_groups_user', 'FOREIGN KEY (user_id) REFERENCES users(id)');

DROP FUNCTION pg_temp.ensure_pk(regclass, text, text);
DROP FUNCTION pg_temp.ensure_unique(regclass, text, text[]);
DROP FUNCTION pg_temp.ensure_fk(regclass, text, text);
DROP FUNCTION pg_temp.try_unique_index(text);

-- +goose Down
-- Pas de retour arrière : supprimer le schéma de référence effacerait toutes les données.
SELECT 1;
