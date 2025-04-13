package mcp

import (
	"fmt"
	"mcp_server/pkg/messages"
	"mcp_server/pkg/transport"
	"sync"
	"sync/atomic"
)

type Client struct {
	transport            transport.Transport
	Name                 string
	Version              string
	Counter              atomic.Uint32
	messageIdLedger      map[uint32]chan messages.Response // 123 => {}
	messageIdMutex       sync.Mutex
	pendingMessages      map[uint32]messages.Response
	pendingMessagesMutex sync.Mutex
}

func NewClient(name string, version string) *Client {
	return &Client{
		Name:            name,
		Version:         version,
		messageIdLedger: map[uint32]chan messages.Response{},
		pendingMessages: map[uint32]messages.Response{},
	}
}

func (c *Client) OnMessage(messageFromServer messages.Response) {
	c.messageIdMutex.Lock()
	defer c.messageIdMutex.Unlock()
	channel, exists := c.messageIdLedger[uint32(messageFromServer.Id)]
	if !exists {
		c.pendingMessagesMutex.Lock()
		c.pendingMessages[uint32(messageFromServer.Id)] = messageFromServer
		c.pendingMessagesMutex.Unlock()
		// log.Printf("Received unexpected message from server: %+v", messageFromServer)
		return
	}

	channel <- messageFromServer
}
func (c *Client) Connect(transport transport.Transport) error {
	initialisationMessageId := c.Counter.Add(1)
	c.transport = transport
	err := c.transport.Start()
	if err != nil {
		return err
	}

	go c.transport.OnMessage(c.OnMessage)

	err = c.transport.Send(messages.Request{
		JsonRPC: "2.0",
		Id:      int(initialisationMessageId),
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    c.Name,
				"version": c.Version,
			},
		},
	})
	if err != nil {
		return err
	}

	// Wait for server response for initialisation
	c.waitForResponse(initialisationMessageId)

	return nil
}

func (c *Client) waitForResponse(messageID uint32) messages.Response {
	responseChannel := make(chan messages.Response)
	defer close(responseChannel)

	// Check if pending messages exist
	c.pendingMessagesMutex.Lock()
	if pendingMessage, exists := c.pendingMessages[messageID]; exists {
		c.pendingMessagesMutex.Unlock()
		// Delete from pending messages
		delete(c.pendingMessages, messageID)

		return pendingMessage
	}
	c.pendingMessagesMutex.Unlock()

	c.messageIdMutex.Lock()
	c.messageIdLedger[messageID] = responseChannel
	c.messageIdMutex.Unlock()

	return <-responseChannel
}

func (c *Client) CallTool(toolName string, arguments map[string]interface{}) ([]ToolCallResponseResult, error) {
	toolCallMessageId := c.Counter.Add(1)

	err := c.transport.Send(messages.Request{
		JsonRPC: "2.0",
		Id:      int(toolCallMessageId),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	})
	if err != nil {
		return nil, err
	}

	toolCallResponse := c.waitForResponse(toolCallMessageId)

	if toolCallResponse.Error != nil {
		return nil, fmt.Errorf("error: %s", toolCallResponse.Error.Message)
	}

	result := []ToolCallResponseResult{}
	toolCallResponseString := toolCallResponse.Result["content"].([]interface{})
	for _, ct := range toolCallResponseString {
		ct := ct.(map[string]interface{})

		text := ct["text"]
		data := ct["data"]
		mimeType := ct["mimeType"]

		result = append(result, ToolCallResponseResult{
			Type:     ct["type"].(string),
			Text:     convertInterfaceToStrPtr(text),
			Data:     convertInterfaceToStrPtr(data),
			MimeType: convertInterfaceToStrPtr(mimeType),
		})
	}

	return result, nil

}

func convertInterfaceToStrPtr(i interface{}) *string {
	if i == nil {
		return nil
	}
	str := i.(string)
	return &str
}

func (c *Client) ListTools() ([]Tool, error) {
	toolsListMessageId := c.Counter.Add(1)

	err := c.transport.Send(messages.Request{
		JsonRPC: "2.0",
		Id:      int(toolsListMessageId),
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	})
	if err != nil {
		return nil, err
	}

	resp := c.waitForResponse(toolsListMessageId)

	if resp.Error != nil {
		return nil, fmt.Errorf("error: %s", resp.Error.Message)
	}

	var tools []Tool

	for _, v := range resp.Result["tools"].([]interface{}) {
		tool := v.(map[string]interface{})
		tools = append(tools, Tool{
			Name:        tool["name"].(string),
			Description: tool["description"].(string),
			InputSchema: tool["inputSchema"].(map[string]interface{}),
		})
	}

	return tools, nil
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type ToolCallResponseResult struct {
	Type     string  `json:"type"`
	Text     *string `json:"text,omitempty"`
	Data     *string `json:"data,omitempty"`
	MimeType *string `json:"mimeType,omitempty"`
}
