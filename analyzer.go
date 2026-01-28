package aierror

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type OpenAIStyleResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func Analyze(errorText string) {
	// Load .env if exists (optional)
	_ = godotenv.Load()

	geminiKey := os.Getenv("GEMINI_API_KEY")
	deepseekKey := os.Getenv("DEEPSEEK_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	systemPrompt := `You are a senior Golang debugging expert.
Analyze Go errors and respond ONLY in this format:

ROOT CAUSE:
WHY IT HAPPENED:
HOW TO FIX:
EXAMPLE FIX CODE:`

	userPrompt := "Error:\n" + errorText

	// 🟣 GEMINI (Priority 1)
	if geminiKey != "" {
		fmt.Println("✨ Using Google Gemini")

		url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + geminiKey

		body := map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"parts": []map[string]string{
						{"text": systemPrompt + "\n" + userPrompt},
					},
				},
			},
		}

		jsonBody, _ := json.Marshal(body)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			fmt.Println("Gemini request failed:", err)
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		var result GeminiResponse
		if err := json.Unmarshal(respBody, &result); err != nil || len(result.Candidates) == 0 {
			fmt.Println("Failed to parse Gemini response")
			fmt.Println(string(respBody))
			return
		}

		PrintFormatted(result.Candidates[0].Content.Parts[0].Text)
		return
	}

	// 🔵 DEEPSEEK (Priority 2)
	if deepseekKey != "" {
		fmt.Println("🧠 Using DeepSeek AI")
		callOpenAIStyleAPI(
			"https://api.deepseek.com/v1/chat/completions",
			"deepseek-chat",
			deepseekKey,
			systemPrompt,
			userPrompt,
		)
		return
	}

	// 🟢 OPENAI (Priority 3)
	if openaiKey != "" {
		fmt.Println("🤖 Using OpenAI")
		callOpenAIStyleAPI(
			"https://api.openai.com/v1/chat/completions",
			"gpt-4o-mini",
			openaiKey,
			systemPrompt,
			userPrompt,
		)
		return
	}

	fmt.Println("❌ No AI API key found. Set GEMINI_API_KEY, DEEPSEEK_API_KEY, or OPENAI_API_KEY")
}

func callOpenAIStyleAPI(apiURL, model, apiKey, systemPrompt, userPrompt string) {
	body := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
	}

	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("Request creation failed:", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("AI request failed:", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result OpenAIStyleResponse
	if err := json.Unmarshal(respBody, &result); err != nil || len(result.Choices) == 0 {
		fmt.Println("Failed to parse AI response")
		fmt.Println(string(respBody))
		return
	}

	PrintFormatted(result.Choices[0].Message.Content)
}
