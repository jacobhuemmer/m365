package keychain

import "strings"

type Fallback struct {
	Primary   Store
	Secondary Store
}

func (f *Fallback) Get() (Blob, bool, error) {
	if f.Primary != nil {
		b, ok, err := f.Primary.Get()
		if err == nil && ok {
			return b, true, nil
		}
	}
	if f.Secondary != nil {
		return f.Secondary.Get()
	}
	return Blob{}, false, nil
}

func (f *Fallback) Put(b Blob) error {
	var err error
	if f.Primary != nil {
		err = f.Primary.Put(b)
		if err == nil {
			return nil
		}
		if !tooLarge(err) {
			return err
		}
	}
	if f.Secondary != nil {
		return f.Secondary.Put(b)
	}
	return err
}

func (f *Fallback) Delete() error {
	var e1, e2 error
	if f.Primary != nil {
		e1 = f.Primary.Delete()
	}
	if f.Secondary != nil {
		e2 = f.Secondary.Delete()
	}
	if e1 != nil {
		return e1
	}
	return e2
}

func tooLarge(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "too big") || strings.Contains(s, "too large")
}
