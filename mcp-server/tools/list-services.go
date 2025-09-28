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

type ServiceFinder struct {
	Namespace string `json:"namespace" description:"Name of the Namespace for which the list of services are requested" required:"true"`
}

// Name of the tool
func (s *ServiceFinder) Name() string {
	return "ServiceFinder"
}

// Description of the tool
func (s *ServiceFinder) Description() string {
	desc := []string{
		"Tool to find the list of services in a given Namespace.",
		"Service is also known as Application.",
	}
	return strings.Join(desc, "\n")
}

func GetServiceFinderTool() (*protocol.Tool, server.ToolHandlerFunc) {
	log.Print("Initializing ServiceFinder tool")

	toolStruct := ServiceFinder{}

	tool, err := protocol.NewTool(
		toolStruct.Name(),
		toolStruct.Description(),
		toolStruct,
	)
	if err != nil {
		log.Fatalf("Failed to create tool: %v", err)
	}

	return tool, handleServiceFinder
}

// Tool execution logic
func handleServiceFinder(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var request ServiceFinder

	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &request); err != nil {
		return nil, err
	}

	services, err := kube.GetServices(request.Namespace) // ✅ Updated function call
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
				Text: fmt.Sprintf("%v", services),
			},
		},
		IsError: false,
	}, nil
}
