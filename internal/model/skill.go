package model

import (
	"errors"
	"time"
)

type Skill struct {
	ID          string    `json:"id"`
	Tool        string    `json:"tool"`
	Category    string    `json:"category"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Difficulty  string    `json:"difficulty"`
	UsageCount  int       `json:"usage_count"`
	AvgRating   float64   `json:"avg_rating"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s Skill) Validate() error {
	if s.Name == "" {
		return errors.New("name is required")
	}
	if s.Tool == "" {
		return errors.New("tool is required")
	}
	if s.Category == "" {
		return errors.New("category is required")
	}
	return nil
}

func (s *Skill) UpdateStats(deltaUsage int, success bool, rating float64) {
	s.UsageCount += deltaUsage
	if rating > 0 {
		if s.AvgRating == 0 {
			s.AvgRating = rating
		} else {
			s.AvgRating = (s.AvgRating*float64(s.UsageCount-deltaUsage) + rating*float64(deltaUsage)) / float64(s.UsageCount)
		}
	}
	s.UpdatedAt = time.Now()
}

type ExampleType string

const (
	ExampleTypeUsage   ExampleType = "usage"
	ExampleTypeSuccess ExampleType = "success"
	ExampleTypeFailure ExampleType = "failure"
)

var validExampleTypes = map[ExampleType]bool{
	ExampleTypeUsage:   true,
	ExampleTypeSuccess: true,
	ExampleTypeFailure: true,
}

func (et ExampleType) String() string {
	return string(et)
}

func (et ExampleType) Valid() bool {
	return validExampleTypes[et]
}

type SkillExample struct {
	ID        string      `json:"id"`
	SkillID   string      `json:"skill_id"`
	Type      ExampleType `json:"type"`
	Title     string      `json:"title"`
	Scenario  string      `json:"scenario"`
	Prompt    string      `json:"prompt"`
	Result    string      `json:"result"`
	Lesson    string      `json:"lesson"`
	Rating    int         `json:"rating"`
	ClientID  string      `json:"client_id"`
	CreatedAt time.Time   `json:"created_at"`
}

func (e SkillExample) Validate() error {
	if !e.Type.Valid() {
		return errors.New("invalid example type: must be usage, success, or failure")
	}
	if e.Rating < 0 || e.Rating > 5 {
		return errors.New("rating must be between 0 and 5")
	}
	return nil
}

type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

func (es ErrorSeverity) String() string {
	return string(es)
}

type ErrorCase struct {
	ID               string      `json:"id"`
	SessionID        string      `json:"session_id"`
	ErrorType        string      `json:"error_type"`
	ErrorFingerprint string      `json:"error_fingerprint"`
	ErrorCount       int         `json:"error_count"`
	RootCause        string      `json:"root_cause"`
	SkillID          string      `json:"skill_id"`
	Resolution       string      `json:"resolution"`
	FileDiff         string      `json:"file_diff"`
	Lesson           string      `json:"lesson"`
	Tags             []string    `json:"tags"`
	Severity         ErrorSeverity `json:"severity"`
	ClientID         string      `json:"client_id"`
	CreatedAt        time.Time   `json:"created_at"`
}

func (e ErrorCase) Validate() error {
	if e.ErrorType == "" {
		return errors.New("error_type is required")
	}
	return nil
}

type ErrorCluster struct {
	ID              string  `json:"id"`
	ErrorType       string  `json:"error_type"`
	Fingerprint     string  `json:"fingerprint"`
	TotalCount      int     `json:"total_count"`
	Sessions        int     `json:"sessions"`
	TopSkill        string  `json:"top_skill"`
	ResolutionRate  float64 `json:"resolution_rate"`
	CommonRootCause string  `json:"common_root_cause"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (c *ErrorCluster) Increment(errors int, resolved bool) {
	c.TotalCount += errors
	c.Sessions++
	if resolved {
		c.ResolutionRate = float64(c.TotalCount-errors+errors) / float64(c.TotalCount)
	} else {
		if c.TotalCount > 0 {
			c.ResolutionRate = float64(c.TotalCount-errors) / float64(c.TotalCount)
		}
	}
	c.UpdatedAt = time.Now()
}

type SkillCombination struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SkillIDs    []string  `json:"skill_ids"`
	UseCase     string    `json:"use_case"`
	Tags        []string  `json:"tags"`
	UsageCount  int       `json:"usage_count"`
	AvgRating   float64   `json:"avg_rating"`
	CreatedAt   time.Time `json:"created_at"`
}

func (c SkillCombination) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if len(c.SkillIDs) < 2 {
		return errors.New("at least 2 skills required for a combination")
	}
	return nil
}

type ChainStep struct {
	Order        int    `json:"order"`
	SkillID      string `json:"skill_id"`
	Trigger      string `json:"trigger"`
	OutputToNext string `json:"output_to_next"`
}

type SkillChain struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Steps       []ChainStep `json:"steps"`
	UseCase     string      `json:"use_case"`
	UsageCount  int         `json:"usage_count"`
	SuccessRate float64     `json:"success_rate"`
	CreatedAt   time.Time   `json:"created_at"`
}

func (c SkillChain) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if len(c.Steps) == 0 {
		return errors.New("at least one step is required")
	}
	return nil
}

type UsageLog struct {
	ID          string    `json:"id"`
	ClientID    string    `json:"client_id"`
	SessionID   string    `json:"session_id"`
	SkillID     string    `json:"skill_id"`
	ChainID     string    `json:"chain_id"`
	Context     string    `json:"context"`
	ProjectType string    `json:"project_type"`
	InvokedAt   time.Time `json:"invoked_at"`
	DurationMs  int64     `json:"duration_ms"`
	Success     bool      `json:"success"`
	Feedback    string    `json:"feedback"`
}

func (u UsageLog) Validate() error {
	if u.ClientID == "" {
		return errors.New("client_id is required")
	}
	if u.SkillID == "" {
		return errors.New("skill_id is required")
	}
	return nil
}

type SkillRecommendation struct {
	ToSkillID string  `json:"to_skill_id"`
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`
	BasedOn   string  `json:"based_on"`
}

type AssociationRule struct {
	Antecedent string  `json:"antecedent"`
	Consequent string  `json:"consequent"`
	Confidence float64 `json:"confidence"`
	Support    int     `json:"support"`
}