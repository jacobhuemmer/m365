package steps

import (
	"testing"

	"github.com/masonhuemmer/m365/acceptance/runtime"
)

func TestMCPAgentGuardsFeature(t *testing.T) {
	runtime.RunFeature(t, "../../features/mcp/agent-guards.feature")
}
