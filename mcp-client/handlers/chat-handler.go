package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"uf/mcp/mcp-client/utils"
	"uf/mcp/pkg/llm"
	"uf/mcp/pkg/mcp"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Http Handler for chat

func ChatHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("ChatHandler is processing the userMsg")

	// Allow CORS for frontend
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")

	var userMsg Message
	var output string
	var exception string

	if err := json.NewDecoder(r.Body).Decode(&userMsg); err != nil {
		exception = fmt.Sprintf("Invalid JSON payload: %v", err)
		log.Printf("Invalid JSON payload %v", err)

	} else if userMsg.Role != "user" || userMsg.Content == "" {
		exception = fmt.Sprintf("Invalid message format: %v", userMsg)
		log.Printf("Invalid message format")
	}

	if exception == "" { // if initial validation is successful
		ctx := r.Context()
		log.Printf("model: %v", utils.GetModel())

		// Select a tool and get the arguments to the selected tool
		selectToolResp, err := llm.SelectTool(ctx, utils.GetModel(), mcp.GetToolListSchema(), userMsg.Content)
		if err != nil {
			exception = fmt.Sprintf("SelectTool error: %v", err)
			log.Printf("SelectTool error %v", err)
		} else {
			fmt.Printf("DBG ChatHandler>> selectToolResp: %v\n", selectToolResp)

			// if no tool available to answer the query, get a generic response from LLM
			if selectToolResp.ToolName == "none" {
				resp, _ := llm.GenericResponse(ctx, utils.GetModel(), userMsg.Content)
				output = fmt.Sprintf("Currently no tool is implemented to answer the query.\n\nHere is a generic response from LLM:\n%s", resp)
			} else {
				// if some arguments missing, report the missing arguments ...
				//log.Printf("DBG ChatHandler>> selectToolResp: %v\n", selectToolResp)

				if len(selectToolResp.MissingArgs) > 0 {
					exception = fmt.Sprintf("Some arguments are missing: %s", selectToolResp.MissingArgs)
				} else {
					// Call the selected tool
					toolOutput, err := mcp.CallTool(ctx, selectToolResp)
					log.Printf("toolOutput: %v\n", toolOutput)

					//fmt.Printf("DBG ChatHandler>> toolOutput: %v\n", toolOutput)
					if err != nil {
						log.Printf("CallTool error %v", err)
						exception = fmt.Sprintf("CallTool error: %v", err)
					} else {
						// call llm to format the output
						formattedOutput, err := llm.FormatOutput(ctx, utils.GetModel(), selectToolResp.ToolName, toolOutput, userMsg.Content)
						if err != nil {
							exception = fmt.Sprintf("FormatOutput error: %v", err)
							log.Printf("FormatOutput error %v", err)
						} else {
							output = formattedOutput
						}
					}
				}
			}
		}
	}

	content := output

	if exception != "" {
		content = exception
	}

	assistantMsg := Message{
		Role:    "assistant",
		Content: content,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(assistantMsg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		log.Printf("Failed to encode response %v", err)
	}
}
