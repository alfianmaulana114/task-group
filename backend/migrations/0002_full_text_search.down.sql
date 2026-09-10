BEGIN;

DROP INDEX IF EXISTS idx_tasks_search;
DROP INDEX IF EXISTS idx_projects_search;
DROP INDEX IF EXISTS idx_comments_search;

COMMIT;
