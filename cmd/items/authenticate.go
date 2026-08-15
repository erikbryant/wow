package main

import (
	"github.com/erikbryant/wow/internal/credentials"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/wowapi"
)

// TODO: This is a duplicate of authenticate() in internal/application/application.go

// authenticate authenticates this session against the WoW web APIs
func authenticate(paths *path.Paths) error {
	clientID, err := credentials.ReadFromKeychain(paths.Secret, "clientID")
	if err != nil {
		return err
	}

	clientSecret, err := credentials.ReadFromKeychain(paths.Secret, "clientSecret")
	if err != nil {
		return err
	}

	err = wowapi.Authenticate(clientID, clientSecret)
	if err != nil {
		return err
	}

	return nil
}
