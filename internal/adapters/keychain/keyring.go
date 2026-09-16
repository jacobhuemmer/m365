package keychain

import (
	"encoding/json"

	"github.com/zalando/go-keyring"
)

const keyringService = "m365-cli"
const keyringUser = "session"

type Keyring struct{}

func (Keyring) Get() (Blob, bool, error) {
	s, err := keyring.Get(keyringService, keyringUser)
	if err == keyring.ErrNotFound {
		return Blob{}, false, nil
	}
	if err != nil {
		return Blob{}, false, err
	}
	var b Blob
	if err := json.Unmarshal([]byte(s), &b); err != nil {
		return Blob{}, false, err
	}
	return b, true, nil
}

func (Keyring) Put(b Blob) error {
	raw, err := json.Marshal(b)
	if err != nil {
		return err
	}
	return keyring.Set(keyringService, keyringUser, string(raw))
}

func (Keyring) Delete() error {
	err := keyring.Delete(keyringService, keyringUser)
	if err == keyring.ErrNotFound {
		return nil
	}
	return err
}
