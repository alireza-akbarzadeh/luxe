package collection

import (
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const (
	maxRuleGroups     = 8
	maxRuleConditions = 32
)

var allowedRuleFields = map[string]map[string]struct{}{
	"category_id": {"eq": {}, "neq": {}, "in": {}},
	"brand_id":    {"eq": {}, "neq": {}, "in": {}},
	"min_price":   {"eq": {}, "gte": {}, "lte": {}},
	"max_price":   {"eq": {}, "gte": {}, "lte": {}},
	"min_rating":  {"eq": {}, "gte": {}, "lte": {}},
	"is_new":      {"eq": {}},
	"in_stock":    {"eq": {}},
	"on_sale":     {"eq": {}},
	"search":      {"eq": {}, "contains": {}},
	"ids":         {"in": {}, "eq": {}},
	"sort":        {"eq": {}},
}

func countRuleConditions(rules *dto.CollectionRules) int {
	if rules == nil {
		return 0
	}
	n := len(rules.Conditions)
	for _, group := range rules.Groups {
		n += countGroupConditions(group)
	}
	return n
}

func countGroupConditions(group dto.CollectionRuleGroup) int {
	n := len(group.Conditions)
	for _, nested := range group.Groups {
		n += countGroupConditions(nested)
	}
	return n
}

func countRuleGroups(rules *dto.CollectionRules) int {
	if rules == nil {
		return 0
	}
	n := len(rules.Groups)
	for _, group := range rules.Groups {
		n += countNestedGroups(group)
	}
	return n
}

func countNestedGroups(group dto.CollectionRuleGroup) int {
	n := len(group.Groups)
	for _, nested := range group.Groups {
		n += countNestedGroups(nested)
	}
	return n
}

// ValidateCollectionRules checks mode/rules constraints before save or preview.
func ValidateCollectionRules(mode string, rules *dto.CollectionRules) error {
	normalized := normalizeMode(mode, "")
	conditionCount := countRuleConditions(rules)
	if (normalized == "dynamic" || normalized == "hybrid") && conditionCount == 0 {
		return utils.ErrBadRequest("dynamic and hybrid collections require at least one rule condition")
	}
	if conditionCount > maxRuleConditions {
		return utils.ErrBadRequest(fmt.Sprintf("rules may have at most %d conditions", maxRuleConditions))
	}
	if countRuleGroups(rules) > maxRuleGroups {
		return utils.ErrBadRequest(fmt.Sprintf("rules may have at most %d groups", maxRuleGroups))
	}
	if rules == nil {
		return nil
	}
	op := strings.ToLower(strings.TrimSpace(rules.Operator))
	if op == "" {
		op = "and"
	}
	if op != "and" && op != "or" {
		return utils.ErrBadRequest("rules operator must be and or or")
	}
	for _, condition := range rules.Conditions {
		if err := validateRuleCondition(condition); err != nil {
			return err
		}
	}
	for _, group := range rules.Groups {
		if err := validateRuleGroup(group, 0); err != nil {
			return err
		}
	}
	return nil
}

func validateRuleGroup(group dto.CollectionRuleGroup, depth int) error {
	if depth > 1 {
		return utils.ErrBadRequest("rules support at most one nested group level")
	}
	op := strings.ToLower(strings.TrimSpace(group.Operator))
	if op == "" {
		op = "and"
	}
	if op != "and" && op != "or" {
		return utils.ErrBadRequest("group operator must be and or or")
	}
	if len(group.Conditions) == 0 && len(group.Groups) == 0 {
		return utils.ErrBadRequest("rule groups must include at least one condition")
	}
	for _, condition := range group.Conditions {
		if err := validateRuleCondition(condition); err != nil {
			return err
		}
	}
	for _, nested := range group.Groups {
		if err := validateRuleGroup(nested, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func validateRuleCondition(condition dto.CollectionRuleCondition) error {
	field := strings.TrimSpace(condition.Field)
	if field == "" {
		return utils.ErrBadRequest("rule condition field is required")
	}
	ops, ok := allowedRuleFields[field]
	if !ok {
		return utils.ErrBadRequest(fmt.Sprintf("unsupported rule field: %s", field))
	}
	operator := strings.ToLower(strings.TrimSpace(condition.Operator))
	if operator == "" {
		operator = "eq"
	}
	if _, allowed := ops[operator]; !allowed {
		return utils.ErrBadRequest(fmt.Sprintf("unsupported operator %s for field %s", operator, field))
	}
	if condition.Value == nil {
		return utils.ErrBadRequest(fmt.Sprintf("rule condition value is required for field %s", field))
	}
	switch field {
	case "category_id", "brand_id":
		if operator == "in" {
			if _, ok := toUintSlice(condition.Value); !ok {
				return utils.ErrBadRequest(fmt.Sprintf("%s in requires an array of ids", field))
			}
			return nil
		}
		if _, ok := toUint(condition.Value); !ok {
			return utils.ErrBadRequest(fmt.Sprintf("%s requires a numeric id", field))
		}
	case "min_price", "max_price", "min_rating":
		if _, ok := toFloat(condition.Value); !ok {
			return utils.ErrBadRequest(fmt.Sprintf("%s requires a number", field))
		}
	case "is_new", "in_stock", "on_sale":
		if _, ok := toBool(condition.Value); !ok {
			return utils.ErrBadRequest(fmt.Sprintf("%s requires a boolean", field))
		}
	case "search", "sort":
		if _, ok := condition.Value.(string); !ok {
			return utils.ErrBadRequest(fmt.Sprintf("%s requires a string", field))
		}
	case "ids":
		if _, ok := toUintSlice(condition.Value); !ok {
			if _, ok := toUint(condition.Value); !ok {
				return utils.ErrBadRequest("ids requires an id or array of ids")
			}
		}
	}
	return nil
}

// normalizePublishStatus maps active/scheduled + future starts_at to scheduled.
func normalizePublishStatus(status string, startsAt *time.Time) string {
	if status == "" {
		status = "draft"
	}
	if status != "active" && status != "scheduled" {
		return status
	}
	if startsAt != nil && startsAt.After(time.Now().UTC()) {
		return "scheduled"
	}
	if status == "scheduled" && (startsAt == nil || !startsAt.After(time.Now().UTC())) {
		return "active"
	}
	return status
}

// expandRuleBranches flattens nested AND/OR rules into AND-only ProductListFilters branches.
func expandRuleBranches(rules *dto.CollectionRules) ([]dto.ProductListFilters, error) {
	if rules == nil {
		return []dto.ProductListFilters{{Status: "active"}}, nil
	}
	base := dto.ProductListFilters{Status: "active"}
	op := strings.ToLower(strings.TrimSpace(rules.Operator))
	if op == "" {
		op = "and"
	}

	leafSets := make([][]dto.CollectionRuleCondition, 0)
	rootConditions := append([]dto.CollectionRuleCondition(nil), rules.Conditions...)

	if len(rules.Groups) == 0 {
		if op == "or" && len(rootConditions) > 1 {
			for _, condition := range rootConditions {
				leafSets = append(leafSets, []dto.CollectionRuleCondition{condition})
			}
		} else {
			leafSets = append(leafSets, rootConditions)
		}
	} else if op == "and" {
		groupBranches, err := expandGroupsAND(rules.Groups)
		if err != nil {
			return nil, err
		}
		for _, branch := range groupBranches {
			merged := append(append([]dto.CollectionRuleCondition(nil), rootConditions...), branch...)
			leafSets = append(leafSets, merged)
		}
	} else {
		// OR at root: each root condition and each group expansion is a branch.
		for _, condition := range rootConditions {
			leafSets = append(leafSets, []dto.CollectionRuleCondition{condition})
		}
		for _, group := range rules.Groups {
			branches, err := expandGroup(group)
			if err != nil {
				return nil, err
			}
			leafSets = append(leafSets, branches...)
		}
		if len(leafSets) == 0 {
			leafSets = append(leafSets, nil)
		}
	}

	if len(leafSets) > maxRuleGroups {
		return nil, utils.ErrBadRequest(fmt.Sprintf("OR expansion exceeds %d branches", maxRuleGroups))
	}

	out := make([]dto.ProductListFilters, 0, len(leafSets))
	for _, conditions := range leafSets {
		filters := base
		for _, condition := range conditions {
			if err := applyRuleCondition(&filters, condition); err != nil {
				return nil, err
			}
		}
		out = append(out, filters)
	}
	if len(out) == 0 {
		out = append(out, base)
	}
	return out, nil
}

func expandGroupsAND(groups []dto.CollectionRuleGroup) ([][]dto.CollectionRuleCondition, error) {
	result := [][]dto.CollectionRuleCondition{{}}
	for _, group := range groups {
		groupBranches, err := expandGroup(group)
		if err != nil {
			return nil, err
		}
		next := make([][]dto.CollectionRuleCondition, 0, len(result)*len(groupBranches))
		for _, prefix := range result {
			for _, branch := range groupBranches {
				merged := append(append([]dto.CollectionRuleCondition(nil), prefix...), branch...)
				next = append(next, merged)
			}
		}
		result = next
		if len(result) > maxRuleGroups {
			return nil, utils.ErrBadRequest(fmt.Sprintf("AND/OR expansion exceeds %d branches", maxRuleGroups))
		}
	}
	return result, nil
}

func expandGroup(group dto.CollectionRuleGroup) ([][]dto.CollectionRuleCondition, error) {
	op := strings.ToLower(strings.TrimSpace(group.Operator))
	if op == "" {
		op = "and"
	}
	conditions := append([]dto.CollectionRuleCondition(nil), group.Conditions...)

	if len(group.Groups) == 0 {
		if op == "or" && len(conditions) > 1 {
			branches := make([][]dto.CollectionRuleCondition, 0, len(conditions))
			for _, condition := range conditions {
				branches = append(branches, []dto.CollectionRuleCondition{condition})
			}
			return branches, nil
		}
		return [][]dto.CollectionRuleCondition{conditions}, nil
	}

	nestedBranches, err := expandGroupsAND(group.Groups)
	if err != nil {
		return nil, err
	}
	if op == "and" {
		out := make([][]dto.CollectionRuleCondition, 0, len(nestedBranches))
		for _, nested := range nestedBranches {
			merged := append(append([]dto.CollectionRuleCondition(nil), conditions...), nested...)
			out = append(out, merged)
		}
		return out, nil
	}

	out := make([][]dto.CollectionRuleCondition, 0)
	for _, condition := range conditions {
		out = append(out, []dto.CollectionRuleCondition{condition})
	}
	out = append(out, nestedBranches...)
	return out, nil
}
