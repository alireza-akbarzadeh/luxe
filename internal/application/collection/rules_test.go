package collection

import (
	"testing"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
)

func TestValidateCollectionRulesRequiresConditionsForDynamic(t *testing.T) {
	err := ValidateCollectionRules("dynamic", nil)
	if err == nil {
		t.Fatal("expected error for empty dynamic rules")
	}
	err = ValidateCollectionRules("manual", nil)
	if err != nil {
		t.Fatalf("manual mode should allow empty rules: %v", err)
	}
}

func TestValidateCollectionRulesRejectsBadField(t *testing.T) {
	err := ValidateCollectionRules("dynamic", &dto.CollectionRules{
		Operator: "and",
		Conditions: []dto.CollectionRuleCondition{
			{Field: "not_a_field", Operator: "eq", Value: true},
		},
	})
	if err == nil {
		t.Fatal("expected unsupported field error")
	}
}

func TestExpandRuleBranchesAND(t *testing.T) {
	branches, err := expandRuleBranches(&dto.CollectionRules{
		Operator: "and",
		Conditions: []dto.CollectionRuleCondition{
			{Field: "is_new", Operator: "eq", Value: true},
			{Field: "on_sale", Operator: "eq", Value: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(branches))
	}
	if branches[0].IsNew == nil || !*branches[0].IsNew {
		t.Fatal("expected is_new true")
	}
	if branches[0].OnSale == nil || !*branches[0].OnSale {
		t.Fatal("expected on_sale true")
	}
}

func TestExpandRuleBranchesOR(t *testing.T) {
	branches, err := expandRuleBranches(&dto.CollectionRules{
		Operator: "or",
		Conditions: []dto.CollectionRuleCondition{
			{Field: "is_new", Operator: "eq", Value: true},
			{Field: "on_sale", Operator: "eq", Value: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}
}

func TestExpandRuleBranchesWithGroup(t *testing.T) {
	branches, err := expandRuleBranches(&dto.CollectionRules{
		Operator: "and",
		Conditions: []dto.CollectionRuleCondition{
			{Field: "in_stock", Operator: "eq", Value: true},
		},
		Groups: []dto.CollectionRuleGroup{
			{
				Operator: "or",
				Conditions: []dto.CollectionRuleCondition{
					{Field: "is_new", Operator: "eq", Value: true},
					{Field: "on_sale", Operator: "eq", Value: true},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches from AND root + OR group, got %d", len(branches))
	}
	for _, branch := range branches {
		if branch.InStock == nil || !*branch.InStock {
			t.Fatal("expected in_stock on every branch")
		}
	}
}

func TestNormalizePublishStatus(t *testing.T) {
	future := time.Now().UTC().Add(48 * time.Hour)
	if got := normalizePublishStatus("active", &future); got != "scheduled" {
		t.Fatalf("expected scheduled, got %s", got)
	}
	past := time.Now().UTC().Add(-48 * time.Hour)
	if got := normalizePublishStatus("scheduled", &past); got != "active" {
		t.Fatalf("expected active, got %s", got)
	}
	if got := normalizePublishStatus("draft", &future); got != "draft" {
		t.Fatalf("expected draft, got %s", got)
	}
}
