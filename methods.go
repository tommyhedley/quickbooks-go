package quickbooks

type Creatable[T any] interface {
	Create(params RequestParameters, object *T) (*T, error)
}
