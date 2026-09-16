package cli

import "testing"

func TestChatAliasVerbs(t *testing.T) {
	d, _, errw := testDeps()
	login(t, d)
	if c := Run([]string{"m365", "chat", "get", "chat-1"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	if c := Run([]string{"m365", "chat", "messages", "chat-1"}, d); c != 0 {
		t.Fatal(errw.String())
	}
}
