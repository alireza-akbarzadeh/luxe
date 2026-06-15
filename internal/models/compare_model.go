package models

import (
	"database/sql/driver"
	"time"

	"github.com/lib/pq"
)

// UintArray maps PostgreSQL int[] columns for GORM scan/value.
type UintArray []uint

func (a *UintArray) Scan(value interface{}) error {
	if value == nil {
		*a = UintArray{}
		return nil
	}

	var arr pq.Int64Array
	if err := arr.Scan(value); err != nil {
		return err
	}

	*a = make(UintArray, len(arr))
	for i, id := range arr {
		(*a)[i] = uint(id)
	}
	return nil
}

func (a UintArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return pq.Array([]int64{}).Value()
	}

	arr := make(pq.Int64Array, len(a))
	for i, id := range a {
		arr[i] = int64(id)
	}
	return arr.Value()
}

type CompareList struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id,omitempty"`
	ProductIDs UintArray `gorm:"type:int[]" json:"product_ids"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
