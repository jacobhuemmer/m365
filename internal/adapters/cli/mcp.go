package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/m365/internal/domain"
)

const mcpHelp = `m365 mcp — stdio MCP for agents

Verbs: serve
serve: JSON-RPC on stdin/stdout. Tools: m365_status, m365_help, m365_run.
Recipe topics: mail-search, teams-find, calendar, files (also MCP prompts).
Writes through m365_run dry-run unless write_opt_in is true.
Do not use --human. Login stays m365 auth login in a terminal.
No session required for --help.
`

type helpIn struct {
	Namespace string `json:"namespace,omitempty" jsonschema:"optional CLI namespace"`
	Verb      string `json:"verb,omitempty" jsonschema:"optional verb"`
	Topic     string `json:"topic,omitempty" jsonschema:"recipe topic: mail-search, teams-find, calendar, or files"`
}

type runIn struct {
	Namespace  string         `json:"namespace,omitempty" jsonschema:"CLI namespace"`
	Verb       string         `json:"verb,omitempty" jsonschema:"CLI verb"`
	Args       []string       `json:"args,omitempty" jsonschema:"positional ids after the verb"`
	Flags      map[string]any `json:"flags,omitempty" jsonschema:"CLI long flag names without dashes"`
	WriteOptIn bool           `json:"write_opt_in,omitempty" jsonschema:"true to perform a real workload write"`
}

func runMCP(args []string, d Deps, format string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" || hasHelp(args) {
		return writeHelp(d.Stdout, mcpHelp)
	}
	if args[0] != "serve" {
		return fail(d, domain.Usagef("unknown mcp verb %q", args[0]))
	}
	if format == "human" {
		return fail(d, domain.Usage("mcp serve is JSON-RPC on stdio; do not use --human"))
	}
	if err := ServeMCP(d); err != nil {
		return fail(d, domain.Service(err.Error()))
	}
	return domain.ExitOK
}

func ServeMCP(d Deps) error {
	return NewMCPServer(d).Run(context.Background(), &mcp.StdioTransport{})
}

func NewMCPServer(d Deps) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "m365", Version: "1.0.0"}, nil)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "m365_status",
		Description: "Signed-in, session usable, per-namespace consent. No tokens. Does not open a browser.",
	}, func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, any, error) {
		return callCLI(d, []string{"m365", "auth", "status"}), nil, nil
	})
	mcp.AddTool(s, &mcp.Tool{
		Name:        "m365_help",
		Description: "CLI help for a namespace or verb, or recipe topic mail-search, teams-find, calendar, files. No session required.",
	}, handleHelp)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "m365_run",
		Description: "Run one CLI namespace+verb with a flag map. Returns that command's JSON. Writes dry-run unless write_opt_in is true. Lookup examples: help topics mail-search, teams-find, calendar, files.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in runIn) (*mcp.CallToolResult, any, error) {
		args, err := buildRunArgs(in.Namespace, in.Verb, in.Args, in.Flags, in.WriteOptIn)
		if err != nil {
			return toolErr(err), nil, nil
		}
		return callCLI(d, args), nil, nil
	})
	for _, name := range recipeNames {
		n := name
		s.AddPrompt(&mcp.Prompt{Name: n, Description: "Lookup recipe " + n}, recipePrompt(n))
	}
	return s
}

func handleHelp(_ context.Context, _ *mcp.CallToolRequest, in helpIn) (*mcp.CallToolResult, any, error) {
	if in.Topic != "" && (in.Namespace != "" || in.Verb != "") {
		return toolErr(domain.Usage("use topic or namespace, not both")), nil, nil
	}
	if in.Topic != "" {
		body, ok := recipe(in.Topic)
		if !ok {
			return toolErr(domain.Usagef("unknown topic %q", in.Topic)), nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: strings.TrimSpace(body)}}}, nil, nil
	}
	if in.Namespace == "" && in.Verb == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: strings.TrimSpace(mcpHelp)}}}, nil, nil
	}
	args := []string{"m365"}
	if in.Namespace != "" {
		args = append(args, in.Namespace)
	}
	if in.Verb != "" {
		args = append(args, in.Verb)
	}
	args = append(args, "--help")
	return callCLI(Deps{}, args), nil, nil
}

func recipePrompt(name string) mcp.PromptHandler {
	return func(context.Context, *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		body, _ := recipe(name)
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{{Role: "user", Content: &mcp.TextContent{Text: strings.TrimSpace(body)}}},
		}, nil
	}
}

func callCLI(d Deps, args []string) *mcp.CallToolResult {
	var out, errw bytes.Buffer
	d.Stdout, d.Stderr = &out, &errw
	code := Run(args, d)
	return toolFromCLI(redact(out.String()), redact(errw.String()), code)
}

func toolFromCLI(stdout, stderr string, code int) *mcp.CallToolResult {
	if code == domain.ExitOK {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: strings.TrimSpace(stdout)}}}
	}
	text := strings.TrimSpace(stderr)
	if text == "" {
		text = `{"class":"usage","message":"command failed"}`
	}
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

func toolErr(err error) *mcp.CallToolResult {
	msg := err.Error()
	cls := domain.ClassOf(err)
	hint := ""
	if de, ok := err.(*domain.Error); ok {
		hint = de.Hint
	}
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(errObj{Class: cls, Message: redact(msg), Hint: redact(hint)})
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: strings.TrimSpace(buf.String())}}}
}
