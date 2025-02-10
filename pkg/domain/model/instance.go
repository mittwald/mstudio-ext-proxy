package model

type ExtensionInstanceContext struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

type ExtensionInstance struct {
	ID      string                   `bson:"_id" json:"id"`
	Enabled bool                     `json:"enabled"`
	Context ExtensionInstanceContext `json:"context"`
	Scopes  []string                 `json:"scopes"`
	Secret  []byte                   `json:"secret"`
}
