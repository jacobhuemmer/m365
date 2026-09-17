package cli

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPStatusUnderTwoSeconds(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	cs := connectMCP(t, d)
	start := time.Now()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "m365_status"})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("status slow")
	}
}
