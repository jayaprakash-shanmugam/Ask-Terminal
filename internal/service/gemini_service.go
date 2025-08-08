package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"askterminal/internal/dto"
)

type GeminiService struct{}

func NewGeminiService() *GeminiService {
	return &GeminiService{}
}

// GetCommandFromGemini sends the query to Gemini API and parses the response
func (g *GeminiService) GetCommandFromGemini(ctx context.Context, systemPrompt, query, apiKey string) (dto.CommandResponse, error) {
	// Updated API URL for Gemini
	geminiURL := "https://generativelanguage.googleapis.com/v1/models/gemini-1.5-flash:generateContent"

	// Create request payload with system prompt as first message
	requestBody := dto.GeminiRequest{
		Contents: []dto.Content{
			{
				Parts: []dto.Part{
					{
						Text: systemPrompt + "\n\nUser: " + query,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return dto.CommandResponse{}, err
	}

	// Add API key to URL
	url := fmt.Sprintf("%s?key=%s", geminiURL, apiKey)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return dto.CommandResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return dto.CommandResponse{}, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return dto.CommandResponse{}, err
	}

	// Check for error status code
	if resp.StatusCode != http.StatusOK {
		return dto.CommandResponse{}, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse Gemini response
	var geminiResp dto.GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return dto.CommandResponse{}, err
	}

	// Extract command JSON from text response
	if len(geminiResp.Candidates) == 0 {
		return dto.CommandResponse{}, fmt.Errorf("no response from Gemini")
	}

	cmdText := geminiResp.Candidates[0].Content.Parts[0].Text

	// Clean up the response text to extract valid JSON
	cmdText = strings.TrimSpace(cmdText)
	if strings.HasPrefix(cmdText, "```json") {
		cmdText = strings.TrimPrefix(cmdText, "```json")
		cmdText = strings.TrimSuffix(cmdText, "```")
	} else if strings.HasPrefix(cmdText, "```") {
		cmdText = strings.TrimPrefix(cmdText, "```")
		cmdText = strings.TrimSuffix(cmdText, "```")
	}

	cmdText = strings.TrimSpace(cmdText)

	// Parse the command
	var cmdResponse dto.CommandResponse
	if err := json.Unmarshal([]byte(cmdText), &cmdResponse); err != nil {
		return dto.CommandResponse{}, fmt.Errorf("failed to parse command response: %v\nRaw response: %s", err, cmdText)
	}

	return cmdResponse, nil
}
