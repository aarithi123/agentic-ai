package tools

import (
	"context"
	"fmt"
	"log"
	"strings"

	"uf/mcp/pkg/kube"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
)

type NamespaceFinder struct{}

// Name of the tool
func (n *NamespaceFinder) Name() string {
	return "NamespaceFinder"
}

// Description of the tool
func (n *NamespaceFinder) Description() string {
	desc := []string{
		"Tool to list all namespaces in the Kubernetes cluster.",
		"Useful for discovering available environments in Minikube or other clusters.",
	}
	return strings.Join(desc, "\n")
}

func GetNamespaceFinderTool() (*protocol.Tool, server.ToolHandlerFunc) {
	log.Print("Initializing NamespaceFinder tool")

	toolStruct := NamespaceFinder{}

	tool, err := protocol.NewTool(
		toolStruct.Name(),
		toolStruct.Description(),
		toolStruct,
	)
	if err != nil {
		log.Fatalf("Failed to create tool: %v", err)
	}

	return tool, handleNamespaceFinder
}

// Tool execution logic
func handleNamespaceFinder(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var request NamespaceFinder

	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &request); err != nil {
		return nil, err
	}

	namespaces, err := kube.GetNamespaces()
	if err != nil {
		return &protocol.CallToolResult{
			Content: []protocol.Content{
				&protocol.TextContent{
					Type: "text",
					Text: fmt.Sprintf(`{"Error":"%v"}`, err),
				},
			},
			IsError: true,
		}, err
	}

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Namespaces:\n%s", strings.Join(namespaces, "\n")),
			},
		},
		IsError: false,
	}, nil
}
