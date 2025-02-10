package proxy

import (
	"encoding/json"
	"net/url"
)

type jsonURL url.URL

func (j *jsonURL) UnmarshalJSON(bytes []byte) error {
	urlStr := ""
	if err := json.Unmarshal(bytes, &urlStr); err != nil {
		return err
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	*j = jsonURL(*parsed)
	return nil
}

type Configuration struct {
	UpstreamURL jsonURL
	StripPrefix string
}

type ConfigurationCollection map[string]Configuration

func (cc *ConfigurationCollection) Decode(value string) error {
	if err := json.Unmarshal([]byte(value), &cc); err != nil {
		return err
	}

	return nil
}
