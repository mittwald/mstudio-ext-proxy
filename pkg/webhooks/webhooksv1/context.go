package webhooksv1

type ContextKind string

const (
	ContextKindCustomer ContextKind = "customer"
	ContextKindProject  ContextKind = "project"
)

type Context struct {
	ID   string      `json:"id"`
	Kind ContextKind `json:"kind"`
}
