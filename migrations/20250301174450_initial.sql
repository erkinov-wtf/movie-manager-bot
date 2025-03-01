-- Create "update_modified_column" function
CREATE FUNCTION "update_modified_column" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;
-- Create "users" table
CREATE TABLE "users" (
  "id" bigserial NOT NULL,
  "first_name" text NULL,
  "last_name" text NULL,
  "username" text NULL,
  "language" text NOT NULL DEFAULT 'en',
  "tmdb_api_key" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "users_username_key" UNIQUE ("username")
);
-- Set comment to table: "users"
COMMENT ON TABLE "users" IS 'Stores user information for authentication and preferences';
-- Create "movies" table
CREATE TABLE "movies" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" bigint NOT NULL,
  "api_id" bigint NOT NULL,
  "title" text NOT NULL,
  "runtime" integer NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_movies_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "check_runtime_positive" CHECK ((runtime IS NULL) OR (runtime > 0))
);
-- Create index "idx_movies_user_api_unique" to table: "movies"
CREATE UNIQUE INDEX "idx_movies_user_api_unique" ON "movies" ("user_id", "api_id") WHERE (deleted_at IS NULL);
-- Set comment to table: "movies"
COMMENT ON TABLE "movies" IS 'Stores movie information tracked by users';
-- Create trigger "update_movies_timestamp"
CREATE TRIGGER "update_movies_timestamp" BEFORE UPDATE ON "movies" FOR EACH ROW EXECUTE FUNCTION "update_modified_column"();
-- Create "tv_shows" table
CREATE TABLE "tv_shows" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" bigint NOT NULL,
  "api_id" bigint NOT NULL,
  "name" text NOT NULL,
  "seasons" integer NOT NULL,
  "episodes" integer NOT NULL,
  "runtime" integer NOT NULL,
  "status" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tv_shows_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "check_tv_positive_values" CHECK ((seasons > 0) AND (episodes > 0) AND (runtime > 0))
);
-- Create index "idx_tv_shows_user_api_unique" to table: "tv_shows"
CREATE UNIQUE INDEX "idx_tv_shows_user_api_unique" ON "tv_shows" ("user_id", "api_id") WHERE (deleted_at IS NULL);
-- Set comment to table: "tv_shows"
COMMENT ON TABLE "tv_shows" IS 'Stores TV show information tracked by users';
-- Create trigger "update_tv_shows_timestamp"
CREATE TRIGGER "update_tv_shows_timestamp" BEFORE UPDATE ON "tv_shows" FOR EACH ROW EXECUTE FUNCTION "update_modified_column"();
-- Create trigger "update_users_timestamp"
CREATE TRIGGER "update_users_timestamp" BEFORE UPDATE ON "users" FOR EACH ROW EXECUTE FUNCTION "update_modified_column"();
-- Create "watchlists" table
CREATE TABLE "watchlists" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" bigint NOT NULL,
  "show_api_id" bigint NOT NULL,
  "type" text NOT NULL,
  "title" text NOT NULL,
  "image" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_watchlists_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_watchlists_user_show_api_unique" to table: "watchlists"
CREATE UNIQUE INDEX "idx_watchlists_user_show_api_unique" ON "watchlists" ("user_id", "show_api_id", "type") WHERE (deleted_at IS NULL);
-- Set comment to table: "watchlists"
COMMENT ON TABLE "watchlists" IS 'Stores shows and movies users want to watch';
-- Create trigger "update_watchlists_timestamp"
CREATE TRIGGER "update_watchlists_timestamp" BEFORE UPDATE ON "watchlists" FOR EACH ROW EXECUTE FUNCTION "update_modified_column"();
