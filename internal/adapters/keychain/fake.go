package keychain

import "sync"

type Fake struct {
	mu sync.Mutex
	v  *Blob
}

func (f *Fake) Get() (Blob, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.v == nil {
		return Blob{}, false, nil
	}
	return *f.v, true, nil
}

func (f *Fake) Put(b Blob) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := b
	f.v = &cp
	return nil
}

func (f *Fake) Delete() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.v = nil
	return nil
}
