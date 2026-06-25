package category

// Category is the category aggregate (domain view).
type Category struct {
	ID       uint
	Name     string
	Slug     string
	ParentID *uint
	IsActive bool
	Level    int
	Path     string
}
