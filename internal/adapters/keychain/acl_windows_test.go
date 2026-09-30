//go:build windows

package keychain

import (
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestFileStoreIsOwnerOnlyOnWindows(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "m365")
	path := filepath.Join(dir, "session.json")
	f := &FileStore{Path: path}
	if err := f.Put(Blob{Account: "user@example.com", Usable: true}); err != nil {
		t.Fatal(err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{dir, path} {
		sd, err := windows.GetNamedSecurityInfo(p, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			t.Fatal(err)
		}
		control, _, _ := sd.Control()
		if control&windows.SE_DACL_PROTECTED == 0 {
			t.Fatalf("%s still inherits its parent ACL", p)
		}
		acl, _, err := sd.DACL()
		if err != nil || acl == nil {
			t.Fatalf("%s dacl: %v", p, err)
		}
		if acl.AceCount != 1 {
			t.Fatalf("%s has %d ACL entries, want 1", p, acl.AceCount)
		}
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(acl, 0, &ace); err != nil {
			t.Fatal(err)
		}
		if !(*windows.SID)(unsafe.Pointer(&ace.SidStart)).Equals(user.User.Sid) {
			t.Fatalf("%s entry is not for the current user", p)
		}
	}
}
