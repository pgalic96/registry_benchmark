package auth

import (
	"log"
	"strings"

	"registry_benchmark/config"

	"github.com/pgalic96/docker-registry-client/registry"
)

// ObtainRegistryCredentials returns username and password for the given registry
func ObtainRegistryCredentials(containerReg config.Registry, filename string) (string, string, error) {
	var password string
	if strings.HasPrefix(containerReg.Platform, "ecr") {
		token, err := GetECRAuthorizationToken(containerReg.AccountID, containerReg.Region)
		if err != nil {
			return "", "", err
		}
		password = strings.TrimPrefix(token, "AWS:")
		if []byte(password[len(password)-1:])[0] == []byte{0}[0] {
			password = password[:len(password)-1]
		}
	} else if strings.HasPrefix(containerReg.Platform, "gcr") {
		log.Println("Entering get auth key")
		password, _ = GetGCRAuthorizationKey(filename)
	} else {
		password = containerReg.Password
	}
	return containerReg.Username, password, nil
}

// AuthenticateRegistry authenticates with the provided registry using config provided in yaml
func AuthenticateRegistry(containerReg config.Registry, filename string) (*registry.Registry, error) {
	username, password, err := ObtainRegistryCredentials(containerReg, filename)
	if err != nil {
		return nil, err
	}
	hub, err := registry.New(containerReg.URL, username, password)
	if err != nil {
		return nil, err
	}
	return hub, nil
}
