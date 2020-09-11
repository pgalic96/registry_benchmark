package auth

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Config contains google cloud auth key
type Config struct {
	GoogleRegistryKey string `yaml:"gcloud-key,omitempty"`
}

func loadConfig(yamlFilename string) (*Config, error) {
	c := Config{}
	yamlFile, err := ioutil.ReadFile(yamlFilename)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return &c, nil
}

// GetGCRAuthorizationKey obtains key for authenticating with Google Cloud Registry
func GetGCRAuthorizationKey(filename string) (string, error) {
	config, _ := loadConfig(filename)
	log.Println(config.GoogleRegistryKey)
	content, err := ioutil.ReadFile(config.GoogleRegistryKey)
	if err != nil {
		return "", err
	}

	text := string(content)
	return text, nil
}
