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

type DeploymentFinder struct {
	Namespace string `json:"namespace" description:"Name of the Namespace for which the list of deployments are requested" required:"true"`
}

// Name of the tool
func (d *DeploymentFinder) Name() string {
	return "DeploymentFinder"
}

// Description of the tool
func (d *DeploymentFinder) Description() string {
	desc := []string{
		"Tool to find the list of deployments in a given Kubernetes Namespace.",
		"Useful for inspecting active applications in Minikube or other clusters.",
	}
	return strings.Join(desc, "\n")
}

func GetDeploymentFinderTool() (*protocol.Tool, server.ToolHandlerFunc) {
	log.Print("Initializing DeploymentFinder tool")

	toolStruct := DeploymentFinder{}

	tool, err := protocol.NewTool(
		toolStruct.Name(),
		toolStruct.Description(),
		toolStruct,
	)
	if err != nil {
		log.Fatalf("Failed to create tool: %v", err)
	}

	return tool, handleDeploymentFinder
}

// Tool execution logic
func handleDeploymentFinder(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var request DeploymentFinder

	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &request); err != nil {
		return nil, err
	}

	deployments, err := kube.GetDeployments(request.Namespace)
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
				Text: fmt.Sprintf("Deployments in namespace '%s':\n%s", request.Namespace, strings.Join(deployments, "\n")),
			},
		},
		IsError: false,
	}, nil
}
