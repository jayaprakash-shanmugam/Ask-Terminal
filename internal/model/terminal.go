package model

import (
	"askterminal/internal/utils"
	"fmt"
	"os"
)

// Terminal represents our Gemini-enhanced terminal
type Terminal struct {
	prompt       string
	history      []string
	workingDir   string
	running      bool
	apiKey       string
	systemPrompt string
}

// NewTerminal creates a new terminal instance
func NewTerminal() *Terminal {
	wd, _ := os.Getwd()

	// Load API key from environment
	apiKey := utils.LoadAPIKey()
	fmt.Println("API Key successfully loaded:", apiKey[:5]+"...")

	// System prompt that instructs Gemini how to interpret commands
	systemPrompt := `You are an AI assistant that translates natural language into terminal commands.
Your task is to interpret user input and convert it to executable shell commands.
Always respond with a JSON object that contains:
1. "command": The main command to execute
2. "args": An array of command arguments
3. "explanation": A brief explanation of what the command does

For example:
User: "show me all text files"
You: {"command": "find", "args": [".", "-type", "f", "-name", "*.txt"], "explanation": "Finding all text files in the current directory"}

User: "create a new folder called projects"
You: {"command": "mkdir", "args": ["projects"], "explanation": "Creating a new directory named 'projects'"}

If you're not sure about a command or the intent isn't clear, respond with {"command": "echo", "args": ["I'm not sure how to do that"], "explanation": "Unclear command intent"}

For security, never suggest dangerous commands that could harm the system.`

	return &Terminal{
		prompt:       "☸ ",
		history:      []string{},
		workingDir:   wd,
		running:      true,
		apiKey:       apiKey,
		systemPrompt: systemPrompt,
	}
}

// Getters to expose private fields
func (t *Terminal) Prompt() string {
	return t.prompt
}

func (t *Terminal) History() []string {
	return t.history
}

func (t *Terminal) WorkingDir() string {
	return t.workingDir
}

func (t *Terminal) IsRunning() bool {
	return t.running
}

func (t *Terminal) APIKey() string {
	return t.apiKey
}

func (t *Terminal) SystemPrompt() string {
	return t.systemPrompt
}

// Setters
func (t *Terminal) SetRunning(running bool) {
	t.running = running
}

func (t *Terminal) SetWorkingDir(dir string) {
	t.workingDir = dir
}

func (t *Terminal) AddToHistory(command string) {
	t.history = append(t.history, command)
}
