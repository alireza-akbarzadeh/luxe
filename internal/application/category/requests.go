package category

// BulkDeleteRequest is the payload for bulk category deletion.
type BulkDeleteRequest struct {
	IDs []uint `json:"ids" validate:"required,min=1"`
}
