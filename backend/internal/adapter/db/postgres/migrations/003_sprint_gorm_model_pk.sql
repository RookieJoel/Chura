-- Switches sprints.id from UUID to a gorm.Model-compatible BIGSERIAL
-- primary key and adds deleted_at for gorm soft-delete support.

ALTER TABLE sprints DROP CONSTRAINT sprints_pkey;
ALTER TABLE sprints DROP COLUMN id;
ALTER TABLE sprints ADD COLUMN id BIGSERIAL PRIMARY KEY;

ALTER TABLE sprints ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_sprints_deleted_at ON sprints (deleted_at);
