package main

import (
	"context"
	"flag"
	"log"
	"uf/mcp/mcp-server/tools"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
)

func main() {
	// Get MCP Server instance...
	mcpServer := getMcpServer()

	// Register AddNumbers tools
	registerTools(mcpServer)

	// start mcp Server
	go func() {
		if err := mcpServer.Run(); err != nil {
			log.Fatalf("MCP server failed to start: %v", err)
		}
	}()

	defer mcpServer.Shutdown(context.Background())

	// Add logic to listen for kill signal & terminate the program
	waitChl := make(chan struct{})
	<-waitChl

}

func registerTools(mcpServer *server.Server) {

	mcpServer.RegisterTool(tools.GetCalculatorTool())
	mcpServer.RegisterTool(tools.GetServiceFinderTool())
	mcpServer.RegisterTool(tools.GetDeploymentFinderTool())
	mcpServer.RegisterTool(tools.GetNamespaceFinderTool())
	mcpServer.RegisterTool(tools.GetIngressFinderTool())
	mcpServer.RegisterTool(tools.GetServiceRestarterTool())
	mcpServer.RegisterTool(tools.GetPodCpuMemoryViewerTool())
}

func getMcpServer() *server.Server {
	// define flag variables for command line arguments
	var addr string
	var endpoint string

	flag.StringVar(&addr, "addr", ":9090", "listen address")
	flag.StringVar(&endpoint, "endpoint", "/mcp", "endpoint")
	flag.Parse()

	// setup a streamable http server transport
	streamableTransport := transport.NewStreamableHTTPServerTransport(
		addr,
		transport.WithStreamableHTTPServerTransportOptionEndpoint(endpoint),
	)

	// new mcp server
	mcpServer, err := server.NewServer(streamableTransport,
		server.WithServerInfo(protocol.Implementation{
			Name:    "minikube-mcp-server",
			Version: "1.0.0",
		}),
	)

	if err != nil {
		log.Panicf("new mcpServer error: %v", err)
	}

	return mcpServer
}
