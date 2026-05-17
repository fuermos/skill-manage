package store

import (
	"testing"
	"time"

	"skill-manage/internal/model"
)

func TestSkillCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSkillStore(db)

	s := model.Skill{
		ID:          "claude-skills-java-coding",
		Tool:        "claude",
		Category:    "skills",
		Name:        "java-coding-standards",
		DisplayName: "Java Coding Standards",
		Summary:     "Java naming and conventions",
		Description: "Comprehensive Java coding conventions",
		Tags:        []string{"java", "style"},
		Difficulty:  "intermediate",
	}

	if err := store.UpsertSkill(&s); err != nil {
		t.Fatalf("UpsertSkill: %v", err)
	}

	got, err := store.GetSkill(s.ID)
	if err != nil {
		t.Fatalf("GetSkill: %v", err)
	}
	if got == nil {
		t.Fatal("expected skill to exist")
	}
	if got.DisplayName != s.DisplayName {
		t.Errorf("DisplayName = %q, want %q", got.DisplayName, s.DisplayName)
	}
	if len(got.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(got.Tags))
	}

	// Update
	s.Summary = "Updated summary"
	if err := store.UpsertSkill(&s); err != nil {
		t.Fatalf("UpsertSkill update: %v", err)
	}
	got, _ = store.GetSkill(s.ID)
	if got.Summary != "Updated summary" {
		t.Errorf("Summary = %q, want 'Updated summary'", got.Summary)
	}
}

func TestSkillList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSkillStore(db)

	skills := []model.Skill{
		{ID: "s1", Tool: "claude", Category: "skills", Name: "java"},
		{ID: "s2", Tool: "claude", Category: "rules", Name: "security-rule"},
		{ID: "s3", Tool: "opencode", Category: "skills", Name: "typescript"},
	}
	for _, s := range skills {
		if err := store.UpsertSkill(&s); err != nil {
			t.Fatalf("UpsertSkill: %v", err)
		}
	}

	all, err := store.ListSkills("", "", "")
	if err != nil {
		t.Fatalf("ListSkills: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 skills, got %d", len(all))
	}

	claude, _ := store.ListSkills("claude", "", "")
	if len(claude) != 2 {
		t.Errorf("expected 2 claude skills, got %d", len(claude))
	}

	opencode, _ := store.ListSkills("opencode", "skills", "")
	if len(opencode) != 1 {
		t.Errorf("expected 1 opencode skill, got %d", len(opencode))
	}
}

func TestSkillExampleCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	skillStore := NewSkillStore(db)

	skillStore.UpsertSkill(&model.Skill{
		ID: "s1", Tool: "claude", Category: "skills", Name: "java",
	})

	ex := model.SkillExample{
		ID:       "ex-001",
		SkillID:  "s1",
		Type:     model.ExampleTypeSuccess,
		Title:    "Fixed NPE",
		Scenario: "NullPointerException in UserService",
		Result:   "Used Optional to handle null",
		Rating:   5,
		ClientID: "win-pc",
	}
	if err := skillStore.AddExample(&ex); err != nil {
		t.Fatalf("AddExample: %v", err)
	}

	examples, err := skillStore.ListExamples("s1", "")
	if err != nil {
		t.Fatalf("ListExamples: %v", err)
	}
	if len(examples) != 1 {
		t.Errorf("expected 1 example, got %d", len(examples))
	}
	if examples[0].Rating != 5 {
		t.Errorf("Rating = %d, want 5", examples[0].Rating)
	}

	successOnly, _ := skillStore.ListExamples("s1", model.ExampleTypeSuccess)
	if len(successOnly) != 1 {
		t.Errorf("expected 1 success example, got %d", len(successOnly))
	}

	failureOnly, _ := skillStore.ListExamples("s1", model.ExampleTypeFailure)
	if len(failureOnly) != 0 {
		t.Errorf("expected 0 failure examples, got %d", len(failureOnly))
	}
}

func TestErrorCaseCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSkillStore(db)

	ec := model.ErrorCase{
		ID:        "err-001",
		SessionID: "sess-001",
		ErrorType: "NullPointerException",
		ErrorFingerprint: "npe-user-service",
		ErrorCount: 3,
		RootCause: "Not using Optional",
		SkillID:   "java-coding",
		Resolution: "Replaced null checks with Optional",
		Tags:      []string{"java", "npe"},
		Severity:  model.SeverityHigh,
		ClientID:  "win-pc",
	}

	if err := store.AddErrorCase(&ec); err != nil {
		t.Fatalf("AddErrorCase: %v", err)
	}

	cases, err := store.ListErrorCases("", 10)
	if err != nil {
		t.Fatalf("ListErrorCases: %v", err)
	}
	if len(cases) != 1 {
		t.Errorf("expected 1 error case, got %d", len(cases))
	}

	cases, _ = store.ListErrorCases("NullPointerException", 10)
	if len(cases) != 1 {
		t.Errorf("expected 1 NPE case, got %d", len(cases))
	}

	cases, _ = store.ListErrorCases("SQLException", 10)
	if len(cases) != 0 {
		t.Errorf("expected 0 SQLException cases, got %d", len(cases))
	}
}

func TestSkillCombinationCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSkillStore(db)

	combo := model.SkillCombination{
		ID:          "combo-001",
		Name:        "Spring Boot Dev Stack",
		Description: "Recommended stack for Spring Boot development",
		SkillIDs:    []string{"java-coding", "jpa-patterns", "springboot-security"},
		UseCase:     "Building REST APIs",
		Tags:        []string{"spring", "backend"},
	}

	if err := store.UpsertCombination(&combo); err != nil {
		t.Fatalf("UpsertCombination: %v", err)
	}

	combos, err := store.ListCombinations("", "")
	if err != nil {
		t.Fatalf("ListCombinations: %v", err)
	}
	if len(combos) != 1 {
		t.Errorf("expected 1 combination, got %d", len(combos))
	}

	got, err := store.GetCombination(combo.ID)
	if err != nil {
		t.Fatalf("GetCombination: %v", err)
	}
	if got == nil {
		t.Fatal("expected combination to exist")
	}
	if len(got.SkillIDs) != 3 {
		t.Errorf("expected 3 skills, got %d", len(got.SkillIDs))
	}

	combos, _ = store.ListCombinations("backend", "")
	if len(combos) != 1 {
		t.Errorf("expected 1 combo with tag backend, got %d", len(combos))
	}
}

func TestSkillChainCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSkillStore(db)

	chain := model.SkillChain{
		ID:          "chain-001",
		Name:        "TDD Workflow",
		Description: "Test-driven development workflow",
		Steps: []model.ChainStep{
			{Order: 1, SkillID: "tdd-workflow", Trigger: "Write feature"},
			{Order: 2, SkillID: "code-review", Trigger: "After TDD cycle"},
			{Order: 3, SkillID: "verification-loop", Trigger: "After code review"},
		},
		UseCase: "Developing new features",
	}

	if err := store.UpsertChain(&chain); err != nil {
		t.Fatalf("UpsertChain: %v", err)
	}

	chains, err := store.ListChains()
	if err != nil {
		t.Fatalf("ListChains: %v", err)
	}
	if len(chains) != 1 {
		t.Errorf("expected 1 chain, got %d", len(chains))
	}

	got, err := store.GetChain(chain.ID)
	if err != nil {
		t.Fatalf("GetChain: %v", err)
	}
	if got == nil {
		t.Fatal("expected chain to exist")
	}
	if len(got.Steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(got.Steps))
	}
	if got.Steps[0].Order != 1 {
		t.Errorf("first step order = %d, want 1", got.Steps[0].Order)
	}
}

func TestUsageLogBatchInsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSkillStore(db)

	logs := []model.UsageLog{
		{ID: "u1", ClientID: "pc-1", SessionID: "s1", SkillID: "java", Success: true, InvokedAt: time.Now()},
		{ID: "u2", ClientID: "pc-1", SessionID: "s1", SkillID: "jpa", Success: true, InvokedAt: time.Now()},
		{ID: "u3", ClientID: "pc-1", SessionID: "s2", SkillID: "java", Success: false, InvokedAt: time.Now()},
	}

	if err := store.BatchInsertUsage(logs); err != nil {
		t.Fatalf("BatchInsertUsage: %v", err)
	}

	stats, err := store.GetUsageStats("java")
	if err != nil {
		t.Fatalf("GetUsageStats: %v", err)
	}
	if stats.TotalUses != 2 {
		t.Errorf("expected 2 total uses for java, got %d", stats.TotalUses)
	}
	if stats.Successes != 1 {
		t.Errorf("expected 1 success for java, got %d", stats.Successes)
	}
	if stats.Failures != 1 {
		t.Errorf("expected 1 failure for java, got %d", stats.Failures)
	}

	stats, _ = store.GetUsageStats("jpa")
	if stats.TotalUses != 1 {
		t.Errorf("expected 1 total use for jpa, got %d", stats.TotalUses)
	}

	stats, _ = store.GetUsageStats("nonexistent")
	if stats.TotalUses != 0 {
		t.Errorf("expected 0 uses for nonexistent, got %d", stats.TotalUses)
	}
}

func TestErrorClusterUpsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSkillStore(db)

	cluster := model.ErrorCluster{
		ID:               "cluster-001",
		ErrorType:        "NullPointerException",
		Fingerprint:      "npe-user-service",
		TotalCount:       5,
		Sessions:         3,
		TopSkill:         "java-coding-standards",
		ResolutionRate:   0.6,
		CommonRootCause:  "缺少Optional空值处理",
	}

	if err := store.UpsertErrorCluster(&cluster); err != nil {
		t.Fatalf("UpsertErrorCluster: %v", err)
	}

	clusters, err := store.ListErrorClusters()
	if err != nil {
		t.Fatalf("ListErrorClusters: %v", err)
	}
	if len(clusters) != 1 {
		t.Errorf("expected 1 cluster, got %d", len(clusters))
	}

	// Update the same cluster
	cluster.TotalCount = 10
	cluster.Sessions = 6
	cluster.ResolutionRate = 0.8
	if err := store.UpsertErrorCluster(&cluster); err != nil {
		t.Fatalf("UpsertErrorCluster update: %v", err)
	}

	clusters, _ = store.ListErrorClusters()
	if clusters[0].TotalCount != 10 {
		t.Errorf("TotalCount = %d, want 10", clusters[0].TotalCount)
	}
}

func TestGetSkillsByIDs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSkillStore(db)

	skills := []model.Skill{
		{ID: "a", Tool: "claude", Category: "skills", Name: "A"},
		{ID: "b", Tool: "claude", Category: "skills", Name: "B"},
		{ID: "c", Tool: "claude", Category: "skills", Name: "C"},
	}
	for _, s := range skills {
		store.UpsertSkill(&s)
	}

	result, err := store.GetSkillsByIDs([]string{"a", "c", "nonexistent"})
	if err != nil {
		t.Fatalf("GetSkillsByIDs: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 skills, got %d", len(result))
	}
}

func TestUpdateSkillStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSkillStore(db)

	s := model.Skill{ID: "s1", Tool: "claude", Category: "skills", Name: "test"}
	store.UpsertSkill(&s)

	if err := store.UpdateSkillUsage("s1", true, 4.5); err != nil {
		t.Fatalf("UpdateSkillUsage: %v", err)
	}

	got, _ := store.GetSkill("s1")
	if got.UsageCount != 1 {
		t.Errorf("UsageCount = %d, want 1", got.UsageCount)
	}
	if got.AvgRating != 4.5 {
		t.Errorf("AvgRating = %f, want 4.5", got.AvgRating)
	}

	store.UpdateSkillUsage("s1", false, 2.0)
	got, _ = store.GetSkill("s1")
	if got.UsageCount != 2 {
		t.Errorf("UsageCount = %d, want 2", got.UsageCount)
	}
}

func TestSearchSkills(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSkillStore(db)

	skills := []model.Skill{
		{ID: "s1", Tool: "claude", Category: "skills", Name: "java-coding", Summary: "Java conventions"},
		{ID: "s2", Tool: "claude", Category: "skills", Name: "jpa-patterns", Summary: "JPA best practices"},
		{ID: "s3", Tool: "claude", Category: "rules", Name: "security", Summary: "Security guidelines for Java"},
	}
	for _, s := range skills {
		store.UpsertSkill(&s)
	}

	result, err := store.SearchSkills("java", "", "")
	if err != nil {
		t.Fatalf("SearchSkills: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 results for 'java', got %d", len(result))
	}

	result, _ = store.SearchSkills("jpa", "", "")
	if len(result) != 1 {
		t.Errorf("expected 1 result for 'jpa', got %d", len(result))
	}

	result, _ = store.SearchSkills("nonexistent", "", "")
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}