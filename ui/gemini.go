package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type GeminiCard struct {
	Front string   `json:"front"`
	Back  string   `json:"back"`
	Hint  string   `json:"hint"`
	Tags  []string `json:"tags"`
}

type GeminiResultMsg struct {
	Cards []GeminiCard
	Topic string
	Err   error
}

type GeminiAdviceMsg struct {
	Advice string
	Err    error
}

// Structs for Gemini API request/response
type GeminiRequest struct {
	Contents         []GeminiContent  `json:"contents"`
	GenerationConfig *GeminiGenConfig `json:"generationConfig,omitempty"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiGenConfig struct {
	ResponseMimeType string `json:"responseMimeType"`
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

func CallGeminiGenerate(apiKey, topic string) ([]GeminiCard, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)

	prompt := fmt.Sprintf(`Generate a set of 6-10 high-quality flashcards to study the topic: "%s".
Return a JSON array of objects, where each object has:
- "front": a question, definition prompt, or concept (keep it short, under 120 chars)
- "back": a concise, accurate answer or explanation (under 300 chars)
- "hint": a helpful short hint or mnemonic (optional, leave empty if not needed)
- "tags": a list of 1-3 tags related to the concept (e.g. ["go", "concurrency"])

Example:
[
  {
    "front": "What is a goroutine?",
    "back": "A lightweight thread of execution managed by the Go runtime.",
    "hint": "Managed by runtime, not OS.",
    "tags": ["go", "concurrency"]
  }
]
`, topic)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: &GeminiGenConfig{
			ResponseMimeType: "application/json",
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API status: %s", resp.Status)
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned no content candidates")
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text

	var cards []GeminiCard
	if err := json.Unmarshal([]byte(rawText), &cards); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from Gemini: %v (raw response: %s)", err, rawText)
	}

	return cards, nil
}

func CallGeminiAdvice(apiKey, statsData string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: statsData},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API status: %s", resp.Status)
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no content candidates")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func GenerateCardsCmd(apiKey, topic string) tea.Cmd {
	return func() tea.Msg {
		cards, err := CallGeminiGenerate(apiKey, topic)
		return GeminiResultMsg{Cards: cards, Topic: topic, Err: err}
	}
}

func GetAdviceCmd(apiKey, statsData string) tea.Cmd {
	return func() tea.Msg {
		advice, err := CallGeminiAdvice(apiKey, statsData)
		return GeminiAdviceMsg{Advice: advice, Err: err}
	}
}
