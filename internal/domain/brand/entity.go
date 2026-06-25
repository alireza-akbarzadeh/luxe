package brand

// Brand is the brand aggregate (domain view).
type Brand struct {
	ID     uint
	Name   string
	Slug   string
	Status string
}
