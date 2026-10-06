-- Reflections are answered per template attribute (key -> text) instead of
-- a single free-text content column.

ALTER TABLE sprint_reflections ADD COLUMN IF NOT EXISTS answers JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE sprint_reflections DROP COLUMN IF EXISTS content;
