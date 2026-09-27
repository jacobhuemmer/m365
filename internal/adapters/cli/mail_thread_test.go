package cli

import (
	"encoding/json"
	"testing"
)

// mail thread returns the whole conversation; bodies only with --bodies,
// and flags may come before or after the id.
func TestMailThread(t *testing.T) {
	for _, tc := range []struct {
		args   []string
		bodies bool
	}{
		{[]string{"msg-1"}, false},
		{[]string{"--bodies", "msg-1"}, true},
		{[]string{"msg-2", "--bodies"}, true},
	} {
		d, out, errw := testDeps()
		loginAll(t, &d)
		out.Reset()
		if code := Run(append([]string{"m365", "mail", "thread"}, tc.args...), d); code != 0 {
			t.Fatalf("%v: exit %d %s", tc.args, code, errw.String())
		}
		var th struct {
			ConversationID string `json:"conversation_id"`
			Count          int    `json:"count"`
			Items          []struct {
				ID   string `json:"id"`
				Body string `json:"body"`
			} `json:"items"`
		}
		if err := json.Unmarshal(out.Bytes(), &th); err != nil {
			t.Fatal(err, out.String())
		}
		if th.ConversationID != "conv-1" || th.Count != 2 || len(th.Items) != 2 {
			t.Fatalf("%v: %s", tc.args, out.String())
		}
		for _, it := range th.Items {
			if (it.Body != "") != tc.bodies {
				t.Fatalf("%v: body %q with bodies=%v", tc.args, it.Body, tc.bodies)
			}
		}
	}
	d, _, _ := testDeps()
	loginAll(t, &d)
	if code := Run([]string{"m365", "mail", "thread"}, d); code != 3 {
		t.Fatalf("missing id: exit %d want 3", code)
	}
}
