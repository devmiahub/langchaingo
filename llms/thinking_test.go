package llms

import (
	"encoding/json"
	"testing"
)

func TestThinkingContent(t *testing.T) {
	t.Run("ThinkingPart creates correct content", func(t *testing.T) {
		tc := ThinkingPart("This is my reasoning")
		if tc.Thinking != "This is my reasoning" {
			t.Errorf("Expected thinking text 'This is my reasoning', got %q", tc.Thinking)
		}
		if tc.Signature != "" {
			t.Errorf("Expected empty signature, got %q", tc.Signature)
		}
	})

	t.Run("ThinkingPartWithSignature creates correct content", func(t *testing.T) {
		tc := ThinkingPartWithSignature("My reasoning", "sig123")
		if tc.Thinking != "My reasoning" {
			t.Errorf("Expected thinking text 'My reasoning', got %q", tc.Thinking)
		}
		if tc.Signature != "sig123" {
			t.Errorf("Expected signature 'sig123', got %q", tc.Signature)
		}
	})

	t.Run("ThinkingContent implements ContentPart", func(t *testing.T) {
		var _ ContentPart = ThinkingContent{}
	})

	t.Run("ThinkingContent String returns thinking text", func(t *testing.T) {
		tc := ThinkingContent{Thinking: "Test thinking", Signature: "sig"}
		if tc.String() != "Test thinking" {
			t.Errorf("Expected String() to return thinking text, got %q", tc.String())
		}
	})
}

func TestThinkingContentMarshalJSON(t *testing.T) {
	t.Run("Marshal with signature", func(t *testing.T) {
		tc := ThinkingContent{
			Thinking:  "Step 1: Analyze the problem",
			Signature: "abc123signature",
		}

		data, err := json.Marshal(tc)
		if err != nil {
			t.Fatalf("Failed to marshal ThinkingContent: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if result["type"] != "thinking" {
			t.Errorf("Expected type 'thinking', got %v", result["type"])
		}

		thinking, ok := result["thinking"].(map[string]any)
		if !ok {
			t.Fatal("Expected 'thinking' field to be a map")
		}

		if thinking["thinking"] != "Step 1: Analyze the problem" {
			t.Errorf("Expected thinking text, got %v", thinking["thinking"])
		}

		if thinking["signature"] != "abc123signature" {
			t.Errorf("Expected signature 'abc123signature', got %v", thinking["signature"])
		}
	})

	t.Run("Marshal without signature", func(t *testing.T) {
		tc := ThinkingContent{
			Thinking: "Just thinking",
		}

		data, err := json.Marshal(tc)
		if err != nil {
			t.Fatalf("Failed to marshal ThinkingContent: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		thinking, ok := result["thinking"].(map[string]any)
		if !ok {
			t.Fatal("Expected 'thinking' field to be a map")
		}

		if _, hasSignature := thinking["signature"]; hasSignature {
			t.Error("Expected no signature field when signature is empty")
		}
	})
}

func TestThinkingContentUnmarshalJSON(t *testing.T) {
	t.Run("Unmarshal with signature", func(t *testing.T) {
		jsonData := `{
			"type": "thinking",
			"thinking": {
				"thinking": "My reasoning process",
				"signature": "verification_token_123"
			}
		}`

		var tc ThinkingContent
		if err := json.Unmarshal([]byte(jsonData), &tc); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if tc.Thinking != "My reasoning process" {
			t.Errorf("Expected thinking 'My reasoning process', got %q", tc.Thinking)
		}

		if tc.Signature != "verification_token_123" {
			t.Errorf("Expected signature 'verification_token_123', got %q", tc.Signature)
		}
	})

	t.Run("Unmarshal without signature", func(t *testing.T) {
		jsonData := `{
			"type": "thinking",
			"thinking": {
				"thinking": "Just a thought"
			}
		}`

		var tc ThinkingContent
		if err := json.Unmarshal([]byte(jsonData), &tc); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if tc.Thinking != "Just a thought" {
			t.Errorf("Expected thinking 'Just a thought', got %q", tc.Thinking)
		}

		if tc.Signature != "" {
			t.Errorf("Expected empty signature, got %q", tc.Signature)
		}
	})

	t.Run("Unmarshal with wrong type fails", func(t *testing.T) {
		jsonData := `{
			"type": "text",
			"thinking": {
				"thinking": "Should fail"
			}
		}`

		var tc ThinkingContent
		err := json.Unmarshal([]byte(jsonData), &tc)
		if err == nil {
			t.Error("Expected error for wrong type, got nil")
		}
	})
}

func TestMessageContentWithThinking(t *testing.T) {
	t.Run("Marshal MessageContent with ThinkingContent part", func(t *testing.T) {
		mc := MessageContent{
			Role: ChatMessageTypeAI,
			Parts: []ContentPart{
				ThinkingContent{
					Thinking:  "Let me think about this",
					Signature: "sig456",
				},
				TextContent{Text: "Here's my answer"},
			},
		}

		data, err := json.Marshal(mc)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		parts, ok := result["parts"].([]any)
		if !ok {
			t.Fatal("Expected 'parts' field to be an array")
		}

		if len(parts) != 2 {
			t.Errorf("Expected 2 parts, got %d", len(parts))
		}
	})

	t.Run("Unmarshal MessageContent with ThinkingContent part", func(t *testing.T) {
		jsonData := `{
			"role": "ai",
			"parts": [
				{
					"type": "thinking",
					"thinking": {
						"thinking": "Processing the request",
						"signature": "token789"
					}
				},
				{
					"type": "text",
					"text": "The answer is 42"
				}
			]
		}`

		var mc MessageContent
		if err := json.Unmarshal([]byte(jsonData), &mc); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if mc.Role != ChatMessageTypeAI {
			t.Errorf("Expected role 'ai', got %q", mc.Role)
		}

		if len(mc.Parts) != 2 {
			t.Errorf("Expected 2 parts, got %d", len(mc.Parts))
		}

		thinking, ok := mc.Parts[0].(ThinkingContent)
		if !ok {
			t.Fatal("Expected first part to be ThinkingContent")
		}

		if thinking.Thinking != "Processing the request" {
			t.Errorf("Expected thinking text 'Processing the request', got %q", thinking.Thinking)
		}

		if thinking.Signature != "token789" {
			t.Errorf("Expected signature 'token789', got %q", thinking.Signature)
		}

		text, ok := mc.Parts[1].(TextContent)
		if !ok {
			t.Fatal("Expected second part to be TextContent")
		}

		if text.Text != "The answer is 42" {
			t.Errorf("Expected text 'The answer is 42', got %q", text.Text)
		}
	})
}

func TestAIChatMessageWithThinking(t *testing.T) {
	t.Run("AIChatMessage with ThinkingContent and ThoughtSignature", func(t *testing.T) {
		msg := AIChatMessage{
			Content:          "The answer is 42",
			ThinkingContent:  "Let me calculate: 6 * 7 = 42",
			ThoughtSignature: "verification_sig",
		}

		if msg.Content != "The answer is 42" {
			t.Errorf("Expected content 'The answer is 42', got %q", msg.Content)
		}

		if msg.ThinkingContent != "Let me calculate: 6 * 7 = 42" {
			t.Errorf("Expected thinking content, got %q", msg.ThinkingContent)
		}

		if msg.ThoughtSignature != "verification_sig" {
			t.Errorf("Expected thought signature, got %q", msg.ThoughtSignature)
		}
	})

	t.Run("AIChatMessage JSON roundtrip with thinking", func(t *testing.T) {
		original := AIChatMessage{
			Content:          "Hello",
			ThinkingContent:  "Reasoning here",
			ThoughtSignature: "sig",
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		var restored AIChatMessage
		if err := json.Unmarshal(data, &restored); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if restored.Content != original.Content {
			t.Errorf("Content mismatch: got %q, want %q", restored.Content, original.Content)
		}

		if restored.ThinkingContent != original.ThinkingContent {
			t.Errorf("ThinkingContent mismatch: got %q, want %q", restored.ThinkingContent, original.ThinkingContent)
		}

		if restored.ThoughtSignature != original.ThoughtSignature {
			t.Errorf("ThoughtSignature mismatch: got %q, want %q", restored.ThoughtSignature, original.ThoughtSignature)
		}
	})
}

func TestContentChoiceWithThinking(t *testing.T) {
	t.Run("ContentChoice contains thinking fields", func(t *testing.T) {
		choice := ContentChoice{
			Content:          "Answer",
			ThinkingContent:  "My reasoning",
			ThoughtSignature: "token123",
			GenerationInfo: map[string]any{
				"ThinkingContent":   "My reasoning",
				"ThoughtSignature":  "token123",
				"ThinkingTokens":    100,
			},
		}

		if choice.ThinkingContent != "My reasoning" {
			t.Errorf("Expected thinking content 'My reasoning', got %q", choice.ThinkingContent)
		}

		if choice.ThoughtSignature != "token123" {
			t.Errorf("Expected thought signature 'token123', got %q", choice.ThoughtSignature)
		}

		if choice.GenerationInfo["ThinkingTokens"] != 100 {
			t.Errorf("Expected thinking tokens 100, got %v", choice.GenerationInfo["ThinkingTokens"])
		}
	})
}

