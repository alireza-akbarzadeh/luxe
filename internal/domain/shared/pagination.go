package shared

// PageParams holds list pagination inputs.
type PageParams struct {
	Limit  int
	Offset int
}

// PageResult holds paginated list metadata.
type PageResult struct {
	Total  int64
	Limit  int
	Offset int
}

// Normalize applies defaults and caps for limit/offset.
func (p PageParams) Normalize(defaultLimit, maxLimit int) PageParams {
	if defaultLimit <= 0 {
		defaultLimit = 20
	}
	if maxLimit <= 0 {
		maxLimit = 100
	}
	limit := p.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}
	return PageParams{Limit: limit, Offset: offset}
}
