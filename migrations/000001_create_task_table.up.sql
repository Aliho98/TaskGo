-- Drop existing table if it exists
DROP TABLE IF EXISTS task CASCADE;
DROP FUNCTION IF EXISTS set_updated_at CASCADE;
-- Step 1: enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Step 2: create the table
CREATE TABLE task (
                      id          UUID         NOT NULL DEFAULT uuid_generate_v4(),
                      title       VARCHAR(255) NOT NULL,
                      description TEXT,
                      status      VARCHAR(50)  NOT NULL DEFAULT 'pending',
                      created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                      updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                      CONSTRAINT task_pkey PRIMARY KEY (id)
);

-- Step 3: create the trigger function
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 4: attach the trigger to the table
CREATE TRIGGER trg_task_updated_at
    BEFORE UPDATE ON task
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();