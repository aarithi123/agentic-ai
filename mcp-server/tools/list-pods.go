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

type PodFinder struct {
	Namespace string `json:"namespace" description:"Name of the Namespace for which the list of pods are requested" required:"true"`
}

// Name of the tool
func (p *PodFinder) Name() string {
	return "PodFinder"
}

// Description of the tool
func (p *PodFinder) Description() string {
	desc := []string{
		"Tool to find the list of pods in a given Kubernetes Namespace.",
		"Useful for inspecting workload status and debugging cluster activity.",
	}
	return strings.Join(desc, "\n")
}

func GetPodFinderTool() (*protocol.Tool, server.ToolHandlerFunc) {
	log.Print("Initializing PodFinder tool")

	toolStruct := PodFinder{}

	tool, err := protocol.NewTool(
		toolStruct.Name(),
		toolStruct.Description(),
		toolStruct,
	)
	if err != nil {
		log.Fatalf("Failed to create tool: %v", err)
	}

	return tool, handlePodFinder
}

// Tool execution logic
func handlePodFinder(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var request PodFinder

	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &request); err != nil {
		return nil, err
	}

	pods, err := kube.GetPods(request.Namespace)
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
				Text: fmt.Sprintf("Pods in namespace '%s':\n%s", request.Namespace, strings.Join(pods, "\n")),
			},
		},
		IsError: false,
	}, nil
}
