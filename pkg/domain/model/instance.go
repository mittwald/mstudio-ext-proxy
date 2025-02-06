package model

type ExtensionInstanceContext struct {
	ID   string
	Kind string
}

type ExtensionInstance struct {
	ID      string `bson:"_id"`
	Enabled bool
	Context ExtensionInstanceContext
	Scopes  []string
	Secret  []byte
}
