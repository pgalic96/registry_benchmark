package imggen

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"

	"gopkg.in/yaml.v3"
)

// ImgGen is a struct containing relevant config for Image Generation tool
type ImgGen struct {
	ImgSizeMb      int  `yaml:"img-size-mb,omitempty"`
	LayerNumber    int  `yaml:"layer-number,omitempty"`
	GenerateRandom bool `yaml:"generate-random,omitempty"`
}

// Config is the config for image generation
type Config struct {
	ImageGeneration ImgGen
}

func loadConfig() (*Config, error) {
	c := Config{}
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return &c, nil
}

// Generate a docker image out of yaml config file
func Generate() {
	log.Printf("Loading config file")
	config, err := loadConfig()
	layerSize := int64((config.ImageGeneration.ImgSizeMb / config.ImageGeneration.LayerNumber) * 1024 * 1024)
	fd, err := os.Create("imggen")
	if err != nil {
		log.Fatal("Failed to create file")
	}
	//for i := 0; i < config.ImageGeneration.LayerNumber; i++ {
	if config.ImageGeneration.GenerateRandom {
		_, err = fd.Seek(layerSize-9, 0)
		randbytes := make([]byte, 8)
		rand.Read(randbytes)
		_, err = fd.Write(randbytes)
		_, err = fd.Write([]byte{0})
		err = fd.Close()
		if err != nil {
			log.Fatal("Failed to close file")
		}
	}
	//}
}
