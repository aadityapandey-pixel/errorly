package aierror

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func Analyze(errorText string) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️ OPENAI_API_KEY not set")
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
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
	}

	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
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

	var result OpenAIResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil || len(result.Choices) == 0 {
		fmt.Println("Failed to parse AI response")
		fmt.Println(string(respBody))
		return
	}

	PrintFormatted(result.Choices[0].Message.Content)
}
