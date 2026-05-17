package model

import (
	"testing"
	"time"
)

func TestSkillValidate(t *testing.T) {
	valid := Skill{
		ID:       "claude-skills-java-coding",
		Tool:     "claude",
		Category: "skills",
		Name:     "java-coding-standards",
		Summary:  "Java coding conventions and best practices",
		Tags:     []string{"java", "style", "convention"},
	}

	if err := valid.Validate(); err != nil {
		t.Errorf("valid skill should not error: %v", err)
	}

	tests := []struct {
		name  string
		skill Skill
	}{
		{"empty name", Skill{ID: "x", Tool: "claude", Category: "skills"}},
		{"empty tool", Skill{ID: "x", Name: "test", Category: "skills"}},
		{"empty category", Skill{ID: "x", Tool: "claude", Name: "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.skill.Validate(); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestSkillExampleValidate(t *testing.T) {
	valid := SkillExample{
		ID:       "ex-001",
		SkillID:  "claude-skills-java-coding",
		Type:     ExampleTypeSuccess,
		Title:    "Fixed NPE with Optional",
		Scenario: "NullPointerException in UserService",
		Rating:   4,
	}

	if err := valid.Validate(); err != nil {
		t.Errorf("valid example should not error: %v", err)
	}

	invalidType := SkillExample{
		ID:      "ex-002",
		SkillID: "claude-skills-java-coding",
		Type:    ExampleType("invalid"),
		Title:   "test",
		Rating:  4,
	}
	if err := invalidType.Validate(); err == nil {
		t.Error("invalid type should error")
	}

	invalidRating := SkillExample{
		ID:      "ex-003",
		SkillID: "claude-skills-java-coding",
		Type:    ExampleTypeSuccess,
		Title:   "test",
		Rating:  6,
	}
	if err := invalidRating.Validate(); err == nil {
		t.Error("rating > 5 should error")
	}

	negativeRating := SkillExample{
		ID:      "ex-004",
		SkillID: "claude-skills-java-coding",
		Type:    ExampleTypeSuccess,
		Title:   "test",
		Rating:  -1,
	}
	if err := negativeRating.Validate(); err == nil {
		t.Error("rating < 0 should error")
	}
}

func TestExampleTypeString(t *testing.T) {
	tests := []struct {
		et   ExampleType
		want string
	}{
		{ExampleTypeUsage, "usage"},
		{ExampleTypeSuccess, "success"},
		{ExampleTypeFailure, "failure"},
		{ExampleType("invalid"), "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.et.String()
			if got != tt.want {
				t.Errorf("ExampleType(%q).String() = %q, want %q", tt.et, got, tt.want)
			}
		})
	}
}

func TestErrorCaseSeverity(t *testing.T) {
	c := ErrorCase{
		ID:        "err-001",
		SessionID: "sess-001",
		ErrorType: "NullPointerException",
		Severity:  SeverityHigh,
		ClientID:  "win-pc",
	}

	if c.Severity.String() != "high" {
		t.Errorf("Severity.String() = %q, want 'high'", c.Severity.String())
	}

	c.Severity = SeverityCritical
	if c.Severity.String() != "critical" {
		t.Errorf("Severity.String() = %q, want 'critical'", c.Severity.String())
	}
}

func TestErrorCaseValidate(t *testing.T) {
	c := ErrorCase{
		ID:        "err-001",
		SessionID: "sess-001",
		ErrorType: "NullPointerException",
		ClientID:  "win-pc",
	}
	if err := c.Validate(); err != nil {
		t.Errorf("valid error case should not error: %v", err)
	}

	noType := ErrorCase{ID: "err-002", ClientID: "win-pc"}
	if err := noType.Validate(); err == nil {
		t.Error("error case without error_type should error")
	}
}

func TestSkillCombinationValidate(t *testing.T) {
	c := SkillCombination{
		ID:       "combo-001",
		Name:     "Spring Boot Dev Stack",
		SkillIDs: []string{"java-coding", "jpa-patterns", "springboot-security"},
	}
	if err := c.Validate(); err != nil {
		t.Errorf("valid combination should not error: %v", err)
	}

	noName := SkillCombination{ID: "combo-002", SkillIDs: []string{"a"}}
	if err := noName.Validate(); err == nil {
		t.Error("combination without name should error")
	}

	noSkills := SkillCombination{ID: "combo-003", Name: "Empty"}
	if err := noSkills.Validate(); err == nil {
		t.Error("combination without skills should error")
	}

	singleSkill := SkillCombination{ID: "combo-004", Name: "Single", SkillIDs: []string{"one"}}
	if err := singleSkill.Validate(); err == nil {
		t.Error("combination with single skill should error")
	}
}

func TestSkillChainStep(t *testing.T) {
	step := ChainStep{
		Order:   1,
		SkillID: "tdd-workflow",
		Trigger: "Write new feature",
	}
	if step.Order != 1 {
		t.Errorf("Order = %d, want 1", step.Order)
	}
}

func TestSkillChainValidate(t *testing.T) {
	c := SkillChain{
		ID:   "chain-001",
		Name: "TDD Workflow",
		Steps: []ChainStep{
			{Order: 1, SkillID: "tdd-workflow", Trigger: "Write feature"},
			{Order: 2, SkillID: "code-review", Trigger: "After TDD cycle"},
		},
	}
	if err := c.Validate(); err != nil {
		t.Errorf("valid chain should not error: %v", err)
	}

	noName := SkillChain{ID: "chain-002", Steps: []ChainStep{{Order: 1}}}
	if err := noName.Validate(); err == nil {
		t.Error("chain without name should error")
	}

	emptySteps := SkillChain{ID: "chain-003", Name: "Empty"}
	if err := emptySteps.Validate(); err == nil {
		t.Error("chain without steps should error")
	}
}

func TestUsageLogValidate(t *testing.T) {
	ul := UsageLog{
		ID:        "usage-001",
		ClientID:  "win-pc",
		SessionID: "sess-001",
		SkillID:   "java-coding",
		Success:   true,
		InvokedAt: time.Now(),
	}
	if err := ul.Validate(); err != nil {
		t.Errorf("valid usage log should not error: %v", err)
	}

	noClient := UsageLog{
		ID:        "usage-002",
		SessionID: "sess-001",
		SkillID:   "java-coding",
		Success:   true,
	}
	if err := noClient.Validate(); err == nil {
		t.Error("usage log without client_id should error")
	}

	noSkill := UsageLog{
		ID:        "usage-003",
		ClientID:  "win-pc",
		SessionID: "sess-001",
		Success:   true,
	}
	if err := noSkill.Validate(); err == nil {
		t.Error("usage log without skill_id should error")
	}
}

func TestErrorClusterUpdate(t *testing.T) {
	cluster := ErrorCluster{
		ID:         "cluster-001",
		ErrorType:  "NullPointerException",
		Fingerprint: "npe-user-service",
		TotalCount: 10,
		Sessions:   5,
		TopSkill:   "java-coding-standards",
		ResolutionRate: 0.8,
	}

	cluster.Increment(2, true)
	if cluster.TotalCount != 12 {
		t.Errorf("TotalCount = %d, want 12", cluster.TotalCount)
	}
	if cluster.Sessions != 6 {
		t.Errorf("Sessions = %d, want 6", cluster.Sessions)
	}

	cluster.Increment(1, false)
	if cluster.ResolutionRate <= 0 {
		t.Error("resolution rate should be > 0 after successful resolutions")
	}
}

func TestSkillUpdateStats(t *testing.T) {
	s := Skill{
		ID:         "java-coding",
		UsageCount: 10,
		AvgRating:  4.0,
	}

	s.UpdateStats(1, true, 5.0)
	if s.UsageCount != 11 {
		t.Errorf("UsageCount = %d, want 11", s.UsageCount)
	}
	if s.AvgRating < 4.0 {
		t.Errorf("AvgRating should increase with high rating, got %f", s.AvgRating)
	}

	s.UpdateStats(1, false, 2.0)
	if s.UsageCount != 12 {
		t.Errorf("UsageCount = %d, want 12", s.UsageCount)
	}
}