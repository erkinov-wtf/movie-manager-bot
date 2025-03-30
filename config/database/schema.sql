-- public.users definition with new UUID id and renamed existing id to tg_id
CREATE TABLE IF NOT EXISTS users
(
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    tg_id        BIGSERIAL   NOT NULL,
    first_name   TEXT,
    last_name    TEXT,
    username     TEXT UNIQUE,
    language     TEXT        NOT NULL DEFAULT 'en',
    tmdb_api_key TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_tg_id_unique UNIQUE (tg_id)
);

COMMENT ON TABLE users IS 'Stores user information for authentication and preferences';

-- public.movies definition with user_id still BIGINT but now referencing users.tg_id
CREATE TABLE IF NOT EXISTS movies
(
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    user_id    BIGINT      NOT NULL,
    api_id     BIGINT      NOT NULL,
    title      TEXT        NOT NULL,
    runtime    INT         NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT movies_pkey PRIMARY KEY (id),
    CONSTRAINT fk_movies_user FOREIGN KEY (user_id) REFERENCES users (tg_id) ON DELETE CASCADE,
    CONSTRAINT check_runtime_positive CHECK (runtime IS NULL OR runtime > 0)
);

CREATE UNIQUE INDEX idx_movies_user_api_unique ON movies USING btree (user_id, api_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE movies IS 'Stores movie information tracked by users';

-- public.tv_shows definition with user_id still BIGINT but now referencing users.tg_id
CREATE TABLE IF NOT EXISTS tv_shows
(
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    user_id    BIGINT      NOT NULL,
    api_id     BIGINT      NOT NULL,
    name       TEXT        NOT NULL,
    seasons    INT         NOT NULL,
    episodes   INT         NOT NULL,
    runtime    INT         NOT NULL,
    status     TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT tv_shows_pkey PRIMARY KEY (id),
    CONSTRAINT fk_tv_shows_user FOREIGN KEY (user_id) REFERENCES users (tg_id) ON DELETE CASCADE,
    CONSTRAINT check_tv_positive_values CHECK (seasons > 0 AND episodes > 0 AND runtime > 0)
);

CREATE UNIQUE INDEX idx_tv_shows_user_api_unique ON tv_shows USING btree (user_id, api_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE tv_shows IS 'Stores TV show information tracked by users';

-- public.watchlists definition with user_id still BIGINT but now referencing users.tg_id
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
    CONSTRAINT fk_watchlists_user FOREIGN KEY (user_id) REFERENCES users (tg_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_watchlists_user_show_api_unique ON watchlists USING btree (user_id, show_api_id, type) WHERE deleted_at IS NULL;

COMMENT ON TABLE watchlists IS 'Stores shows and movies users want to watch';

-- Create worker_states table to track the state of workers
CREATE TABLE IF NOT EXISTS worker_states
(
    id              UUID                     NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    worker_id       VARCHAR(255)             NOT NULL UNIQUE,
    worker_type     VARCHAR(50)              NOT NULL,
    status          VARCHAR(20)              NOT NULL DEFAULT 'idle',
    last_check_time TIMESTAMP WITH TIME ZONE,
    next_check_time TIMESTAMP WITH TIME ZONE,
    error           TEXT,
    shows_checked   INTEGER                  NOT NULL DEFAULT 0,
    updates_found   INTEGER                  NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on worker_id
CREATE INDEX IF NOT EXISTS idx_worker_states_worker_id ON worker_states (worker_id);

-- Create worker_tasks table to track individual tasks performed by workers
CREATE TABLE IF NOT EXISTS worker_tasks
(
    id            UUID                     NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    worker_id     VARCHAR(255)             NOT NULL,
    task_type     VARCHAR(50)              NOT NULL,
    status        VARCHAR(20)              NOT NULL DEFAULT 'running',
    start_time    TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time      TIMESTAMP WITH TIME ZONE,
    duration_ms   BIGINT,
    error         TEXT,
    show_id       BIGINT,
    user_id       BIGINT,
    shows_checked INTEGER,
    updates_found INTEGER,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_worker_tasks_worker_id FOREIGN KEY (worker_id) REFERENCES worker_states (worker_id) ON DELETE NO ACTION
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_worker_tasks_worker_id ON worker_tasks (worker_id);
CREATE INDEX IF NOT EXISTS idx_worker_tasks_status ON worker_tasks (status);
CREATE INDEX IF NOT EXISTS idx_worker_tasks_created_at ON worker_tasks (created_at);

-- Create a view for easy monitoring of worker performance
CREATE OR REPLACE VIEW worker_performance AS
SELECT w.worker_id,
       w.worker_type,
       w.status,
       w.last_check_time,
       w.next_check_time,
       w.shows_checked,
       w.updates_found,
       w.error,
       COUNT(t.id)                                         AS total_tasks,
       AVG(t.duration_ms)                                  AS avg_task_duration_ms,
       MAX(t.duration_ms)                                  AS max_task_duration_ms,
       SUM(CASE WHEN t.status = 'error' THEN 1 ELSE 0 END) AS error_tasks,
       MAX(t.created_at)                                   AS last_task_time
FROM worker_states w
         LEFT JOIN
     worker_tasks t ON w.worker_id = t.worker_id
GROUP BY w.id, w.worker_id, w.worker_type, w.status, w.last_check_time,
         w.next_check_time, w.shows_checked, w.updates_found, w.error;

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