package keychain

import "github.com/masonhuemmer/m365/internal/app/auth"

type Blob = auth.Blob

type Store interface {
	Get() (Blob, bool, error)
	Put(Blob) error
	Delete() error
}
