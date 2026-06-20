package dto

// ImportRowResult reports what happened to a single Excel row.
type ImportRowResult struct {
	Row    int    `json:"row"`              // 1-based Excel row number (row 1 = header)
	Status string `json:"status"`           // "created" | "skipped" | "failed"
	ID     *uint  `json:"id,omitempty"`     // ID of the created record (when status=created)
	Name   string `json:"name,omitempty"`   // display value from the row for easy identification
	Error  string `json:"error,omitempty"`  // human-readable reason (when status=failed/skipped)
}

// ImportSummary is returned after a bulk import from Excel.
type ImportSummary struct {
	TotalRows int               `json:"total_rows"`   // data rows parsed (header excluded)
	Created   int               `json:"created"`      // successfully inserted
	Failed    int               `json:"failed"`       // validation or DB errors
	Skipped   int               `json:"skipped"`      // intentionally skipped (e.g. empty row)
	Rows      []ImportRowResult `json:"rows"`         // per-row detail
}
