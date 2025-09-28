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

type ServiceRestarter struct {
	Namespace string `json:"namespace" description:"Name of the Namespace where the service is hosted" required:"true"`
	Service   string `json:"service" description:"Name of the service (Deployment) to be restarted" required:"true"`
}

// Name of the tool
func (s *ServiceRestarter) Name() string {
	return "ServiceRestarter"
}

// Description of the tool
func (s *ServiceRestarter) Description() string {
	desc := []string{
		"Tool to restart a Kubernetes Deployment in a given Namespace.",
		"This triggers a rolling restart by patching the deployment's pod template.",
		"Compatible with Minikube and other Kubernetes clusters.",
	}
	return strings.Join(desc, "\n")
}

func GetServiceRestarterTool() (*protocol.Tool, server.ToolHandlerFunc) {
	log.Print("Initializing ServiceRestarter tool")

	toolStruct := ServiceRestarter{}

	tool, err := protocol.NewTool(
		toolStruct.Name(),
		toolStruct.Description(),
		toolStruct,
	)
	if err != nil {
		log.Fatalf("Failed to create tool: %v", err)
	}

	return tool, handleServiceRestarter
}

// Tool execution logic
func handleServiceRestarter(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var request ServiceRestarter

	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &request); err != nil {
		return nil, err
	}

	resultMsg, err := kube.RestartApplication(request.Namespace, request.Service)
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
				Text: resultMsg,
			},
		},
		IsError: false,
	}, nil
}
