# Golang MCP Client

This repository provides a Go-based client library for the [Model Context Protocol (MCP)](https://example.com/)—an open standard that lets you connect AI systems with data sources in a consistent, scalable way. With this client, you can:

- Establish a transport layer using SSE.
- Connect to an MCP server.
- List and call available tools exposed via MCP
- Handle authentication and authoirization with the MCP server

## Overview

MCP aims to solve the “data silo” problem for AI by standardizing how models and data sources communicate. Rather than writing bespoke connectors for every tool, you can build once against MCP and then reuse that connection for many data sources.
After that, you instantiate the client, connect, and call tools to exchange data.

## Prerequisites

- Go 1.21+ (or a more recent version).
- An MCP server to connect to (either local or remote).
- Network connectivity to the MCP server, if you are using SSE or another network-based transport.

## Installation

Install the client package in your Go project:

```bash
git clone [repository-url]
```

## Usage

Below is a simple example that demonstrates how to:

1. Create a transport (SSE in this case).
2. Instantiate an MCP client.
3. Connect the client to the MCP server.
4. List available tools.
5. Call a tool and handle the response.

```go
package main

import (
    "log"
    "mcp_server/pkg/mcp"
    "mcp_server/pkg/transport"
)

func main() {
    // 1. Create transport (SSE example)
    transport := transport.NewSSETransport("http://localhost:8777", map[string]string{
        "Authorization": "Bearer abcd",
    })

    // 2. Instantiate an MCP client
    client := mcp.NewClient("spark", "1.0.0")

    // 3. Connect to the MCP server
    err := client.Connect(transport)
    if err != nil {
        log.Fatal(err)
    }

    // 4. List available tools
    tools, err := client.ListTools()
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Tools: %+v", tools)

    // 5. Call a tool and process its response
    res, err := client.CallTool("query", map[string]interface{}{
        "query": "SELECT * FROM employees",
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Response: %s", *res[0].Text)
}
```

### Choosing a Transport

- **SSE Transport** (`NewSSETransport`): Communicates over HTTP using Server-Sent Events (SSE). Useful when the MCP server is hosted remotely, and you want to stream data or logs over a persistent HTTP connection.
- **Stdio Transport** TBD

### Authentication

If your MCP server requires authentication, you can pass the relevant tokens or credentials in the transport configuration. In the example above, an HTTP Authorization header is set to `Bearer abcd`. Adapt that as needed for your environment.

## Project Structure

Typical structure might look like:

```
.
├── pkg
│   ├── mcp
│   │   ├── client.go        // MCP client logic
│   │   └── ...
│   └── transport
│       ├── sse_transport.go // SSE transport implementation
│       ├── stdio.go         // STDIO transport (if implemented)
│       └── ...
├── cmd
│   └── sample_client
│       └── main.go          // Example usage
└── ...
```

## Contributing

Contributions are welcome! Feel free to open issues or PRs if you have improvements to suggest, bug fixes, or new features to add. Whether it’s fixing documentation, improving the client’s interface, or proposing new transports, all feedback is appreciated.
