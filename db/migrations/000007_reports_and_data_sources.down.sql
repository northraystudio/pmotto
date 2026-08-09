DROP TABLE IF EXISTS executive_reports;
DROP TABLE IF EXISTS weekly_reports;
DROP TABLE IF EXISTS daily_reports;
DROP TABLE IF EXISTS project_progress;

-- FK を先に落とさないと source_type 列を削除できない。
ALTER TABLE projects
  DROP FOREIGN KEY fk_projects_source_type;

-- up と対称に、backlog 設定を元の列へ戻してから新しい2列を落とす。
ALTER TABLE projects
  ADD COLUMN backlog_project_id VARCHAR(255) AFTER status;

UPDATE projects
   SET backlog_project_id = source_value
 WHERE source_type = 'backlog'
   AND source_value IS NOT NULL;

ALTER TABLE projects
  DROP COLUMN source_value,
  DROP COLUMN source_type;

DROP TABLE IF EXISTS data_source_types;
