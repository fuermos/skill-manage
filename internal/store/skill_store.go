package store

import (
	"database/sql"
	"encoding/json"

	"skill-manage/internal/model"
)

type SkillStore struct {
	db *sql.DB
}

type UsageStats struct {
	TotalUses int `json:"total_uses"`
	Successes int `json:"successes"`
	Failures  int `json:"failures"`
}

func NewSkillStore(db *sql.DB) *SkillStore {
	return &SkillStore{db: db}
}

func (s *SkillStore) UpsertSkill(sk *model.Skill) error {
	tagsJSON, _ := json.Marshal(sk.Tags)
	_, err := s.db.Exec(`
		INSERT INTO skills (id, tool, category, name, display_name, summary, description, tags, difficulty)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			tool = excluded.tool,
			category = excluded.category,
			name = excluded.name,
			display_name = excluded.display_name,
			summary = excluded.summary,
			description = excluded.description,
			tags = excluded.tags,
			difficulty = excluded.difficulty,
			updated_at = CURRENT_TIMESTAMP
	`, sk.ID, sk.Tool, sk.Category, sk.Name, sk.DisplayName, sk.Summary, sk.Description, string(tagsJSON), sk.Difficulty)
	return err
}

func (s *SkillStore) GetSkill(id string) (*model.Skill, error) {
	var sk model.Skill
	var tagsJSON string
	err := s.db.QueryRow(
		"SELECT id, tool, category, name, display_name, summary, description, tags, difficulty, usage_count, avg_rating, created_at, updated_at FROM skills WHERE id = ?",
		id,
	).Scan(&sk.ID, &sk.Tool, &sk.Category, &sk.Name, &sk.DisplayName, &sk.Summary, &sk.Description, &tagsJSON, &sk.Difficulty, &sk.UsageCount, &sk.AvgRating, &sk.CreatedAt, &sk.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(tagsJSON), &sk.Tags)
	return &sk, nil
}

func (s *SkillStore) ListSkills(tool, category, search string) ([]model.Skill, error) {
	return s.SearchSkills(search, tool, category)
}

func (s *SkillStore) SearchSkills(query, tool, category string) ([]model.Skill, error) {
	sql := "SELECT id, tool, category, name, display_name, summary, description, tags, difficulty, usage_count, avg_rating, created_at, updated_at FROM skills WHERE 1=1"
	var args []interface{}

	if tool != "" {
		sql += " AND tool = ?"
		args = append(args, tool)
	}
	if category != "" {
		sql += " AND category = ?"
		args = append(args, category)
	}
	if query != "" {
		sql += " AND (name LIKE ? OR summary LIKE ? OR description LIKE ?)"
		q := "%" + query + "%"
		args = append(args, q, q, q)
	}
	sql += " ORDER BY usage_count DESC, name ASC"

	return s.querySkills(s.db.Query(sql, args...))
}

func (s *SkillStore) GetSkillsByIDs(ids []string) ([]model.Skill, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	sql := "SELECT id, tool, category, name, display_name, summary, description, tags, difficulty, usage_count, avg_rating, created_at, updated_at FROM skills WHERE id IN ("
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			sql += ","
		}
		sql += "?"
		args[i] = id
	}
	sql += ")"
	return s.querySkills(s.db.Query(sql, args...))
}

func (s *SkillStore) querySkills(rows *sql.Rows, err error) ([]model.Skill, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []model.Skill
	for rows.Next() {
		var sk model.Skill
		var tagsJSON string
		if err := rows.Scan(&sk.ID, &sk.Tool, &sk.Category, &sk.Name, &sk.DisplayName, &sk.Summary, &sk.Description, &tagsJSON, &sk.Difficulty, &sk.UsageCount, &sk.AvgRating, &sk.CreatedAt, &sk.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(tagsJSON), &sk.Tags)
		skills = append(skills, sk)
	}
	return skills, rows.Err()
}

func (s *SkillStore) UpdateSkillUsage(skillID string, success bool, rating float64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int
	var avg float64
	err = tx.QueryRow("SELECT usage_count, avg_rating FROM skills WHERE id = ?", skillID).Scan(&count, &avg)
	if err != nil {
		return err
	}

	newCount := count + 1
	newAvg := avg
	if rating > 0 {
		if avg == 0 {
			newAvg = rating
		} else {
			newAvg = (avg*float64(count) + rating) / float64(newCount)
		}
	}

	_, err = tx.Exec("UPDATE skills SET usage_count = ?, avg_rating = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newCount, newAvg, skillID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SkillStore) AddExample(ex *model.SkillExample) error {
	_, err := s.db.Exec(`
		INSERT INTO skill_examples (id, skill_id, type, title, scenario, prompt, result, lesson, rating, client_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ex.ID, ex.SkillID, ex.Type.String(), ex.Title, ex.Scenario, ex.Prompt, ex.Result, ex.Lesson, ex.Rating, ex.ClientID)
	return err
}

func (s *SkillStore) ListExamples(skillID string, exampleType model.ExampleType) ([]model.SkillExample, error) {
	query := "SELECT id, skill_id, type, title, scenario, prompt, result, lesson, rating, client_id, created_at FROM skill_examples WHERE skill_id = ?"
	var args []interface{}
	args = append(args, skillID)

	if exampleType != "" && exampleType.Valid() {
		query += " AND type = ?"
		args = append(args, exampleType.String())
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var examples []model.SkillExample
	for rows.Next() {
		var ex model.SkillExample
		var typeStr string
		if err := rows.Scan(&ex.ID, &ex.SkillID, &typeStr, &ex.Title, &ex.Scenario, &ex.Prompt, &ex.Result, &ex.Lesson, &ex.Rating, &ex.ClientID, &ex.CreatedAt); err != nil {
			return nil, err
		}
		ex.Type = model.ExampleType(typeStr)
		examples = append(examples, ex)
	}
	return examples, rows.Err()
}

func (s *SkillStore) AddErrorCase(ec *model.ErrorCase) error {
	tagsJSON, _ := json.Marshal(ec.Tags)
	_, err := s.db.Exec(`
		INSERT INTO error_cases (id, session_id, error_type, error_fingerprint, error_count, root_cause, skill_id, resolution, file_diff, lesson, tags, severity, client_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ec.ID, ec.SessionID, ec.ErrorType, ec.ErrorFingerprint, ec.ErrorCount, ec.RootCause, ec.SkillID, ec.Resolution, ec.FileDiff, ec.Lesson, string(tagsJSON), string(ec.Severity), ec.ClientID)
	return err
}

func (s *SkillStore) ListErrorCases(errorType string, limit int) ([]model.ErrorCase, error) {
	query := "SELECT id, session_id, error_type, error_fingerprint, error_count, root_cause, skill_id, resolution, file_diff, lesson, tags, severity, client_id, created_at FROM error_cases"
	var args []interface{}

	if errorType != "" {
		query += " WHERE error_type = ?"
		args = append(args, errorType)
	}
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []model.ErrorCase
	for rows.Next() {
		var ec model.ErrorCase
		var tagsJSON, severity string
		if err := rows.Scan(&ec.ID, &ec.SessionID, &ec.ErrorType, &ec.ErrorFingerprint, &ec.ErrorCount, &ec.RootCause, &ec.SkillID, &ec.Resolution, &ec.FileDiff, &ec.Lesson, &tagsJSON, &severity, &ec.ClientID, &ec.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(tagsJSON), &ec.Tags)
		ec.Severity = model.ErrorSeverity(severity)
		cases = append(cases, ec)
	}
	return cases, rows.Err()
}

func (s *SkillStore) UpsertCombination(c *model.SkillCombination) error {
	skillIDsJSON, _ := json.Marshal(c.SkillIDs)
	tagsJSON, _ := json.Marshal(c.Tags)
	_, err := s.db.Exec(`
		INSERT INTO skill_combinations (id, name, description, skill_ids, use_case, tags)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			skill_ids = excluded.skill_ids,
			use_case = excluded.use_case,
			tags = excluded.tags
	`, c.ID, c.Name, c.Description, string(skillIDsJSON), c.UseCase, string(tagsJSON))
	return err
}

func (s *SkillStore) ListCombinations(tag string, search string) ([]model.SkillCombination, error) {
	query := "SELECT id, name, description, skill_ids, use_case, tags, usage_count, avg_rating, created_at FROM skill_combinations WHERE 1=1"
	var args []interface{}

	if tag != "" {
		query += " AND tags LIKE ?"
		args = append(args, "%"+tag+"%")
	}
	if search != "" {
		query += " AND (name LIKE ? OR description LIKE ? OR use_case LIKE ?)"
		q := "%" + search + "%"
		args = append(args, q, q, q)
	}
	query += " ORDER BY usage_count DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var combos []model.SkillCombination
	for rows.Next() {
		var c model.SkillCombination
		var skillIDsJSON, tagsJSON string
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &skillIDsJSON, &c.UseCase, &tagsJSON, &c.UsageCount, &c.AvgRating, &c.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(skillIDsJSON), &c.SkillIDs)
		json.Unmarshal([]byte(tagsJSON), &c.Tags)
		combos = append(combos, c)
	}
	return combos, rows.Err()
}

func (s *SkillStore) GetCombination(id string) (*model.SkillCombination, error) {
	var c model.SkillCombination
	var skillIDsJSON, tagsJSON string
	err := s.db.QueryRow(
		"SELECT id, name, description, skill_ids, use_case, tags, usage_count, avg_rating, created_at FROM skill_combinations WHERE id = ?", id,
	).Scan(&c.ID, &c.Name, &c.Description, &skillIDsJSON, &c.UseCase, &tagsJSON, &c.UsageCount, &c.AvgRating, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(skillIDsJSON), &c.SkillIDs)
	json.Unmarshal([]byte(tagsJSON), &c.Tags)
	return &c, nil
}

func (s *SkillStore) UpsertChain(c *model.SkillChain) error {
	stepsJSON, _ := json.Marshal(c.Steps)
	_, err := s.db.Exec(`
		INSERT INTO skill_chains (id, name, description, steps, use_case)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			steps = excluded.steps,
			use_case = excluded.use_case
	`, c.ID, c.Name, c.Description, string(stepsJSON), c.UseCase)
	return err
}

func (s *SkillStore) ListChains() ([]model.SkillChain, error) {
	rows, err := s.db.Query(
		"SELECT id, name, description, steps, use_case, usage_count, success_rate, created_at FROM skill_chains ORDER BY usage_count DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chains []model.SkillChain
	for rows.Next() {
		var c model.SkillChain
		var stepsJSON string
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &stepsJSON, &c.UseCase, &c.UsageCount, &c.SuccessRate, &c.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(stepsJSON), &c.Steps)
		chains = append(chains, c)
	}
	return chains, rows.Err()
}

func (s *SkillStore) GetChain(id string) (*model.SkillChain, error) {
	var c model.SkillChain
	var stepsJSON string
	err := s.db.QueryRow(
		"SELECT id, name, description, steps, use_case, usage_count, success_rate, created_at FROM skill_chains WHERE id = ?", id,
	).Scan(&c.ID, &c.Name, &c.Description, &stepsJSON, &c.UseCase, &c.UsageCount, &c.SuccessRate, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(stepsJSON), &c.Steps)
	return &c, nil
}

func (s *SkillStore) BatchInsertUsage(logs []model.UsageLog) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, l := range logs {
		_, err := tx.Exec(`
			INSERT INTO skill_usage_log (id, client_id, session_id, skill_id, chain_id, context, project_type, invoked_at, duration_ms, success, feedback)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, l.ID, l.ClientID, l.SessionID, l.SkillID, l.ChainID, l.Context, l.ProjectType, l.InvokedAt, l.DurationMs, boolToInt(l.Success), l.Feedback)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SkillStore) GetUsageStats(skillID string) (UsageStats, error) {
	var stats UsageStats
	err := s.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END), 0), COALESCE(SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END), 0)
		 FROM skill_usage_log WHERE skill_id = ?`, skillID,
	).Scan(&stats.TotalUses, &stats.Successes, &stats.Failures)
	return stats, err
}

func (s *SkillStore) UpsertErrorCluster(c *model.ErrorCluster) error {
	_, err := s.db.Exec(`
		INSERT INTO error_clusters (id, error_type, fingerprint, total_count, sessions, top_skill, resolution_rate, common_root_cause, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(fingerprint) DO UPDATE SET
			total_count = excluded.total_count,
			sessions = excluded.sessions,
			top_skill = excluded.top_skill,
			resolution_rate = excluded.resolution_rate,
			common_root_cause = excluded.common_root_cause,
			updated_at = CURRENT_TIMESTAMP
	`, c.ID, c.ErrorType, c.Fingerprint, c.TotalCount, c.Sessions, c.TopSkill, c.ResolutionRate, c.CommonRootCause)
	return err
}

func (s *SkillStore) ListErrorClusters() ([]model.ErrorCluster, error) {
	rows, err := s.db.Query(
		"SELECT id, error_type, fingerprint, total_count, sessions, top_skill, resolution_rate, common_root_cause, updated_at FROM error_clusters ORDER BY total_count DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []model.ErrorCluster
	for rows.Next() {
		var c model.ErrorCluster
		if err := rows.Scan(&c.ID, &c.ErrorType, &c.Fingerprint, &c.TotalCount, &c.Sessions, &c.TopSkill, &c.ResolutionRate, &c.CommonRootCause, &c.UpdatedAt); err != nil {
			return nil, err
		}
		clusters = append(clusters, c)
	}
	return clusters, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}