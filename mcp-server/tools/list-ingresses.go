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

type IngressFinder struct {
	Namespace string `json:"namespace" description:"Name of the Namespace where the Ingresses are defined" required:"true"`
}

// Name of the tool
func (i *IngressFinder) Name() string {
	return "IngressFinder"
}

// Description of the tool
func (i *IngressFinder) Description() string {
	desc := []string{
		"Tool to list all Ingresses in a given Kubernetes Namespace.",
		"Compatible with Minikube and other Kubernetes clusters.",
	}
	return strings.Join(desc, "\n")
}

func GetIngressFinderTool() (*protocol.Tool, server.ToolHandlerFunc) {
	log.Print("Initializing IngressFinder tool")

	toolStruct := IngressFinder{}

	tool, err := protocol.NewTool(
		toolStruct.Name(),
		toolStruct.Description(),
		toolStruct,
	)
	if err != nil {
		log.Fatalf("Failed to create tool: %v", err)
	}

	return tool, handleIngressFinder
}

// Tool execution logic
func handleIngressFinder(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var request IngressFinder

	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &request); err != nil {
		return nil, err
	}

	ingresses, err := kube.GetIngresses(request.Namespace)
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
				Text: fmt.Sprintf("Ingresses in namespace '%s':\n%s", request.Namespace, strings.Join(ingresses, "\n")),
			},
		},
		IsError: false,
	}, nil
}
