package keychain

import "testing"

func TestFakeDoesNotUseKeychain(t *testing.T) {
	f := &Fake{}
	_, ok, err := f.Get()
	if err != nil || ok {
		t.Fatal("empty")
	}
	if err := f.Put(Blob{Account: "a", Usable: true}); err != nil {
		t.Fatal(err)
	}
	b, ok, err := f.Get()
	if err != nil || !ok || b.Account != "a" {
		t.Fatalf("%+v", b)
	}
	if err := f.Delete(); err != nil {
		t.Fatal(err)
	}
	_, ok, _ = f.Get()
	if ok {
		t.Fatal("deleted")
	}
}
