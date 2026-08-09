-- 進捗の取得元をマルチデータソース化し、n8n が生成するレポートの保存先を用意する。
--
-- 取得元は projects.source_type / source_value の2列で表す。種別を data_source_types
-- マスタに切り出すのは、種別追加をコード変更なしのデータ追加で行えるようにするため
-- （role_functions と同じ考え方）。FK を id ではなく code に張っているのは、n8n の
-- ワークフローが `WHERE source_type = 'spreadsheet'` と素直に書けるようにするため。

CREATE TABLE data_source_types (
  id          INT AUTO_INCREMENT PRIMARY KEY,
  code        VARCHAR(50)  NOT NULL UNIQUE,            -- projects.source_type から参照される業務キー
  label       VARCHAR(100) NOT NULL,                   -- 画面のセレクトに出す表示名
  value_label VARCHAR(100) NOT NULL,                   -- source_value 入力欄のラベル（種別で意味が変わるため）
  sort_order  INT     NOT NULL DEFAULT 0,
  is_active   BOOLEAN NOT NULL DEFAULT true,           -- 削除は論理削除のみ（過去のプロジェクト設定を壊さない）
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

INSERT INTO data_source_types (code, label, value_label, sort_order) VALUES
  ('spreadsheet', 'Google スプレッドシート', 'スプレッドシートID',        1),
  ('backlog',     'Backlog',                'Backlog プロジェクトキー', 2);

-- backlog_project_id は source_type='backlog' + source_value と意味が重複するため廃止する。
-- 既存値は取りこぼさないよう、列を落とす前に新しい2列へ移送する。
ALTER TABLE projects
  ADD COLUMN source_type  VARCHAR(50)  AFTER status,   -- data_source_types.code。未設定なら収集対象外
  ADD COLUMN source_value VARCHAR(255) AFTER source_type;

UPDATE projects
   SET source_type  = 'backlog',
       source_value = backlog_project_id
 WHERE backlog_project_id IS NOT NULL
   AND backlog_project_id <> '';

ALTER TABLE projects
  DROP COLUMN backlog_project_id;

ALTER TABLE projects
  ADD CONSTRAINT fk_projects_source_type
    FOREIGN KEY (source_type) REFERENCES data_source_types(code);

-- 取得元から収集したタスク進捗（日次ワークフローが当日分を洗い替えする）
CREATE TABLE project_progress (
  id           INT AUTO_INCREMENT PRIMARY KEY,
  project_id   INT NOT NULL,
  phase        VARCHAR(100),                           -- 取得元上の管理単位
  task         TEXT,
  status       VARCHAR(50),
  progress     DECIMAL(5, 2),
  deadline     DATE,
  register_day DATE NOT NULL,                          -- 収集日
  created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_project_day (project_id, register_day),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- 以下3テーブルは n8n が直接書き込む。API は読み取りのみ行う。
-- UNIQUE KEY は各ワークフローの ON DUPLICATE KEY UPDATE が依存している。

CREATE TABLE daily_reports (
  id               INT AUTO_INCREMENT PRIMARY KEY,
  project_id       INT  NOT NULL,
  report_date      DATE NOT NULL,
  overall_progress DECIMAL(5, 2),
  phase_summary    JSON,
  risks            JSON,
  trend            JSON,
  ai_comment       TEXT,
  created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uq_project_date (project_id, report_date),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE weekly_reports (
  id              INT AUTO_INCREMENT PRIMARY KEY,
  project_id      INT  NOT NULL,
  week_start_date DATE NOT NULL,
  week_end_date   DATE NOT NULL,
  summary         JSON,
  ai_comment      TEXT,
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uq_project_week (project_id, week_start_date),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- 全プロジェクト横断。プロジェクトに紐づかないため FK は持たない。
CREATE TABLE executive_reports (
  id          INT AUTO_INCREMENT PRIMARY KEY,
  report_date DATE        NOT NULL,
  report_type VARCHAR(50) NOT NULL DEFAULT 'weekly',
  content     JSON,
  ai_comment  TEXT,
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uq_date_type (report_date, report_type)
);
