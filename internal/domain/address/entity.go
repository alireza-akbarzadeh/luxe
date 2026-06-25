package address

// Address is a user shipping/billing address (stub).
type Address struct {
	ID          uint
	UserID      uint
	AddressType string
	IsDefault   bool
}
