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

type AIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func Analyze(errorText string) {
	// Try loading .env (optional)
	_ = godotenv.Load()

	deepseekKey := os.Getenv("DEEPSEEK_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	var apiURL string
	var apiKey string
	var model string

	// Decide provider automatically
	if deepseekKey != "" {
		apiURL = "https://api.deepseek.com/v1/chat/completions"
		apiKey = deepseekKey
		model = "deepseek-chat"
		fmt.Println("🧠 Using DeepSeek AI")
	} else if openaiKey != "" {
		apiURL = "https://api.openai.com/v1/chat/completions"
		apiKey = openaiKey
		model = "gpt-4o-mini"
		fmt.Println("🤖 Using OpenAI")
	} else {
		fmt.Println("❌ No AI API key found. Set OPENAI_API_KEY or DEEPSEEK_API_KEY")
		return
	}

	systemPrompt := `You are a senior Golang debugging expert.
Analyze Go errors and respond ONLY in this format:

ROOT CAUSE:
WHY IT HAPPENED:
HOW TO FIX:
EXAMPLE FIX CODE:`

	userPrompt := "Error:\n" + errorText

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

	var result AIResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil || len(result.Choices) == 0 {
		fmt.Println("Failed to parse AI response")
		fmt.Println(string(respBody))
		return
	}

	PrintFormatted(result.Choices[0].Message.Content)
}
