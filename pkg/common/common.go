package common

import (
	"log"
	"net/http"
	"os"
	custom "uf/mcp/pkg/transport"

	"github.com/ThinkInAIXYZ/go-mcp/client"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	"github.com/tmc/langchaingo/llms/openai"
)

func GetModel() *openai.LLM {
	// get llm model
	// Get LLM Userid & Password from environment variables

	var llmUrl, user, password, llmToken string

	if v, found := os.LookupEnv("LLM_URL"); found {
		llmUrl = v
	} else {
		log.Fatalf("env variable LLM_URL is required")
	}

	/*
		if v, found := os.LookupEnv("LLM_USER"); found {
			user = v
		} else {
			log.Fatalf("env variable LLM_USER is required")
		}

		if v, found := os.LookupEnv("LLM_PASSWORD"); found {
			password = v
		} else {
			log.Fatalf("env variable LLM_PASSWORD is required")
		}
	*/

	if v, found := os.LookupEnv("LLM_TOKEN"); found {
		llmToken = v
	} else {
		log.Fatalf("env variable LLM_TOKEN is required")
	}

	// Setup LLM Client...
	ct := custom.NewCustomTransport()
	ct.User = user
	ct.Password = password
	ct.Token = llmToken

	ct.Path = "v1/chat/completions"
	client := &http.Client{
		Transport: ct,
	}

	ct.AlterBody = true
	//ct.Debug = true

	llm, err := openai.New(openai.WithBaseURL(llmUrl),
		openai.WithHTTPClient(client),
		openai.WithModel("gpt-4o"),
		openai.WithToken(llmToken),
	)

	if err != nil {
		log.Fatalf("1. %v", err)
	}
	return llm
}

func GetMcpClients() map[string]*client.Client {
	mcpClients := make(map[string]*client.Client)

	// define a map of environemnt variables to mcp client names
	envUrls := map[string]string{
		"OCP_MCP_URL": "ocp",
		//"ARGOCD_MCP_URL": "argocd",
	}

	for envName, clientName := range envUrls {
		// Get mcpUrl from environment variable
		mcpUrl, found := os.LookupEnv(envName)

		if !found {
			log.Printf("env variable %s is not found, skipping mcp client %s", envName, clientName)
			continue
		}

		ct := custom.NewCustomTransport()
		//ct.Debug = true

		httpClient := &http.Client{
			Transport: ct,
		}

		transportClient, err := transport.NewStreamableHTTPClientTransport(
			mcpUrl,
			transport.WithStreamableHTTPClientOptionHTTPClient(httpClient),
		)

		if err != nil {
			log.Printf("Failed to create transport client: %v", err)
			continue
		}

		// Initialize MCP client
		mcpClient, err := client.NewClient(transportClient)
		if err != nil {
			log.Printf("Failed to create MCP client: %v", err)
			continue
		}

		mcpClients[clientName] = mcpClient
	}

	return mcpClients
}
