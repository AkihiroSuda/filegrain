package builder

type Builder interface {
	Build(img, refName string) error
}
