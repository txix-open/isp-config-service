package rqlite

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/rqlite/rqlite/v8/auth"
	"github.com/txix-open/isp-kit/json"
)

func credentialsStore(credentials []auth.Credential) (*auth.CredentialsStore, error) {
	data, err := json.Marshal(credentials)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal credentials")
	}

	store := auth.NewCredentialsStore()
	err = store.Load(bytes.NewReader(data))
	if err != nil {
		return nil, errors.WithMessage(err, "load rqlite credentials")
	}

	return store, nil
}
