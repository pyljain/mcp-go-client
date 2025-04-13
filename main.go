package main

import (
	"log"
	"mcp_server/pkg/mcp"
	"mcp_server/pkg/transport"
)

func main() {
	transport := transport.NewSSETransport("http://localhost:8777", map[string]string{
		"Authorization": "Bearer abcd",
	})

	client := mcp.NewClient("spark", "1.0.0")

	err := client.Connect(transport)
	if err != nil {
		log.Fatal(err)
	}

	tools, err := client.ListTools()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Tools are %+v", tools)

	res, err := client.CallTool("query", map[string]interface{}{
		"query": "SELECT * FROM employees",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response is %s", *res[0].Text)
}
