package aierror

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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

	// 🟣 GEMINI
	if geminiKey != "" {
		fmt.Println("✨ Using Google Gemini")

		url := "https://generativelanguage.googleapis.com/v1/models/gemini-1.5-flash:generateContent?key=" + geminiKey

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
		if err == nil {
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)

			var result GeminiResponse
			if json.Unmarshal(respBody, &result) == nil && len(result.Candidates) > 0 {
				PrintFormatted(result.Candidates[0].Content.Parts[0].Text)
				return
			}
		}

		fmt.Println("⚠️ Gemini failed, falling back...")
	}

	// 🔵 DEEPSEEK
	if deepseekKey != "" {
		fmt.Println("🧠 Using DeepSeek AI")
		if callOpenAIStyleAPI("https://api.deepseek.com/v1/chat/completions", "deepseek-chat", deepseekKey, systemPrompt, userPrompt) {
			return
		}
		fmt.Println("⚠️ DeepSeek failed, falling back...")
	}

	// 🟢 OPENAI
	if openaiKey != "" {
		fmt.Println("🤖 Using OpenAI")
		if callOpenAIStyleAPI("https://api.openai.com/v1/chat/completions", "gpt-4o-mini", openaiKey, systemPrompt, userPrompt) {
			return
		}
		fmt.Println("⚠️ OpenAI failed, falling back...")
	}

	// 🖥 OLLAMA FALLBACK (FREE LOCAL AI)
	fmt.Println("🖥 No working cloud AI found, switching to Local Ollama... with phi3 modal")
	callOllama(systemPrompt + "\n" + userPrompt)
}

func callOpenAIStyleAPI(apiURL, model, apiKey, systemPrompt, userPrompt string) bool {
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
		return false
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result OpenAIStyleResponse
	if json.Unmarshal(respBody, &result) != nil || len(result.Choices) == 0 {
		return false
	}

	PrintFormatted(result.Choices[0].Message.Content)
	return true
}

func callOllama(prompt string) {
	body := map[string]interface{}{
		"model":  "phi3",
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"num_predict": 200,
		},
	}

	jsonBody, _ := json.Marshal(body)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("❌ Ollama not running. Install Ollama and run: ollama pull phi3")
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if json.Unmarshal(respBody, &result) != nil {
		fmt.Println("Failed to parse Ollama response")
		fmt.Println(string(respBody))
		return
	}

	if response, ok := result["response"].(string); ok && response != "" {
		PrintFormatted(response)
	} else {
		fmt.Println("⚠️ Ollama returned empty response")
	}
}
