package auth

import (
	"io/ioutil"
	"registry_benchmark/config"
)

var registryConfig *config.Config

// Config contains google cloud auth key
type Config struct {
	GoogleRegistryKey string `yaml:"gcloud-key,omitempty"`
}

// GetGCRAuthorizationKey obtains key for authenticating with Google Cloud Registry
func GetGCRAuthorizationKey(filename string) (string, error) {
	registryConfig, _ := config.LoadConfig(filename)
	content, err := ioutil.ReadFile(registryConfig.GoogleRegistryKey)
	if err != nil {
		return "", err
	}

	text := string(content)
	return text, nil
}
