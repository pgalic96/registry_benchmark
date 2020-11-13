package config

import (
	"io/ioutil"
	"log"
	"registry_benchmark/imggen"

	"gopkg.in/yaml.v3"
)

// Registry is the struct for single registry config
type Registry struct {
	Platform     string
	ImageURL     string `yaml:"image-url,omitempty"`
	URL          string `yaml:"registry-url,omitempty"`
	Username     string
	Password     string
	Repository   string
	AccountID    string `yaml:"account-id,omitempty"`
	Region       string
	WithManifest bool `yaml:"upload-manifest,omitempty"`
}

// Config is the configuration for the benchmark
type Config struct {
	Registries        []Registry
	ImageGeneration   imggen.ImgGen       `yaml:"image-generation,omitempty"`
	ImageName         string              `yaml:"image-name,omitempty"`
	Iterations        int                 `yaml:"iterations,omitempty"`
	StorageURL        string              `yaml:"storage-url,omitempty"`
	PullSourceFolder  string              `yaml:"pull-source-folder,omitempty"`
	ReplayerConfig    TraceReplayerConfig `yaml:"trace-replayer,omitempty"`
	GoogleRegistryKey string              `yaml:"gcloud-key,omitempty"`
}

// LoadConfig is the function for loading configuration from yaml file
func LoadConfig(filename string) (*Config, error) {
	c := Config{}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return &c, nil
}
