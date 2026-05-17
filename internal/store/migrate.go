package store

import "database/sql"

func Migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS meta (
			key   TEXT PRIMARY KEY,
			value TEXT
		)`,
		`INSERT OR IGNORE INTO meta (key, value) VALUES ('revision', '0')`,
		`INSERT OR IGNORE INTO meta (key, value) VALUES ('updated_at', datetime('now'))`,

		`CREATE TABLE IF NOT EXISTS changelog (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			revision   INTEGER NOT NULL,
			path       TEXT NOT NULL,
			action     TEXT NOT NULL,
			hash       TEXT NOT NULL DEFAULT '',
			client_id  TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_changelog_revision ON changelog(revision)`,

		`CREATE TABLE IF NOT EXISTS files (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			path       TEXT UNIQUE NOT NULL,
			tool       TEXT NOT NULL DEFAULT '',
			category   TEXT NOT NULL DEFAULT '',
			hash       TEXT NOT NULL DEFAULT '',
			version    INTEGER DEFAULT 0,
			size       INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_files_tool ON files(tool)`,
		`CREATE INDEX IF NOT EXISTS idx_files_category ON files(category)`,

		`CREATE TABLE IF NOT EXISTS skills (
			id           TEXT PRIMARY KEY,
			tool         TEXT NOT NULL DEFAULT '',
			category     TEXT NOT NULL DEFAULT '',
			name         TEXT NOT NULL DEFAULT '',
			display_name TEXT NOT NULL DEFAULT '',
			summary      TEXT NOT NULL DEFAULT '',
			description  TEXT NOT NULL DEFAULT '',
			tags         TEXT NOT NULL DEFAULT '[]',
			difficulty   TEXT NOT NULL DEFAULT '',
			usage_count  INTEGER DEFAULT 0,
			avg_rating   REAL DEFAULT 0,
			created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_skills_tool ON skills(tool)`,
		`CREATE INDEX IF NOT EXISTS idx_skills_name ON skills(name)`,

		`CREATE TABLE IF NOT EXISTS skill_examples (
			id         TEXT PRIMARY KEY,
			skill_id   TEXT NOT NULL REFERENCES skills(id),
			type       TEXT NOT NULL DEFAULT 'usage',
			title      TEXT NOT NULL DEFAULT '',
			scenario   TEXT NOT NULL DEFAULT '',
			prompt     TEXT NOT NULL DEFAULT '',
			result     TEXT NOT NULL DEFAULT '',
			lesson     TEXT NOT NULL DEFAULT '',
			rating     INTEGER DEFAULT 0,
			client_id  TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_examples_skill ON skill_examples(skill_id)`,

		`CREATE TABLE IF NOT EXISTS error_cases (
			id                TEXT PRIMARY KEY,
			session_id        TEXT NOT NULL DEFAULT '',
			error_type        TEXT NOT NULL,
			error_fingerprint TEXT NOT NULL DEFAULT '',
			error_count       INTEGER DEFAULT 0,
			root_cause        TEXT NOT NULL DEFAULT '',
			skill_id          TEXT NOT NULL DEFAULT '',
			resolution        TEXT NOT NULL DEFAULT '',
			file_diff         TEXT NOT NULL DEFAULT '',
			lesson            TEXT NOT NULL DEFAULT '',
			tags              TEXT NOT NULL DEFAULT '[]',
			severity          TEXT NOT NULL DEFAULT 'medium',
			client_id         TEXT NOT NULL DEFAULT '',
			created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_error_cases_type ON error_cases(error_type)`,
		`CREATE INDEX IF NOT EXISTS idx_error_cases_skill ON error_cases(skill_id)`,

		`CREATE TABLE IF NOT EXISTS error_clusters (
			id                TEXT PRIMARY KEY,
			error_type        TEXT NOT NULL DEFAULT '',
			fingerprint       TEXT UNIQUE NOT NULL,
			total_count       INTEGER DEFAULT 0,
			sessions          INTEGER DEFAULT 0,
			top_skill         TEXT NOT NULL DEFAULT '',
			resolution_rate   REAL DEFAULT 0,
			common_root_cause TEXT NOT NULL DEFAULT '',
			updated_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS skill_combinations (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			skill_ids   TEXT NOT NULL DEFAULT '[]',
			use_case    TEXT NOT NULL DEFAULT '',
			tags        TEXT NOT NULL DEFAULT '[]',
			usage_count INTEGER DEFAULT 0,
			avg_rating  REAL DEFAULT 0,
			created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS skill_chains (
			id           TEXT PRIMARY KEY,
			name         TEXT NOT NULL DEFAULT '',
			description  TEXT NOT NULL DEFAULT '',
			steps        TEXT NOT NULL DEFAULT '[]',
			use_case     TEXT NOT NULL DEFAULT '',
			usage_count  INTEGER DEFAULT 0,
			success_rate REAL DEFAULT 0,
			created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS skill_usage_log (
			id           TEXT PRIMARY KEY,
			client_id    TEXT NOT NULL DEFAULT '',
			session_id   TEXT NOT NULL DEFAULT '',
			skill_id     TEXT NOT NULL DEFAULT '',
			chain_id     TEXT NOT NULL DEFAULT '',
			context      TEXT NOT NULL DEFAULT '',
			project_type TEXT NOT NULL DEFAULT '',
			invoked_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			duration_ms  INTEGER DEFAULT 0,
			success      INTEGER DEFAULT 0,
			feedback     TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_skill ON skill_usage_log(skill_id)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_client ON skill_usage_log(client_id)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return err
		}
	}
	return nil
}