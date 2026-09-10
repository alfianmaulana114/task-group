BEGIN;

CREATE INDEX IF NOT EXISTS idx_tasks_search ON tasks USING GIN (
  to_tsvector('simple', COALESCE(title,'') || ' ' || COALESCE(description,'') || ' ' || COALESCE(task_key,''))
) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_projects_search ON projects USING GIN (
  to_tsvector('simple', COALESCE(name,'') || ' ' || COALESCE(description,''))
) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_comments_search ON comments USING GIN (
  to_tsvector('simple', body)
);

COMMIT;
