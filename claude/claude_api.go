package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type ClaudeApi struct {
	apiKey      string
	client      *http.Client
	model       string
	maxTokens   int
	temperature float32
}

const ROLE_USER = "user"
const ROLE_ASSISTANT = "assistant"

const CLAUDE_MODEL = "claude-sonnet-4-0"
const CLAUDE_45_MODEL = "claude-sonnet-4-5-20250929"
const CLAUDE_API_URL = "https://api.anthropic.com/v1/messages"
const DEFAULT_TEMPERATURE = 0.01
const MAX_TOKENS = 64000
const DEFAULT_MAX_TOKENS = 1000

type ClaudeMessageRequest struct {
	Model         string         `json:"model"`
	System        string         `json:"system"`
	Messages      ClaudeMessages `json:"messages"`
	MaxTokens     int            `json:"max_tokens"`
	Temperature   float32        `json:"temperature,omitempty"`
	StopSequences []string       `json:"stop_sequences,omitempty"`
	Thinking      *struct {
		Type         string `json:"type"`
		BudgetTokens int    `json:"budget_tokens,omitempty"`
	} `json:"thinking,omitempty"`
}

type ClaudeMessages []ClaudeMessage

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type ClaudeMessageResponse struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Content      []Content `json:"content"`
	Model        string    `json:"model"`
	StopReason   string    `json:"stop_reason"`
	StopSequence *string   `json:"stop_sequence"`
	Usage        Usage     `json:"usage"`
}

type ClaudeMessageErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
	Type string `json:"type"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func NewClaudeClient(apiKey string, proxyDSN string, defaultModel string) (api *ClaudeApi, err error) {
	transport := &http.Transport{}
	if proxyDSN != "" {
		proxyURL, err := url.Parse(proxyDSN)
		if err != nil {
			return nil, fmt.Errorf("new gruta client proxy dsn error: %s", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Minute,
	}
	api = &ClaudeApi{
		apiKey:      apiKey,
		client:      client,
		model:       defaultModel,
		maxTokens:   DEFAULT_MAX_TOKENS,
		temperature: DEFAULT_TEMPERATURE,
	}
	return api, nil
}

func (c *ClaudeApi) SendMessage(claudeMessages ClaudeMessages, systemMessage string) (*ClaudeMessageResponse, error) {
	log.Printf("🤖 [GRUTA_API] Preparing request - Model: %s, Messages: %d, MaxTokens: %d", c.model, len(claudeMessages), min(c.maxTokens, MAX_TOKENS))
	log.Printf("🤖 [GRUTA_API] System message length: %d characters", len(systemMessage))

	request := ClaudeMessageRequest{
		Model:       c.model,
		System:      systemMessage,
		Messages:    claudeMessages,
		MaxTokens:   min(c.maxTokens, MAX_TOKENS),
		Temperature: c.temperature,
	}

	log.Printf("🤖 [GRUTA_API] Sending request to API...")
	return c.DoRequest(request)
}

func (c *ClaudeApi) DoRequest(request ClaudeMessageRequest) (*ClaudeMessageResponse, error) {
	log.Printf("📤 [GRUTA_API] Marshaling request body...")
	reqBody, err := json.Marshal(request)
	if err != nil {
		log.Printf("❌ [GRUTA_API] Error marshaling request: %v", err)
		return nil, err
	}
	log.Printf("📤 [GRUTA_API] Request body size: %d bytes", len(reqBody))

	requestText := request.System
	for _, msg := range request.Messages {
		requestText += msg.Content
	}

	log.Printf("🌐 [GRUTA_API] Creating HTTP request to %s", CLAUDE_API_URL)
	httpReq, err := http.NewRequest("POST", CLAUDE_API_URL, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("❌ [GRUTA_API] Error creating HTTP request: %v", err)
		return nil, err
	}

	log.Printf("🔑 [GRUTA_API] Setting request headers...")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	log.Printf("📡 [GRUTA_API] Executing HTTP request...")
	startTime := time.Now()
	resp, err := c.client.Do(httpReq)
	requestDuration := time.Since(startTime)
	log.Printf("⏱️ [GRUTA_API] Request completed in %v", requestDuration)

	if err != nil {
		log.Printf("❌ [GRUTA_API] HTTP request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("📥 [GRUTA_API] Received response - Status: %d %s", resp.StatusCode, resp.Status)

	log.Printf("📖 [GRUTA_API] Reading response body...")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [GRUTA_API] Error reading response body: %v", err)
		return nil, err
	}
	log.Printf("📖 [GRUTA_API] Response body size: %d bytes", len(body))

	if resp.StatusCode != 200 {
		log.Printf("❌ [GRUTA_API] Non-200 status code: %d", resp.StatusCode)
		log.Printf("❌ [GRUTA_API] Error response body: %s", string(body))
		log.Printf("❌ [GRUTA_API] Error api key was: %s", c.apiKey[20:])

		if resp.StatusCode == 529 {
			log.Printf("⚠️ [GRUTA_API] API overloaded (529), returning special error")
			return nil, fmt.Errorf("gruta_overloaded_529")
		}

		var respData ClaudeMessageErrorResponse
		err = json.Unmarshal(body, &respData)
		if err != nil {
			log.Printf("❌ [GRUTA_API] Error unmarshaling error response: %v", err)
			return nil, fmt.Errorf("gruta SendMessage status code non 200, %d, unmarshall err: %s, body: %s", resp.StatusCode, err, string(body))
		}
		log.Printf("❌ [GRUTA_API]  API error - Type: %s, Message: %s", respData.Error.Type, respData.Error.Message)
		return nil, fmt.Errorf("SendMessage status not 200(%d) error: message: %s, type: %s", resp.StatusCode, respData.Error.Message, respData.Error.Type)
	}

	log.Printf("🔄 [GRUTA_API] Parsing successful response...")
	var respData ClaudeMessageResponse
	err = json.Unmarshal(body, &respData)
	if err != nil {
		log.Printf("❌ [GRUTA_API] Error unmarshaling response: %v", err)
		log.Printf("❌ [GRUTA_API] Response body: %s", string(body))
		return nil, fmt.Errorf("gruta SendMessage unmarshall err: %s, body: %s", err, string(body))
	}

	log.Printf("✅ [GRUTA_API] Successfully parsed response")
	log.Printf("📊 [GRUTA_API] Token usage - Input: %d, Output: %d", respData.Usage.InputTokens, respData.Usage.OutputTokens)
	log.Printf("🏁 [GRUTA_API] Stop reason: %s", respData.StopReason)
	if len(respData.Content) > 0 {
		log.Printf("📝 [GRUTA_API] Response content length: %d characters", len(respData.Content[0].Text))
	}

	return &respData, nil
}
func (s *ClaudeApi) SetMaxTokens(maxTokens int) {
	s.maxTokens = maxTokens
}
func (s *ClaudeApi) SetTemperature(temp float32) {
	s.temperature = temp
}
