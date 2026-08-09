DROP TABLE IF EXISTS executive_reports;
DROP TABLE IF EXISTS weekly_reports;
DROP TABLE IF EXISTS daily_reports;
DROP TABLE IF EXISTS project_progress;

-- FK を先に落とさないと source_type 列を削除できない。
ALTER TABLE projects
  DROP FOREIGN KEY fk_projects_source_type;

ALTER TABLE projects
  DROP COLUMN source_value,
  DROP COLUMN source_type,
  ADD COLUMN backlog_project_id VARCHAR(255) AFTER status;

DROP TABLE IF EXISTS data_source_types;
