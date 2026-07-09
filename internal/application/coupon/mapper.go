package coupon

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	domaincoupon "github.com/alireza-akbarzadeh/luxe/internal/domain/coupon"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

func normalizeApplicationType(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case constants.CouponApplicationAutomatic:
		return constants.CouponApplicationAutomatic
	case constants.CouponApplicationBOGO:
		return constants.CouponApplicationBOGO
	default:
		return constants.CouponApplicationCode
	}
}

func generateCouponCode(appType string) string {
	prefix := "AUTO"
	switch appType {
	case constants.CouponApplicationBOGO:
		prefix = "BOGO"
	case constants.CouponApplicationAutomatic:
		prefix = "AUTO"
	default:
		prefix = "PROMO"
	}
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	return strings.ToUpper(prefix + "-" + hex.EncodeToString(buf))
}

func resolveCouponCode(code, appType string) string {
	trimmed := strings.TrimSpace(strings.ToUpper(code))
	if trimmed != "" {
		return trimmed
	}
	return generateCouponCode(appType)
}

func conditionsFromRequest(req *dto.CouponConditionsRequest) models.CouponConditions {
	if req == nil {
		return models.CouponConditions{}
	}
	return models.CouponConditions{
		FirstOrderOnly:  req.FirstOrderOnly,
		CategoryIDs:     req.CategoryIDs,
		ProductIDs:      req.ProductIDs,
		UserIDs:         req.UserIDs,
		CustomerSegment: req.CustomerSegment,
		MinItemQuantity: req.MinItemQuantity,
	}
}

func applyConditionsRequest(coupon *models.Coupon, req *dto.CouponConditionsRequest) {
	if req == nil {
		return
	}
	coupon.Conditions = models.MarshalCouponConditions(conditionsFromRequest(req))
}

func couponFromModel(c models.Coupon) domaincoupon.Coupon {
	conditions := models.ParseCouponConditions(c.Conditions)
	appType := normalizeApplicationType(c.ApplicationType)

	return domaincoupon.Coupon{
		ApplicationType:        appType,
		DiscountType:           c.DiscountType,
		DiscountValue:          c.DiscountValue,
		MinimumOrderAmount:     c.MinimumOrderAmount,
		MaxDiscountAmount:      c.MaxDiscountAmount,
		UsageLimit:             c.UsageLimit,
		UsedCount:              c.UsedCount,
		IsActive:               c.IsActive,
		StartDate:              c.StartDate,
		EndDate:                c.EndDate,
		BogoBuyQuantity:        c.BogoBuyQuantity,
		BogoGetQuantity:        c.BogoGetQuantity,
		BogoGetDiscountPercent: c.BogoGetDiscountPercent,
		MinItemQuantity:        conditions.MinItemQuantity,
	}
}
