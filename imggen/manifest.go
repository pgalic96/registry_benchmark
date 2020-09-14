package imggen

import (
	"crypto/rand"
	"log"
	"os"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
)

var (
	configSize = 1024
)

// GenerateManifest generates manifest for imggen generated layers and returns config layer digest
func GenerateManifest(items []os.FileInfo, yamlFilename string) (*schema2.DeserializedManifest, digest.Digest) {

	layers := make([]distribution.Descriptor, len(items))
	for i, item := range items {
		digest := digest.NewDigestFromHex(
			"sha256",
			item.Name(),
		)
		layer := distribution.Descriptor{
			Digest:    digest,
			MediaType: schema2.MediaTypeLayer,
			Size:      item.Size(),
		}
		layers[i] = layer
	}

	manifest := schema2.Manifest{
		Versioned: schema2.SchemaVersion,
		Layers:    layers,
		Config:    createConfigFile(yamlFilename),
	}

	deserializedManifest, _ := schema2.FromStruct(manifest)
	return deserializedManifest, deserializedManifest.Config.Digest
}

func createConfigFile(yamlFilename string) distribution.Descriptor {
	fd, err := Create("config-file")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	size := int64(configSize)
	fd.Seek(size-9, 0)

	randbytes := make([]byte, 8)
	rand.Read(randbytes)

	fd.Write(randbytes)
	fd.Write([]byte{0})

	err = fd.Close()
	if err != nil {
		log.Fatal("Failed to close file")
	}
	configDigest, err := sha256Digest("config-file")
	if err != nil {
		log.Fatal(err)
	}
	digest := digest.NewDigestFromHex(
		"sha256",
		configDigest,
	)

	err = os.Rename("config-file", configDigest)
	if err != nil {
		log.Fatal(err)
	}

	return distribution.Descriptor{
		Digest:    digest,
		MediaType: schema2.MediaTypeImageConfig,
		Size:      size,
	}
}
