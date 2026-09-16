package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestAuthStatusSignedOutLoginLogout(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"m365", "auth", "status"}, d)
	if code != 0 {
		t.Fatal(errw.String())
	}
	var st map[string]any
	if err := json.Unmarshal(out.Bytes(), &st); err != nil {
		t.Fatal(err, out.String())
	}
	if st["signed_in"] != false {
		t.Fatal(st)
	}
	if strings.Contains(out.String(), "token") {
		t.Fatal("token leaked")
	}
	out.Reset()
	code = Run([]string{"m365", "auth", "login"}, d)
	if code != 0 {
		t.Fatal(errw.String())
	}
	out.Reset()
	_ = Run([]string{"m365", "auth", "status"}, d)
	_ = json.Unmarshal(out.Bytes(), &st)
	if st["signed_in"] != true {
		t.Fatal(st)
	}
	out.Reset()
	if c := Run([]string{"m365", "auth", "logout"}, d); c != 0 {
		t.Fatal(c)
	}
	out.Reset()
	_ = Run([]string{"m365", "auth", "status"}, d)
	_ = json.Unmarshal(out.Bytes(), &st)
	if st["signed_in"] != false {
		t.Fatal(st)
	}
	if c := Run([]string{"m365", "mail", "list"}, d); c != domain.ExitAuth {
		t.Fatal(c, errw.String())
	}
}
