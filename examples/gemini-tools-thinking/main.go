package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/devmiahub/langchaingo/llms"
	"github.com/devmiahub/langchaingo/llms/googleai"
)

// WeatherTool simulates getting weather information
type WeatherInput struct {
	Location string `json:"location"`
	Unit     string `json:"unit,omitempty"`
}

func getWeather(location, unit string) string {
	if unit == "" {
		unit = "celsius"
	}
	// Simulated weather data
	weatherData := map[string]string{
		"São Paulo": fmt.Sprintf("22°%s, partly cloudy", unit[:1]),
		"New York":  fmt.Sprintf("18°%s, sunny", unit[:1]),
		"Tokyo":     fmt.Sprintf("25°%s, clear", unit[:1]),
		"London":    fmt.Sprintf("15°%s, rainy", unit[:1]),
		"Paris":     fmt.Sprintf("17°%s, cloudy", unit[:1]),
	}

	if weather, ok := weatherData[location]; ok {
		return weather
	}
	return fmt.Sprintf("20°%s, unknown conditions", unit[:1])
}

// CalculatorTool for mathematical operations
type CalculatorInput struct {
	Operation string  `json:"operation"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
}

func calculate(operation string, a, b float64) string {
	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return "Error: division by zero"
		}
		result = a / b
	default:
		return "Error: unknown operation"
	}
	return fmt.Sprintf("%.2f", result)
}

func main() {
	ctx := context.Background()

	// Get API key from environment
	apiKey := ""

	// Create GoogleAI client with Gemini model
	// Available models: gemini-2.0-flash, gemini-2.5-pro-preview-06-05, etc.
	// When Gemini 3.0 is released, change to "gemini-3.0-flash" or similar
	client, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithDefaultModel("gemini-3-pro-preview"), // Use gemini-2.5-pro-preview-06-05 for latest
	)
	if err != nil {
		log.Fatalf("Failed to create GoogleAI client: %v", err)
	}
	defer client.Close()

	// Check if model supports reasoning
	fmt.Printf("Model supports reasoning: %v\n\n", client.SupportsReasoning())

	// Define tools
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city name, e.g. São Paulo, New York",
						},
						"unit": map[string]any{
							"type":        "string",
							"enum":        []string{"celsius", "fahrenheit"},
							"description": "The temperature unit",
						},
					},
					"required": []string{"location"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculator",
				Description: "Perform basic mathematical operations",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"operation": map[string]any{
							"type":        "string",
							"enum":        []string{"add", "subtract", "multiply", "divide"},
							"description": "The mathematical operation to perform",
						},
						"a": map[string]any{
							"type":        "number",
							"description": "The first number",
						},
						"b": map[string]any{
							"type":        "number",
							"description": "The second number",
						},
					},
					"required": []string{"operation", "a", "b"},
				},
			},
		},
	}

	// Initial user message
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What's the weather in Tokyo? Also, what is 25 multiplied by 17? Think step by step."),
			},
		},
	}

	fmt.Println("=== Testing Gemini with Tools and Thinking ===")
	fmt.Println("User: What's the weather in Tokyo? Also, what is 25 multiplied by 17? Think step by step.")
	fmt.Println()

	// First call - model should request tool calls
	resp, err := client.GenerateContent(ctx, messages,
		llms.WithTools(tools),
		llms.WithMaxTokens(1024),
	)
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	// Process response
	for i, choice := range resp.Choices {
		fmt.Printf("--- Choice %d ---\n", i)

		// Check for thinking content (Gemini 3.0 feature)
		if choice.ThinkingContent != "" {
			fmt.Printf("🧠 Thinking: %s\n", choice.ThinkingContent)
		}
		if choice.ThoughtSignature != "" {
			fmt.Printf("🔐 Thought Signature: %s...\n", truncate(choice.ThoughtSignature, 50))
		}

		// Check metadata for thinking info
		if genInfo := choice.GenerationInfo; genInfo != nil {
			if tc, ok := genInfo["ThinkingContent"].(string); ok && tc != "" {
				fmt.Printf("📝 Thinking (from metadata): %s\n", tc)
			}
			if ts, ok := genInfo["ThoughtSignature"].(string); ok && ts != "" {
				fmt.Printf("🔑 Signature (from metadata): %s...\n", truncate(ts, 50))
			}
		}

		// Process tool calls
		if len(choice.ToolCalls) > 0 {
			fmt.Printf("\n📞 Tool Calls: %d\n", len(choice.ToolCalls))

			// Add assistant message with tool calls to history
			aiParts := []llms.ContentPart{}
			// Note: We don't include ThinkingContent here because the Google API
			// doesn't accept it back in subsequent requests
			if choice.Content != "" {
				aiParts = append(aiParts, llms.TextPart(choice.Content))
			}
			// Include tool calls with their thought signatures (required for Gemini 3)
			for _, tc := range choice.ToolCalls {
				// Log thought signature if present
				if tc.ThoughtSignature != "" {
					fmt.Printf("\n  🔐 Tool call has thought signature (len=%d)\n", len(tc.ThoughtSignature))
				}
				aiParts = append(aiParts, tc)
			}
			messages = append(messages, llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: aiParts,
			})

			// Collect all tool responses in a single message (Google API requirement)
			toolResponseParts := []llms.ContentPart{}

			// Execute each tool call
			for _, toolCall := range choice.ToolCalls {
				fmt.Printf("\n  🔧 Tool: %s\n", toolCall.FunctionCall.Name)
				fmt.Printf("     Arguments: %s\n", toolCall.FunctionCall.Arguments)

				var result string
				switch toolCall.FunctionCall.Name {
				case "get_weather":
					var input WeatherInput
					if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &input); err != nil {
						result = fmt.Sprintf("Error parsing arguments: %v", err)
					} else {
						result = getWeather(input.Location, input.Unit)
					}
				case "calculator":
					var input CalculatorInput
					if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &input); err != nil {
						result = fmt.Sprintf("Error parsing arguments: %v", err)
					} else {
						result = calculate(input.Operation, input.A, input.B)
					}
				default:
					result = "Unknown tool"
				}

				fmt.Printf("     Result: %s\n", result)

				// Add tool response to parts collection
				// Note: Google API doesn't use ToolCallID, only Name and Content
				toolResponseParts = append(toolResponseParts, llms.ToolCallResponse{
					Name:    toolCall.FunctionCall.Name,
					Content: result,
				})
			}

			// Add all tool responses as a single message (required by Google API)
			messages = append(messages, llms.MessageContent{
				Role:  llms.ChatMessageTypeTool,
				Parts: toolResponseParts,
			})

			// Make follow-up call with tool results
			fmt.Println("\n=== Follow-up call with tool results ===")

			// Write debug info to file
			debugFile, _ := os.Create("debug.txt")
			fmt.Fprintf(debugFile, "Sending %d messages\n", len(messages))
			for i, msg := range messages {
				fmt.Fprintf(debugFile, "Message %d: Role=%s, Parts=%d\n", i, msg.Role, len(msg.Parts))
				for j, part := range msg.Parts {
					fmt.Fprintf(debugFile, "  Part %d: %T\n", j, part)
				}
			}
			debugFile.Close()

			resp2, err := client.GenerateContent(ctx, messages,
				llms.WithTools(tools),
				llms.WithMaxTokens(1024),
			)
			if err != nil {
				log.Printf("ERROR: %v\n", err)
				return
			}

			for _, choice2 := range resp2.Choices {
				// Check for thinking in follow-up
				if choice2.ThinkingContent != "" {
					fmt.Printf("🧠 Thinking: %s\n", choice2.ThinkingContent)
				}

				fmt.Printf("\n🤖 Assistant: %s\n", choice2.Content)

				// Print token usage
				if genInfo := choice2.GenerationInfo; genInfo != nil {
					fmt.Println("\n📊 Token Usage:")
					if pt, ok := genInfo["PromptTokens"].(int32); ok {
						fmt.Printf("   Prompt tokens: %d\n", pt)
					}
					if ct, ok := genInfo["CompletionTokens"].(int32); ok {
						fmt.Printf("   Completion tokens: %d\n", ct)
					}
					if tt, ok := genInfo["TotalTokens"].(int32); ok {
						fmt.Printf("   Total tokens: %d\n", tt)
					}
				}
			}
		} else {
			// No tool calls, just show content
			fmt.Printf("\n🤖 Assistant: %s\n", choice.Content)
		}
	}

	fmt.Println("\n=== Test Complete ===")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
