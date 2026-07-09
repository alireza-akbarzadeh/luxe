package models

import (
	"encoding/json"

	"gorm.io/datatypes"
)

// CouponConditions stores optional eligibility rules for a promotion.
type CouponConditions struct {
	FirstOrderOnly  bool   `json:"first_order_only,omitempty"`
	CategoryIDs     []uint `json:"category_ids,omitempty"`
	ProductIDs      []uint `json:"product_ids,omitempty"`
	UserIDs         []uint `json:"user_ids,omitempty"`
	CustomerSegment string `json:"customer_segment,omitempty"`
	MinItemQuantity int    `json:"min_item_quantity,omitempty"`
}

// ParseCouponConditions unmarshals JSONB conditions from a coupon row.
func ParseCouponConditions(raw datatypes.JSON) CouponConditions {
	if len(raw) == 0 {
		return CouponConditions{}
	}
	var conditions CouponConditions
	_ = json.Unmarshal(raw, &conditions)
	return conditions
}

// CouponConditionsIsEmpty reports whether no eligibility rules are configured.
func CouponConditionsIsEmpty(conditions CouponConditions) bool {
	return !conditions.FirstOrderOnly &&
		len(conditions.CategoryIDs) == 0 &&
		len(conditions.ProductIDs) == 0 &&
		len(conditions.UserIDs) == 0 &&
		conditions.CustomerSegment == "" &&
		conditions.MinItemQuantity == 0
}

// CouponConditionsAllowsUser reports whether a user may use a restricted promotion.
func CouponConditionsAllowsUser(conditions CouponConditions, userID uint) bool {
	if len(conditions.UserIDs) == 0 {
		return true
	}
	for _, id := range conditions.UserIDs {
		if id == userID {
			return true
		}
	}
	return false
}

// MarshalCouponConditions encodes conditions for persistence.
func MarshalCouponConditions(conditions CouponConditions) datatypes.JSON {
	if CouponConditionsIsEmpty(conditions) {
		return datatypes.JSON([]byte("{}"))
	}
	b, err := json.Marshal(conditions)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(b)
}
