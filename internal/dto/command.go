package dto

// CommandResponse represents the structured output from Gemini
type CommandResponse struct {
	Command     string   `json:"command"`
	Args        []string `json:"args,omitempty"`
	Explanation string   `json:"explanation,omitempty"`
}

// GeminiRequest represents a request to the Gemini API
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

// Content represents the content of a Gemini request
type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

// Part represents a part of a Gemini request content
type Part struct {
	Text string `json:"text"`
}

// GeminiResponse represents a response from the Gemini API
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

// Candidate represents a candidate response from Gemini
type Candidate struct {
	Content Content `json:"content"`
}
