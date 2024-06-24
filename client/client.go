package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/abayomipopoola/debugr/utils"
)

const (
	apiURL = "https://api.anthropic.com/v1/messages"
	model  = "claude-3-5-sonnet-20240620"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
}

type Action struct {
	Type    string
	Content string
	Path    string
}

type FileContext struct {
	Files []File
}

type File struct {
	Path     string
	Content  string
	Language string
}

// Structs for request and response handling
type request struct {
	Model     string    `json:"model"`
	System    string    `json:"system"`
	Messages  []message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type message struct {
	Role    string    `json:"role"`
	Content []content `json:"content"`
}

type content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type response struct {
	Content []content `json:"content"`
}

func NewClient(key string) (*Client, error) {
	return &Client{
		apiKey: key,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (c *Client) Prompt(prompt string, context *FileContext) ([]Action, error) {
	utils.Log("Sending prompt to AI: %s", prompt)

	systemPrompt := `You are a helpful debugging and testing assistant. 
	Provide concise, actionable suggestions for debugging or testing the described issue. 
	Each suggestion should be a separate, executable command, a clear instruction, or a file operation.
	For file operations, use the following format:
	CREATE_FILE:<path>:<content>
	or
	MODIFY_FILE:<path>:<content>
	When providing code, ensure it is complete, well-formatted, and includes proper indentation.
	If you can't provide a specific action, suggest a general approach or next step.`

	userMessage := prompt
	if context != nil && len(context.Files) > 0 {
		var contextBuilder strings.Builder
		contextBuilder.WriteString("Context:\n")
		for _, file := range context.Files {
			contextBuilder.WriteString(fmt.Sprintf("File: %s\nLanguage: %s\nContent:\n```\n%s\n```\n\n",
				file.Path, file.Language, file.Content))
		}
		contextBuilder.WriteString("Request: " + prompt)
		userMessage = contextBuilder.String()
	}

	messages := []message{
		{
			Role: "user",
			Content: []content{
				{
					Type: "text",
					Text: userMessage,
				},
			},
		},
	}

	resp, err := c.do(&request{
		Model:     model,
		System:    systemPrompt,
		MaxTokens: 2048,
		Messages:  messages,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Content) == 0 {
		return nil, nil
	}

	actions := parseActions(resp.Content[0].Text)
	utils.Log("Received %d actions", len(actions))
	return actions, nil
}

func (c *Client) do(r *request) (*response, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	utils.Log("Sending request to API: %s", string(data))

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	utils.Log("Response status: %s", resp.Status)
	utils.Log("Response body: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &apiResp, nil
}

func parseActions(text string) []Action {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var actions []Action
	var currentAction Action
	var codeBlock strings.Builder
	inCodeBlock := false
	inFileAction := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "CREATE_FILE:") || strings.HasPrefix(line, "MODIFY_FILE:") {
			if inFileAction {
				currentAction.Content = codeBlock.String()
				actions = append(actions, currentAction)
				codeBlock.Reset()
			}
			parts := strings.SplitN(line, ":", 3)
			if len(parts) >= 2 {
				currentAction = Action{
					Type: parts[0],
					Path: parts[1],
				}
				inFileAction = true
				inCodeBlock = false
			}
		} else if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			if !inCodeBlock && inFileAction {
				currentAction.Content = codeBlock.String()
				actions = append(actions, currentAction)
				codeBlock.Reset()
				inFileAction = false
			}
		} else if inCodeBlock || inFileAction {
			codeBlock.WriteString(line + "\n")
		} else if strings.HasPrefix(line, "$") || strings.HasPrefix(line, "go ") {
			actions = append(actions, Action{
				Type:    "COMMAND",
				Content: strings.TrimPrefix(line, "$ "),
			})
		} else {
			actions = append(actions, Action{
				Type:    "EXPLANATION",
				Content: line,
			})
		}
	}

	if inFileAction {
		currentAction.Content = codeBlock.String()
		actions = append(actions, currentAction)
	}

	return actions
}
