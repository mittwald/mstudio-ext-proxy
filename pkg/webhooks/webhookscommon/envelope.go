package webhookscommon

type Envelope struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}
