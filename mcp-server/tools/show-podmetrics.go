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

type PodCpuMemoryViewer struct {
	Namespace string `json:"namespace" description:"Namespace to inspect pod CPU and memory usage" required:"true"`
}

// Name of the tool
func (p *PodCpuMemoryViewer) Name() string {
	return "PodCpuMemoryViewer"
}

// Description of the tool
func (p *PodCpuMemoryViewer) Description() string {
	desc := []string{
		"Displays CPU and memory usage for all pods in a given Kubernetes namespace.",
		"Requires metrics-server to be enabled (e.g., in Minikube).",
	}
	return strings.Join(desc, "\n")
}

func GetPodCpuMemoryViewerTool() (*protocol.Tool, server.ToolHandlerFunc) {
	log.Print("Initializing PodCpuMemoryViewer tool")

	toolStruct := PodCpuMemoryViewer{}

	tool, err := protocol.NewTool(
		toolStruct.Name(),
		toolStruct.Description(),
		toolStruct,
	)
	if err != nil {
		log.Fatalf("Failed to create PodCpuMemoryViewer tool: %v", err)
		return nil, nil
	}

	return tool, handlePodCpuMemoryViewer
}

// Tool execution logic
func handlePodCpuMemoryViewer(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var request PodCpuMemoryViewer

	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &request); err != nil {
		return nil, err
	}

	metrics, err := kube.GetPodCpuMemory(request.Namespace)
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

	var lines []string
	for _, m := range metrics {
		lines = append(lines, fmt.Sprintf("Pod: %s | CPU: %s | Memory: %s", m.Name, m.CPU, m.Memory))
	}

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: strings.Join(lines, "\n"),
			},
		},
		IsError: false,
	}, nil
}
