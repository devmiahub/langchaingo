//nolint:all
package googleai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tmc/langchaingo/llms"
)

// Gemini3RestClient provides REST API access for Gemini 3 models that require
// thought signatures for function calling.
type Gemini3RestClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewGemini3RestClient creates a new REST client for Gemini 3 API.
func NewGemini3RestClient(apiKey string) *Gemini3RestClient {
	return &Gemini3RestClient{
		apiKey:     apiKey,
		baseURL:    "https://generativelanguage.googleapis.com/v1beta",
		httpClient: http.DefaultClient,
	}
}

// REST API request/response types for Gemini 3 with thought signature support
type gemini3Request struct {
	Contents          []gemini3Content         `json:"contents"`
	Tools             []gemini3Tool            `json:"tools,omitempty"`
	GenerationConfig  *gemini3GenerationConfig `json:"generationConfig,omitempty"`
	SystemInstruction *gemini3Content          `json:"systemInstruction,omitempty"`
}

type gemini3Content struct {
	Role  string        `json:"role"`
	Parts []gemini3Part `json:"parts"`
}

type gemini3Part struct {
	Text             string                   `json:"text,omitempty"`
	FunctionCall     *gemini3FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *gemini3FunctionResponse `json:"functionResponse,omitempty"`
	// ThoughtSignature is a sibling field of functionCall in the part
	// It must be included when sending function calls back to Gemini 3
	ThoughtSignature string `json:"thoughtSignature,omitempty"`
}

type gemini3FunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type gemini3FunctionResponse struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type gemini3Tool struct {
	FunctionDeclarations []gemini3FunctionDeclaration `json:"functionDeclarations"`
}

type gemini3FunctionDeclaration struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type gemini3GenerationConfig struct {
	Temperature     float32  `json:"temperature,omitempty"`
	TopP            float32  `json:"topP,omitempty"`
	TopK            int32    `json:"topK,omitempty"`
	MaxOutputTokens int32    `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type gemini3Response struct {
	Candidates    []gemini3Candidate `json:"candidates"`
	UsageMetadata *gemini3Usage      `json:"usageMetadata,omitempty"`
}

type gemini3Candidate struct {
	Content       *gemini3Content `json:"content"`
	FinishReason  string          `json:"finishReason"`
	SafetyRatings []any           `json:"safetyRatings,omitempty"`
}

type gemini3Usage struct {
	PromptTokenCount     int32 `json:"promptTokenCount"`
	CandidatesTokenCount int32 `json:"candidatesTokenCount"`
	TotalTokenCount      int32 `json:"totalTokenCount"`
}

// GenerateContent calls the Gemini 3 REST API with thought signature support.
func (c *Gemini3RestClient) GenerateContent(
	ctx context.Context,
	model string,
	messages []llms.MessageContent,
	opts *llms.CallOptions,
) (*llms.ContentResponse, error) {
	// Build the request
	req, err := c.buildRequest(messages, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Make the HTTP request
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, model, c.apiKey)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug: uncomment to see request body
	// fmt.Printf("DEBUG Request body: %s\n", string(reqBody))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Debug: uncomment to see raw response
	// fmt.Printf("DEBUG Raw response: %s\n", string(body))

	// Parse the response
	var geminiResp gemini3Response
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to ContentResponse
	return c.convertResponse(&geminiResp)
}

func (c *Gemini3RestClient) buildRequest(messages []llms.MessageContent, opts *llms.CallOptions) (*gemini3Request, error) {
	req := &gemini3Request{
		Contents: make([]gemini3Content, 0, len(messages)),
	}

	// Convert tools
	if len(opts.Tools) > 0 {
		funcDecls := make([]gemini3FunctionDeclaration, 0, len(opts.Tools))
		for _, tool := range opts.Tools {
			if tool.Type == "function" && tool.Function != nil {
				params, _ := tool.Function.Parameters.(map[string]any)
				funcDecls = append(funcDecls, gemini3FunctionDeclaration{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  params,
				})
			}
		}
		if len(funcDecls) > 0 {
			req.Tools = []gemini3Tool{{FunctionDeclarations: funcDecls}}
		}
	}

	// Generation config
	req.GenerationConfig = &gemini3GenerationConfig{
		Temperature:     float32(opts.Temperature),
		TopP:            float32(opts.TopP),
		TopK:            int32(opts.TopK),
		MaxOutputTokens: int32(opts.MaxTokens),
		StopSequences:   opts.StopWords,
	}

	// Convert messages
	for _, msg := range messages {
		role := convertRoleToGemini3(msg.Role)
		if role == "system" {
			// System messages go to SystemInstruction
			parts := make([]gemini3Part, 0, len(msg.Parts))
			for _, p := range msg.Parts {
				if tc, ok := p.(llms.TextContent); ok {
					parts = append(parts, gemini3Part{Text: tc.Text})
				}
			}
			if len(parts) > 0 {
				req.SystemInstruction = &gemini3Content{
					Role:  "user", // System instruction uses user role in content
					Parts: parts,
				}
			}
			continue
		}

		content := gemini3Content{
			Role:  role,
			Parts: make([]gemini3Part, 0, len(msg.Parts)),
		}

		for _, p := range msg.Parts {
			switch part := p.(type) {
			case llms.TextContent:
				content.Parts = append(content.Parts, gemini3Part{Text: part.Text})
			case llms.ToolCall:
				args := make(map[string]any)
				if part.FunctionCall != nil {
					_ = json.Unmarshal([]byte(part.FunctionCall.Arguments), &args)
				}
				fc := &gemini3FunctionCall{
					Name: part.FunctionCall.Name,
					Args: args,
				}
				// Include thought signature as sibling field of functionCall
				// This is required for Gemini 3 Pro
				content.Parts = append(content.Parts, gemini3Part{
					FunctionCall:     fc,
					ThoughtSignature: part.ThoughtSignature,
				})
			case llms.ToolCallResponse:
				content.Parts = append(content.Parts, gemini3Part{
					FunctionResponse: &gemini3FunctionResponse{
						Name: part.Name,
						Response: map[string]any{
							"response": part.Content,
						},
					},
				})
			}
		}

		if len(content.Parts) > 0 {
			req.Contents = append(req.Contents, content)
		}
	}

	return req, nil
}

func (c *Gemini3RestClient) convertResponse(resp *gemini3Response) (*llms.ContentResponse, error) {
	contentResp := &llms.ContentResponse{
		Choices: make([]*llms.ContentChoice, 0, len(resp.Candidates)),
	}

	for _, candidate := range resp.Candidates {
		choice := &llms.ContentChoice{
			StopReason:     candidate.FinishReason,
			GenerationInfo: make(map[string]any),
			ToolCalls:      make([]llms.ToolCall, 0),
		}

		if resp.UsageMetadata != nil {
			choice.GenerationInfo["PromptTokens"] = resp.UsageMetadata.PromptTokenCount
			choice.GenerationInfo["CompletionTokens"] = resp.UsageMetadata.CandidatesTokenCount
			choice.GenerationInfo["TotalTokens"] = resp.UsageMetadata.TotalTokenCount
		}

		if candidate.Content != nil {
			var textBuilder bytes.Buffer
			var lastThoughtSig string
			var toolCallIdx int

			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					textBuilder.WriteString(part.Text)
				}
				if part.FunctionCall != nil {
					args, _ := json.Marshal(part.FunctionCall.Args)
					toolCall := llms.ToolCall{
						ID:   fmt.Sprintf("call_%d", toolCallIdx),
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      part.FunctionCall.Name,
							Arguments: string(args),
						},
					}
					// Capture thought signature from sibling field in the same part
					if part.ThoughtSignature != "" {
						toolCall.ThoughtSignature = part.ThoughtSignature
						lastThoughtSig = part.ThoughtSignature
					}
					choice.ToolCalls = append(choice.ToolCalls, toolCall)
					toolCallIdx++
				}
			}

			choice.Content = textBuilder.String()

			// Store thought signature at choice level
			if lastThoughtSig != "" {
				choice.ThoughtSignature = lastThoughtSig
			}
		}

		// Set FuncCall for backwards compatibility
		if len(choice.ToolCalls) > 0 {
			choice.FuncCall = choice.ToolCalls[0].FunctionCall
		}

		contentResp.Choices = append(contentResp.Choices, choice)
	}

	return contentResp, nil
}

func convertRoleToGemini3(role llms.ChatMessageType) string {
	switch role {
	case llms.ChatMessageTypeSystem:
		return "system"
	case llms.ChatMessageTypeHuman:
		return "user"
	case llms.ChatMessageTypeAI:
		return "model"
	case llms.ChatMessageTypeTool:
		return "user" // Tool responses are sent as user messages in Gemini API
	default:
		return "user"
	}
}

// IsGemini3Model checks if the model name indicates a Gemini 3 model that
// requires thought signatures for function calling.
func IsGemini3Model(modelName string) bool {
	// Gemini 3 models have "gemini-3" in the name
	return len(modelName) >= 8 && (modelName[:8] == "gemini-3" ||
		// Also check for preview versions
		(len(modelName) >= 6 && modelName[:6] == "gemini" &&
			(len(modelName) >= 12 && modelName[7:12] == "3-pro" ||
				len(modelName) >= 14 && modelName[7:14] == "3-flash")))
}
