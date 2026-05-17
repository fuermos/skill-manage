package store

import (
	"database/sql"
	"time"

	"skill-manage/internal/model"
)

type SyncStore struct {
	db *sql.DB
}

func NewSyncStore(db *sql.DB) *SyncStore {
	return &SyncStore{db: db}
}

func (s *SyncStore) GetRevision() (int64, error) {
	var rev int64
	err := s.db.QueryRow("SELECT CAST(value AS INTEGER) FROM meta WHERE key = 'revision'").Scan(&rev)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return rev, err
}

func (s *SyncStore) IncrementRevision() (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var current int64
	err = tx.QueryRow("SELECT CAST(value AS INTEGER) FROM meta WHERE key = 'revision'").Scan(&current)
	if err != nil {
		return 0, err
	}

	newRev := current + 1
	_, err = tx.Exec("UPDATE meta SET value = ? WHERE key = 'revision'", newRev)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec("UPDATE meta SET value = ? WHERE key = 'updated_at'", time.Now().UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return newRev, nil
}

func (s *SyncStore) GetRevisionInfo() (model.RevisionInfo, error) {
	rev, err := s.GetRevision()
	if err != nil {
		return model.RevisionInfo{}, err
	}

	var updatedAt time.Time
	row := s.db.QueryRow("SELECT value FROM meta WHERE key = 'updated_at'")
	var timeStr string
	if err := row.Scan(&timeStr); err == nil {
		updatedAt, _ = time.Parse("2006-01-02 15:04:05", timeStr)
	}

	return model.RevisionInfo{
		Revision:  rev,
		UpdatedAt: updatedAt,
	}, nil
}

func (s *SyncStore) UpsertFile(f model.FileInfo, version int64) error {
	hash := f.Hash()
	_, err := s.db.Exec(`
		INSERT INTO files (path, tool, category, hash, version, size, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(path) DO UPDATE SET
			hash = excluded.hash,
			version = excluded.version,
			size = excluded.size,
			updated_at = CURRENT_TIMESTAMP
	`, f.Path, f.Tool, f.Category, hash, version, len(f.Content))
	return err
}

func (s *SyncStore) GetFile(path string) (*model.FileInfo, error) {
	var f model.FileInfo
	err := s.db.QueryRow(
		"SELECT path, tool, category, hash, version, size FROM files WHERE path = ?", path,
	).Scan(&f.Path, &f.Tool, &f.Category, &f.ContentHash, &f.Version, &f.Size)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (s *SyncStore) DeleteFile(path string, version int64) error {
	_, err := s.db.Exec("DELETE FROM files WHERE path = ?", path)
	return err
}

func (s *SyncStore) ListFiles(tool, category string) ([]model.FileInfo, error) {
	query := "SELECT path, tool, category, hash, version, size FROM files WHERE 1=1"
	var args []interface{}

	if tool != "" {
		query += " AND tool = ?"
		args = append(args, tool)
	}
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	query += " ORDER BY path"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []model.FileInfo
	for rows.Next() {
		var f model.FileInfo
		if err := rows.Scan(&f.Path, &f.Tool, &f.Category, &f.ContentHash, &f.Version, &f.Size); err != nil {
			return nil, err
		}
		f.Content = []byte{}
		files = append(files, f)
	}
	return files, rows.Err()
}

func (s *SyncStore) WriteChangelog(entry model.ChangelogEntry) error {
	_, err := s.db.Exec(
		"INSERT INTO changelog (revision, path, action, hash, client_id) VALUES (?, ?, ?, ?, ?)",
		entry.Revision, entry.Path, entry.Action.String(), entry.Hash, entry.ClientID,
	)
	return err
}

func (s *SyncStore) GetChangesSince(sinceRevision int64) ([]model.ChangelogEntry, error) {
	rows, err := s.db.Query(
		"SELECT revision, path, action, hash, client_id, created_at FROM changelog WHERE revision > ? ORDER BY revision ASC",
		sinceRevision,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.ChangelogEntry
	for rows.Next() {
		var e model.ChangelogEntry
		var actionStr string
		var createdAt time.Time
		if err := rows.Scan(&e.Revision, &e.Path, &actionStr, &e.Hash, &e.ClientID, &createdAt); err != nil {
			return nil, err
		}
		e.Action = parseChangeType(actionStr)
		e.CreatedAt = createdAt
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *SyncStore) BatchUpsertFiles(files []model.FileInfo, version int64, clientID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, f := range files {
		hash := f.Hash()
		_, err := tx.Exec(`
			INSERT INTO files (path, tool, category, hash, version, size, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(path) DO UPDATE SET
				hash = excluded.hash,
				version = excluded.version,
				size = excluded.size,
				updated_at = CURRENT_TIMESTAMP
		`, f.Path, f.Tool, f.Category, hash, version, len(f.Content))

		if err != nil {
			return err
		}

		changelog := model.ChangelogEntry{
			Revision: version,
			Path:     f.Path,
			Action:   model.ChangeTypeCreate,
			Hash:     hash,
			ClientID: clientID,
		}
		if err := s.writeChangelogTx(tx, changelog); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SyncStore) writeChangelogTx(tx *sql.Tx, entry model.ChangelogEntry) error {
	_, err := tx.Exec(
		"INSERT INTO changelog (revision, path, action, hash, client_id) VALUES (?, ?, ?, ?, ?)",
		entry.Revision, entry.Path, entry.Action.String(), entry.Hash, entry.ClientID,
	)
	return err
}

func (s *SyncStore) GetFilesByPaths(paths []string) ([]model.FileInfo, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	query := "SELECT path, tool, category, hash, version, size FROM files WHERE path IN ("
	args := make([]interface{}, len(paths))
	for i, p := range paths {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = p
	}
	query += ")"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []model.FileInfo
	for rows.Next() {
		var f model.FileInfo
		if err := rows.Scan(&f.Path, &f.Tool, &f.Category, &f.ContentHash, &f.Version, &f.Size); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func parseChangeType(s string) model.ChangeType {
	switch s {
	case "create":
		return model.ChangeTypeCreate
	case "modify":
		return model.ChangeTypeModify
	case "delete":
		return model.ChangeTypeDelete
	default:
		return model.ChangeType(0)
	}
}