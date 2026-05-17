package engine

import (
	"testing"

	"skill-manage/internal/model"
)

type recomTestCase struct {
	name   string
	skillID string
	logs   []model.UsageLog
	want   []string
}

func TestRecommendByCoOccurrence(t *testing.T) {
	logs := []model.UsageLog{
		{SessionID: "s1", SkillID: "java-coding"},
		{SessionID: "s1", SkillID: "jpa-patterns"},
		{SessionID: "s1", SkillID: "springboot-security"},

		{SessionID: "s2", SkillID: "java-coding"},
		{SessionID: "s2", SkillID: "jpa-patterns"},

		{SessionID: "s3", SkillID: "tdd-workflow"},
		{SessionID: "s3", SkillID: "verification-loop"},
		{SessionID: "s3", SkillID: "java-coding"},

		{SessionID: "s4", SkillID: "jpa-patterns"},
		{SessionID: "s4", SkillID: "springboot-security"},
	}

	engine := NewRecommendEngine()

	recommended := engine.Recommend("java-coding", logs, 5)
	if len(recommended) == 0 {
		t.Fatal("expected recommendations")
	}

	top := recommended[0].ToSkillID
	if top != "jpa-patterns" {
		t.Errorf("top recommendation should be jpa-patterns, got %s", top)
	}
}

func TestRecommendNoData(t *testing.T) {
	engine := NewRecommendEngine()
	result := engine.Recommend("unknown-skill", []model.UsageLog{}, 3)
	if len(result) != 0 {
		t.Error("expected no recommendations for unknown skill")
	}
}

func TestAssociationRules(t *testing.T) {
	logs := []model.UsageLog{
		{SessionID: "s1", SkillID: "A"},
		{SessionID: "s1", SkillID: "B"},
		{SessionID: "s1", SkillID: "C"},

		{SessionID: "s2", SkillID: "A"},
		{SessionID: "s2", SkillID: "B"},

		{SessionID: "s3", SkillID: "A"},
		{SessionID: "s3", SkillID: "C"},

		{SessionID: "s4", SkillID: "D"},
		{SessionID: "s4", SkillID: "E"},
	}

	engine := NewRecommendEngine()
	sessions := groupBySession(logs)
	rules := engine.MineRules(sessions, 0.5, 2)

	found := false
	for _, r := range rules {
		if r.Antecedent == "A" && r.Consequent == "B" {
			found = true
			if r.Confidence < 0.6 {
				t.Errorf("A→B confidence should be >= 0.6, got %.2f", r.Confidence)
			}
		}
	}
	if !found {
		t.Error("expected rule A→B to be found")
	}
}

func TestRecommendByTags(t *testing.T) {
	skills := []model.Skill{
		{ID: "java-coding", Tags: []string{"java", "style", "convention"}},
		{ID: "jpa-patterns", Tags: []string{"java", "jpa", "performance"}},
		{ID: "springboot-security", Tags: []string{"java", "security", "spring"}},
		{ID: "typescript-rules", Tags: []string{"typescript", "style"}},
		{ID: "python-django", Tags: []string{"python", "web"}},
	}

	engine := NewRecommendEngine()
	result := engine.RecommendByTags("java-coding", skills, 3)

	if len(result) == 0 {
		t.Fatal("expected tag-based recommendations")
	}

	for _, r := range result {
		if r.ToSkillID == "java-coding" {
			t.Error("should not recommend the same skill")
		}
		if r.ToSkillID == "python-django" {
			t.Error("python-django should not be recommended for java skill")
		}
	}
}

func TestChainBasedRecommend(t *testing.T) {
	chains := []model.SkillChain{
		{
			ID:   "chain-1",
			Steps: []model.ChainStep{
				{Order: 1, SkillID: "tdd-workflow"},
				{Order: 2, SkillID: "code-review"},
				{Order: 3, SkillID: "verification-loop"},
			},
		},
		{
			ID:   "chain-2",
			Steps: []model.ChainStep{
				{Order: 1, SkillID: "java-coding"},
				{Order: 2, SkillID: "jpa-patterns"},
			},
		},
	}

	engine := NewRecommendEngine()
	result := engine.RecommendByChains("tdd-workflow", chains)

	if len(result) == 0 {
		t.Fatal("expected chain-based recommendations")
	}
	if result[0].ToSkillID != "code-review" {
		t.Errorf("next should be code-review, got %s", result[0].ToSkillID)
	}
}

func TestMergeRecommendations(t *testing.T) {
	coRec := []model.SkillRecommendation{
		{ToSkillID: "A", Score: 0.8, BasedOn: "co-occurrence"},
		{ToSkillID: "B", Score: 0.6, BasedOn: "co-occurrence"},
	}
	tagRec := []model.SkillRecommendation{
		{ToSkillID: "B", Score: 0.9, BasedOn: "tags"},
		{ToSkillID: "C", Score: 0.4, BasedOn: "tags"},
	}
	chainRec := []model.SkillRecommendation{
		{ToSkillID: "A", Score: 0.5, BasedOn: "chain"},
	}

	engine := NewRecommendEngine()
	merged := engine.Merge(coRec, tagRec, chainRec)

	ids := make(map[string]bool)
	for _, m := range merged {
		ids[m.ToSkillID] = true
	}
	if !ids["A"] || !ids["B"] || !ids["C"] {
		t.Error("merged result should contain A, B, C")
	}
	if len(merged) != 3 {
		t.Errorf("expected 3 unique recommendations, got %d", len(merged))
	}

	if merged[0].ToSkillID != "B" {
		t.Errorf("top should be B (0.9+0.6=1.5), got %s", merged[0].ToSkillID)
	}
}

func TestSortByScore(t *testing.T) {
	recs := []model.SkillRecommendation{
		{ToSkillID: "C", Score: 0.3},
		{ToSkillID: "A", Score: 0.9},
		{ToSkillID: "B", Score: 0.5},
	}

	sortByScore(recs)
	if recs[0].ToSkillID != "A" {
		t.Errorf("first should be A, got %s", recs[0].ToSkillID)
	}
	if recs[1].ToSkillID != "B" {
		t.Errorf("second should be B, got %s", recs[1].ToSkillID)
	}
	if recs[2].ToSkillID != "C" {
		t.Errorf("third should be C, got %s", recs[2].ToSkillID)
	}
}