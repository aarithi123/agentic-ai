package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"uf/mcp/mcp-client/handlers"
	"uf/mcp/mcp-client/utils"
	"uf/mcp/pkg/mcp"
)

func main() {
	// Declare the static file directory & routes
	http.Handle("/", http.FileServer(http.Dir("./AppRoot/static")))

	// REST API endpoint for chat
	http.HandleFunc("/chat", handlers.ChatHandler)

	// Bring up the http listener
	address := ":8080"
	if a, ok := os.LookupEnv("WEB_PORT"); ok {
		address = fmt.Sprintf(":%s", a)
	}

	log.Printf("Listening on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		panic(err)
	}
}

func init() {
	// Initialize the list of tools avaibale to this application
	// init() method not used in utils package to enable testing of individual functions
	utils.InitializeConfiguration()
	mcp.InitalizeTools()
	//handlers.InitializTemplates()
}
