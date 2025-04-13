package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mcp_server/pkg/messages"
	"net/http"
	"regexp"
	"time"
)

type SSETransport struct {
	Url             string
	Headers         map[string]string
	Endpoint        string
	MessageCallback func(message messages.Response)
}

func NewSSETransport(url string, headers map[string]string) *SSETransport {
	return &SSETransport{
		Url:     url,
		Headers: headers,
	}
}

func (t *SSETransport) Start() error {
	req, err := http.NewRequest(http.MethodGet, t.Url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	for hName, hValue := range t.Headers {
		req.Header.Add(hName, hValue)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("expected status code 200 but got %d", resp.StatusCode)
	}

	go func() {
		respBuf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(respBuf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				log.Printf("Error reading response body: %v", err)
				break
			}

			msg := string(respBuf[:n])

			extractedMessage, err := extractSSEEvent(msg)
			if err != nil {
				log.Printf("Error extracting SSE message: %v", err)
				break
			}

			if extractedMessage.Event == "endpoint" {
				t.Endpoint = extractedMessage.Data
			}

			if extractedMessage.Event == "message" {
				var msg messages.Response
				err = json.Unmarshal([]byte(extractedMessage.Data), &msg)
				if err != nil {
					log.Printf("Error unmarshalling message: %v", err)
					break
				}
				t.MessageCallback(msg)
			}
		}
	}()

	return nil
}

func (t *SSETransport) OnMessage(callback func(message messages.Response)) {
	t.MessageCallback = callback
}

func (t *SSETransport) Send(message messages.Request) error {
	for {
		if t.Endpoint != "" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	requestBodyBuffer := bytes.NewBuffer(nil)

	err := json.NewEncoder(requestBodyBuffer).Encode(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", t.Url, t.Endpoint), requestBodyBuffer)
	if err != nil {
		return err
	}

	client := &http.Client{}
	for hName, hValue := range t.Headers {
		req.Header.Add(hName, hValue)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 202 {
		return fmt.Errorf("expected status code 202 but got %d", resp.StatusCode)
	}

	return nil
}

var sseRegex = regexp.MustCompile(`event:\s*(.+)\ndata:\s*(.+)\n\n`)

func extractSSEEvent(input string) (SSEMessage, error) {
	matches := sseRegex.FindStringSubmatch(input)
	if len(matches) != 3 {
		return SSEMessage{}, errors.New("invalid SSE message format")
	}

	return SSEMessage{
		Event: matches[1],
		Data:  matches[2],
	}, nil
}

type SSEMessage struct {
	Event string
	Data  string
}
