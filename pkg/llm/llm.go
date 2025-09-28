package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms/openai"
)

const (
	maxTokens = 8192
)

type SelectedToolInfo struct {
	ToolName    string         `json:"tool_name"`
	ToolArgs    map[string]any `json:"tool_args"`
	MissingArgs []string       `json:"missing_args"`
}

var (
	llmUrl   = os.Getenv("LLM_URL")
	llmToken = os.Getenv("LLM_TOKEN")
)

var (
	toolSelectionPromptlines = []string{
		"You are a software engineer experienced in developing RESTful applications.",
		"You are familiar with JSON documents and JSON schema.",
		"Using the schema of the Tools provided, decide if a tool from the list can be used to answer the user's query or complete the user's task.",
		"Your response must be a valid JSON string",
		"There are three possible responses:",
		`1. If a matching tool is found, the response should be {"tool_name": <name of the tool found>, "tool_args": <argument to the selected tool>}`,
		`2. If a closely matching tool is available, but some arguments needed are missing then response should be {"tool_name": <name of the tool selected>, "missing_args", <list of missing arguments>}`,
		`3. If no tool can be used to meet the user's query or task, the response should be {"tool_name": "none"}`,
		`4. If the input to the LLM is in JSON format, the response should in natutal language in a readble nice text format.`,
	}

	toolSelectionPrompt = strings.Join(toolSelectionPromptlines, "\n")
)

// Function to select the tool based on the query - Original code

/* Hard coded request to test openai
func SelectTool(ctx context.Context, llm *openai.LLM, toolListSchema string, query string) (*SelectedToolInfo, error) {
	// Hardcoded payload matching Bruno API test
	payload := map[string]interface{}{
		"model":      "gpt-4o",
		"stream":     false,
		"max_tokens": 8192,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `You are a software engineer experienced in developing RESTful applications.
You are familiar with JSON documents and JSON schema.
Using the schema of the Tools provided, decide if a tool from the list can be used to answer the user's query or complete the user's task.
Your response must be a valid JSON string
There are three possible responses:
1. If a matching tool is found, the response should be {"tool_name": <name of the tool found>, "tool_args": <argument to the selected tool>}
2. If a closely matching tool is available, but some arguments needed are missing then response should be {"tool_name": <name of the tool selected>, "missing_args": <list of missing arguments>}
3. If no tool can be used to meet the user's query or task, the response should be {"tool_name": "none"}
4. If the input to the LLM is in JSON format, the response should in natural language in a readable nice text format.`,
			},
			{
				"role":    "user",
				"content": `Tools: {"tools":[{"description":"Tool to perform arithmetic operation on two numbers.\nValid operations are add, subtract, multiply, divide, and exponentiate.\nThe input to this tool should be two floating point numbers","inputSchema":{"type":"object","properties":{"a":{"type":"number","description":"First operand of the operation"},"b":{"type":"number","description":"Second operand of the operation"},"operation":{"type":"string","description":"Operation to be performed on the two numbers"}},"required":["operation","a","b"]},"name":"Calculator"}]}`,
			},
			{
				"role":    "user",
				"content": "what is minikube?",
			},
		},
	}

	// Marshal payload
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Prepare request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "<<API-KEY>>")

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Extract JSON from assistant message
	var raw map[string]interface{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	// Parse assistant message
	choices, ok := raw["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := message["content"].(string)

	return buildToolSelectionResponse(content)
}
*/

func SelectTool(ctx context.Context, llm *openai.LLM, toolListSchema string, query string) (*SelectedToolInfo, error) {
	// Construct raw OpenAI-compatible message payload
	messages := []map[string]string{
		{
			"role":    "system",
			"content": toolSelectionPrompt,
		},
		{
			"role":    "user",
			"content": fmt.Sprintf("Tools: %s", toolListSchema),
		},
		{
			"role":    "user",
			"content": query,
		},
	}

	// Prepare request body
	payload := map[string]interface{}{
		"model":      "gpt-4o",
		"messages":   messages,
		"max_tokens": maxTokens,
		"stream":     false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", llmUrl, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", llmToken))

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var raw map[string]interface{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	choices, ok := raw["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := message["content"].(string)

	// Convert assistant response to SelectedToolInfo
	return buildToolSelectionResponse(content)
}

var (
	outputFormatPromptlines = []string{
		"You are a skilled text formatter.",
		"Use the provided Text to format a response. If the Text contains a list of items, output should be a numbered list of items.",
		"if the provided Text has icons or images or picture, keep them as-is in their respective context",
	}

	outputFormatPrompt = strings.Join(outputFormatPromptlines, "\n")
)

func FormatOutput(ctx context.Context, llm *openai.LLM, toolName, output, input string) (string, error) {
	// Construct OpenAI-compatible message payload
	messages := []map[string]string{
		{
			"role":    "system",
			"content": outputFormatPrompt,
		},
		{
			"role":    "user",
			"content": fmt.Sprintf("Text: %s", output),
		},
	}

	// Prepare request body
	payload := map[string]interface{}{
		"model":      "gpt-4o",
		"messages":   messages,
		"max_tokens": maxTokens,
		"stream":     false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", llmUrl, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", llmToken))

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var raw map[string]interface{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return "", fmt.Errorf("invalid JSON response: %w", err)
	}

	choices, ok := raw["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := message["content"].(string)

	return content, nil
}

func GenericResponse(ctx context.Context, llm *openai.LLM, query string) (string, error) {
	// Construct raw OpenAI-compatible message payload
	messages := []map[string]string{
		{
			"role":    "user",
			"content": fmt.Sprintf("Query: %s", query),
		},
	}

	// Prepare request body
	payload := map[string]interface{}{
		"model":      "gpt-4o",
		"messages":   messages,
		"max_tokens": maxTokens,
		"stream":     false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", llmUrl, strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", llmToken))

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var raw map[string]interface{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return "", fmt.Errorf("invalid JSON response: %w", err)
	}

	choices, ok := raw["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := message["content"].(string)

	return content, nil
}

func buildToolSelectionResponse(llmResp string) (*SelectedToolInfo, error) {
	jsonDoc, found := extractJson(llmResp)

	if !found {
		return nil, fmt.Errorf("response is not json document.\nOuput from llm:\n%s", llmResp)
	}

	toolInfo := SelectedToolInfo{}

	if err := json.Unmarshal([]byte(jsonDoc), &toolInfo); err != nil {
		fmt.Printf("DBG-llm.go> %s\nErr:%v\n", jsonDoc, err)
		return nil, fmt.Errorf("llm unable to select a tool.\nOuput from llm:\n%s", llmResp)
	}

	return &toolInfo, nil
}

var (
	promptlines = []string{
		"You are a software engineer experienced in developing RESTful applications.",
		"You are familiar with JSON documents.",
		"Extract the values for arguments 'namespace' and 'service' from the given Description.",
		"Service is also known as Application.",
		"Your response must a valid JSON string.",
		"It is important that the output must contain only the JSON string. Don't generate any verbose text.",
		"There are two possible responses:",
		`1. If both arguments are found, the response should be {"namespace": <extracted value for namespace>, "service": <extracted value for service>}`,
		`2. If some arguments are missing then response should be {"missing_args", <list of missing arguments>}`,
	}

	prompt = strings.Join(promptlines, "\n")
)

func GetRestartArguments(ctx context.Context, woDetails string) (string, error) {
	// Construct OpenAI-compatible message payload
	messages := []map[string]string{
		{
			"role":    "system",
			"content": prompt,
		},
		{
			"role":    "user",
			"content": fmt.Sprintf("Description: %s", woDetails),
		},
	}

	// Prepare request body
	payload := map[string]interface{}{
		"model":      "gpt-4o",
		"messages":   messages,
		"max_tokens": maxTokens,
		"stream":     false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", llmUrl, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", llmToken))

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var raw map[string]interface{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return "", fmt.Errorf("invalid JSON response: %w", err)
	}

	choices, ok := raw["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := message["content"].(string)

	// Extract JSON from response content
	output, _ := extractJson(content)

	return output, nil
}

func extractJson(s string) (string, bool) {
	var result string
	start := strings.Index(s, "{")

	if start < 0 {
		return result, false
	}

	end := strings.LastIndex(s, "}")

	if start < 0 {
		return result, false
	}

	return s[start : end+1], true
}
