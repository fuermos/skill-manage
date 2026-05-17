package engine

import (
	"sort"

	"skill-manage/internal/model"
)

type RecommendEngine struct {
	coOccurWeight float64
	tagWeight     float64
	chainWeight   float64
}

func NewRecommendEngine() *RecommendEngine {
	return &RecommendEngine{
		coOccurWeight: 0.4,
		tagWeight:     0.3,
		chainWeight:   0.3,
	}
}

func (e *RecommendEngine) Recommend(skillID string, logs []model.UsageLog, limit int) []model.SkillRecommendation {
	sessions := groupBySession(logs)
	coRec := e.recommendByCoOccur(skillID, sessions)
	tagRec := e.RecommendByTags(skillID, nil, limit)
	return e.Merge(coRec, tagRec, nil)[:min(limit, len(coRec)+len(tagRec))]
}

func (e *RecommendEngine) recommendByCoOccur(skillID string, sessions map[string][]model.UsageLog) []model.SkillRecommendation {
	counts := make(map[string]int)
	total := 0

	for _, sessionSkills := range sessions {
		hasTarget := false
		for _, s := range sessionSkills {
			if s.SkillID == skillID {
				hasTarget = true
				break
			}
		}
		if !hasTarget {
			continue
		}
		total++
		for _, s := range sessionSkills {
			if s.SkillID != skillID {
				counts[s.SkillID]++
			}
		}
	}

	var recs []model.SkillRecommendation
	for id, count := range counts {
		confidence := float64(count) / float64(total)
		recs = append(recs, model.SkillRecommendation{
			ToSkillID: id,
			Score:     e.coOccurWeight * confidence,
			Reason:    "frequently used together",
			BasedOn:   "co-occurrence",
		})
	}

	sortByScore(recs)
	return recs
}

func (e *RecommendEngine) RecommendByTags(skillID string, allSkills []model.Skill, limit int) []model.SkillRecommendation {
	var targetTags []string
	for _, s := range allSkills {
		if s.ID == skillID {
			targetTags = s.Tags
			break
		}
	}
	if len(targetTags) == 0 {
		return nil
	}

	tagSet := make(map[string]bool)
	for _, t := range targetTags {
		tagSet[t] = true
	}

	type scored struct {
		id    string
		score float64
	}
	var scores []scored
	for _, s := range allSkills {
		if s.ID == skillID {
			continue
		}
		common := 0
		for _, t := range s.Tags {
			if tagSet[t] {
				common++
			}
		}
		if common == 0 {
			continue
		}
		sim := float64(common) / float64(len(targetTags)+len(s.Tags)-common)
		scores = append(scores, scored{id: s.ID, score: sim})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	var recs []model.SkillRecommendation
	for i, sc := range scores {
		if limit > 0 && i >= limit {
			break
		}
		recs = append(recs, model.SkillRecommendation{
			ToSkillID: sc.id,
			Score:     e.tagWeight * sc.score,
			Reason:    "similar tags",
			BasedOn:   "tags",
		})
	}
	return recs
}

func (e *RecommendEngine) RecommendByChains(skillID string, chains []model.SkillChain) []model.SkillRecommendation {
	var recs []model.SkillRecommendation

	for _, chain := range chains {
		for i, step := range chain.Steps {
			if step.SkillID == skillID && i+1 < len(chain.Steps) {
				next := chain.Steps[i+1]
				recs = append(recs, model.SkillRecommendation{
					ToSkillID: next.SkillID,
					Score:     e.chainWeight * float64(chain.UsageCount+1) / float64(len(chain.Steps)),
					Reason:    "next step in chain: " + chain.Name,
					BasedOn:   "chain",
				})
			}
		}
	}

	sortByScore(recs)
	return recs
}

func (e *RecommendEngine) MineRules(sessions map[string][]model.UsageLog, minConfidence float64, minSupport int) []model.AssociationRule {
	type pair struct{ a, b string }
	counts := make(map[pair]int)
	singleCounts := make(map[string]int)

	for _, skills := range sessions {
		seen := make(map[string]bool)
		for _, s := range skills {
			singleCounts[s.SkillID]++
			seen[s.SkillID] = true
		}
		ids := make([]string, 0, len(seen))
		for id := range seen {
			ids = append(ids, id)
		}
		for i := 0; i < len(ids); i++ {
			for j := i + 1; j < len(ids); j++ {
				counts[pair{ids[i], ids[j]}]++
				counts[pair{ids[j], ids[i]}]++
			}
		}
	}

	var rules []model.AssociationRule
	for p, pairCount := range counts {
		conf := float64(pairCount) / float64(singleCounts[p.a])
		if conf >= minConfidence && pairCount >= minSupport {
			rules = append(rules, model.AssociationRule{
				Antecedent: p.a,
				Consequent: p.b,
				Confidence: conf,
				Support:    pairCount,
			})
		}
	}

	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Confidence > rules[j].Confidence
	})
	return rules
}

func (e *RecommendEngine) Merge(coRec, tagRec, chainRec []model.SkillRecommendation) []model.SkillRecommendation {
	merged := make(map[string]float64)
	reasons := make(map[string]string)
	bases := make(map[string]string)

	add := func(recs []model.SkillRecommendation) {
		for _, r := range recs {
			merged[r.ToSkillID] += r.Score
			if existing, ok := reasons[r.ToSkillID]; ok {
				reasons[r.ToSkillID] = existing + "; " + r.Reason
			} else {
				reasons[r.ToSkillID] = r.Reason
			}
			if existing, ok := bases[r.ToSkillID]; ok {
				bases[r.ToSkillID] = existing + "+" + r.BasedOn
			} else {
				bases[r.ToSkillID] = r.BasedOn
			}
		}
	}

	add(coRec)
	add(tagRec)
	add(chainRec)

	var result []model.SkillRecommendation
	for id, score := range merged {
		result = append(result, model.SkillRecommendation{
			ToSkillID: id,
			Score:     score,
			Reason:    reasons[id],
			BasedOn:   bases[id],
		})
	}

	sortByScore(result)
	return result
}

func sortByScore(recs []model.SkillRecommendation) {
	sort.Slice(recs, func(i, j int) bool {
		return recs[i].Score > recs[j].Score
	})
}

func groupBySession(logs []model.UsageLog) map[string][]model.UsageLog {
	sessions := make(map[string][]model.UsageLog)
	for _, l := range logs {
		sessions[l.SessionID] = append(sessions[l.SessionID], l)
	}
	return sessions
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}