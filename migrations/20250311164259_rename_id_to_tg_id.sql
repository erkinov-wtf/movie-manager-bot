-- Step 1: Rename the existing id column to tg_id
ALTER TABLE users RENAME COLUMN id TO tg_id;

-- Step 2: Drop all foreign key constraints that reference users.id (the primary key)
ALTER TABLE movies DROP CONSTRAINT IF EXISTS fk_movies_user;
ALTER TABLE tv_shows DROP CONSTRAINT IF EXISTS fk_tv_shows_user;
ALTER TABLE watchlists DROP CONSTRAINT IF EXISTS fk_watchlists_user;

-- Step 3: Drop the primary key constraint on users (users_pkey)
ALTER TABLE users DROP CONSTRAINT users_pkey CASCADE;

-- Step 4: Ensure tg_id remains BIGSERIAL and unique
ALTER TABLE users ALTER COLUMN tg_id SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_tg_id_unique UNIQUE (tg_id);

-- Step 5: Add the new UUID column (only if it does not already exist)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'id') THEN
ALTER TABLE users ADD COLUMN id UUID DEFAULT gen_random_uuid();
END IF;
END $$;

-- Step 6: Set the new UUID column as the primary key
ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (id);

-- Step 7: Re-add foreign key constraints to reference tg_id instead of id
ALTER TABLE movies ADD CONSTRAINT fk_movies_user FOREIGN KEY (user_id) REFERENCES users (tg_id) ON DELETE CASCADE;
ALTER TABLE tv_shows ADD CONSTRAINT fk_tv_shows_user FOREIGN KEY (user_id) REFERENCES users (tg_id) ON DELETE CASCADE;
ALTER TABLE watchlists ADD CONSTRAINT fk_watchlists_user FOREIGN KEY (user_id) REFERENCES users (tg_id) ON DELETE CASCADE;
