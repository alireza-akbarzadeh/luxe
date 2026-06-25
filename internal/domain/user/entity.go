package user

// User is the account aggregate root (stub for migration).
type User struct {
	ID        uint
	Email     string
	FirstName string
	LastName  string
	Phone     string
	Role      string
	IsActive  bool
}
