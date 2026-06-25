package utils

import "strconv"

func ParseStrToUnit(userIDStr string) (uint, error) {
	id64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, ErrBadRequest("invalid user ID format: must be numeric")
	}
	return uint(id64), nil
}
