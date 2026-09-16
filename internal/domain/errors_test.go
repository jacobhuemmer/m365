package domain

import "testing"

func TestExitCodes(t *testing.T) {
	cases := []struct {
		err  error
		code int
		cls  string
	}{
		{nil, ExitOK, ""},
		{Usage("bad"), ExitUsage, ClassUsage},
		{Auth("no"), ExitAuth, ClassAuth},
		{Service("down"), ExitService, ClassService},
		{NotFound("id"), ExitNotFound, ClassNotFound},
	}
	for _, c := range cases {
		if got := ExitOf(c.err); got != c.code {
			t.Fatalf("exit %v: got %d want %d", c.err, got, c.code)
		}
		if got := ClassOf(c.err); got != c.cls {
			t.Fatalf("class %v: got %q want %q", c.err, got, c.cls)
		}
	}
}
