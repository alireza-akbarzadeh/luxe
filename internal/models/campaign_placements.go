package models

import (
	"encoding/json"

	"gorm.io/datatypes"
)

// CampaignPlacements references merchandising assets included in a campaign.
type CampaignPlacements struct {
	FlashDealIDs  []uint `json:"flash_deal_ids,omitempty"`
	SectionIDs    []uint `json:"section_ids,omitempty"`
	CollectionIDs []uint `json:"collection_ids,omitempty"`
}

// ParseCampaignPlacements unmarshals JSONB placements from a campaign row.
func ParseCampaignPlacements(raw datatypes.JSON) CampaignPlacements {
	if len(raw) == 0 {
		return CampaignPlacements{}
	}
	var placements CampaignPlacements
	_ = json.Unmarshal(raw, &placements)
	return placements
}

// MarshalCampaignPlacements encodes placements for persistence.
func MarshalCampaignPlacements(placements CampaignPlacements) datatypes.JSON {
	if len(placements.FlashDealIDs) == 0 &&
		len(placements.SectionIDs) == 0 &&
		len(placements.CollectionIDs) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	b, err := json.Marshal(placements)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(b)
}
