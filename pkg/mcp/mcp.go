package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"uf/mcp/pkg/common"
	"uf/mcp/pkg/llm"

	"github.com/ThinkInAIXYZ/go-mcp/client"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
)

type ToolInfo struct {
	ToolName     *protocol.Tool
	SourceClient *client.Client
	envURL       string
}

var (
	// Serialized version of the Tools struct
	toolList string

	// Cross reference table of tools. toolName -> Tool struct
	toolsTable = make(map[string]ToolInfo)
)

func GetToolListSchema() string {
	return toolList
}

func CallTool(ctx context.Context, selectedTool *llm.SelectedToolInfo) (string, error) {
	toolInfo, ok := toolsTable[selectedTool.ToolName]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", selectedTool.ToolName)
	}

	targetClient := toolInfo.SourceClient
	request := protocol.NewCallToolRequest(selectedTool.ToolName, selectedTool.ToolArgs)

	result, err := targetClient.CallTool(ctx, request)
	if err != nil {
		return "", err
	}

	textContent := result.Content[0].(*protocol.TextContent)
	if result.IsError {
		return textContent.Text, fmt.Errorf("error calling tool: %s", textContent.Text)
	}

	return textContent.Text, nil
}

func InitalizeTools() {
	allToolsForSchema := []*protocol.Tool{}
	toolsTable = make(map[string]ToolInfo)
	mcpClients := common.GetMcpClients()

	if len(mcpClients) == 0 {
		log.Println("No mcp clients found")
		toolList = `{tools: []}`
		return
	}

	for serverName, client := range mcpClients {
		result, err := client.ListTools(context.Background())
		if err != nil {
			log.Printf("Failed to list tools from MCP client '%s': %v", serverName, err)
			continue
		}

		for _, tool := range result.Tools {
			if _, exists := toolsTable[tool.Name]; exists {
				log.Fatalf("Tool '%s' already exists in MCP client '%s'", tool.Name, serverName)
			}

			toolsTable[tool.Name] = ToolInfo{
				ToolName:     tool,
				SourceClient: client,
			}
			allToolsForSchema = append(allToolsForSchema, tool)
		}
	}

	jsonDoc, err := json.Marshal(struct {
		Tools []*protocol.Tool `json:"tools"`
	}{
		Tools: allToolsForSchema,
	})

	if err != nil {
		log.Fatalf("ToolList Marshal. Err: %v\n", err)
	}

	toolList = string(jsonDoc)
	//log.Printf("ToolList: %s", toolList)
}
