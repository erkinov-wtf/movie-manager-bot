-- public.users definition
CREATE TABLE IF NOT EXISTS users
(
    id           BIGSERIAL   NOT NULL,
    first_name    TEXT,
    last_name    TEXT,
    username     TEXT UNIQUE,
    language     TEXT        NOT NULL DEFAULT 'en',
    tmdb_api_key TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_pkey PRIMARY KEY (id)
);

COMMENT ON TABLE users IS 'Stores user information for authentication and preferences';

-- public.movies definition
CREATE TABLE IF NOT EXISTS movies
(
    id         UUID        NOT NULL DEFAULT gen_random_uuid(), -- Added default UUID generation
    user_id    BIGINT      NOT NULL,                           -- Changed to BIGINT for consistency with users.id
    api_id     BIGINT      NOT NULL,
    title      TEXT        NOT NULL,
    runtime    INT,                                            -- Changed from INT8 to INT which is usually sufficient for runtime in minutes
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT movies_pkey PRIMARY KEY (id),
    CONSTRAINT fk_movies_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT check_runtime_positive CHECK (runtime IS NULL OR runtime > 0)
);

CREATE UNIQUE INDEX idx_movies_user_api_unique ON movies USING btree (user_id, api_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE movies IS 'Stores movie information tracked by users';

-- public.tv_shows definition
CREATE TABLE IF NOT EXISTS tv_shows
(
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    user_id    BIGINT      NOT NULL, -- Changed to BIGINT for consistency
    api_id     BIGINT      NOT NULL,
    name       TEXT        NOT NULL, -- Changed varchar to TEXT for consistency
    seasons    INT         NOT NULL,
    episodes   INT         NOT NULL,
    runtime    INT         NOT NULL,
    status     TEXT        NOT NULL, -- Changed varchar to TEXT for consistency
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT tv_shows_pkey PRIMARY KEY (id),
    CONSTRAINT fk_tv_shows_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT check_tv_positive_values CHECK (seasons > 0 AND episodes > 0 AND runtime > 0)
);

CREATE UNIQUE INDEX idx_tv_shows_user_api_unique ON tv_shows USING btree (user_id, api_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE tv_shows IS 'Stores TV show information tracked by users';

-- public.watchlists definition
CREATE TABLE IF NOT EXISTS watchlists
(
    id          UUID        NOT NULL DEFAULT gen_random_uuid(),
    user_id     BIGINT      NOT NULL,
    show_api_id BIGINT      NOT NULL,
    type        TEXT        NOT NULL,
    title       TEXT        NOT NULL,
    image       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,

    CONSTRAINT watchlists_pkey PRIMARY KEY (id),
    CONSTRAINT fk_watchlists_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_watchlists_user_show_api_unique ON watchlists USING btree (user_id, show_api_id, type) WHERE deleted_at IS NULL;

COMMENT ON TABLE watchlists IS 'Stores shows and movies users want to watch';

-- Function for updating timestamps
CREATE OR REPLACE FUNCTION update_modified_column()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at columns
CREATE TRIGGER update_users_timestamp
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_movies_timestamp
    BEFORE UPDATE
    ON movies
    FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_tv_shows_timestamp
    BEFORE UPDATE
    ON tv_shows
    FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_watchlists_timestamp
    BEFORE UPDATE
    ON watchlists
    FOR EACH ROW
EXECUTE FUNCTION update_modified_column();