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
		conditions.CustomerSegment == "" &&
		conditions.MinItemQuantity == 0
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
